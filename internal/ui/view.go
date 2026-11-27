package ui

import (
	"fmt"
	"github.com/charmbracelet/x/ansi"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/download"
	"github.com/3mro777o/qari/internal/store"
	"github.com/charmbracelet/lipgloss"
)

// zone is a clickable rectangle recorded while painting.
type zone struct{ kind, i, x0, y0, x1, y1 int }

const (
	zTab = iota + 1
	zRow
	zPick
	zSeek
	zPlay
	zBox
)

// painter renders one frame of a copy of the model and records click zones.
type painter struct {
	*Model
	s     styles
	zones []zone
}

func (m Model) View() string {
	if m.Headless {
		return m.statusLine()
	}
	v, _ := m.paint()
	return v
}

func (m Model) paint() (string, []zone) {
	p := &painter{Model: &m, s: m.th.s}
	if m.loading || m.loadErr != nil {
		return fitFrame(p.splash(), m.w, m.h), nil
	}
	lines := append(p.header(), p.canvas(2, m.h-2-m.dockH()-m.gap())...)
	lines = append(lines, make([]string, m.gap())...)
	lines = append(lines, p.dock(m.h-m.dockH())...)
	if m.ov != ovNone {
		lines = p.overlay(lines)
	}
	return fitFrame(lines, m.w, m.h), p.zones
}

func (p *painter) zone(kind, i, x0, y0, x1, y1 int) {
	p.zones = append(p.zones, zone{kind, i, x0, y0, x1, y1})
}

func (p *painter) kv(k, v string) string { return p.s.Key.Render(k) + " " + p.s.Muted.Render(v) }

// join joins the non-empty parts with a muted " · ".
func (p *painter) join(parts ...string) string {
	return strings.Join(slices.DeleteFunc(parts, func(s string) bool { return width(s) == 0 }), p.s.Muted.Render(" · "))
}

// fitJoin joins as many leading parts as fit in w cells.
func (p *painter) fitJoin(w int, parts ...string) string {
	for n := len(parts); n > 0; n-- {
		if s := p.join(parts[:n]...); width(s) <= w {
			return s
		}
	}
	return ""
}

// row lays out a list row: selection rail, left part, right part flush right.
func (p *painter) row(sel bool, left, right string, w int) string {
	rail := "   "
	if sel {
		rail = p.s.AccentB.Render("▌› ")
	}
	if right != "" {
		right += " "
	}
	return lr(rail+left, right, w)
}

func (p *painter) name(sel bool, s string) string {
	if sel {
		return p.s.Sel.Render(s)
	}
	return p.s.Text.Render(s)
}

func (p *painter) arabic(s string, w int, sel bool) string {
	if p.NoArabic || s == "" {
		return ""
	}
	if sel {
		return p.s.Accent.Render(isolate(s, w))
	}
	return p.s.Sub.Render(isolate(s, w))
}

func (p *painter) rule(w int) string { return p.s.Rule.Render(strings.Repeat("─", max(w, 0))) }

// status is the mark for a surah: playing, paused, finished, started or not.
func (p *painter) status(playing bool, frac float64, started bool) string {
	switch {
	case playing && p.paused:
		return p.s.Accent2.Render("‖")
	case playing:
		return p.s.AccentB.Render("▶")
	case frac >= 1:
		return p.s.Success.Render("✓")
	case started:
		return p.s.Accent2.Render("◔")
	}
	return p.s.Muted.Render("·")
}

func (p *painter) trackStatus(t store.Track) string {
	pr := p.Store.Get(t)
	return p.status(p.kind == playSurah && p.track == t, fraction(pr), pr != nil && pr.Position >= 5)
}

// surahCell is "001  Al-Fatihah    الفاتحة" with the Arabic in a straight column.
func (p *painter) surahCell(su api.SurahEntry, sel, hot bool, w int) string {
	num := p.s.Muted.Render(fmt.Sprintf("%03d", su.ID))
	if sel || hot {
		num = p.s.Accent.Render(fmt.Sprintf("%03d", su.ID))
	}
	ar := ""
	if w >= 54 {
		ar = p.arabic(su.Arabic, 20, sel)
	}
	if ar == "" {
		return num + "  " + p.name(sel, su.Name)
	}
	col := 17 - 3*b2i(w < 64)
	return num + "  " + fit(p.name(sel, trunc(su.Name, col)), col) + "  " + ar
}

func (p *painter) empty(h int, title, hint string) []string {
	out := make([]string, max((h-3)/2, 0), h)
	return append(out, center(p.s.Accent.Render("✦")+"  "+p.s.Bold.Render(title), p.w), "", center(p.s.Sub.Render(hint), p.w))
}

func (p *painter) splash() []string {
	s, bw := p.s, min(p.w-4, 68)
	b, top := s.Accent, s.AccentB.Render("✦ qari")
	body := []string{"", center(p.spin.View()+"  "+s.Bold.Render("Connecting to mp3quran.net"), bw-4),
		center(s.Sub.Render("Loading qaris, surahs, riwayat and radio stations"), bw-4), ""}
	foot := s.Muted.Render("cached for a day after the first load")
	if p.loadErr != nil {
		b, top, foot = s.Live, s.Live.Render("● Could not load the catalog"), p.kv("enter", "retry")
		body = []string{""}
		for _, l := range wrap(p.loadErr.Error(), bw-4, 3) {
			body = append(body, s.Text.Render(l))
		}
		body = append(body, "", s.Sub.Render("Some networks block mp3quran.net. If so, use a proxy:"),
			"  "+s.AccentB.Render("qari --proxy socks5://127.0.0.1:1080"), "")
	}
	out := make([]string, max((p.h-len(body)-2)/2, 0))
	out = append(out, center(border(bw, true, top, "", b), p.w))
	for _, l := range body {
		out = append(out, center(boxed(bw, l, b), p.w))
	}
	return append(out, center(border(bw, false, foot, p.kv("q", "quit"), b), p.w))
}

func (p *painter) header() []string {
	s, w := p.s, p.w
	gap, brand := 3, "✦ qari"
	if w < 112 {
		gap = 2
	}
	if w < 82 {
		gap, brand = 1, "✦"
	}
	line := " " + s.AccentB.Render(brand) + spaces(gap+1)
	if w < 74 { // room for all seven tabs
		line = " "
	}
	a0, a1 := 0, 0
	for i, name := range tabNames {
		seg := s.Muted.Render(fmt.Sprint(i+1)) + " " + s.Sub.Render(name)
		if i == p.tab {
			seg = s.Accent.Render(fmt.Sprint(i+1, " ", name))
			if p.page == i {
				seg = s.AccentB.Render(fmt.Sprint(i+1, " ", name))
			}
			a0, a1 = width(line), width(line)+width(seg)
		}
		p.zone(zTab, i, width(line), 0, width(line)+width(seg)-1, 0)
		line += seg + spaces(gap)
	}
	var right []string
	if p.Client != nil && p.Client.Offline() {
		right = append(right, s.Live.Render("OFFLINE"))
	}
	if done, total, frac, _ := p.Downloads.Progress(); total > 0 {
		right = append(right, s.Accent2.Render(fmt.Sprintf("↓ %d/%d %.0f%%", done, total, frac*100)))
	}
	if w >= 112 {
		right = append(right, s.Muted.Render(p.th.ID))
	}
	help := s.Key.Render("?")
	if w >= 76 {
		help = p.kv("?", "help")
	}
	l1 := lr(line, p.join(append(right, help)...)+" ", w)

	l := p.list()
	switch {
	case p.searching && p.ov == ovNone:
		hint := s.Sub.Render(fmt.Sprint(p.count(), " matches")) + s.Muted.Render(" · enter done · esc clear ")
		p.input.Width = max(8, w-width(hint)-6)
		return []string{l1, lr(" "+s.AccentB.Render("/")+" "+bidi(p.input.View()), hint, w)}
	case l.query != "":
		hint := s.Sub.Render(fmt.Sprint(p.count(), " matches")) + s.Muted.Render(" · / edit · esc clear ")
		return []string{l1, lr(" "+s.AccentB.Render("Filter:")+" "+s.Bold.Render(bidi(l.query)), hint, w)}
	}
	return []string{l1, p.rule(a0) + s.Accent.Render(strings.Repeat("━", a1-a0)) + p.rule(w-a1)}
}

