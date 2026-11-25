// Package ui is the Bubble Tea interface: tabs of searchable lists on top and
// a player bar at the bottom.
package ui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/player"
	"github.com/3mro777o/qari/internal/popular"
	"github.com/3mro777o/qari/internal/store"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Options struct {
	Client   *api.Client
	Lang     string
	Store    *store.Store
	Player   *player.Player // nil when mpv is not installed
	Popular  map[int]popular.Entry
	NoArabic bool
	Version  string
}

const (
	tabHome = iota
	tabReciters
	tabSurahs
	tabRadio
)

var tabNames = []string{"Home", "Qaris", "Surahs", "Radio"}

type screenKind int

const (
	scrHome screenKind = iota
	scrReciters
	scrMoshafs
	scrSurahs
	scrAllSurahs
	scrRadio
)

type screen struct {
	kind    screenKind
	list    *listView
	reciter *api.ReciterEntry
	moshaf  *api.Moshaf
	surah   int // scrReciters: only reciters who recorded this surah
}

// Row payloads.
type (
	moshafChoice struct {
		r *api.ReciterEntry
		m *api.Moshaf
	}
	surahChoice struct {
		r *api.ReciterEntry
		m *api.Moshaf
		n int
	}
	allSurah  int
	resumeRow store.Track
)

type nowPlaying struct {
	track   store.Track
	reciter *api.ReciterEntry
	moshaf  *api.Moshaf
	radio   *api.Radio
	ev      player.Event
}

func (n *nowPlaying) isRadio() bool { return n.radio != nil }

var (
	sortModes    = []string{"popular", "a-z", "recent"}
	repeatModes  = []string{"next", "one", "off"}
	repeatLabel  = map[string]string{"next": "autoplay", "one": "repeat surah", "off": "stop after surah"}
	sleepOptions = []time.Duration{0, 15 * time.Minute, 30 * time.Minute, 45 * time.Minute, 60 * time.Minute, -1} // -1: end of surah
	speeds       = []float64{0.75, 0.9, 1, 1.1, 1.25, 1.5, 1.75, 2}
)

type Model struct {
	opt     Options
	cat     *api.Catalog
	loadErr error
	spin    spinner.Model
	w, h    int

	tab       int
	stacks    [4][]*screen
	searching bool
	input     textinput.Model

	now      *nowPlaying
	speedIdx int
	volume   float64
	sleepIdx int
	sleepAt  time.Time
	lastSave time.Time

	status    string
	statusErr bool
	statusAt  time.Time
	showHelp  bool
	arabic    bool
}

type (
	catalogMsg struct {
		cat *api.Catalog
		err error
	}
	playerMsg player.Event
	tickMsg   time.Time
)

func New(opt Options) Model {
	in := textinput.New()
	in.Prompt = "/ "
	in.PromptStyle = sAccent
	in.Placeholder = "search in English, العربية, or any listed language"
	in.CharLimit = 64
	sp := spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(sAccent))
	s := &opt.Store.Settings
	if s.Sort == "" {
		s.Sort = "popular"
	}
	if s.Repeat == "" {
		s.Repeat = "next"
	}
	return Model{opt: opt, spin: sp, input: in, speedIdx: 2, volume: 100, arabic: !opt.NoArabic}
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.spin.Tick, tick(), m.loadCatalog()}
	if m.opt.Player != nil {
		cmds = append(cmds, listen(m.opt.Player))
	}
	return tea.Batch(cmds...)
}

func (m Model) loadCatalog() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		cat, err := api.LoadCatalog(ctx, m.opt.Client, m.opt.Lang)
		return catalogMsg{cat, err}
	}
}

