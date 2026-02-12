// Package raster provides raster grid operations for the GeoFlow language.
package raster

import (
	"fmt"

	"github.com/rogue780/geoflow/internal/object"
)

// GetExports returns all exported functions for the geo.raster module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		"empty":     &object.Builtin{Name: "raster.empty", Fn: rasterEmpty},
		"fromArray": &object.Builtin{Name: "raster.fromArray", Fn: rasterFromArray},
	}
}

// rasterEmpty creates an empty raster with given dimensions and an optional fill value.
// raster.empty(width, height, boundsMap, value=0)
func rasterEmpty(args ...object.Object) object.Object {
	if len(args) < 3 || len(args) > 4 {
		return &object.Error{Message: "raster.empty expects 3-4 arguments (width, height, bounds, [value])"}
	}

	width, ok := args[0].(*object.Integer)
	if !ok {
		return &object.Error{Message: "raster.empty: width must be an integer"}
	}
	height, ok := args[1].(*object.Integer)
	if !ok {
		return &object.Error{Message: "raster.empty: height must be an integer"}
	}

	if width.Value <= 0 || height.Value <= 0 {
		return &object.Error{Message: "raster.empty: width and height must be positive"}
	}

	bounds, err := parseBounds(args[2])
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("raster.empty: %s", err)}
	}

	fillValue := 0.0
	if len(args) == 4 {
		switch v := args[3].(type) {
		case *object.Float:
			fillValue = v.Value
		case *object.Integer:
			fillValue = float64(v.Value)
		default:
			return &object.Error{Message: "raster.empty: fill value must be a number"}
		}
	}

	w := int(width.Value)
	h := int(height.Value)
	data := make([][]float64, h)
	for i := range data {
		row := make([]float64, w)
		if fillValue != 0 {
			for j := range row {
				row[j] = fillValue
			}
		}
		data[i] = row
	}

	return &object.Raster{
		Data:   data,
		Width:  w,
		Height: h,
		Bounds: bounds,
	}
}

// rasterFromArray creates a raster from a flat data list + width + height.
// raster.fromArray(data, width, height)
func rasterFromArray(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "raster.fromArray expects 3 arguments (data, width, height)"}
	}

	dataList, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "raster.fromArray: data must be a list"}
	}
	width, ok := args[1].(*object.Integer)
	if !ok {
		return &object.Error{Message: "raster.fromArray: width must be an integer"}
	}
	height, ok := args[2].(*object.Integer)
	if !ok {
		return &object.Error{Message: "raster.fromArray: height must be an integer"}
	}

	w := int(width.Value)
	h := int(height.Value)

	if w <= 0 || h <= 0 {
		return &object.Error{Message: "raster.fromArray: width and height must be positive"}
	}

	expected := w * h
	if len(dataList.Elements) != expected {
		return &object.Error{Message: fmt.Sprintf("raster.fromArray: data list length (%d) must equal width*height (%d)", len(dataList.Elements), expected)}
	}

	data := make([][]float64, h)
	idx := 0
	for i := 0; i < h; i++ {
		row := make([]float64, w)
		for j := 0; j < w; j++ {
			val, err := toFloat64(dataList.Elements[idx])
			if err != nil {
				return &object.Error{Message: fmt.Sprintf("raster.fromArray: element at index %d: %s", idx, err)}
			}
			row[j] = val
			idx++
		}
		data[i] = row
	}

	return &object.Raster{
		Data:   data,
		Width:  w,
		Height: h,
	}
}

// parseBounds extracts a BBox from a Map with keys "minX", "minY", "maxX", "maxY"
// or directly from a BBox object.
func parseBounds(obj object.Object) (*object.BBox, error) {
	switch b := obj.(type) {
	case *object.BBox:
		return b, nil
	case *object.Map:
		minX, err := getMapFloat(b, "minX")
		if err != nil {
			return nil, fmt.Errorf("bounds: %s", err)
		}
		minY, err := getMapFloat(b, "minY")
		if err != nil {
			return nil, fmt.Errorf("bounds: %s", err)
		}
		maxX, err := getMapFloat(b, "maxX")
		if err != nil {
			return nil, fmt.Errorf("bounds: %s", err)
		}
		maxY, err := getMapFloat(b, "maxY")
		if err != nil {
			return nil, fmt.Errorf("bounds: %s", err)
		}
		return &object.BBox{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}, nil
	default:
		return nil, fmt.Errorf("bounds must be a BBox or a Map with minX, minY, maxX, maxY keys")
	}
}

// getMapFloat extracts a float64 value from a Map by string key.
func getMapFloat(m *object.Map, key string) (float64, error) {
	v, ok := m.Get(key)
	if !ok {
		return 0, fmt.Errorf("missing key %q", key)
	}
	return toFloat64(v)
}

// toFloat64 converts an Integer or Float to float64.
func toFloat64(obj object.Object) (float64, error) {
	switch v := obj.(type) {
	case *object.Float:
		return v.Value, nil
	case *object.Integer:
		return float64(v.Value), nil
	default:
		return 0, fmt.Errorf("expected a number, got %s", obj.Type())
	}
}
