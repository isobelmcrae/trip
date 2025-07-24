package rendermaps

import (
	"encoding/json"
	"fmt"
	"reflect"

	_ "embed"

	"github.com/paulmach/orb/geojson"
)

type StyleLayer struct {
	ID          string         `json:"id"`
	Ref         string         `json:"ref"`
	Type        string         `json:"type"`
	SourceLayer string         `json:"source-layer"`
	MinZoom     float64        `json:"minzoom"`
	MaxZoom     float64        `json:"maxzoom"`
	Filter      []any          `json:"filter"`
	Paint       map[string]any `json:"paint"`
	Layout      map[string]any `json:"layout"`

	AppliesTo FilterFunc
}

type styleJSON struct {
	Name      string            `json:"name"`
	Constants map[string]string `json:"constants"`
	Layers    []json.RawMessage `json:"layers"`
}

type Styler struct {
	styleByID    map[string]*StyleLayer
	styleByLayer map[string][]*StyleLayer
	StyleName    string
}

func NewStyler(styleData []byte) (*Styler, error) {
	var s styleJSON
	if err := json.Unmarshal(styleData, &s); err != nil {
		return nil, fmt.Errorf("failed to unmarshal style: %w", err)
	}

	styler := &Styler{
		styleByID:    make(map[string]*StyleLayer),
		styleByLayer: make(map[string][]*StyleLayer),
		StyleName:    s.Name,
	}

	if s.Constants == nil {
		s.Constants = make(map[string]string)
	}

	for _, rawLayer := range s.Layers {
		// 1. Replace constants in the raw JSON of the layer.
		processedLayerBytes, err := replaceConstantsInJSON(s.Constants, rawLayer)
		if err != nil {
			return nil, fmt.Errorf("failed to replace constants for layer: %w", err)
		}

		var layer StyleLayer
		if err := json.Unmarshal(processedLayerBytes, &layer); err != nil {
			return nil, fmt.Errorf("failed to unmarshal layer from %s: %w", string(processedLayerBytes), err)
		}

		// 2. Handle layer referencing (`ref` property).
		if layer.Ref != "" {
			if refLayer, ok := styler.styleByID[layer.Ref]; ok {
				// Inherit properties from the referenced layer if they are not set.
				if layer.Type == "" {
					layer.Type = refLayer.Type
				}
				if layer.SourceLayer == "" {
					layer.SourceLayer = refLayer.SourceLayer
				}
				if layer.MinZoom == 0 && refLayer.MinZoom != 0 {
					layer.MinZoom = refLayer.MinZoom
				}
				if layer.MaxZoom == 0 && refLayer.MaxZoom != 0 {
					layer.MaxZoom = refLayer.MaxZoom
				}
				if layer.Filter == nil {
					layer.Filter = refLayer.Filter
				}
			}
		}

		// 3. Compile the filter expression into a function.
		layer.AppliesTo, err = CompileFilter(layer.Filter)
		if err != nil {
			layer.AppliesTo = func(f *geojson.Feature) bool { return true }
		}

		// 4. Index the processed layer for fast lookup.
		if layer.SourceLayer != "" {
			styler.styleByLayer[layer.SourceLayer] = append(styler.styleByLayer[layer.SourceLayer], &layer)
		}
		styler.styleByID[layer.ID] = &layer
	}

	return styler, nil
}

// can return nil
func (s *Styler) GetStyleFor(sourceLayerName string, feature *geojson.Feature) *StyleLayer {
	layers, ok := s.styleByLayer[sourceLayerName]
	if !ok {
		return nil
	}

	for _, layer := range layers {
		if layer.AppliesTo(feature) {
			return layer
		}
	}

	return nil
}

func (l *StyleLayer) GetPaintProperty(key string, defaultValue string) string {
	if l.Paint == nil {
		return defaultValue
	}
	if value, ok := l.Paint[key]; ok {
		if strValue, isString := value.(string); isString {
			return strValue
		}
	}
	return defaultValue
}

func replaceConstantsInJSON(constants map[string]string, raw json.RawMessage) (json.RawMessage, error) {
	var node any
	if err := json.Unmarshal(raw, &node); err != nil {
		return nil, err
	}
	doReplace(constants, node)
	return json.Marshal(node)
}

func doReplace(constants map[string]string, node any) {
	switch n := node.(type) {
	case map[string]any:
		for key, val := range n {
			if strVal, ok := val.(string); ok && len(strVal) > 0 && strVal[0] == '@' {
				if constVal, found := constants[strVal]; found {
					n[key] = constVal
				}
			} else {
				doReplace(constants, val) // Recurse
			}
		}
	case []any:
		for _, item := range n {
			doReplace(constants, item) // Recurse
		}
	}
}