func listen(p *player.Player) tea.Cmd {
	return func() tea.Msg { return playerMsg(<-p.Events) }
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// ---- update ----------------------------------------------------------------

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.input.Width = max(m.w-6, 10)
		return m, nil

	case spinner.TickMsg:
		if m.cat != nil || m.loadErr != nil {
			return m, nil
		}
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case catalogMsg:
		if msg.err != nil {
			m.loadErr = msg.err
			return m, nil
		}
		m.cat = msg.cat
		for _, r := range m.cat.Reciters {
			if p, ok := m.opt.Popular[r.ID]; ok {
				r.Listens, r.Known = p.Listens, p.Known
			}
		}
		m.stacks = [4][]*screen{
			{{kind: scrHome}},
			{{kind: scrReciters}},
			{{kind: scrAllSurahs}},
			{{kind: scrRadio}},
		}
		m.refreshAll()
		if m.opt.Client.Offline() {
			m.flash("Offline: showing the saved catalog", true)
		}
		return m, nil

	case playerMsg:
		m.onPlayer(player.Event(msg))
		return m, listen(m.opt.Player)

	case tickMsg:
		if m.sleepIdx > 0 && sleepOptions[m.sleepIdx] > 0 && !m.sleepAt.IsZero() && time.Now().After(m.sleepAt) {
			m.sleepIdx, m.sleepAt = 0, time.Time{}
			if m.now != nil && !m.now.ev.Paused {
				m.opt.Player.TogglePause()
			}
			m.flash("Sleep timer: paused", false)
		}
		if time.Since(m.lastSave) > 15*time.Second {
			m.save()
		}
		if m.status != "" && time.Since(m.statusAt) > 4*time.Second {
			m.status = ""
		}
		return m, tick()

	case tea.KeyMsg:
		return m.onKey(msg)
	}
	return m, nil
}

func (m *Model) flash(s string, isErr bool) {
	m.status, m.statusErr, m.statusAt = s, isErr, time.Now()
}

func (m *Model) save() {
	m.lastSave = time.Now()
	if err := m.opt.Store.Save(); err != nil {
		m.flash("Could not save progress: "+err.Error(), true)
	}
}

func (m *Model) top() *screen {
	st := m.stacks[m.tab]
	if len(st) == 0 {
		return nil
	}
	return st[len(st)-1]
}

func (m *Model) push(s *screen) {
	s.list = newList(m.rowsFor(s), "Nothing here")
	m.stacks[m.tab] = append(m.stacks[m.tab], s)
}

func (m *Model) pop() {
	if len(m.stacks[m.tab]) > 1 {
		m.stacks[m.tab] = m.stacks[m.tab][:len(m.stacks[m.tab])-1]
	}
}

// refreshAll rebuilds the rows of every open screen, e.g. after a setting,
// favourite or progress change.
func (m *Model) refreshAll() {
	for t := range m.stacks {
		for _, s := range m.stacks[t] {
			rows := m.rowsFor(s)
			if s.list == nil {
				s.list = newList(rows, "Nothing here")
			} else {
				s.list.setRows(rows)
			}
		}
	}
}