func (p *painter) canvas(y, h int) []string {
	var out []string
	switch p.page {
	case pHome:
		out = p.home(y, h)
	case pDownloads:
		out = p.downloads(y, h)
	case pHistory:
		out = p.history(y, h)
	default:
		out = p.flat(y, h)
	}
	for len(out) < h {
		out = append(out, "")
	}
	return out[:h]
}

// page describes a list page: heading lines, rows, and an optional side
// panel shown on wide terminals.
type page struct {
	head  []string
	n     int
	row   func(i, w int, sel bool) string
	side  func(i, w, h int) []string
	split int // percent of the width for the list when the side panel shows
	empty [2]string
}

func (p *painter) flat(y, h int) []string {
	pg := p.spec()
	out := append(pg.head, p.rule(p.w))
	vis := h - len(out)
	l := p.list()
	l.clamp(pg.n, vis)
	if pg.n == 0 {
		return append(out, p.empty(vis, pg.empty[0], pg.empty[1])...)
	}
	w := p.w
	if pg.side != nil && p.w >= 100 {
		w = p.w * pg.split / 100
	}
	var rows []string
	for i := l.offset; i < min(pg.n, l.offset+vis); i++ {
		p.zone(zRow, i, 0, y+len(out)+len(rows), w-1, y+len(out)+len(rows))
		rows = append(rows, pg.row(i, w, i == l.cursor))
	}
	if w < p.w {
		rows = columns(rows, pg.side(l.cursor, p.w-w-1, vis), w, p.w-w-1, vis, p.s.Rule.Render("│"))
	}
	return append(out, rows...)
}

func (p *painter) spec() page {
	s, w, q := p.s, p.w, p.list().query
	noMatch := fmt.Sprintf("No matches for %q", q)
	switch p.page {
	case pQaris:
		items := p.qaris(q)
		return page{
			head:  []string{p.scopeRibbon(s.AccentB.Render(fmt.Sprintf("%s · %d", map[bool]string{false: "Well-known", true: "All qaris"}[p.Store.Settings.ShowAll], len(items))))},
			n:     len(items),
			row:   func(i, w int, sel bool) string { return p.qariRow(items[i], 0, w, sel) },
			side:  func(i, w, h int) []string { return p.qariSide(items[i], w, h) },
			split: 58, empty: [2]string{noMatch, p.scopeHint()},
		}
	case pSurahs:
		items, heard := p.surahs(nil, q), p.heard(func(*store.Progress) bool { return true })
		return page{
			head: []string{lr(" "+s.AccentB.Render(fmt.Sprintf("Surahs · %d", len(items))), p.kv("enter", "choose a qari")+" ", w)},
			n:    len(items),
			row: func(i, w int, sel bool) string {
				su := items[i]
				f, ok := heard[su.ID]
				hot := p.kind == playSurah && p.track.Surah == su.ID
				return p.row(sel, p.status(hot, f, ok)+" "+p.surahCell(su, sel, hot, w), p.join(s.Muted.Render(pages(su)), s.Muted.Render(revelation(su))), w)
			},
			side:  func(i, w, h int) []string { return p.surahSide(items[i], heard, w, h) },
			split: 58, empty: [2]string{noMatch, "Try an English or Arabic name, or a number 1-114"},
		}
	case pRadio:
		items := p.radios()
		right := p.kv("enter", "listen")
		if p.kind == playRadio {
			right = s.Live.Render("● LIVE")
		}
		return page{
			head:  []string{lr(" "+s.AccentB.Render(fmt.Sprintf("Live Quran Radio · %d stations", len(items))), right+" ", w)},
			n:     len(items),
			row:   func(i, w int, sel bool) string { return p.radioRow(items[i], w, sel) },
			side:  func(i, w, h int) []string { return p.radioSide(items[i], w) },
			split: 62, empty: [2]string{noMatch, "esc clears the search"},
		}
	case pQari:
		return p.qariPage(q)
	case pSettings:
		return p.settingsPage()
	}
	return p.surahPage(q)
}

func (p *painter) radioRow(r api.Radio, w int, sel bool) string {
	s := p.s
	icon, right := s.Muted.Render("♪"), ""
	if p.Store.FavRadios[r.ID] {
		icon = s.Fav.Render("★")
	}
	if p.kind == playRadio && p.radio == r {
		icon, right = s.Live.Render("●"), s.Live.Render("● LIVE")
		if p.paused {
			icon, right = s.Accent2.Render("‖"), s.Accent2.Render("paused")
		}
	}
	return p.row(sel, icon+" "+p.name(sel, r.Name), right, w)
}

// scopeRibbon is the qari filter line: scope, sort and riwayah, and the keys
// that change them. The a label says what pressing it will do.
func (p *painter) scopeRibbon(lead string) string {
	st, s := p.Store.Settings, p.s
	rw := ""
	if st.Rewaya != 0 {
		rw = s.Muted.Render("Riwayah: ") + s.Accent2.Render(riwayah(p.cat, api.Moshaf{RewayaID: st.Rewaya}, true))
	}
	left := p.join(" "+lead, s.Muted.Render("Sort: ")+s.Text.Render(sortNames[mode(sortModes, st.Sort)]), rw)
	act := map[bool]string{false: "show all", true: "well-known only"}[st.ShowAll]
	right := p.fitJoin(p.w-width(left)-2, s.Key.Render("a")+" "+s.Accent.Render(act), p.kv("s", "sort"), p.kv("w", "riwayah"), p.kv("f", "favorite"))
	return lr(left, right+" ", p.w)
}

func (p *painter) scopeHint() string {
	if !p.Store.Settings.ShowAll {
		return "a shows all qaris · w riwayah · esc clears the search"
	}
	return "w riwayah · esc clears the search"
}