type FilterFunc func(feature *geojson.Feature) bool

func CompileFilter(filter []any) (FilterFunc, error) {
	if len(filter) == 0 {
		return func(f *geojson.Feature) bool { return true }, nil
	}

	op, ok := filter[0].(string)
	if !ok {
		return nil, fmt.Errorf("filter operator must be a string, got %T", filter[0])
	}

	switch op {
	case "all":
		subFilters, err := compileSubFilters(filter)
		if err != nil {
			return nil, err
		}
		return func(f *geojson.Feature) bool {
			for _, sf := range subFilters {
				if !sf(f) {
					return false
				}
			}
			return true
		}, nil

	case "any":
		subFilters, err := compileSubFilters(filter)
		if err != nil {
			return nil, err
		}
		return func(f *geojson.Feature) bool {
			for _, sf := range subFilters {
				if sf(f) {
					return true
				}
			}
			return false
		}, nil

	case "none":
		subFilters, err := compileSubFilters(filter)
		if err != nil {
			return nil, err
		}
		return func(f *geojson.Feature) bool {
			for _, sf := range subFilters {
				if sf(f) {
					return false
				}
			}
			return true
		}, nil

	case "==", "!=":
		if len(filter) != 3 {
			return nil, fmt.Errorf("'%s' filter expects 2 arguments, got %d", op, len(filter)-1)
		}
		key, ok := filter[1].(string)
		if !ok {
			return nil, fmt.Errorf("'%s' filter key must be a string, got %T", op, filter[1])
		}
		val := filter[2]
		return func(f *geojson.Feature) bool {
			prop, exists := f.Properties[key]
			isEqual := exists && reflect.DeepEqual(prop, val)
			if op == "==" {
				return isEqual
			}
			return !isEqual
		}, nil

	case ">", ">=", "<", "<=":
		if len(filter) != 3 {
			return nil, fmt.Errorf("'%s' filter expects 2 arguments", op)
		}
		key, ok := filter[1].(string)
		if !ok {
			return nil, fmt.Errorf("'%s' filter key must be a string", op)
		}
		filterVal, ok := filter[2].(float64) // JSON numbers are parsed as float64
		if !ok {
			return nil, fmt.Errorf("'%s' filter value must be a number, got %T", op, filter[2])
		}
		return func(f *geojson.Feature) bool {
			prop, exists := f.Properties[key]
			if !exists {
				return false
			}
			propVal, ok := prop.(float64)
			if !ok { // Property exists but is not a number
				return false
			}
			switch op {
			case ">":
				return propVal > filterVal
			case ">=":
				return propVal >= filterVal
			case "<":
				return propVal < filterVal
			case "<=":
				return propVal <= filterVal
			}
			return false
		}, nil

	case "in", "!in":
		if len(filter) < 3 {
			return nil, fmt.Errorf("'%s' filter expects at least 2 arguments", op)
		}
		key, ok := filter[1].(string)
		if !ok {
			return nil, fmt.Errorf("'%s' filter key must be a string", op)
		}
		values := make(map[any]struct{})
		for _, v := range filter[2:] {
			values[v] = struct{}{}
		}
		return func(f *geojson.Feature) bool {
			prop, exists := f.Properties[key]
			if !exists {
				return op == "!in" // 'in' is false, '!in' is true if key is missing
			}
			_, found := values[prop]
			if op == "in" {
				return found
			}
			return !found
		}, nil

	case "has", "!has":
		if len(filter) != 2 {
			return nil, fmt.Errorf("'%s' filter expects 1 argument", op)
		}
		key, ok := filter[1].(string)
		if !ok {
			return nil, fmt.Errorf("'%s' filter key must be a string", op)
		}
		return func(f *geojson.Feature) bool {
			_, exists := f.Properties[key]
			if op == "has" {
				return exists
			}
			return !exists
		}, nil

	default:
		return nil, fmt.Errorf("unsupported filter operator: %s", op)
	}
}

func compileSubFilters(filter []any) ([]FilterFunc, error) {
	subFilters := make([]FilterFunc, 0, len(filter)-1)
	for i, subFilterExpr := range filter[1:] {
		expr, ok := subFilterExpr.([]any)
		if !ok {
			return nil, fmt.Errorf("sub-filter at index %d is not a valid expression", i)
		}
		subFilter, err := CompileFilter(expr)
		if err != nil {
			return nil, err
		}
		subFilters = append(subFilters, subFilter)
	}
	return subFilters, nil
}