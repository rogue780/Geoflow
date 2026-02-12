// Package s2 provides S2 spherical geometry cell functions for GeoFlow.
package s2

import (
	"fmt"
	"math"
	"strconv"

	"github.com/rogue780/geoflow/internal/object"
)

// GetExports returns all exported functions for the geo.s2 module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		"fromLatLon": &object.Builtin{Name: "s2.fromLatLon", Fn: s2FromLatLon},
		"level":      &object.Builtin{Name: "s2.level", Fn: s2Level},
		"isValid":    &object.Builtin{Name: "s2.isValid", Fn: s2IsValid},
		"toPoint":    &object.Builtin{Name: "s2.toPoint", Fn: s2ToPoint},
		"contains":   &object.Builtin{Name: "s2.contains", Fn: s2Contains},
		"toToken":    &object.Builtin{Name: "s2.toToken", Fn: s2ToToken},
		"fromToken":  &object.Builtin{Name: "s2.fromToken", Fn: s2FromToken},
	}
}

// maxLevel is the maximum S2 cell level (0-30).
const maxLevel = 30

// s2Hash encodes lat/lon at a given level into a uint64 by quantizing coordinates.
// This is a simplified spatial hash, not the real S2 algorithm.
// The encoding uses: top 5 bits for level, remaining bits for face + position.
func s2Hash(lat, lon float64, level int) uint64 {
	// Determine face (0-5) of the cube based on the dominant axis
	// Convert lat/lon to 3D unit sphere coordinates
	latRad := lat * math.Pi / 180.0
	lonRad := lon * math.Pi / 180.0
	x := math.Cos(latRad) * math.Cos(lonRad)
	y := math.Cos(latRad) * math.Sin(lonRad)
	z := math.Sin(latRad)

	ax := math.Abs(x)
	ay := math.Abs(y)
	az := math.Abs(z)

	var face uint64
	var u, v float64
	if ax >= ay && ax >= az {
		if x > 0 {
			face = 0
		} else {
			face = 3
		}
		u = y / ax
		v = z / ax
	} else if ay >= ax && ay >= az {
		if y > 0 {
			face = 1
		} else {
			face = 4
		}
		u = x / ay
		v = z / ay
	} else {
		if z > 0 {
			face = 2
		} else {
			face = 5
		}
		u = x / az
		v = y / az
	}

	// Normalize u, v from [-1, 1] to [0, 1]
	normU := (u + 1.0) / 2.0
	normV := (v + 1.0) / 2.0
	normU = math.Max(0, math.Min(1, normU))
	normV = math.Max(0, math.Min(1, normV))

	// Quantize based on level
	gridSize := uint64(1) << uint(level)
	uCell := uint64(normU * float64(gridSize))
	vCell := uint64(normV * float64(gridSize))
	if uCell >= gridSize {
		uCell = gridSize - 1
	}
	if vCell >= gridSize {
		vCell = gridSize - 1
	}

	// Encode: level in top 5 bits, face in next 3 bits, then u/v cells
	return (uint64(level) << 59) | (face << 56) | (uCell << 28) | vCell
}

// s2DecodeCells extracts the level, face, u cell, and v cell from an S2 cell ID.
func s2DecodeCells(cellID uint64) (level int, face uint64, uCell, vCell uint64) {
	level = int((cellID >> 59) & 0x1F)
	face = (cellID >> 56) & 0x07
	uCell = (cellID >> 28) & 0x0FFFFFFF
	vCell = cellID & 0x0FFFFFFF
	return
}

// s2CellToLatLon converts S2 cell components back to approximate lat/lon.
func s2CellToLatLon(level int, face, uCell, vCell uint64) (float64, float64) {
	gridSize := float64(uint64(1) << uint(level))
	normU := (float64(uCell) + 0.5) / gridSize
	normV := (float64(vCell) + 0.5) / gridSize

	// Map back from [0, 1] to [-1, 1]
	u := normU*2.0 - 1.0
	v := normV*2.0 - 1.0

	// Recover 3D coordinates from face, u, v
	var x, y, z float64
	switch face {
	case 0:
		x, y, z = 1, u, v
	case 1:
		x, y, z = u, 1, v
	case 2:
		x, y, z = u, v, 1
	case 3:
		x, y, z = -1, u, v
	case 4:
		x, y, z = u, -1, v
	case 5:
		x, y, z = u, v, -1
	}

	// Normalize to unit sphere
	mag := math.Sqrt(x*x + y*y + z*z)
	if mag == 0 {
		return 0, 0
	}
	x /= mag
	y /= mag
	z /= mag

	lat := math.Asin(z) * 180.0 / math.Pi
	lon := math.Atan2(y, x) * 180.0 / math.Pi
	return lat, lon
}

