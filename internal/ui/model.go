// Package ui is qari's terminal interface (Bubble Tea).
package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/desktop"
	"github.com/3mro777o/qari/internal/download"
	"github.com/3mro777o/qari/internal/norm"
	"github.com/3mro777o/qari/internal/player"
	"github.com/3mro777o/qari/internal/popular"
	"github.com/3mro777o/qari/internal/store"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Options struct {
	Client    *api.Client
	Lang      string
	Store     *store.Store
	Player    *player.Player // nil when mpv is missing
	Downloads *download.Manager
	Desktop   *desktop.Desktop // nil without a session bus
	Popular   map[int]popular.Entry
	NoArabic  bool
	Theme     string // overrides the saved theme
	Version   string
	Catalog   *api.Catalog // preloaded (tests, direct commands)
	Start     *Start       // play this right away (qari play, radio, random)
	Headless  bool         // --no-ui: one status line, quit when playback stops
}

// Pages: the six tabs, then the two detail pages.
const (
	pHome = iota
	pQaris
	pSurahs
	pRadio
	pDownloads
	pHistory
	pQari
	pSurah
)

var tabNames = []string{"Home", "Qaris", "Surahs", "Radio", "Downloads", "History"}

const (
	ovNone = iota
	ovHelp
	ovRiwayah
	ovRecording
	ovTheme
	ovRandom // recipe list
	ovEdit   // recipe editor
	ovField  // multi-select for one editor field
	ovAyahs  // ayah repeat
)

const (
	playNone = iota
	playSurah
	playRadio
)

var (
	speeds = []float64{0.75, 1, 1.25, 1.5, 1.75, 2}
	sleeps = []int{0, 15, 30, 45, 60, -1} // minutes; -1 = end of surah
)

type listState struct {
	cursor, offset int
	query          string
}

// clamp keeps the cursor inside n items and visible in a window of vis rows.
func (l *listState) clamp(n, vis int) {
	l.cursor = max(min(l.cursor, n-1), 0)
	vis = max(vis, 1)
	l.offset = max(min(l.offset, l.cursor, n-vis), l.cursor-vis+1, 0)
}

type Model struct {
	Options
	w, h      int
	th        Theme
	themeWas  Theme // restored when the theme picker is cancelled
	loading   bool
	loadErr   error
	cat       *api.Catalog
	radioFold []string
	riwayat   []riwayaOption
	spin      spinner.Model
	input     textinput.Model
	searching bool

	page, tab, prev int // prev: where the qari page returns to
	qari            *api.ReciterEntry
	moshaf          api.Moshaf
	surah           int
	lists           [8]listState
	ov              int
	pick            listState
	helpOff         int

	kind      int
	track     store.Track
	radio     api.Radio
	pos, dur  float64
	paused    bool
	buffering bool
	seq       int
	vol       float64
	muted     bool
	speed     int
	sleep     int
	sleepEnd  time.Time
	sleepSeq  int
	lastSave  time.Time

	toast          string
	toastID        int
	confirm, armed string // a destructive or large action waits for a second press
	drag           bool   // dragging on the progress bar
	seekX0, seekX1 int

	recipe   *store.Recipe   // active random recipe, nil when off
	shufHist []store.Entry   // what the recipe played, for p
	shufAt   int             // index of what plays now in shufHist
	shufSeen map[string]bool // reciter/surah or station picked since the pool was last exhausted
	draft    store.Recipe    // recipe being edited
	draftIdx int             // its index in Store.Recipes, -1 when new
	field    int             // editor field open in ovField
	fieldQ   string          // search in ovField
	editCur  int             // editor cursor while a field is open
	typedWas string          // the recipe name before typing, restored by esc

	reads     []api.TimingRead            // recordings with ayah timings, nil until first needed
	timings   map[[2]int][]api.AyahTiming // read id, surah -> ayah timings
	rep       *repeat                     // ayah repeat in progress
	abFrom    int                         // ayah marked by i, waiting for the end mark
	abEach    int                         // times each ayah plays in an A-B repeat
	gapSeq    int                         // pause between repeats in progress
	hifz      store.Hifz                  // values in the repeat popup
	hifzTrack store.Track                 // the surah the popup repeats
	hifzT     []api.AyahTiming            // its timings
	hifzTyped bool                        // a digit was just typed into the popup
}