func (m Model) onKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Keys typed faster than a frame arrive as one message; outside the search
	// box each one is a separate command.
	if k.Type == tea.KeyRunes && len(k.Runes) > 1 && !m.searching && !k.Paste {
		var model tea.Model = m
		var cmds []tea.Cmd
		for _, r := range k.Runes {
			var cmd tea.Cmd
			model, cmd = model.(Model).onKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			cmds = append(cmds, cmd)
		}
		return model, tea.Batch(cmds...)
	}
	key := k.String()
	if key == "ctrl+c" {
		return m.quit()
	}
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}
	if m.cat == nil {
		if key == "q" || key == "esc" {
			return m.quit()
		}
		return m, nil
	}
	s := m.top()

	if m.searching {
		switch key {
		case "esc":
			m.searching = false
			m.input.Blur()
			m.input.SetValue("")
			s.list.setQuery("")
		case "enter":
			m.searching = false
			m.input.Blur()
			if s.list.selected() != nil {
				return m.activate()
			}
		case "up", "ctrl+p":
			s.list.move(-1)
		case "down", "ctrl+n":
			s.list.move(1)
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(k)
			s.list.setQuery(m.input.Value())
			return m, cmd
		}
		return m, nil
	}

	p := m.opt.Player
	switch key {
	case "q":
		return m.quit()
	case "?":
		m.showHelp = true
	case "tab", "shift+tab", "1", "2", "3", "4":
		switch key {
		case "tab":
			m.tab = (m.tab + 1) % len(tabNames)
		case "shift+tab":
			m.tab = (m.tab + len(tabNames) - 1) % len(tabNames)
		default:
			m.tab = int(key[0] - '1')
		}
		m.input.SetValue(m.top().list.query)
	case "/":
		m.searching = true
		m.input.SetValue(s.list.query)
		m.input.CursorEnd()
		return m, m.input.Focus()
	case "esc", "backspace":
		if s.list.query != "" {
			s.list.setQuery("")
			m.input.SetValue("")
		} else {
			m.pop()
			m.input.SetValue(m.top().list.query)
		}
	case "up", "k":
		s.list.move(-1)
	case "down", "j":
		s.list.move(1)
	case "pgup", "ctrl+u":
		s.list.move(-m.listHeight() / 2)
	case "pgdown", "ctrl+d":
		s.list.move(m.listHeight() / 2)
	case "home", "g":
		s.list.home()
	case "end", "G":
		s.list.end()
	case "enter":
		return m.activate()
	case "f":
		m.toggleFavorite()
	case "a":
		st := &m.opt.Store.Settings
		st.ShowAll = !st.ShowAll
		m.refreshAll()
		if st.ShowAll {
			m.flash("Showing every qari", false)
		} else {
			m.flash("Showing well-known qaris with a complete Quran", false)
		}
	case "s":
		st := &m.opt.Store.Settings
		st.Sort = next(sortModes, st.Sort)
		m.refreshAll()
		s.list.home()
		m.flash("Sort: "+st.Sort, false)
	case "w":
		m.cycleRewaya()
	case "A":
		m.arabic = !m.arabic
	case "o":
		m.openNowPlaying()
	}

	if p == nil || m.now == nil {
		switch key {
		case " ", "left", "right", "n", "p", "x", "[", "]":
			if p == nil {
				m.flash("mpv is not installed, so nothing can play", true)
			}
		}
		if key != "r" && key != "t" && key != "-" && key != "+" && key != "=" && key != "<" && key != ">" {
			return m, nil
		}
	}
	switch key {
	case " ":
		p.TogglePause()
	case "left":
		p.Seek(-10)
	case "right":
		p.Seek(10)
	case "[":
		p.Seek(-60)
	case "]":
		p.Seek(60)
	case "n":
		m.step(1)
	case "p":
		m.step(-1)
	case "x":
		m.stop()
	case "r":
		st := &m.opt.Store.Settings
		st.Repeat = next(repeatModes, st.Repeat)
		m.flash("After a surah: "+repeatLabel[st.Repeat], false)
	case "t":
		m.sleepIdx = (m.sleepIdx + 1) % len(sleepOptions)
		d := sleepOptions[m.sleepIdx]
		switch {
		case d == 0:
			m.sleepAt = time.Time{}
			m.flash("Sleep timer off", false)
		case d < 0:
			m.sleepAt = time.Time{}
			m.flash("Sleep timer: pause at the end of this surah", false)
		default:
			m.sleepAt = time.Now().Add(d)
			m.flash(fmt.Sprintf("Sleep timer: pause in %d min", int(d.Minutes())), false)
		}
	case "-", "+", "=":
		if p != nil {
			d := 5.0
			if key == "-" {
				d = -5
			}
			m.volume = p.AddVolume(d)
			m.flash(fmt.Sprintf("Volume %d%%", int(m.volume)), false)
		}
	case "<", ">":
		if p != nil {
			if key == "<" {
				m.speedIdx = max(m.speedIdx-1, 0)
			} else {
				m.speedIdx = min(m.speedIdx+1, len(speeds)-1)
			}
			p.SetSpeed(speeds[m.speedIdx])
			m.flash(fmt.Sprintf("Speed %.2gx", speeds[m.speedIdx]), false)
		}
	}
	return m, nil
}

func next(list []string, cur string) string {
	for i, v := range list {
		if v == cur {
			return list[(i+1)%len(list)]
		}
	}
	return list[0]
}

func (m Model) quit() (tea.Model, tea.Cmd) {
	if m.opt.Player != nil {
		m.opt.Player.Close()
	}
	m.opt.Store.Save()
	return m, tea.Quit
}

func (m *Model) cycleRewaya() {
	ids := []int{0}
	seen := map[int]bool{}
	for _, r := range m.cat.Reciters {
		for _, ms := range r.Moshaf {
			if !seen[ms.RewayaID] {
				seen[ms.RewayaID] = true
				ids = append(ids, ms.RewayaID)
			}
		}
	}
	sort.Ints(ids[1:])
	st := &m.opt.Store.Settings
	for i, id := range ids {
		if id == st.Rewaya {
			st.Rewaya = ids[(i+1)%len(ids)]
			break
		}
	}
	m.refreshAll()
	if st.Rewaya == 0 {
		m.flash("Riwayah: all", false)
	} else {
		m.flash("Riwayah: "+m.cat.Riwayat[st.Rewaya], false)
	}
}

