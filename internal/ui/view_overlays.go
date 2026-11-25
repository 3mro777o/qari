package ui

import "fmt"

func (m Model) renderOverlay() []string {
	switch m.overlay {
	case overlayHelp:
		return m.renderHelpModal()
	case overlayRiwayah:
		return m.renderRiwayahModal()
	case overlayRecording:
		return m.renderRecordingModal()
	default:
		return nil
	}
}

func (m Model) renderHelpModal() []string {
	s := m.theme.Styles

	type kv struct{ k, d string }
	navRows := []kv{
		{"1-4 / tab", "Home, Qaris, Surahs, Radio"},
		{"j k / up dn", "Move selection (pgup/pgdn page)"},
		{"g / G", "Jump to top / bottom"},
		{"enter", "Open qari/surah or play"},
		{"/", "Search (English, Arabic, #)"},
		{"esc", "Close popup / clear / back"},
	}
	playRows := []kv{
		{"space", "Pause / resume (or last track)"},
		{"left / right", "Seek -10s / +10s"},
		{"shift+arrows", "Seek -1m / +1m (also < / >)"},
		{"n / p", "Next / previous surah"},
		{"+ / -   m", "Volume +/-5%, mute toggle"},
		{"[ / ]", "Speed 0.75x - 2.0x"},
		{"r", "After surah: auto/repeat/stop"},
		{"t", "Sleep: 15-60m or end of surah"},
		{"x", "Stop playback"},
		{"o", "Jump to what is playing"},
	}
	libRows := []kv{
		{"f", "Favorite / unfavorite qari"},
		{"a", "Toggle well-known vs all qaris"},
		{"s", "Sort: Popular, A-Z, Updated"},
		{"w", "Filter qaris by riwayah"},
		{"e / v", "Choose / cycle qari recording"},
		{"i", "Open qari page from surah view"},
	}
	appRows := []kv{
		{"c", "Cycle theme (" + m.theme.Name + ")"},
		{"A", "Show / hide Arabic surah names"},
		{"?", "Show / hide this help screen"},
		{"q / ctrl+c", "Quit and save progress"},
	}

	topRight := s.Muted.Render(m.theme.Name)
	if sv := shortVersion(m.version); sv != "" && sv != "dev" {
		topRight = s.Muted.Render(m.theme.Name + " · " + sv)
	}

	// Wide two-column help floats cleanly inside the canvas with zero scrolling needed.
	if m.width >= 100 && m.height >= 24 {
		boxW := min(m.width-6, 98)
		innerW := boxW - 4
		colW := (innerW - 3) / 2
		rightW := innerW - colW - 3
		keyW := 13

		renderSec := func(title string, rows []kv, cw int) []string {
			out := []string{s.SectionTitle.Render("✦ " + title)}
			for _, r := range rows {
				k := padRight(s.Key.Render(r.k), keyW)
				d := s.Text.Render(trunc(r.d, max(8, cw-keyW-1)))
				out = append(out, " "+k+d)
			}
			return out
		}

		leftCol := renderSec("NAVIGATION", navRows, colW)
		leftCol = append(leftCol, "")
		leftCol = append(leftCol, renderSec("LIBRARY & FILTERS", libRows, colW)...)

		rightCol := renderSec("PLAYBACK", playRows, rightW)
		rightCol = append(rightCol, "")
		rightCol = append(rightCol, renderSec("APPEARANCE & APP", appRows, rightW)...)

		bodyH := max(len(leftCol), len(rightCol))
		div := " " + s.Rule.Render("│") + " "
		joined := joinColumns(leftCol, rightCol, colW, rightW, bodyH, div)

		lines := make([]string, 0, bodyH+4)
		lines = append(lines, frameTopLine(boxW, s.Brand.Render("✦ Keyboard Shortcuts"), topRight, s.BorderModal))
		for _, jl := range joined {
			lines = append(lines, frameContentLine(boxW, jl, s.BorderModal))
		}
		lines = append(lines, frameContentLine(boxW, "", s.BorderModal))
		lines = append(lines, frameContentLine(boxW, center(s.Muted.Render("Arabic keyboard shortcuts follow key positions (no layout switch needed)"), innerW), s.BorderModal))
		lines = append(lines, frameBottomLine(boxW, s.Key.Render("esc")+" "+s.Muted.Render("or ")+s.Key.Render("?")+" "+s.Muted.Render("close"), s.Muted.Render("qari"), s.BorderModal))
		return lines
	}

	// Compact / narrow scrollable single-column help modal.
	boxW := min(m.width-2, 72)
	if boxW < 46 {
		boxW = m.width
	}
	innerW := boxW - 4
	maxBodyH := max(6, m.height-4)
	if m.height >= 24 {
		maxBodyH = min(maxBodyH, m.height-8)
	}

	allSecs := []struct {
		title string
		rows  []kv
	}{
		{"NAVIGATION", navRows},
		{"PLAYBACK", playRows},
		{"LIBRARY & FILTERS", libRows},
		{"APPEARANCE & APP", appRows},
	}

	var body []string
	keyCol := 14
	for si, sec := range allSecs {
		if si > 0 {
			body = append(body, "")
		}
		body = append(body, s.SectionTitle.Render("✦ "+sec.title))
		for _, r := range sec.rows {
			k := padRight(s.Key.Render(r.k), keyCol)
			d := s.Text.Render(trunc(r.d, max(10, innerW-keyCol-1)))
			body = append(body, " "+k+d)
		}
	}
	body = append(body, "")
	body = append(body, s.Muted.Render(trunc("Arabic keyboard shortcuts follow key positions.", innerW)))

	bodyH := min(len(body), maxBodyH)
	maxOff := max(0, len(body)-bodyH)
	off := m.helpOffset
	if off > maxOff {
		off = maxOff
	}
	if off < 0 {
		off = 0
	}

	lines := make([]string, 0, bodyH+2)
	lines = append(lines, frameTopLine(boxW, s.Brand.Render("✦ Keyboard Shortcuts"), topRight, s.BorderModal))
	for i := 0; i < bodyH; i++ {
		lines = append(lines, frameContentLine(boxW, body[off+i], s.BorderModal))
	}

	botLeft := s.Key.Render("esc") + " " + s.Muted.Render("close")
	if maxOff > 0 {
		botLeft = s.Key.Render("j/k") + " " + s.Muted.Render("scroll · ") + botLeft
	}
	botRight := ""
	if maxOff > 0 {
		botRight = s.Muted.Render(fmt.Sprintf("%d-%d/%d", off+1, off+bodyH, len(body)))
	}
	lines = append(lines, frameBottomLine(boxW, botLeft, botRight, s.BorderModal))
	return lines
}