type (
	catalogMsg struct {
		cat *api.Catalog
		err error
	}
	playedMsg struct {
		seq int
		err error
	}
	toastMsg int
	sleepMsg int
	dlMsg    struct{}
)

func New(o Options) Model {
	m := Model{Options: o, w: 100, h: 30, vol: 100, speed: 1, loading: o.Catalog == nil}
	i := themeIndex(o.Theme)
	if i < 0 {
		i = max(themeIndex(o.Store.Settings.Theme), 0)
	}
	m.th = themes[i]
	m.spin = spinner.New(spinner.WithSpinner(spinner.Spinner{Frames: []string{"✦", "·"}, FPS: time.Second / 4}))
	m.input = textinput.New()
	m.input.Prompt, m.input.CharLimit = "", 80
	m.restyle()
	if o.Catalog != nil {
		m.setCatalog(o.Catalog)
	}
	return m
}

func (m *Model) restyle() {
	s := m.th.s
	m.spin.Style, m.input.TextStyle, m.input.PlaceholderStyle, m.input.Cursor.Style = s.AccentB, s.Text, s.Muted, s.Accent
}

func (m *Model) setCatalog(cat *api.Catalog) {
	for _, r := range cat.Reciters {
		e := m.Popular[r.ID]
		r.Listens, r.Known = e.Listens, e.Known
	}
	m.cat, m.loading, m.loadErr = cat, false, nil
	m.radioFold = nil
	for _, r := range cat.Radios {
		m.radioFold = append(m.radioFold, norm.Fold(r.Name))
	}
	m.riwayat = riwayaOptions(cat)
	if _, ok := cat.Riwayat[m.Store.Settings.Rewaya]; !ok {
		m.Store.Settings.Rewaya = 0
	}
}

func wait[T any](ch <-chan T) tea.Cmd {
	return func() tea.Msg {
		if v, ok := <-ch; ok {
			return v
		}
		return nil
	}
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{func() tea.Msg { <-m.Downloads.Changed; return dlMsg{} }}
	if m.loading {
		cmds = append(cmds, m.spin.Tick, m.load())
	}
	if m.Player != nil {
		cmds = append(cmds, wait(m.Player.Events))
	}
	if m.Desktop != nil {
		cmds = append(cmds, wait(m.Desktop.Cmds))
	}
	if m.Start != nil {
		cmds = append(cmds, func() tea.Msg { return startMsg{} })
	}
	return tea.Batch(cmds...)
}

