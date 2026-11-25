package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ThemeInfo describes a color theme for CLI listing and in-app display.
type ThemeInfo struct {
	ID   string
	Name string
	Desc string
}

// Theme defines the semantic color tokens and pre-computed styles for a visual theme.
// No theme paints large background areas, preserving translucent terminals over wallpapers.
type Theme struct {
	ID      string
	Name    string
	Desc    string
	IsANSI  bool // strictly 16-color ANSI palette (0-15 or terminal default)
	IsLight bool // designed for light terminal backgrounds
	HighCon bool // high contrast accessibility theme

	Accent  lipgloss.Color // primary brand & active focus color
	Accent2 lipgloss.Color // secondary calm highlight (badges, metadata, secondary progress)
	Success lipgloss.Color // finished checkmarks, confirmations, playing indicator
	Text    lipgloss.Color // primary foreground text
	Subtext lipgloss.Color // secondary descriptive text
	Muted   lipgloss.Color // dim hints, inactive elements, subtle metadata
	Border  lipgloss.Color // idle hairline dividers & frames
	Fav     lipgloss.Color // favorite star color
	Live    lipgloss.Color // live radio dot, offline badge, errors
	Warn    lipgloss.Color // sleep timer / warning accents

	Styles ThemeStyles
}

// ThemeStyles holds pre-computed Lip Gloss styles so View() allocates minimally.
type ThemeStyles struct {
	Brand        lipgloss.Style
	TabActive    lipgloss.Style
	TabInactive  lipgloss.Style
	TabNum       lipgloss.Style
	Rule         lipgloss.Style
	RuleAccent   lipgloss.Style
	SectionTitle lipgloss.Style
	Text         lipgloss.Style
	TextBold     lipgloss.Style
	Subtext      lipgloss.Style
	Muted        lipgloss.Style
	Accent       lipgloss.Style
	AccentBold   lipgloss.Style
	Accent2      lipgloss.Style
	Success      lipgloss.Style
	Fav          lipgloss.Style
	Live         lipgloss.Style
	Warn         lipgloss.Style
	Key          lipgloss.Style
	SelectedRail lipgloss.Style
	SelectedText lipgloss.Style
	SelectedSub  lipgloss.Style
	Arabic       lipgloss.Style
	ArabicSel    lipgloss.Style
	BorderIdle   lipgloss.Style
	BorderPlay   lipgloss.Style
	BorderPause  lipgloss.Style
	BorderLive   lipgloss.Style
	BorderModal  lipgloss.Style
	ProgFill     lipgloss.Style
	ProgHead     lipgloss.Style
	ProgEmpty    lipgloss.Style
	Toast        lipgloss.Style
	OfflineBadge lipgloss.Style
}

