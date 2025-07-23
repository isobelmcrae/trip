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
}
