// Package player drives one long-lived mpv process over its JSON IPC socket.
package player

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Event is sent to the UI whenever playback state changes.
type Event struct {
	Pos, Duration float64
	Paused        bool
	Loading       bool  // a file is opening and has not started yet
	Ended         bool  // the current file played to the end
	Err           error // the file failed to play or mpv exited
	Tag           int   // the tag of the Play call this event belongs to
}

type Player struct {
	Events chan Event

	mu     sync.Mutex
	cmd    *exec.Cmd
	conn   net.Conn
	sock   string
	state  Event
	relay  *Relay
	speed  float64
	volume float64
	buffer int // MiB of stream mpv may hold, ahead and behind

	precise bool // report every position change, not whole seconds
}

// New creates a player. relay may be nil; when set, stream URLs are routed
// through it (mpv itself cannot use SOCKS proxies).
func New(relay *Relay) *Player {
	return &Player{Events: make(chan Event, 64), relay: relay, speed: 1, volume: 100, buffer: 150}
}

// SetBuffer limits how much of a stream mpv keeps in memory, in MiB.
func (p *Player) SetBuffer(mb int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.buffer = mb
	if p.conn != nil {
		p.send("set_property", "demuxer-max-bytes", fmt.Sprintf("%dMiB", mb))
		p.send("set_property", "demuxer-max-back-bytes", fmt.Sprintf("%dMiB", mb))
	}
}

// Available reports whether mpv is installed.
func Available() bool {
	_, err := exec.LookPath("mpv")
	return err == nil
}

func (p *Player) start() error {
	if p.conn != nil {
		return nil
	}
	dir, err := os.MkdirTemp("", "qari-")
	if err != nil {
		return err
	}
	p.sock = ipcPath(dir)
	p.cmd = exec.Command("mpv",
		"--idle=yes", "--no-video", "--no-terminal", "--keep-open=no",
		"--no-resume-playback", "--save-position-on-quit=no",
		// qari publishes its own MPRIS player; scripts such as mpv-mpris
		// would register a second one.
		"--load-scripts=no",
		"--input-ipc-server="+p.sock,
		fmt.Sprintf("--demuxer-max-bytes=%dMiB", p.buffer),
		fmt.Sprintf("--demuxer-max-back-bytes=%dMiB", p.buffer),
	)
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("starting mpv: %w", err)
	}
	for i := 0; i < 100; i++ { // wait up to 5s for the socket
		if p.conn, err = dialIPC(p.sock); err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if p.conn == nil {
		p.cmd.Process.Kill()
		return errors.New("mpv did not open its IPC socket")
	}
	go p.read(p.conn)
	go func(cmd *exec.Cmd, dir string) {
		cmd.Wait()
		os.RemoveAll(dir)
		p.mu.Lock()
		if p.cmd == cmd {
			p.conn, p.cmd = nil, nil
			p.mu.Unlock()
			p.emit(Event{Err: errors.New("mpv exited")})
			return
		}
		p.mu.Unlock()
	}(p.cmd, dir)
	for i, prop := range []string{"time-pos", "duration", "pause"} {
		p.send("observe_property", i+1, prop)
	}
	return nil
}

func (p *Player) send(args ...any) error {
	if p.conn == nil {
		return errors.New("mpv is not running")
	}
	b, _ := json.Marshal(map[string]any{"command": args})
	_, err := p.conn.Write(append(b, '\n'))
	return err
}

