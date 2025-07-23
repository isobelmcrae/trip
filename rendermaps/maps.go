package rendermaps

import (
	"math"
	"sync"

	"github.com/paulmach/orb"
)

const (
	TileSourceURL = "http://mapscii.me/"
	POIMarker     = "◉"
	ProjectSize   = 256.0
	MaxLat        = 85.0511

	MaxZoom       = 14.0 // any zoom higher than 14 breaks the rendering...
	MinZoom       = 1.0
)

func RenderMap(width, height int, lat, lon float64, zoom float64) (string, error) {
	canvas := NewCanvas(width*pixelWidthPerChar, height*pixelHeightPerChar) // Canvas is in pixels (2x4 per char)
	labelBuffer := NewLabelBuffer()

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

	drawOrder := []string{"landuse", "water", "building", "road", "admin", "place_label", "poi_label"}
	for _, layerName := range drawOrder {
		for _, job := range jobs {
			renderTileLayer(canvas, labelBuffer, job.tile, job.pos, tileSize, zoom, layerName)
		}
	}
	return canvas.Frame(), nil
}
