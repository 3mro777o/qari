package ui

import (
	"fmt"
	"slices"
	"time"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/keymap"
	"github.com/3mro777o/qari/internal/store"
	tea "github.com/charmbracelet/bubbletea"
)

type binding struct {
	keys        []string
	label, desc string
	do          func(m *Model, key string) tea.Cmd
}

type group struct {
	name string
	b    []binding
}

// groups is every shortcut; the help screen is generated from it.
var groups []group

func init() {
	nop := func(f func(m *Model)) func(*Model, string) tea.Cmd {
		return func(m *Model, _ string) tea.Cmd { f(m); return nil }
	}
	groups = []group{
		{"NAVIGATION", []binding{
			{[]string{"1", "2", "3", "4", "5"}, "1-5", "Go to tab: Home ... Downloads",
				func(m *Model, k string) tea.Cmd { m.goTab(int(k[0] - '1')); return nil }},
			{[]string{"tab", "shift+tab"}, "tab", "Next / previous tab (shift+tab)",
				func(m *Model, k string) tea.Cmd { m.goTab((m.tab + 5 + sign(k, "tab")) % 5); return nil }},
			{[]string{"j", "down", "k", "up"}, "j k", "Move down / up (arrow keys too)",
				func(m *Model, k string) tea.Cmd { m.move(sign(k, "j", "down")); return nil }},
			{[]string{"ctrl+d", "pgdown", "ctrl+u", "pgup"}, "ctrl+d u", "Half page down / up (pgdn pgup)",
				func(m *Model, k string) tea.Cmd { m.move(sign(k, "ctrl+d", "pgdown") * m.listRows() / 2); return nil }},
			{[]string{"g", "home", "G", "end"}, "g G", "First / last row (home end)",
				func(m *Model, k string) tea.Cmd { m.move(sign(k, "G", "end") * 99999); return nil }},
			{[]string{"enter", "l"}, "enter l", "Open, or play", (*Model).activateKey},
			{[]string{"esc", "h", "backspace"}, "esc h", "Back (clears the search first)", nop((*Model).escape)},
			{[]string{"/"}, "/", "Search (English, Arabic, number)", nop((*Model).startSearch)},
			{[]string{"o"}, "o", "Go to what is playing", nop((*Model).jump)},
		}},
		{"PLAYBACK", []binding{
			{[]string{" "}, "space", "Play / pause, or resume last", func(m *Model, _ string) tea.Cmd { return m.togglePause() }},
			{[]string{"left", "right"}, "left right", "Seek 10 seconds",
				func(m *Model, k string) tea.Cmd { return m.seekBy(float64(10 * sign(k, "right"))) }},
			{[]string{"<", ">", "shift+left", "shift+right"}, ", .", "Seek 1 minute (shift+arrows too)",
				func(m *Model, k string) tea.Cmd { return m.seekBy(float64(60 * sign(k, ">", "shift+right"))) }},
			{[]string{"n", "p"}, "n p", "Next / previous surah",
				func(m *Model, k string) tea.Cmd { return m.step(sign(k, "n")) }},
			{[]string{"+", "=", "-"}, "+ -", "Volume", func(m *Model, k string) tea.Cmd { return m.volume(float64(5 * sign(k, "+", "="))) }},
			{[]string{"m"}, "m", "Mute", func(m *Model, _ string) tea.Cmd { return m.mute() }},
			{[]string{"[", "]"}, "[ ]", "Slower / faster (0.75x-2x)",
				func(m *Model, k string) tea.Cmd { return m.setSpeed(m.speed + sign(k, "]")) }},
			{[]string{"r"}, "r", "After a surah: next/repeat/stop", (*Model).cycleRepeat},
			{[]string{"t"}, "t", "Sleep: 15-60 min / end of surah", (*Model).cycleSleep},
		}},
		{"LIBRARY", []binding{
			{[]string{"f"}, "f", "Favorite qari or radio (pins it on Home)", (*Model).favorite},
			{[]string{"a"}, "a", "Well-known qaris / all qaris", (*Model).showAll},
			{[]string{"s"}, "s", "Sort: popular, A-Z, updated", (*Model).cycleSort},
			{[]string{"w"}, "w", "Filter by riwayah", nop(func(m *Model) {
				m.openPick(ovRiwayah, slices.IndexFunc(m.riwayat, func(o riwayaOption) bool { return o.id == m.Store.Settings.Rewaya }))
			})},
			{[]string{"v"}, "v", "Choose a qari's recording", (*Model).chooseRecording},
			{[]string{"d"}, "d", "Download surah or recording", (*Model).download},
			{[]string{"x"}, "x", "Remove: from Continue Listening, a download, or cancel it", (*Model).remove},
		}},
		{"APP", []binding{
			{[]string{"c"}, "c", "Color theme (live preview)", nop(func(m *Model) {
				m.themeWas = m.th
				m.openPick(ovTheme, themeIndex(m.th.ID))
			})},
			{[]string{"?"}, "?", "This help", nop(func(m *Model) { m.ov, m.helpOff = ovHelp, 0 })},
			{[]string{"q"}, "q", "Quit (your place is saved)", func(m *Model, _ string) tea.Cmd { return m.quit() }},
		}},
	}
}

