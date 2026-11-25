package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const chrome = 7 // header, rule, crumb, rule, 2 player lines, help line

func (m Model) listHeight() int { return max(m.h-chrome, 1) }

func (m Model) View() string {
	if m.w == 0 {
		return ""
	}
	if m.loadErr != nil {
		return m.errorView()
	}
	if m.cat == nil {
		return "\n  " + m.spin.View() + sText.Render(" Loading qaris from mp3quran.net…")
	}
	if m.showHelp {
		return m.helpView()
	}
	rule := sRule.Render(strings.Repeat("─", m.w))
	parts := []string{
		m.headerView(),
		rule,
		m.crumbView(),
		m.listView(),
		rule,
		m.playerView(),
		m.footerView(),
	}
	return strings.Join(parts, "\n")
}

func (m Model) headerView() string {
	left := sTitle.Render(" ✦ qari ")
	for i, name := range tabNames {
		label := fmt.Sprintf("%d %s", i+1, name)
		if i == m.tab {
			left += sTabOn.Render(label)
		} else {
			left += sTab.Render(label)
		}
	}
	right := ""
	if m.opt.Client.Offline() {
		right = sGold.Render("offline ")
	}
	gap := max(m.w-lipgloss.Width(left)-lipgloss.Width(right), 1)
	return left + strings.Repeat(" ", gap) + right
}

func (m Model) crumbView() string {
	s := m.top()
	var title, info string
	st := m.opt.Store.Settings
	switch s.kind {
	case scrHome:
		title = "Home"
	case scrReciters:
		list, hidden := m.filteredReciters(s.surah)
		title = "Qaris"
		if s.surah > 0 {
			title = "Qaris of " + m.cat.SurahName(s.surah)
		}
		info = fmt.Sprintf("%d shown", len(list))
		if hidden > 0 {
			info += fmt.Sprintf(" · %d hidden (a)", hidden)
		}
		info += " · sort " + st.Sort
		if st.Rewaya != 0 {
			info += " · " + m.cat.Riwayat[st.Rewaya]
		}
	case scrMoshafs:
		title = s.reciter.Name
		info = "choose a recitation"
	case scrSurahs:
		title = s.reciter.Name + sDim.Render(" › ") + sText.Render(m.rewayaShort(s.moshaf))
		info = fmt.Sprintf("%d surahs", s.moshaf.SurahTotal)
		if m.opt.Store.Favorites[s.reciter.ID] {
			info = sGold.Render("★ ") + sDim.Render(info)
		}
	case scrAllSurahs:
		title = "Surahs"
		info = "enter: choose a qari"
	case scrRadio:
		title = "Radio"
		info = fmt.Sprintf("%d live stations", len(m.cat.Radios))
	}
	if m.searching {
		return " " + m.input.View()
	}
	line := " " + sText.Bold(true).Render(title)
	if s.list.query != "" {
		line += sAccent.Render("  /"+s.list.query) + sDim.Render(fmt.Sprintf(" (%d) esc clears", len(s.list.visible)))
	} else if info != "" {
		line += sDim.Render("  " + info)
	}
	return ansi.Truncate(line, m.w, "…")
}

func (m Model) listView() string {
	h := m.listHeight()
	body := m.top().list.view(m.w, h, m.showArabic())
	n := strings.Count(body, "\n") + 1
	if n < h {
		body += strings.Repeat("\n", h-n)
	}
	return body
}

func (m Model) playerView() string {
	if m.opt.Player == nil {
		return sErr.Render(" mpv was not found.") + sDim.Render(" Browsing works; install mpv to listen.") + "\n"
	}
	if m.now == nil {
		return sDim.Render(" Nothing playing — pick a surah and press enter") + "\n " + m.modesView()
	}
	n := m.now
	ev := n.ev
	icon := "▶"
	switch {
	case ev.Loading:
		icon = "…"
	case ev.Paused:
		icon = "‖"
	}
	var title, sub string
	if n.isRadio() {
		title, sub = strings.TrimSpace(n.radio.Name), "live radio"
	} else {
		e := m.cat.Surahs[n.track.Surah-1]
		title = e.Name
		if m.showArabic() && e.Arabic != "" {
			title += "  ⁧" + e.Arabic + "⁩"
		}
		sub = n.reciter.Name + " · " + m.rewayaShort(n.moshaf)
	}
	line1 := " " + sAccent.Bold(true).Render(icon+" "+title) + sDim.Render("  "+sub)
	line1 = ansi.Truncate(line1, m.w, "…")

	var line2 string
	if n.isRadio() || ev.Duration <= 0 {
		line2 = " " + sDim.Render(clock(ev.Pos)) + "  " + m.modesView()
	} else {
		modes := m.modesView()
		times := clock(ev.Pos) + " / " + clock(ev.Duration)
		barW := m.w - lipgloss.Width(times) - lipgloss.Width(modes) - 6
		line2 = " " + sText.Render(times) + "  " + bar(ev.Pos/ev.Duration, barW) + "  " + modes
	}
	return line1 + "\n" + line2
}

