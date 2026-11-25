package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// chrome is the height of everything but the list: header, rule, crumb,
// player card (4) and footer, plus a spacer under the crumb when tall enough.
func (m Model) chrome() int {
	if m.h >= 24 {
		return 9
	}
	return 8
}

func (m Model) listHeight() int { return max(m.h-m.chrome(), 1) }

func (m Model) View() string {
	if m.w == 0 {
		return ""
	}
	if m.loadErr != nil {
		return m.errorView()
	}
	if m.cat == nil {
		return "\n   " + m.spin.View() + sText.Render(" Loading qaris from mp3quran.net…")
	}
	if m.showHelp {
		return m.helpView()
	}
	header, rule := m.headerView()
	parts := []string{header, rule, m.crumbView()}
	if m.chrome() == 9 {
		parts = append(parts, "")
	}
	parts = append(parts,
		m.listView(),
		m.playerView(),
		m.footerView(),
	)
	return strings.Join(parts, "\n")
}

// headerView returns the title line and the rule under it; the rule turns
// into a gold underline beneath the active tab.
func (m Model) headerView() (string, string) {
	left := sTitle.Render(" ✦ qari") + "    "
	ulFrom, ulTo := 0, 0
	for i, name := range tabNames {
		digit := sDim.Render(fmt.Sprint(i + 1))
		if i == m.tab {
			ulFrom = lipgloss.Width(left) + 2
			ulTo = ulFrom + len(name)
			left += sGold.Render(fmt.Sprint(i+1)) + " " + sText.Bold(true).Render(name)
		} else {
			left += digit + " " + sDim.Render(name)
		}
		left += "    "
	}
	right := ""
	if m.opt.Client.Offline() {
		right = sErr.Render("offline ●") + " "
	}
	gap := max(m.w-lipgloss.Width(left)-lipgloss.Width(right), 1)
	header := ansi.Truncate(left+strings.Repeat(" ", gap)+right, m.w, "")
	ulFrom, ulTo = min(ulFrom, m.w), min(ulTo, m.w)
	rule := sRule.Render(strings.Repeat("─", ulFrom)) + sGold.Render(strings.Repeat("━", ulTo-ulFrom)) +
		sRule.Render(strings.Repeat("─", m.w-ulTo))
	return header, rule
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
			info += fmt.Sprintf("   %d hidden (a)", hidden)
		}
		info += "   sort: " + st.Sort
		if st.Rewaya != 0 {
			info += "   " + m.cat.Riwayat[st.Rewaya]
		}
	case scrMoshafs:
		title = s.reciter.Name
		info = "choose a recitation"
	case scrSurahs:
		title = s.reciter.Name + sGold.Render(" › ") + sText.Render(m.rewayaShort(s.moshaf))
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
		line += sAccent.Render("   /"+s.list.query) + sDim.Render("   esc clears")
	} else if info != "" {
		line += sDim.Render("   " + info)
	}
	pos := ""
	if n := s.list.count(); n > 0 {
		pos = sDim.Render(fmt.Sprintf("%d of %d ", s.list.position(), n))
	}
	line = ansi.Truncate(line, max(m.w-lipgloss.Width(pos)-1, 0), "…")
	return line + strings.Repeat(" ", max(m.w-lipgloss.Width(line)-lipgloss.Width(pos), 1)) + pos
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

// playerView is a framed card. While a surah plays, the frame is gilded and
// its name sits in the top edge like the heading of a surah in a mushaf.
func (m Model) playerView() string {
	inner := max(m.w-4, 10) // border and padding
	var line1, line2, inset string
	frame := hairline
	switch {
	case m.opt.Player == nil:
		line1 = sErr.Render("mpv was not found.") + sDim.Render(" Browsing works; install mpv to listen.")
	case m.now == nil:
		line1 = sDim.Render("Nothing playing. Pick a surah and press enter.")
		line2 = m.modesView(inner)
	default:
		n, ev := m.now, m.now.ev
		icon := "▶"
		frame = gold
		switch {
		case ev.Loading:
			icon, frame = "…", goldDim
		case ev.Paused:
			icon, frame = "‖", goldDim
		}
		var title, sub string
		if n.isRadio() {
			title, sub, inset = strings.TrimSpace(n.radio.Name), "live radio", "live radio"
		} else {
			e := m.cat.Surahs[n.track.Surah-1]
			title = e.Name
			inset = e.Name
			if m.showArabic() && e.Arabic != "" {
				inset = "\u2067" + e.Arabic + "\u2069"
			}
			sub = n.reciter.Name + " (" + m.rewayaShort(n.moshaf) + ")"
		}
		modes := m.modesView(inner / 2)
		left := sGold.Render(icon) + "  " + sText.Bold(true).Render(title) + sDim.Render("   "+sub)
		left = ansi.Truncate(left, max(inner-lipgloss.Width(modes)-3, 0), "…")
		line1 = left + strings.Repeat(" ", max(inner-lipgloss.Width(left)-lipgloss.Width(modes), 1)) + modes
		if n.isRadio() || ev.Duration <= 0 {
			line2 = sText.Render(clock(ev.Pos)) + sDim.Render("   live")
		} else {
			cur, total := clock(ev.Pos), clock(ev.Duration)
			barW := inner - lipgloss.Width(cur) - lipgloss.Width(total) - 2
			line2 = sText.Render(cur) + " " + bar(ev.Pos/ev.Duration, barW, lipgloss.NewStyle().Foreground(frame)) + " " + sDim.Render(total)
		}
	}
	bc := lipgloss.NewStyle().Foreground(frame)
	card := lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), false, true, true, true).
		BorderForeground(frame).Padding(0, 1).Width(m.w - 2)
	body := card.Render(ansi.Truncate(line1, inner, "…") + "\n" + ansi.Truncate(line2, inner, "…"))
	return m.cardTop(inset, bc) + "\n" + body
}

