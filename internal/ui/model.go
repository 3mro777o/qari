package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/norm"
	"github.com/3mro777o/qari/internal/player"
	"github.com/3mro777o/qari/internal/popular"
	"github.com/3mro777o/qari/internal/store"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Options configures the qari TUI model.
type Options struct {
	Client   *api.Client
	Lang     string
	Store    *store.Store
	Player   *player.Player
	Popular  map[int]popular.Entry
	NoArabic bool
	Theme    string
	Version  string
	Catalog  *api.Catalog // optional preloaded catalog (for tests)
}

type tabKind int

const (
	tabHome tabKind = iota
	tabQaris
	tabSurahs
	tabRadio
)

type viewKind int

const (
	viewTabs viewKind = iota
	viewQariDetail
	viewSurahDetail
)

type overlayKind int

const (
	overlayNone overlayKind = iota
	overlayHelp
	overlayRiwayah
	overlayRecording
)

type playKind int

const (
	playNone playKind = iota
	playSurah
	playRadio
)

type sleepMode int

const (
	sleepOff sleepMode = iota
	sleep15
	sleep30
	sleep45
	sleep60
	sleepEndOfSurah
)

var speedSteps = []float64{0.75, 1.0, 1.25, 1.5, 1.75, 2.0}

// listState tracks cursor, scroll offset, and search query for a single list view.
type listState struct {
	cursor int
	offset int
	query  string
}

func (l *listState) clamp(n, visible int) {
	if n <= 0 {
		l.cursor, l.offset = 0, 0
		return
	}
	if l.cursor < 0 {
		l.cursor = 0
	}
	if l.cursor >= n {
		l.cursor = n - 1
	}
	if visible <= 0 {
		visible = 1
	}
	if l.cursor < l.offset {
		l.offset = l.cursor
	}
	if l.cursor >= l.offset+visible {
		l.offset = l.cursor - visible + 1
	}
	maxOff := max(0, n-visible)
	if l.offset > maxOff {
		l.offset = maxOff
	}
	if l.offset < 0 {
		l.offset = 0
	}
}

func (l *listState) move(delta, n, visible int) {
	if n <= 0 {
		return
	}
	l.cursor += delta
	l.clamp(n, visible)
}

// homeItem represents a single selectable row on the Home screen.
type homeItemKind int

const (
	homeContinue homeItemKind = iota
	homeFavorite
	homePopular
)

type homeItem struct {
	kind     homeItemKind
	progress store.Progress
	reciter  *api.ReciterEntry
}

// Model is the root Bubble Tea model for qari.
type Model struct {
	client   *api.Client
	lang     string
	store    *store.Store
	player   *player.Player
	popular  map[int]popular.Entry
	version  string
	noArabic bool

	// Terminal dimensions
	width  int
	height int

	// Theme
	themeIdx int
	theme    Theme

	// Loading / catalog state
	loading     bool
	loadErr     error
	cat         *api.Catalog
	radioFolded []string
	riwayaOpts  []riwayaOption
	spinner     spinner.Model

	// Navigation
	tab     tabKind
	view    viewKind
	overlay overlayKind

	// Back-navigation context
	returnTab     tabKind
	returnView    viewKind
	returnSurahID int

	// Detail views
	detailReciter *api.ReciterEntry
	detailMoshaf  api.Moshaf
	detailSurahID int

	// Lists
	homeList        listState
	qarisList       listState
	surahsList      listState
	radioList       listState
	qariDetailList  listState
	surahDetailList listState

	// Overlays
	helpOffset      int
	riwayaCursor    int
	riwayaOffset    int
	recordingCursor int
	recordingOffset int

	// Search input
	searching   bool
	searchInput textinput.Model

	// Playback
	playKind    playKind
	playTrack   store.Track
	playRadio   api.Radio
	playPos     float64
	playDur     float64
	paused      bool
	playLoading bool
	playErr     string
	playSeq     int
	volume      float64
	muted       bool
	speedIdx    int
	lastSave    time.Time

	// Sleep timer
	sleepMode  sleepMode
	sleepUntil time.Time

	// Status toast
	toast   string
	toastID int
}

