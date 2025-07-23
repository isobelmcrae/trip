package rendermaps

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/mvt"
	"github.com/paulmach/orb/geojson"
	"github.com/tidwall/rtree"
)

type StyledFeature struct {
	Geometry     orb.Geometry
	Style        *StyleLayer
	Color, Label string
}

type Tile struct {
	mvt.Layer
	Rtree *rtree.RTree
}

type TileSource struct {
	url        string
	styler     *Styler
	client     *http.Client
	cache      map[string]*Tile
	colorCache map[string]string
	mu         sync.Mutex
}

func newTileSource(url string, styler *Styler) *TileSource {
	return &TileSource{
		url: url, client: &http.Client{Timeout: 10 * time.Second}, styler: styler,
		cache: make(map[string]*Tile), colorCache: make(map[string]string),
	}
}

func (ts *TileSource) GetTile(z, x, y int) (*Tile, error) {
	key := fmt.Sprintf("%d-%d-%d", z, x, y)
	ts.mu.Lock()
	if tile, ok := ts.cache[key]; ok {
		ts.mu.Unlock()
		return tile, nil
	}
	ts.mu.Unlock()
	url := fmt.Sprintf("%s%d/%d/%d.pbf", ts.url, z, x, y)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "MapSCII-Go-MVP/1.0")
	resp, err := ts.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	tile := &Tile{Rtree: &rtree.RTree{}}
	if err := tile.Load(body, ts.styler, ts.colorCache); err != nil {
		return &Tile{Rtree: &rtree.RTree{}}, nil
	}
	ts.mu.Lock()
	ts.cache[key] = tile
	ts.mu.Unlock()
	return tile, nil
}

func (t *Tile) Load(buffer []byte, styler *Styler, colorCache map[string]string) error {
	gz, err := gzip.NewReader(bytes.NewReader(buffer))
	var data []byte
	if err == nil {
		data, err = ioutil.ReadAll(gz)
		gz.Close()
	} else {
		data = buffer
	}
	if err != nil {
		return err
	}
	layers, err := mvt.Unmarshal(data)
	if err != nil {
		return err
	}
		fmt.Printf("  [Tile.Load] Tile contains %d layers. Processing...\n", len(layers)) // <-- ADD THIS
	for _, layer := range layers {
		fmt.Printf("    [Tile.Load] Processing layer: '%s' with %d features.\n", layer.Name, len(layer.Features))
		featureAddedCount := 0 // <-- ADD THIS
		for _, feature := range layer.Features {
			if feature.Properties == nil {
				feature.Properties = make(map[string]interface{})
			}
			switch feature.Geometry.(type) {
			case orb.Point, orb.MultiPoint:
				feature.Properties["$type"] = "Point"
			case orb.LineString, orb.MultiLineString:
				feature.Properties["$type"] = "LineString"
			case orb.Polygon, orb.MultiPolygon:
				feature.Properties["$type"] = "Polygon"
			}
			// END ADDED BLOCK

			style := styler.GetStyleFor(layer.Name, feature)
			if style == nil {
				if layer.Name == "water" {
					fmt.Printf("      - Style NOT FOUND for feature in layer '%s' with props: %v\n", layer.Name, feature.Properties)
				}
				continue
			}
			featureAddedCount++ // <-- ADD THIS
			colorStr := style.GetPaintProperty("line-color", style.GetPaintProperty("fill-color", "#ffffff"))
			colorCode, ok := colorCache[colorStr]
			if !ok {
				colorCode = hexToANSI(colorStr)
				colorCache[colorStr] = colorCode
			}
			label, _ := feature.Properties["name"].(string)
			styledFeat := &StyledFeature{
				Geometry: feature.Geometry, Style: style, Color: colorCode, Label: label,
			}
			bounds := feature.Geometry.Bound()
			fmt.Printf("      - Style '%s' FOUND. Inserting feature into R-Tree for layer '%s'.\n", style.ID, layer.Name)
			t.Rtree.Insert([2]float64{bounds.Min.X(), bounds.Min.Y()}, [2]float64{bounds.Max.X(), bounds.Max.Y()}, styledFeat)
		}
		// ADD THIS
		if featureAddedCount > 0 {
			fmt.Printf("    [Tile.Load] Finished processing layer '%s', ADDED %d features to R-Tree.\n", layer.Name, featureAddedCount)
		}
	}
	if len(layers) > 0 {
		t.Layer = *layers[0]
	}
	return nil
}

