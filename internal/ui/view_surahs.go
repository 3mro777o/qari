package ui

import (
	"fmt"
	"strings"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/store"
)

// renderSurahs renders the 114 Surahs browser with optional wide preview panel.
func (m Model) renderSurahs(ch int) []string {
	s := m.theme.Styles
	w := m.width
	items := m.filteredSurahs()
	vis := max(1, ch-2)

	ls := m.surahsList
	ls.clamp(len(items), vis)

	title := fmt.Sprintf("Surahs · %d", len(items))
	if m.surahsList.query != "" {
		title = fmt.Sprintf("Surahs · %d of 114", len(items))
	}
	arabAction := "hide Arabic"
	if m.noArabic {
		arabAction = "show Arabic"
	}
	rightRib := s.Key.Render("A") + " " + s.Muted.Render(arabAction)
	if w >= 72 {
		rightRib += s.Muted.Render(" · ") + s.Key.Render("enter") + " " + s.Muted.Render("choose qari")
	}
	rightRib += " "

	ribbon := rowLR(" "+s.SectionTitle.Render(title), rightRib, w)
	rule := s.Rule.Render(strings.Repeat("─", w))

	if len(items) == 0 {
		out := []string{ribbon, rule}
		out = append(out, m.renderEmptyState(vis,
			fmt.Sprintf("No surahs match %q", m.surahsList.query),
			"Try English name, Arabic name, or surah number (1-114) · esc clear",
		)...)
		return out
	}

	if w >= 100 {
		leftW := (w * 58) / 100
		rightW := w - leftW - 1
		leftLines := m.renderSurahRows(items, ls, leftW, vis)
		var selSurah api.SurahEntry
		if ls.cursor >= 0 && ls.cursor < len(items) {
			selSurah = items[ls.cursor]
		}
		rightLines := m.renderSurahPreview(selSurah, rightW, vis)
		body := joinColumns(leftLines, rightLines, leftW, rightW, vis, s.Rule.Render("│"))
		return append([]string{ribbon, rule}, body...)
	}

	body := m.renderSurahRows(items, ls, w, vis)
	return append([]string{ribbon, rule}, body...)
}

func (m Model) renderSurahRows(items []api.SurahEntry, ls listState, w, vis int) []string {
	s := m.theme.Styles
	out := make([]string, 0, vis)
	end := min(len(items), ls.offset+vis)

	for i := ls.offset; i < end; i++ {
		su := items[i]
		sel := i == ls.cursor
		fin, inProg, _ := surahProgressSummary(m.store, su.ID)
		isPlaying := m.playKind == playSurah && m.playTrack.Surah == su.ID

		rail := "   "
		nameStyle := s.Text
		if sel {
			rail = s.SelectedRail.Render("▌› ")
			nameStyle = s.SelectedText
		}

		status := s.Muted.Render("·")
		if isPlaying {
			if m.paused {
				status = s.Accent2.Render("‖")
			} else {
				status = s.AccentBold.Render("▶")
			}
		} else if fin {
			status = s.Success.Render("✓")
		} else if inProg {
			status = s.Accent2.Render("◔")
		}

		num := s.Muted.Render(fmt.Sprintf("%03d", su.ID))
		if sel || isPlaying {
			num = s.Accent.Render(fmt.Sprintf("%03d", su.ID))
		}

		arabicPart := ""
		nameField := nameStyle.Render(su.Name)
		if !m.noArabic && su.Arabic != "" && w >= 54 {
			arStyle := s.Arabic
			if sel {
				arStyle = s.ArabicSel
			}
			nameCol := 17
			if w < 64 {
				nameCol = 14
			}
			nameField = padRight(nameStyle.Render(trunc(su.Name, nameCol)), nameCol)
			arabicPart = "  " + arStyle.Render(fitArabic(su.Arabic, 20))
		}

		left := rail + status + " " + num + "  " + nameField + arabicPart
		right := s.Muted.Render(revelationLabel(su.Makkia))
		if pages := surahPagesLabel(su); pages != "" && w >= 66 {
			right = s.Muted.Render(revelationLabel(su.Makkia) + " · " + pages)
		}
		right += " "

		out = append(out, rowLR(left, right, w))
	}
	for len(out) < vis {
		out = append(out, strings.Repeat(" ", w))
	}
	return out
}

