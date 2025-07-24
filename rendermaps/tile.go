package rendermaps

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/mvt"
	"github.com/tidwall/rtree"

	_ "embed"
)

// parser and compiler ofr Mapbox Map Style files
// https://www.mapbox.com/mapbox-gl-style-spec/

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

//go:embed style.json
var styleJson []byte

var (
	gStyler = makeNewStyler()
	gTs = NewTileSource(TileSourceURL, gStyler)
)

func makeNewStyler() (*Styler) {
	styler, err := NewStyler(styleJson)
	if err != nil {
		log.Panicf("Failed to load embedded style file': %v", err)
	}
	return styler
}

// TODO migrate to cache directory

func NewTileSource(url string, styler *Styler) *TileSource {
	return &TileSource{
		url: url, client: &http.Client{Timeout: 10 * time.Second}, styler: styler,
		cache: make(map[string]*Tile), colorCache: make(map[string]string),
	}
}

func (ts *TileSource) GetTile(z, x, y int) (*Tile, error) {
	var err error
	var tile *Tile
	var ok bool
	
	key := fmt.Sprintf("%d-%d-%d", z, x, y)

	//bench := time.Now()
	ts.mu.Lock()
	if tile, ok = ts.cache[key]; ok {
		ts.mu.Unlock()
		// cached
	} else {
		ts.mu.Unlock()
		var body []byte

		if body, err = cacheGetKey(key); err == nil {
			// cached
		} else {
			url := fmt.Sprintf("%s%d/%d/%d.pbf", ts.url, z, x, y)
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("User-Agent", "MapSCII-Go-MVP/1.0")
			resp, err := ts.client.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			cacheInsertKey(key, body) // cache the response
		}

		ts.mu.Lock()
		tile = &Tile{Rtree: &rtree.RTree{}}
		if err := tile.Load(body, ts.styler, ts.colorCache); err != nil {
			return &Tile{Rtree: &rtree.RTree{}}, nil
		}

		ts.cache[key] = tile
		ts.mu.Unlock()
	}

	return tile, nil
}

func (t *Tile) Load(buffer []byte, styler *Styler, colorCache map[string]string) error {
	gz, err := gzip.NewReader(bytes.NewReader(buffer))
	var data []byte
	if err == nil {
		data, err = io.ReadAll(gz)
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
	for _, layer := range layers {
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

			style := styler.GetStyleFor(layer.Name, feature)
			if style == nil {
				continue
			}

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
			t.Rtree.Insert([2]float64{bounds.Min.X(), bounds.Min.Y()}, [2]float64{bounds.Max.X(), bounds.Max.Y()}, styledFeat)
		}
	}
	if len(layers) > 0 {
		t.Layer = *layers[0]
	}
	return nil
}

type LabelBuffer struct {
	tree *rtree.RTree
}

func NewLabelBuffer() *LabelBuffer {
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

func baseZoom(zoom float64) int {
	return min(int(TileRange), int(math.Floor(zoom)))
}

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