type StyleLayer struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	SourceLayer string                 `json:"source-layer"` // <-- FIX HERE
	Ref         string                 `json:"ref"`
	Filter      []interface{}          `json:"filter"`
	Paint       map[string]interface{} `json:"paint"`
	MinZoom     float64                `json:"minzoom"` // <-- FIX HERE
	MaxZoom     float64                `json:"maxzoom"` // <-- FIX HERE
	AppliesTo   func(feature *geojson.Feature) bool `json:"-"` // Ignore this field for JSON
}

func (sl *StyleLayer) GetPaintProperty(key, fallback string) string {
	if sl.Paint == nil {
		return fallback
	}
	if val, ok := sl.Paint[key].(string); ok {
		return val
	}
	return fallback
}

type Styler struct {
	styleByLayer map[string][]*StyleLayer
}

func newStyler(stylePath string) (*Styler, error) {
	data, err := ioutil.ReadFile(stylePath)
	if err != nil {
		return nil, err
	}
	var styleDef struct {
		Layers []*StyleLayer `json:"layers"`
	}
	if err := json.Unmarshal(data, &styleDef); err != nil {
		return nil, err
	}

	s := &Styler{styleByLayer: make(map[string][]*StyleLayer)}
	styleByID := make(map[string]*StyleLayer)
	for _, layer := range styleDef.Layers { // This is fine for populating the map
		styleByID[layer.ID] = layer
	}

	// --- THIS IS THE PART TO CHANGE ---
	for i := range styleDef.Layers {
		layer := styleDef.Layers[i] // layer is now a pointer to the actual slice element

		if layer.Ref != "" {
			if ref, ok := styleByID[layer.Ref]; ok {
				if layer.Type == "" {
					layer.Type = ref.Type
				}
				if layer.SourceLayer == "" {
					layer.SourceLayer = ref.SourceLayer
				}
				if layer.Filter == nil {
					layer.Filter = ref.Filter
				}
                // Also copy paint properties for refs
				if layer.Paint == nil {
					layer.Paint = ref.Paint
				}
			}
		}
		layer.AppliesTo = compileFilter(layer.Filter)
		s.styleByLayer[layer.SourceLayer] = append(s.styleByLayer[layer.SourceLayer], layer)
	}
	return s, nil
}

func (s *Styler) GetStyleFor(layerName string, feature *geojson.Feature) *StyleLayer {
	// Let's first check if there are ANY styles for this layer name
	styles, ok := s.styleByLayer[layerName]
	if !ok {
		// This will tell us if the layerName itself is the problem
		fmt.Printf("    -> No styles registered for source-layer '%s'\n", layerName)
		return nil
	}

	for _, style := range styles {
		// ADD THIS LOGGING BLOCK
		fmt.Printf("    - Checking style '%s' (filter: %v)... ", style.ID, style.Filter)
		if style.AppliesTo(feature) {
			fmt.Println("MATCH!")
			return style
		}
		fmt.Println("NO MATCH.")
		// END LOGGING BLOCK
	}
	return nil
}