func (m Model) renderSurahPreview(su api.SurahEntry, w, h int) []string {
	s := m.theme.Styles
	if su.ID == 0 || m.cat == nil {
		return nil
	}
	pad := "  "
	inner := max(8, w-3)
	var lines []string

	lines = append(lines, pad+s.AccentBold.Render(fmt.Sprintf("%03d · %s", su.ID, su.Name)))
	if !m.noArabic && su.Arabic != "" {
		lines = append(lines, pad+s.ArabicSel.Render(fitArabic(su.Arabic, inner)))
	}

	meta := revelationLabel(su.Makkia)
	if su.StartPage > 0 {
		if su.EndPage <= su.StartPage {
			meta += fmt.Sprintf(" · Page %d (1 page)", su.StartPage)
		} else {
			pCount := su.EndPage - su.StartPage + 1
			meta += fmt.Sprintf(" · Pages %d-%d (%d pages)", su.StartPage, su.EndPage, pCount)
		}
	}
	lines = append(lines, pad+s.Subtext.Render(trunc(meta, inner)))

	// Show user listening status on this surah
	fin, inProg, frac := surahProgressSummary(m.store, su.ID)
	if fin {
		lines = append(lines, pad+s.Success.Render("✓ Completed"))
	} else if inProg {
		bar := renderMiniBar(frac, 10, s.ProgFill, s.ProgEmpty)
		lines = append(lines, pad+bar+" "+s.Accent2.Render(fmt.Sprintf("◔ %d%% heard", int(frac*100+0.5))))
	}

	qaris := filterQarisForSurah(m.cat, m.store.Favorites, su.ID, m.store.Settings.ShowAll, m.store.Settings.Rewaya, sortPopular, "")
	lines = append(lines, "")
	lines = append(lines, pad+s.SectionTitle.Render(fmt.Sprintf("TOP QARIS (%d AVAILABLE)", len(qaris))))

	limit := min(len(qaris), max(1, h-len(lines)-2))
	for i := 0; i < limit; i++ {
		q := qaris[i]
		star := "  "
		if m.store.Favorites[q.Reciter.ID] {
			star = s.Fav.Render("★ ")
		}
		rw := shortRiwayahName(m.cat, q.Moshaf)
		lines = append(lines, rowLR(pad+star+s.Text.Render(q.Reciter.Name), s.Muted.Render(rw)+" ", w))
	}

	for len(lines) < h-1 {
		lines = append(lines, "")
	}
	if len(lines) < h {
		lines = append(lines, pad+s.Key.Render("enter")+" "+s.Muted.Render("choose a qari & play directly"))
	}
	return lines[:min(len(lines), h)]
}

