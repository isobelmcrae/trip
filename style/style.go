package style

import "github.com/charmbracelet/lipgloss"

var (
    Primary = lipgloss.Color("#61AFEF")
    Accent = lipgloss.Color("#E06C75")
    Gray = lipgloss.Color("#5C6370")
    Highlight = lipgloss.Color("#98C379")

    Padding = lipgloss.NewStyle().Padding(0,1)
    Border = lipgloss.RoundedBorder()
)