// qariRow is a qari in a list; rank > 0 numbers it.
func (p *painter) qariRow(r *api.ReciterEntry, rank, w int, sel bool) string {
	s := p.s
	left := ""
	if rank > 0 {
		left = s.Muted.Render(fmt.Sprintf("%02d ", rank))
	}
	left += map[bool]string{true: s.Fav.Render("★ "), false: "  "}[p.Store.Favorites[r.ID]]
	if p.kind == playSurah && p.track.Reciter == r.ID {
		left += p.status(true, 0, false) + " "
	}
	var meta []string
	if mo, ok := best(p.cat, r, p.Store.Settings.Rewaya, 0, false); ok && w >= 54 {
		rw := riwayah(p.cat, mo, true)
		if len(r.Moshaf) > 1 {
			rw += fmt.Sprintf(" +%d", len(r.Moshaf)-1)
		}
		meta = append(meta, s.Muted.Render(rw))
		if !r.Complete {
			meta = append(meta, s.Warn.Render(fmt.Sprintf("%d/114", mo.SurahTotal)))
		}
	}
	if mode(sortModes, p.Store.Settings.Sort) == 2 && len(r.Date) >= 7 {
		meta = append(meta, s.Sub.Render(r.Date[:7]))
	} else {
		meta = append(meta, s.Sub.Render(count(r.Listens)))
	}
	return p.row(sel, left+p.name(sel, r.Name), p.join(meta...), w)
}

func (p *painter) qariSide(r *api.ReciterEntry, w, h int) []string {
	s := p.s
	fav := s.Muted.Render("f favorite")
	if p.Store.Favorites[r.ID] {
		fav = s.Fav.Render("★ Favorite")
	}
	done := map[bool]string{true: s.Success.Render("✓ Complete Quran"), false: s.Warn.Render("Partial recordings")}[r.Complete]
	out := []string{"  " + s.AccentB.Render(r.Name), "  " + p.join(fav, done, s.Sub.Render(count(r.Listens)+map[bool]string{true: " listens"}[r.Listens > 0]))}
	heard := p.heard(func(pr *store.Progress) bool { return pr.Reciter == r.ID })
	if fin := finished(heard); len(heard) > 0 {
		out = append(out, "", "  "+mini(float64(fin)/114, 12, s.Accent, s.Rule)+" "+p.join(s.Sub.Render(fmt.Sprintf("✓ %d/114 finished", fin)), s.Sub.Render(fmt.Sprintf("◔ %d started", len(heard)-fin))))
	}
	out = append(out, "", "  "+s.AccentB.Render(fmt.Sprintf("RECORDINGS (%d)", len(r.Moshaf))))
	pref, _ := best(p.cat, r, p.Store.Settings.Rewaya, 0, false)
	for i, mo := range r.Moshaf {
		if len(out) >= h-2 {
			out = append(out, "  "+s.Muted.Render(fmt.Sprintf(".. and %d more", len(r.Moshaf)-i)))
			break
		}
		mark, st := s.Muted.Render("·"), s.Text
		if mo.ID == pref.ID {
			mark, st = s.Accent.Render("✦"), s.Bold
		}
		out = append(out, lr("  "+mark+" "+st.Render(label(p.cat, mo, false)), p.saved(r.ID, mo)+s.Sub.Render(fmt.Sprintf("%d surahs ", mo.SurahTotal)), w))
	}
	for len(out) < h-1 {
		out = append(out, "")
	}
	return append(out, "  "+p.kv("enter", "open recordings & surahs"))
}

// saved is "↓ 12 · " when some surahs of mo are downloaded.
func (p *painter) saved(reciter int, mo api.Moshaf) string {
	n := 0
	for _, s := range surahIDs(mo) {
		n += b2i(p.Downloads.Has(store.Track{Reciter: reciter, Moshaf: mo.ID, Surah: s}))
	}
	if n == 0 {
		return ""
	}
	return p.s.Accent2.Render(fmt.Sprint("↓ ", n)) + p.s.Muted.Render(" · ")
}

func (p *painter) surahSide(su api.SurahEntry, heard map[int]float64, w, h int) []string {
	s := p.s
	out := []string{"  " + s.AccentB.Render(fmt.Sprintf("%03d · %s", su.ID, su.Name))}
	if ar := p.arabic(su.Arabic, w-4, true); ar != "" {
		out = append(out, "  "+ar)
	}
	meta := revelation(su)
	if n := su.EndPage - su.StartPage + 1; su.StartPage > 0 {
		meta += fmt.Sprintf(" · %s from page %d", plural(max(n, 1), "page"), su.StartPage)
	}
	out = append(out, "  "+s.Sub.Render(meta))
	switch f, ok := heard[su.ID]; {
	case f >= 1:
		out = append(out, "  "+s.Success.Render("✓ Finished"))
	case ok:
		out = append(out, "  "+mini(f, 10, s.Accent, s.Rule)+" "+s.Accent2.Render(fmt.Sprintf("◔ %.0f%% heard", f*100)))
	}
	for len(out) < h-1 {
		out = append(out, "")
	}
	return append(out, "  "+p.kv("enter", "choose a qari and play"))
}

func (p *painter) radioSide(r api.Radio, w int) []string {
	s := p.s
	var out []string
	if p.kind == playRadio {
		out = append(out, "  "+s.Live.Render("● NOW STREAMING"))
		for _, l := range wrap(p.radio.Name, w-4, 2) {
			out = append(out, "  "+s.Bold.Render(l))
		}
		out = append(out, "")
	}
	out = append(out, "  "+s.AccentB.Render("SELECTED STATION"))
	for _, l := range wrap(r.Name, w-4, 3) {
		out = append(out, "  "+s.Text.Render(l))
	}
	return append(out, "", "  "+s.Muted.Render("A continuous live broadcast."), "  "+s.Muted.Render("Radio does not change surah progress."),
		"", "  "+p.kv("enter", "tune in"), "  "+p.kv("space", "pause or resume"))
}

