package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// The palette borrows from an illuminated mushaf read at night: gold leaf
// for ornament and what is playing, lapis for things you can press,
// parchment for text, and sand for everything secondary.
var (
	gold      = lipgloss.AdaptiveColor{Light: "#8A6516", Dark: "#D9B36C"}
	goldDim   = lipgloss.AdaptiveColor{Light: "#B89A5E", Dark: "#7D6A45"} // paused: the gilding dims
	lapis     = lipgloss.AdaptiveColor{Light: "#2F4C9E", Dark: "#8EA6E8"}
	parchment = lipgloss.AdaptiveColor{Light: "#2B2519", Dark: "#ECE4D2"}
	sand      = lipgloss.AdaptiveColor{Light: "#766C59", Dark: "#948B78"}
	hairline  = lipgloss.AdaptiveColor{Light: "#D6CDBB", Dark: "#3D3A33"}
	sage      = lipgloss.AdaptiveColor{Light: "#4E7A3A", Dark: "#A3BE8C"}
	carnelian = lipgloss.AdaptiveColor{Light: "#A3402E", Dark: "#E38B78"}
	selBg     = lipgloss.AdaptiveColor{Light: "#E6E9F4", Dark: "#262C3F"}

	sTitle    = lipgloss.NewStyle().Foreground(gold).Bold(true)
	sDim      = lipgloss.NewStyle().Foreground(sand)
	sText     = lipgloss.NewStyle().Foreground(parchment)
	sGold     = lipgloss.NewStyle().Foreground(gold)
	sAccent   = lipgloss.NewStyle().Foreground(lapis)
	sDone     = lipgloss.NewStyle().Foreground(sage)
	sHeader   = lipgloss.NewStyle().Foreground(gold).Bold(true)
	sErr      = lipgloss.NewStyle().Foreground(carnelian)
	sRule     = lipgloss.NewStyle().Foreground(hairline)
	sKey      = lipgloss.NewStyle().Foreground(lapis)
	sBarEmpty = lipgloss.NewStyle().Foreground(hairline)
)

// band is a section title followed by a hairline to the given width, like
// the framed surah headings of a mushaf.
func band(title string, width int) string {
	t := " " + sHeader.Render(title) + " "
	return t + sRule.Render(strings.Repeat("─", max(width-lipgloss.Width(t)-1, 0)))
}
