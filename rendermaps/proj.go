package rendermaps

import (
	"math"

	"github.com/paulmach/orb"
)

const (
	pixelWidthPerChar = 2
	pixelHeightPerChar = 4
)

// yMercatorNormalized calculates the normalized y-coordinate in Mercator projection (from 0 to 1).
// This is derived from the formula in mapscii/src/utils.js ll2tile().
func yMercatorNormalized(lat float64) float64 {
	latRad := lat * math.Pi / 180
	return (1 - math.Log(math.Tan(latRad)+1/math.Cos(latRad))/math.Pi) / 2
}

// FocusOn calculates the optimal center latitude, longitude, and zoom level to fit
// two coordinates within a given view size. The view size is provided in terminal
// characters (width, height).
func FocusOn(lat1, lon1, lat2, lon2 float64, viewWidthChars, viewHeightChars int) (centerLat, centerLon, zoom float64) {
	// Apply some padding so the points are not at the very edge of the map.
	// A value of 0.8 means the bounding box will take up 80% of the view.
	padding := 0.8
	viewWidthPixels := float64(viewWidthChars*pixelWidthPerChar) * padding
	viewHeightPixels := float64(viewHeightChars*pixelHeightPerChar) * padding

	// If the two points are (almost) the same, we can't calculate a span.
	// Default to a fixed high zoom level centered on the point.
	if math.Abs(lat1-lat2) < 1e-6 && math.Abs(lon1-lon2) < 1e-6 {
		return lat1, lon1, MaxZoom // A good default zoom for a single point
	}

	// 1. Calculate the bounding box, spans, and center point.
	minLat := math.Min(lat1, lat2)
	maxLat := math.Max(lat1, lat2)
	centerLat = (minLat + maxLat) / 2

	// The latitude span in normalized Mercator coordinates.
	latSpanNorm := math.Abs(yMercatorNormalized(maxLat) - yMercatorNormalized(minLat))

	// Handle longitude carefully due to the antimeridian (180° longitude).
	var lonSpan float64
	if math.Abs(lon1-lon2) > 180 {
		// The shortest path crosses the antimeridian.
		maxLonVal := math.Max(lon1, lon2)
		minLonVal := math.Min(lon1, lon2)
		lonSpan = 360 - (maxLonVal - minLonVal)
		centerLon = (maxLonVal + minLonVal + 360) / 2
		if centerLon > 180 {
			centerLon -= 360
		}
	} else {
		// The path does not cross the antimeridian.
		lonSpan = math.Abs(lon1 - lon2)
		centerLon = (lon1 + lon2) / 2
	}

	// 2. Calculate the required zoom level.
	// This is done by finding the necessary "world size" in pixels for both
	// width and height, taking the smaller of the two, and then converting that
	// back to a zoom level.
	var worldSizeForLon, worldSizeForLat float64

	if lonSpan > 0 {
		worldSizeForLon = viewWidthPixels * 360 / lonSpan
	}
	if latSpanNorm > 0 {
		worldSizeForLat = viewHeightPixels / latSpanNorm
	}

	// Determine the constraining world size.
	var worldSize float64
	if worldSizeForLon > 0 && worldSizeForLat > 0 {
		worldSize = math.Min(worldSizeForLon, worldSizeForLat)
	} else if worldSizeForLon > 0 {
		worldSize = worldSizeForLon
	} else {
		worldSize = worldSizeForLat
	}

	// If worldSize is still zero, we can't calculate a zoom.
	if worldSize <= 0 {
		return centerLat, centerLon, MaxZoom
	}

	// Convert world size to zoom level using the formula: worldSize = tileSize * 2^zoom
	zoom = math.Log2(worldSize / ProjectSize)

	// Clamp zoom to the valid range.
	if zoom > MaxZoom {
		zoom = MaxZoom
	}
	if zoom < MinZoom {
		zoom = MinZoom
	}

	return centerLat, centerLon, zoom
}

