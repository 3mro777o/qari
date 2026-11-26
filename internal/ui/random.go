package ui

// Random play from recipes (saved filters), and the listening history.

import (
	"cmp"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/norm"
	"github.com/3mro777o/qari/internal/store"
	tea "github.com/charmbracelet/bubbletea"
)

func span(a, b int) []int {
	var out []int
	for i := a; i <= b; i++ {
		out = append(out, i)
	}
	return out
}

var (
	daily    = []int{18, 32, 36, 56, 67} // Al-Kahf, As-Sajdah, Ya-Sin, Al-Waqi'ah, Al-Mulk
	builtins = []store.Recipe{
		{Name: "Anything", Known: true, Riwayat: []int{1}, Styles: []int{1}},
		{Name: "Favorite qaris", Qaris: "fav"},
		{Name: "Short surahs", Known: true, Riwayat: []int{1}, Surahs: span(78, 114)},
		{Name: "Daily surahs", Known: true, Surahs: daily},
		{Name: "Mujawwad", Styles: []int{2}},
		{Name: "Random radio", Radio: true},
		{Name: "Favorite radios", Radio: true, Qaris: "fav"},
	}
	// surahSets are the quick choices in the surah field.
	surahSets = []struct {
		short, long string
		ids         []int
	}{
		{"Juz Amma", "Juz Amma (78-114)", span(78, 114)},
		{"Juz 29-30", "Juz 29-30 (67-114)", span(67, 114)},
		{"Daily surahs", "Daily: Kahf, Sajdah, Ya-Sin, Waqi'ah, Mulk", daily},
	}
	styleIDs = []int{1, 2, 3, 4}
)

func styleLabel(id int) string {
	return map[int]string{1: "Murattal", 2: "Mujawwad", 3: "Teaching", 4: "Featured"}[id]
}

// recipes is the built-in recipes, then the user's.
func (m *Model) recipes() []store.Recipe { return append(slices.Clone(builtins), m.Store.Recipes...) }

func (m *Model) recipeName() string {
	if m.recipe != nil {
		return m.recipe.Name
	}
	return ""
}

func (m *Model) openRandom() {
	m.openPick(ovRandom, max(slices.IndexFunc(m.recipes(), func(r store.Recipe) bool { return r.Name == m.Store.Settings.Recipe }), 0))
}

func (m *Model) startRecipe(r store.Recipe) tea.Cmd {
	m.recipe, m.shufHist, m.shufAt, m.shufSeen = &r, nil, -1, map[string]bool{}
	m.Store.Settings.Recipe = r.Name
	return m.shuffleNext()
}

// shuffleNext replays the next entry after going back with p, else picks a
// new one from the recipe.
func (m *Model) shuffleNext() tea.Cmd {
	if m.shufAt+1 < len(m.shufHist) {
		m.shufAt++
		return m.playEntry(m.shufHist[m.shufAt], false)
	}
	r := *m.recipe
	pool := m.pool(r)
	if len(pool) == 0 {
		m.recipe = nil
		return m.flash(m.emptyMsg(r))
	}
	e := pickFrom(pool, m.shufSeen, rand.IntN)
	m.shufHist = append(m.shufHist, e)
	m.shufAt = len(m.shufHist) - 1
	return m.playEntry(e, true) // a random pick starts at the beginning
}

func (m *Model) shuffleBack() tea.Cmd {
	if m.shufAt <= 0 {
		return m.flash("Nothing earlier in this random session")
	}
	m.shufAt--
	return m.playEntry(m.shufHist[m.shufAt], false)
}

func (m *Model) playEntry(e store.Entry, fromStart bool) tea.Cmd {
	if e.Radio == 0 {
		return m.playSurah(e.Track, fromStart)
	}
	if r, ok := m.station(e.Radio); ok {
		return m.playRadio(r)
	}
	return m.flash("Station not found")
}

func (m *Model) station(id int) (api.Radio, bool) {
	i := slices.IndexFunc(m.cat.Radios, func(r api.Radio) bool { return r.ID == id })
	if i < 0 {
		return api.Radio{}, false
	}
	return m.cat.Radios[i], true
}

