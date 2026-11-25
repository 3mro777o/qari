package ui

import (
	"fmt"
	"strings"

	"github.com/3mro777o/qari/internal/api"
)

// renderHome renders the personal listening sanctuary: Continue Listening, Favorites, Most Listened.
func (m Model) renderHome(ch int) []string {
	if m.cat == nil {
		return nil
	}
	items := m.homeItems()
	if len(items) == 0 {
		return m.renderEmptyState(ch,
			fmt.Sprintf("No matches for %q", m.homeList.query),
			"Press esc to clear search, or switch to 2 Qaris / 3 Surahs",
		)
	}

	if m.width >= 96 {
		return m.renderHomeWide(items, ch)
	}
	return m.renderHomeCompact(items, ch)
}

func (m Model) renderHomeWide(items []homeItem, ch int) []string {
	s := m.theme.Styles
	leftW := (m.width * 54) / 100
	rightW := m.width - leftW - 1
	div := s.Rule.Render("│")

	// Partition items into Left (Continue + Favorites) and Right (Most Listened), preserving global indices.
	type indexed struct {
		idx  int
		item homeItem
	}
	var cont, favs, pops []indexed
	for i, it := range items {
		switch it.kind {
		case homeContinue:
			cont = append(cont, indexed{i, it})
		case homeFavorite:
			favs = append(favs, indexed{i, it})
		case homePopular:
			pops = append(pops, indexed{i, it})
		}
	}

	// Build Left Column rows and track cursor line for vertical scrolling if needed.
	type colRow struct {
		text  string
		isSel bool
	}
	var leftRows []colRow
	tallCards := ch >= 18

	// Section: Continue Listening
	contHead := " " + s.SectionTitle.Render("✦ CONTINUE LISTENING")
	contSub := ""
	if len(cont) > 0 {
		contSub = s.Muted.Render("enter to resume ")
	}
	leftRows = append(leftRows, colRow{text: rowLR(contHead, contSub, leftW)})

	if len(cont) == 0 {
		leftRows = append(leftRows, colRow{text: "   " + s.TextBold.Render("Welcome to qari")})
		leftRows = append(leftRows, colRow{text: "   " + s.Subtext.Render("Pick a qari on the right, or browse all 114 surahs.")})
		leftRows = append(leftRows, colRow{text: "   " + s.Muted.Render("Your place in every surah is saved automatically.")})
		if ch >= 20 {
			leftRows = append(leftRows, colRow{text: ""})
			leftRows = append(leftRows, colRow{text: "   " + s.Key.Render("2") + " " + s.Subtext.Render("Qaris") + s.Muted.Render(" · ") + s.Key.Render("3") + " " + s.Subtext.Render("Surahs") + s.Muted.Render(" · ") + s.Key.Render("4") + " " + s.Subtext.Render("Radio") + s.Muted.Render(" · ") + s.Key.Render("/") + " " + s.Subtext.Render("Search")})
		}
	} else {
		for _, ci := range cont {
			sel := ci.idx == m.homeList.cursor
			if tallCards {
				l1, l2 := m.renderContinueCard(ci.item, leftW, sel)
				leftRows = append(leftRows, colRow{text: l1, isSel: sel})
				leftRows = append(leftRows, colRow{text: l2, isSel: sel})
			} else {
				leftRows = append(leftRows, colRow{text: m.renderContinueRow(ci.item, leftW, sel), isSel: sel})
			}
		}
	}

	leftRows = append(leftRows, colRow{text: ""})

	// Section: Favorite Qaris
	favHead := " " + s.Fav.Render("★") + " " + s.SectionTitle.Render("FAVORITE QARIS")
	favSub := s.Muted.Render(fmt.Sprintf("%d pinned ", len(favs)))
	leftRows = append(leftRows, colRow{text: rowLR(favHead, favSub, leftW)})
	if len(favs) == 0 {
		leftRows = append(leftRows, colRow{text: "   " + s.Muted.Render("Press f on any qari to pin them here")})
	} else {
		for _, fi := range favs {
			sel := fi.idx == m.homeList.cursor
			leftRows = append(leftRows, colRow{text: m.renderHomeQariRow(fi.item.reciter, 0, leftW, sel), isSel: sel})
		}
	}

	// Build Right Column: Most Listened Qaris
	var rightRows []colRow
	scopeAction := "a show all"
	if m.store.Settings.ShowAll {
		scopeAction = "a well-known only"
	}
	popHead := " " + s.SectionTitle.Render("✦ MOST LISTENED")
	popRight := s.Muted.Render(scopeAction + " ")
	rightRows = append(rightRows, colRow{text: rowLR(popHead, popRight, rightW)})

	for rank, pi := range pops {
		sel := pi.idx == m.homeList.cursor
		rightRows = append(rightRows, colRow{text: m.renderHomeQariRow(pi.item.reciter, rank+1, rightW, sel), isSel: sel})
	}

	// Add subtle footer summary on right column when there is room.
	finTotal := 0
	for _, p := range m.store.Progress {
		if p.Finished {
			finTotal++
		}
	}
	if len(rightRows)+2 <= ch {
		rightRows = append(rightRows, colRow{text: ""})
		favWord := "favorites"
		if len(m.store.Favorites) == 1 {
			favWord = "favorite"
		}
		summary := fmt.Sprintf(" %d qaris · %d %s", len(m.cat.Reciters), len(m.store.Favorites), favWord)
		if finTotal > 0 {
			sWord := "surahs"
			if finTotal == 1 {
				sWord = "surah"
			}
			summary += fmt.Sprintf(" · %d %s finished", finTotal, sWord)
		}
		rightRows = append(rightRows, colRow{text: s.Muted.Render(summary)})
	}

	sliceWindow := func(rows []colRow, height int) []string {
		if len(rows) <= height {
			out := make([]string, len(rows))
			for i, r := range rows {
				out[i] = r.text
			}
			return out
		}
		selRow := -1
		for i, r := range rows {
			if r.isSel {
				selRow = i
				break
			}
		}
		start := 0
		if selRow >= 0 {
			// Keep header row 0 pinned if near top, else center selection cleanly.
			if selRow >= height-2 {
				start = min(len(rows)-height, max(0, selRow-height/2))
			}
		}
		out := make([]string, height)
		for i := 0; i < height; i++ {
			out[i] = rows[start+i].text
		}
		return out
	}

	leftLines := sliceWindow(leftRows, ch)
	rightLines := sliceWindow(rightRows, ch)
	return joinColumns(leftLines, rightLines, leftW, rightW, ch, div)
}

