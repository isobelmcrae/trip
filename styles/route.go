package styles

import (
	lg "github.com/charmbracelet/lipgloss"
)

var (
    RouteTitle = lg.NewStyle().
        Align(lg.Left).
        Bold(true)
    RouteLegBox = lg.NewStyle().
        BorderStyle(lg.RoundedBorder()).
        Align(lg.Left)
)

func CreateLineHighlight(transit string) lg.Style {
    colour := LgColourForLine(transit)

    style := lg.NewStyle().Background(colour).Bold(true)
    return style
}

func FormatRouteLeg(width int, transit string) lg.Style {
    return lg.NewStyle().
        BorderStyle(lg.RoundedBorder()).
        Align(lg.Left).
        Width(width).
        PaddingRight(1).
        PaddingLeft(1).
        BorderForeground(LgColourForLine(transit)) // mid fix

}
