package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/isobelmcrae/trip/api"
	"github.com/isobelmcrae/trip/rendermaps"
)

func (s *routeState) renderLeg(l api.Leg) {
	width, height := s.root.Main.GetWidth(), s.root.Main.GetHeight()

	centerLat, centerLon, zoom := rendermaps.FocusOn(
        l.Origin.Coord[0], l.Origin.Coord[1],
        l.Destination.Coord[0], l.Destination.Coord[1],
        width, height,
    )

    canvas, err := rendermaps.RenderMap(width - 4, height - 2, centerLat, centerLon, zoom)

    if err != nil {
        log.Error("Error rendering map", "err", err)
        s.root.Main.SetContent("Error rendering map")
        return
    }

	//fmt.Println(lipgloss.JoinVertical(lipgloss.Right, lipgloss.JoinHorizontal(lipgloss.Center, m.Render(dogImg))))

    s.root.Main.SetContent(
		lipgloss.JoinVertical(lipgloss.Right, lipgloss.JoinHorizontal(lipgloss.Center, canvas)),
	)
}
