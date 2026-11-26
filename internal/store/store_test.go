package store

import (
	"path/filepath"
	"slices"
	"testing"
	"time"
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

func TestRecipesAndHistoryRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s, _ := Open(path)
	s.Recipes = []Recipe{{Name: "Short", Known: true, Riwayat: []int{1}, Surahs: []int{78, 79}}}
	tr := Track{Reciter: 1, Moshaf: 2, Surah: 18}
	for i := range HistoryMax + 5 {
		s.Log(Entry{Time: time.Now(), Track: Track{Reciter: 1, Moshaf: 2, Surah: i%114 + 1}})
	}
	s.Log(Entry{Track: tr, Recipe: "Short"})
	s.Log(Entry{Track: tr, Recipe: "Short"}) // a restart refreshes, not duplicates
	s.Log(Entry{Radio: 7})
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	s2, _ := Open(path)
	if len(s2.Recipes) != 1 || !slices.Equal(s2.Recipes[0].Surahs, []int{78, 79}) || !s2.Recipes[0].Known {
		t.Fatalf("recipes: %+v", s2.Recipes)
	}
	h := s2.History
	if len(h) != HistoryMax || h[len(h)-1].Radio != 7 || h[len(h)-2].Track != tr || h[len(h)-2].Recipe != "Short" || h[len(h)-3].Track == tr {
		t.Fatalf("history: %d entries, tail %+v", len(h), h[len(h)-3:])
	}
}
