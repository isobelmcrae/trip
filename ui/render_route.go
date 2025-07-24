package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/isobelmcrae/trip/api"
	"github.com/isobelmcrae/trip/rendermaps"
	"github.com/isobelmcrae/trip/styles"
)

func (s *routeState) renderLeg(legs []api.Leg, legIdx int) {
	l := legs[legIdx]
	
	width, height := s.root.Main.GetWidth(), s.root.Main.GetHeight()

	// focus on the leg's origin and destination
	centerLat, centerLon, zoom := rendermaps.FocusOn(
        l.Origin.Coord[0], l.Origin.Coord[1],
        l.Destination.Coord[0], l.Destination.Coord[1],
        width, height,
    )

	// we don't wanna zoom in further than this
	if zoom > 14 {
		zoom = 14
	}

	renderer := rendermaps.RenderMap(width - 4, height - 2, centerLat, centerLon, zoom)
	renderer.Draw([]string{"landuse", "water", "building", "road", "admin"})

	// draw the leg as a red line
	for leg := range legs {
		// TODO(iso): make this actually do the colours of the transport type
		//            this is just testing data for now
		
		var hex string

		hex = styles.HexColourForLine(legs[leg].Transportation.DisassembledName)
		
		l := legs[leg]
		renderPartLeg(s, renderer, l, centerLat, centerLon, zoom, hex)
	}

	// but still draw the rest of the lines too
	renderer.Draw([]string{"place_label", "poi_label"})

	frame := renderer.Frame()

    s.root.Main.SetContent(
		lipgloss.JoinVertical(lipgloss.Right, lipgloss.JoinHorizontal(lipgloss.Center, frame)),
	)
}

func renderPartLeg(
	s *routeState,
	renderer *rendermaps.Renderer, l api.Leg,
	centerLat float64, centerLon float64, zoom float64, hex string,
) {
	// no more naive method
	/* renderer.Canvas.SplatLineGeo(
		l.Origin.Coord[0], l.Origin.Coord[1],
		l.Destination.Coord[0], l.Destination.Coord[1],
		centerLat, centerLon,
		zoom, hex,
	) */

	for i := 1; i < len(l.StopSequence); i++ {
		stop := l.StopSequence[i].ID
		prevStop := l.StopSequence[i-1].ID

		points, err := s.root.Client.GetJourneyLeg(prevStop, stop)
		if err != nil {
			continue // skip this leg if we can't find a path
		}

		pointsLen := len(points)
		for j := 0; j < pointsLen-1; j++ {
			renderer.Canvas.SplatLineGeo(
				points[j][0], points[j][1],
				points[j+1][0], points[j+1][1],
				centerLat, centerLon,
				zoom, hex,
			)
		}
	}
}
