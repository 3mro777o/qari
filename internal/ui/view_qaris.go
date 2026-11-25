package ui

import (
	"fmt"
	"strings"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/store"
)

// renderQaris renders the Qaris library tab with live filter ribbon and wide master-detail preview.
func (m Model) renderQaris(ch int) []string {
	s := m.theme.Styles
	w := m.width
	items := m.filteredQaris()
	vis := max(1, ch-2)

	// Clamp local view offset safely (View is value receiver).
	ls := m.qarisList
	ls.clamp(len(items), vis)

	// Ribbon line: scope + sort + riwayah on left; dynamic `a` action label on right!
	scopeText := fmt.Sprintf("Well-known · %d", len(items))
	toggleAction := "show all"
	if m.store.Settings.ShowAll {
		scopeText = fmt.Sprintf("All qaris · %d", len(items))
		toggleAction = "well-known only"
	}
	leftParts := []string{" " + s.SectionTitle.Render(scopeText)}
	leftParts = append(leftParts, s.Muted.Render("Sort: ")+s.Text.Render(sortLabel(m.store.Settings.Sort)))
	if m.store.Settings.Rewaya != 0 {
		leftParts = append(leftParts, s.Muted.Render("Riwayah: ")+s.Accent2.Render(riwayaFilterName(m.cat, m.store.Settings.Rewaya)))
	}
	leftRib := strings.Join(leftParts, s.Muted.Render(" · "))

	var rightParts []string
	rightParts = append(rightParts, s.Key.Render("a")+" "+s.Accent.Render(toggleAction))
	if w >= 74 {
		rightParts = append(rightParts, s.Key.Render("s")+" "+s.Muted.Render("sort"))
		rightParts = append(rightParts, s.Key.Render("w")+" "+s.Muted.Render("riwayah"))
	}
	if w >= 92 {
		rightParts = append(rightParts, s.Key.Render("f")+" "+s.Muted.Render("fav"))
	}
	rightRib := strings.Join(rightParts, s.Muted.Render(" · ")) + " "

	ribbon := rowLR(leftRib, rightRib, w)
	rule := s.Rule.Render(strings.Repeat("─", w))

	if len(items) == 0 {
		hint := "Press esc to clear search"
		if !m.store.Settings.ShowAll {
			hint = "Press a to show all qaris · w riwayah · esc clear"
		} else if m.store.Settings.Rewaya != 0 {
			hint = "Press w to choose all riwayat · esc clear"
		}
		title := "No qaris match current filters"
		if m.qarisList.query != "" {
			title = fmt.Sprintf("No qaris match %q", m.qarisList.query)
		}
		out := []string{ribbon, rule}
		out = append(out, m.renderEmptyState(vis, title, hint)...)
		return out
	}

	if w >= 100 {
		leftW := (w * 58) / 100
		rightW := w - leftW - 1
		leftLines := m.renderQariRows(items, ls, leftW, vis)
		var selRec *api.ReciterEntry
		if ls.cursor >= 0 && ls.cursor < len(items) {
			selRec = items[ls.cursor]
		}
		rightLines := m.renderQariPreview(selRec, rightW, vis)
		body := joinColumns(leftLines, rightLines, leftW, rightW, vis, s.Rule.Render("│"))
		return append([]string{ribbon, rule}, body...)
	}

	body := m.renderQariRows(items, ls, w, vis)
	return append([]string{ribbon, rule}, body...)
}

func (m Model) renderQariRows(items []*api.ReciterEntry, ls listState, w, vis int) []string {
	s := m.theme.Styles
	out := make([]string, 0, vis)
	end := min(len(items), ls.offset+vis)

	for i := ls.offset; i < end; i++ {
		r := items[i]
		sel := i == ls.cursor
		rail := "   "
		nameStyle := s.Text
		if sel {
			rail = s.SelectedRail.Render("▌› ")
			nameStyle = s.SelectedText
		}

		favMark := "  "
		if m.store.Favorites[r.ID] {
			favMark = s.Fav.Render("★ ")
		}

		playMark := ""
		if m.playKind == playSurah && m.playTrack.Reciter == r.ID {
			if m.paused {
				playMark = s.Accent2.Render("‖ ")
			} else {
				playMark = s.AccentBold.Render("▶ ")
			}
		}

		left := rail + favMark + playMark + nameStyle.Render(r.Name)

		// Right-aligned contextual metadata
		var meta []string
		if mo, ok := bestMoshaf(m.cat, r, m.store.Settings.Rewaya, 0, false); ok {
			if rw := shortRiwayahName(m.cat, mo); rw != "" && w >= 54 {
				if len(r.Moshaf) > 1 {
					meta = append(meta, s.Muted.Render(fmt.Sprintf("%s +%d", rw, len(r.Moshaf)-1)))
				} else {
					meta = append(meta, s.Muted.Render(rw))
				}
			}
			if !r.Complete {
				meta = append(meta, s.Warn.Render(fmt.Sprintf("%d/114", mo.SurahTotal)))
			}
		}

		switch normalizeSort(m.store.Settings.Sort) {
		case sortUpdated:
			if d := shortDate(r.Date); d != "" {
				meta = append(meta, s.Subtext.Render(d))
			}
		default:
			if lsStr := formatListens(r.Listens); lsStr != "" {
				meta = append(meta, s.Subtext.Render(lsStr))
			}
		}

		right := ""
		if len(meta) > 0 {
			right = strings.Join(meta, s.Muted.Render(" · ")) + " "
		}
		out = append(out, rowLR(left, right, w))
	}

	// Scroll position whisper if list scrolls and there's spare room, else blank lines.
	for len(out) < vis {
		out = append(out, strings.Repeat(" ", w))
	}
	return out
}

