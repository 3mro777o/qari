package ui

// Catalog queries: filtering, sorting and choosing recordings.

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/norm"
	"github.com/3mro777o/qari/internal/store"
)

var (
	sortModes   = []string{"popular", "name", "updated"}
	sortNames   = []string{"Popular", "A-Z", "Updated"}
	repeatModes = []string{"next", "one", "off"}
	repeatNames = []string{"auto", "repeat", "stop"}
	repeatLong  = []string{"Autoplay next surah", "Repeat surah", "Stop after surah"}
)

// mode returns the index of cur in modes (0 if unknown).
func mode(modes []string, cur string) int { return max(slices.Index(modes, cur), 0) }

func styleName(mo api.Moshaf) string {
	if i := mo.MoshafType % 10; i >= 1 && i <= 3 {
		return []string{"Murattal", "Mujawwad", "Teaching"}[i-1]
	}
	if _, after, ok := strings.Cut(mo.Name, " - "); ok {
		return strings.TrimSpace(after)
	}
	return ""
}

// riwayah is the recording's riwayah without the "Rewayat" prefix; short
// cuts it to the reader's name ("Hafs").
func riwayah(cat *api.Catalog, mo api.Moshaf, short bool) string {
	name := cat.Riwayat[mo.RewayaID]
	if name == "" {
		name, _, _ = strings.Cut(mo.Name, " - ")
	}
	for _, p := range []string{"Rewayat ", "Riwayat ", "rewayat ", "riwayat "} {
		name = strings.TrimPrefix(strings.TrimSpace(name), p)
	}
	if short {
		for _, sep := range []string{" A'n ", " An ", " 'An ", " an "} {
			if before, _, ok := strings.Cut(name, sep); ok && before != "" {
				return before
			}
		}
	}
	return name
}

// label is "Hafs A'n Assem · Murattal" (short: "Hafs · Murattal").
func label(cat *api.Catalog, mo api.Moshaf, short bool) string {
	r, st := riwayah(cat, mo, short), styleName(mo)
	if r == "" {
		return mo.Name
	}
	if st == "" || strings.Contains(strings.ToLower(r), strings.ToLower(st)) {
		return r
	}
	return r + " · " + st
}

func surahIDs(mo api.Moshaf) []int {
	var out []int
	for n := range mo.Surahs() {
		if n >= 1 && n <= 114 {
			out = append(out, n)
		}
	}
	slices.Sort(out)
	return out
}

func findMoshaf(r *api.ReciterEntry, id int) (api.Moshaf, bool) {
	if r != nil {
		for _, mo := range r.Moshaf {
			if mo.ID == id {
				return mo, true
			}
		}
	}
	return api.Moshaf{}, false
}

