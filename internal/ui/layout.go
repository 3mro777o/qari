package ui

// Width-exact text helpers. Every frame must be exactly W×H cells: a line one
// cell too wide wraps and scrolls the whole screen.

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	rli = "⁧" // right-to-left isolate: keeps Arabic from reordering the columns around it
	pdi = "⁩"
)

func width(s string) int { return ansi.StringWidth(s) }

func spaces(n int) string { return strings.Repeat(" ", max(n, 0)) }

// trunc clips s to w cells with ".." and keeps bidi isolates balanced.
func trunc(s string, w int) string {
	if width(s) <= w {
		return s
	}
	if w <= 0 {
		return ""
	}
	tail := ".."
	if w <= 2 {
		tail = ""
	}
	out := ansi.Truncate(s, w-len(tail), "")
	if n := strings.Count(out, rli) - strings.Count(out, pdi); n > 0 {
		out += strings.Repeat(pdi, n)
	}
	return out + tail
}

// fit pads or clips s to exactly w cells.
func fit(s string, w int) string {
	s = trunc(s, w)
	return s + spaces(w-width(s))
}

func center(s string, w int) string {
	s = trunc(s, w)
	l := (w - width(s)) / 2
	return spaces(l) + s + spaces(w-width(s)-l)
}

// lr puts left flush left and right flush right in exactly w cells; the left
// side is clipped first.
func lr(left, right string, w int) string {
	rw := width(right)
	if rw == 0 {
		return fit(left, w)
	}
	if rw >= w-1 {
		return fit(right, w)
	}
	left = trunc(left, w-rw-1)
	return left + spaces(w-width(left)-rw) + right
}

// isolate wraps Arabic text in bidi isolates after clipping it to w cells.
func isolate(s string, w int) string {
	if s == "" || w <= 0 {
		return ""
	}
	return rli + trunc(s, w) + pdi
}

// fitFrame makes lines exactly w×h.
func fitFrame(lines []string, w, h int) string {
	out := make([]string, h)
	for i := range out {
		if i < len(lines) {
			out[i] = fit(lines[i], w)
		} else {
			out[i] = spaces(w)
		}
	}
	return strings.Join(out, "\n")
}

// columns joins two blocks side by side with a divider.
func columns(left, right []string, lw, rw, h int, div string) []string {
	out := make([]string, h)
	for i := range out {
		var l, r string
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}
		out[i] = fit(l, lw) + div + fit(r, rw)
	}
	return out
}

// border draws "╭─ left ───── right ─╮" (or the bottom with ╰ ╯) in w cells.
func border(w int, top bool, left, right string, b lipgloss.Style) string {
	cl, cr := "╰", "╯"
	if top {
		cl, cr = "╭", "╮"
	}
	inner := w - 4
	if right != "" {
		right = " " + trunc(right, inner-8) + " "
	}
	if left != "" {
		left = " " + trunc(left, inner-width(right)-2) + " "
	}
	fill := inner - width(left) - width(right)
	if fill < 0 {
		left, right, fill = "", "", inner
	}
	return b.Render(cl+"─") + left + b.Render(strings.Repeat("─", fill)) + right + b.Render("─"+cr)
}

// boxed renders "│ content │" in w cells.
func boxed(w int, content string, b lipgloss.Style) string {
	return b.Render("│") + " " + fit(content, w-4) + " " + b.Render("│")
}

// bar is a playhead bar ━━━●───; mini is a block bar ▓▓░░.
func bar(frac float64, w int, fill, head, empty lipgloss.Style) string {
	if w <= 0 {
		return ""
	}
	pos := int(math.Round(min(max(frac, 0), 1) * float64(w-1)))
	return fill.Render(strings.Repeat("━", pos)) + head.Render("●") + empty.Render(strings.Repeat("─", w-pos-1))
}

func mini(frac float64, w int, fill, empty lipgloss.Style) string {
	n := int(math.Round(min(max(frac, 0), 1) * float64(w)))
	if frac > 0 && n == 0 {
		n = 1
	}
	return fill.Render(strings.Repeat("▓", n)) + empty.Render(strings.Repeat("░", w-n))
}

// clock formats seconds as M:SS or H:MM:SS.
func clock(sec float64) string {
	if sec < 0 || math.IsNaN(sec) || math.IsInf(sec, 0) {
		sec = 0
	}
	t := int(sec + 0.5)
	if t >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", t/3600, t/60%60, t%60)
	}
	return fmt.Sprintf("%d:%02d", t/60, t%60)
}

func count(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1e6)
	case n >= 1000:
		return fmt.Sprintf("%.1fk", float64(n)/1e3)
	case n > 0:
		return fmt.Sprint(n)
	}
	return ""
}

func bytes(n float64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f GB", n/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", n/(1<<20))
	}
	return fmt.Sprintf("%.0f KB", n/(1<<10))
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}

// wrap splits text into at most n lines of w cells.
func wrap(text string, w, n int) []string {
	var out []string
	for _, word := range strings.Fields(text) {
		if k := len(out) - 1; k >= 0 && width(out[k])+1+width(word) <= w {
			out[k] += " " + word
		} else if len(out) < n {
			out = append(out, trunc(word, w))
		} else {
			break
		}
	}
	return out
}