// sign is 1 if k is one of plus, else -1.
func sign(k string, plus ...string) int {
	if slices.Contains(plus, k) {
		return 1
	}
	return -1
}

// keyOf names a key press; a character typed on the Arabic layout becomes
// the US key in the same position, so shortcuts work on both.
func keyOf(msg tea.KeyMsg) string {
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && !msg.Alt {
		return string(keymap.LatinKey(msg.Runes[0]))
	}
	return msg.String()
}

func (m *Model) onKey(msg tea.KeyMsg) tea.Cmd {
	if msg.Type == tea.KeyRunes && len(msg.Runes) > 1 && !m.searching {
		// Fast typing arrives as one message: "/" followed by search text.
		var cmds []tea.Cmd
		for _, r := range msg.Runes {
			cmds = append(cmds, m.onKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}))
		}
		return tea.Batch(cmds...)
	}
	k := keyOf(msg)
	m.armed, m.confirm = m.confirm, ""
	switch {
	case k == "ctrl+c":
		return m.quit()
	case m.loading || m.loadErr != nil:
		if k == "q" {
			return tea.Quit
		}
		if m.loadErr != nil && (k == "enter" || k == "r") {
			m.loading, m.loadErr = true, nil
			return tea.Batch(m.spin.Tick, m.load())
		}
		return nil
	case m.searching:
		return m.searchKey(msg)
	case m.ov == ovHelp:
		switch k {
		case "j", "down", "k", "up":
			m.helpOff = max(m.helpOff+sign(k, "j", "down"), 0)
		default:
			m.ov = ovNone
		}
		return nil
	case m.ov != ovNone:
		return m.pickKey(k)
	}
	for _, g := range groups {
		for _, b := range g.b {
			if slices.Contains(b.keys, k) {
				return b.do(m, k)
			}
		}
	}
	return nil
}

// searchKey edits the search text; keys here are raw, not layout-mapped.
func (m *Model) searchKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		m.list().query = ""
		fallthrough
	case tea.KeyEnter:
		m.searching = false
		m.input.Blur()
		m.clamp()
		return nil
	case tea.KeyDown, tea.KeyUp:
		m.move(sign(msg.String(), "down"))
		return nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if l := m.list(); l.query != m.input.Value() {
		*l = listState{query: m.input.Value()}
	}
	return cmd
}

func (m *Model) pickKey(k string) tea.Cmd {
	switch k {
	case "esc", "q", "h", "backspace", "w", "v", "c":
		m.closePick()
	case "j", "down", "k", "up":
		m.pickMove(sign(k, "j", "down"))
	case "g", "home", "G", "end":
		m.pickMove(999 * sign(k, "G", "end"))
	case "enter", "l":
		return m.choose()
	case "d":
		if m.ov == ovRecording {
			return m.download("")
		}
	case "x":
		if m.ov == ovRecording {
			return m.remove("")
		}
	}
	return nil
}

