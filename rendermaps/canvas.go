package rendermaps

import (
	"sort"
	"strings"

	"github.com/flywave/go-earcut"
	"github.com/mattn/go-runewidth"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/simplify"
)

func renderTileLayer(canvas *Canvas, lb *LabelBuffer, tile *Tile, pos orb.Point, tileSize, zoom float64, layerName string) {
	if tile.Extent == 0 {
		return // Skip empty/invalid tiles
	}
	scale := float64(tile.Extent) / tileSize
	tile.Rtree.Search(
		[2]float64{-pos.X() * scale, -pos.Y() * scale},
		[2]float64{(float64(canvas.width) - pos.X()) * scale, (float64(canvas.height) - pos.Y()) * scale},
		func(_, _ [2]float64, data interface{}) bool {
			feature := data.(*StyledFeature)
			if feature.Style.SourceLayer == layerName {
				drawFeature(canvas, lb, feature, pos, scale, zoom)
			}
			return true
		},
	)
}

func drawFeature(canvas *Canvas, lb *LabelBuffer, feature *StyledFeature, pos orb.Point, scale, zoom float64) {
	if (feature.Style.MinZoom != 0 && zoom < feature.Style.MinZoom) || (feature.Style.MaxZoom != 0 && zoom > feature.Style.MaxZoom) {
		return
	}

	transform := func(p orb.Point) orb.Point {
		return orb.Point{pos.X() + p.X()/scale, pos.Y() + p.Y()/scale}
	}

	switch feature.Style.Type {
	case "fill":
		if polygon, ok := feature.Geometry.(orb.Polygon); ok {
			var rings []orb.Ring
			for _, ring := range polygon {
				transformedRing := make(orb.Ring, len(ring))
				for i, p := range ring {
					transformedRing[i] = transform(p)
				}
				rings = append(rings, transformedRing)
			}
			canvas.Polygon(rings, feature.Color)
		}
	case "line":
		if ls, ok := feature.Geometry.(orb.LineString); ok {
			points := make(orb.LineString, len(ls))
			for i, p := range ls {
				points[i] = transform(p)
			}
			simplifier := simplify.DouglasPeucker(0.5)
			simplifiedGeom := simplifier.Simplify(points)

			if simplifiedLine, ok := simplifiedGeom.(orb.LineString); ok {
				canvas.Polyline(simplifiedLine, feature.Color)
			}
			// simplified, _ := planar.Simplify(nil, 0.5, points...)
			// canvas.Polyline(simplified, feature.Color)
		}
	case "symbol":
		label := feature.Label
		if label == "" {
			label = POIMarker
		}
		if p, ok := feature.Geometry.(orb.Point); ok {
			tp := transform(p)
			charX, charY := int(tp.X()/2), int(tp.Y()/4)
			if lb.WriteIfPossible(label, charX, charY) {
				canvas.Text(label, int(tp.X()), int(tp.Y()), feature.Color)
			}
		}
	}
}

type Canvas struct {
	width, height int
	pixelBuffer   []byte
	charBuffer    map[int]rune
	colorBuffer   map[int]string
	brailleMap    [4][2]byte
}

func NewCanvas(width, height int) *Canvas {
	size := (width / 2) * (height / 4)
	return &Canvas{
		width: width, height: height, pixelBuffer: make([]byte, size),
		charBuffer: make(map[int]rune), colorBuffer: make(map[int]string),
		brailleMap: [4][2]byte{{0x01, 0x08}, {0x02, 0x10}, {0x04, 0x20}, {0x40, 0x80}},
	}
}

func (c *Canvas) project(x, y int) (int, bool) {
	if x < 0 || x >= c.width || y < 0 || y >= c.height {
		return 0, false
	}
	return (x / 2) + (c.width/2)*(y/4), true
}

func (c *Canvas) SetPixel(x, y int, color string) {
	if idx, ok := c.project(x, y); ok {
		c.pixelBuffer[idx] |= c.brailleMap[y%4][x%2]
		if _, exists := c.colorBuffer[idx]; !exists {
			c.colorBuffer[idx] = color
		}
	}
}

func (c *Canvas) setPixelSplat(x, y int, color string) {
	if idx, ok := c.project(x, y); ok {
		// c.pixelBuffer[idx] |= 0xff // splat!
		// c.pixelBuffer[idx] |= c.brailleMap[y%4][x%2]
		// c.charBuffer[idx] = '•'
		c.charBuffer[idx] = '⬤'
		c.colorBuffer[idx] = color
	}
}

func (c *Canvas) Text(text string, x, y int, color string) {
	// Center text
	x -= (runewidth.StringWidth(text) / 2) * 2
	for i, r := range text {
		if idx, ok := c.project(x+i*2, y); ok {
			c.charBuffer[idx] = r
			c.colorBuffer[idx] = color
		}
	}
}

