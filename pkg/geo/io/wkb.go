package io

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strings"

	"github.com/rogue780/geoflow/internal/object"
)

// WKB geometry type constants.
const (
	wkbPoint      uint32 = 1
	wkbLineString uint32 = 2
	wkbPolygon    uint32 = 3
)

// WKB byte order: little-endian.
const wkbLittleEndian byte = 0x01

// GetWKBExports returns all WKB I/O functions.
func GetWKBExports() map[string]object.Object {
	return map[string]object.Object{
		"toWKB":   &object.Builtin{Name: "geoio.toWKB", Fn: toWKB},
		"fromWKB": &object.Builtin{Name: "geoio.fromWKB", Fn: fromWKB},
	}
}

// ── WKB Output ──

// toWKB converts a geometry to a hex-encoded WKB string.
func toWKB(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "geoio.toWKB expects 1 argument (geometry)"}
	}
	geom, ok := args[0].(object.Geometry)
	if !ok {
		return &object.Error{Message: "geoio.toWKB: argument must be a geometry"}
	}

	data, err := geometryToWKB(geom)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("WKB encode error: %s", err)}
	}
	return &object.String{Value: strings.ToUpper(hex.EncodeToString(data))}
}

func geometryToWKB(geom object.Geometry) ([]byte, error) {
	var buf bytes.Buffer
	switch g := geom.(type) {
	case *object.Point:
		buf.WriteByte(wkbLittleEndian)
		if err := binary.Write(&buf, binary.LittleEndian, wkbPoint); err != nil {
			return nil, err
		}
		if err := binary.Write(&buf, binary.LittleEndian, g.Coord.X); err != nil {
			return nil, err
		}
		if err := binary.Write(&buf, binary.LittleEndian, g.Coord.Y); err != nil {
			return nil, err
		}
	case *object.LineString:
		buf.WriteByte(wkbLittleEndian)
		if err := binary.Write(&buf, binary.LittleEndian, wkbLineString); err != nil {
			return nil, err
		}
		numPoints := uint32(len(g.Coords))
		if err := binary.Write(&buf, binary.LittleEndian, numPoints); err != nil {
			return nil, err
		}
		for _, c := range g.Coords {
			if err := binary.Write(&buf, binary.LittleEndian, c.X); err != nil {
				return nil, err
			}
			if err := binary.Write(&buf, binary.LittleEndian, c.Y); err != nil {
				return nil, err
			}
		}
	case *object.Polygon:
		buf.WriteByte(wkbLittleEndian)
		if err := binary.Write(&buf, binary.LittleEndian, wkbPolygon); err != nil {
			return nil, err
		}
		numRings := uint32(1 + len(g.InteriorRings))
		if err := binary.Write(&buf, binary.LittleEndian, numRings); err != nil {
			return nil, err
		}
		// Write exterior ring
		if err := writeWKBRing(&buf, g.ExteriorRing); err != nil {
			return nil, err
		}
		// Write interior rings
		for _, ring := range g.InteriorRings {
			if err := writeWKBRing(&buf, ring); err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("unsupported geometry type for WKB: %s", geom.GeomType())
	}
	return buf.Bytes(), nil
}

func writeWKBRing(buf *bytes.Buffer, ring []object.Coordinate) error {
	numPoints := uint32(len(ring))
	if err := binary.Write(buf, binary.LittleEndian, numPoints); err != nil {
		return err
	}
	for _, c := range ring {
		if err := binary.Write(buf, binary.LittleEndian, c.X); err != nil {
			return err
		}
		if err := binary.Write(buf, binary.LittleEndian, c.Y); err != nil {
			return err
		}
	}
	return nil
}

// ── WKB Input ──

// fromWKB parses a hex-encoded WKB string and returns a geometry object.
func fromWKB(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "geoio.fromWKB expects 1 argument (hex string)"}
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "geoio.fromWKB: argument must be a string"}
	}

	data, err := hex.DecodeString(s.Value)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("WKB hex decode error: %s", err)}
	}

	geom, err := parseWKBGeometry(data)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("WKB parse error: %s", err)}
	}
	return geom
}