// tileCoord represents a Web Mercator tile coordinate at a specific zoom level.
type tileCoord struct {
	X, Y, Z float64
}

/* func clampLat(lat float64) float64 {
	// The Mercator projection is undefined at the poles.
	// This range is the standard for web maps (e.g., Google, OpenStreetMap).
	const maxMercatorLat = 85.05112878
	if lat > maxMercatorLat {
		return maxMercatorLat
	}
	if lat < -maxMercatorLat {
		return -maxMercatorLat
	}
	return lat
} */

// llToTile converts a latitude/longitude pair to a tile coordinate at a given integer zoom level.
// This is a port of the formula from mapscii/src/utils.js.
/* func llToTile(lon, lat float64, zoom int) tileCoord {
	clampedLat := clampLat(lat)

	zoomFloat := float64(zoom)
	x := (lon + 180) / 360 * math.Pow(2, zoomFloat)

	latRad := clampedLat * math.Pi / 180 // Use the clamped latitude
	y := (1 - math.Log(math.Tan(latRad)+1/math.Cos(latRad))/math.Pi) / 2 * math.Pow(2, zoomFloat)

	return tileCoord{X: x, Y: y, Z: zoomFloat}
}

func tileSizeAtZoom(zoom float64, baseZoom int) float64 {
	return 256.0 * math.Pow(2, zoom-float64(baseZoom))
} */

// GeoToPixel converts a single lat/lon coordinate to an absolute (x, y) pixel coordinate on the canvas.
// It takes into account the map's center, zoom, and the canvas dimensions.
func geoToPixel(lat, lon float64, centerLat, centerLon, zoom float64, canvasWidth, canvasHeight int) orb.Point {
	// --- Step 1: Define projection constants and helpers ---
	const TILE_SIZE = 256.0
	clampLat := func(lat float64) float64 {
		const maxMercatorLat = 85.05112878
		if lat > maxMercatorLat {
			return maxMercatorLat
		}
		if lat < -maxMercatorLat {
			return -maxMercatorLat
		}
		return lat
	}

	// --- Step 2: Calculate intermediate projection values ---
	// Get the integer zoom level for base tile calculations.
	baseZoomVal := int(math.Floor(zoom))
	
	// Calculate the effective size of a tile in pixels for the given fractional zoom.
	effectiveTileSize := TILE_SIZE * math.Pow(2, zoom-float64(baseZoomVal))

	// --- Step 3: Project Center and Target Point to Mercator Tile Coordinates ---
	// Project the map's center point.
	//centerLonRad := centerLon * math.Pi / 180
	centerLatRad := clampLat(centerLat) * math.Pi / 180
	centerTileX := (centerLon + 180) / 360 * math.Pow(2, float64(baseZoomVal))
	centerTileY := (1 - math.Log(math.Tan(centerLatRad)+1/math.Cos(centerLatRad))/math.Pi) / 2 * math.Pow(2, float64(baseZoomVal))

	// Project the target point (the one we want to draw).
	//lonRad := lon * math.Pi / 180
	latRad := clampLat(lat) * math.Pi / 180
	pointTileX := (lon + 180) / 360 * math.Pow(2, float64(baseZoomVal))
	pointTileY := (1 - math.Log(math.Tan(latRad)+1/math.Cos(latRad))/math.Pi) / 2 * math.Pow(2, float64(baseZoomVal))

	// --- Step 4: Calculate the pixel offset from the center of the screen ---
	// Find the distance between the center and the target point in tile units.
	dxInTiles := pointTileX - centerTileX
	dyInTiles := pointTileY - centerTileY

	// Convert this tile distance into a pixel distance.
	dxInPixels := dxInTiles * effectiveTileSize
	dyInPixels := dyInTiles * effectiveTileSize

	// The final coordinate is the center of the canvas plus the pixel distance.
	finalX := float64(canvasWidth)/2 + dxInPixels
	finalY := float64(canvasHeight)/2 + dyInPixels

	return orb.Point{finalX, finalY}
}