func bar(f float64, w int) string {
	if w < 4 {
		return ""
	}
	f = min(max(f, 0), 1)
	full := int(f * float64(w))
	return sBarFull.Render(strings.Repeat("━", full)) + sBarEmpty.Render(strings.Repeat("─", w-full))
}

func (m Model) modesView() string {
	parts := []string{
		fmt.Sprintf("vol %d", int(m.volume)),
		fmt.Sprintf("%.2gx", speeds[m.speedIdx]),
		repeatLabel[m.opt.Store.Settings.Repeat],
	}
	switch d := sleepOptions[m.sleepIdx]; {
	case d < 0:
		parts = append(parts, "sleep: end of surah")
	case d > 0:
		left := time.Until(m.sleepAt).Round(time.Minute)
		parts = append(parts, fmt.Sprintf("sleep %dm", int(left.Minutes())))
	}
	return sDim.Render(strings.Join(parts, " · "))
}

func (m Model) footerView() string {
	if m.status != "" {
		st := sGold
		if m.statusErr {
			st = sErr
		}
		return ansi.Truncate(" "+st.Render(m.status), m.w, "…")
	}
	keys := [][2]string{{"enter", "open/play"}, {"/", "search"}, {"esc", "back"}, {"space", "pause"}, {"←→", "seek"}, {"n/p", "next/prev"}, {"f", "favorite"}, {"a", "show all"}, {"?", "help"}, {"q", "quit"}}
	if m.searching {
		keys = [][2]string{{"type", "to filter"}, {"↑↓", "move"}, {"enter", "open"}, {"esc", "cancel"}}
	}
	var b strings.Builder
	b.WriteString(" ")
	for _, k := range keys {
		item := sKey.Render(k[0]) + sDim.Render(" "+k[1]+"  ")
		if lipgloss.Width(b.String())+lipgloss.Width(item) > m.w {
			break
		}
		b.WriteString(item)
	}
	return b.String()
}

func (m Model) helpView() string {
	sections := []struct {
		title string
		keys  [][2]string
	}{
		{"Browse", [][2]string{
			{"tab / 1-4", "switch tab"}, {"↑↓ j k", "move"}, {"pgup pgdn", "page"}, {"g G", "top / bottom"},
			{"enter", "open or play"}, {"esc", "clear search, then go back"}, {"/", "search (English, Arabic…)"},
			{"o", "open what is playing"},
		}},
		{"Filters", [][2]string{
			{"a", "all qaris / well-known complete ones"}, {"s", "sort: popular, a-z, recently updated"},
			{"w", "cycle riwayah (Hafs, Warsh, …)"}, {"f", "add / remove favorite"}, {"A", "show / hide Arabic surah names"},
		}},
		{"Playback", [][2]string{
			{"space", "pause / resume"}, {"← →", "seek 10 s"}, {"[ ]", "seek 1 min"}, {"n p", "next / previous surah"},
			{"x", "stop"}, {"- +", "volume"}, {"< >", "speed"}, {"r", "after a surah: autoplay / repeat / stop"},
			{"t", "sleep timer: 15-60 min or end of surah"},
		}},
	}
	var blocks []string
	for _, s := range sections {
		var b strings.Builder
		b.WriteString(sHeader.Render(s.title) + "\n")
		for _, k := range s.keys {
			b.WriteString("  " + sKey.Render(fmt.Sprintf("%-10s", k[0])) + " " + sText.Render(k[1]) + "\n")
		}
		blocks = append(blocks, b.String())
	}
	var body string
	if m.w >= 110 {
		left := lipgloss.JoinVertical(lipgloss.Left, blocks[0], blocks[1])
		body = lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(56).Render(left), blocks[2])
	} else {
		body = lipgloss.JoinVertical(lipgloss.Left, blocks...)
	}
	out := "\n " + sTitle.Render("qari "+m.opt.Version) + sDim.Render("  listen to the Quran from mp3quran.net") + "\n\n" +
		lipgloss.NewStyle().PaddingLeft(1).Render(body) + "\n" +
		sDim.Render(" Progress is saved automatically. Popularity: way2quran.com, quranicaudio.com, quran.com.") + "\n" +
		sDim.Render(" Press any key to close.")
	lines := strings.Split(out, "\n")
	if len(lines) > m.h { // tiny terminal: keep the top
		lines = lines[:m.h]
	}
	return strings.Join(lines, "\n")
}

func (m Model) errorView() string {
	msg := m.loadErr.Error()
	return "\n " + sErr.Bold(true).Render("Could not load the catalog from mp3quran.net") + "\n\n " +
		sText.Render(ansi.Wrap(msg, max(m.w-4, 20), "")) + "\n\n " +
		sDim.Render("Some networks block mp3quran.net. Try another network or a proxy, e.g.") + "\n   " +
		sKey.Render("qari --proxy socks5://127.0.0.1:40000") + "\n\n " + sDim.Render("q quits")
}

// showArabic: the Arabic column is redundant when names are already Arabic.
func (m Model) showArabic() bool { return m.arabic && m.cat.Lang != "ar" }