func (m *Model) pickN() int {
	switch m.ov {
	case ovRiwayah:
		return len(m.riwayat)
	case ovRecording:
		return len(m.qari.Moshaf)
	}
	return len(themes)
}

func (m *Model) pickRows() int { return max(3, min(12, m.h-8)) }

func (m *Model) openPick(ov, cursor int) {
	m.ov, m.pick = ov, listState{cursor: cursor}
	m.pick.clamp(m.pickN(), m.pickRows())
}

func (m *Model) pickMove(d int) {
	m.pick.cursor += d
	m.pick.clamp(m.pickN(), m.pickRows())
	if m.ov == ovTheme {
		m.th = themes[m.pick.cursor]
		m.restyle()
	}
}

func (m *Model) closePick() {
	if m.ov == ovTheme {
		m.th = m.themeWas
		m.restyle()
	}
	m.ov = ovNone
}

func (m *Model) choose() tea.Cmd {
	i, ov := m.pick.cursor, m.ov
	m.ov = ovNone
	switch ov {
	case ovRiwayah:
		m.Store.Settings.Rewaya = m.riwayat[i].id
		m.lists[pQaris], m.lists[pSurah] = listState{query: m.lists[pQaris].query}, listState{query: m.lists[pSurah].query}
		m.save()
		return m.flash("Riwayah: " + m.riwayat[i].name)
	case ovRecording:
		m.moshaf = m.qari.Moshaf[i]
		m.lists[pQari] = listState{}
		return m.flash("Recording: " + label(m.cat, m.moshaf, false))
	case ovTheme:
		m.th = themes[i]
		m.restyle()
		m.Store.Settings.Theme = m.th.ID
		m.save()
		return m.flash("Theme: " + m.th.Name)
	}
	return nil
}

func (m *Model) move(d int) {
	m.list().cursor += d
	m.clamp()
}

func (m *Model) activateKey(string) tea.Cmd { return m.activate() }

func (m *Model) activate() tea.Cmd {
	i, q := m.list().cursor, m.list().query
	switch m.page {
	case pSurahs:
		if items := m.surahs(nil, q); i < len(items) {
			m.page, m.surah, m.lists[pSurah] = pSurah, items[i].ID, listState{}
		}
	case pRadio:
		if items := m.radios(); i < len(items) {
			return m.playRadio(items[i])
		}
	default:
		if it := m.homeSel(); it != nil && it.radio != nil {
			return m.playRadio(*it.radio)
		}
		rec, mo, n := m.selected()
		switch {
		case rec == nil:
		case n > 0:
			return m.playSurah(store.Track{Reciter: rec.ID, Moshaf: mo.ID, Surah: n}, false)
		case m.page != pQari:
			m.openQari(rec, mo.ID, 0)
		}
	}
	return nil
}

func (m *Model) escape() {
	if l := m.list(); l.query != "" {
		*l = listState{}
	} else {
		m.back()
	}
}

func (m *Model) startSearch() {
	m.searching = true
	m.input.SetValue(m.list().query)
	m.input.CursorEnd()
	m.input.Focus()
}

func (m *Model) jump() {
	switch {
	case m.kind == playRadio:
		m.goTab(pRadio)
		m.lists[pRadio] = listState{cursor: slices.Index(m.cat.Radios, m.radio)}
		m.clamp()
	case m.kind == playSurah:
		m.openQari(m.cat.Reciter(m.track.Reciter), m.track.Moshaf, m.track.Surah)
	case m.Store.Last != nil && m.cat.Reciter(m.Store.Last.Reciter) != nil:
		m.openQari(m.cat.Reciter(m.Store.Last.Reciter), m.Store.Last.Moshaf, m.Store.Last.Surah)
	}
}

