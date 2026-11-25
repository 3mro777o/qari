package ui

import (
	"fmt"
	"time"

	"github.com/3mro777o/qari/internal/keymap"
	"github.com/3mro777o/qari/internal/store"
	tea "github.com/charmbracelet/bubbletea"
)

// normalizeKey maps single-rune Arabic-layout shortcuts to their US-layout key
// while leaving multi-char/control keys and search input untouched.
func normalizeKey(msg tea.KeyMsg) string {
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && !msg.Alt {
		r := msg.Runes[0]
		// Preserve ASCII ',' and '.' as well as mapped Arabic keys:
		// keymap.LatinKey maps ','->'<' and '.'->'>' (Arabic shift keys),
		// which we intentionally treat identically for 1-minute seek.
		if r == ' ' {
			return " "
		}
		return string(keymap.LatinKey(r))
	}
	return msg.String()
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Ctrl+C always exits cleanly.
	if msg.String() == "ctrl+c" {
		m.saveStore()
		return m, tea.Quit
	}

	// Loading / error states
	if m.loading {
		if normalizeKey(msg) == "q" {
			return m, tea.Quit
		}
		return m, nil
	}
	if m.loadErr != nil {
		switch normalizeKey(msg) {
		case "q":
			return m, tea.Quit
		case "r", "R", "enter":
			if m.client != nil {
				m.loadErr = nil
				m.loading = true
				return m, tea.Batch(m.spinner.Tick, loadCatalogCmd(m.client, m.lang))
			}
		case "c":
			return m.cycleTheme()
		}
		return m, nil
	}

	// Focused search bar gets raw keys (NO LatinKey translation so Arabic search works!)
	if m.searching {
		return m.handleSearchKey(msg)
	}

	// Active modal overlay gets priority.
	if m.overlay != overlayNone {
		return m.handleOverlayKey(msg)
	}

	key := normalizeKey(msg)

	switch key {
	case "q":
		m.saveStore()
		return m, tea.Quit

	case "?":
		m.overlay = overlayHelp
		m.helpOffset = 0
		return m, nil

	case "c":
		return m.cycleTheme()

	case "A":
		m.noArabic = !m.noArabic
		m.store.Settings.NoArabic = m.noArabic
		m.saveStore()
		if m.noArabic {
			return m, m.setToast("Arabic surah names hidden")
		}
		return m, m.setToast("Arabic surah names shown")

	// Tabs
	case "1":
		m.switchTab(tabHome)
		return m, nil
	case "2":
		m.switchTab(tabQaris)
		return m, nil
	case "3":
		m.switchTab(tabSurahs)
		return m, nil
	case "4":
		m.switchTab(tabRadio)
		return m, nil
	case "tab":
		m.switchTab(tabKind((int(m.tab) + 1) % 4))
		return m, nil
	case "shift+tab":
		m.switchTab(tabKind((int(m.tab) + 3) % 4))
		return m, nil

	// Search & Back
	case "/":
		m.searching = true
		m.searchInput.SetValue(m.activeList().query)
		m.searchInput.CursorEnd()
		m.searchInput.Focus()
		return m, nil

	case "esc":
		al := m.activeList()
		if al.query != "" {
			al.query = ""
			al.cursor, al.offset = 0, 0
			return m, m.setToast("Search cleared")
		}
		if m.view == viewQariDetail {
			if m.returnView == viewSurahDetail && m.returnSurahID > 0 {
				m.view = viewSurahDetail
				m.detailSurahID = m.returnSurahID
			} else {
				m.view = viewTabs
				m.tab = m.returnTab
			}
			m.clampActiveList()
			return m, nil
		}
		if m.view == viewSurahDetail {
			m.view = viewTabs
			m.tab = m.returnTab
			m.clampActiveList()
			return m, nil
		}
		return m, nil

	// List navigation
	case "j", "down":
		m.activeList().move(1, m.activeListCount(), m.activeVisibleRows())
		return m, nil
	case "k", "up":
		m.activeList().move(-1, m.activeListCount(), m.activeVisibleRows())
		return m, nil
	case "pgdown", "ctrl+d":
		step := max(1, m.activeVisibleRows()/2)
		m.activeList().move(step, m.activeListCount(), m.activeVisibleRows())
		return m, nil
	case "pgup", "ctrl+u":
		step := max(1, m.activeVisibleRows()/2)
		m.activeList().move(-step, m.activeListCount(), m.activeVisibleRows())
		return m, nil
	case "g", "home":
		al := m.activeList()
		al.cursor, al.offset = 0, 0
		return m, nil
	case "G", "end":
		n := m.activeListCount()
		if n > 0 {
			al := m.activeList()
			al.cursor = n - 1
			al.clamp(n, m.activeVisibleRows())
		}
		return m, nil

	case "enter":
		return m.handleActivate()

	// Library & Filters
	case "f":
		return m.toggleFavorite()
	case "a":
		if m.view == viewTabs && m.tab == tabRadio {
			return m, nil
		}
		return m.toggleShowAll()
	case "s":
		if m.view == viewTabs && m.tab == tabRadio {
			return m, nil
		}
		return m.cycleSort()
	case "w":
		if m.view == viewTabs && m.tab == tabRadio {
			return m, nil
		}
		return m.openRiwayahPicker()
	case "e":
		if m.view == viewQariDetail {
			return m.openRecordingPicker()
		}
		return m, nil
	case "v":
		if m.view == viewQariDetail {
			return m.cycleRecording()
		}
		return m, nil
	case "i":
		if m.view == viewSurahDetail {
			items := m.filteredSurahDetailQaris()
			if m.surahDetailList.cursor >= 0 && m.surahDetailList.cursor < len(items) {
				it := items[m.surahDetailList.cursor]
				m.openQari(it.Reciter, it.Moshaf.ID, m.detailSurahID)
			}
		}
		return m, nil

	// Playback controls
	case " ":
		return m.togglePauseOrResume()
	case "left":
		return m.seekBy(-10)
	case "right":
		return m.seekBy(+10)
	case "shift+left", "<":
		return m.seekBy(-60)
	case "shift+right", ">":
		return m.seekBy(+60)
	case "n":
		return m.stepSurah(+1)
	case "p":
		return m.stepSurah(-1)
	case "+", "=":
		return m.adjustVolume(+5)
	case "-":
		return m.adjustVolume(-5)
	case "m":
		return m.toggleMute()
	case "[":
		return m.adjustSpeed(-1)
	case "]":
		return m.adjustSpeed(+1)
	case "x":
		return m.stopPlayback()
	case "r":
		return m.cycleRepeat()
	case "t":
		return m.cycleSleepTimer()
	case "o":
		return m.jumpToPlaying()
	}

	return m, nil
}