func (m Model) renderHomeCompact(items []homeItem, ch int) []string {
	s := m.theme.Styles
	w := m.width

	type row struct {
		text  string
		isSel bool
	}
	var rows []row
	lastKind := homeItemKind(-1)

	for i, it := range items {
		if it.kind != lastKind {
			if lastKind != -1 {
				rows = append(rows, row{text: ""})
			}
			switch it.kind {
			case homeContinue:
				rows = append(rows, row{text: rowLR(" "+s.SectionTitle.Render("✦ CONTINUE LISTENING"), s.Muted.Render("enter resume "), w)})
			case homeFavorite:
				rows = append(rows, row{text: " " + s.Fav.Render("★") + " " + s.SectionTitle.Render("FAVORITE QARIS")})
			case homePopular:
				lbl := "a show all"
				if m.store.Settings.ShowAll {
					lbl = "a well-known only"
				}
				rows = append(rows, row{text: rowLR(" "+s.SectionTitle.Render("✦ MOST LISTENED"), s.Muted.Render(lbl+" "), w)})
			}
			lastKind = it.kind
		}
		sel := i == m.homeList.cursor
		switch it.kind {
		case homeContinue:
			rows = append(rows, row{text: m.renderContinueRow(it, w, sel), isSel: sel})
		default:
			rows = append(rows, row{text: m.renderHomeQariRow(it.reciter, 0, w, sel), isSel: sel})
		}
	}

	if len(rows) <= ch {
		out := make([]string, len(rows))
		for i, r := range rows {
			out[i] = r.text
		}
		return out
	}

	selRow := 0
	for i, r := range rows {
		if r.isSel {
			selRow = i
			break
		}
	}
	start := 0
	if selRow >= ch-1 {
		start = min(len(rows)-ch, max(0, selRow-ch/2))
	}
	out := make([]string, ch)
	for i := 0; i < ch; i++ {
		out[i] = rows[start+i].text
	}
	return out
}

