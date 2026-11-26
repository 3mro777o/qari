// Package store keeps the listener's state: favourites, where each surah was
// left off, and listening history. It is one small JSON file.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Track identifies one surah in one moshaf of one reciter.
type Track struct {
	Reciter int `json:"reciter"`
	Moshaf  int `json:"moshaf"`
	Surah   int `json:"surah"`
}

func (t Track) key() string { return fmt.Sprintf("%d/%d/%d", t.Reciter, t.Moshaf, t.Surah) }

type Progress struct {
	Track
	Position float64   `json:"position"` // seconds
	Duration float64   `json:"duration"`
	Finished bool      `json:"finished"`
	Updated  time.Time `json:"updated"`
}

// Fraction is how much of the track was heard, 0..1.
func (p Progress) Fraction() float64 {
	if p.Finished {
		return 1
	}
	if p.Duration <= 0 {
		return 0
	}
	return min(p.Position/p.Duration, 1)
}

type Data struct {
	Favorites map[int]bool         `json:"favorites"` // reciter ids
	Progress  map[string]*Progress `json:"progress"`
	Plays     map[int]int          `json:"plays"` // reciter id -> surahs started
	Last      *Track               `json:"last,omitempty"`
	Settings  Settings             `json:"settings"`
}

type Settings struct {
	ShowAll bool   `json:"show_all"` // show incomplete and little-known reciters
	Sort    string `json:"sort"`
	Rewaya  int    `json:"rewaya"`
	Repeat  string `json:"repeat"`
	Theme   string `json:"theme,omitempty"`
}

type Store struct {
	mu   sync.Mutex
	path string
	Data
}

func Open(path string) (*Store, error) {
	s := &Store{path: path}
	b, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(b, &s.Data); err != nil {
			// Keep the unreadable file instead of overwriting the user's history.
			os.Rename(path, path+".broken")
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if s.Favorites == nil {
		s.Favorites = map[int]bool{}
	}
	if s.Progress == nil {
		s.Progress = map[string]*Progress{}
	}
	if s.Plays == nil {
		s.Plays = map[int]int{}
	}
	return s, nil
}

func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := json.MarshalIndent(s.Data, "", " ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) Get(t Track) *Progress { return s.Progress[t.key()] }

// Update records the playback position of t.
func (s *Store) Update(t Track, pos, dur float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.Progress[t.key()]
	if p == nil {
		p = &Progress{Track: t}
		s.Progress[t.key()] = p
	}
	p.Position, p.Updated = pos, time.Now()
	if dur > 0 {
		p.Duration = dur
	}
	// Within the last 10 seconds counts as finished.
	if p.Duration > 0 && pos >= p.Duration-10 {
		p.Finished = true
	}
	s.Last = &t
}

// Forget drops t's saved position, so it leaves Continue Listening.
func (s *Store) Forget(t Track) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Progress, t.key())
	if s.Last != nil && *s.Last == t {
		s.Last = nil
	}
}

func (s *Store) MarkFinished(t Track) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p := s.Progress[t.key()]; p != nil {
		p.Finished, p.Position = true, 0
	}
}

// ResumeAt is where playback of t should start: the saved position, unless
// the surah was finished or barely started.
func (s *Store) ResumeAt(t Track) float64 {
	p := s.Get(t)
	if p == nil || p.Finished || p.Position < 5 {
		return 0
	}
	return max(p.Position-3, 0) // a little overlap helps pick the thread back up
}

// Recent returns unfinished tracks, most recent first.
func (s *Store) Recent(n int) []Progress {
	var out []Progress
	for _, p := range s.Progress {
		if !p.Finished && p.Position >= 5 {
			out = append(out, *p)
		}
	}
	for i := 1; i < len(out); i++ { // small n, insertion sort is fine
		for j := i; j > 0 && out[j].Updated.After(out[j-1].Updated); j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	if len(out) > n {
		out = out[:n]
	}
	return out
}
