package terminal

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Check if the terminal has a dark background
var IsDarkBg = termenv.HasDarkBackground()

var (
	ColorHiGreen   lipgloss.Style
	ColorHiMagenta lipgloss.Style
	ColorHiRed     lipgloss.Style
	ColorHiYellow  lipgloss.Style
	ColorHiCyan    lipgloss.Style
	ColorHiBlue    lipgloss.Style
)

func init() {
	if IsDarkBg {
		ColorHiGreen = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
		ColorHiMagenta = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF"))
		ColorHiRed = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		ColorHiYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
		ColorHiCyan = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
		ColorHiBlue = lipgloss.NewStyle().Foreground(lipgloss.Color("#0000FF"))
	} else {
		ColorHiGreen = lipgloss.NewStyle().Foreground(lipgloss.Color("#008000"))
		ColorHiMagenta = lipgloss.NewStyle().Foreground(lipgloss.Color("#800080"))
		ColorHiRed = lipgloss.NewStyle().Foreground(lipgloss.Color("#800000"))
		ColorHiYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("#808000"))
		ColorHiCyan = lipgloss.NewStyle().Foreground(lipgloss.Color("#008080"))
		ColorHiBlue = lipgloss.NewStyle().Foreground(lipgloss.Color("#000080"))
	}
}