func (m *Model) load() tea.Cmd {
	c, lang := m.Client, m.Lang
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		cat, err := api.LoadCatalog(ctx, c, lang)
		return catalogMsg{cat, err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmd := m.update(msg)
	m.syncDesktop()
	if m.headlessDone() {
		return m, tea.Batch(cmd, tea.Quit)
	}
	return m, cmd
}

func (m *Model) update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = max(msg.Width, 40), max(msg.Height, 12)
		m.clamp()
	case spinner.TickMsg:
		if m.loading || m.buffering {
			var cmd tea.Cmd
			m.spin, cmd = m.spin.Update(msg)
			return cmd
		}
	case catalogMsg:
		m.loading = false
		if m.loadErr = msg.err; msg.err == nil {
			m.setCatalog(msg.cat)
		}
	case startMsg:
		return m.begin()
	case Notice:
		m.toast = string(msg)
		m.toastID++
		id := m.toastID
		return tea.Tick(8*time.Second, func(time.Time) tea.Msg { return toastMsg(id) })
	case toastMsg:
		if int(msg) == m.toastID {
			m.toast = ""
		}
	case sleepMsg:
		return m.sleepTick(int(msg))
	case timingMsg:
		return m.onTimings(msg)
	case gapMsg:
		m.gapEnd(int(msg))
	case playedMsg:
		if msg.seq == m.seq && msg.err != nil {
			m.buffering = false
			return m.flash("Playback failed: " + msg.err.Error())
		}
	case player.Event:
		return tea.Batch(wait(m.Player.Events), m.onPlayer(msg))
	case dlMsg:
		return func() tea.Msg { <-m.Downloads.Changed; return dlMsg{} }
	case desktop.Cmd:
		return tea.Batch(wait(m.Desktop.Cmds), m.onDesktop(msg))
	case tea.MouseMsg:
		return m.onMouse(msg)
	case tea.KeyMsg:
		return m.onKey(msg)
	}
	return nil
}

func (m *Model) flash(s string) tea.Cmd {
	m.toast = s
	m.toastID++
	id := m.toastID
	return tea.Tick(1600*time.Millisecond, func(time.Time) tea.Msg { return toastMsg(id) })
}

func (m *Model) save() {
	m.Store.Save()
	m.lastSave = time.Now()
}

func (m *Model) quit() tea.Cmd {
	m.save()
	return tea.Quit
}

func (m *Model) onPlayer(ev player.Event) tea.Cmd {
	if ev.Err != nil {
		m.buffering = false
		return m.flash(ev.Err.Error())
	}
	if m.kind == playNone {
		return nil
	}
	m.buffering, m.paused = ev.Loading, ev.Paused
	if ev.Pos > 0 || !ev.Loading {
		m.pos = ev.Pos
	}
	if ev.Duration > 0 {
		m.dur = ev.Duration
	}
	if m.kind != playSurah {
		return nil
	}
	if !ev.Loading && ev.Pos > 0 && ev.Duration > 0 {
		m.Store.Update(m.track, ev.Pos, ev.Duration)
		if time.Since(m.lastSave) > 15*time.Second {
			m.save()
		}
	}
	if !ev.Ended {
		if m.rep != nil && !ev.Loading {
			return m.repeatTick(ev.Pos)
		}
		return nil
	}
	m.Store.MarkFinished(m.track)
	m.save()
	if sleeps[m.sleep] < 0 {
		m.sleep, m.kind = 0, playNone
		return m.flash("Sleep timer: stopped after the surah")
	}
	if m.recipe != nil {
		return m.shuffleNext()
	}
	switch mode(repeatModes, m.Store.Settings.Repeat) {
	case 1:
		return m.playSurah(m.track, true)
	case 0:
		if next, ok := m.adjacent(1); ok {
			return m.playSurah(next, true)
		}
	}
	m.kind = playNone
	return m.flash("Surah finished")
}

// playSurah starts tr, from the saved position unless fromStart. A saved
// download is played from disk.
func (m *Model) playSurah(tr store.Track, fromStart bool) tea.Cmd {
	start := 0.0
	if !fromStart {
		start = m.Store.ResumeAt(tr)
	}
	return m.playAt(tr, start)
}