// pickFrom picks a random entry not seen since the pool was last exhausted.
func pickFrom(pool []store.Entry, seen map[string]bool, intn func(int) int) store.Entry {
	key := func(e store.Entry) string { return fmt.Sprint(e.Radio, "/", e.Reciter, "/", e.Surah) }
	fresh := slices.DeleteFunc(slices.Clone(pool), func(e store.Entry) bool { return seen[key(e)] })
	if len(fresh) == 0 {
		clear(seen)
		fresh = pool
	}
	e := fresh[intn(len(fresh))]
	seen[key(e)] = true
	return e
}

// pool is everything r can play: each matching (qari, surah) once, in the
// qari's preferred recording among those that pass the filters.
func (m *Model) pool(r store.Recipe) []store.Entry {
	var out []store.Entry
	if r.Radio {
		for _, st := range m.cat.Radios {
			if r.Qaris != "fav" || m.Store.FavRadios[st.ID] {
				out = append(out, store.Entry{Radio: st.ID})
			}
		}
		return out
	}
	ids := r.Surahs
	if len(ids) == 0 {
		ids = span(1, 114)
	}
	for _, rec := range m.cat.Reciters {
		if (r.Known && !wellKnown(rec)) || (r.Qaris == "fav" && !m.Store.Favorites[rec.ID]) || (r.Qaris == "set" && !slices.Contains(r.Reciters, rec.ID)) {
			continue
		}
		left := *rec
		left.Moshaf = slices.DeleteFunc(slices.Clone(rec.Moshaf), func(mo api.Moshaf) bool {
			return (len(r.Riwayat) > 0 && !slices.Contains(r.Riwayat, mo.RewayaID)) || (len(r.Styles) > 0 && !slices.Contains(r.Styles, mo.MoshafType%10))
		})
		var order []api.Moshaf
		var sets []map[int]bool
		for len(left.Moshaf) > 0 { // most preferred recording first
			mo, _ := best(m.cat, &left, 0, 0, false)
			order, sets = append(order, mo), append(sets, mo.Surahs())
			left.Moshaf = slices.DeleteFunc(left.Moshaf, func(x api.Moshaf) bool { return x.ID == mo.ID })
		}
		for _, s := range ids {
			if i := slices.IndexFunc(sets, func(set map[int]bool) bool { return set[s] }); i >= 0 {
				out = append(out, store.Entry{Track: store.Track{Reciter: rec.ID, Moshaf: order[i].ID, Surah: s}})
			}
		}
	}
	return out
}

func (m *Model) emptyMsg(r store.Recipe) string {
	switch {
	case r.Radio && r.Qaris == "fav":
		return "No favorite stations yet: press f on a station in 4 Radio"
	case !r.Radio && r.Qaris == "fav" && len(m.Store.Favorites) == 0:
		return "No favorite qaris yet: press f on a qari"
	}
	return fmt.Sprintf("Nothing matches %q: press z, then e to change it", r.Name)
}

// names is "Hafs, Warsh", or "3 riwayat" for more than two, or all when empty.
func names(ids []int, all, many string, name func(int) string) string {
	switch {
	case len(ids) == 0:
		return all
	case len(ids) > 2:
		return fmt.Sprint(len(ids), " ", many)
	}
	var out []string
	for _, id := range ids {
		out = append(out, name(id))
	}
	return strings.Join(out, ", ")
}

func (m *Model) riwayatDesc(r store.Recipe) string {
	return names(r.Riwayat, "all riwayat", "riwayat", func(id int) string { return riwayah(m.cat, api.Moshaf{RewayaID: id}, true) })
}

func stylesDesc(r store.Recipe) string { return names(r.Styles, "all styles", "styles", styleLabel) }

func (m *Model) surahsDesc(r store.Recipe) string {
	for _, set := range surahSets {
		if slices.Equal(r.Surahs, set.ids) {
			return set.short
		}
	}
	return names(r.Surahs, "all surahs", "surahs", m.cat.SurahName)
}