func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.searching = false
		m.searchInput.Blur()
		al := m.activeList()
		al.query = ""
		al.cursor, al.offset = 0, 0
		return m, nil

	case tea.KeyEnter:
		m.searching = false
		m.searchInput.Blur()
		m.clampActiveList()
		return m, nil

	case tea.KeyDown:
		m.activeList().move(1, m.activeListCount(), m.activeVisibleRows())
		return m, nil

	case tea.KeyUp:
		m.activeList().move(-1, m.activeListCount(), m.activeVisibleRows())
		return m, nil
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	al := m.activeList()
	if al.query != m.searchInput.Value() {
		al.query = m.searchInput.Value()
		al.cursor, al.offset = 0, 0
	}
	return m, cmd
}

func (m Model) handleOverlayKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := normalizeKey(msg)

	switch m.overlay {
	case overlayHelp:
		switch key {
		case "esc", "?", "q", "enter":
			m.overlay = overlayNone
		case "j", "down":
			m.helpOffset++
		case "k", "up":
			if m.helpOffset > 0 {
				m.helpOffset--
			}
		case "g", "home":
			m.helpOffset = 0
		case "G", "end":
			m.helpOffset = 999
		}
		return m, nil

	case overlayRiwayah:
		n := len(m.riwayaOpts)
		vis := max(3, min(12, m.height-8))
		switch key {
		case "esc", "w", "q":
			m.overlay = overlayNone
		case "j", "down":
			if m.riwayaCursor < n-1 {
				m.riwayaCursor++
			}
		case "k", "up":
			if m.riwayaCursor > 0 {
				m.riwayaCursor--
			}
		case "g", "home":
			m.riwayaCursor = 0
		case "G", "end":
			if n > 0 {
				m.riwayaCursor = n - 1
			}
		case "enter":
			if m.riwayaCursor >= 0 && m.riwayaCursor < n {
				opt := m.riwayaOpts[m.riwayaCursor]
				m.store.Settings.Rewaya = opt.ID
				m.saveStore()
				m.qarisList.cursor, m.qarisList.offset = 0, 0
				m.surahDetailList.cursor, m.surahDetailList.offset = 0, 0
				m.overlay = overlayNone
				return m, m.setToast("Riwayah: " + opt.Name)
			}
		}
		if m.riwayaCursor < m.riwayaOffset {
			m.riwayaOffset = m.riwayaCursor
		}
		if m.riwayaCursor >= m.riwayaOffset+vis {
			m.riwayaOffset = m.riwayaCursor - vis + 1
		}
		return m, nil

	case overlayRecording:
		if m.detailReciter == nil {
			m.overlay = overlayNone
			return m, nil
		}
		n := len(m.detailReciter.Moshaf)
		vis := max(3, min(10, m.height-8))
		switch key {
		case "esc", "e", "q":
			m.overlay = overlayNone
		case "j", "down":
			if m.recordingCursor < n-1 {
				m.recordingCursor++
			}
		case "k", "up":
			if m.recordingCursor > 0 {
				m.recordingCursor--
			}
		case "g", "home":
			m.recordingCursor = 0
		case "G", "end":
			if n > 0 {
				m.recordingCursor = n - 1
			}
		case "enter":
			if m.recordingCursor >= 0 && m.recordingCursor < n {
				chosen := m.detailReciter.Moshaf[m.recordingCursor]
				m.detailMoshaf = chosen
				m.qariDetailList.cursor, m.qariDetailList.offset = 0, 0
				m.overlay = overlayNone
				return m, m.setToast("Recording: " + moshafDisplayLabel(m.cat, chosen))
			}
		}
		if m.recordingCursor < m.recordingOffset {
			m.recordingOffset = m.recordingCursor
		}
		if m.recordingCursor >= m.recordingOffset+vis {
			m.recordingOffset = m.recordingCursor - vis + 1
		}
		return m, nil
	}

	return m, nil
}

