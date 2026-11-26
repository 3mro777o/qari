package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme is a named palette. No theme paints backgrounds, so translucent
// terminals stay see-through; hierarchy comes from foreground colors only.
type Theme struct {
	ID, Name, Desc string
	// accent, accent2, success, text, subtext, muted, border, favorite, live, warn
	c [10]string
	s styles
}

type styles struct {
	Text, Bold, Sub, Muted, Accent, AccentB, Accent2, Success, Fav, Live, Warn, Key, Rule, Sel lipgloss.Style
}

var themes = []Theme{
	{ID: "mihrab", Name: "Mihrab", Desc: "Warm gilded gold, alabaster and sage (default)",
		c: [10]string{"#DFAA4B", "#6BB892", "#6BB892", "#EDE6D6", "#AEA596", "#736B5E", "#524A3E", "#F0C153", "#E06C5A", "#E3A857"}},
	{ID: "andalus", Name: "Al-Andalus", Desc: "Alhambra terracotta, glazed turquoise and ivory",
		c: [10]string{"#E08A5A", "#4ABBB2", "#6EC490", "#EFECE4", "#A6B0B6", "#6B767E", "#404E56", "#E8BA5A", "#E56B6F", "#E8BA5A"}},
	{ID: "lapis", Name: "Samarkand Lapis", Desc: "Timurid lapis azure, gold leaf and moonlit silk",
		c: [10]string{"#74A8F8", "#E3BC64", "#6FC79A", "#E6ECF6", "#9BA9C0", "#66738B", "#3D4B66", "#E8C468", "#EA7575", "#E3BC64"}},
	{ID: "madinah", Name: "Madinah", Desc: "Luminous emerald, soft brass and warm pearl",
		c: [10]string{"#52CA8E", "#DCBC6C", "#66DA9E", "#E8F0EC", "#9BB0A5", "#637A6E", "#344C40", "#E6C66C", "#E27268", "#DCBC6C"}},
	{ID: "contrast", Name: "High Contrast", Desc: "White, vivid gold and electric cyan, crisp borders",
		c: [10]string{"#FFD700", "#00F0FF", "#39FF14", "#FFFFFF", "#E0E0E0", "#B0B0B0", "#B8B8B8", "#FFD700", "#FF5555", "#FFD700"}},
}

func init() {
	for i := range themes {
		t := &themes[i]
		fg := func(n int) lipgloss.Style { return lipgloss.NewStyle().Foreground(lipgloss.Color(t.c[n])) }
		t.s = styles{
			Accent: fg(0), AccentB: fg(0).Bold(true), Accent2: fg(1), Success: fg(2).Bold(true),
			Text: fg(3), Bold: fg(3).Bold(true), Sub: fg(4), Muted: fg(5), Rule: fg(6),
			Fav: fg(7), Live: fg(8).Bold(true), Warn: fg(9), Key: fg(1).Bold(true), Sel: fg(3).Bold(true),
		}
		if t.ID == "contrast" {
			t.s.Sel = fg(0).Bold(true).Underline(true)
		}
	}
}

// Themes lists the built-in themes.
func Themes() []Theme { return themes }

// ValidTheme reports whether name is a theme id or name.
func ValidTheme(name string) bool { return themeIndex(name) >= 0 }

func themeIndex(name string) int {
	name = strings.ToLower(strings.TrimSpace(name))
	for i, t := range themes {
		if t.ID == name || strings.ToLower(t.Name) == name {
			return i
		}
	}
	return -1
}