// playAt starts tr at start seconds; it ends any ayah repeat.
func (m *Model) playAt(tr store.Track, start float64) tea.Cmd {
	if m.Player == nil {
		return m.flash("Install mpv to play audio")
	}
	rec := m.cat.Reciter(tr.Reciter)
	mo, ok := findMoshaf(rec, tr.Moshaf)
	if !ok {
		return m.flash("Recording not found")
	}
	m.endRepeat()
	url := mo.URL(tr.Surah)
	if m.Downloads.Has(tr) {
		url = m.Downloads.Path(tr)
	}
	// The cursor follows the playing surah (next, shuffle) when it sat on the
	// previous one, so d and x act on what is playing.
	if r, cmo, n := m.selected(); m.page == pQari && m.kind == playSurah && r != nil &&
		(store.Track{Reciter: r.ID, Moshaf: cmo.ID, Surah: n}) == m.track {
		m.openQari(rec, mo.ID, tr.Surah)
	}
	m.kind, m.track, m.pos, m.dur = playSurah, tr, start, 0
	if p := m.Store.Get(tr); p != nil {
		m.dur = p.Duration
	}
	m.Store.Plays[tr.Reciter]++
	m.Store.Last = &tr
	m.Store.Log(store.Entry{Time: time.Now(), Track: tr, Recipe: m.recipeName()})
	m.save()
	return m.play(url, start, fmt.Sprintf("%03d %s - %s", tr.Surah, m.cat.SurahName(tr.Surah), rec.Name), speeds[m.speed])
}

func (m *Model) playRadio(r api.Radio) tea.Cmd {
	if m.Player == nil {
		return m.flash("Install mpv to play audio")
	}
	m.endRepeat()
	m.kind, m.radio, m.pos, m.dur = playRadio, r, 0, 0
	m.Store.Log(store.Entry{Time: time.Now(), Radio: r.ID, Recipe: m.recipeName()})
	m.save()
	return m.play(r.URL, 0, r.Name, 1)
}

func (m *Model) play(url string, start float64, title string, speed float64) tea.Cmd {
	m.seq++
	m.paused, m.buffering = false, true
	p, seq, mute := m.Player, m.seq, m.muted
	return tea.Batch(m.spin.Tick, func() tea.Msg {
		err := p.Play(url, start, title)
		if err == nil {
			p.SetSpeed(speed)
			p.SetMute(mute)
			p.AddVolume(0)
		}
		return playedMsg{seq, err}
	})
}

// adjacent is the next (d=1) or previous (d=-1) surah of the playing recording.
func (m *Model) adjacent(d int) (store.Track, bool) {
	mo, ok := findMoshaf(m.cat.Reciter(m.track.Reciter), m.track.Moshaf)
	if !ok {
		return store.Track{}, false
	}
	ids := surahIDs(mo)
	for i, id := range ids {
		if id == m.track.Surah && i+d >= 0 && i+d < len(ids) {
			next := m.track
			next.Surah = ids[i+d]
			return next, true
		}
	}
	return store.Track{}, false
}

func (m *Model) sleepTick(id int) tea.Cmd {
	if id != m.sleepSeq || sleeps[m.sleep] <= 0 {
		return nil
	}
	if time.Now().Before(m.sleepEnd) {
		return m.sleepCmd()
	}
	m.sleep = 0
	if m.kind != playNone && !m.paused {
		m.Player.TogglePause()
		m.paused = true
	}
	m.save()
	return m.flash("Sleep timer paused playback")
}

func (m *Model) sleepCmd() tea.Cmd {
	id := m.sleepSeq
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return sleepMsg(id) })
}

func (m *Model) onDesktop(c desktop.Cmd) tea.Cmd {
	switch c.Name {
	case "playpause":
		return m.togglePause()
	case "play", "pause":
		if (c.Name == "play") == (m.kind == playNone || m.paused) {
			return m.togglePause()
		}
	case "next", "previous":
		return m.step(map[string]int{"next": 1, "previous": -1}[c.Name])
	case "stop":
		m.stop()
	case "seek":
		return m.seekBy(c.Arg)
	case "setpos":
		if m.kind == playSurah {
			m.Player.SeekTo(c.Arg)
			m.pos = c.Arg
		}
	case "quit":
		return m.quit()
	}
	return nil
}

func (m *Model) stop() {
	if m.kind != playNone {
		m.endRepeat()
		m.Player.Stop()
		m.save()
		m.kind, m.paused, m.buffering, m.pos, m.dur = playNone, false, false, 0, 0
	}
}