func (p *painter) qariPage(q string) page {
	s, w, r, mo := p.s, p.w, p.qari, p.moshaf
	items := p.surahs(surahIDs(mo), q)
	back := map[int]string{pHome: "Home", pSurah: fmt.Sprintf("Surah %03d", p.surah), pDownloads: "Downloads"}[p.prev]
	fav, favAct := "", "favorite"
	if p.Store.Favorites[r.ID] {
		fav, favAct = " "+s.Fav.Render("★"), "unfavorite"
	}
	l1 := lr(" "+s.Muted.Render("‹ "+cmpOr(back, "Qaris")+" › ")+s.AccentB.Render(r.Name)+fav, p.join(p.kv("f", favAct), p.kv("esc", "back"))+" ", w)
	rec := " " + s.Accent.Render("✦") + " " + s.Bold.Render(label(p.cat, mo, false)) + s.Muted.Render(fmt.Sprintf(" (%d surahs)", mo.SurahTotal))
	if n := len(r.Moshaf); n > 1 {
		i := slices.IndexFunc(r.Moshaf, func(x api.Moshaf) bool { return x.ID == mo.ID })
		rec += "  " + s.Muted.Render(fmt.Sprintf("recording %d of %d · ", i+1, n)) + p.kv("v", "other recordings")
	}
	fin := finished(p.heard(func(pr *store.Progress) bool { return pr.Reciter == r.ID && pr.Moshaf == mo.ID }))
	l2 := lr(rec, p.saved(r.ID, mo)+mini(float64(fin)/float64(max(mo.SurahTotal, 1)), 8, s.Success, s.Rule)+" "+s.Sub.Render(fmt.Sprintf("%d/%d ", fin, mo.SurahTotal)), w)
	jobs := map[store.Track]download.Job{}
	for _, j := range p.Downloads.Jobs() {
		jobs[j.Track] = j
	}
	return page{
		head: []string{l1, l2},
		n:    len(items),
		row: func(i, w int, sel bool) string {
			su := items[i]
			t := store.Track{Reciter: r.ID, Moshaf: mo.ID, Surah: su.ID}
			pr, hot := p.Store.Get(t), p.kind == playSurah && p.track == t
			var right string
			switch j, ok := jobs[t]; {
			case ok:
				right = p.jobState(j)
			case hot && p.dur > 0:
				right = mini(p.pos/p.dur, 8, s.Accent, s.Rule) + " " + s.AccentB.Render(clock(p.pos)+"/"+clock(p.dur))
			case pr != nil && !pr.Finished && pr.Position >= 5 && pr.Duration > 0:
				right = mini(pr.Fraction(), 8, s.Accent, s.Rule) + " " + s.Sub.Render(clock(pr.Position)+"/"+clock(pr.Duration))
			case pr != nil && pr.Finished:
				right = p.join(s.Success.Render("✓ heard"), s.Muted.Render(revelation(su)))
			default:
				right = p.join(s.Muted.Render(map[bool]string{true: pages(su)}[w >= 74]), s.Muted.Render(revelation(su)))
			}
			if p.Downloads.Has(t) {
				right = p.join(s.Accent2.Render("↓"), right)
			}
			return p.row(sel, p.trackStatus(t)+" "+p.surahCell(su, sel, hot, w), right, w)
		},
		empty: [2]string{fmt.Sprintf("No surahs match %q", q), "Search by name or number · esc clears"},
	}
}

// bidi isolates typed text that contains Arabic.
func bidi(s string) string {
	if strings.ContainsFunc(s, func(r rune) bool { return unicode.In(r, unicode.Arabic) }) {
		return rli + s + pdi
	}
	return s
}

func fraction(pr *store.Progress) float64 {
	if pr == nil {
		return 0
	}
	return pr.Fraction()
}

func cmpOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func (p *painter) surahPage(q string) page {
	s, w := p.s, p.w
	su := p.cat.Surahs[p.surah-1]
	items := p.qarisFor(p.surah, q)
	l1 := lr(" "+s.Muted.Render("‹ Surahs › ")+s.AccentB.Render(fmt.Sprintf("%03d · %s", su.ID, su.Name))+"  "+p.arabic(su.Arabic, 20, true),
		p.join(s.Muted.Render(pages(su)), s.Muted.Render(revelation(su)), p.kv("esc", "back"))+" ", w)
	return page{
		head: []string{l1, p.scopeRibbon(s.Bold.Render(plural(len(items), "qari")))},
		n:    len(items),
		row: func(i, w int, sel bool) string {
			r, mo := items[i].r, items[i].mo
			t := store.Track{Reciter: r.ID, Moshaf: mo.ID, Surah: p.surah}
			mark := p.trackStatus(t) + " "
			if p.Store.Favorites[r.ID] {
				mark += s.Fav.Render("★ ")
			}
			meta := []string{s.Sub.Render(label(p.cat, mo, true))}
			if p.Downloads.Has(t) {
				meta = append(meta, s.Accent2.Render("↓"))
			}
			if pr := p.Store.Get(t); pr != nil && !pr.Finished && pr.Position >= 5 {
				meta = append(meta, s.Accent2.Render(fmt.Sprintf("%.0f%%", pr.Fraction()*100)))
			} else if w >= 72 {
				meta = append(meta, s.Muted.Render(count(r.Listens)))
			}
			return p.row(sel, mark+p.name(sel, r.Name), p.join(meta...), w)
		},
		empty: [2]string{"No qaris for this surah here", p.scopeHint()},
	}
}

// srow is a row of a sectioned list; i is the item index, or -1 for headings.
type srow struct {
	text string
	i    int
}

// window shows h rows of a sectioned list, scrolled to keep the cursor in view.
func (p *painter) window(rows []srow, x, y, w, h int) []string {
	cur, start := p.list().cursor, 0
	if sel := slices.IndexFunc(rows, func(r srow) bool { return r.i == cur }); len(rows) > h && sel >= h-2 {
		start = min(len(rows)-h, sel-h/2)
	}
	var out []string
	for j := start; j < min(len(rows), start+h); j++ {
		if rows[j].i >= 0 {
			p.zone(zRow, rows[j].i, x, y+j-start, x+w-1, y+j-start)
		}
		out = append(out, rows[j].text)
	}
	return out
}

func (p *painter) home(y, h int) []string {
	s, items, cur := p.s, p.homeItems(), p.list().cursor
	if len(items) == 0 {
		return p.empty(h, fmt.Sprintf("No matches for %q", p.list().query), "esc clears the search · 2 Qaris · 3 Surahs")
	}
	wide := p.w >= 96
	lw, rw := p.w, 0
	if wide {
		lw = p.w * 54 / 100
		rw = p.w - lw - 1
	}
	var sec [3][]srow
	rank := 0
	for i, it := range items {
		sel := i == cur
		switch {
		case it.kind == 0 && wide && h >= 18:
			a, b := p.continueCard(it, lw, sel)
			sec[0] = append(sec[0], srow{a, i}, srow{b, i})
		case it.kind == 0:
			sec[0] = append(sec[0], srow{p.continueRow(it, lw, sel), i})
		case it.radio != nil:
			sec[1] = append(sec[1], srow{p.radioRow(*it.radio, lw, sel), i})
		case it.kind == 1:
			sec[1] = append(sec[1], srow{p.qariRow(it.rec, 0, lw, sel), i})
		default:
			rank++
			sec[2] = append(sec[2], srow{p.qariRow(it.rec, rank, max(rw, lw*b2i(!wide)), sel), i})
		}
	}
	act := map[bool]string{false: "show all", true: "well-known only"}[p.Store.Settings.ShowAll]
	heads := [3][2]string{
		{s.AccentB.Render("✦ CONTINUE LISTENING"), s.Muted.Render(map[bool]string{true: "enter resumes"}[len(sec[0]) > 0])},
		{s.Fav.Render("★") + " " + s.AccentB.Render("FAVORITES"), s.Muted.Render(plural(len(sec[1]), "favorite"))},
		{s.AccentB.Render("✦ MOST LISTENED"), s.Muted.Render("a " + act)},
	}
	if wide {
		if len(sec[0]) == 0 {
			sec[0] = []srow{{"   " + s.Bold.Render("Welcome to qari"), -1},
				{"   " + s.Sub.Render("Pick a qari on the right, or browse all 114 surahs."), -1},
				{"   " + s.Muted.Render("Your place in every surah is saved automatically."), -1}}
		}
		if len(sec[1]) == 0 {
			sec[1] = []srow{{"   " + s.Muted.Render("Press f on a qari or radio to pin it here"), -1}}
		}
	}
	var left, right []srow
	for k := range sec {
		dst, w := &left, lw
		if wide && k == 2 {
			dst, w = &right, rw
		}
		if len(sec[k]) > 0 {
			if len(*dst) > 0 {
				*dst = append(*dst, srow{"", -1})
			}
			*dst = append(append(*dst, srow{lr(" "+heads[k][0], heads[k][1]+" ", w), -1}), sec[k]...)
		}
	}
	if !wide {
		return p.window(left, 0, y, p.w, h)
	}
	if len(right)+2 <= h {
		fin := 0
		for _, pr := range p.Store.Progress {
			fin += b2i(pr.Finished)
		}
		right = append(right, srow{"", -1}, srow{" " + p.join(s.Muted.Render(plural(len(p.cat.Reciters), "qari")), s.Muted.Render(plural(len(p.Store.Favorites)+len(p.Store.FavRadios), "favorite")), s.Muted.Render(plural(fin, "surah")+" finished")), -1})
	}
	return columns(p.window(left, 0, y, lw, h), p.window(right, lw+1, y, rw, h), lw, rw, h, s.Rule.Render("│"))
}