func (m Model) renderQariPreview(r *api.ReciterEntry, w, h int) []string {
	s := m.theme.Styles
	if r == nil {
		return nil
	}
	var lines []string
	pad := "  "
	inner := max(8, w-3)

	// Qari Name (display language ONLY — never Arabic!)
	lines = append(lines, pad+s.AccentBold.Render(trunc(r.Name, inner)))

	// Status badges
	var badges []string
	if m.store.Favorites[r.ID] {
		badges = append(badges, s.Fav.Render("★ Favorite"))
	} else {
		badges = append(badges, s.Muted.Render("f favorite"))
	}
	if r.Complete {
		badges = append(badges, s.Success.Render("✓ Complete Quran"))
	} else {
		badges = append(badges, s.Warn.Render("Partial recordings"))
	}
	if ls := formatListens(r.Listens); ls != "" {
		badges = append(badges, s.Subtext.Render(ls+" listens"))
	}
	lines = append(lines, pad+trunc(strings.Join(badges, s.Muted.Render(" · ")), inner))

	// User listening progress with this qari
	fin, prog := reciterProgressCounts(m.store, r.ID)
	if fin > 0 || prog > 0 {
		lines = append(lines, "")
		bar := renderMiniBar(float64(fin)/114.0, min(14, max(6, inner-24)), s.ProgFill, s.ProgEmpty)
		stat := fmt.Sprintf("✓ %d/114 finished", fin)
		if prog > 0 {
			stat += fmt.Sprintf(" · ◔ %d active", prog)
		}
		lines = append(lines, pad+bar+" "+s.Subtext.Render(stat))
	}

	lines = append(lines, "")
	lines = append(lines, pad+s.SectionTitle.Render(fmt.Sprintf("RECORDINGS (%d)", len(r.Moshaf))))

	pref, _ := bestMoshaf(m.cat, r, m.store.Settings.Rewaya, 0, false)
	for _, mo := range r.Moshaf {
		if len(lines) >= h-2 {
			rem := len(r.Moshaf) - (len(lines) - 5)
			if rem > 0 {
				lines = append(lines, pad+s.Muted.Render(fmt.Sprintf(".. and %d more", rem)))
			}
			break
		}
		mark := s.Muted.Render("·")
		lblStyle := s.Text
		if mo.ID == pref.ID {
			mark = s.Accent.Render("✦")
			lblStyle = s.TextBold
		}
		lbl := moshafDisplayLabel(m.cat, mo)
		cnt := s.Subtext.Render(fmt.Sprintf("%d surahs", mo.SurahTotal))
		lines = append(lines, rowLR(pad+mark+" "+lblStyle.Render(lbl), cnt+" ", w))
	}

	// Pin helpful cue near bottom when height permits.
	for len(lines) < h-1 {
		lines = append(lines, "")
	}
	if len(lines) < h {
		lines = append(lines, pad+s.Key.Render("enter")+" "+s.Muted.Render("open recordings & surahs"))
	}
	return lines[:min(len(lines), h)]
}

