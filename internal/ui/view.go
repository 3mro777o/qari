package ui

import (
	"fmt"
	"strings"
)

// View renders the complete terminal frame at exact m.width × m.height dimensions.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	if m.loading {
		return fitFrame(m.renderLoading(), m.width, m.height)
	}
	if m.loadErr != nil {
		return fitFrame(m.renderError(), m.width, m.height)
	}

	header := m.renderHeader()
	canvas := m.renderCanvas()
	dock := m.renderDock()

	lines := make([]string, 0, m.height)
	lines = append(lines, header...)
	lines = append(lines, canvas...)
	lines = append(lines, dock...)

	base := fitFrame(strings.Join(lines, "\n"), m.width, m.height)

	if m.overlay != overlayNone {
		modal := m.renderOverlay()
		base = overlayCenter(base, modal, m.width, m.height)
	}
	return fitFrame(base, m.width, m.height)
}

func (m Model) renderLoading() string {
	s := m.theme.Styles
	w, h := m.width, m.height
	boxW := min(w-4, 52)
	if boxW < 30 {
		boxW = w
	}

	var box []string
	box = append(box, frameTopLine(boxW, s.Brand.Render("✦ qari"), s.Muted.Render(shortVersion(m.version)), s.BorderPlay))
	box = append(box, frameContentLine(boxW, "", s.BorderPlay))
	box = append(box, frameContentLine(boxW, center(m.spinner.View()+"  "+s.TextBold.Render("Connecting to mp3quran.net"), boxW-4), s.BorderPlay))
	box = append(box, frameContentLine(boxW, center(s.Subtext.Render("Loading qaris, surahs, riwayat and radio stations"), boxW-4), s.BorderPlay))
	box = append(box, frameContentLine(boxW, "", s.BorderPlay))
	box = append(box, frameBottomLine(boxW, s.Muted.Render("cached for 24h after first load"), s.Muted.Render("q quit"), s.BorderPlay))

	return centerBlockVertically(box, w, h)
}

func (m Model) renderError() string {
	s := m.theme.Styles
	w, h := m.width, m.height
	boxW := min(w-4, 68)
	if boxW < 40 {
		boxW = w
	}
	innerW := boxW - 4

	var box []string
	box = append(box, frameTopLine(boxW, s.Live.Render("● Could not load catalog"), s.Muted.Render("qari"), s.BorderLive))
	box = append(box, frameContentLine(boxW, "", s.BorderLive))
	for _, wl := range wrapWords(m.loadErr.Error(), innerW, 3) {
		box = append(box, frameContentLine(boxW, s.Text.Render(wl), s.BorderLive))
	}
	box = append(box, frameContentLine(boxW, "", s.BorderLive))
	box = append(box, frameContentLine(boxW, s.Subtext.Render("Some networks block mp3quran.net. If so, run with a proxy:"), s.BorderLive))
	box = append(box, frameContentLine(boxW, "  "+s.AccentBold.Render("qari --proxy socks5://127.0.0.1:1080"), s.BorderLive))
	box = append(box, frameContentLine(boxW, "", s.BorderLive))
	box = append(box, frameBottomLine(boxW,
		s.Key.Render("enter")+" "+s.Muted.Render("retry")+" · "+s.Key.Render("c")+" "+s.Muted.Render("theme"),
		s.Key.Render("q")+" "+s.Muted.Render("quit"),
		s.BorderLive,
	))

	return centerBlockVertically(box, w, h)
}

func centerBlockVertically(lines []string, w, h int) string {
	top := max(0, (h-len(lines))/2)
	out := make([]string, 0, h)
	blank := strings.Repeat(" ", w)
	for i := 0; i < top; i++ {
		out = append(out, blank)
	}
	for _, l := range lines {
		out = append(out, center(l, w))
	}
	for len(out) < h {
		out = append(out, blank)
	}
	return strings.Join(out[:h], "\n")
}