func (p *painter) continueIcon(pr *store.Progress) string {
	return p.status(p.kind == playSurah && p.track == pr.Track, pr.Fraction(), true)
}

func (p *painter) continueRow(it homeItem, w int, sel bool) string {
	pr := it.prog
	right := ""
	if pr.Duration > 0 {
		right = p.s.Sub.Render(clock(pr.Position) + "/" + clock(pr.Duration))
	}
	return p.row(sel, p.continueIcon(pr)+" "+p.s.Muted.Render(fmt.Sprintf("%03d ", pr.Surah))+p.name(sel, p.cat.SurahName(pr.Surah))+
		p.s.Muted.Render(" · ")+p.s.Sub.Render(it.rec.Name), right, w)
}

// continueCard is a two-line Continue Listening entry with a progress bar.
func (p *painter) continueCard(it homeItem, w int, sel bool) (string, string) {
	s, pr := p.s, it.prog
	ar := ""
	if w >= 46 {
		ar = "  " + p.arabic(p.cat.Surahs[pr.Surah-1].Arabic, 16, sel)
	}
	right := ""
	switch {
	case pr.Finished:
		right = s.Success.Render("finished")
	case pr.Duration > 0:
		right = s.Sub.Render(clock(pr.Duration-pr.Position) + " left")
	}
	l1 := p.row(sel, p.continueIcon(pr)+" "+s.Muted.Render(fmt.Sprintf("%03d", pr.Surah))+" "+p.name(sel, p.cat.SurahName(pr.Surah))+ar, right, w)
	mo, _ := findMoshaf(it.rec, pr.Moshaf)
	rail := "   "
	if sel {
		rail = s.AccentB.Render("▌  ")
	}
	right = ""
	if pr.Duration > 0 || pr.Finished {
		right = mini(pr.Fraction(), 8, s.Accent, s.Rule) + " " + s.Muted.Render(fmt.Sprintf("%3.0f%% ", pr.Fraction()*100))
	}
	return l1, lr(rail+"  "+s.Sub.Render(it.rec.Name)+s.Muted.Render(" · "+riwayah(p.cat, mo, true)), right, w)
}

func (p *painter) jobState(j download.Job) string {
	s := p.s
	switch j.State {
	case download.Active:
		f := 0.0
		if j.Size > 0 {
			f = float64(j.Done) / float64(j.Size)
		}
		return s.Accent2.Render("↓ ") + mini(f, 10, s.Accent, s.Rule) + " " + s.Sub.Render(fmt.Sprintf("%3.0f%% · %s/s", f*100, bytes(j.Speed)))
	case download.Failed:
		return s.Live.Render("failed: " + trunc(j.Err.Error(), 32))
	}
	return s.Muted.Render("queued")
}

func (p *painter) downloads(y, h int) []string {
	s, w, cur := p.s, p.w, p.list().cursor
	items := p.dlItems()
	var total float64
	var rows []srow
	for i, it := range items {
		sel := i == cur
		if j := it.job; j != nil {
			if i == 0 {
				done, all, frac, speed := p.Downloads.Progress()
				right := mini(frac, 14, s.Accent, s.Rule) + " " + s.Sub.Render(fmt.Sprintf("%.0f%%", frac*100))
				if speed > 0 {
					right += s.Muted.Render(" · ") + s.Sub.Render(bytes(speed)+"/s")
				}
				rows = append(rows, srow{lr(" "+s.AccentB.Render(fmt.Sprintf("↓ DOWNLOADING  %d of %d", min(done+1, all), all)), right+" ", w), -1})
			}
			icon := map[download.State]string{download.Active: s.AccentB.Render("↓"), download.Queued: s.Muted.Render("·"), download.Failed: s.Live.Render("!")}[j.State]
			name := ""
			if r := p.cat.Reciter(j.Reciter); r != nil {
				name = r.Name
			}
			rows = append(rows, srow{p.row(sel, icon+" "+s.Muted.Render(fmt.Sprintf("%03d ", j.Surah))+p.name(sel, p.cat.SurahName(j.Surah))+s.Muted.Render(" · "+name), p.jobState(*j), w), i})
			continue
		}
		g := it.group
		if i == 0 || items[i-1].job != nil {
			if i > 0 {
				rows = append(rows, srow{"", -1})
			}
			rows = append(rows, srow{" " + s.AccentB.Render("✦ SAVED FOR OFFLINE"), -1})
		}
		r := p.cat.Reciter(g.Reciter)
		mo, _ := findMoshaf(r, g.Moshaf)
		total += float64(g.Bytes)
		rows = append(rows, srow{p.row(sel, "  "+p.name(sel, r.Name)+s.Muted.Render(" · "+label(p.cat, mo, true)),
			p.join(s.Sub.Render(fmt.Sprintf("%d/%d surahs", g.Surahs, mo.SurahTotal)), s.Muted.Render(bytes(float64(g.Bytes)))), w), i})
	}
	if len(items) == 0 {
		rows = append(rows, srow{" " + s.AccentB.Render("✦ SAVED FOR OFFLINE"), -1}, srow{"", -1},
			srow{"   " + s.Text.Render("Nothing saved yet."), -1},
			srow{"   " + s.Sub.Render("Press d on a surah to save it, or on a qari to save a whole recording."), -1},
			srow{"   " + s.Muted.Render("Saved surahs play without a network connection."), -1})
	}
	head := lr(" "+s.AccentB.Render("Downloads")+s.Muted.Render(" · "+bytes(total)+" on disk"), p.fitJoin(w/2, p.kv("enter", "open"), p.kv("x", "delete or cancel"))+" ", w)
	return append([]string{head, p.rule(w)}, p.window(rows, 0, y+2, w, h-2)...)
}

