package style

import "github.com/charmbracelet/lipgloss"

var ListTitle = lipgloss.NewStyle().
    Align(lipgloss.Center).
    Border(lipgloss.RoundedBorder()).
    Background(nil).
    Foreground(lipgloss.Color("5")).
    Padding(2, 0, 0, 2)

