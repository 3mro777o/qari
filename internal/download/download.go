// Package download saves surahs to disk for offline listening, one file at
// a time. A file is written under a temporary name and renamed only when
// complete, so a cancel or crash never leaves a broken surah behind.
package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/3mro777o/qari/internal/store"
)

type State int

const (
	Queued State = iota
	Active
	Failed
)

type Job struct {
	store.Track
	URL        string
	State      State
	Done, Size int64   // bytes; Size is -1 when unknown
	Speed      float64 // bytes per second
	Err        error
	cancel     context.CancelFunc
}

// Group is one saved recording: how many of its surahs are on disk.
type Group struct {
	Reciter, Moshaf, Surahs int
	Bytes                   int64
}

type Manager struct {
	Dir     string
	Changed chan struct{} // receives after any change worth redrawing
	client  *http.Client

	mu       sync.Mutex
	jobs     []*Job
	have     map[store.Track]int64
	batch    int // jobs queued since the queue was last empty
	finished int
	busy     bool
}

// New indexes the surahs already saved in dir (<qari>/<recording>/<NNN>.mp3).
func New(dir string, client *http.Client) *Manager {
	m := &Manager{Dir: dir, client: client, Changed: make(chan struct{}, 1), have: map[store.Track]int64{}}
	filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		var t store.Track
		rel, _ := filepath.Rel(dir, p)
		if err == nil && !d.IsDir() {
			if n, _ := fmt.Sscanf(filepath.ToSlash(rel), "%d/%d/%d.mp3", &t.Reciter, &t.Moshaf, &t.Surah); n == 3 && strings.HasSuffix(p, ".mp3") {
				if info, err := d.Info(); err == nil {
					m.have[t] = info.Size()
				}
			}
		}
		return nil
	})
	return m
}

func (m *Manager) Path(t store.Track) string {
	return filepath.Join(m.Dir, strconv.Itoa(t.Reciter), strconv.Itoa(t.Moshaf), fmt.Sprintf("%03d.mp3", t.Surah))
}

func (m *Manager) Has(t store.Track) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.have[t]
	return ok
}

// Add queues t. It returns false if t is already saved or queued.
func (m *Manager) Add(t store.Track, url string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.have[t]; ok {
		return false
	}
	if j := m.find(t); j != nil {
		if j.State != Failed {
			return false
		}
		j.State, j.Err = Queued, nil // retry
	} else {
		m.jobs = append(m.jobs, &Job{Track: t, URL: url, Size: -1})
	}
	m.batch++
	if !m.busy {
		m.busy = true
		go m.run()
	}
	m.notify()
	return true
}

// Cancel stops and forgets the job for t, if any.
func (m *Manager) Cancel(t store.Track) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	j := m.find(t)
	if j == nil {
		return false
	}
	if j.cancel != nil {
		j.cancel()
	}
	if j.State != Failed {
		m.batch--
	}
	m.drop(j)
	m.notify()
	return true
}

// Delete removes the saved file for t.
func (m *Manager) Delete(t store.Track) error {
	m.mu.Lock()
	delete(m.have, t)
	m.mu.Unlock()
	p := m.Path(t)
	err := os.Remove(p)
	os.Remove(filepath.Dir(p)) // only succeeds once the folders are empty
	os.Remove(filepath.Dir(filepath.Dir(p)))
	m.notify()
	return err
}

// Jobs returns a copy of the queue: active and queued first, then failed.
func (m *Manager) Jobs() []Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Job, 0, len(m.jobs))
	for _, j := range m.jobs {
		out = append(out, *j)
	}
	rank := map[State]int{Active: 0, Queued: 1, Failed: 2}
	slices.SortStableFunc(out, func(a, b Job) int { return rank[a.State] - rank[b.State] })
	return out
}

