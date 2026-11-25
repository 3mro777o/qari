package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	rli = "\u2067" // Right-to-Left Isolate
	pdi = "\u2069" // Pop Directional Isolate
)

// dispWidth measures the visible terminal cell width of s (ANSI-aware).
func dispWidth(s string) int {
	return ansi.StringWidth(s)
}

// trunc clips s to at most maxW terminal cells, appending ".." when shortened.
// Uses pure-ASCII ellipsis to avoid East Asian Ambiguous width mismatches.
func trunc(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	if dispWidth(s) <= maxW {
		return s
	}
	if maxW <= 2 {
		return ansi.Truncate(s, maxW, "")
	}
	return ansi.Truncate(s, maxW, "..")
}

// padRight returns s fitted to exact width w (truncating if wider, padding with unstyled spaces if narrower).
func padRight(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if dispWidth(s) > w {
		s = trunc(s, w)
	}
	if pad := w - dispWidth(s); pad > 0 {
		s += strings.Repeat(" ", pad)
	}
	return s
}

// padLeft aligns s to the right within exact width w.
func padLeft(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if dispWidth(s) > w {
		s = trunc(s, w)
	}
	if pad := w - dispWidth(s); pad > 0 {
		return strings.Repeat(" ", pad) + s
	}
	return s
}

