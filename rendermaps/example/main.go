package main

import (
	"fmt"
	"time"

	"github.com/isobelmcrae/trip/rendermaps"
)

func main() {
	t0 := time.Now()
	
	str, err := rendermaps.RenderMap(120, 40, -33.88402424, 151.20620308, 14)	
	if err != nil {
		panic(err)
	}

	println(str)

	t1 := time.Now()
	str, err = rendermaps.RenderMap(120, 40, -33.88402424, 151.20620308, 14)	
	if err != nil {
		panic(err)
	}
	t2 := time.Now()

	fmt.Println(str)
	fmt.Println("Rendered map (cold) in", t1.Sub(t0))
	fmt.Println("Rendered map (hot) in", t2.Sub(t1))

	// render the thing, but nudge the lat long and move around in a circle

	// actually do this in a circle
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
