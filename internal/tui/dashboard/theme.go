package dashboard

import "charm.land/lipgloss/v2"

var (
	phosphorBackground = lipgloss.Color("#09140D")
	phosphorText       = lipgloss.Color("#A3C9A8")
	phosphorMuted      = lipgloss.Color("#77967D")
	phosphorBorder     = lipgloss.Color("#365C42")
	phosphorBright     = lipgloss.Color("#D0F0C0")
	phosphorSelection  = lipgloss.Color("#203D2A")

	titleStyle = lipgloss.NewStyle().Foreground(phosphorBright).Bold(true)
	mutedStyle = lipgloss.NewStyle().Foreground(phosphorMuted)

	screenStyle = lipgloss.NewStyle().
			Foreground(phosphorText).
			Background(phosphorBackground)

	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(phosphorBorder).
			BorderBackground(phosphorBackground)
)
