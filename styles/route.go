package styles

import (
	lg "github.com/charmbracelet/lipgloss"
	"github.com/isobelmcrae/trip/api"
)

var (
    RouteTitle = lg.NewStyle().
        Align(lg.Left).
        Bold(true)
    RouteLegBox = lg.NewStyle().
        BorderStyle(lg.RoundedBorder()).
        Align(lg.Left)
)

func CreateLineHighlight(api.Transportation) lg.Style {
    return lg.NewStyle()
}

func FormatRouteLeg(width int) lg.Style {
    return lg.NewStyle().BorderStyle(lg.NormalBorder()).Align(lg.Left).Width(width)
}