// renderHeader renders the 2-line architectural navigation rail + segmented underline / search bar.
func (m Model) renderHeader() []string {
	s := m.theme.Styles
	w := m.width

	brand := " " + s.Brand.Render("✦ qari")
	brandW := dispWidth(brand)

	tabGap := "   "
	leadGap := "    "
	if w < 72 {
		tabGap = " "
		leadGap = "  "
	} else if w < 90 {
		tabGap = "  "
		leadGap = "   "
	}

	tabNames := []string{"Home", "Qaris", "Surahs", "Radio"}
	var leftBuf strings.Builder
	leftBuf.WriteString(brand)
	leftBuf.WriteString(leadGap)
	curCol := brandW + len(leadGap)

	activeStart, activeEnd := 0, 0
	for i, name := range tabNames {
		if i > 0 {
			leftBuf.WriteString(tabGap)
			curCol += len(tabGap)
		}
		numStr := fmt.Sprintf("%d", i+1)
		var seg string
		if tabKind(i) == m.tab && m.view == viewTabs {
			seg = s.AccentBold.Render(numStr) + " " + s.TabActive.Render(name)
			activeStart = curCol
			activeEnd = curCol + dispWidth(seg)
		} else if tabKind(i) == m.tab {
			// Inside a detail view spawned from this tab: keep subtle accent
			seg = s.Accent.Render(numStr) + " " + s.Accent.Render(name)
			activeStart = curCol
			activeEnd = curCol + dispWidth(seg)
		} else {
			seg = s.TabNum.Render(numStr) + " " + s.TabInactive.Render(name)
		}
		leftBuf.WriteString(seg)
		curCol += dispWidth(seg)
	}

	var rightParts []string
	if m.isOffline() {
		rightParts = append(rightParts, s.OfflineBadge.Render("OFFLINE"))
	}
	if w >= 80 {
		rightParts = append(rightParts, s.Muted.Render(m.theme.ID))
	}
	if w >= 66 {
		rightParts = append(rightParts, s.Key.Render("?")+" "+s.Muted.Render("help"))
	} else {
		rightParts = append(rightParts, s.Key.Render("?"))
	}
	right := strings.Join(rightParts, s.Muted.Render(" · ")) + " "

	line1 := rowLR(leftBuf.String(), right, w)

	// Line 2: focused search bar, active search filter pill, or architectural underline rule.
	var line2 string
	al := m.activeList()
	if m.searching {
		count := m.activeListCount()
		prompt := " " + s.AccentBold.Render("/") + " "
		matchStr := fmt.Sprintf("%d", count)
		rightHint := s.Subtext.Render(matchStr+" matches") + s.Muted.Render(" · enter done · esc clear ")
		if w < 72 {
			rightHint = s.Subtext.Render(matchStr) + s.Muted.Render(" · esc ")
		}
		inputMax := max(8, w-dispWidth(prompt)-dispWidth(rightHint)-2)
		m.searchInput.Width = inputMax
		left := prompt + m.searchInput.View()
		line2 = rowLR(left, rightHint, w)
	} else if al.query != "" {
		count := m.activeListCount()
		left := " " + s.AccentBold.Render("Filter:") + " " + s.TextBold.Render(al.query)
		rightHint := s.Subtext.Render(fmt.Sprintf("%d matches", count)) + s.Muted.Render(" · / edit · esc clear ")
		if w < 72 {
			rightHint = s.Subtext.Render(fmt.Sprintf("%d", count)) + s.Muted.Render(" · esc clear ")
		}
		line2 = rowLR(left, rightHint, w)
	} else {
		line2 = renderUnderlineRule(w, activeStart, activeEnd, s.Rule, s.RuleAccent)
	}

	return []string{line1, line2}
}

func renderUnderlineRule(w, activeStart, activeEnd int, ruleStyle, accentStyle interface{ Render(...string) string }) string {
	if activeEnd <= activeStart || activeStart < 0 || activeEnd > w {
		return ruleStyle.Render(strings.Repeat("─", w))
	}
	var b strings.Builder
	if activeStart > 0 {
		b.WriteString(ruleStyle.Render(strings.Repeat("─", activeStart)))
	}
	b.WriteString(accentStyle.Render(strings.Repeat("━", activeEnd-activeStart)))
	if activeEnd < w {
		b.WriteString(ruleStyle.Render(strings.Repeat("─", w-activeEnd)))
	}
	return b.String()
}