// syncDesktop tells media keys and bar widgets what is playing.
func (m *Model) syncDesktop() {
	if m.Desktop == nil {
		return
	}
	st := desktop.State{Status: "Stopped", Rate: speeds[m.speed], Volume: m.vol / 100}
	if m.muted {
		st.Volume = 0
	}
	switch m.kind {
	case playSurah:
		rec := m.cat.Reciter(m.track.Reciter)
		mo, _ := findMoshaf(rec, m.track.Moshaf)
		st.Title = fmt.Sprintf("%s (%d)", m.cat.SurahName(m.track.Surah), m.track.Surah)
		st.Artist, st.Album, st.Length = rec.Name, label(m.cat, mo, false), m.dur
		st.CanNext, st.CanPrev = true, true
	case playRadio:
		st.Title, st.Artist, st.Rate = m.radio.Name, "Live Quran radio", 1
	}
	if m.kind != playNone {
		st.Status, st.Pos = "Playing", m.pos
		if m.paused {
			st.Status = "Paused"
		}
	}
	m.Desktop.Set(st)
}

// homeItem is a Home row: a surah to continue (prog set) or a qari.
type homeItem struct {
	prog  *store.Progress
	rec   *api.ReciterEntry
	radio *api.Radio
	kind  int // 0 continue, 1 favorite (qari or radio), 2 most listened
}

// homeSel is the selected Home row, or nil off Home.
func (m *Model) homeSel() *homeItem {
	if m.page != pHome || m.ov != ovNone {
		return nil
	}
	if items, i := m.homeItems(), m.list().cursor; i < len(items) {
		return &items[i]
	}
	return nil
}

func (m *Model) homeItems() []homeItem {
	q := m.lists[pHome].query
	var out []homeItem
	recent := m.Store.Recent(5)
	if len(recent) == 0 && m.Store.Last != nil {
		recent = []store.Progress{{Track: *m.Store.Last}}
		if p := m.Store.Get(*m.Store.Last); p != nil {
			recent[0] = *p
		}
	}
	for i := range recent {
		p := &recent[i]
		if r := m.cat.Reciter(p.Reciter); r != nil && (norm.Match(r.Search, q) || len(m.surahs([]int{p.Surah}, q)) > 0) {
			out = append(out, homeItem{p, r, nil, 0})
		}
	}
	all := m.qaris(q)
	for _, r := range all {
		if m.Store.Favorites[r.ID] {
			out = append(out, homeItem{nil, r, nil, 1})
		}
	}
	for i := range m.cat.Radios {
		if r := &m.cat.Radios[i]; m.Store.FavRadios[r.ID] && norm.Match(m.radioFold[i], q) {
			out = append(out, homeItem{nil, nil, r, 1})
		}
	}
	limit := max(16, m.h-9)
	for i, r := range all {
		if i < limit {
			out = append(out, homeItem{nil, r, nil, 2})
		}
	}
	return out
}

// dlItem is a Downloads row: a queued job or a saved recording.
type dlItem struct {
	job   *download.Job
	group *download.Group
}

func (m *Model) dlItems() []dlItem {
	var out []dlItem
	for _, j := range m.Downloads.Jobs() {
		out = append(out, dlItem{job: &j})
	}
	for _, g := range m.Downloads.Saved() {
		if m.cat.Reciter(g.Reciter) != nil {
			out = append(out, dlItem{group: &g})
		}
	}
	return out
}

func (m *Model) radios() []api.Radio {
	var out []api.Radio
	for i, r := range m.cat.Radios {
		if norm.Match(m.radioFold[i], m.lists[pRadio].query) {
			out = append(out, r)
		}
	}
	return out
}

