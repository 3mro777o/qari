package ui

// Direct commands from the command line (qari play, radio, random), the
// one-line view of --no-ui, and names for Tab completion.

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/norm"
	"github.com/3mro777o/qari/internal/store"
	tea "github.com/charmbracelet/bubbletea"
)

// Start is what to play as soon as the app opens.
type Start struct {
	Entry  store.Entry   // a surah (Track) or a station (Radio)
	Recipe *store.Recipe // random play from this recipe instead
	begun  bool
}

// Notice is a message shown for a while in the footer, such as a new release.
type Notice string

type startMsg struct{}

// NotFound is the error for a name that matches nothing, or several things.
type NotFound struct {
	What, Query string
	Candidates  []string
}

func (e *NotFound) Error() string {
	if len(e.Candidates) == 0 {
		return fmt.Sprintf("no %s matches %q", e.What, e.Query)
	}
	return fmt.Sprintf("%q matches several %ss, be more specific:\n  %s", e.Query, e.What, strings.Join(e.Candidates, "\n  "))
}

type named struct{ name, search string }

// choose picks the item q names: one whose name is q, else the only one
// with q as whole words, else the only one matching q at all.
func choose(what, q string, items []named) (int, error) {
	var hits, words []int
	qw := strings.Fields(norm.Fold(q))
	for i, it := range items {
		if norm.Compact(it.name) == norm.Compact(q) {
			return i, nil
		}
		if !norm.Match(it.search, q) {
			continue
		}
		hits = append(hits, i)
		if f := strings.Fields(norm.Fold(it.name)); !slices.ContainsFunc(qw, func(w string) bool { return !slices.Contains(f, w) }) {
			words = append(words, i)
		}
	}
	switch {
	case len(words) == 1:
		return words[0], nil
	case len(hits) == 1:
		return hits[0], nil
	}
	e := &NotFound{What: what, Query: q}
	for i, h := range hits {
		if i == 5 {
			e.Candidates = append(e.Candidates, fmt.Sprintf("... and %d more", len(hits)-5))
			break
		}
		e.Candidates = append(e.Candidates, items[h].name)
	}
	return -1, e
}

// Resolve turns a command (play, radio or random) and its words into what
// to play. o.Catalog must be loaded.
func Resolve(o Options, cmd string, args []string) (*Start, error) {
	m := New(o)
	q := strings.Join(args, " ")
	switch cmd {
	case "play":
		return m.resolvePlay(args)
	case "radio":
		if q == "" {
			return nil, errors.New("name a station, e.g. qari radio cairo")
		}
		i, err := choose("station", q, m.stationItems())
		if err != nil {
			return nil, err
		}
		return &Start{Entry: store.Entry{Radio: m.cat.Radios[i].ID}}, nil
	case "random":
		rs := m.recipes()
		i := max(slices.IndexFunc(rs, func(r store.Recipe) bool { return r.Name == m.Store.Settings.Recipe }), 0)
		if q != "" {
			var items []named
			for _, r := range rs {
				items = append(items, named{r.Name, norm.Fold(r.Name)})
			}
			var err error
			if i, err = choose("recipe", q, items); err != nil {
				return nil, err
			}
		}
		return &Start{Recipe: &rs[i]}, nil
	}
	return nil, fmt.Errorf("unknown command %q", cmd)
}

// resolvePlay reads "<qari words> [surah words]": the longest leading words
// that name a qari, so names work without quotes.
func (m *Model) resolvePlay(args []string) (*Start, error) {
	if len(args) == 0 {
		return nil, errors.New("name a qari, e.g. qari play alafasi 18")
	}
	var firstErr error
	for i := len(args); i >= 1; i-- {
		rec, err := m.findQari(strings.Join(args[:i], " "))
		if err == nil {
			var s *Start
			if s, err = m.startFor(rec, strings.Join(args[i:], " ")); err == nil {
				return s, nil
			}
		}
		// Report the surah of the longest qari match, else the qari itself.
		if firstErr == nil && (rec != nil || i == 1) {
			firstErr = err
		}
	}
	return nil, firstErr
}

func (m *Model) qariItems() ([]*api.ReciterEntry, []named) {
	recs := slices.Clone(m.cat.Reciters)
	slices.SortStableFunc(recs, func(a, b *api.ReciterEntry) int { return cmp.Or(b.Listens-a.Listens, cmp.Compare(a.Name, b.Name)) })
	var items []named
	for _, r := range recs {
		items = append(items, named{r.Name, r.Search})
	}
	return recs, items
}

func (m *Model) findQari(q string) (*api.ReciterEntry, error) {
	if id, err := strconv.Atoi(q); err == nil && m.cat.Reciter(id) != nil {
		return m.cat.Reciter(id), nil
	}
	recs, items := m.qariItems()
	i, err := choose("qari", q, items)
	if err != nil {
		return nil, err
	}
	return recs[i], nil
}