func (m Model) renderCanvas() []string {
	ch := m.canvasHeight()
	var lines []string
	switch m.view {
	case viewQariDetail:
		lines = m.renderQariDetail(ch)
	case viewSurahDetail:
		lines = m.renderSurahDetail(ch)
	default:
		switch m.tab {
		case tabHome:
			lines = m.renderHome(ch)
		case tabQaris:
			lines = m.renderQaris(ch)
		case tabSurahs:
			lines = m.renderSurahs(ch)
		case tabRadio:
			lines = m.renderRadio(ch)
		}
	}
	if len(lines) > ch {
		lines = lines[:ch]
	}
	blank := strings.Repeat(" ", m.width)
	for len(lines) < ch {
		lines = append(lines, blank)
	}
	return lines
}

// renderDock renders the living bottom player frame (4 lines normal, 3 lines compact).
func (m Model) renderDock() []string {
	s := m.theme.Styles
	w := m.width
	compact := m.dockHeight() == 3

	border := s.BorderIdle
	switch {
	case m.playKind == playRadio:
		if m.paused {
			border = s.BorderPause
		} else {
			border = s.BorderLive
		}
	case m.playKind == playSurah:
		if m.paused {
			border = s.BorderPause
		} else {
			border = s.BorderPlay
		}
	}

	// Build top-border left title & right telemetry.
	var topLeft, topRight string
	var midLeft, midRight string
	var barLine string

	// Telemetry badges (speed, volume/mute, sleep)
	var tele []string
	if m.sleepMode != sleepOff {
		tele = append(tele, s.Warn.Render("sleep "+m.sleepRemainingText()))
	}
	if m.playKind == playSurah && (speedSteps[m.speedIdx] != 1.0 || w >= 86) {
		tele = append(tele, s.Subtext.Render(formatSpeed(speedSteps[m.speedIdx])))
	}
	if m.player != nil {
		if m.muted {
			tele = append(tele, s.Warn.Render("muted"))
		} else {
			tele = append(tele, s.Subtext.Render(fmt.Sprintf("vol %d%%", int(m.volume+0.5))))
		}
	}
	topRight = strings.Join(tele, s.Muted.Render(" · "))

	switch {
	case m.player == nil:
		topLeft = s.Warn.Render("✦ mpv not installed")
		topRight = s.Muted.Render("browse mode")
		midLeft = s.Subtext.Render("Install mpv to play audio; browsing and search work normally")
		midRight = s.Key.Render("?") + " " + s.Muted.Render("keys")
		barLine = s.Muted.Render(strings.Repeat("─", max(1, w-4)))

	case m.playKind == playSurah && m.cat != nil:
		surahName := m.cat.SurahName(m.playTrack.Surah)
		rec := m.cat.Reciter(m.playTrack.Reciter)
		recName := ""
		moshafLabel := ""
		if rec != nil {
			recName = rec.Name
			if mo, ok := findMoshafByID(rec, m.playTrack.Moshaf); ok {
				moshafLabel = moshafCompactLabel(m.cat, mo)
			}
		}

		icon := s.AccentBold.Render("▶")
		stateNote := ""
		if m.playLoading {
			icon = m.spinner.View()
			stateNote = s.Muted.Render(" loading..")
		} else if m.paused {
			icon = s.Accent2.Render("‖")
			stateNote = s.Muted.Render(" paused")
		}
		arabicSuffix := ""
		if !m.noArabic && w >= 84 && m.playTrack.Surah >= 1 && m.playTrack.Surah <= len(m.cat.Surahs) {
			if ar := m.cat.Surahs[m.playTrack.Surah-1].Arabic; ar != "" {
				arabicSuffix = "  " + s.ArabicSel.Render(fitArabic(ar, 20))
			}
		}
		topLeft = icon + " " + s.TextBold.Render(fmt.Sprintf("%03d · %s", m.playTrack.Surah, surahName)) + arabicSuffix + stateNote

		midLeft = s.Text.Render(recName)
		if moshafLabel != "" {
			midLeft += s.Muted.Render(" · ") + s.Subtext.Render(moshafLabel)
		}
		timeStr := formatDuration(m.playPos) + " / " + formatDuration(m.playDur)
		if m.playDur <= 0 {
			timeStr = formatDuration(m.playPos)
		}
		modeStr := repeatLabel(m.store.Settings.Repeat)
		midRight = s.Text.Render(timeStr) + s.Muted.Render(" · ") + s.Accent2.Render(modeStr)

		frac := 0.0
		if m.playDur > 0 {
			frac = m.playPos / m.playDur
		}
		barLine = renderProgressBar(frac, max(4, w-4), s.ProgFill, s.ProgHead, s.ProgEmpty)

	case m.playKind == playRadio:
		icon := s.Live.Render("● LIVE")
		if m.playLoading {
			icon = m.spinner.View() + " " + s.Live.Render("CONNECTING")
		} else if m.paused {
			icon = s.Accent2.Render("‖ PAUSED")
		}
		topLeft = icon + " " + s.TextBold.Render(m.playRadio.Name)
		midLeft = s.Subtext.Render("Live Quran radio stream")
		midRight = s.Key.Render("space") + " " + s.Muted.Render("pause · ") + s.Key.Render("x") + " " + s.Muted.Render("stop")
		liveW := max(4, w-4)
		tag := " " + s.Live.Render("● LIVE STREAM") + " "
		if m.paused {
			tag = " " + s.Accent2.Render("‖ STREAM PAUSED") + " "
		}
		tagW := dispWidth(tag)
		side := max(0, (liveW-tagW)/2)
		rem := max(0, liveW-tagW-side)
		barLine = s.BorderIdle.Render(strings.Repeat("─", side)) + tag + s.BorderIdle.Render(strings.Repeat("─", rem))

	default:
		// Idle state: offer last listened track cleanly if available.
		if m.store.Last != nil && m.cat != nil {
			tr := *m.store.Last
			surahName := m.cat.SurahName(tr.Surah)
			recName := ""
			if rec := m.cat.Reciter(tr.Reciter); rec != nil {
				recName = rec.Name
			}
			topLeft = s.Muted.Render("✦ Last:") + " " + s.Text.Render(fmt.Sprintf("%03d · %s", tr.Surah, surahName))
			if topRight == "" {
				topRight = s.Key.Render("space") + " " + s.Muted.Render("resume")
			} else {
				topRight = s.Key.Render("space") + " " + s.Muted.Render("resume · ") + topRight
			}
			midLeft = s.Subtext.Render(recName)
			frac := 0.0
			if p := m.store.Get(tr); p != nil {
				frac = p.Fraction()
				if p.Finished {
					midRight = s.Success.Render("✓ finished")
				} else if p.Duration > 0 {
					midRight = s.Subtext.Render(formatDuration(p.Position) + " / " + formatDuration(p.Duration))
				}
			}
			barLine = renderProgressBar(frac, max(4, w-4), s.Subtext, s.Subtext, s.ProgEmpty)
		} else {
			topLeft = s.Muted.Render("✦ Ready to listen")
			midLeft = s.Subtext.Render("Choose a qari, a surah, or live radio and press enter")
			midRight = s.Key.Render("/") + " " + s.Muted.Render("search · ") + s.Key.Render("c") + " " + s.Muted.Render("theme")
			barLine = s.BorderIdle.Render(strings.Repeat("─", max(4, w-4)))
		}
	}

	// Bottom border: toast notification takes priority on the left, else contextual key hints.
	var botLeft, botRight string
	if m.toast != "" {
		botLeft = s.Toast.Render("✦ " + m.toast)
		if w >= 78 {
			botRight = s.Key.Render("?") + " " + s.Muted.Render("help")
		}
	} else {
		botLeft, botRight = m.dockHints(w)
	}

	if compact {
		// 3-line compact dock for heights < 20 (e.g. 60x16, 64x18)
		innerW := max(4, w-4)
		var row string
		if m.playKind == playSurah {
			frac := 0.0
			if m.playDur > 0 {
				frac = m.playPos / m.playDur
			}
			miniW := min(14, max(6, innerW/5))
			mini := renderMiniBar(frac, miniW, s.ProgFill, s.ProgEmpty)
			row = rowLR(midLeft, mini+" "+midRight, innerW)
		} else {
			row = rowLR(midLeft, midRight, innerW)
		}
		return []string{
			frameTopLine(w, topLeft, topRight, border),
			frameContentLine(w, row, border),
			frameBottomLine(w, botLeft, botRight, border),
		}
	}

	innerW := max(4, w-4)
	return []string{
		frameTopLine(w, topLeft, topRight, border),
		frameContentLine(w, rowLR(midLeft, midRight, innerW), border),
		frameContentLine(w, barLine, border),
		frameBottomLine(w, botLeft, botRight, border),
	}
}