func (m *Model) toggleFavorite() {
	var r *api.ReciterEntry
	if sel := m.top().list.selected(); sel != nil {
		switch d := sel.data.(type) {
		case *api.ReciterEntry:
			r = d
		case moshafChoice:
			r = d.r
		case surahChoice:
			r = d.r
		}
	}
	if r == nil {
		r = m.top().reciter
	}
	if r == nil {
		return
	}
	fav := m.opt.Store.Favorites
	if fav[r.ID] {
		delete(fav, r.ID)
		m.flash("Removed from favorites: "+r.Name, false)
	} else {
		fav[r.ID] = true
		m.flash("Added to favorites: "+r.Name, false)
	}
	m.refreshAll()
	m.save()
}

// activate handles enter on the selected row.
func (m Model) activate() (tea.Model, tea.Cmd) {
	s := m.top()
	sel := s.list.selected()
	if sel == nil {
		return m, nil
	}
	switch d := sel.data.(type) {
	case *api.ReciterEntry:
		if s.surah > 0 {
			if ms := m.bestMoshaf(d, s.surah); ms != nil {
				m.play(d, ms, s.surah)
			}
			return m, nil
		}
		m.openReciter(d)
	case moshafChoice:
		m.push(&screen{kind: scrSurahs, reciter: d.r, moshaf: d.m})
		m.jumpToResume()
	case surahChoice:
		m.play(d.r, d.m, d.n)
	case allSurah:
		m.push(&screen{kind: scrReciters, surah: int(d)})
	case *api.Radio:
		m.playRadio(d)
	case resumeRow:
		r := m.cat.Reciter(d.Reciter)
		if r == nil {
			return m, nil
		}
		for i := range r.Moshaf {
			if r.Moshaf[i].ID == d.Moshaf {
				m.play(r, &r.Moshaf[i], d.Surah)
			}
		}
	}
	m.input.SetValue(m.top().list.query)
	return m, nil
}

func (m *Model) openReciter(r *api.ReciterEntry) {
	ms := m.moshafsFor(r)
	if len(ms) == 1 {
		m.push(&screen{kind: scrSurahs, reciter: r, moshaf: ms[0]})
		m.jumpToResume()
		return
	}
	m.push(&screen{kind: scrMoshafs, reciter: r})
}

// jumpToResume puts the cursor on the surah last played in this moshaf.
func (m *Model) jumpToResume() {
	s := m.top()
	if s.kind != scrSurahs {
		return
	}
	var best *store.Progress
	for _, p := range m.opt.Store.Progress {
		if p.Reciter == s.reciter.ID && p.Moshaf == s.moshaf.ID && (best == nil || p.Updated.After(best.Updated)) {
			best = p
		}
	}
	if best == nil {
		return
	}
	n := best.Surah
	if best.Finished {
		n = nextSurah(s.moshaf, n, 1)
	}
	for i, idx := range s.list.visible {
		if c, ok := s.list.rows[idx].data.(surahChoice); ok && c.n == n {
			s.list.cursor = i
		}
	}
}

func (m *Model) openNowPlaying() {
	if m.now == nil || m.now.isRadio() {
		return
	}
	m.tab = tabReciters
	m.stacks[m.tab] = m.stacks[m.tab][:1]
	m.push(&screen{kind: scrSurahs, reciter: m.now.reciter, moshaf: m.now.moshaf})
	for i, idx := range m.top().list.visible {
		if c, ok := m.top().list.rows[idx].data.(surahChoice); ok && c.n == m.now.track.Surah {
			m.top().list.cursor = i
		}
	}
	m.input.SetValue("")
}

// moshafsFor lists a reciter's recordings, best first: the chosen riwayah
// (Hafs when none is chosen), complete, then Murattal before other styles.
func (m *Model) moshafsFor(r *api.ReciterEntry) []*api.Moshaf {
	var out []*api.Moshaf
	for i := range r.Moshaf {
		out = append(out, &r.Moshaf[i])
	}
	rw := m.opt.Store.Settings.Rewaya
	if rw == 0 {
		rw = 1 // Hafs an Asim, the most widely recited
	}
	style := func(ms *api.Moshaf) int { // moshaf_type is riwayah*10 + style; 1 = Murattal
		if s := ms.MoshafType % 10; s >= 1 {
			return s
		}
		return 9
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if (a.RewayaID == rw) != (b.RewayaID == rw) {
			return a.RewayaID == rw
		}
		if (a.SurahTotal >= 114) != (b.SurahTotal >= 114) {
			return a.SurahTotal >= 114
		}
		return style(a) < style(b)
	})
	return out
}