// count is the number of rows in the current page's list.
func (m *Model) count() int {
	q := m.list().query
	switch m.page {
	case pHome:
		return len(m.homeItems())
	case pQaris:
		return len(m.qaris(q))
	case pSurahs:
		return len(m.surahs(nil, q))
	case pRadio:
		return len(m.radios())
	case pDownloads:
		return len(m.dlItems())
	case pHistory:
		return len(m.histItems())
	case pQari:
		return len(m.surahs(surahIDs(m.moshaf), q))
	case pSurah:
		return len(m.qarisFor(m.surah, q))
	}
	return 0
}

func (m *Model) list() *listState { return &m.lists[m.page] }

func (m *Model) dockH() int { return 3 + b2i(m.h >= 20) }

// gap is the blank line between the list and the player, when there is room.
func (m *Model) gap() int { return b2i(m.h >= 20) }

// listRows is how many list rows fit under the page's heading lines.
func (m *Model) listRows() int {
	return max(1, m.h-2-m.dockH()-m.gap()-2-b2i(m.page >= pQari))
}

func (m *Model) clamp() {
	if m.cat != nil {
		m.list().clamp(m.count(), m.listRows())
	}
}

// selected is what the cursor points at: a qari, its recording, and a
// surah when the row is one surah.
func (m *Model) selected() (rec *api.ReciterEntry, mo api.Moshaf, surah int) {
	i, q := m.list().cursor, m.list().query
	switch m.page {
	case pHome:
		if items := m.homeItems(); i < len(items) {
			if it := items[i]; it.prog != nil {
				mo, _ = findMoshaf(it.rec, it.prog.Moshaf)
				return it.rec, mo, it.prog.Surah
			}
			rec = items[i].rec
		}
	case pQaris:
		if items := m.qaris(q); i < len(items) {
			rec = items[i]
		}
	case pDownloads:
		if items := m.dlItems(); i < len(items) {
			t := store.Track{}
			if it := items[i]; it.job != nil {
				t = it.job.Track
			} else {
				t = store.Track{Reciter: it.group.Reciter, Moshaf: it.group.Moshaf}
			}
			rec = m.cat.Reciter(t.Reciter)
			mo, _ = findMoshaf(rec, t.Moshaf)
			return rec, mo, t.Surah
		}
	case pHistory:
		if items := m.histItems(); i < len(items) && items[i].Radio == 0 {
			rec = m.cat.Reciter(items[i].Reciter)
			mo, _ = findMoshaf(rec, items[i].Moshaf)
			return rec, mo, items[i].Surah
		}
		return nil, mo, 0
	case pQari:
		if items := m.surahs(surahIDs(m.moshaf), q); i < len(items) {
			return m.qari, m.moshaf, items[i].ID
		}
		return m.qari, m.moshaf, 0
	case pSurah:
		if items := m.qarisFor(m.surah, q); i < len(items) {
			return items[i].r, items[i].mo, m.surah
		}
	}
	if rec != nil {
		mo, _ = best(m.cat, rec, m.Store.Settings.Rewaya, 0, false)
	}
	return rec, mo, 0
}

func (m *Model) openQari(rec *api.ReciterEntry, moshafID, focus int) {
	if m.page != pQari {
		m.prev = m.page
	}
	mo, ok := findMoshaf(rec, moshafID)
	if !ok {
		mo, _ = best(m.cat, rec, m.Store.Settings.Rewaya, focus, false)
	}
	m.qari, m.moshaf, m.page = rec, mo, pQari
	m.lists[pQari] = listState{}
	for i, s := range m.surahs(surahIDs(mo), "") {
		if s.ID == focus {
			m.lists[pQari].cursor = i
		}
	}
	m.clamp()
}

func (m *Model) goTab(t int) {
	m.tab, m.page = t, t
	m.clamp()
}

func (m *Model) back() {
	switch {
	case m.page == pQari && m.prev == pSurah:
		m.page, m.prev = pSurah, m.tab
	case m.page >= pQari:
		m.page = m.tab
	}
	m.clamp()
}