func (m Model) dockHints(w int) (string, string) {
	s := m.theme.Styles
	kv := func(k, v string) string { return s.Key.Render(k) + " " + s.Muted.Render(v) }
	sep := s.Muted.Render(" · ")

	if m.playKind == playSurah {
		if w < 78 {
			return strings.Join([]string{kv("space", "pause"), kv("n/p", "surah"), kv("c", "theme")}, sep), kv("?", "help")
		}
		return strings.Join([]string{
			kv("space", "pause"), kv("<-/->", "seek"), kv("n/p", "surah"),
			kv("r", repeatLabel(m.store.Settings.Repeat)), kv("t", "sleep"), kv("o", "jump"),
		}, sep), kv("c", "theme") + sep + kv("?", "help")
	}
	if m.playKind == playRadio {
		return strings.Join([]string{
			kv("space", "pause"), kv("x", "stop"), kv("+/-", "vol"), kv("m", "mute"),
		}, sep), kv("c", "theme") + sep + kv("?", "help")
	}

	switch m.view {
	case viewQariDetail:
		hints := []string{kv("enter", "play"), kv("/", "search"), kv("f", "fav")}
		if m.detailReciter != nil && len(m.detailReciter.Moshaf) > 1 {
			hints = append(hints, kv("e", "recording"))
		}
		hints = append(hints, kv("esc", "back"))
		return strings.Join(hints, sep), kv("c", "theme") + sep + kv("?", "help")
	case viewSurahDetail:
		return strings.Join([]string{
			kv("enter", "play"), kv("i", "qari"), kv("f", "fav"), kv("esc", "back"),
		}, sep), kv("c", "theme") + sep + kv("?", "help")
	default:
		switch m.tab {
		case tabRadio:
			if w < 78 {
				return strings.Join([]string{kv("enter", "listen"), kv("/", "search"), kv("c", "theme")}, sep), kv("?", "help")
			}
			return strings.Join([]string{
				kv("enter", "listen"), kv("/", "search"), kv("c", "theme"), kv("q", "quit"),
			}, sep), kv("?", "help")
		case tabSurahs:
			if w < 78 {
				return strings.Join([]string{kv("enter", "open"), kv("/", "search"), kv("A", "arabic")}, sep), kv("?", "help")
			}
			return strings.Join([]string{
				kv("enter", "choose qari"), kv("/", "search"), kv("A", "arabic"), kv("c", "theme"), kv("q", "quit"),
			}, sep), kv("?", "help")
		default:
			if w < 78 {
				return strings.Join([]string{kv("enter", "open"), kv("/", "search"), kv("c", "theme")}, sep), kv("?", "help")
			}
			return strings.Join([]string{
				kv("enter", "open"), kv("/", "search"), kv("f", "fav"), kv("c", "theme"), kv("q", "quit"),
			}, sep), kv("?", "help")
		}
	}
}

func formatSpeed(spd float64) string {
	if spd == float64(int(spd)) {
		return fmt.Sprintf("%.1fx", spd)
	}
	return fmt.Sprintf("%.2gx", spd)
}

func shortVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || strings.HasPrefix(v, "v0.0.0-") || len(v) > 14 {
		return "dev"
	}
	return v
}

func wrapWords(text string, maxW, maxLines int) []string {
	if maxW <= 4 {
		return []string{trunc(text, maxW)}
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	cur := words[0]
	for _, wd := range words[1:] {
		if dispWidth(cur)+1+dispWidth(wd) <= maxW {
			cur += " " + wd
		} else {
			lines = append(lines, cur)
			cur = wd
			if len(lines) >= maxLines-1 {
				break
			}
		}
	}
	lines = append(lines, cur)
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	for i := range lines {
		lines[i] = trunc(lines[i], maxW)
	}
	return lines
}
