package rendermaps

import (
	"fmt"
	"image/color"

	sm "github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
	"github.com/isobelmcrae/trip/api"
)

func TripRender(journey api.Journey) error {
	ctx := sm.NewContext()

	if len(journey.Legs) == 0 {
		return fmt.Errorf("cannot render a trip with no legs")
	}
	
	panic("set api key in this function")
	ctx.SetTileProvider(&sm.TileProvider{
		TileSize: 512,
		URLPattern: "https://api.maptiler.com/maps/toner-v2/%[2]d/%[3]d/%[4]d.png?key=",
	})
	ctx.OverrideAttribution("")
	
	ctx.SetSize(800, 600)

	var objects []sm.MapObject

	pathColor := color.RGBA{R: 0x00, G: 0x00, B: 0xFF, A: 0xBB}
	startMarkerColor := color.RGBA{R: 0x00, G: 0xFF, B: 0x00, A: 0xFF}
	endMarkerColor := color.RGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF}

	latLngFromCoord := func(coord []float64) (s2.LatLng, bool) {
		if len(coord) == 2 {
			return s2.LatLngFromDegrees(coord[0], coord[1]), true
		}
		return s2.LatLng{}, false
	}

	for _, leg := range journey.Legs {
		var legPathPoints []s2.LatLng

		if p, ok := latLngFromCoord(leg.Origin.Coord); ok {
			legPathPoints = append(legPathPoints, p)
		}
		for _, stop := range leg.StopSequence {
			if p, ok := latLngFromCoord(stop.Coord); ok {
				legPathPoints = append(legPathPoints, p)
			}
		}
		if p, ok := latLngFromCoord(leg.Destination.Coord); ok {
			legPathPoints = append(legPathPoints, p)
		}

		if len(legPathPoints) > 1 {
			objects = append(objects, sm.NewPath(legPathPoints, pathColor, 3.0))
		}
	}

	if len(journey.Legs) > 0 {
		firstLeg := journey.Legs[0]
		if p, ok := latLngFromCoord(firstLeg.Origin.Coord); ok {
			objects = append(objects, sm.NewMarker(p, startMarkerColor, 16.0))
		}

		lastLeg := journey.Legs[len(journey.Legs)-1]
		if p, ok := latLngFromCoord(lastLeg.Destination.Coord); ok {
			objects = append(objects, sm.NewMarker(p, endMarkerColor, 16.0))
		}
	}

	if len(objects) == 0 {
		return fmt.Errorf("cannot render trip: the provided journey contains no valid coordinates to draw")
	}

	for _, object := range objects {
		ctx.AddObject(object)
	}

	img, err := ctx.Render()
	if err != nil {
		return fmt.Errorf("failed to render map: %w", err)
	}

	if err := gg.SavePNG("trip_map.png", img); err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	fmt.Println("Successfully rendered and saved trip_map.png")
	return nil
}