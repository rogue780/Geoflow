// Package h3 provides H3 hexagonal hierarchical spatial index functions for GeoFlow.
package h3

import (
	"fmt"
	"math"
	"strconv"

	"github.com/rogue780/geoflow/internal/object"
)

// GetExports returns all exported functions for the geo.h3 module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		"fromLatLon":  &object.Builtin{Name: "h3.fromLatLon", Fn: h3FromLatLon},
		"resolution":  &object.Builtin{Name: "h3.resolution", Fn: h3Resolution},
		"isValid":     &object.Builtin{Name: "h3.isValid", Fn: h3IsValid},
		"toPoint":     &object.Builtin{Name: "h3.toPoint", Fn: h3ToPoint},
		"kRing":       &object.Builtin{Name: "h3.kRing", Fn: h3KRing},
		"distance":    &object.Builtin{Name: "h3.distance", Fn: h3Distance},
		"toString":    &object.Builtin{Name: "h3.toString", Fn: h3ToString},
		"fromString":  &object.Builtin{Name: "h3.fromString", Fn: h3FromString},
	}
}

// maxResolution is the maximum supported H3 resolution (0-15).
const maxResolution = 15

// h3Hash encodes lat/lon at a given resolution into a uint64 by quantizing coordinates.
// This is a simplified spatial hash, not the real H3 algorithm.
func h3Hash(lat, lon float64, resolution int) uint64 {
	// Normalize lat [-90, 90] and lon [-180, 180] into [0, 1] range
	normLat := (lat + 90.0) / 180.0
	normLon := (lon + 180.0) / 360.0

	// Clamp to valid range
	normLat = math.Max(0, math.Min(1, normLat))
	normLon = math.Max(0, math.Min(1, normLon))

	// Quantize based on resolution: higher resolution = more grid cells
	gridSize := uint64(1) << uint(resolution+1) // 2^(res+1) cells per axis
	latCell := uint64(normLat * float64(gridSize))
	lonCell := uint64(normLon * float64(gridSize))

	// Clamp cell indices
	if latCell >= gridSize {
		latCell = gridSize - 1
	}
	if lonCell >= gridSize {
		lonCell = gridSize - 1
	}

	// Encode: resolution in top 4 bits, then interleave lat/lon cells
	// Top 4 bits: resolution (0-15), next 30 bits: lat cells, bottom 30 bits: lon cells
	return (uint64(resolution) << 60) | (latCell << 30) | lonCell
}

// h3DecodeCells extracts the resolution, lat cell, and lon cell from an H3 index.
func h3DecodeCells(index uint64) (resolution int, latCell, lonCell uint64) {
	resolution = int((index >> 60) & 0x0F)
	latCell = (index >> 30) & 0x3FFFFFFF
	lonCell = index & 0x3FFFFFFF
	return
}

// h3CellToLatLon converts H3 cell indices back to approximate lat/lon (cell center).
func h3CellToLatLon(resolution int, latCell, lonCell uint64) (float64, float64) {
	gridSize := float64(uint64(1) << uint(resolution+1))
	lat := (float64(latCell)+0.5)/gridSize*180.0 - 90.0
	lon := (float64(lonCell)+0.5)/gridSize*360.0 - 180.0
	return lat, lon
}

// h3FromLatLon converts lat, lon, resolution into an H3Index.
func h3FromLatLon(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "h3.fromLatLon expects 3 arguments (lat, lon, resolution)"}
	}

	lat, err := toFloat64(args[0])
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("h3.fromLatLon: lat: %s", err)}
	}
	lon, errL := toFloat64(args[1])
	if errL != nil {
		return &object.Error{Message: fmt.Sprintf("h3.fromLatLon: lon: %s", errL)}
	}
	res, ok := args[2].(*object.Integer)
	if !ok {
		return &object.Error{Message: "h3.fromLatLon: resolution must be an integer"}
	}
	if res.Value < 0 || res.Value > maxResolution {
		return &object.Error{Message: fmt.Sprintf("h3.fromLatLon: resolution must be 0-%d", maxResolution)}
	}

	hash := h3Hash(lat, lon, int(res.Value))
	return &object.H3Index{Value: hash, Resolution: int(res.Value)}
}

// h3Resolution returns the resolution of an H3Index.
func h3Resolution(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "h3.resolution expects 1 argument (H3Index)"}
	}
	h, ok := args[0].(*object.H3Index)
	if !ok {
		return &object.Error{Message: "h3.resolution: argument must be an H3Index"}
	}
	return &object.Integer{Value: int64(h.Resolution)}
}