// cardTop is the card's top edge with the inset centered in it.
func (m Model) cardTop(inset string, bc lipgloss.Style) string {
	span := m.w - 2
	if inset == "" || ansi.StringWidth(inset)+6 > span {
		return bc.Render("╭" + strings.Repeat("─", span) + "╮")
	}
	mid := sGold.Render(" ✦ ") + sGold.Bold(true).Render(inset) + sGold.Render(" ✦ ")
	rest := span - lipgloss.Width(mid)
	return bc.Render("╭"+strings.Repeat("─", rest/2)) + mid + bc.Render(strings.Repeat("─", rest-rest/2)+"╮")
}

func bar(f float64, w int, fill lipgloss.Style) string {
	if w < 4 {
		return ""
	}
	f = min(max(f, 0), 1)
	full := min(int(f*float64(w)), w-1)
	return fill.Render(strings.Repeat("━", full)) + sAccent.Render("●") + sBarEmpty.Render(strings.Repeat("─", w-full-1))
}

// modesView shows volume, speed, after-surah mode and sleep timer, dropping
// the least important ones when space is short.
func (m Model) modesView(room int) string {
	vol := fmt.Sprintf("vol %d", int(m.volume))
	if m.muted {
		vol = "muted"
	}
	parts := []string{vol, fmt.Sprintf("%.2g×", speeds[m.speedIdx]), "after: " + repeatLabel[m.opt.Store.Settings.Repeat]}
	switch d := sleepOptions[m.sleepIdx]; {
	case d < 0:
		parts = append(parts, "sleep: end of surah")
	case d > 0:
		parts = append(parts, fmt.Sprintf("sleep %dm", max(int(time.Until(m.sleepAt).Minutes()+0.5), 0)))
	}
	for len(parts) > 1 && lipgloss.Width(strings.Join(parts, "   ")) > room {
		parts = parts[1:]
	}
	return sDim.Render(strings.Join(parts, "   "))
}

func (m Model) footerView() string {
	if m.status != "" {
		st := sGold
		if m.statusErr {
			st = sErr
		}
		return ansi.Truncate(" "+st.Render(m.status), m.w, "…")
	}
	all := "show all"
	if m.opt.Store.Settings.ShowAll {
		all = "well-known only"
	}
	var keys [][2]string
	switch {
	case m.searching:
		keys = [][2]string{{"type", "to filter"}, {"↑↓", "move"}, {"enter", "open"}, {"esc", "cancel"}}
	case m.now != nil:
		pause := "pause"
		if m.now.ev.Paused {
			pause = "resume"
		}
		keys = [][2]string{{"space", pause}, {"←→", "seek"}, {"n/p", "next/prev surah"}, {"/", "search"}, {"f", "favorite"}, {"a", all}, {"?", "all keys"}, {"q", "quit"}}
	default:
		keys = [][2]string{{"enter", "open/play"}, {"/", "search"}, {"esc", "back"}, {"f", "favorite"}, {"a", all}, {"s", "sort"}, {"?", "all keys"}, {"q", "quit"}}
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
			{"space", "pause / resume"}, {"← →", "seek 10 s (shift: 1 min)"}, {"n p", "next / previous surah"},
			{"x", "stop"}, {"- +", "volume"}, {"m", "mute"}, {"[ ]", "slower / faster"}, {"r", "after a surah: autoplay / repeat / stop"},
			{"t", "sleep timer: 15-60 min or end of surah"},
		}},
	}
	var blocks []string
	for _, s := range sections {
		var b strings.Builder
		b.WriteString(band(s.title, min(52, m.w)) + "\n")
		for _, k := range s.keys {
			b.WriteString("   " + sKey.Render(fmt.Sprintf("%-10s", k[0])) + " " + sText.Render(k[1]) + "\n")
		}
		blocks = append(blocks, b.String())
	}
	var body string
	if m.w >= 110 {
		left := lipgloss.JoinVertical(lipgloss.Left, blocks[0], blocks[1])
		body = lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(54).Render(left), blocks[2])
	} else {
		body = lipgloss.JoinVertical(lipgloss.Left, blocks...)
	}
	out := "\n" + sTitle.Render(" ✦ qari "+m.opt.Version) + sDim.Render("   listen to the Quran from mp3quran.net") + "\n\n" +
		body + "\n" +
		sDim.Render(" Keys also work on the Arabic keyboard layout: the key position counts, not the letter.") + "\n" +
		sDim.Render(" Progress is saved automatically. Popularity: way2quran.com, quranicaudio.com, quran.com.") + "\n" +
		sDim.Render(" Press any key to close.")
	lines := strings.Split(out, "\n")
	if len(lines) > m.h { // tiny terminal: keep the top
		lines = lines[:m.h]
	}
	for i, l := range lines {
		lines[i] = ansi.Truncate(l, m.w, "…")
	}
	return strings.Join(lines, "\n")
}

func (m Model) errorView() string {
	msg := m.loadErr.Error()
	return "\n" + sTitle.Render(" ✦ qari") + "\n\n " + sErr.Bold(true).Render("Could not load the catalog from mp3quran.net") + "\n\n " +
		sText.Render(ansi.Wrap(msg, max(m.w-4, 20), "")) + "\n\n " +
		sDim.Render("Some networks block mp3quran.net. Try another network or a proxy, e.g.") + "\n   " +
		sKey.Render("qari --proxy socks5://127.0.0.1:40000") + "\n\n " + sDim.Render("q quits")
}

// showArabic: the Arabic column is redundant when names are already Arabic.
func (m Model) showArabic() bool { return m.arabic && m.cat.Lang != "ar" }