func (m *Model) togglePause() tea.Cmd {
	switch {
	case m.Player == nil:
		return m.flash("Install mpv to play audio")
	case m.kind != playNone:
		m.Player.TogglePause()
		if m.paused = !m.paused; m.paused {
			m.save()
		}
	case m.Store.Last != nil && m.cat.Reciter(m.Store.Last.Reciter) != nil:
		return m.playSurah(*m.Store.Last, false)
	default:
		return m.flash("Choose a surah or a station first")
	}
	return nil
}

func (m *Model) seekBy(sec float64) tea.Cmd {
	if m.kind == playSurah {
		m.Player.Seek(sec)
		m.pos = max(0, m.pos+sec)
		if m.dur > 0 {
			m.pos = min(m.pos, m.dur)
		}
	}
	return nil
}

func (m *Model) step(d int) tea.Cmd {
	if m.kind != playSurah {
		return nil
	}
	if next, ok := m.adjacent(d); ok {
		return m.playSurah(next, false)
	}
	return m.flash("No more surahs in this recording")
}

func (m *Model) volume(d float64) tea.Cmd {
	if m.Player == nil {
		return nil
	}
	m.vol = m.Player.AddVolume(d)
	if m.muted && d > 0 {
		return m.mute()
	}
	return m.flash(fmt.Sprintf("Volume %.0f%%", m.vol))
}

func (m *Model) mute() tea.Cmd {
	if m.Player == nil {
		return nil
	}
	m.muted = !m.muted
	m.Player.SetMute(m.muted)
	if m.muted {
		return m.flash("Muted")
	}
	return m.flash(fmt.Sprintf("Volume %.0f%%", m.vol))
}

func (m *Model) setSpeed(i int) tea.Cmd {
	if m.Player == nil {
		return nil
	}
	m.speed = max(0, min(i, len(speeds)-1))
	if m.kind != playRadio {
		m.Player.SetSpeed(speeds[m.speed])
	}
	return m.flash(fmt.Sprintf("Speed %gx", speeds[m.speed]))
}

func (m *Model) cycleRepeat(string) tea.Cmd {
	i := (mode(repeatModes, m.Store.Settings.Repeat) + 1) % len(repeatModes)
	m.Store.Settings.Repeat = repeatModes[i]
	m.save()
	return m.flash("After a surah: " + repeatLong[i])
}

func (m *Model) cycleSleep(string) tea.Cmd {
	m.sleep = (m.sleep + 1) % len(sleeps)
	m.sleepSeq++
	switch mins := sleeps[m.sleep]; mins {
	case 0:
		return m.flash("Sleep timer off")
	case -1:
		return m.flash("Sleep timer: end of this surah")
	default:
		m.sleepEnd = time.Now().Add(time.Duration(mins) * time.Minute)
		return tea.Batch(m.flash(fmt.Sprintf("Sleep timer: %d min", mins)), m.sleepCmd())
	}
}

func (m *Model) favorite(string) tea.Cmd {
	var radio *api.Radio
	if it := m.homeSel(); it != nil {
		radio = it.radio
	} else if items, i := m.radios(), m.list().cursor; m.page == pRadio && m.ov == ovNone && i < len(items) {
		radio = &items[i]
	}
	if radio != nil {
		on := !m.Store.FavRadios[radio.ID]
		if on {
			m.Store.FavRadios[radio.ID] = true
		} else {
			delete(m.Store.FavRadios, radio.ID)
		}
		m.save()
		m.clamp()
		return m.flash(map[bool]string{true: "★ On Home: ", false: "Removed from Home: "}[on] + radio.Name)
	}
	rec, _, _ := m.selected()
	if rec == nil {
		return nil
	}
	on := !m.Store.Favorites[rec.ID]
	if on {
		m.Store.Favorites[rec.ID] = true
	} else {
		delete(m.Store.Favorites, rec.ID)
	}
	m.save()
	m.clamp()
	if on {
		return m.flash("★ Favorite: " + rec.Name)
	}
	return m.flash("Removed from favorites: " + rec.Name)
}