// h3IsValid checks whether an H3Index has a valid resolution.
func h3IsValid(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "h3.isValid expects 1 argument (H3Index)"}
	}
	h, ok := args[0].(*object.H3Index)
	if !ok {
		return object.NativeBoolToBooleanObject(false)
	}
	valid := h.Resolution >= 0 && h.Resolution <= maxResolution
	return object.NativeBoolToBooleanObject(valid)
}

// h3ToPoint returns the centroid Point of an H3 cell.
func h3ToPoint(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "h3.toPoint expects 1 argument (H3Index)"}
	}
	h, ok := args[0].(*object.H3Index)
	if !ok {
		return &object.Error{Message: "h3.toPoint: argument must be an H3Index"}
	}

	res, latCell, lonCell := h3DecodeCells(h.Value)
	lat, lon := h3CellToLatLon(res, latCell, lonCell)
	return &object.Point{Coord: object.Coordinate{X: lon, Y: lat}}
}

// h3KRing returns the k-ring (disk) of H3 indices around a given index.
// Simplified: returns a list containing neighbors by offsetting lat/lon cells.
func h3KRing(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "h3.kRing expects 2 arguments (H3Index, k)"}
	}
	h, ok := args[0].(*object.H3Index)
	if !ok {
		return &object.Error{Message: "h3.kRing: first argument must be an H3Index"}
	}
	k, ok := args[1].(*object.Integer)
	if !ok {
		return &object.Error{Message: "h3.kRing: second argument must be an integer"}
	}

	if k.Value < 0 {
		return &object.Error{Message: "h3.kRing: k must be non-negative"}
	}

	resolution, latCell, lonCell := h3DecodeCells(h.Value)
	gridSize := uint64(1) << uint(resolution+1)
	kVal := int(k.Value)

	var results []object.Object
	for di := -kVal; di <= kVal; di++ {
		for dj := -kVal; dj <= kVal; dj++ {
			newLat := int64(latCell) + int64(di)
			newLon := int64(lonCell) + int64(dj)
			// Wrap around
			if newLat < 0 {
				newLat = 0
			}
			if newLat >= int64(gridSize) {
				newLat = int64(gridSize) - 1
			}
			if newLon < 0 {
				newLon += int64(gridSize)
			}
			if newLon >= int64(gridSize) {
				newLon -= int64(gridSize)
			}
			idx := (uint64(resolution) << 60) | (uint64(newLat) << 30) | uint64(newLon)
			results = append(results, &object.H3Index{Value: idx, Resolution: resolution})
		}
	}

	return &object.List{Elements: results}
}

// h3Distance computes the approximate grid distance between two H3 indices.
func h3Distance(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "h3.distance expects 2 arguments (H3Index, H3Index)"}
	}
	a, ok := args[0].(*object.H3Index)
	if !ok {
		return &object.Error{Message: "h3.distance: first argument must be an H3Index"}
	}
	b, ok := args[1].(*object.H3Index)
	if !ok {
		return &object.Error{Message: "h3.distance: second argument must be an H3Index"}
	}

	_, aLat, aLon := h3DecodeCells(a.Value)
	_, bLat, bLon := h3DecodeCells(b.Value)

	dLat := int64(aLat) - int64(bLat)
	dLon := int64(aLon) - int64(bLon)
	if dLat < 0 {
		dLat = -dLat
	}
	if dLon < 0 {
		dLon = -dLon
	}

	// Chebyshev distance (L-inf) as grid distance
	dist := dLat
	if dLon > dist {
		dist = dLon
	}
	return &object.Integer{Value: dist}
}

// h3ToString converts an H3Index to a hex string representation.
func h3ToString(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "h3.toString expects 1 argument (H3Index)"}
	}
	h, ok := args[0].(*object.H3Index)
	if !ok {
		return &object.Error{Message: "h3.toString: argument must be an H3Index"}
	}
	return &object.String{Value: fmt.Sprintf("%016x", h.Value)}
}

// h3FromString parses a hex string into an H3Index.
func h3FromString(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "h3.fromString expects 1 argument (string)"}
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "h3.fromString: argument must be a string"}
	}

	val, err := strconv.ParseUint(s.Value, 16, 64)
	if err != nil {
		return &object.Result{
			Value: &object.String{Value: fmt.Sprintf("invalid H3 hex string: %s", err)},
			IsOk:  false,
		}
	}

	resolution := int((val >> 60) & 0x0F)
	return &object.Result{
		Value: &object.H3Index{Value: val, Resolution: resolution},
		IsOk:  true,
	}
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