func (m *Model) findSurah(q string) (int, error) {
	if n, err := strconv.Atoi(q); err == nil {
		if n < 1 || n > 114 {
			return 0, fmt.Errorf("surah %d does not exist (1-114)", n)
		}
		return n, nil
	}
	var items []named
	for _, s := range m.cat.Surahs {
		items = append(items, named{s.Name, s.Search})
	}
	i, err := choose("surah", q, items)
	return i + 1, err
}

func (m *Model) stationItems() []named {
	var items []named
	for i, r := range m.cat.Radios {
		items = append(items, named{r.Name, m.radioFold[i]})
	}
	return items
}

// startFor plays surah sq of rec in its best recording; with no surah, what
// was last heard of rec, else the first surah.
func (m *Model) startFor(rec *api.ReciterEntry, sq string) (*Start, error) {
	if sq == "" {
		var last *store.Progress
		for _, p := range m.Store.Progress {
			if p.Reciter == rec.ID && (last == nil || p.Updated.After(last.Updated)) {
				last = p
			}
		}
		if last != nil {
			if _, ok := findMoshaf(rec, last.Moshaf); ok {
				return &Start{Entry: store.Entry{Track: last.Track}}, nil
			}
		}
		mo, ok := best(m.cat, rec, m.Store.Settings.Rewaya, 0, false)
		if ids := surahIDs(mo); ok && len(ids) > 0 {
			return &Start{Entry: store.Entry{Track: store.Track{Reciter: rec.ID, Moshaf: mo.ID, Surah: ids[0]}}}, nil
		}
		return nil, fmt.Errorf("%s has no recordings", rec.Name)
	}
	n, err := m.findSurah(sq)
	if err != nil {
		return nil, err
	}
	mo, ok := best(m.cat, rec, m.Store.Settings.Rewaya, n, false)
	if !ok {
		return nil, fmt.Errorf("%s has no recording of %s (%d)", rec.Name, m.cat.SurahName(n), n)
	}
	return &Start{Entry: store.Entry{Track: store.Track{Reciter: rec.ID, Moshaf: mo.ID, Surah: n}}}, nil
}

// begin plays the Start; a surah opens its qari's page.
func (m *Model) begin() tea.Cmd {
	s := m.Start
	s.begun = true
	if s.Recipe != nil {
		return m.startRecipe(*s.Recipe)
	}
	if rec := m.cat.Reciter(s.Entry.Reciter); s.Entry.Radio == 0 && rec != nil {
		m.openQari(rec, s.Entry.Moshaf, s.Entry.Surah)
	}
	return m.playEntry(s.Entry, false)
}

// headlessDone: in --no-ui mode, nothing plays any more.
func (m *Model) headlessDone() bool {
	return m.Headless && m.Start != nil && m.Start.begun && m.kind == playNone
}

// statusLine is the whole view in --no-ui mode: title · elapsed/total.
func (m Model) statusLine() string {
	var parts []string
	switch m.kind {
	case playSurah:
		parts = append(parts, fmt.Sprintf("%s (%d) · %s", m.cat.SurahName(m.track.Surah), m.track.Surah, m.cat.Reciter(m.track.Reciter).Name))
	case playRadio:
		parts = append(parts, m.radio.Name)
	}
	if m.kind != playNone {
		t := clock(m.pos)
		if m.dur > 0 && m.kind == playSurah {
			t += "/" + clock(m.dur)
		}
		switch {
		case m.buffering:
			t += " loading"
		case m.paused:
			t += " paused"
		}
		parts = append(parts, t)
	}
	if m.recipe != nil {
		parts = append(parts, "random: "+m.recipe.Name)
	}
	if m.toast != "" {
		parts = append(parts, m.toast)
	}
	return strings.Join(parts, " · ")
}

// Cand is a Tab-completion candidate.
type Cand struct{ Value, Desc, Search string }

// Names lists completion candidates for what: "qari", "surah", "station"
// or "recipe". For "surah", qari narrows the list to that qari's surahs.
func Names(o Options, what, qari string) []Cand {
	m := New(o)
	var out []Cand
	switch what {
	case "qari":
		recs, _ := m.qariItems()
		for _, r := range recs {
			mo, _ := best(m.cat, r, m.Store.Settings.Rewaya, 0, false)
			out = append(out, Cand{r.Name, label(m.cat, mo, true), r.Search})
		}
	case "surah":
		ids := span(1, 114)
		if rec, err := m.findQari(qari); err == nil {
			var set []int
			for _, mo := range rec.Moshaf {
				set = append(set, surahIDs(mo)...)
			}
			ids = slices.DeleteFunc(ids, func(n int) bool { return !slices.Contains(set, n) })
		}
		for _, n := range ids {
			s := m.cat.Surahs[n-1]
			out = append(out, Cand{strconv.Itoa(n), s.Name, s.Search})
		}
	case "station":
		for i, r := range m.cat.Radios {
			out = append(out, Cand{r.Name, "radio", m.radioFold[i]})
		}
	case "recipe":
		for _, r := range m.recipes() {
			out = append(out, Cand{r.Name, m.recipeDesc(r), norm.Fold(r.Name)})
		}
	}
	return out
}