func (m Model) renderContinueCard(it homeItem, w int, sel bool) (string, string) {
	s := m.theme.Styles
	p := it.progress
	surahName := m.cat.SurahName(p.Surah)
	isPlaying := m.playKind == playSurah && m.playTrack == p.Track

	rail1 := "   "
	rail2 := "   "
	nameStyle := s.TextBold
	if sel {
		rail1 = s.SelectedRail.Render("▌› ")
		rail2 = s.SelectedRail.Render("▌  ")
		nameStyle = s.SelectedText
	}

	icon := s.Accent2.Render("◔")
	if isPlaying {
		if m.paused {
			icon = s.Accent2.Render("‖")
		} else {
			icon = s.AccentBold.Render("▶")
		}
	} else if p.Finished {
		icon = s.Success.Render("✓")
	}

	arabicPart := ""
	if !m.noArabic && w >= 46 && p.Surah >= 1 && p.Surah <= len(m.cat.Surahs) {
		if ar := m.cat.Surahs[p.Surah-1].Arabic; ar != "" {
			arStyle := s.Arabic
			if sel {
				arStyle = s.ArabicSel
			}
			arabicPart = "  " + arStyle.Render(fitArabic(ar, 16))
		}
	}
	left1 := rail1 + icon + " " + s.Muted.Render(fmt.Sprintf("%03d", p.Surah)) + " " + nameStyle.Render(surahName) + arabicPart

	var right1 string
	if p.Finished {
		right1 = s.Success.Render("finished") + " "
	} else if p.Duration > 0 {
		rem := max(0, p.Duration-p.Position)
		right1 = s.Subtext.Render(formatDuration(rem)+" left") + " "
	} else if p.Position > 0 {
		right1 = s.Subtext.Render(formatDuration(p.Position)) + " "
	}

	recName := ""
	moshafLbl := ""
	if it.reciter != nil {
		recName = it.reciter.Name
		if mo, ok := findMoshafByID(it.reciter, p.Moshaf); ok {
			moshafLbl = shortRiwayahName(m.cat, mo)
		}
	}
	left2 := rail2 + "  " + s.Subtext.Render(recName)
	if moshafLbl != "" {
		left2 += s.Muted.Render(" · " + moshafLbl)
	}
	right2 := ""
	if p.Duration > 0 || p.Finished {
		pct := int(p.Fraction()*100 + 0.5)
		right2 = renderMiniBar(p.Fraction(), 8, s.ProgFill, s.ProgEmpty) + " " + s.Muted.Render(fmt.Sprintf("%3d%% ", pct))
	}

	return rowLR(left1, right1, w), rowLR(left2, right2, w)
}

func (m Model) renderContinueRow(it homeItem, w int, sel bool) string {
	s := m.theme.Styles
	p := it.progress
	surahName := m.cat.SurahName(p.Surah)
	rail := "   "
	nameStyle := s.TextBold
	if sel {
		rail = s.SelectedRail.Render("▌› ")
		nameStyle = s.SelectedText
	}
	icon := s.Accent2.Render("◔")
	if m.playKind == playSurah && m.playTrack == p.Track {
		if m.paused {
			icon = s.Accent2.Render("‖")
		} else {
			icon = s.AccentBold.Render("▶")
		}
	} else if p.Finished {
		icon = s.Success.Render("✓")
	}
	recName := ""
	if it.reciter != nil {
		recName = it.reciter.Name
	}
	left := rail + icon + " " + s.Muted.Render(fmt.Sprintf("%03d", p.Surah)) + " " + nameStyle.Render(surahName) + s.Muted.Render(" · ") + s.Subtext.Render(recName)
	right := ""
	if p.Duration > 0 {
		right = s.Subtext.Render(formatDuration(p.Position)+"/"+formatDuration(p.Duration)) + " "
	}
	return rowLR(left, right, w)
}

func (m Model) renderHomeQariRow(r *api.ReciterEntry, rank int, w int, sel bool) string {
	s := m.theme.Styles
	if r == nil {
		return strings.Repeat(" ", w)
	}
	rail := "   "
	nameStyle := s.Text
	if sel {
		rail = s.SelectedRail.Render("▌› ")
		nameStyle = s.SelectedText
	}
	prefix := rail
	if rank > 0 {
		prefix += s.Muted.Render(fmt.Sprintf("%02d ", rank))
	}
	if m.store.Favorites[r.ID] {
		prefix += s.Fav.Render("★ ")
	} else if rank == 0 {
		prefix += "  "
	}

	left := prefix + nameStyle.Render(r.Name)

	var meta []string
	if m.playKind == playSurah && m.playTrack.Reciter == r.ID {
		meta = append(meta, s.AccentBold.Render("▶"))
	}
	if len(r.Moshaf) > 1 {
		meta = append(meta, s.Muted.Render(fmt.Sprintf("%d rec", len(r.Moshaf))))
	} else if len(r.Moshaf) == 1 {
		if rw := shortRiwayahName(m.cat, r.Moshaf[0]); rw != "" {
			meta = append(meta, s.Muted.Render(rw))
		}
	}
	if ls := formatListens(r.Listens); ls != "" {
		meta = append(meta, s.Subtext.Render(ls))
	}
	right := ""
	if len(meta) > 0 {
		right = strings.Join(meta, s.Muted.Render(" · ")) + " "
	}
	return rowLR(left, right, w)
}

func (m Model) renderEmptyState(ch int, title, subtitle string) []string {
	s := m.theme.Styles
	w := m.width
	top := max(0, (ch-3)/2)
	lines := make([]string, 0, ch)
	blank := strings.Repeat(" ", w)
	for i := 0; i < top; i++ {
		lines = append(lines, blank)
	}
	lines = append(lines, center(s.Accent.Render("✦")+"  "+s.TextBold.Render(title), w))
	lines = append(lines, blank)
	lines = append(lines, center(s.Subtext.Render(subtitle), w))
	for len(lines) < ch {
		lines = append(lines, blank)
	}
	return lines[:ch]
}