func parseWKBGeometry(data []byte) (object.Geometry, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("WKB data too short")
	}

	byteOrder := data[0]
	var order binary.ByteOrder
	switch byteOrder {
	case 0x01:
		order = binary.LittleEndian
	case 0x00:
		order = binary.BigEndian
	default:
		return nil, fmt.Errorf("invalid WKB byte order: 0x%02x", byteOrder)
	}

	geomType := order.Uint32(data[1:5])
	reader := bytes.NewReader(data[5:])

	switch geomType {
	case wkbPoint:
		return readWKBPoint(reader, order)
	case wkbLineString:
		return readWKBLineString(reader, order)
	case wkbPolygon:
		return readWKBPolygon(reader, order)
	default:
		return nil, fmt.Errorf("unsupported WKB geometry type: %d", geomType)
	}
}

func readWKBPoint(r *bytes.Reader, order binary.ByteOrder) (*object.Point, error) {
	var x, y float64
	if err := binary.Read(r, order, &x); err != nil {
		return nil, fmt.Errorf("reading Point X: %w", err)
	}
	if err := binary.Read(r, order, &y); err != nil {
		return nil, fmt.Errorf("reading Point Y: %w", err)
	}
	if math.IsNaN(x) || math.IsNaN(y) {
		return nil, fmt.Errorf("Point contains NaN coordinates")
	}
	return &object.Point{Coord: object.Coordinate{X: x, Y: y}}, nil
}

func readWKBLineString(r *bytes.Reader, order binary.ByteOrder) (*object.LineString, error) {
	var numPoints uint32
	if err := binary.Read(r, order, &numPoints); err != nil {
		return nil, fmt.Errorf("reading LineString numPoints: %w", err)
	}
	coords := make([]object.Coordinate, numPoints)
	for i := uint32(0); i < numPoints; i++ {
		var x, y float64
		if err := binary.Read(r, order, &x); err != nil {
			return nil, fmt.Errorf("reading LineString point %d X: %w", i, err)
		}
		if err := binary.Read(r, order, &y); err != nil {
			return nil, fmt.Errorf("reading LineString point %d Y: %w", i, err)
		}
		coords[i] = object.Coordinate{X: x, Y: y}
	}
	return &object.LineString{Coords: coords}, nil
}

func readWKBPolygon(r *bytes.Reader, order binary.ByteOrder) (*object.Polygon, error) {
	var numRings uint32
	if err := binary.Read(r, order, &numRings); err != nil {
		return nil, fmt.Errorf("reading Polygon numRings: %w", err)
	}
	if numRings < 1 {
		return nil, fmt.Errorf("Polygon must have at least 1 ring")
	}
	poly := &object.Polygon{}
	for i := uint32(0); i < numRings; i++ {
		ring, err := readWKBRing(r, order)
		if err != nil {
			return nil, fmt.Errorf("reading ring %d: %w", i, err)
		}
		if i == 0 {
			poly.ExteriorRing = ring
		} else {
			poly.InteriorRings = append(poly.InteriorRings, ring)
		}
	}
	return poly, nil
}

func readWKBRing(r *bytes.Reader, order binary.ByteOrder) ([]object.Coordinate, error) {
	var numPoints uint32
	if err := binary.Read(r, order, &numPoints); err != nil {
		return nil, fmt.Errorf("reading ring numPoints: %w", err)
	}
	coords := make([]object.Coordinate, numPoints)
	for i := uint32(0); i < numPoints; i++ {
		var x, y float64
		if err := binary.Read(r, order, &x); err != nil {
			return nil, fmt.Errorf("reading ring point %d X: %w", i, err)
		}
		if err := binary.Read(r, order, &y); err != nil {
			return nil, fmt.Errorf("reading ring point %d Y: %w", i, err)
		}
		coords[i] = object.Coordinate{X: x, Y: y}
	}
	return coords, nil
}