// Progress reports the current batch: finished and total jobs, the overall
// fraction, and the current speed.
func (m *Manager) Progress() (done, total int, frac, speed float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	part := 0.0
	for _, j := range m.jobs {
		if j.State == Active {
			speed = j.Speed
			if j.Size > 0 {
				part = float64(j.Done) / float64(j.Size)
			}
		}
	}
	if m.batch > 0 {
		frac = (float64(m.finished) + part) / float64(m.batch)
	}
	return m.finished, m.batch, frac, speed
}

// Saved lists saved recordings, ordered by qari and recording id.
func (m *Manager) Saved() []Group {
	m.mu.Lock()
	idx := map[[2]int]*Group{}
	for t, n := range m.have {
		g := idx[[2]int{t.Reciter, t.Moshaf}]
		if g == nil {
			g = &Group{Reciter: t.Reciter, Moshaf: t.Moshaf}
			idx[[2]int{t.Reciter, t.Moshaf}] = g
		}
		g.Surahs++
		g.Bytes += n
	}
	m.mu.Unlock()
	out := make([]Group, 0, len(idx))
	for _, g := range idx {
		out = append(out, *g)
	}
	slices.SortFunc(out, func(a, b Group) int {
		if a.Reciter != b.Reciter {
			return a.Reciter - b.Reciter
		}
		return a.Moshaf - b.Moshaf
	})
	return out
}

func (m *Manager) find(t store.Track) *Job {
	for _, j := range m.jobs {
		if j.Track == t {
			return j
		}
	}
	return nil
}

func (m *Manager) drop(j *Job) {
	m.jobs = slices.DeleteFunc(m.jobs, func(x *Job) bool { return x == j })
	m.settle()
}

// settle starts a new batch count once nothing is left to download.
func (m *Manager) settle() {
	if !slices.ContainsFunc(m.jobs, func(x *Job) bool { return x.State != Failed }) {
		m.batch, m.finished = 0, 0
	}
}

func (m *Manager) notify() {
	select {
	case m.Changed <- struct{}{}:
	default:
	}
}

func (m *Manager) run() {
	for {
		m.mu.Lock()
		i := slices.IndexFunc(m.jobs, func(j *Job) bool { return j.State == Queued })
		if i < 0 {
			m.busy = false
			m.mu.Unlock()
			return
		}
		j := m.jobs[i]
		ctx, cancel := context.WithCancel(context.Background())
		j.State, j.cancel = Active, cancel
		m.mu.Unlock()

		err := m.fetch(ctx, j)
		cancelled := ctx.Err() != nil
		cancel()
		m.mu.Lock()
		switch {
		case cancelled: // Cancel already dropped it
		case err != nil:
			j.State, j.Err, j.cancel = Failed, err, nil
			m.batch--
			m.settle()
		default:
			m.have[j.Track] = j.Done
			m.finished++
			m.drop(j)
		}
		m.notify()
		m.mu.Unlock()
	}
}

func (m *Manager) fetch(ctx context.Context, j *Job) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, j.URL, nil)
	if err != nil {
		return err
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/") {
		return errors.New("not audio (blocked by a network filter?)")
	}
	path := m.Path(j.Track)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	m.mu.Lock()
	j.Size = resp.ContentLength
	m.mu.Unlock()
	buf := make([]byte, 64<<10)
	last, lastDone := time.Now(), int64(0)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, err = f.Write(buf[:n]); err != nil {
				break
			}
		}
		m.mu.Lock()
		j.Done += int64(n)
		if dt := time.Since(last).Seconds(); dt >= 0.5 {
			v := float64(j.Done-lastDone) / dt
			if j.Speed > 0 {
				v = 0.6*j.Speed + 0.4*v
			}
			j.Speed, last, lastDone = v, time.Now(), j.Done
			m.notify()
		}
		done, size := j.Done, j.Size
		m.mu.Unlock()
		if rerr == io.EOF {
			if size > 0 && done != size {
				err = io.ErrUnexpectedEOF
			}
			break
		}
		if rerr != nil {
			err = rerr
			break
		}
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}
