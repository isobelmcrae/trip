package main

import "github.com/isobelmcrae/trip/rendermaps"

func main() {
	str, err := rendermaps.RenderMap(120, 40, -33.88402424, 151.20620308, 14)	
	if err != nil {
		panic(err)
	}

	println(str)
}