// Async messages
type catalogLoadedMsg struct {
	cat *api.Catalog
	err error
}

type playerEventMsg struct {
	ev player.Event
}

type playStartedMsg struct {
	seq int
	err error
}

type toastClearMsg struct {
	id int
}

type sleepTickMsg struct {
	now time.Time
}

// New constructs a new qari TUI model.
func New(opts Options) Model {
	st := opts.Store
	if st == nil {
		st = &store.Store{}
		st.Favorites = map[int]bool{}
		st.Progress = map[string]*store.Progress{}
		st.Plays = map[int]int{}
	}
	st.Settings.Sort = normalizeSort(st.Settings.Sort)
	st.Settings.Repeat = normalizeRepeat(st.Settings.Repeat)

	// Theme precedence: explicit CLI/env > saved setting > default (mihrab)
	themeIdx := 0
	if opts.Theme != "" {
		if idx, ok := themeIndexByName(opts.Theme); ok {
			themeIdx = idx
		}
	} else if st.Settings.Theme != "" {
		if idx, ok := themeIndexByName(st.Settings.Theme); ok {
			themeIdx = idx
		}
	}

	noArabic := opts.NoArabic || st.Settings.NoArabic

	sp := spinner.New()
	sp.Spinner = spinner.Spinner{
		Frames: []string{"✦", "·", "✦", "·"},
		FPS:    time.Second / 4,
	}

	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 80

	m := Model{
		client:      opts.Client,
		lang:        opts.Lang,
		store:       st,
		player:      opts.Player,
		popular:     opts.Popular,
		version:     opts.Version,
		noArabic:    noArabic,
		width:       100,
		height:      30,
		themeIdx:    themeIdx,
		theme:       themeByIndex(themeIdx),
		loading:     opts.Catalog == nil,
		spinner:     sp,
		searchInput: ti,
		volume:      100,
		speedIdx:    1, // 1.0x
	}
	m.applyThemeStyles()

	if opts.Catalog != nil {
		m.setCatalog(opts.Catalog)
	}
	return m
}

func (m *Model) applyThemeStyles() {
	m.spinner.Style = m.theme.Styles.AccentBold
	m.searchInput.TextStyle = m.theme.Styles.Text
	m.searchInput.PlaceholderStyle = m.theme.Styles.Muted
	m.searchInput.Cursor.Style = m.theme.Styles.Accent
}

func (m *Model) setCatalog(cat *api.Catalog) {
	applyPopularity(cat, m.popular)
	m.cat = cat
	m.loading = false
	m.loadErr = nil
	m.radioFolded = make([]string, len(cat.Radios))
	for i, r := range cat.Radios {
		m.radioFolded[i] = norm.Fold(r.Name)
	}
	m.riwayaOpts = buildRiwayaOptions(cat)
	// Clear saved riwayah if it no longer exists in this catalog.
	if m.store.Settings.Rewaya != 0 {
		found := false
		for _, o := range m.riwayaOpts {
			if o.ID == m.store.Settings.Rewaya {
				found = true
				break
			}
		}
		if !found {
			m.store.Settings.Rewaya = 0
		}
	}
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	if m.loading && m.client != nil {
		cmds = append(cmds, m.spinner.Tick, loadCatalogCmd(m.client, m.lang))
	}
	if m.player != nil {
		cmds = append(cmds, waitPlayerEventCmd(m.player.Events))
	}
	return tea.Batch(cmds...)
}

func loadCatalogCmd(c *api.Client, lang string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		cat, err := api.LoadCatalog(ctx, c, lang)
		return catalogLoadedMsg{cat: cat, err: err}
	}
}

