package main

import (
	"fmt"
	"time"

	"github.com/isobelmcrae/trip/rendermaps"
)

func main() {
	t0 := time.Now()

	str := rendermaps.RenderMapOneshot(120, 40, -33.88402424, 151.20620308, 14)

	println(str)

	t1 := time.Now()
	str = rendermaps.RenderMapOneshot(120, 40, -33.88402424, 151.20620308, 14)
	t2 := time.Now()

	fmt.Println(str)
	fmt.Println("Rendered map (cold) in", t1.Sub(t0))
	fmt.Println("Rendered map (hot) in", t2.Sub(t1))

	centerLat, centerLon, zoom := rendermaps.FocusOn(
		-33.884179, 151.207215, // central
		-33.861351, 151.210295, // circular quay
		120, 40,
	)

	renderer := rendermaps.RenderMap(120, 40, centerLat, centerLon, zoom)
	renderer.Draw([]string{"landuse", "water", "building", "road", "admin"})

	renderer.Canvas.SplatLineGeo(
		-33.884179, 151.207215, // central
		-33.861351, 151.210295, // circular quay
		centerLat, centerLon,
		zoom, "#ff0000", // red line
	)

	renderer.Draw([]string{"place_label", "poi_label"})

	str = renderer.Frame()
	fmt.Println(str)
	
	/* if err != nil {
		panic(err)
	}

	canvas.RedLineGeo(
		-33.884179, 151.207215, // central
		-33.861351, 151.210295, // circular quay
		centerLat, centerLon,
		zoom,
	)

	str = canvas.Frame()
	fmt.Println(str) */

	/* for i := 0; i < 360; i++ {
		lat := -33.88402424 + 0.01 * float64(i) * 0.017453292519943295 // 0.01 degrees in radians
		lon := 151.20620308 + 0.01 * float64(i) * 0.017453292519943295 // 0.01 degrees in radians
		str, err := rendermaps.RenderMap(120, 40, lat, lon, 14)
		if err != nil {
			panic(err)
		}
		os.Stdout.WriteString(str + "\n")
		os.Stdout.Sync()

		time.Sleep(100 * time.Millisecond)
	} */
}
