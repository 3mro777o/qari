package ui

import "github.com/charmbracelet/lipgloss"

var (
	accent = lipgloss.AdaptiveColor{Light: "#0F766E", Dark: "#5EEAD4"}
	gold   = lipgloss.AdaptiveColor{Light: "#B45309", Dark: "#FBBF24"}
	dim    = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#7C8591"}
	faint  = lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#374151"}
	text   = lipgloss.AdaptiveColor{Light: "#111827", Dark: "#E5E7EB"}
	selBg  = lipgloss.AdaptiveColor{Light: "#CCFBF1", Dark: "#134E4A"}
	danger = lipgloss.AdaptiveColor{Light: "#B91C1C", Dark: "#F87171"}

	sTitle    = lipgloss.NewStyle().Foreground(accent).Bold(true)
	sTab      = lipgloss.NewStyle().Foreground(dim).Padding(0, 1)
	sTabOn    = lipgloss.NewStyle().Foreground(accent).Bold(true).Padding(0, 1).Underline(true)
	sDim      = lipgloss.NewStyle().Foreground(dim)
	sText     = lipgloss.NewStyle().Foreground(text)
	sGold     = lipgloss.NewStyle().Foreground(gold)
	sAccent   = lipgloss.NewStyle().Foreground(accent)
	sHeader   = lipgloss.NewStyle().Foreground(gold).Bold(true)
	sSel      = lipgloss.NewStyle().Background(selBg)
	sErr      = lipgloss.NewStyle().Foreground(danger)
	sRule     = lipgloss.NewStyle().Foreground(faint)
	sKey      = lipgloss.NewStyle().Foreground(accent)
	sBarFull  = lipgloss.NewStyle().Foreground(accent)
	sBarEmpty = lipgloss.NewStyle().Foreground(faint)
)