func (m *Model) qarisDesc(r store.Recipe) string {
	switch {
	case r.Radio:
		return map[bool]string{false: "all stations", true: "favorite stations"}[r.Qaris == "fav"]
	case r.Qaris == "fav":
		return "favorite qaris"
	case r.Qaris == "set" && len(r.Reciters) > 1:
		return plural(len(r.Reciters), "qari")
	case r.Qaris == "set":
		return names(r.Reciters, "", "qaris", func(id int) string {
			if rec := m.cat.Reciter(id); rec != nil {
				return rec.Name
			}
			return fmt.Sprint("qari ", id)
		})
	}
	return "all qaris"
}

// recipeDesc is a short summary that leaves out the "all" parts.
func (m *Model) recipeDesc(r store.Recipe) string {
	q := m.qarisDesc(r)
	if r.Known && !r.Radio {
		q = map[bool]string{true: "well-known", false: q + " (well-known)"}[q == "all qaris"]
	}
	parts := []string{q}
	if !r.Radio {
		parts = append(parts, m.riwayatDesc(r), stylesDesc(r), m.surahsDesc(r))
	}
	parts = slices.DeleteFunc(parts, func(s string) bool { return strings.HasPrefix(s, "all ") && s != "all stations" })
	return cmp.Or(strings.Join(parts, " · "), "anything")
}

// Recipe editor fields.
const (
	fName = iota
	fKind
	fQaris
	fKnown
	fRiwayat
	fStyles
	fSurahs
	fSave
)

func (m *Model) editFields() []int {
	if m.draft.Radio {
		return []int{fName, fKind, fQaris, fSave}
	}
	return []int{fName, fKind, fQaris, fKnown, fRiwayat, fStyles, fSurahs, fSave}
}

// editRow is the label and value of an editor field.
func (m *Model) editRow(f int) (string, string) {
	d := m.draft
	switch f {
	case fName:
		return "Name", d.Name
	case fKind:
		return "Plays", map[bool]string{false: "Surahs", true: "Radio stations"}[d.Radio]
	case fQaris:
		return map[bool]string{false: "Qaris", true: "Stations"}[d.Radio], m.qarisDesc(d)
	case fKnown:
		return "Well-known", map[bool]string{false: "any qari", true: "well-known qaris only"}[d.Known]
	case fRiwayat:
		return "Riwayat", m.riwayatDesc(d)
	case fStyles:
		return "Styles", stylesDesc(d)
	case fSurahs:
		return "Surahs", m.surahsDesc(d)
	}
	return "", "Save recipe"
}

type fieldItem struct {
	label, right string
	on           bool
	set          func(r *store.Recipe)
}

func toggle(s []int, v int) []int {
	if i := slices.Index(s, v); i >= 0 {
		return slices.Delete(s, i, i+1)
	}
	s = append(s, v)
	slices.Sort(s)
	return s
}