var themeDefs = []Theme{
	{
		ID:      "mihrab",
		Name:    "Mihrab",
		Desc:    "Warm gilded gold, alabaster & sage on dark terminals (default)",
		Accent:  lipgloss.Color("#DFAA4B"),
		Accent2: lipgloss.Color("#6BB892"),
		Success: lipgloss.Color("#6BB892"),
		Text:    lipgloss.Color("#EDE6D6"),
		Subtext: lipgloss.Color("#AEA596"),
		Muted:   lipgloss.Color("#736B5E"),
		Border:  lipgloss.Color("#524A3E"),
		Fav:     lipgloss.Color("#F0C153"),
		Live:    lipgloss.Color("#E06C5A"),
		Warn:    lipgloss.Color("#E3A857"),
	},
	{
		ID:      "andalus",
		Name:    "Al-Andalus",
		Desc:    "Alhambra terracotta, glazed turquoise & courtyard ivory",
		Accent:  lipgloss.Color("#E08A5A"),
		Accent2: lipgloss.Color("#4ABBB2"),
		Success: lipgloss.Color("#6EC490"),
		Text:    lipgloss.Color("#EFECE4"),
		Subtext: lipgloss.Color("#A6B0B6"),
		Muted:   lipgloss.Color("#6B767E"),
		Border:  lipgloss.Color("#404E56"),
		Fav:     lipgloss.Color("#E8BA5A"),
		Live:    lipgloss.Color("#E56B6F"),
		Warn:    lipgloss.Color("#E8BA5A"),
	},
	{
		ID:      "lapis",
		Name:    "Samarkand Lapis",
		Desc:    "Timurid lapis azure, antique gold leaf & moonlit silk",
		Accent:  lipgloss.Color("#74A8F8"),
		Accent2: lipgloss.Color("#E3BC64"),
		Success: lipgloss.Color("#6FC79A"),
		Text:    lipgloss.Color("#E6ECF6"),
		Subtext: lipgloss.Color("#9BA9C0"),
		Muted:   lipgloss.Color("#66738B"),
		Border:  lipgloss.Color("#3D4B66"),
		Fav:     lipgloss.Color("#E8C468"),
		Live:    lipgloss.Color("#EA7575"),
		Warn:    lipgloss.Color("#E3BC64"),
	},
	{
		ID:      "madinah",
		Name:    "Madinah",
		Desc:    "Luminous emerald, soft brass gold & warm pearl",
		Accent:  lipgloss.Color("#52CA8E"),
		Accent2: lipgloss.Color("#DCBC6C"),
		Success: lipgloss.Color("#66DA9E"),
		Text:    lipgloss.Color("#E8F0EC"),
		Subtext: lipgloss.Color("#9BB0A5"),
		Muted:   lipgloss.Color("#637A6E"),
		Border:  lipgloss.Color("#344C40"),
		Fav:     lipgloss.Color("#E6C66C"),
		Live:    lipgloss.Color("#E27268"),
		Warn:    lipgloss.Color("#DCBC6C"),
	},
	{
		ID:      "parchment",
		Name:    "Parchment",
		Desc:    "Light-background manuscript palette: walnut ink, ochre & forest green",
		IsLight: true,
		Accent:  lipgloss.Color("#9C560C"),
		Accent2: lipgloss.Color("#1A6B49"),
		Success: lipgloss.Color("#1A6B49"),
		Text:    lipgloss.Color("#241E17"),
		Subtext: lipgloss.Color("#574C3F"),
		Muted:   lipgloss.Color("#85796A"),
		Border:  lipgloss.Color("#B8AB96"),
		Fav:     lipgloss.Color("#B36A00"),
		Live:    lipgloss.Color("#B8332A"),
		Warn:    lipgloss.Color("#9C560C"),
	},
	{
		ID:      "ansi",
		Name:    "Terminal",
		Desc:    "Uses only the terminal's 16 ANSI colors (follows your terminal theme)",
		IsANSI:  true,
		Accent:  lipgloss.Color("3"),  // yellow
		Accent2: lipgloss.Color("6"),  // cyan
		Success: lipgloss.Color("2"),  // green
		Text:    lipgloss.Color(""),   // terminal default foreground
		Subtext: lipgloss.Color("7"),  // standard light/secondary
		Muted:   lipgloss.Color("8"),  // bright black (dim)
		Border:  lipgloss.Color("8"),  // bright black
		Fav:     lipgloss.Color("11"), // bright yellow
		Live:    lipgloss.Color("1"),  // red
		Warn:    lipgloss.Color("3"),  // yellow
	},
	{
		ID:      "contrast",
		Name:    "High Contrast",
		Desc:    "Maximum contrast white, vivid gold, electric cyan & crisp borders",
		HighCon: true,
		Accent:  lipgloss.Color("#FFD700"),
		Accent2: lipgloss.Color("#00F0FF"),
		Success: lipgloss.Color("#39FF14"),
		Text:    lipgloss.Color("#FFFFFF"),
		Subtext: lipgloss.Color("#E0E0E0"),
		Muted:   lipgloss.Color("#B0B0B0"),
		Border:  lipgloss.Color("#B8B8B8"),
		Fav:     lipgloss.Color("#FFD700"),
		Live:    lipgloss.Color("#FF5555"),
		Warn:    lipgloss.Color("#FFD700"),
	},
}