func (m *Model) switchTab(t tabKind) {
	m.tab = t
	m.view = viewTabs
	m.clampActiveList()
}

func (m Model) handleActivate() (tea.Model, tea.Cmd) {
	switch m.view {
	case viewQariDetail:
		surahs := m.filteredQariDetailSurahs()
		if m.qariDetailList.cursor < 0 || m.qariDetailList.cursor >= len(surahs) || m.detailReciter == nil {
			return m, nil
		}
		s := surahs[m.qariDetailList.cursor]
		tr := store.Track{Reciter: m.detailReciter.ID, Moshaf: m.detailMoshaf.ID, Surah: s.ID}
		return m, m.startSurah(tr, false)

	case viewSurahDetail:
		items := m.filteredSurahDetailQaris()
		if m.surahDetailList.cursor < 0 || m.surahDetailList.cursor >= len(items) {
			return m, nil
		}
		it := items[m.surahDetailList.cursor]
		tr := store.Track{Reciter: it.Reciter.ID, Moshaf: it.Moshaf.ID, Surah: m.detailSurahID}
		return m, m.startSurah(tr, false)

	default:
		switch m.tab {
		case tabHome:
			items := m.homeItems()
			if m.homeList.cursor < 0 || m.homeList.cursor >= len(items) {
				return m, nil
			}
			it := items[m.homeList.cursor]
			if it.kind == homeContinue {
				return m, m.startSurah(it.progress.Track, false)
			}
			m.openQari(it.reciter, 0, 0)
			return m, nil

		case tabQaris:
			items := m.filteredQaris()
			if m.qarisList.cursor < 0 || m.qarisList.cursor >= len(items) {
				return m, nil
			}
			m.openQari(items[m.qarisList.cursor], 0, 0)
			return m, nil

		case tabSurahs:
			items := m.filteredSurahs()
			if m.surahsList.cursor < 0 || m.surahsList.cursor >= len(items) {
				return m, nil
			}
			m.openSurah(items[m.surahsList.cursor].ID)
			return m, nil

		case tabRadio:
			items := m.filteredRadios()
			if m.radioList.cursor < 0 || m.radioList.cursor >= len(items) {
				return m, nil
			}
			return m, m.startRadio(items[m.radioList.cursor])
		}
	}
	return m, nil
}

func (m Model) toggleFavorite() (tea.Model, tea.Cmd) {
	rec := m.selectedReciterForFavorite()
	if rec == nil {
		return m, nil
	}
	if m.store.Favorites[rec.ID] {
		delete(m.store.Favorites, rec.ID)
		m.saveStore()
		m.clampActiveList()
		return m, m.setToast("Removed from favorites: " + rec.Name)
	}
	m.store.Favorites[rec.ID] = true
	m.saveStore()
	return m, m.setToast("★ Added to favorites: " + rec.Name)
}