func waitPlayerEventCmd(ch <-chan player.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return nil
		}
		return playerEventMsg{ev: ev}
	}
}

func sleepTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return sleepTickMsg{now: t}
	})
}

func (m *Model) setToast(msg string) tea.Cmd {
	m.toast = msg
	m.toastID++
	id := m.toastID
	return tea.Tick(1600*time.Millisecond, func(time.Time) tea.Msg {
		return toastClearMsg{id: id}
	})
}

func (m *Model) saveStore() {
	if m.store != nil {
		_ = m.store.Save()
		m.lastSave = time.Now()
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = max(msg.Width, 40)
		m.height = max(msg.Height, 12)
		m.clampActiveList()
		return m, nil

	case spinner.TickMsg:
		if m.loading || m.playLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case catalogLoadedMsg:
		if msg.err != nil {
			m.loading = false
			m.loadErr = msg.err
			return m, nil
		}
		m.setCatalog(msg.cat)
		return m, nil

	case toastClearMsg:
		if msg.id == m.toastID {
			m.toast = ""
		}
		return m, nil

	case sleepTickMsg:
		if m.sleepMode == sleepOff || m.sleepMode == sleepEndOfSurah {
			return m, nil
		}
		if !msg.now.Before(m.sleepUntil) {
			m.sleepMode = sleepOff
			m.sleepUntil = time.Time{}
			if m.playKind != playNone && !m.paused && m.player != nil {
				m.player.TogglePause()
				m.paused = true
			}
			m.saveStore()
			return m, m.setToast("Sleep timer paused playback")
		}
		return m, sleepTickCmd()

	case playStartedMsg:
		if msg.seq != m.playSeq {
			return m, nil
		}
		if msg.err != nil {
			m.playLoading = false
			m.playErr = msg.err.Error()
			return m, m.setToast("Playback failed: " + msg.err.Error())
		}
		return m, nil

	case playerEventMsg:
		var cmds []tea.Cmd
		if m.player != nil {
			cmds = append(cmds, waitPlayerEventCmd(m.player.Events))
		}
		if cmd := m.handlePlayerEvent(msg.ev); cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	if m.searching {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *Model) handlePlayerEvent(ev player.Event) tea.Cmd {
	if ev.Err != nil {
		m.playLoading = false
		m.playErr = ev.Err.Error()
		return m.setToast(ev.Err.Error())
	}
	if m.playKind == playNone {
		return nil
	}
	m.playErr = ""
	m.playLoading = ev.Loading
	m.paused = ev.Paused
	if ev.Pos > 0 || !ev.Loading {
		m.playPos = ev.Pos
	}
	if ev.Duration > 0 {
		m.playDur = ev.Duration
	}

	if m.playKind == playSurah && !ev.Loading && ev.Pos > 0 && ev.Duration > 0 {
		m.store.Update(m.playTrack, ev.Pos, ev.Duration)
		if time.Since(m.lastSave) >= 15*time.Second {
			m.saveStore()
		}
	}

	if ev.Ended && m.playKind == playSurah {
		m.store.MarkFinished(m.playTrack)
		m.saveStore()

		if m.sleepMode == sleepEndOfSurah {
			m.sleepMode = sleepOff
			m.playKind = playNone
			m.playLoading = false
			m.paused = false
			return m.setToast("Sleep timer: stopped after surah")
		}

		switch normalizeRepeat(m.store.Settings.Repeat) {
		case repeatOne:
			return m.startSurah(m.playTrack, true)
		case repeatNext:
			if next, ok := m.adjacentSurah(m.playTrack, +1); ok {
				return m.startSurah(next, true)
			}
			m.playKind = playNone
			m.playLoading = false
			return m.setToast("Finished recording")
		default: // repeatOff
			m.playKind = playNone
			m.playLoading = false
			return m.setToast("Surah finished")
		}
	}
	return nil
}

// startSurah begins playing track; if fromStart is true, starts at 0 instead of ResumeAt.
func (m *Model) startSurah(tr store.Track, fromStart bool) tea.Cmd {
	if m.player == nil {
		return m.setToast("Install mpv to play audio")
	}
	if m.cat == nil {
		return nil
	}
	rec := m.cat.Reciter(tr.Reciter)
	if rec == nil {
		return m.setToast("Qari not found in catalog")
	}
	moshaf, ok := findMoshafByID(rec, tr.Moshaf)
	if !ok {
		return m.setToast("Recording not found")
	}
	startAt := 0.0
	if !fromStart {
		startAt = m.store.ResumeAt(tr)
	}
	url := moshaf.URL(tr.Surah)
	title := fmt.Sprintf("%03d %s - %s", tr.Surah, m.cat.SurahName(tr.Surah), rec.Name)

	m.playKind = playSurah
	m.playTrack = tr
	m.playPos = startAt
	if p := m.store.Get(tr); p != nil && p.Duration > 0 {
		m.playDur = p.Duration
	} else {
		m.playDur = 0
	}
	m.paused = false
	m.playLoading = true
	m.playErr = ""
	m.playSeq++
	seq := m.playSeq

	m.store.Plays[tr.Reciter]++
	m.store.Last = &tr
	m.saveStore()

	p := m.player
	speed := speedSteps[m.speedIdx]
	muted := m.muted
	vol := m.volume

	playCmd := func() tea.Msg {
		err := p.Play(url, startAt, title)
		if err == nil {
			if speed != 1.0 {
				p.SetSpeed(speed)
			}
			if muted {
				p.SetMute(true)
			}
			if vol != 100 {
				p.AddVolume(0) // re-applies stored volume to newly started mpv
			}
		}
		return playStartedMsg{seq: seq, err: err}
	}
	return tea.Batch(playCmd, m.spinner.Tick)
}

func (m *Model) startRadio(r api.Radio) tea.Cmd {
	if m.player == nil {
		return m.setToast("Install mpv to play audio")
	}
	m.playKind = playRadio
	m.playRadio = r
	m.playPos = 0
	m.playDur = 0
	m.paused = false
	m.playLoading = true
	m.playErr = ""
	m.playSeq++
	seq := m.playSeq

	p := m.player
	muted := m.muted
	vol := m.volume

	playCmd := func() tea.Msg {
		err := p.Play(r.URL, 0, r.Name)
		if err == nil {
			p.SetSpeed(1.0)
			if muted {
				p.SetMute(true)
			}
			if vol != 100 {
				p.AddVolume(0)
			}
		}
		return playStartedMsg{seq: seq, err: err}
	}
	return tea.Batch(playCmd, m.spinner.Tick)
}

func (m *Model) adjacentSurah(tr store.Track, delta int) (store.Track, bool) {
	if m.cat == nil {
		return store.Track{}, false
	}
	rec := m.cat.Reciter(tr.Reciter)
	if rec == nil {
		return store.Track{}, false
	}
	moshaf, ok := findMoshafByID(rec, tr.Moshaf)
	if !ok {
		return store.Track{}, false
	}
	ids := moshafSortedSurahs(moshaf)
	for i, id := range ids {
		if id == tr.Surah {
			ni := i + delta
			if ni >= 0 && ni < len(ids) {
				next := tr
				next.Surah = ids[ni]
				return next, true
			}
			return store.Track{}, false
		}
	}
	return store.Track{}, false
}

func (m *Model) activeList() *listState {
	switch m.view {
	case viewQariDetail:
		return &m.qariDetailList
	case viewSurahDetail:
		return &m.surahDetailList
	default:
		switch m.tab {
		case tabHome:
			return &m.homeList
		case tabQaris:
			return &m.qarisList
		case tabSurahs:
			return &m.surahsList
		case tabRadio:
			return &m.radioList
		}
	}
	return &m.homeList
}

func (m *Model) homeItems() []homeItem {
	if m.cat == nil {
		return nil
	}
	q := strings.TrimSpace(m.homeList.query)
	var items []homeItem

	// 1. Continue Listening (unfinished tracks, newest first; fallback to Last if present)
	recent := m.store.Recent(5)
	if len(recent) == 0 && m.store.Last != nil {
		if p := m.store.Get(*m.store.Last); p != nil {
			recent = []store.Progress{*p}
		} else {
			recent = []store.Progress{{Track: *m.store.Last}}
		}
	}
	for _, p := range recent {
		rec := m.cat.Reciter(p.Reciter)
		if rec == nil {
			continue
		}
		if q != "" {
			s := api.SurahEntry{}
			if p.Surah >= 1 && p.Surah <= len(m.cat.Surahs) {
				s = m.cat.Surahs[p.Surah-1]
			}
			if !norm.Match(rec.Search, q) && !matchSurah(s, q) {
				continue
			}
		}
		items = append(items, homeItem{kind: homeContinue, progress: p, reciter: rec})
	}

	// 2. Favorite Qaris
	var favs []*api.ReciterEntry
	for _, r := range m.cat.Reciters {
		if !m.store.Favorites[r.ID] {
			continue
		}
		if q != "" && !norm.Match(r.Search, q) {
			continue
		}
		favs = append(favs, r)
	}
	sortQaris(favs, sortPopular)
	for _, r := range favs {
		items = append(items, homeItem{kind: homeFavorite, reciter: r})
	}

	// 3. Most Listened Qaris (curated complete qaris, or filtered if searching)
	pop := filterQaris(m.cat, m.store.Favorites, m.store.Settings.ShowAll, m.store.Settings.Rewaya, sortPopular, q)
	limit := max(16, m.canvasHeight()-3)
	if q != "" {
		limit = max(30, limit)
	}
	if len(pop) > limit {
		pop = pop[:limit]
	}
	for _, r := range pop {
		items = append(items, homeItem{kind: homePopular, reciter: r})
	}
	return items
}

func (m *Model) filteredQaris() []*api.ReciterEntry {
	return filterQaris(m.cat, m.store.Favorites, m.store.Settings.ShowAll, m.store.Settings.Rewaya, m.store.Settings.Sort, m.qarisList.query)
}

func (m *Model) filteredSurahs() []api.SurahEntry {
	return filterSurahs(m.cat, m.surahsList.query)
}

func (m *Model) filteredRadios() []api.Radio {
	if m.cat == nil {
		return nil
	}
	q := m.radioList.query
	if strings.TrimSpace(q) == "" {
		return m.cat.Radios
	}
	var out []api.Radio
	for i, r := range m.cat.Radios {
		folded := ""
		if i < len(m.radioFolded) {
			folded = m.radioFolded[i]
		} else {
			folded = norm.Fold(r.Name)
		}
		if norm.Match(folded, q) {
			out = append(out, r)
		}
	}
	return out
}

func (m *Model) filteredQariDetailSurahs() []api.SurahEntry {
	if m.cat == nil || m.detailReciter == nil {
		return nil
	}
	return filterMoshafSurahs(m.cat, m.detailMoshaf, m.qariDetailList.query)
}

func (m *Model) filteredSurahDetailQaris() []surahQariItem {
	return filterQarisForSurah(m.cat, m.store.Favorites, m.detailSurahID, m.store.Settings.ShowAll, m.store.Settings.Rewaya, m.store.Settings.Sort, m.surahDetailList.query)
}

func (m *Model) activeListCount() int {
	switch m.view {
	case viewQariDetail:
		return len(m.filteredQariDetailSurahs())
	case viewSurahDetail:
		return len(m.filteredSurahDetailQaris())
	default:
		switch m.tab {
		case tabHome:
			return len(m.homeItems())
		case tabQaris:
			return len(m.filteredQaris())
		case tabSurahs:
			return len(m.filteredSurahs())
		case tabRadio:
			return len(m.filteredRadios())
		}
	}
	return 0
}

func (m *Model) dockHeight() int {
	if m.height < 20 {
		return 3
	}
	return 4
}

func (m *Model) canvasHeight() int {
	// Header = 2 lines, Dock = 3 or 4 lines.
	return max(4, m.height-2-m.dockHeight())
}

func (m *Model) activeVisibleRows() int {
	ch := m.canvasHeight()
	switch m.view {
	case viewQariDetail:
		return max(1, ch-3)
	case viewSurahDetail:
		return max(1, ch-3)
	default:
		switch m.tab {
		case tabHome:
			return max(1, ch-2)
		case tabQaris, tabSurahs, tabRadio:
			return max(1, ch-2)
		}
	}
	return max(1, ch-2)
}

func (m *Model) clampActiveList() {
	m.activeList().clamp(m.activeListCount(), m.activeVisibleRows())
}

func (m *Model) openQari(rec *api.ReciterEntry, preferredMoshafID int, focusSurah int) {
	if rec == nil || m.cat == nil {
		return
	}
	if m.view == viewTabs {
		m.returnTab = m.tab
		m.returnView = viewTabs
	} else if m.view == viewSurahDetail {
		m.returnView = viewSurahDetail
		m.returnSurahID = m.detailSurahID
	}

	var moshaf api.Moshaf
	var ok bool
	if preferredMoshafID > 0 {
		moshaf, ok = findMoshafByID(rec, preferredMoshafID)
	}
	if !ok {
		moshaf, ok = bestMoshaf(m.cat, rec, m.store.Settings.Rewaya, focusSurah, false)
	}
	if !ok && len(rec.Moshaf) > 0 {
		moshaf = rec.Moshaf[0]
	}

	m.detailReciter = rec
	m.detailMoshaf = moshaf
	m.view = viewQariDetail
	m.qariDetailList = listState{}

	if focusSurah > 0 {
		surahs := m.filteredQariDetailSurahs()
		for i, s := range surahs {
			if s.ID == focusSurah {
				m.qariDetailList.cursor = i
				break
			}
		}
		m.qariDetailList.clamp(len(surahs), m.activeVisibleRows())
	}
}

func (m *Model) openSurah(surahID int) {
	if m.cat == nil || surahID < 1 || surahID > 114 {
		return
	}
	m.returnTab = m.tab
	m.returnView = viewTabs
	m.detailSurahID = surahID
	m.view = viewSurahDetail
	m.surahDetailList = listState{}
}

func (m *Model) selectedReciterForFavorite() *api.ReciterEntry {
	switch m.view {
	case viewQariDetail:
		return m.detailReciter
	case viewSurahDetail:
		items := m.filteredSurahDetailQaris()
		if m.surahDetailList.cursor >= 0 && m.surahDetailList.cursor < len(items) {
			return items[m.surahDetailList.cursor].Reciter
		}
	default:
		switch m.tab {
		case tabHome:
			items := m.homeItems()
			if m.homeList.cursor >= 0 && m.homeList.cursor < len(items) {
				return items[m.homeList.cursor].reciter
			}
		case tabQaris:
			items := m.filteredQaris()
			if m.qarisList.cursor >= 0 && m.qarisList.cursor < len(items) {
				return items[m.qarisList.cursor]
			}
		}
	}
	return nil
}

func (m *Model) sleepRemainingText() string {
	switch m.sleepMode {
	case sleepOff:
		return ""
	case sleepEndOfSurah:
		return "end of surah"
	default:
		rem := time.Until(m.sleepUntil)
		if rem < 0 {
			rem = 0
		}
		return formatDuration(rem.Seconds())
	}
}

func (m *Model) isOffline() bool {
	return m.client != nil && m.client.Offline()
}
