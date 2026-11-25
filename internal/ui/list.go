package ui

import (
	"strings"

	"github.com/3mro777o/qari/internal/norm"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// row is one line of a list. Header rows are section titles the cursor skips.
type row struct {
	mark   string // left gutter, e.g. a star or progress glyph
	markSt lipgloss.Style
	title  string
	sub    string // dim text after the title
	meta   string // right-aligned, before the Arabic name
	metaSt lipgloss.Style
	arabic string
	header bool
	search string // folded text the query matches against
	data   any
}

// listView is a scrollable, searchable list.
type listView struct {
	rows    []row
	visible []int // indexes into rows that match the query
	cursor  int   // index into visible
	offset  int
	query   string
	empty   string // shown when nothing matches
}

func newList(rows []row, empty string) *listView {
	l := &listView{empty: empty}
	l.setRows(rows)
	return l
}

// setRows replaces the rows, keeping the cursor on the same data if possible.
func (l *listView) setRows(rows []row) {
	var keep any
	if r := l.selected(); r != nil {
		keep = r.data
	}
	l.rows = rows
	l.refilter()
	if keep != nil {
		for i, idx := range l.visible {
			if l.rows[idx].data == keep {
				l.cursor = i
				break
			}
		}
	}
}

func (l *listView) setQuery(q string) {
	l.query = q
	l.cursor, l.offset = 0, 0
	l.refilter()
}

func (l *listView) refilter() {
	l.visible = l.visible[:0]
	q := strings.TrimSpace(l.query)
	for i, r := range l.rows {
		if q != "" && (r.header || !norm.Match(r.search, q)) {
			continue
		}
		l.visible = append(l.visible, i)
	}
	l.cursor = min(l.cursor, max(len(l.visible)-1, 0))
	l.skipHeaders(1)
}

func (l *listView) selected() *row {
	if l.cursor < 0 || l.cursor >= len(l.visible) {
		return nil
	}
	r := &l.rows[l.visible[l.cursor]]
	if r.header {
		return nil
	}
	return r
}

func (l *listView) skipHeaders(dir int) {
	for l.cursor >= 0 && l.cursor < len(l.visible) && l.rows[l.visible[l.cursor]].header {
		l.cursor += dir
	}
	if l.cursor >= len(l.visible) || l.cursor < 0 {
		// ran off the end: search the other way
		l.cursor = min(max(l.cursor, 0), len(l.visible)-1)
		for l.cursor >= 0 && l.cursor < len(l.visible) && l.rows[l.visible[l.cursor]].header {
			l.cursor -= dir
		}
		l.cursor = max(l.cursor, 0)
	}
}

func (l *listView) move(d int) {
	if len(l.visible) == 0 {
		return
	}
	dir := 1
	if d < 0 {
		dir = -1
	}
	l.cursor = min(max(l.cursor+d, 0), len(l.visible)-1)
	l.skipHeaders(dir)
}

func (l *listView) home() { l.cursor = 0; l.skipHeaders(1) }
func (l *listView) end()  { l.cursor = len(l.visible) - 1; l.skipHeaders(-1) }

func (l *listView) view(width, height int, showArabic bool) string {
	if len(l.visible) == 0 {
		return sDim.Render("  " + l.empty)
	}
	if l.cursor < l.offset {
		l.offset = l.cursor
	}
	if l.cursor >= l.offset+height {
		l.offset = l.cursor - height + 1
	}
	// Keep the header above the first item visible.
	if l.offset == 1 && l.rows[l.visible[0]].header {
		l.offset = 0
	}
	l.offset = max(min(l.offset, len(l.visible)-height), 0)

	var b strings.Builder
	end := min(l.offset+height, len(l.visible))
	for i := l.offset; i < end; i++ {
		r := l.rows[l.visible[i]]
		b.WriteString(l.renderRow(r, i == l.cursor, width, showArabic))
		if i < end-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (l *listView) renderRow(r row, sel bool, width int, showArabic bool) string {
	if r.header {
		return " " + sHeader.Render(r.title)
	}
	// Every segment carries the row background, since an inner style's reset
	// would otherwise cut a background set on the whole line.
	st := func(s lipgloss.Style) lipgloss.Style {
		if sel {
			return s.Background(selBg)
		}
		return s
	}
	plain := lipgloss.NewStyle()
	mark := r.mark
	if mark == "" {
		mark = " "
	}
	gutter := st(plain).Render(" ")
	if sel {
		gutter = st(sAccent).Render("▌")
	}
	left := gutter + st(r.markSt).Render(mark) + st(plain).Render(" ")

	// Right column: meta and Arabic name. On narrow screens drop the meta
	// first, then the Arabic name, so the title keeps at least 16 cells.
	meta, ar := r.meta, ""
	if showArabic {
		ar = r.arabic
	}
	rightWidth := func() int {
		w := ansi.StringWidth(meta) + ansi.StringWidth(ar)
		if meta != "" && ar != "" {
			w += 2
		}
		return w
	}
	if width-4-rightWidth() < 16 {
		meta = ""
	}
	if width-4-rightWidth() < 16 {
		ar = ""
	}
	rw := rightWidth()
	room := width - 3 - rw - 2
	full := r.title
	if r.sub != "" {
		full += "  " + r.sub
	}
	full = ansi.Truncate(full, max(room, 0), "…")
	tw := ansi.StringWidth(full)
	t, sub := full, ""
	if len(full) > len(r.title) && strings.HasPrefix(full, r.title) {
		t, sub = r.title, full[len(r.title):]
	}
	title := st(sText.Bold(sel)).Render(t) + st(sDim).Render(sub)

	gap := st(plain).Render(strings.Repeat(" ", max(width-3-tw-rw-1, 1)))
	right := ""
	if meta != "" {
		right += st(r.metaSt).Render(meta)
	}
	if ar != "" {
		if meta != "" {
			right += st(plain).Render("  ")
		}
		// RLI..PDI isolate the Arabic run so bidi terminals keep the columns.
		right += st(sText).Render("\u2067" + ar + "\u2069")
	}
	return left + title + gap + right + st(plain).Render(" ")
}

// count is the number of selectable rows shown; position is the cursor's
// 1-based index among them.
func (l *listView) count() int {
	n := 0
	for _, i := range l.visible {
		if !l.rows[i].header {
			n++
		}
	}
	return n
}

func (l *listView) position() int {
	n := 0
	for k := 0; k <= l.cursor && k < len(l.visible); k++ {
		if !l.rows[l.visible[k]].header {
			n++
		}
	}
	return n
}