func (m Model) renderRiwayahModal() []string {
	s := m.theme.Styles
	boxW := min(m.width-4, 56)
	if boxW < 36 {
		boxW = m.width
	}
	innerW := boxW - 4
	vis := max(3, min(12, m.height-6))
	if vis > len(m.riwayaOpts) {
		vis = len(m.riwayaOpts)
	}

	off := m.riwayaOffset
	if m.riwayaCursor < off {
		off = m.riwayaCursor
	}
	if m.riwayaCursor >= off+vis {
		off = m.riwayaCursor - vis + 1
	}
	if maxOff := max(0, len(m.riwayaOpts)-vis); off > maxOff {
		off = maxOff
	}

	lines := make([]string, 0, vis+3)
	lines = append(lines, frameTopLine(boxW, s.SectionTitle.Render("✦ Filter by Riwayah"), s.Muted.Render(fmt.Sprintf("%d options", len(m.riwayaOpts))), s.BorderModal))

	for i := off; i < off+vis && i < len(m.riwayaOpts); i++ {
		opt := m.riwayaOpts[i]
		sel := i == m.riwayaCursor
		active := opt.ID == m.store.Settings.Rewaya

		rail := "  "
		nameStyle := s.Text
		if sel {
			rail = s.SelectedRail.Render("› ")
			nameStyle = s.SelectedText
		}
		check := "  "
		if active {
			check = s.Success.Render("✓ ")
		}
		left := rail + check + nameStyle.Render(opt.Name)
		right := s.Muted.Render(fmt.Sprintf("%d qaris", opt.Count))
		lines = append(lines, frameContentLine(boxW, rowLR(left, right, innerW), s.BorderModal))
	}

	lines = append(lines, frameBottomLine(boxW,
		s.Key.Render("enter")+" "+s.Muted.Render("apply · ")+s.Key.Render("esc")+" "+s.Muted.Render("cancel"),
		s.Muted.Render(fmt.Sprintf("%d/%d", m.riwayaCursor+1, max(1, len(m.riwayaOpts)))),
		s.BorderModal,
	))
	return lines
}

func (m Model) renderRecordingModal() []string {
	s := m.theme.Styles
	if m.detailReciter == nil {
		return nil
	}
	moshafs := m.detailReciter.Moshaf
	boxW := min(m.width-4, 64)
	if boxW < 40 {
		boxW = m.width
	}
	innerW := boxW - 4
	vis := max(3, min(10, m.height-6))
	if vis > len(moshafs) {
		vis = len(moshafs)
	}

	off := m.recordingOffset
	if m.recordingCursor < off {
		off = m.recordingCursor
	}
	if m.recordingCursor >= off+vis {
		off = m.recordingCursor - vis + 1
	}
	if maxOff := max(0, len(moshafs)-vis); off > maxOff {
		off = maxOff
	}

	lines := make([]string, 0, vis+3)
	lines = append(lines, frameTopLine(boxW, s.SectionTitle.Render("✦ Choose Recording"), s.Muted.Render(trunc(m.detailReciter.Name, 24)), s.BorderModal))

	for i := off; i < off+vis && i < len(moshafs); i++ {
		mo := moshafs[i]
		sel := i == m.recordingCursor
		active := mo.ID == m.detailMoshaf.ID

		rail := "  "
		nameStyle := s.Text
		if sel {
			rail = s.SelectedRail.Render("› ")
			nameStyle = s.SelectedText
		}
		check := "  "
		if active {
			check = s.Success.Render("✓ ")
		}
		lbl := moshafDisplayLabel(m.cat, mo)
		left := rail + check + nameStyle.Render(lbl)
		right := s.Subtext.Render(fmt.Sprintf("%d surahs", mo.SurahTotal))
		lines = append(lines, frameContentLine(boxW, rowLR(left, right, innerW), s.BorderModal))
	}

	lines = append(lines, frameBottomLine(boxW,
		s.Key.Render("enter")+" "+s.Muted.Render("select · ")+s.Key.Render("esc")+" "+s.Muted.Render("cancel"),
		s.Muted.Render(fmt.Sprintf("%d/%d", m.recordingCursor+1, max(1, len(moshafs)))),
		s.BorderModal,
	))
	return lines
}