func (p *painter) history(y, h int) []string {
	s, w, cur := p.s, p.w, p.list().cursor
	items := p.histItems()
	var rows []srow
	day, now := "", time.Now()
	for i, e := range items {
		if d := dayName(e.Time, now); d != day {
			if day != "" {
				rows = append(rows, srow{"", -1})
			}
			rows, day = append(rows, srow{" " + s.AccentB.Render("✦ "+strings.ToUpper(d)), -1}), d
		}
		rows = append(rows, srow{p.histRow(e.Entry, w, i == cur), i})
	}
	if len(items) == 0 {
		hint := []string{"Nothing played yet.", "Every surah and station you start shows up here, newest first.", "Press z to play something at random."}
		if q := p.list().query; q != "" {
			hint = []string{fmt.Sprintf("No matches for %q", q), "esc clears the search"}
		}
		rows = append(rows, srow{"", -1})
		for i, l := range hint {
			rows = append(rows, srow{"   " + []lipgloss.Style{s.Text, s.Sub, s.Muted}[i].Render(l), -1})
		}
	}
	head := lr(" "+s.AccentB.Render("History")+s.Muted.Render(fmt.Sprintf(" · %d", len(items))), p.fitJoin(w/2, p.kv("enter", "play again"), p.kv("x", "remove"))+" ", w)
	return append([]string{head, p.rule(w)}, p.window(rows, 0, y+2, w, h-2)...)
}

// histRow is "▶ 14:05 018 Al-Kahf · Qari" with where it was left off.
func (p *painter) histRow(e store.Entry, w int, sel bool) string {
	s := p.s
	at := s.Muted.Render(e.Time.Local().Format("15:04") + " ")
	var right []string
	if e.Recipe != "" && w >= 70 {
		right = append(right, s.Accent2.Render("⤮ "+trunc(bidi(e.Recipe), 18)))
	}
	if e.Radio > 0 {
		r, _ := p.station(e.Radio)
		icon := s.Muted.Render("♪")
		if p.kind == playRadio && p.radio.ID == r.ID {
			icon = s.Live.Render("●")
		}
		return p.row(sel, icon+" "+at+p.name(sel, bidi(r.Name)), p.join(append(right, s.Muted.Render("radio"))...), w)
	}
	pr := p.Store.Get(e.Track)
	switch {
	case p.kind == playSurah && p.track == e.Track && p.dur > 0:
		right = append(right, s.AccentB.Render(clock(p.pos)+"/"+clock(p.dur)))
	case pr != nil && pr.Finished:
		right = append(right, s.Success.Render("✓ heard"))
	case pr != nil && pr.Position >= 5 && pr.Duration > 0:
		right = append(right, mini(pr.Fraction(), 6, s.Accent, s.Rule)+" "+s.Sub.Render(clock(pr.Position)+"/"+clock(pr.Duration)))
	}
	return p.row(sel, p.trackStatus(e.Track)+" "+at+s.Muted.Render(fmt.Sprintf("%03d ", e.Surah))+p.name(sel, p.cat.SurahName(e.Surah))+
		s.Muted.Render(" · ")+s.Sub.Render(p.cat.Reciter(e.Reciter).Name), p.join(right...), w)
}

func (p *painter) dock(y int) []string {
	s, w, inner := p.s, p.w, p.w-4
	b := s.Rule
	var tl, tr, ml, mr, line string
	var tele []string
	if mins := sleeps[p.sleep]; mins != 0 {
		left := "end of surah"
		if mins > 0 {
			left = clock(p.sleepEnd.Sub(time.Now()).Seconds())
		}
		tele = append(tele, s.Warn.Render("sleep "+left))
	}
	if p.kind == playSurah && (p.speed != 1 || w >= 86) {
		tele = append(tele, s.Sub.Render(fmt.Sprintf("%gx", speeds[p.speed])))
	}
	if p.muted {
		tele = append(tele, s.Warn.Render("muted"))
	} else if p.Player != nil {
		tele = append(tele, s.Sub.Render(fmt.Sprintf("vol %.0f%%", p.vol)))
	}
	tr = p.join(tele...)
	frac := 0.0
	if p.dur > 0 {
		frac = p.pos / p.dur
	}
	switch {
	case p.Player == nil:
		tl, tr = s.Warn.Render("✦ mpv not installed"), s.Muted.Render("browse mode")
		ml, line = s.Sub.Render("Install mpv to play audio; browsing and search still work"), p.rule(inner)
	case p.kind == playSurah:
		b = map[bool]lipgloss.Style{false: s.Accent, true: s.Accent2}[p.paused]
		rec := p.cat.Reciter(p.track.Reciter)
		mo, _ := findMoshaf(rec, p.track.Moshaf)
		icon, note := s.AccentB.Render("▶"), ""
		if p.buffering {
			icon, note = p.spin.View(), s.Muted.Render(" loading..")
		} else if p.rep != nil && p.rep.gapping {
			icon, note = s.Accent2.Render("‖"), s.Muted.Render(" gap")
		} else if p.paused {
			icon, note = s.Accent2.Render("‖"), s.Muted.Render(" paused")
		}
		ar := ""
		if w >= 84 {
			ar = "  " + p.arabic(p.cat.Surahs[p.track.Surah-1].Arabic, 20, true)
		}
		tl = icon + " " + s.Bold.Render(fmt.Sprintf("%03d · %s", p.track.Surah, p.cat.SurahName(p.track.Surah))) + ar + note
		ml = s.Text.Render(rec.Name) + s.Muted.Render(" · ") + s.Sub.Render(label(p.cat, mo, true))
		t := clock(p.pos)
		if p.dur > 0 {
			t += " / " + clock(p.dur)
		}
		after := repeatNames[mode(repeatModes, p.Store.Settings.Repeat)]
		if p.recipe != nil {
			after = "⤮ " + trunc(bidi(p.recipe.Name), max(inner/4, 8))
		}
		switch {
		case p.rep != nil:
			after = p.rep.status()
		case p.abFrom > 0:
			after = fmt.Sprintf("from ayah %d · i marks the last", p.abFrom)
		}
		mr = s.Text.Render(t) + s.Muted.Render(" · ") + s.Accent2.Render(after)
		line = bar(frac, inner, s.Accent, s.AccentB, s.Rule)
	case p.kind == playRadio:
		b = map[bool]lipgloss.Style{false: s.Live, true: s.Accent2}[p.paused]
		icon, tag := s.Live.Render("● LIVE"), s.Live.Render(" ● LIVE STREAM ")
		if p.buffering {
			icon = p.spin.View() + " " + s.Live.Render("CONNECTING")
		} else if p.paused {
			icon, tag = s.Accent2.Render("‖ PAUSED"), s.Accent2.Render(" ‖ STREAM PAUSED ")
		}
		tl, ml = icon+" "+s.Bold.Render(p.radio.Name), s.Sub.Render("Live Quran radio")
		if p.recipe != nil {
			ml += s.Muted.Render(" · ") + s.Accent2.Render("⤮ "+bidi(p.recipe.Name)) + s.Muted.Render(" · ") + p.kv("n", "another")
		}
		side := (inner - width(tag)) / 2
		line = p.rule(side) + tag + p.rule(inner-side-width(tag))
	case p.Store.Last != nil && p.cat.Reciter(p.Store.Last.Reciter) != nil:
		t := *p.Store.Last
		tl = s.Muted.Render("✦ Last:") + " " + s.Text.Render(fmt.Sprintf("%03d · %s", t.Surah, p.cat.SurahName(t.Surah)))
		tr = p.join(p.kv("space", "resume"), tr)
		ml = s.Sub.Render(p.cat.Reciter(t.Reciter).Name)
		pr := p.Store.Get(t)
		switch {
		case pr != nil && pr.Finished:
			mr = s.Success.Render("✓ finished")
		case pr != nil && pr.Duration > 0:
			mr = s.Sub.Render(clock(pr.Position) + " / " + clock(pr.Duration))
		}
		line = bar(fraction(pr), inner, s.Sub, s.Sub, s.Rule)
	default:
		tl, ml, line = s.Muted.Render("✦ Ready to listen"), s.Sub.Render("Choose a qari, a surah or a radio station and press enter"), p.rule(inner)
	}
	if p.kind != playNone {
		p.zone(zPlay, 0, 3, y, 3+width(tl), y)
	}
	bl, br := s.Success.Render("✦ "+p.toast), ""
	if p.toast == "" {
		bl = p.footer(w - 14)
	}
	if w >= 78 || p.toast == "" {
		br = p.kv("?", "help")
	}
	top, bottom := border(w, true, tl, tr, b), border(w, false, bl, br, b)
	if p.dockH() == 3 {
		right := mr
		if p.kind == playSurah {
			mw := min(14, max(6, inner/5))
			right = mini(frac, mw, s.Accent, s.Rule) + " " + mr
			x := 2 + inner - width(right)
			p.zone(zSeek, 0, x, y+1, x+mw-1, y+1)
		}
		return []string{top, boxed(w, lr(ml, right, inner), b), bottom}
	}
	if p.kind == playSurah {
		p.zone(zSeek, 0, 2, y+2, 1+inner, y+2)
	}
	return []string{top, boxed(w, lr(ml, mr, inner), b), boxed(w, line, b), bottom}
}

