package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/norm"
	"github.com/3mro777o/qari/internal/store"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) rowsFor(s *screen) []row {
	switch s.kind {
	case scrHome:
		return m.homeRows()
	case scrReciters:
		return m.reciterRows(s.surah)
	case scrMoshafs:
		return m.moshafRows(s.reciter)
	case scrSurahs:
		return m.surahRows(s.reciter, s.moshaf)
	case scrAllSurahs:
		return m.allSurahRows()
	case scrRadio:
		return m.radioRows()
	}
	return nil
}

// shown says whether a reciter passes the default filter: a complete
// recording by a well-known reciter. Favourites always pass.
func (m *Model) shown(r *api.ReciterEntry) bool {
	return m.opt.Store.Settings.ShowAll || m.opt.Store.Favorites[r.ID] || (r.Complete && (r.Known || r.Listens > 0))
}

func (m *Model) hasRewaya(r *api.ReciterEntry, surah int) bool {
	rw := m.opt.Store.Settings.Rewaya
	for _, ms := range r.Moshaf {
		if (rw == 0 || ms.RewayaID == rw) && (surah == 0 || ms.Surahs()[surah]) {
			return true
		}
	}
	return false
}

// filteredReciters applies the filters and the chosen sort; hidden counts
// reciters removed only by the "well-known, complete" filter.
func (m *Model) filteredReciters(surah int) (list []*api.ReciterEntry, hidden int) {
	for _, r := range m.cat.Reciters {
		if !m.hasRewaya(r, surah) {
			continue
		}
		if !m.shown(r) {
			hidden++
			continue
		}
		list = append(list, r)
	}
	plays := m.opt.Store.Plays
	switch m.opt.Store.Settings.Sort {
	case "a-z":
		sort.SliceStable(list, func(i, j int) bool { return norm.Fold(list[i].Name) < norm.Fold(list[j].Name) })
	case "recent":
		sort.SliceStable(list, func(i, j int) bool { return list[i].Date > list[j].Date })
	default:
		sort.SliceStable(list, func(i, j int) bool {
			a, b := list[i], list[j]
			if a.Listens != b.Listens {
				return a.Listens > b.Listens
			}
			if a.Known != b.Known {
				return a.Known
			}
			if plays[a.ID] != plays[b.ID] {
				return plays[a.ID] > plays[b.ID]
			}
			return norm.Fold(a.Name) < norm.Fold(b.Name)
		})
	}
	return list, hidden
}

func (m *Model) reciterRow(r *api.ReciterEntry, surah int) row {
	// The Arabic name is searchable (r.Search) but not shown, to keep rows calm.
	rw := row{title: r.Name, search: r.Search, data: r, metaSt: sDim}
	if m.opt.Store.Favorites[r.ID] {
		rw.mark, rw.markSt = "★", sGold
	}
	if m.now != nil && m.now.reciter == r {
		rw.mark, rw.markSt = "♪", sAccent
	}
	var meta []string
	if r.Listens > 0 {
		meta = append(meta, compact(r.Listens)+" plays")
	}
	if len(r.Moshaf) > 1 {
		meta = append(meta, plural(len(r.Moshaf), "recitation"))
	}
	if surah == 0 {
		total := 0
		for _, ms := range r.Moshaf {
			total = max(total, ms.SurahTotal)
		}
		if total >= 114 {
			meta = append(meta, "complete")
		} else {
			meta = append(meta, fmt.Sprintf("%d/114", total))
		}
	} else if ms := m.bestMoshaf(r, surah); ms != nil {
		rw.sub = m.rewayaShort(ms)
		if p := m.opt.Store.Get(store.Track{Reciter: r.ID, Moshaf: ms.ID, Surah: surah}); p != nil {
			meta = append(meta, progressLabel(p))
		}
	}
	rw.meta = strings.Join(meta, " · ")
	return rw
}

func (m *Model) reciterRows(surah int) []row {
	list, _ := m.filteredReciters(surah)
	rows := make([]row, 0, len(list))
	for _, r := range list {
		rows = append(rows, m.reciterRow(r, surah))
	}
	return rows
}

// rewayaShort turns "Rewayat Hafs A'n Assem - Murattal" into "Hafs · Murattal".
func (m *Model) rewayaShort(ms *api.Moshaf) string {
	name := ms.Name
	riw, style, ok := strings.Cut(name, " - ")
	if !ok {
		return name
	}
	riw = strings.TrimPrefix(riw, "Rewayat ")
	if i := strings.Index(riw, " A'n"); i > 0 {
		riw = riw[:i]
	}
	style = strings.NewReplacer("Almusshaf Al Mojawwad", "Mujawwad", "Almusshaf Al Mo'lim", "Teaching").Replace(style)
	return riw + " · " + style
}