func (p *Player) read(conn net.Conn) {
	sc := bufio.NewScanner(conn)
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	for sc.Scan() {
		var m struct {
			Event  string          `json:"event"`
			Name   string          `json:"name"`
			Data   json.RawMessage `json:"data"`
			Reason string          `json:"reason"`
			Error  string          `json:"file_error"`
		}
		if json.Unmarshal(sc.Bytes(), &m) != nil {
			continue
		}
		p.mu.Lock()
		s := &p.state
		s.Ended, s.Err = false, nil
		switch m.Event {
		case "property-change":
			var f float64
			var b bool
			switch m.Name {
			case "time-pos":
				if s.Loading {
					p.mu.Unlock()
					continue // still the previous file's position
				}
				if json.Unmarshal(m.Data, &f) == nil {
					// mpv reports many times a second; the UI only needs whole seconds.
					same := int(f) == int(s.Pos) && !s.Loading && !p.precise
					s.Pos = f
					if same {
						p.mu.Unlock()
						continue
					}
				}
			case "duration":
				if json.Unmarshal(m.Data, &f) == nil {
					s.Duration = f
				}
			case "pause":
				if json.Unmarshal(m.Data, &b) == nil {
					s.Paused = b
				}
			}
		case "end-file":
			switch m.Reason {
			case "eof":
				s.Ended = true
			case "error":
				s.Err = fmt.Errorf("could not play: %s", m.Error)
				s.Loading = false
			default:
				p.mu.Unlock()
				continue // stopped or replaced on purpose
			}
		case "file-loaded":
			s.Loading = false
		default:
			p.mu.Unlock()
			continue
		}
		ev := *s
		p.mu.Unlock()
		p.emit(ev)
	}
}

func (p *Player) emit(ev Event) {
	select {
	case p.Events <- ev:
	default: // UI is behind; position updates are safe to drop
		if ev.Ended || ev.Err != nil {
			p.Events <- ev
		}
	}
}

// Play loads url (or a local file), starting at start seconds. title is
// mpv's media title. Events from now on carry tag, so the caller can drop
// ones still queued from the previous file.
func (p *Player) Play(url string, start float64, title string, tag int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.start(); err != nil {
		return err
	}
	if p.relay != nil && strings.HasPrefix(url, "http") {
		url = p.relay.Wrap(url)
	}
	p.state = Event{Loading: true, Tag: tag}
	// Options are set as properties first: loadfile's per-file options moved
	// between mpv versions, properties work on all of them.
	p.send("set_property", "start", fmt.Sprintf("%.1f", start))
	p.send("set_property", "force-media-title", title)
	p.send("set_property", "pause", false)
	return p.send("loadfile", url, "replace")
}

func (p *Player) TogglePause()       { p.cmdLocked("cycle", "pause") }
func (p *Player) Seek(sec float64)   { p.cmdLocked("seek", sec, "relative") }
func (p *Player) SeekTo(sec float64) { p.cmdLocked("seek", sec, "absolute") }
func (p *Player) Stop()              { p.cmdLocked("stop") }

// AddVolume changes volume by d percent and returns the new value.
func (p *Player) AddVolume(d float64) float64 {
	p.mu.Lock()
	p.volume = min(max(p.volume+d, 0), 130)
	v := p.volume
	p.mu.Unlock()
	p.cmdLocked("set_property", "volume", v)
	return v
}

func (p *Player) SetMute(on bool)  { p.cmdLocked("set_property", "mute", on) }
func (p *Player) SetPause(on bool) { p.cmdLocked("set_property", "pause", on) }

// Precise makes position events sub-second, for ayah repeat.
func (p *Player) Precise(on bool) {
	p.mu.Lock()
	p.precise = on
	p.mu.Unlock()
}

// SetSpeed sets playback speed (pitch is preserved by mpv).
func (p *Player) SetSpeed(s float64) {
	p.mu.Lock()
	p.speed = s
	p.mu.Unlock()
	p.cmdLocked("set_property", "speed", s)
}

func (p *Player) cmdLocked(args ...any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.send(args...)
}

// Close quits mpv.
func (p *Player) Close() {
	p.mu.Lock()
	cmd := p.cmd
	p.send("quit")
	p.cmd = nil
	p.mu.Unlock()
	if cmd != nil {
		done := make(chan struct{})
		go func() { cmd.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			cmd.Process.Kill()
		}
	}
}