// footer shows the few keys that matter here.
func (p *painter) footer(w int) string {
	var h []string
	if p.kind != playNone { // idle: the dock's top border already says "space resume"
		h = append(h, p.kv("space", map[bool]string{false: "pause", true: "play"}[p.paused]))
	}
	if p.recipe != nil {
		h = append(h, p.kv("n/p", "random"))
	} else if p.kind == playSurah {
		h = append(h, p.kv("n/p", "surah"))
	}
	if p.kind == playSurah {
		h = append(h, p.kv("<-/->", "seek"))
	}
	switch {
	case p.rep != nil:
		h = append([]string{p.kv("esc", "stop repeat"), p.kv("u", "times each")}, h...)
	case p.abFrom > 0:
		h = append([]string{p.kv("i", "mark last ayah"), p.kv("esc", "cancel")}, h...)
	case p.kind == playSurah:
		h = append(h, p.kv("e", "repeat ayahs"), p.kv("i", "mark A-B"))
	}
	var first, last []string
	switch p.page {
	case pQari:
		first = []string{p.kv("enter", "play"), p.kv("d", "download")}
		if len(p.qari.Moshaf) > 1 {
			last = append(last, p.kv("v", "recording"))
		}
		last = append(last, p.kv("esc", "back"))
	case pSurah:
		first, last = []string{p.kv("enter", "play"), p.kv("d", "download")}, []string{p.kv("esc", "back")}
	case pSurahs:
		first = []string{p.kv("enter", "choose qari"), p.kv("/", "search")}
	case pRadio:
		first = []string{p.kv("enter", "listen"), p.kv("f", "add to home"), p.kv("/", "search")}
	case pDownloads:
		first = []string{p.kv("enter", "open"), p.kv("x", "delete")}
	case pHistory:
		first = []string{p.kv("enter", "play again"), p.kv("x", "remove"), p.kv("/", "search")}
	case pSettings:
		first = []string{p.kv("enter", "change"), p.kv("← →", "adjust")}
	case pHome:
		if it := p.homeSel(); it != nil && it.prog != nil {
			first = []string{p.kv("enter", "resume"), p.kv("x", "remove")}
		} else if it != nil && it.radio != nil {
			first = []string{p.kv("enter", "listen"), p.kv("f", "remove")}
		} else if it != nil && p.Store.Favorites[it.rec.ID] {
			first = []string{p.kv("enter", "open"), p.kv("f", "unfavorite")}
		} else {
			first = []string{p.kv("enter", "open"), p.kv("f", "favorite")}
		}
		last = []string{p.kv("z", "random"), p.kv("/", "search"), p.kv("c", "theme")}
	default:
		first, last = []string{p.kv("enter", "open"), p.kv("/", "search")}, []string{p.kv("f", "favorite")}
	}
	h = append(append(first, h...), last...)
	return p.fitJoin(w, append(h, p.kv("+/-", "volume"), p.kv("tab", "next tab"))...)
}

func (p *painter) overlay(base []string) []string {
	p.zones = nil // the page under a popup is not clickable
	box := p.help()
	if p.ov != ovHelp {
		box = p.picker()
	}
	bw := width(box[0])
	top, left := max((p.h-len(box))/2, 0), max((p.w-bw)/2, 0)
	for i := range p.zones {
		z := &p.zones[i]
		z.x0, z.x1, z.y0, z.y1 = z.x0+left, z.x1+left, z.y0+top, z.y1+top
	}
	// The box goes first: zones are searched from the end, so rows win.
	p.zones = append([]zone{{zBox, 0, left, top, left + bw - 1, top + len(box) - 1}}, p.zones...)
	for i, l := range box {
		if top+i < len(base) {
			// Keep the page visible on both sides, mending Arabic isolates cut in half.
			b := base[top+i]
			lp, rp := ansi.Truncate(b, left, ""), ansi.TruncateLeft(b, left+bw, "")
			if strings.Count(lp, "\u2067") > strings.Count(lp, "\u2069") {
				lp += "\u2069"
			}
			if strings.Count(rp, "\u2069") > strings.Count(rp, "\u2067") {
				rp = strings.Replace(rp, "\u2069", "", 1)
			}
			base[top+i] = lp + spaces(left-width(lp)) + "\x1b[0m" + l + rp
		}
	}
	return base
}

