package picker

import "github.com/charmbracelet/lipgloss"

var (
	paletteLightPurple = lipgloss.Color("183")
	paletteAccentCyan  = lipgloss.Color("51")
	paletteMutedGray   = lipgloss.Color("245")

	titleStyle     = lipgloss.NewStyle().Foreground(paletteAccentCyan).Bold(true)
	inputStyle     = lipgloss.NewStyle().Foreground(paletteLightPurple)
	itemStyle      = lipgloss.NewStyle().Foreground(paletteMutedGray)
	cursorStyle    = lipgloss.NewStyle().Foreground(paletteAccentCyan).Bold(true)
	selectedStyle  = lipgloss.NewStyle().Foreground(paletteAccentCyan).Bold(true)
	helpStyle      = lipgloss.NewStyle().Foreground(paletteMutedGray).Italic(true)
	emptyStyle     = lipgloss.NewStyle().Foreground(paletteMutedGray).Italic(true)
	highlightStyle = lipgloss.NewStyle().Foreground(paletteLightPurple).Bold(true)
)