var compiledThemes []Theme

func init() {
	compiledThemes = make([]Theme, len(themeDefs))
	for i, t := range themeDefs {
		compiledThemes[i] = compileTheme(t)
	}
}

func compileTheme(t Theme) Theme {
	fg := func(c lipgloss.Color) lipgloss.Style {
		if c == "" {
			return lipgloss.NewStyle()
		}
		return lipgloss.NewStyle().Foreground(c)
	}

	subStyle := fg(t.Subtext)
	if t.IsANSI {
		// On light terminals ANSI "7" can be faint/white; terminal default + Faint adapts to both light & dark.
		subStyle = lipgloss.NewStyle().Faint(true)
	}

	selText := fg(t.Text).Bold(true)
	if t.HighCon {
		selText = fg(t.Accent).Bold(true).Underline(true)
	} else if t.IsANSI {
		selText = fg(t.Accent).Bold(true)
	}

	t.Styles = ThemeStyles{
		Brand:        fg(t.Accent).Bold(true),
		TabActive:    fg(t.Accent).Bold(true),
		TabInactive:  subStyle,
		TabNum:       fg(t.Muted),
		Rule:         fg(t.Border),
		RuleAccent:   fg(t.Accent),
		SectionTitle: fg(t.Accent).Bold(true),
		Text:         fg(t.Text),
		TextBold:     fg(t.Text).Bold(true),
		Subtext:      subStyle,
		Muted:        fg(t.Muted),
		Accent:       fg(t.Accent),
		AccentBold:   fg(t.Accent).Bold(true),
		Accent2:      fg(t.Accent2),
		Success:      fg(t.Success).Bold(true),
		Fav:          fg(t.Fav),
		Live:         fg(t.Live).Bold(true),
		Warn:         fg(t.Warn),
		Key:          fg(t.Accent2).Bold(true),
		SelectedRail: fg(t.Accent).Bold(true),
		SelectedText: selText,
		SelectedSub:  fg(t.Accent2),
		Arabic:       fg(t.Subtext),
		ArabicSel:    fg(t.Accent),
		BorderIdle:   fg(t.Border),
		BorderPlay:   fg(t.Accent),
		BorderPause:  fg(t.Accent2),
		BorderLive:   fg(t.Live),
		BorderModal:  fg(t.Accent),
		ProgFill:     fg(t.Accent),
		ProgHead:     fg(t.Accent).Bold(true),
		ProgEmpty:    fg(t.Border),
		Toast:        fg(t.Success).Bold(true),
		OfflineBadge: fg(t.Live).Bold(true),
	}
	if t.IsANSI {
		t.Styles.Arabic = lipgloss.NewStyle().Faint(true)
	}
	return t
}

// Themes returns metadata for all available color themes.
func Themes() []ThemeInfo {
	out := make([]ThemeInfo, len(compiledThemes))
	for i, t := range compiledThemes {
		out[i] = ThemeInfo{ID: t.ID, Name: t.Name, Desc: t.Desc}
	}
	return out
}

// ValidTheme reports whether name matches a known theme ID (case-insensitive).
func ValidTheme(name string) bool {
	_, ok := themeIndexByName(name)
	return ok
}

func themeIndexByName(name string) (int, bool) {
	key := strings.ToLower(strings.TrimSpace(name))
	for i, t := range compiledThemes {
		if t.ID == key || strings.ToLower(t.Name) == key {
			return i, true
		}
	}
	return 0, false
}

func themeByIndex(i int) Theme {
	if i < 0 || i >= len(compiledThemes) {
		return compiledThemes[0]
	}
	return compiledThemes[i]
}