func (p *painter) picker() []string {
	s, bw := p.s, min(p.w-4, 64)
	var title, info, foot string
	act, esc := "choose", "cancel"
	var row func(i int, sel bool) (string, string)
	var search []string
	check := func(on bool) string { return map[bool]string{true: s.Success.Render("✓ "), false: "  "}[on] }
	switch p.ov {
	case ovRiwayah:
		title, info = "✦ Filter by Riwayah", plural(len(p.riwayat)-1, "riwayah")
		row = func(i int, sel bool) (string, string) {
			o := p.riwayat[i]
			return check(o.id == p.Store.Settings.Rewaya) + p.name(sel, o.name), s.Muted.Render(plural(o.count, "qari"))
		}
	case ovRecording:
		title, info, foot = "✦ Choose Recording", trunc(p.qari.Name, 24), p.kv("d", "download")+s.Muted.Render(" · ")+p.kv("x", "delete")+s.Muted.Render(" · ")
		row = func(i int, sel bool) (string, string) {
			mo := p.qari.Moshaf[i]
			return check(mo.ID == p.moshaf.ID) + p.name(sel, label(p.cat, mo, false)), p.saved(p.qari.ID, mo) + s.Sub.Render(fmt.Sprintf("%d surahs", mo.SurahTotal))
		}
	case ovLang:
		title, info = "✦ Name Language", "qaris and surahs"
		row = func(i int, sel bool) (string, string) {
			l := p.langs[i]
			return check(l.Locale == p.Lang) + p.name(sel, l.Language), s.Muted.Render(l.Locale)
		}
	case ovTheme:
		title, info = "✦ Color Theme", "live preview"
		row = func(i int, sel bool) (string, string) {
			t := themes[i]
			return check(t.ID == p.themeWas.ID) + fit(p.name(sel, t.Name), 16) + s.Muted.Render(t.Desc), t.s.Accent.Render("●") + t.s.Accent2.Render("●") + t.s.Fav.Render("●")
		}
	case ovRandom:
		rs := p.recipes()
		title, info, act = "⤮ Random", fmt.Sprintf("%d built-in · %d yours", len(builtins), len(p.Store.Recipes)), "play"
		row = func(i int, sel bool) (string, string) {
			r := rs[i]
			return check(r.Name == p.Store.Settings.Recipe) + p.name(sel, bidi(r.Name)), s.Muted.Render(trunc(p.recipeDesc(r), (bw-8)*11/20))
		}
	case ovEdit:
		title, info, act, esc = "⤮ New recipe", "what it plays", "change", "back"
		if p.draftIdx >= 0 {
			title = "⤮ Edit recipe"
		}
		foot = p.kv("s", "save") + s.Muted.Render(" · ")
		fs := p.editFields()
		row = func(i int, sel bool) (string, string) {
			f := fs[i]
			if f == fSave {
				return s.Success.Render("✓ ") + p.name(sel, "Save recipe"), s.Muted.Render("or press s")
			}
			k, v := p.editRow(f)
			v = bidi(v)
			if f == fName && p.searching {
				p.input.Width = bw - 22
				v = bidi(p.input.View())
			}
			hint := map[int]string{fName: "type", fKind: "toggle", fKnown: "toggle"}[f]
			return "  " + fit(s.Muted.Render(k), 12) + p.name(sel, v), s.Muted.Render(cmpOr(hint, "choose"))
		}
	case ovField:
		k, _ := p.editRow(p.field)
		title, info, act, esc = "⤮ "+k, "/ search", "toggle", "back"
		items := p.fieldItems()
		row = func(i int, sel bool) (string, string) {
			it := items[i]
			return check(it.on) + p.name(sel, it.label), s.Muted.Render(it.right)
		}
		if p.searching || p.fieldQ != "" {
			v := p.fieldQ
			if p.searching {
				p.input.Width = bw - 10
				v = p.input.View()
			}
			search = []string{boxed(bw, s.AccentB.Render("/ ")+bidi(v), s.Accent)}
		}
	}
	if p.ov == ovAyahs {
		n := lastAyah(p.hifzT)
		title, info, act = "✦ Repeat ayahs", fmt.Sprintf("%03d %s · %s", p.hifzTrack.Surah, trunc(p.cat.SurahName(p.hifzTrack.Surah), 14), plural(n, "ayah")), "start"
		foot = p.kv("<-/->", "change") + s.Muted.Render(" · ")
		row = func(i int, sel bool) (string, string) {
			hint := map[int]string{0: fmt.Sprintf("1-%d", n), 1: fmt.Sprintf("1-%d", n), 3: "0 = until stopped", 4: "silence before a repeat"}[i]
			return "  " + fit(s.Muted.Render(hifzFields(&p.hifz, n)[i].label), 13) + p.name(sel, hifzValue(p.hifz, i)), s.Muted.Render(hint)
		}
	}
	n := p.pickN()
	vis := min(p.pickRows(), n)
	l := p.pick
	l.clamp(n, vis)
	out := append([]string{border(bw, true, s.AccentB.Render(title), s.Muted.Render(info), s.Accent)}, search...)
	if n == 0 {
		out = append(out, boxed(bw, s.Muted.Render("  No matches · esc clears the search"), s.Accent))
	}
	for i := l.offset; i < l.offset+vis; i++ {
		rail := "  "
		if i == l.cursor {
			rail = s.AccentB.Render("› ")
		}
		left, right := row(i, i == l.cursor)
		p.zone(zPick, i, 1, len(out), bw-2, len(out))
		out = append(out, boxed(bw, lr(rail+left, right, bw-4), s.Accent))
	}
	foot += p.kv("enter", act) + s.Muted.Render(" · ") + p.kv("esc", esc)
	if p.ov == ovRandom {
		foot = p.fitJoin(bw-12, p.kv("enter", act), p.kv("a", "new"), p.kv("e", "edit"), p.kv("c", "copy"), p.kv("x", "delete"))
	}
	return append(out, border(bw, false, foot, s.Muted.Render(fmt.Sprintf("%d/%d", l.cursor+1, n)), s.Accent))
}

func (p *painter) help() []string {
	s := p.s
	sec := func(g group, w int) []string {
		out := []string{s.AccentB.Render("✦ " + g.name)}
		for _, b := range g.b {
			out = append(out, " "+fit(s.Key.Render(b.label), 12)+s.Text.Render(trunc(b.desc, w-13)))
		}
		return out
	}
	top := s.Muted.Render(p.th.Name)
	if len(p.Version) <= 12 { // release tags only, not Go pseudo-versions
		top = p.join(top, s.Muted.Render(p.Version))
	}
	note := s.Muted.Render("Shortcuts follow key positions, so they work on the Arabic layout too.")
	if p.w >= 100 && p.h >= 24 {
		bw := min(p.w-6, 98)
		cw := (bw - 7) / 2
		l := append(append(sec(groups[0], cw), ""), sec(groups[2], cw)...)
		r := append(append(sec(groups[1], cw), ""), sec(groups[3], cw)...)
		out := []string{border(bw, true, s.AccentB.Render("✦ Keyboard Shortcuts"), top, s.Accent)}
		for _, line := range columns(l, r, cw, bw-4-cw-3, max(len(l), len(r)), " "+s.Rule.Render("│")+" ") {
			out = append(out, boxed(bw, line, s.Accent))
		}
		out = append(out, boxed(bw, "", s.Accent), boxed(bw, center(note, bw-4), s.Accent))
		return append(out, border(bw, false, p.kv("esc", "close"), s.Muted.Render("mouse: click, wheel, drag the bar"), s.Accent))
	}
	bw := min(p.w-2, 72)
	var body []string
	for i, g := range groups {
		if i > 0 {
			body = append(body, "")
		}
		body = append(body, sec(g, bw-4)...)
	}
	body = append(body, "", note)
	vis := min(len(body), p.h-2)
	off := min(p.helpOff, len(body)-vis)
	out := []string{border(bw, true, s.AccentB.Render("✦ Keyboard Shortcuts"), top, s.Accent)}
	for _, line := range body[off : off+vis] {
		out = append(out, boxed(bw, line, s.Accent))
	}
	return append(out, border(bw, false, p.kv("j/k", "scroll")+s.Muted.Render(" · ")+p.kv("esc", "close"), s.Muted.Render(fmt.Sprintf("%d-%d/%d", off+1, off+vis, len(body))), s.Accent))
}