func (c *Canvas) Frame() string {
	var sb strings.Builder
	termReset := "\x1B[0m"
	currentColor := ""
	for y := 0; y < c.height/4; y++ {
		for x := 0; x < c.width/2; x++ {
			idx := x + y*(c.width/2)
			colorCode := c.colorBuffer[idx]
			if colorCode != currentColor {
				sb.WriteString(termReset)
				if colorCode != "" {
					sb.WriteString(colorCode)
				}
				currentColor = colorCode
			}
			if char, ok := c.charBuffer[idx]; ok {
				sb.WriteRune(char)
			} else if pixelVal := c.pixelBuffer[idx]; pixelVal > 0 {
				sb.WriteRune(rune(0x2800 + int(pixelVal)))
			} else {
				sb.WriteRune(' ')
			}
		}
		sb.WriteString(termReset)
		currentColor = ""
		if y < c.height/4-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

func (c *Canvas) Polyline(points []orb.Point, color string) {
	for i := 0; i < len(points)-1; i++ {
		c.line(points[i], points[i+1], color)
	}
}

var redColour = ("#ff0000")

// to be used after everything is rendered
func (c *Canvas) SplatLineGeo(
	originLat, originLon, destLat, destLon float64,
	mapCenterLat, mapCenterLon,
	mapZoom float64, colour string,
) {
	canvasP1 := geoToPixel(
		originLat, originLon,
		mapCenterLat, mapCenterLon, mapZoom,
		c.width, c.height,
	)

	canvasP2 := geoToPixel(
		destLat, destLon,
		mapCenterLat, mapCenterLon, mapZoom,
		c.width, c.height,
	)

	c.line(canvasP1, canvasP2, hexToANSI(colour), true)
}

func (c *Canvas) line(p1, p2 orb.Point, color string, impl ...bool) {
	var setPixel bool
	if len(impl) > 0 && impl[0] {
		setPixel = true
	} else {
		setPixel = false
	}

	x0, y0, x1, y1 := int(p1.X()), int(p1.Y()), int(p2.X()), int(p2.Y())

	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	sx := -1
	if x0 < x1 {
		sx = 1
	}

	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	dy = -dy // This is correct for the algorithm's error term calculation

	sy := -1
	if y0 < y1 {
		sy = 1
	}

	err := dx + dy

	for {
		if setPixel {
			c.setPixelSplat(x0, y0, color)
		} else {
			c.SetPixel(x0, y0, color)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func (c *Canvas) Polygon(rings []orb.Ring, color string) {
	if len(rings) == 0 {
		return
	}
	var vertices []float64
	var holes []int
	for i, ring := range rings {
		if i > 0 {
			holes = append(holes, len(vertices)/2)
		}
		for _, p := range ring {
			vertices = append(vertices, p.X(), p.Y())
		}
	}
	triangles, err := earcut.Earcut(vertices, holes, 2)
	if err != nil {
		panic("unreachable")
	}
	for i := 0; i < len(triangles); i += 3 {
		p1Idx, p2Idx, p3Idx := triangles[i]*2, triangles[i+1]*2, triangles[i+2]*2
		p1 := orb.Point{vertices[p1Idx], vertices[p1Idx+1]}
		p2 := orb.Point{vertices[p2Idx], vertices[p2Idx+1]}
		p3 := orb.Point{vertices[p3Idx], vertices[p3Idx+1]}
		c.filledTriangle(p1, p2, p3, color)
	}
}

func bresenham(p1, p2 orb.Point) []orb.Point {
	var points []orb.Point
	x0, y0, x1, y1 := int(p1.X()), int(p1.Y()), int(p2.X()), int(p2.Y())

	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	sx := -1
	if x0 < x1 {
		sx = 1
	}

	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	dy = -dy // This is part of the algorithm for all octants
	sy := -1
	if y0 < y1 {
		sy = 1
	}

	err := dx + dy
	for {
		points = append(points, orb.Point{float64(x0), float64(y0)})
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
	return points
}

func (c *Canvas) filledTriangle(p1, p2, p3 orb.Point, color string) {
	// Get all points on the triangle's edges
	edge1 := bresenham(p1, p2)
	edge2 := bresenham(p2, p3)
	edge3 := bresenham(p3, p1)

	allPoints := append(edge1, edge2...)
	allPoints = append(allPoints, edge3...)

	// Sort points primarily by Y, then by X
	sort.Slice(allPoints, func(i, j int) bool {
		if allPoints[i].Y() == allPoints[j].Y() {
			return allPoints[i].X() < allPoints[j].X()
		}
		return allPoints[i].Y() < allPoints[j].Y()
	})

	// Fill between the points on each scanline
	if len(allPoints) == 0 {
		return
	}

	for i := 0; i < len(allPoints)-1; {
		pStart := allPoints[i]
		pEnd := pStart

		// Find the last point on the same scanline
		j := i
		for j < len(allPoints) && allPoints[j].Y() == pStart.Y() {
			pEnd = allPoints[j]
			j++
		}

		// Draw the horizontal line
		y := int(pStart.Y())
		for x := int(pStart.X()); x <= int(pEnd.X()); x++ {
			c.SetPixel(x, y, color)
		}

		i = j // Move to the next scanline
	}
}