// renderSurahDetail shows all qaris who recorded detailSurahID, ready to play directly on Enter.
func (m Model) renderSurahDetail(ch int) []string {
	s := m.theme.Styles
	w := m.width
	if m.cat == nil || m.detailSurahID < 1 || m.detailSurahID > len(m.cat.Surahs) {
		return nil
	}
	su := m.cat.Surahs[m.detailSurahID-1]

	// Line 1: Surah Masthead
	arabicPart := ""
	if !m.noArabic && su.Arabic != "" && w >= 66 {
		arabicPart = "  " + s.ArabicSel.Render(fitArabic(su.Arabic, 20))
	}
	h1Left := " " + s.Muted.Render("‹ Surahs › ") + s.AccentBold.Render(fmt.Sprintf("%03d · %s", su.ID, su.Name)) + arabicPart
	h1Right := s.Muted.Render(revelationLabel(su.Makkia))
	if pages := surahPagesLabel(su); pages != "" && w >= 76 {
		h1Right += s.Muted.Render(" · " + pages)
	}
	h1Right += s.Muted.Render(" · ") + s.Key.Render("esc") + " " + s.Muted.Render("back ")
	line1 := rowLR(h1Left, h1Right, w)

	// Line 2: Qari Filter Ribbon (same consistent controls!)
	items := m.filteredSurahDetailQaris()
	scopeLbl := fmt.Sprintf("%d qaris (well-known)", len(items))
	toggleAction := "show all"
	if m.store.Settings.ShowAll {
		scopeLbl = fmt.Sprintf("%d qaris (all)", len(items))
		toggleAction = "well-known only"
	}
	lParts := []string{" " + s.TextBold.Render(scopeLbl), s.Muted.Render("Sort: ") + s.Text.Render(sortLabel(m.store.Settings.Sort))}
	if m.store.Settings.Rewaya != 0 {
		lParts = append(lParts, s.Muted.Render("Riwayah: ")+s.Accent2.Render(riwayaFilterName(m.cat, m.store.Settings.Rewaya)))
	}
	rParts := []string{s.Key.Render("a") + " " + s.Accent.Render(toggleAction)}
	if w >= 78 {
		rParts = append(rParts, s.Key.Render("s")+" "+s.Muted.Render("sort"))
		rParts = append(rParts, s.Key.Render("w")+" "+s.Muted.Render("riwayah"))
		rParts = append(rParts, s.Key.Render("i")+" "+s.Muted.Render("qari page"))
	}
	line2 := rowLR(strings.Join(lParts, s.Muted.Render(" · ")), strings.Join(rParts, s.Muted.Render(" · "))+" ", w)
	line3 := s.Rule.Render(strings.Repeat("─", w))

	vis := max(1, ch-3)
	ls := m.surahDetailList
	ls.clamp(len(items), vis)

	if len(items) == 0 {
		hint := "Press esc to clear search"
		if !m.store.Settings.ShowAll {
			hint = "Press a to show all qaris · w riwayah · esc clear"
		} else if m.store.Settings.Rewaya != 0 {
			hint = "Press w to choose all riwayat · esc clear"
		}
		out := []string{line1, line2, line3}
		out = append(out, m.renderEmptyState(vis, "No qaris match for this surah", hint)...)
		return out
	}

	out := make([]string, 0, ch)
	out = append(out, line1, line2, line3)

	end := min(len(items), ls.offset+vis)
	for i := ls.offset; i < end; i++ {
		it := items[i]
		r := it.Reciter
		mo := it.Moshaf
		sel := i == ls.cursor

		tr := store.Track{Reciter: r.ID, Moshaf: mo.ID, Surah: su.ID}
		prog := m.store.Get(tr)
		isPlaying := m.playKind == playSurah && m.playTrack == tr

		rail := "   "
		nameStyle := s.Text
		if sel {
			rail = s.SelectedRail.Render("▌› ")
			nameStyle = s.SelectedText
		}

		mark := "  "
		switch {
		case isPlaying && m.paused:
			mark = s.Accent2.Render("‖ ")
		case isPlaying:
			mark = s.AccentBold.Render("▶ ")
		case prog != nil && prog.Finished:
			mark = s.Success.Render("✓ ")
		case prog != nil && prog.Position >= 5:
			mark = s.Accent2.Render("◔ ")
		case m.store.Favorites[r.ID]:
			mark = s.Fav.Render("★ ")
		}
		favExtra := ""
		if m.store.Favorites[r.ID] && (isPlaying || (prog != nil && (prog.Finished || prog.Position >= 5))) {
			favExtra = s.Fav.Render("★ ")
		}

		left := rail + mark + favExtra + nameStyle.Render(r.Name)

		var meta []string
		if lbl := moshafCompactLabel(m.cat, mo); lbl != "" {
			meta = append(meta, s.Subtext.Render(lbl))
		}
		if prog != nil && !prog.Finished && prog.Position >= 5 && prog.Duration > 0 {
			meta = append(meta, s.Accent2.Render(fmt.Sprintf("%d%%", int(prog.Fraction()*100+0.5))))
		} else if lsStr := formatListens(r.Listens); lsStr != "" && w >= 72 {
			meta = append(meta, s.Muted.Render(lsStr))
		}
		right := strings.Join(meta, s.Muted.Render(" · ")) + " "

		out = append(out, rowLR(left, right, w))
	}
	return out
}