func (m *Model) showAll(string) tea.Cmd {
	st := &m.Store.Settings
	st.ShowAll = !st.ShowAll
	m.save()
	m.clamp()
	if st.ShowAll {
		return m.flash("Showing all qaris")
	}
	return m.flash("Showing well-known qaris")
}

func (m *Model) cycleSort(string) tea.Cmd {
	i := (mode(sortModes, m.Store.Settings.Sort) + 1) % len(sortModes)
	m.Store.Settings.Sort = sortModes[i]
	m.save()
	m.clamp()
	return m.flash("Sort: " + sortNames[i])
}

func (m *Model) chooseRecording(string) tea.Cmd {
	if m.page != pQari {
		rec, mo, _ := m.selected()
		if rec == nil {
			return nil
		}
		m.openQari(rec, mo.ID, 0)
	}
	m.openPick(ovRecording, slices.IndexFunc(m.qari.Moshaf, func(mo api.Moshaf) bool { return mo.ID == m.moshaf.ID }))
	return nil
}

// target is what d and x act on: the highlighted recording in the
// recording picker, else the row under the cursor.
func (m *Model) target() (rec *api.ReciterEntry, mo api.Moshaf, surah int) {
	if m.ov == ovRecording {
		return m.qari, m.qari.Moshaf[m.pick.cursor], 0
	}
	return m.selected()
}

// twice asks for a second press of the same key before a big action.
func (m *Model) twice(id, ask string) (bool, tea.Cmd) {
	if m.armed == id {
		return true, nil
	}
	m.confirm = id
	return false, m.flash(ask)
}

func (m *Model) download(string) tea.Cmd {
	rec, mo, n := m.target()
	if rec == nil || mo.ID == 0 {
		return nil
	}
	ids := []int{n}
	if n == 0 {
		ids = slices.DeleteFunc(surahIDs(mo), func(s int) bool { return m.Downloads.Has(store.Track{Reciter: rec.ID, Moshaf: mo.ID, Surah: s}) })
		if len(ids) == 0 {
			return m.flash("Already downloaded")
		}
		if ok, cmd := m.twice(fmt.Sprint("d", mo.ID), fmt.Sprintf("Press d again to download %s of %s", plural(len(ids), "surah"), rec.Name)); !ok {
			return cmd
		}
	}
	added := 0
	for _, s := range ids {
		added += b2i(m.Downloads.Add(store.Track{Reciter: rec.ID, Moshaf: mo.ID, Surah: s}, mo.URL(s)))
	}
	if added == 0 {
		return m.flash("Already downloaded or queued")
	}
	return m.flash(fmt.Sprintf("Downloading %s · 5 to watch", plural(added, "surah")))
}

func (m *Model) remove(string) tea.Cmd {
	if it := m.homeSel(); it != nil && it.prog != nil {
		m.Store.Forget(it.prog.Track)
		m.clamp()
		return m.flash("Removed from Continue Listening")
	}
	rec, mo, n := m.target()
	if rec == nil {
		return nil
	}
	var saved []store.Track
	for _, s := range surahIDs(mo) {
		if t := (store.Track{Reciter: rec.ID, Moshaf: mo.ID, Surah: s}); (n == 0 || n == s) && m.Downloads.Has(t) {
			saved = append(saved, t)
		}
	}
	switch {
	case n > 0 && m.Downloads.Cancel(store.Track{Reciter: rec.ID, Moshaf: mo.ID, Surah: n}):
		return m.flash("Download cancelled")
	case len(saved) == 0:
		return m.flash("Nothing downloaded here")
	case len(saved) > 1:
		if ok, cmd := m.twice(fmt.Sprint("x", mo.ID), fmt.Sprintf("Press x again to delete %s", plural(len(saved), "downloaded surah"))); !ok {
			return cmd
		}
	}
	for _, t := range saved {
		m.Downloads.Delete(t)
	}
	m.clamp()
	return m.flash("Deleted " + plural(len(saved), "download"))
}