func (m *Model) bestMoshaf(r *api.ReciterEntry, surah int) *api.Moshaf {
	for _, ms := range m.moshafsFor(r) {
		if ms.Surahs()[surah] {
			return ms
		}
	}
	return nil
}

func nextSurah(ms *api.Moshaf, cur, dir int) int {
	have := ms.Surahs()
	for n := cur + dir; n >= 1 && n <= 114; n += dir {
		if have[n] {
			return n
		}
	}
	return 0
}

// ---- playback --------------------------------------------------------------

func (m *Model) play(r *api.ReciterEntry, ms *api.Moshaf, n int) {
	if m.opt.Player == nil {
		m.flash("mpv is not installed: install it to play audio (e.g. pacman -S mpv, apt install mpv, brew install mpv)", true)
		return
	}
	m.saveCurrent()
	t := store.Track{Reciter: r.ID, Moshaf: ms.ID, Surah: n}
	start := m.opt.Store.ResumeAt(t)
	title := m.cat.SurahName(n) + " — " + r.Name
	if err := m.opt.Player.Play(ms.URL(n), start, title); err != nil {
		m.flash(err.Error(), true)
		return
	}
	m.opt.Player.SetSpeed(speeds[m.speedIdx])
	m.now = &nowPlaying{track: t, reciter: r, moshaf: ms, ev: player.Event{Loading: true, Pos: start}}
	m.opt.Store.Plays[r.ID]++
	m.opt.Store.Last = &t
	if start > 0 {
		m.flash("Resuming at "+clock(start), false)
	}
	m.refreshAll()
}

func (m *Model) playRadio(r *api.Radio) {
	if m.opt.Player == nil {
		m.flash("mpv is not installed, so nothing can play", true)
		return
	}
	m.saveCurrent()
	if err := m.opt.Player.Play(r.URL, 0, r.Name); err != nil {
		m.flash(err.Error(), true)
		return
	}
	m.now = &nowPlaying{radio: r, ev: player.Event{Loading: true}}
}

func (m *Model) stop() {
	m.saveCurrent()
	m.opt.Player.Stop()
	m.now = nil
	m.refreshAll()
}

// saveCurrent stores the position of the playing surah before switching away.
func (m *Model) saveCurrent() {
	if m.now != nil && !m.now.isRadio() && m.now.ev.Pos > 0 {
		m.opt.Store.Update(m.now.track, m.now.ev.Pos, m.now.ev.Duration)
		m.save()
	}
}

func (m *Model) step(dir int) {
	if m.now == nil || m.now.isRadio() {
		return
	}
	n := nextSurah(m.now.moshaf, m.now.track.Surah, dir)
	if n == 0 {
		m.flash("No more surahs in this recitation", false)
		return
	}
	m.play(m.now.reciter, m.now.moshaf, n)
}

func (m *Model) onPlayer(ev player.Event) {
	if m.now == nil {
		return
	}
	if ev.Err != nil {
		m.flash(ev.Err.Error(), true)
		m.now.ev.Loading = false
		return
	}
	m.now.ev = ev
	if m.now.isRadio() {
		return
	}
	if !ev.Ended {
		if ev.Pos > 0 {
			m.opt.Store.Update(m.now.track, ev.Pos, ev.Duration)
		}
		return
	}
	m.opt.Store.MarkFinished(m.now.track)
	m.save()
	if sleepOptions[m.sleepIdx] < 0 {
		m.sleepIdx = 0
		m.now = nil
		m.flash("Sleep timer: stopped after the surah", false)
		m.refreshAll()
		return
	}
	switch m.opt.Store.Settings.Repeat {
	case "one":
		m.play(m.now.reciter, m.now.moshaf, m.now.track.Surah)
	case "next":
		if n := nextSurah(m.now.moshaf, m.now.track.Surah, 1); n != 0 {
			m.play(m.now.reciter, m.now.moshaf, n)
			return
		}
		m.flash("Finished the last surah of this recitation", false)
		m.now = nil
	default:
		m.now = nil
	}
	m.refreshAll()
}

func clock(sec float64) string {
	s := int(sec)
	if s >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", s/3600, s/60%60, s%60)
	}
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}

func compact(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1e6)
	case n >= 1000:
		return fmt.Sprintf("%dk", n/1000)
	}
	return fmt.Sprint(n)
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return fmt.Sprintf("%d %ss", n, word)
}

var _ = strings.TrimSpace