// compileFilter recursively compiles a Mapbox GL filter expression into a function.
func compileFilter(filter []interface{}) func(f *geojson.Feature) bool {
	// If the filter is empty or nil, it passes all features.
	if len(filter) == 0 {
		return func(*geojson.Feature) bool { return true }
	}

	// The first element should be the operator string.
	op, ok := filter[0].(string)
	if !ok {
		// This is a malformed filter; fail open to match the original JS.
		return func(*geojson.Feature) bool { return true }
	}

	switch op {
	// Recursive cases
	case "all":
		var subFilters []func(*geojson.Feature) bool
		for _, subFilterExpr := range filter[1:] {
			if sub, ok := subFilterExpr.([]interface{}); ok {
				subFilters = append(subFilters, compileFilter(sub))
			}
		}
		return func(f *geojson.Feature) bool {
			for _, subf := range subFilters {
				if !subf(f) {
					return false // short-circuit on first failure
				}
			}
			return true
		}
	case "any":
		var subFilters []func(*geojson.Feature) bool
		for _, subFilterExpr := range filter[1:] {
			if sub, ok := subFilterExpr.([]interface{}); ok {
				subFilters = append(subFilters, compileFilter(sub))
			}
		}
		return func(f *geojson.Feature) bool {
			for _, subf := range subFilters {
				if subf(f) {
					return true // short-circuit on first success
				}
			}
			return false
		}
	case "none":
		var subFilters []func(*geojson.Feature) bool
		for _, subFilterExpr := range filter[1:] {
			if sub, ok := subFilterExpr.([]interface{}); ok {
				subFilters = append(subFilters, compileFilter(sub))
			}
		}
		return func(f *geojson.Feature) bool {
			for _, subf := range subFilters {
				if subf(f) {
					return false // short-circuit on first success (as it's a "none" check)
				}
			}
			return true
		}

	// Base cases
	case "==":
		if len(filter) < 3 { return func(*geojson.Feature) bool { return false } }
		key, _ := filter[1].(string)
		val := filter[2]
		return func(f *geojson.Feature) bool {
			propVal, propOk := f.Properties[key]
			// ADD THIS
			if key == "$type" {
				fmt.Printf(`      Filter: ["==", "%s", "%v"] -> Feature has '%s': %v. Result: %t`+"\n", key, val, key, propVal, propOk && propVal == val)
			}
			if !propOk {
				return false // Property doesn't exist, can't be equal
			}
			// Handle case where JSON numbers are float64
			if prop, ok := propVal.(float64); ok {
				if v, ok := val.(float64); ok { return prop == v }
				if v, ok := val.(int); ok { return prop == float64(v) }
			}
			return propVal == val
		}
	case "!=":
		if len(filter) < 3 { return func(*geojson.Feature) bool { return false } }
		key, _ := filter[1].(string)
		val := filter[2]
		return func(f *geojson.Feature) bool {
			// Handle case where JSON numbers are float64
			if prop, ok := f.Properties[key].(float64); ok {
				if v, ok := val.(float64); ok { return prop != v }
				if v, ok := val.(int); ok { return prop != float64(v) }
			}
			return f.Properties[key] != val
		}
	case "in":
		if len(filter) < 3 { return func(*geojson.Feature) bool { return false } }
		key, _ := filter[1].(string)
		values := make(map[interface{}]bool)
		for _, v := range filter[2:] {
			values[v] = true
		}
		return func(f *geojson.Feature) bool {
			prop, ok := f.Properties[key]
			return ok && values[prop]
		}
	case "!in":
		if len(filter) < 3 { return func(*geojson.Feature) bool { return false } }
		key, _ := filter[1].(string)
		values := make(map[interface{}]bool)
		for _, v := range filter[2:] {
			values[v] = true
		}
		return func(f *geojson.Feature) bool {
			prop, ok := f.Properties[key]
			return !ok || !values[prop]
		}
	case "has":
		if len(filter) < 2 { return func(*geojson.Feature) bool { return false } }
		key, _ := filter[1].(string)
		return func(f *geojson.Feature) bool {
			_, ok := f.Properties[key]
			return ok
		}
	case "!has":
		if len(filter) < 2 { return func(*geojson.Feature) bool { return false } }
		key, _ := filter[1].(string)
		return func(f *geojson.Feature) bool {
			_, ok := f.Properties[key]
			return !ok
		}

	default:
		// For any unknown operator, pass all features.
		return func(*geojson.Feature) bool { return true }
	}
}

type LabelBuffer struct {
	tree *rtree.RTree
}

func newLabelBuffer() *LabelBuffer {
	return &LabelBuffer{tree: &rtree.RTree{}}
}

func (lb *LabelBuffer) WriteIfPossible(text string, x, y int) bool {
	width := runewidth.StringWidth(text)
	bounds := [2][2]float64{{float64(x - 1), float64(y - 1)}, {float64(x + width + 1), float64(y + 1)}}
	var collision bool
	lb.tree.Search(bounds[0], bounds[1], func(_, _ [2]float64, _ interface{}) bool {
		collision = true
		return false
	})
	if !collision {
		lb.tree.Insert(bounds[0], bounds[1], text)
		return true
	}
	return false
}

func baseZoom(zoom float64) int { return int(math.Floor(zoom)) }

func tilesizeAtZoom(zoom float64) float64 {
	return ProjectSize * math.Pow(2, zoom-float64(baseZoom(zoom)))
}

func ll2tile(lon, lat float64, zoom int) (float64, float64) {
	latRad := lat * math.Pi / 180
	n := math.Pow(2, float64(zoom))
	xtile := (lon + 180) / 360 * n
	ytile := (1 - math.Asinh(math.Tan(latRad))/math.Pi) / 2 * n
	return xtile, ytile
}

func hexToANSI(hex string) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	var r, g, b uint8
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}