// s2FromLatLon converts lat, lon, level into an S2CellId.
func s2FromLatLon(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "s2.fromLatLon expects 3 arguments (lat, lon, level)"}
	}

	lat, err := toFloat64(args[0])
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("s2.fromLatLon: lat: %s", err)}
	}
	lon, errL := toFloat64(args[1])
	if errL != nil {
		return &object.Error{Message: fmt.Sprintf("s2.fromLatLon: lon: %s", errL)}
	}
	lvl, ok := args[2].(*object.Integer)
	if !ok {
		return &object.Error{Message: "s2.fromLatLon: level must be an integer"}
	}
	if lvl.Value < 0 || lvl.Value > maxLevel {
		return &object.Error{Message: fmt.Sprintf("s2.fromLatLon: level must be 0-%d", maxLevel)}
	}

	hash := s2Hash(lat, lon, int(lvl.Value))
	return &object.S2CellId{Value: hash, Level: int(lvl.Value)}
}

// s2Level returns the level of an S2CellId.
func s2Level(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "s2.level expects 1 argument (S2CellId)"}
	}
	s, ok := args[0].(*object.S2CellId)
	if !ok {
		return &object.Error{Message: "s2.level: argument must be an S2CellId"}
	}
	return &object.Integer{Value: int64(s.Level)}
}

// s2IsValid checks whether an S2CellId has a valid level and face.
func s2IsValid(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "s2.isValid expects 1 argument (S2CellId)"}
	}
	s, ok := args[0].(*object.S2CellId)
	if !ok {
		return object.NativeBoolToBooleanObject(false)
	}
	valid := s.Level >= 0 && s.Level <= maxLevel
	if valid {
		_, face, _, _ := s2DecodeCells(s.Value)
		valid = face <= 5
	}
	return object.NativeBoolToBooleanObject(valid)
}

// s2ToPoint returns the centroid Point of an S2 cell.
func s2ToPoint(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "s2.toPoint expects 1 argument (S2CellId)"}
	}
	s, ok := args[0].(*object.S2CellId)
	if !ok {
		return &object.Error{Message: "s2.toPoint: argument must be an S2CellId"}
	}

	level, face, uCell, vCell := s2DecodeCells(s.Value)
	lat, lon := s2CellToLatLon(level, face, uCell, vCell)
	return &object.Point{Coord: object.Coordinate{X: lon, Y: lat}}
}

// s2Contains checks whether a cell contains a given point.
// It does this by encoding the point at the same level and comparing cells.
func s2Contains(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "s2.contains expects 2 arguments (S2CellId, Point)"}
	}
	s, ok := args[0].(*object.S2CellId)
	if !ok {
		return &object.Error{Message: "s2.contains: first argument must be an S2CellId"}
	}
	pt, ok := args[1].(*object.Point)
	if !ok {
		return &object.Error{Message: "s2.contains: second argument must be a Point"}
	}

	// Encode the point at the same level as the cell
	pointHash := s2Hash(pt.Coord.Y, pt.Coord.X, s.Level)
	return object.NativeBoolToBooleanObject(pointHash == s.Value)
}

// s2ToToken converts an S2CellId to a hex string token.
func s2ToToken(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "s2.toToken expects 1 argument (S2CellId)"}
	}
	s, ok := args[0].(*object.S2CellId)
	if !ok {
		return &object.Error{Message: "s2.toToken: argument must be an S2CellId"}
	}
	return &object.String{Value: fmt.Sprintf("%016x", s.Value)}
}

// s2FromToken parses a hex token string into an S2CellId.
func s2FromToken(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "s2.fromToken expects 1 argument (string)"}
	}
	str, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "s2.fromToken: argument must be a string"}
	}

	val, err := strconv.ParseUint(str.Value, 16, 64)
	if err != nil {
		return &object.Result{
			Value: &object.String{Value: fmt.Sprintf("invalid S2 token: %s", err)},
			IsOk:  false,
		}
	}

	level := int((val >> 59) & 0x1F)
	return &object.Result{
		Value: &object.S2CellId{Value: val, Level: level},
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