// fieldItems are the rows of the open field's multi-select; the "all" and
// quick choices hide while searching.
func (m *Model) fieldItems() []fieldItem {
	d, q := &m.draft, m.fieldQ
	var out []fieldItem
	add := func(label, right string, on bool, set func(*store.Recipe)) {
		out = append(out, fieldItem{label, right, on, set})
	}
	quick := q == ""
	switch m.field {
	case fQaris:
		if d.Radio {
			add("All stations", plural(len(m.cat.Radios), "station"), d.Qaris != "fav", func(r *store.Recipe) { r.Qaris = "" })
			add("Favorite stations", plural(len(m.Store.FavRadios), "favorite"), d.Qaris == "fav", func(r *store.Recipe) { r.Qaris = "fav" })
			break
		}
		if quick {
			add("All qaris", "", d.Qaris == "", func(r *store.Recipe) { r.Qaris, r.Reciters = "", nil })
			add("Favorite qaris", plural(len(m.Store.Favorites), "favorite"), d.Qaris == "fav", func(r *store.Recipe) { r.Qaris, r.Reciters = "fav", nil })
		}
		recs := slices.Clone(m.cat.Reciters)
		slices.SortStableFunc(recs, func(a, b *api.ReciterEntry) int { return cmp.Or(b.Listens-a.Listens, cmp.Compare(a.Name, b.Name)) })
		for _, rec := range recs {
			if !norm.Match(rec.Search, q) {
				continue
			}
			id, right := rec.ID, count(rec.Listens)
			if m.Store.Favorites[id] {
				right = strings.TrimSpace("★ " + right)
			}
			add(rec.Name, right, d.Qaris == "set" && slices.Contains(d.Reciters, id), func(r *store.Recipe) {
				if r.Qaris != "set" {
					r.Qaris, r.Reciters = "set", nil
				}
				if r.Reciters = toggle(r.Reciters, id); len(r.Reciters) == 0 {
					r.Qaris = ""
				}
			})
		}
	case fRiwayat:
		if quick {
			add("All riwayat", "", len(d.Riwayat) == 0, func(r *store.Recipe) { r.Riwayat = nil })
		}
		for _, o := range m.riwayat[1:] {
			if id := o.id; norm.Match(norm.Fold(o.name), q) {
				add(o.name, plural(o.count, "qari"), slices.Contains(d.Riwayat, id), func(r *store.Recipe) { r.Riwayat = toggle(r.Riwayat, id) })
			}
		}
	case fStyles:
		if quick {
			add("All styles", "", len(d.Styles) == 0, func(r *store.Recipe) { r.Styles = nil })
		}
		for _, id := range styleIDs {
			if norm.Match(norm.Fold(styleLabel(id)), q) {
				add(styleLabel(id), "", slices.Contains(d.Styles, id), func(r *store.Recipe) { r.Styles = toggle(r.Styles, id) })
			}
		}
	case fSurahs:
		if quick {
			add("All surahs", "", len(d.Surahs) == 0, func(r *store.Recipe) { r.Surahs = nil })
			for _, set := range surahSets {
				add(set.long, plural(len(set.ids), "surah"), slices.Equal(d.Surahs, set.ids), func(r *store.Recipe) { r.Surahs = slices.Clone(set.ids) })
			}
		}
		for _, su := range m.surahs(nil, q) {
			id := su.ID
			add(fmt.Sprintf("%03d  %s", id, su.Name), revelation(su), slices.Contains(d.Surahs, id), func(r *store.Recipe) { r.Surahs = toggle(r.Surahs, id) })
		}
	}
	return out
}

// edit opens the editor on a copy of r; idx is its place in Store.Recipes,
// -1 for a new recipe.
func (m *Model) edit(r store.Recipe, idx int) {
	r.Reciters, r.Riwayat, r.Styles, r.Surahs = slices.Clone(r.Reciters), slices.Clone(r.Riwayat), slices.Clone(r.Styles), slices.Clone(r.Surahs)
	m.draft, m.draftIdx = r, idx
	m.openPick(ovEdit, 0)
}

// recipeKey handles the keys of the recipe popups that differ from a plain picker.
func (m *Model) recipeKey(k string) (tea.Cmd, bool) {
	switch m.ov {
	case ovRandom:
		rs, i := m.recipes(), m.pick.cursor
		u := i - len(builtins) // index among the user's recipes
		cp := rs[i]
		cp.Name += " copy"
		switch k {
		case "z":
			m.ov = ovNone
		case "a":
			m.edit(store.Recipe{Name: "My recipe", Known: true}, -1)
		case "e":
			if u < 0 {
				m.edit(cp, -1)
				return m.flash("Built-in recipes are fixed: editing a copy"), true
			}
			m.edit(rs[i], u)
		case "c":
			m.edit(cp, -1)
		case "x":
			if u < 0 {
				return m.flash("Built-in recipes cannot be deleted"), true
			}
			if ok, cmd := m.twice(fmt.Sprint("xr", u), "Press x again to delete "+rs[i].Name); !ok {
				return cmd, true
			}
			m.Store.Recipes = slices.Delete(m.Store.Recipes, u, u+1)
			m.save()
			m.pick.clamp(m.pickN(), m.pickRows())
			return m.flash("Deleted " + rs[i].Name), true
		default:
			return nil, false
		}
		return nil, true
	case ovEdit:
		if k == "s" {
			return m.saveDraft(), true
		}
	case ovField:
		switch {
		case k == "/":
			m.startSearch()
			return nil, true
		case k == " ":
			return m.recipeChoose(), true
		case k == "esc" && m.fieldQ != "":
			m.fieldQ, m.pick = "", listState{}
			return nil, true
		}
	}
	return nil, false
}