// renderQariDetail renders a single Qari's recording selector and surah list.
func (m Model) renderQariDetail(ch int) []string {
	s := m.theme.Styles
	w := m.width
	rec := m.detailReciter
	if rec == nil || m.cat == nil {
		return nil
	}

	// Line 1: Breadcrumb + Qari name + favorite toggle
	backLabel := "Qaris"
	if m.returnView == viewSurahDetail {
		backLabel = fmt.Sprintf("Surah %03d", m.returnSurahID)
	} else if m.returnTab == tabHome {
		backLabel = "Home"
	}
	favStar := ""
	if m.store.Favorites[rec.ID] {
		favStar = " " + s.Fav.Render("★")
	}
	h1Left := " " + s.Muted.Render("‹ "+backLabel+" › ") + s.AccentBold.Render(rec.Name) + favStar
	favHint := "favorite"
	if m.store.Favorites[rec.ID] {
		favHint = "unfavorite"
	}
	h1Right := s.Key.Render("f") + " " + s.Muted.Render(favHint) + s.Muted.Render(" · ") + s.Key.Render("esc") + " " + s.Muted.Render("back ")
	line1 := rowLR(h1Left, h1Right, w)

	// Line 2: Active Recording (Moshaf) selector + completion progress
	mo := m.detailMoshaf
	moIdx := 1
	for i, x := range rec.Moshaf {
		if x.ID == mo.ID {
			moIdx = i + 1
			break
		}
	}
	recLbl := moshafDisplayLabel(m.cat, mo)
	h2Left := " " + s.Accent.Render("✦") + " " + s.TextBold.Render(recLbl) + s.Muted.Render(fmt.Sprintf(" (%d surahs)", mo.SurahTotal))
	if len(rec.Moshaf) > 1 {
		h2Left += "  " + s.Accent2.Render(fmt.Sprintf("[%d/%d", moIdx, len(rec.Moshaf))) + " " + s.Key.Render("e") + " " + s.Muted.Render("switch") + s.Accent2.Render("]")
	}
	fin, _ := moshafProgressCounts(m.store, rec.ID, mo.ID)
	total := max(1, mo.SurahTotal)
	h2Right := ""
	if w >= 72 {
		bar := renderMiniBar(float64(fin)/float64(total), 8, s.Success, s.ProgEmpty)
		h2Right = bar + " " + s.Subtext.Render(fmt.Sprintf("%d/%d", fin, mo.SurahTotal)) + " "
	} else {
		h2Right = s.Subtext.Render(fmt.Sprintf("%d/%d ", fin, mo.SurahTotal))
	}
	line2 := rowLR(h2Left, h2Right, w)

	// Line 3: Hairline rule
	line3 := s.Rule.Render(strings.Repeat("─", w))

	vis := max(1, ch-3)
	surahs := m.filteredQariDetailSurahs()
	ls := m.qariDetailList
	ls.clamp(len(surahs), vis)

	if len(surahs) == 0 {
		out := []string{line1, line2, line3}
		out = append(out, m.renderEmptyState(vis,
			fmt.Sprintf("No surahs match %q", m.qariDetailList.query),
			"Search by English name, Arabic name, or surah number (e.g. 18) · esc clear",
		)...)
		return out
	}

	out := make([]string, 0, ch)
	out = append(out, line1, line2, line3)

	end := min(len(surahs), ls.offset+vis)
	for i := ls.offset; i < end; i++ {
		su := surahs[i]
		sel := i == ls.cursor
		tr := store.Track{Reciter: rec.ID, Moshaf: mo.ID, Surah: su.ID}
		prog := m.store.Get(tr)
		isPlaying := m.playKind == playSurah && m.playTrack == tr

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
		} else if prog != nil && prog.Finished {
			status = s.Success.Render("✓")
		} else if prog != nil && prog.Position >= 5 {
			status = s.Accent2.Render("◔")
		}

		num := s.Muted.Render(fmt.Sprintf("%03d", su.ID))
		if sel || isPlaying {
			num = s.Accent.Render(fmt.Sprintf("%03d", su.ID))
		}

		// Right metadata: live/saved progress bar if active/in-progress, else revelation & pages.
		var right string
		if isPlaying && m.playDur > 0 {
			bar := renderMiniBar(m.playPos/m.playDur, 8, s.ProgFill, s.ProgEmpty)
			right = bar + " " + s.AccentBold.Render(formatDuration(m.playPos)+"/"+formatDuration(m.playDur)) + " "
		} else if prog != nil && !prog.Finished && prog.Position >= 5 {
			if prog.Duration > 0 {
				bar := renderMiniBar(prog.Fraction(), 8, s.ProgFill, s.ProgEmpty)
				right = bar + " " + s.Subtext.Render(formatDuration(prog.Position)+"/"+formatDuration(prog.Duration)) + " "
			} else {
				right = s.Subtext.Render(formatDuration(prog.Position)) + " "
			}
		} else if prog != nil && prog.Finished {
			meta := s.Success.Render("✓ heard")
			if w >= 80 {
				meta = s.Muted.Render(revelationLabel(su.Makkia)+" · ") + meta
			}
			right = meta + " "
		} else {
			if w >= 74 {
				pages := surahPagesLabel(su)
				if pages != "" {
					right = s.Muted.Render(revelationLabel(su.Makkia)+" · "+pages) + " "
				} else {
					right = s.Muted.Render(revelationLabel(su.Makkia)) + " "
				}
			} else {
				right = s.Muted.Render(revelationLabel(su.Makkia)) + " "
			}
		}

		arabicPart := ""
		nameField := nameStyle.Render(su.Name)
		if !m.noArabic && su.Arabic != "" && w >= 60 {
			arStyle := s.Arabic
			if sel {
				arStyle = s.ArabicSel
			}
			nameCol := 18
			if w < 72 {
				nameCol = 15
			}
			nameField = padRight(nameStyle.Render(trunc(su.Name, nameCol)), nameCol)
			arabicPart = "  " + arStyle.Render(fitArabic(su.Arabic, 20))
		}

		left := rail + status + " " + num + "  " + nameField + arabicPart
		out = append(out, rowLR(left, right, w))
	}

	return out
}