func (m Model) toggleShowAll() (tea.Model, tea.Cmd) {
	m.store.Settings.ShowAll = !m.store.Settings.ShowAll
	m.saveStore()
	m.clampActiveList()
	if m.store.Settings.ShowAll {
		return m, m.setToast("Showing all qaris")
	}
	return m, m.setToast("Showing well-known complete qaris")
}

func (m Model) cycleSort() (tea.Model, tea.Cmd) {
	m.store.Settings.Sort = nextSort(m.store.Settings.Sort)
	m.saveStore()
	m.clampActiveList()
	return m, m.setToast("Sort: " + sortLabel(m.store.Settings.Sort))
}

func (m Model) openRiwayahPicker() (tea.Model, tea.Cmd) {
	if len(m.riwayaOpts) == 0 {
		return m, nil
	}
	m.overlay = overlayRiwayah
	m.riwayaCursor = 0
	m.riwayaOffset = 0
	for i, o := range m.riwayaOpts {
		if o.ID == m.store.Settings.Rewaya {
			m.riwayaCursor = i
			break
		}
	}
	vis := max(3, min(12, m.height-8))
	if m.riwayaCursor >= vis {
		m.riwayaOffset = m.riwayaCursor - vis + 1
	}
	return m, nil
}

func (m Model) openRecordingPicker() (tea.Model, tea.Cmd) {
	if m.detailReciter == nil || len(m.detailReciter.Moshaf) <= 1 {
		return m, m.setToast("Only one recording available")
	}
	m.overlay = overlayRecording
	m.recordingCursor = 0
	m.recordingOffset = 0
	for i, mo := range m.detailReciter.Moshaf {
		if mo.ID == m.detailMoshaf.ID {
			m.recordingCursor = i
			break
		}
	}
	return m, nil
}

func (m Model) cycleRecording() (tea.Model, tea.Cmd) {
	if m.detailReciter == nil || len(m.detailReciter.Moshaf) <= 1 {
		return m, m.setToast("Only one recording available")
	}
	idx := 0
	for i, mo := range m.detailReciter.Moshaf {
		if mo.ID == m.detailMoshaf.ID {
			idx = (i + 1) % len(m.detailReciter.Moshaf)
			break
		}
	}
	m.detailMoshaf = m.detailReciter.Moshaf[idx]
	m.qariDetailList.cursor, m.qariDetailList.offset = 0, 0
	return m, m.setToast("Recording: " + moshafDisplayLabel(m.cat, m.detailMoshaf))
}

func (m Model) cycleTheme() (tea.Model, tea.Cmd) {
	m.themeIdx = (m.themeIdx + 1) % len(compiledThemes)
	m.theme = themeByIndex(m.themeIdx)
	m.applyThemeStyles()
	m.store.Settings.Theme = m.theme.ID
	m.saveStore()
	return m, m.setToast(fmt.Sprintf("Theme: %s (%s)", m.theme.Name, m.theme.ID))
}

func (m Model) togglePauseOrResume() (tea.Model, tea.Cmd) {
	if m.player == nil {
		return m, m.setToast("Install mpv to play audio")
	}
	if m.playKind == playNone {
		if m.store.Last != nil && m.cat != nil {
			return m, m.startSurah(*m.store.Last, false)
		}
		return m, m.setToast("Nothing playing yet")
	}
	m.player.TogglePause()
	m.paused = !m.paused
	if m.paused {
		m.saveStore()
		return m, m.setToast("Paused")
	}
	return m, m.setToast("Resumed")
}

func (m Model) seekBy(deltaSec float64) (tea.Model, tea.Cmd) {
	if m.player == nil || m.playKind != playSurah {
		return m, nil
	}
	m.player.Seek(deltaSec)
	m.playPos = max(0, m.playPos+deltaSec)
	if m.playDur > 0 && m.playPos > m.playDur {
		m.playPos = m.playDur
	}
	if deltaSec > 0 {
		return m, m.setToast(fmt.Sprintf("Seek +%ds", int(deltaSec)))
	}
	return m, m.setToast(fmt.Sprintf("Seek %ds", int(deltaSec)))
}

func (m Model) stepSurah(delta int) (tea.Model, tea.Cmd) {
	if m.playKind != playSurah {
		return m, nil
	}
	next, ok := m.adjacentSurah(m.playTrack, delta)
	if !ok {
		if delta > 0 {
			return m, m.setToast("Already at last surah in recording")
		}
		return m, m.setToast("Already at first surah in recording")
	}
	return m, m.startSurah(next, false)
}

