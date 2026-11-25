package store

import (
	"path/filepath"
	"testing"
)

func TestResumeRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	tr := Track{Reciter: 1, Moshaf: 2, Surah: 18}
	s.Update(tr, 600, 3000)
	s.Favorites[7] = true
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	s2, _ := Open(path)
	if got := s2.ResumeAt(tr); got != 597 {
		t.Fatalf("ResumeAt = %v, want 597", got)
	}
	if !s2.Favorites[7] || len(s2.Recent(5)) != 1 {
		t.Fatal("favorites or recent lost")
	}
	s2.Update(tr, 2995, 3000)
	if s2.ResumeAt(tr) != 0 || len(s2.Recent(5)) != 0 {
		t.Fatal("finished surah should restart and leave Recent")
	}
}