func (m *Model) moshafRows(r *api.ReciterEntry) []row {
	var rows []row
	for _, ms := range m.moshafsFor(r) {
		rw := row{title: m.rewayaShort(ms), search: norm.Fold(ms.Name), data: moshafChoice{r, ms}, metaSt: sDim}
		if ms.SurahTotal >= 114 {
			rw.meta = "complete"
		} else {
			rw.meta = fmt.Sprintf("%d/114 surahs", ms.SurahTotal)
		}
		if m.now != nil && m.now.moshaf == ms {
			rw.mark, rw.markSt = "♪", sAccent
		}
		rows = append(rows, rw)
	}
	return rows
}

func progressLabel(p *store.Progress) string {
	if p.Finished {
		return "listened"
	}
	if p.Duration > 0 {
		return fmt.Sprintf("%s / %s", clock(p.Position), clock(p.Duration))
	}
	return clock(p.Position)
}

func progressMark(p *store.Progress) (string, lipgloss.Style) {
	switch {
	case p == nil:
		return "", sDim
	case p.Finished:
		return "✓", sAccent
	case p.Position < 5:
		return "", sDim
	}
	return "◔", sGold // one glyph: ◑ and ◕ render double-width in some fonts
}

func (m *Model) surahLabel(n int) (title, arabic, search string) {
	e := m.cat.Surahs[n-1]
	return fmt.Sprintf("%3d  %s", n, e.Name), e.Arabic, e.Search + " " + fmt.Sprint(n)
}

func (m *Model) surahRows(r *api.ReciterEntry, ms *api.Moshaf) []row {
	have := ms.Surahs()
	rows := []row{}
	for n := 1; n <= 114; n++ {
		if !have[n] {
			continue
		}
		title, ar, search := m.surahLabel(n)
		rw := row{title: title, arabic: ar, search: search, data: surahChoice{r, ms, n}, metaSt: sDim}
		p := m.opt.Store.Get(store.Track{Reciter: r.ID, Moshaf: ms.ID, Surah: n})
		rw.mark, rw.markSt = progressMark(p)
		if p != nil && (p.Finished || p.Position >= 5) {
			rw.meta = progressLabel(p)
		}
		if m.now != nil && m.now.moshaf == ms && m.now.track.Surah == n {
			rw.mark, rw.markSt = "♪", sAccent
		}
		rows = append(rows, rw)
	}
	return rows
}

func (m *Model) allSurahRows() []row {
	rows := make([]row, 0, 114)
	for n := 1; n <= 114; n++ {
		title, ar, search := m.surahLabel(n)
		e := m.cat.Surahs[n-1]
		place := "Medinan"
		if e.Makkia == 1 {
			place = "Meccan"
		}
		pages := e.EndPage - e.StartPage + 1
		rows = append(rows, row{title: title, sub: place, arabic: ar, search: search, data: allSurah(n),
			meta: plural(pages, "page"), metaSt: sDim})
	}
	return rows
}

func (m *Model) radioRows() []row {
	rows := make([]row, 0, len(m.cat.Radios))
	for i := range m.cat.Radios {
		r := &m.cat.Radios[i]
		rw := row{title: strings.TrimSpace(r.Name), search: norm.Fold(r.Name), data: r}
		if m.now != nil && m.now.radio == r {
			rw.mark, rw.markSt = "♪", sAccent
		}
		rows = append(rows, rw)
	}
	return rows
}

func (m *Model) homeRows() []row {
	var rows []row
	if recent := m.opt.Store.Recent(8); len(recent) > 0 {
		rows = append(rows, row{header: true, title: "Continue listening"})
		for _, p := range recent {
			r := m.cat.Reciter(p.Reciter)
			if r == nil || p.Surah < 1 || p.Surah > 114 {
				continue
			}
			mark, st := progressMark(&p)
			title, ar, search := m.surahLabel(p.Surah)
			sub := r.Name
			if len(r.Moshaf) > 1 {
				for i := range r.Moshaf {
					if r.Moshaf[i].ID == p.Moshaf {
						sub += " · " + m.rewayaShort(&r.Moshaf[i])
					}
				}
			}
			rows = append(rows, row{mark: mark, markSt: st, title: strings.TrimSpace(title), sub: sub,
				meta: progressLabel(&p), metaSt: sDim, arabic: ar, search: search + " " + r.Search, data: resumeRow(p.Track)})
		}
	}
	var favs []*api.ReciterEntry
	for _, r := range m.cat.Reciters {
		if m.opt.Store.Favorites[r.ID] {
			favs = append(favs, r)
		}
	}
	if len(favs) > 0 {
		sort.Slice(favs, func(i, j int) bool { return norm.Fold(favs[i].Name) < norm.Fold(favs[j].Name) })
		rows = append(rows, row{header: true, title: "Favorite qaris"})
		for _, r := range favs {
			rows = append(rows, m.reciterRow(r, 0))
		}
	}
	rows = append(rows, row{header: true, title: "Most listened"})
	saved := m.opt.Store.Settings.Sort
	m.opt.Store.Settings.Sort = "popular"
	list, _ := m.filteredReciters(0)
	m.opt.Store.Settings.Sort = saved
	for i, r := range list {
		if i == 12 {
			break
		}
		rows = append(rows, m.reciterRow(r, 0))
	}
	return rows
}