func (m Model) adjustVolume(delta float64) (tea.Model, tea.Cmd) {
	if m.player == nil {
		return m, m.setToast("Install mpv to play audio")
	}
	v := m.player.AddVolume(delta)
	m.volume = v
	if m.muted && delta > 0 {
		m.muted = false
		m.player.SetMute(false)
	}
	return m, m.setToast(fmt.Sprintf("Volume %d%%", int(v+0.5)))
}

func (m Model) toggleMute() (tea.Model, tea.Cmd) {
	if m.player == nil {
		return m, m.setToast("Install mpv to play audio")
	}
	m.muted = !m.muted
	m.player.SetMute(m.muted)
	if m.muted {
		return m, m.setToast("Muted")
	}
	return m, m.setToast(fmt.Sprintf("Unmuted (%d%%)", int(m.volume+0.5)))
}

func (m Model) adjustSpeed(delta int) (tea.Model, tea.Cmd) {
	if m.player == nil {
		return m, m.setToast("Install mpv to play audio")
	}
	next := m.speedIdx + delta
	if next < 0 {
		next = 0
	}
	if next >= len(speedSteps) {
		next = len(speedSteps) - 1
	}
	m.speedIdx = next
	spd := speedSteps[m.speedIdx]
	m.player.SetSpeed(spd)
	return m, m.setToast("Speed " + formatSpeed(spd))
}

func (m Model) stopPlayback() (tea.Model, tea.Cmd) {
	if m.playKind == playNone {
		return m, nil
	}
	if m.player != nil {
		m.player.Stop()
	}
	m.saveStore()
	m.playKind = playNone
	m.playLoading = false
	m.paused = false
	m.playPos = 0
	m.playDur = 0
	m.playErr = ""
	return m, m.setToast("Stopped")
}

func (m Model) cycleRepeat() (tea.Model, tea.Cmd) {
	m.store.Settings.Repeat = nextRepeat(m.store.Settings.Repeat)
	m.saveStore()
	return m, m.setToast("After surah: " + repeatLongLabel(m.store.Settings.Repeat))
}

func (m Model) cycleSleepTimer() (tea.Model, tea.Cmd) {
	m.sleepMode = (m.sleepMode + 1) % 6
	switch m.sleepMode {
	case sleepOff:
		m.sleepUntil = time.Time{}
		return m, m.setToast("Sleep timer: off")
	case sleep15:
		m.sleepUntil = time.Now().Add(15 * time.Minute)
		return m, tea.Batch(m.setToast("Sleep timer: 15 min"), sleepTickCmd())
	case sleep30:
		m.sleepUntil = time.Now().Add(30 * time.Minute)
		return m, tea.Batch(m.setToast("Sleep timer: 30 min"), sleepTickCmd())
	case sleep45:
		m.sleepUntil = time.Now().Add(45 * time.Minute)
		return m, tea.Batch(m.setToast("Sleep timer: 45 min"), sleepTickCmd())
	case sleep60:
		m.sleepUntil = time.Now().Add(60 * time.Minute)
		return m, tea.Batch(m.setToast("Sleep timer: 60 min"), sleepTickCmd())
	case sleepEndOfSurah:
		m.sleepUntil = time.Time{}
		return m, m.setToast("Sleep timer: end of surah")
	}
	return m, nil
}

func (m Model) jumpToPlaying() (tea.Model, tea.Cmd) {
	switch m.playKind {
	case playSurah:
		if m.cat == nil {
			return m, nil
		}
		rec := m.cat.Reciter(m.playTrack.Reciter)
		if rec == nil {
			return m, nil
		}
		m.openQari(rec, m.playTrack.Moshaf, m.playTrack.Surah)
		return m, m.setToast("Jumped to playing surah")
	case playRadio:
		m.switchTab(tabRadio)
		m.radioList.query = ""
		radios := m.filteredRadios()
		for i, r := range radios {
			if r.URL == m.playRadio.URL {
				m.radioList.cursor = i
				break
			}
		}
		m.radioList.clamp(len(radios), m.activeVisibleRows())
		return m, m.setToast("Jumped to live radio")
	default:
		if m.store.Last != nil && m.cat != nil {
			if rec := m.cat.Reciter(m.store.Last.Reciter); rec != nil {
				m.openQari(rec, m.store.Last.Moshaf, m.store.Last.Surah)
				return m, m.setToast("Jumped to last listened surah")
			}
		}
		return m, m.setToast("Nothing playing")
	}
}
