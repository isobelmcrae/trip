package rendermaps

import (
	"math"
	"sync"

	"github.com/paulmach/orb"
)

const (
	TileSourceURL = "http://mapscii.me/"
	TileRange     = 14
	POIMarker     = "◉"
	ProjectSize   = 256.0
	MaxLat        = 85.0511

	MaxZoom = 17.0
	MinZoom = 1.0
)

type tileJob struct {
	tile *Tile
	pos  orb.Point
}

type Renderer struct {
	Canvas      *Canvas
	labelBuffer *LabelBuffer
	jobs        []tileJob

	tileSize float64
	zoom float64
}

func RenderMap(width, height int, lat, lon float64, zoom float64) *Renderer {
	canvas := NewCanvas(width*pixelWidthPerChar, height*pixelHeightPerChar) // Canvas is in pixels (2x4 per char)
	labelBuffer := NewLabelBuffer()

	z := baseZoom(zoom)
	centerX, centerY := ll2tile(lon, lat, z)
	tileSize := tilesizeAtZoom(zoom)
	gridSize := math.Pow(2, float64(z))

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

	return &Renderer{
		Canvas:      canvas,
		labelBuffer: labelBuffer,
		jobs:        jobs,
		tileSize:    tileSize,
		zoom:        zoom,
	}
}

func RenderMapOneshot(width, height int, lat, lon float64, zoom float64) string {
	renderer := RenderMap(width, height, lat, lon, zoom)
	drawOrder := []string{"landuse", "water", "building", "road", "admin", "place_label", "poi_label"}
	renderer.Draw(drawOrder)
	return renderer.Frame()
}

// drawOrder := []string{"landuse", "water", "building", "road", "admin", "place_label", "poi_label"}

func (r *Renderer) Draw(drawOrder []string) {
	for _, layerName := range drawOrder {
		for _, job := range r.jobs {
			renderTileLayer(r.Canvas, r.labelBuffer, job.tile, job.pos, r.tileSize, r.zoom, layerName)
		}
	}
}

func (r *Renderer) Frame() string {
	return r.Canvas.Frame()
}

/* func RenderMapString(width, height int, lat, lon float64, zoom float64) (string, error) {
	canvas, err := RenderMap(width, height, lat, lon, zoom)
	if err != nil {
		return "", err
	}
	return canvas.Frame(), nil
} */
