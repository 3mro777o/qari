package ui

import (
	"fmt"
	"strings"

	"github.com/3mro777o/qari/internal/api"
)

// renderRadio renders the Live Quran Radio browser tab.
func (m Model) renderRadio(ch int) []string {
	s := m.theme.Styles
	w := m.width
	items := m.filteredRadios()
	vis := max(1, ch-2)

	ls := m.radioList
	ls.clamp(len(items), vis)

	title := fmt.Sprintf("Live Quran Radio · %d stations", len(items))
	rightRib := s.Key.Render("enter") + " " + s.Muted.Render("listen")
	if m.playKind == playRadio {
		rightRib = s.Live.Render("● LIVE") + s.Muted.Render(" · ") + s.Key.Render("x") + " " + s.Muted.Render("stop")
	}
	rightRib += " "

	ribbon := rowLR(" "+s.SectionTitle.Render(title), rightRib, w)
	rule := s.Rule.Render(strings.Repeat("─", w))

	if len(items) == 0 {
		msg := "No radio stations available"
		hint := "Check your connection or try another language"
		if m.radioList.query != "" {
			msg = fmt.Sprintf("No stations match %q", m.radioList.query)
			hint = "Press esc to clear search"
		}
		out := []string{ribbon, rule}
		out = append(out, m.renderEmptyState(vis, msg, hint)...)
		return out
	}

	if w >= 100 {
		leftW := (w * 62) / 100
		rightW := w - leftW - 1
		leftLines := m.renderRadioRows(items, ls, leftW, vis)
		var selRadio api.Radio
		if ls.cursor >= 0 && ls.cursor < len(items) {
			selRadio = items[ls.cursor]
		}
		rightLines := m.renderRadioPreview(selRadio, rightW, vis)
		body := joinColumns(leftLines, rightLines, leftW, rightW, vis, s.Rule.Render("│"))
		return append([]string{ribbon, rule}, body...)
	}

	body := m.renderRadioRows(items, ls, w, vis)
	return append([]string{ribbon, rule}, body...)
}

func (m Model) renderRadioRows(items []api.Radio, ls listState, w, vis int) []string {
	s := m.theme.Styles
	out := make([]string, 0, vis)
	end := min(len(items), ls.offset+vis)

	for i := ls.offset; i < end; i++ {
		r := items[i]
		sel := i == ls.cursor
		isPlaying := m.playKind == playRadio && m.playRadio.URL == r.URL

		rail := "   "
		nameStyle := s.Text
		if sel {
			rail = s.SelectedRail.Render("▌› ")
			nameStyle = s.SelectedText
		}

		icon := s.Muted.Render("♪")
		right := ""
		if isPlaying {
			if m.paused {
				icon = s.Accent2.Render("‖")
				right = s.Accent2.Render("paused ")
			} else {
				icon = s.Live.Render("●")
				right = s.Live.Render("● LIVE ")
			}
		}

		left := rail + icon + "  " + nameStyle.Render(r.Name)
		out = append(out, rowLR(left, right, w))
	}
	for len(out) < vis {
		out = append(out, strings.Repeat(" ", w))
	}
	return out
}

func (m Model) renderRadioPreview(selRadio api.Radio, w, h int) []string {
	s := m.theme.Styles
	pad := "  "
	inner := max(8, w-3)
	var lines []string

	if m.playKind == playRadio {
		lines = append(lines, pad+s.Live.Render("● NOW STREAMING"))
		for _, wl := range wrapWords(m.playRadio.Name, inner, 2) {
			lines = append(lines, pad+s.TextBold.Render(wl))
		}
		lines = append(lines, "")
	}

	lines = append(lines, pad+s.SectionTitle.Render("SELECTED STATION"))
	for _, wl := range wrapWords(selRadio.Name, inner, 3) {
		lines = append(lines, pad+s.Text.Render(wl))
	}
	lines = append(lines, "")
	lines = append(lines, pad+s.Muted.Render("Continuous live Quran broadcast."))
	lines = append(lines, pad+s.Muted.Render("Live streams do not affect surah progress."))

	lines = append(lines, "")
	lines = append(lines, pad+s.Key.Render("enter")+"  "+s.Subtext.Render("Tune in to station"))
	lines = append(lines, pad+s.Key.Render("space")+"  "+s.Subtext.Render("Pause / resume stream"))
	lines = append(lines, pad+s.Key.Render("x")+"      "+s.Subtext.Render("Stop stream"))
	lines = append(lines, pad+s.Key.Render("+/- m")+"  "+s.Subtext.Render("Volume & mute"))

	if len(lines) > h {
		lines = lines[:h]
	}
	return lines
}