// center centers s horizontally within exact width w using unstyled spaces.
func center(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if dispWidth(s) > w {
		s = trunc(s, w)
	}
	sw := dispWidth(s)
	left := (w - sw) / 2
	right := w - sw - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// rowLR lays out left flush-left and right flush-right in exact width w.
// Guarantees at least 1 space gap when both sides are present; truncates left first.
func rowLR(left, right string, w int) string {
	if w <= 0 {
		return ""
	}
	rw := dispWidth(right)
	if rw >= w {
		return padRight(right, w)
	}
	if rw == 0 {
		return padRight(left, w)
	}
	maxLeft := w - rw - 1
	if maxLeft <= 0 {
		return padLeft(right, w)
	}
	if dispWidth(left) > maxLeft {
		left = trunc(left, maxLeft)
	}
	gap := w - dispWidth(left) - rw
	if gap < 0 {
		gap = 0
	}
	return left + strings.Repeat(" ", gap) + right
}

// wrapArabic isolates Arabic text with U+2067..U+2069 so bidi terminals preserve column alignment.
func wrapArabic(s string) string {
	if s == "" {
		return ""
	}
	return rli + s + pdi
}

// fitArabic truncates plain Arabic text to maxW cells BEFORE wrapping in bidi isolates,
// ensuring isolates are always balanced and never cut mid-sequence.
func fitArabic(s string, maxW int) string {
	if s == "" || maxW <= 0 {
		return ""
	}
	if dispWidth(s) > maxW {
		s = trunc(s, maxW)
	}
	return wrapArabic(s)
}

// fitFrame guarantees the rendered screen is EXACTLY w cells wide × h lines tall,
// using unstyled spaces for padding so translucent terminals stay transparent.
func fitFrame(content string, w, h int) string {
	if w <= 0 || h <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if len(lines) > h {
		lines = lines[:h]
	}
	blank := strings.Repeat(" ", w)
	for len(lines) < h {
		lines = append(lines, blank)
	}
	var b strings.Builder
	b.Grow((w + 16) * h)
	for i, line := range lines {
		if i > 0 {
			b.WriteByte('\n')
		}
		lw := dispWidth(line)
		if lw > w {
			line = ansi.Truncate(line, w, "")
			lw = dispWidth(line)
		}
		b.WriteString(line)
		if lw < w {
			b.WriteString(strings.Repeat(" ", w-lw))
		}
	}
	return b.String()
}

// joinColumns joins two vertical slices of exact-width lines side by side with a 1-col divider.
func joinColumns(leftLines, rightLines []string, leftW, rightW, height int, div string) []string {
	out := make([]string, height)
	for i := 0; i < height; i++ {
		var l, r string
		if i < len(leftLines) {
			l = padRight(leftLines[i], leftW)
		} else {
			l = strings.Repeat(" ", leftW)
		}
		if i < len(rightLines) {
			r = padRight(rightLines[i], rightW)
		} else {
			r = strings.Repeat(" ", rightW)
		}
		out[i] = l + div + r
	}
	return out
}

// frameTopLine renders "╭─ <left> ──── <right> ─╮" in exact width w.
func frameTopLine(w int, left, right string, border lipgloss.Style) string {
	return frameBorderLine(w, "╭", "╮", left, right, border)
}

// frameBottomLine renders "╰─ <left> ──── <right> ─╯" in exact width w.
func frameBottomLine(w int, left, right string, border lipgloss.Style) string {
	return frameBorderLine(w, "╰", "╯", left, right, border)
}

func frameBorderLine(w int, cornerL, cornerR, left, right string, border lipgloss.Style) string {
	if w < 4 {
		return padRight(border.Render(strings.Repeat("─", w)), w)
	}
	inner := w - 2
	// Reserve minimum 1 dash on each side if labels exist.
	rw := dispWidth(right)
	lw := dispWidth(left)

	var leftSeg, rightSeg string
	if rw > 0 {
		// Fit right segment if excessively wide.
		if rw > inner-3 {
			right = trunc(right, max(0, inner-3))
			rw = dispWidth(right)
		}
		if rw > 0 {
			rightSeg = " " + right + " "
		}
	}
	rSegW := dispWidth(rightSeg)

	if lw > 0 {
		// Leave room for "─" + leftSeg + "─" + rightSeg + "─".
		maxLeft := inner - rSegW - 5
		if maxLeft < 0 {
			maxLeft = 0
		}
		if lw > maxLeft {
			left = trunc(left, maxLeft)
			lw = dispWidth(left)
		}
		if lw > 0 {
			leftSeg = " " + left + " "
		}
	}
	lSegW := dispWidth(leftSeg)

	// Build: cornerL + "─" + leftSeg + fill("─") + rightSeg + "─" + cornerR
	if lSegW == 0 && rSegW == 0 {
		return border.Render(cornerL + strings.Repeat("─", inner) + cornerR)
	}
	lead := 1
	trail := 1
	fill := inner - lead - trail - lSegW - rSegW
	if fill < 0 {
		// Fallback: drop rightSeg if ultra narrow
		rightSeg = ""
		rSegW = 0
		fill = inner - lead - trail - lSegW
		if fill < 0 {
			leftSeg = ""
			lSegW = 0
			fill = inner - lead - trail
		}
	}
	var b strings.Builder
	b.WriteString(border.Render(cornerL + strings.Repeat("─", lead)))
	b.WriteString(leftSeg)
	if fill > 0 {
		b.WriteString(border.Render(strings.Repeat("─", fill)))
	}
	b.WriteString(rightSeg)
	b.WriteString(border.Render(strings.Repeat("─", trail) + cornerR))
	return padRight(b.String(), w)
}

// frameContentLine renders "│ " + content (padded to w-4) + " │" in exact width w.
func frameContentLine(w int, content string, border lipgloss.Style) string {
	if w < 4 {
		return padRight(content, w)
	}
	inner := padRight(content, w-4)
	return border.Render("│") + " " + inner + " " + border.Render("│")
}

// renderProgressBar draws a sleek exact-width playhead bar using safe glyphs (━, ●, ─).
func renderProgressBar(frac float64, w int, fillStyle, headStyle, emptyStyle lipgloss.Style) string {
	if w <= 0 {
		return ""
	}
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	if w == 1 {
		if frac > 0 {
			return headStyle.Render("●")
		}
		return emptyStyle.Render("─")
	}
	pos := int(math.Round(frac * float64(w-1)))
	if pos < 0 {
		pos = 0
	}
	if pos >= w {
		pos = w - 1
	}
	var b strings.Builder
	if pos > 0 {
		b.WriteString(fillStyle.Render(strings.Repeat("━", pos)))
	}
	b.WriteString(headStyle.Render("●"))
	if rem := w - pos - 1; rem > 0 {
		b.WriteString(emptyStyle.Render(strings.Repeat("─", rem)))
	}
	return b.String()
}

// renderMiniBar draws a compact exact-width progress bar using safe block glyphs (▓, ░).
func renderMiniBar(frac float64, w int, fillStyle, emptyStyle lipgloss.Style) string {
	if w <= 0 {
		return ""
	}
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	filled := int(math.Round(frac * float64(w)))
	if frac > 0 && filled == 0 {
		filled = 1
	}
	if filled > w {
		filled = w
	}
	var b strings.Builder
	if filled > 0 {
		b.WriteString(fillStyle.Render(strings.Repeat("▓", filled)))
	}
	if rem := w - filled; rem > 0 {
		b.WriteString(emptyStyle.Render(strings.Repeat("░", rem)))
	}
	return b.String()
}

// formatDuration formats seconds as M:SS or H:MM:SS.
func formatDuration(sec float64) string {
	if sec < 0 || math.IsNaN(sec) || math.IsInf(sec, 0) {
		return "0:00"
	}
	total := int(sec + 0.5)
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// formatListens formats a popularity listen count compactly (e.g. 54.8k, 1.2M).
func formatListens(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 10_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	case n > 0:
		return fmt.Sprintf("%d", n)
	default:
		return ""
	}
}

// overlayCenter composites a rendered modal box centered over base lines cleanly.
// Rows above and below the modal preserve the live view; overlaid rows use clean
// translucent margins so Arabic bidi isolates are never split mid-sequence.
func overlayCenter(base string, modalLines []string, w, h int) string {
	baseLines := strings.Split(fitFrame(base, w, h), "\n")
	mh := len(modalLines)
	if mh == 0 {
		return strings.Join(baseLines, "\n")
	}
	mw := 0
	for _, ml := range modalLines {
		if dw := dispWidth(ml); dw > mw {
			mw = dw
		}
	}
	if mw > w {
		mw = w
	}
	if mh > h {
		modalLines = modalLines[:h]
		mh = h
	}
	top := (h - mh) / 2
	left := (w - mw) / 2
	right := w - left - mw
	padL := strings.Repeat(" ", left)
	padR := strings.Repeat(" ", right)

	for i := 0; i < mh; i++ {
		row := top + i
		if row < 0 || row >= len(baseLines) {
			continue
		}
		baseLines[row] = padL + padRight(modalLines[i], mw) + padR
	}
	return strings.Join(baseLines, "\n")
}
