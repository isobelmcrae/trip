package rendermaps

import (
	"fmt"
	"math"
	"sync"

	"github.com/paulmach/orb"
)

const (
	TileSourceURL = "http://mapscii.me/"
	POIMarker     = "◉"
	MaxZoom       = 18.0
	MinZoom       = 1.0
	ProjectSize   = 256.0
	MaxLat        = 85.0511
)

func RenderMap(width, height int, lat, lon float64, zoom float64) (string, error) {
	// 1. Setup: Load style, create tile source, and prepare the canvas.
	canvas := NewCanvas(width*2, height*4) // Canvas is in pixels (2x4 per char)
	labelBuffer := NewLabelBuffer()

	// 2. Tile Calculation: Determine which tiles are visible in the viewport.
	z := baseZoom(zoom)
	centerX, centerY := ll2tile(lon, lat, z)
	tileSize := tilesizeAtZoom(zoom)
	gridSize := math.Pow(2, float64(z))

	type tileJob struct {
		tile *Tile
		pos  orb.Point
	}
	fetchedTiles := make(chan tileJob)
	var wg sync.WaitGroup

	// 3. Fetching: Concurrently fetch all visible tiles.
	for ty := math.Floor(centerY) - 1; ty <= math.Floor(centerY)+1; ty++ {
		for tx := math.Floor(centerX) - 1; tx <= math.Floor(centerX)+1; tx++ {
			tileX := int(math.Mod(tx, gridSize))
			if tileX < 0 {
				tileX += int(gridSize)
			}
			tileY := int(ty)

			if tileY < 0 || tileY >= int(gridSize) {
				continue
			}

			wg.Add(1)
			go func(z, x, y int, tx, ty float64) {
				defer wg.Done()
				tile, err := gTs.GetTile(z, x, y)
				if err == nil {
					pos := orb.Point{
						float64(canvas.width)/2 - (centerX-tx)*tileSize,
						float64(canvas.height)/2 - (centerY-ty)*tileSize,
					}
					fetchedTiles <- tileJob{tile: tile, pos: pos}
				}
			}(z, tileX, tileY, tx, ty)
		}
	}

	go func() {
		wg.Wait()
		close(fetchedTiles)
	}()

	var jobs []tileJob
	for job := range fetchedTiles {
		jobs = append(jobs, job)
	}

	// 4. Rendering: Draw features from each tile onto the canvas in a specific order.
	drawOrder := []string{"landuse", "water", "building", "road", "admin", "place_label", "poi_label"}
	fmt.Println("--- Starting Rendering Loop ---") // <-- ADD THIS
	for _, layerName := range drawOrder {
		for _, job := range jobs {
			fmt.Printf("Attempting to render layer '%s' for tile at pos %v\n", layerName, job.pos)
			renderTileLayer(canvas, labelBuffer, job.tile, job.pos, tileSize, zoom, layerName)
		}
	}
	fmt.Println("--- Finished Rendering Loop ---") // <-- ADD THIS
	// 5. Final Output: Convert the canvas's pixel buffer into a printable string.
	return canvas.Frame(), nil
}