// best picks a recording of r: the chosen riwayah, else Hafs, complete,
// Murattal, most surahs. surah > 0 requires that surah; strict drops other
// riwayat.
func best(cat *api.Catalog, r *api.ReciterEntry, rewaya, surah int, strict bool) (api.Moshaf, bool) {
	key := func(mo api.Moshaf) []int {
		hafs := mo.RewayaID == 1 || strings.Contains(strings.ToLower(mo.Name+cat.Riwayat[mo.RewayaID]), "hafs")
		style := map[int]int{1: 3, 2: 2, 3: 1}[mo.MoshafType%10]
		return []int{b2i(rewaya > 0 && mo.RewayaID == rewaya), b2i(hafs), b2i(mo.SurahTotal >= 114), style, mo.SurahTotal, -mo.ID}
	}
	var out api.Moshaf
	found := false
	for _, mo := range r.Moshaf {
		if (strict && rewaya > 0 && mo.RewayaID != rewaya) || (surah > 0 && !mo.Surahs()[surah]) {
			continue
		}
		if !found || slices.Compare(key(mo), key(out)) > 0 {
			out, found = mo, true
		}
	}
	return out, found
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// qaris filters by scope (well-known or all; favorites always), riwayah and
// query, then sorts.
func (m *Model) qaris(query string) []*api.ReciterEntry {
	st := m.Store.Settings
	var out []*api.ReciterEntry
	for _, r := range m.cat.Reciters {
		if (st.ShowAll || wellKnown(r) || m.Store.Favorites[r.ID]) && norm.Match(r.Search, query) &&
			(st.Rewaya == 0 || slices.ContainsFunc(r.Moshaf, func(mo api.Moshaf) bool { return mo.RewayaID == st.Rewaya })) {
			out = append(out, r)
		}
	}
	by := mode(sortModes, st.Sort)
	slices.SortStableFunc(out, func(a, b *api.ReciterEntry) int {
		name := cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
		switch by {
		case 1:
			return cmp.Or(name, a.ID-b.ID)
		case 2:
			return cmp.Or(cmp.Compare(b.Date, a.Date), b.Listens-a.Listens, name)
		}
		return cmp.Or(b.Listens-a.Listens, b2i(b.Known)-b2i(a.Known), name, a.ID-b.ID)
	})
	return out
}

// wellKnown: a complete recording by a reciter with known popularity.
func wellKnown(r *api.ReciterEntry) bool { return r.Complete && (r.Known || r.Listens > 0) }

type pick struct {
	r  *api.ReciterEntry
	mo api.Moshaf
}

// qarisFor lists the qaris who recorded surah n, each with their best recording of it.
func (m *Model) qarisFor(n int, query string) []pick {
	var out []pick
	for _, r := range m.qaris(query) {
		if mo, ok := best(m.cat, r, m.Store.Settings.Rewaya, n, true); ok {
			out = append(out, pick{r, mo})
		}
	}
	return out
}

// surahs filters surahs ids (nil = all 114) by name in any language or number;
// an exact number comes first.
func (m *Model) surahs(ids []int, query string) []api.SurahEntry {
	q := strings.TrimSpace(query)
	num, _ := strconv.Atoi(q)
	var out []api.SurahEntry
	for i, s := range m.cat.Surahs {
		s.ID = i + 1
		match := norm.Match(s.Search, q)
		if num > 0 { // a number matches surah numbers only: 1 finds 1, 10-19, 100-114
			match = strings.HasPrefix(strconv.Itoa(s.ID), q) || strings.HasPrefix(fmt.Sprintf("%03d", s.ID), q)
		}
		if (ids == nil || slices.Contains(ids, s.ID)) && match {
			if s.ID == num {
				out = append([]api.SurahEntry{s}, out...)
			} else {
				out = append(out, s)
			}
		}
	}
	return out
}

type riwayaOption struct {
	id    int
	name  string
	count int
}

func riwayaOptions(cat *api.Catalog) []riwayaOption {
	counts := map[int]int{}
	for _, r := range cat.Reciters {
		seen := map[int]bool{}
		for _, mo := range r.Moshaf {
			if mo.RewayaID > 0 && !seen[mo.RewayaID] {
				seen[mo.RewayaID] = true
				counts[mo.RewayaID]++
			}
		}
	}
	var out []riwayaOption
	for id, c := range counts {
		out = append(out, riwayaOption{id, cmp.Or(riwayah(cat, api.Moshaf{RewayaID: id}, false), fmt.Sprint("Riwayah ", id)), c})
	}
	slices.SortFunc(out, func(a, b riwayaOption) int { return cmp.Or(b.count-a.count, cmp.Compare(a.name, b.name)) })
	return append([]riwayaOption{{0, "All riwayat", len(cat.Reciters)}}, out...)
}

// heard returns, per surah, the best listened fraction among progress
// entries that match (1 = finished).
func (m *Model) heard(match func(p *store.Progress) bool) map[int]float64 {
	out := map[int]float64{}
	for _, p := range m.Store.Progress {
		if match(p) && (p.Finished || p.Position >= 5) {
			out[p.Surah] = max(out[p.Surah], p.Fraction())
		}
	}
	return out
}

func finished(h map[int]float64) (n int) {
	for _, f := range h {
		n += b2i(f >= 1)
	}
	return n
}

func revelation(s api.SurahEntry) string {
	if s.Makkia == 1 {
		return "Meccan"
	}
	return "Medinan"
}

func pages(s api.SurahEntry) string {
	switch {
	case s.StartPage <= 0:
		return ""
	case s.EndPage <= s.StartPage:
		return fmt.Sprintf("p.%d", s.StartPage)
	}
	return fmt.Sprintf("pp.%d-%d", s.StartPage, s.EndPage)
}
