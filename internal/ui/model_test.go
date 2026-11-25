package ui

import (
	"path/filepath"
	"testing"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/store"
)

func TestMoshafOrder(t *testing.T) {
	st, _ := store.Open(filepath.Join(t.TempDir(), "s.json"))
	m := New(Options{Store: st})
	r := &api.ReciterEntry{Reciter: api.Reciter{Moshaf: []api.Moshaf{
		{ID: 1, RewayaID: 5, MoshafType: 51, SurahTotal: 114}, // Qalon, complete
		{ID: 2, RewayaID: 1, MoshafType: 13, SurahTotal: 114}, // Hafs, Mo'lim
		{ID: 3, RewayaID: 1, MoshafType: 11, SurahTotal: 114}, // Hafs, Murattal
		{ID: 4, RewayaID: 1, MoshafType: 11, SurahTotal: 30},  // Hafs, partial
	}}}
	got := []int{}
	for _, ms := range m.moshafsFor(r) {
		got = append(got, ms.ID)
	}
	if want := []int{3, 2, 4, 1}; !equal(got, want) {
		t.Errorf("default order %v, want %v", got, want)
	}
	st.Settings.Rewaya = 5
	if first := m.moshafsFor(r)[0].ID; first != 1 {
		t.Errorf("with Qalon chosen, first = %d, want 1", first)
	}
}

func TestNextSurah(t *testing.T) {
	ms := &api.Moshaf{SurahList: "1,18,36,114"}
	cases := [][3]int{{1, 1, 18}, {18, 1, 36}, {114, 1, 0}, {36, -1, 18}, {1, -1, 0}}
	for _, c := range cases {
		if got := nextSurah(ms, c[0], c[1]); got != c[2] {
			t.Errorf("nextSurah(%d, %d) = %d, want %d", c[0], c[1], got, c[2])
		}
	}
}

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
