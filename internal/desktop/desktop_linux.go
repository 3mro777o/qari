package desktop

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

const (
	path      = "/org/mpris/MediaPlayer2"
	rootIface = "org.mpris.MediaPlayer2"
	playIface = "org.mpris.MediaPlayer2.Player"
)

// Desktop is an MPRIS player on the session bus. A nil *Desktop is valid
// and does nothing (no session bus, or not Linux).
type Desktop struct {
	Cmds   chan Cmd
	conn   *dbus.Conn
	props  *prop.Properties
	notify bool

	mu      sync.Mutex
	last    State
	lastAt  time.Time
	track   int
	notifID uint32
}

// Start registers org.mpris.MediaPlayer2.qari, or nil if there is no session bus.
func Start(notify bool) *Desktop {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil
	}
	d := &Desktop{Cmds: make(chan Cmd, 8), conn: conn, notify: notify, last: State{Status: "Stopped"}}
	ro := func(v any) *prop.Prop { return &prop.Prop{Value: v, Emit: prop.EmitTrue} }
	d.props, err = prop.Export(conn, path, prop.Map{
		rootIface: {
			"CanQuit": ro(true), "CanRaise": ro(false), "HasTrackList": ro(false),
			"Identity": ro("qari"), "SupportedUriSchemes": ro([]string{}), "SupportedMimeTypes": ro([]string{}),
		},
		playIface: {
			"PlaybackStatus": ro("Stopped"), "Metadata": ro(map[string]dbus.Variant{}),
			"Rate": ro(1.0), "MinimumRate": ro(0.75), "MaximumRate": ro(2.0), "Volume": ro(1.0),
			"Position":  {Value: int64(0), Emit: prop.EmitFalse},
			"CanGoNext": ro(false), "CanGoPrevious": ro(false), "CanPlay": ro(true), "CanPause": ro(true),
			"CanSeek": ro(true), "CanControl": ro(true),
		},
	})
	if err != nil {
		conn.Close()
		return nil
	}
	conn.Export(root{d}, path, rootIface)
	// Seek is named SeekTo in Go so it does not look like io.Seeker to vet.
	conn.ExportWithMap(player{d}, map[string]string{"SeekTo": "Seek"}, path, playIface)
	methods := introspect.Methods(player{})
	for i := range methods {
		if methods[i].Name == "SeekTo" {
			methods[i].Name = "Seek"
		}
	}
	conn.Export(introspect.NewIntrospectable(&introspect.Node{Name: path, Interfaces: []introspect.Interface{
		introspect.IntrospectData, prop.IntrospectData,
		{Name: rootIface, Methods: introspect.Methods(root{}), Properties: d.props.Introspection(rootIface)},
		{Name: playIface, Methods: methods, Properties: d.props.Introspection(playIface),
			Signals: []introspect.Signal{{Name: "Seeked", Args: []introspect.Arg{{Name: "Position", Type: "x"}}}}},
	}}), path, "org.freedesktop.DBus.Introspectable")
	name := "org.mpris.MediaPlayer2.qari"
	if r, err := conn.RequestName(name, dbus.NameFlagDoNotQueue); err != nil || r != dbus.RequestNameReplyPrimaryOwner {
		// A second qari: MPRIS names other instances like this.
		conn.RequestName(fmt.Sprintf("%s.instance%d", name, os.Getpid()), dbus.NameFlagDoNotQueue)
	}
	return d
}

func (d *Desktop) Close() {
	if d != nil {
		d.conn.Close()
	}
}

// Set publishes s, emitting only what changed. A jump in position that
// playback does not explain is announced as a seek.
func (d *Desktop) Set(s State) {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	old, now := d.last, time.Now()
	set := func(name string, changed bool, v any) {
		if changed {
			d.props.SetMust(playIface, name, v)
		}
	}
	newTrack := s.Title != old.Title
	if newTrack {
		d.track++
	}
	set("PlaybackStatus", s.Status != old.Status, s.Status)
	set("Metadata", newTrack || s.Artist != old.Artist || int(s.Length) != int(old.Length), d.metadata(s))
	set("Rate", s.Rate != old.Rate, s.Rate)
	set("Volume", s.Volume != old.Volume, s.Volume)
	set("CanGoNext", s.CanNext != old.CanNext, s.CanNext)
	set("CanGoPrevious", s.CanPrev != old.CanPrev, s.CanPrev)
	pos := int64(s.Pos * 1e6)
	d.props.SetMust(playIface, "Position", pos)
	expect := old.Pos
	if old.Status == "Playing" {
		expect += now.Sub(d.lastAt).Seconds() * max(old.Rate, 1)
	}
	if !newTrack && s.Status != "Stopped" && math.Abs(s.Pos-expect) > 3 {
		d.conn.Emit(path, playIface+".Seeked", pos)
	}
	if newTrack && s.Title != "" && d.notify {
		go d.notifyNow(s.Title, s.Artist)
	}
	d.last, d.lastAt = s, now
}

func (d *Desktop) metadata(s State) map[string]dbus.Variant {
	if s.Title == "" {
		return map[string]dbus.Variant{}
	}
	return map[string]dbus.Variant{
		"mpris:trackid": dbus.MakeVariant(dbus.ObjectPath(fmt.Sprintf("/org/qari/track/%d", d.track))),
		"xesam:title":   dbus.MakeVariant(s.Title),
		"xesam:artist":  dbus.MakeVariant([]string{s.Artist}),
		"xesam:album":   dbus.MakeVariant(s.Album),
		"mpris:length":  dbus.MakeVariant(int64(s.Length * 1e6)),
	}
}

// notifyNow shows a desktop notification, replacing the previous one.
func (d *Desktop) notifyNow(summary, body string) {
	d.mu.Lock()
	id := d.notifID
	d.mu.Unlock()
	obj := d.conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")
	call := obj.Call("org.freedesktop.Notifications.Notify", 0, "qari", id, "audio-x-generic",
		summary, body, []string{}, map[string]dbus.Variant{}, int32(5000))
	if call.Err == nil {
		d.mu.Lock()
		call.Store(&d.notifID)
		d.mu.Unlock()
	}
}

func (d *Desktop) send(name string, arg float64) {
	select {
	case d.Cmds <- Cmd{name, arg}:
	default:
	}
}

// root and player hold only the exported D-Bus methods of each interface.
type root struct{ d *Desktop }

func (r root) Raise() *dbus.Error { return nil }
func (r root) Quit() *dbus.Error  { r.d.send("quit", 0); return nil }

type player struct{ d *Desktop }

func (p player) PlayPause() *dbus.Error      { p.d.send("playpause", 0); return nil }
func (p player) Play() *dbus.Error           { p.d.send("play", 0); return nil }
func (p player) Pause() *dbus.Error          { p.d.send("pause", 0); return nil }
func (p player) Next() *dbus.Error           { p.d.send("next", 0); return nil }
func (p player) Previous() *dbus.Error       { p.d.send("previous", 0); return nil }
func (p player) Stop() *dbus.Error           { p.d.send("stop", 0); return nil }
func (p player) SeekTo(us int64) *dbus.Error { p.d.send("seek", float64(us)/1e6); return nil }
func (p player) OpenUri(string) *dbus.Error  { return nil }
func (p player) SetPosition(_ dbus.ObjectPath, us int64) *dbus.Error {
	p.d.send("setpos", float64(us)/1e6)
	return nil
}