func (m *Model) recipeChoose() tea.Cmd {
	i := m.pick.cursor
	switch m.ov {
	case ovRandom:
		m.ov = ovNone
		return m.startRecipe(m.recipes()[i])
	case ovEdit:
		switch f := m.editFields()[i]; f {
		case fName:
			m.startSearch()
		case fKind:
			if m.draft.Radio = !m.draft.Radio; m.draft.Radio && m.draft.Qaris == "set" {
				m.draft.Qaris = ""
			}
			m.pick.clamp(m.pickN(), m.pickRows())
		case fKnown:
			m.draft.Known = !m.draft.Known
		case fSave:
			return m.saveDraft()
		default:
			m.editCur, m.field, m.fieldQ = i, f, ""
			m.openPick(ovField, 0)
		}
	case ovField:
		if items := m.fieldItems(); i < len(items) {
			items[i].set(&m.draft)
		}
	}
	return nil
}

func (m *Model) saveDraft() tea.Cmd {
	d := m.draft
	if d.Name = strings.TrimSpace(d.Name); d.Name == "" {
		return m.flash("Give the recipe a name first")
	}
	self := -1
	if m.draftIdx >= 0 {
		self = len(builtins) + m.draftIdx
	}
	rs, base := m.recipes(), d.Name
	taken := func(name string) bool {
		for i, r := range rs {
			if i != self && r.Name == name {
				return true
			}
		}
		return false
	}
	for n := 2; taken(d.Name); n++ {
		d.Name = fmt.Sprintf("%s %d", base, n)
	}
	if m.draftIdx >= 0 {
		m.Store.Recipes[m.draftIdx] = d
	} else {
		m.Store.Recipes = append(m.Store.Recipes, d)
		m.draftIdx = len(m.Store.Recipes) - 1
	}
	m.Store.Settings.Recipe = d.Name // z then enter plays it
	m.save()
	m.openPick(ovRandom, len(builtins)+m.draftIdx)
	if n := len(m.pool(d)); n > 0 {
		return m.flash(fmt.Sprintf("Saved %s · %d to choose from", d.Name, n))
	}
	return m.flash("Saved " + d.Name + ", but nothing matches it yet")
}

// histItem is a History row: an entry and its index in Store.History.
type histItem struct {
	store.Entry
	idx int
}

// histItems is the history, newest first, filtered by the search. Entries
// the catalog no longer has are left out.
func (m *Model) histItems() []histItem {
	q := m.lists[pHistory].query
	var out []histItem
	for i := len(m.Store.History) - 1; i >= 0; i-- {
		e := m.Store.History[i]
		var hay []string
		if e.Radio > 0 {
			r, ok := m.station(e.Radio)
			if !ok {
				continue
			}
			hay = append(hay, r.Name)
		} else {
			rec := m.cat.Reciter(e.Reciter)
			if rec == nil || e.Surah < 1 || e.Surah > len(m.cat.Surahs) {
				continue
			}
			hay = append(hay, rec.Search, m.cat.Surahs[e.Surah-1].Search, fmt.Sprint(e.Surah))
		}
		if norm.Match(norm.Fold(strings.Join(append(hay, e.Recipe), " ")), q) {
			out = append(out, histItem{e, i})
		}
	}
	return out
}

// dayName is "Today", "Yesterday" or the date of t.
func dayName(t, now time.Time) string {
	t, now = t.Local(), now.Local()
	switch t.Format("2006-01-02") {
	case now.Format("2006-01-02"):
		return "Today"
	case now.AddDate(0, 0, -1).Format("2006-01-02"):
		return "Yesterday"
	}
	return t.Format("Mon 2 Jan 2006")
}
