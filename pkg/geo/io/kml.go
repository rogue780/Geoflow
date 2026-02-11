package io

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/rogue780/geoflow/internal/object"
)

// GetKMLExports returns all KML I/O functions.
func GetKMLExports() map[string]object.Object {
	return map[string]object.Object{
		"toKML":   &object.Builtin{Name: "geoio.toKML", Fn: toKML},
		"fromKML": &object.Builtin{Name: "geoio.fromKML", Fn: fromKML},
	}
}

// ── KML Output ──

// toKML converts a geometry to a KML XML string.
func toKML(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "geoio.toKML expects 1 argument (geometry)"}
	}
	geom, ok := args[0].(object.Geometry)
	if !ok {
		return &object.Error{Message: "geoio.toKML: argument must be a geometry"}
	}
	result := geometryToKML(geom)
	return &object.String{Value: result}
}

func geometryToKML(geom object.Geometry) string {
	switch g := geom.(type) {
	case *object.Point:
		return fmt.Sprintf("<Point><coordinates>%s</coordinates></Point>",
			coordToKML(g.Coord))
	case *object.LineString:
		return fmt.Sprintf("<LineString><coordinates>%s</coordinates></LineString>",
			coordsToKMLString(g.Coords))
	case *object.Polygon:
		var sb strings.Builder
		sb.WriteString("<Polygon>")
		sb.WriteString("<outerBoundaryIs><LinearRing><coordinates>")
		sb.WriteString(coordsToKMLString(g.ExteriorRing))
		sb.WriteString("</coordinates></LinearRing></outerBoundaryIs>")
		for _, ring := range g.InteriorRings {
			sb.WriteString("<innerBoundaryIs><LinearRing><coordinates>")
			sb.WriteString(coordsToKMLString(ring))
			sb.WriteString("</coordinates></LinearRing></innerBoundaryIs>")
		}
		sb.WriteString("</Polygon>")
		return sb.String()
	case *object.MultiPoint:
		var sb strings.Builder
		sb.WriteString("<MultiGeometry>")
		for _, p := range g.Points {
			sb.WriteString(geometryToKML(p))
		}
		sb.WriteString("</MultiGeometry>")
		return sb.String()
	case *object.MultiLineString:
		var sb strings.Builder
		sb.WriteString("<MultiGeometry>")
		for _, l := range g.Lines {
			sb.WriteString(geometryToKML(l))
		}
		sb.WriteString("</MultiGeometry>")
		return sb.String()
	case *object.MultiPolygon:
		var sb strings.Builder
		sb.WriteString("<MultiGeometry>")
		for _, p := range g.Polygons {
			sb.WriteString(geometryToKML(p))
		}
		sb.WriteString("</MultiGeometry>")
		return sb.String()
	case *object.GeometryCollection:
		var sb strings.Builder
		sb.WriteString("<MultiGeometry>")
		for _, child := range g.Geometries {
			sb.WriteString(geometryToKML(child))
		}
		sb.WriteString("</MultiGeometry>")
		return sb.String()
	default:
		return ""
	}
}

func coordToKML(c object.Coordinate) string {
	if c.HasZ {
		return fmt.Sprintf("%g,%g,%g", c.X, c.Y, c.Z)
	}
	return fmt.Sprintf("%g,%g,0", c.X, c.Y)
}

func coordsToKMLString(coords []object.Coordinate) string {
	parts := make([]string, len(coords))
	for i, c := range coords {
		parts[i] = coordToKML(c)
	}
	return strings.Join(parts, " ")
}

// ── KML Input ──

// kmlPoint, kmlLineString, kmlPolygon, kmlMultiGeometry are XML structures for parsing.
type kmlPoint struct {
	XMLName     xml.Name `xml:"Point"`
	Coordinates string   `xml:"coordinates"`
}

type kmlLineString struct {
	XMLName     xml.Name `xml:"LineString"`
	Coordinates string   `xml:"coordinates"`
}

type kmlLinearRing struct {
	XMLName     xml.Name `xml:"LinearRing"`
	Coordinates string   `xml:"coordinates"`
}

type kmlOuterBoundary struct {
	XMLName    xml.Name      `xml:"outerBoundaryIs"`
	LinearRing kmlLinearRing `xml:"LinearRing"`
}

type kmlInnerBoundary struct {
	XMLName    xml.Name      `xml:"innerBoundaryIs"`
	LinearRing kmlLinearRing `xml:"LinearRing"`
}

type kmlPolygon struct {
	XMLName        xml.Name           `xml:"Polygon"`
	OuterBoundary  kmlOuterBoundary   `xml:"outerBoundaryIs"`
	InnerBoundaries []kmlInnerBoundary `xml:"innerBoundaryIs"`
}

type kmlMultiGeometry struct {
	XMLName xml.Name `xml:"MultiGeometry"`
	Inner   []byte   `xml:",innerxml"`
}

// fromKML parses a KML XML string and returns a geometry object.
func fromKML(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "geoio.fromKML expects 1 argument (string)"}
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "geoio.fromKML: argument must be a string"}
	}

	geom, err := parseKMLGeometry(strings.TrimSpace(s.Value))
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("KML parse error: %s", err)}
	}
	return geom
}

func parseKMLGeometry(kmlStr string) (object.Geometry, error) {
	trimmed := strings.TrimSpace(kmlStr)

	// Try Point
	if strings.HasPrefix(trimmed, "<Point") {
		var pt kmlPoint
		if err := xml.Unmarshal([]byte(trimmed), &pt); err != nil {
			return nil, fmt.Errorf("parsing Point: %w", err)
		}
		coords, err := parseKMLCoordinates(pt.Coordinates)
		if err != nil {
			return nil, err
		}
		if len(coords) < 1 {
			return nil, fmt.Errorf("Point requires at least 1 coordinate")
		}
		return &object.Point{Coord: coords[0]}, nil
	}

	// Try LineString
	if strings.HasPrefix(trimmed, "<LineString") {
		var ls kmlLineString
		if err := xml.Unmarshal([]byte(trimmed), &ls); err != nil {
			return nil, fmt.Errorf("parsing LineString: %w", err)
		}
		coords, err := parseKMLCoordinates(ls.Coordinates)
		if err != nil {
			return nil, err
		}
		if len(coords) < 2 {
			return nil, fmt.Errorf("LineString requires at least 2 coordinates")
		}
		return &object.LineString{Coords: coords}, nil
	}

	// Try Polygon
	if strings.HasPrefix(trimmed, "<Polygon") {
		var poly kmlPolygon
		if err := xml.Unmarshal([]byte(trimmed), &poly); err != nil {
			return nil, fmt.Errorf("parsing Polygon: %w", err)
		}
		extCoords, err := parseKMLCoordinates(poly.OuterBoundary.LinearRing.Coordinates)
		if err != nil {
			return nil, fmt.Errorf("parsing outer boundary: %w", err)
		}
		result := &object.Polygon{ExteriorRing: extCoords}
		for _, inner := range poly.InnerBoundaries {
			intCoords, err := parseKMLCoordinates(inner.LinearRing.Coordinates)
			if err != nil {
				return nil, fmt.Errorf("parsing inner boundary: %w", err)
			}
			result.InteriorRings = append(result.InteriorRings, intCoords)
		}
		return result, nil
	}

	// Try MultiGeometry
	if strings.HasPrefix(trimmed, "<MultiGeometry") {
		var mg kmlMultiGeometry
		if err := xml.Unmarshal([]byte(trimmed), &mg); err != nil {
			return nil, fmt.Errorf("parsing MultiGeometry: %w", err)
		}
		return parseMultiGeometryInner(string(mg.Inner))
	}

	return nil, fmt.Errorf("unrecognized KML geometry element")
}

// parseMultiGeometryInner parses the inner XML of a MultiGeometry element.
func parseMultiGeometryInner(inner string) (object.Geometry, error) {
	// Use a decoder to extract child elements
	decoder := xml.NewDecoder(strings.NewReader(inner))
	var geometries []object.Geometry

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		// Read the full element by re-encoding start + content
		var rawContent strings.Builder
		rawContent.WriteString("<")
		rawContent.WriteString(se.Name.Local)
		for _, attr := range se.Attr {
			rawContent.WriteString(fmt.Sprintf(` %s="%s"`, attr.Name.Local, attr.Value))
		}
		rawContent.WriteString(">")
		// Read until the matching end element
		depth := 1
		for depth > 0 {
			innerTok, err := decoder.Token()
			if err != nil {
				return nil, fmt.Errorf("unexpected end of MultiGeometry content")
			}
			switch t := innerTok.(type) {
			case xml.StartElement:
				depth++
				rawContent.WriteString("<")
				rawContent.WriteString(t.Name.Local)
				for _, attr := range t.Attr {
					rawContent.WriteString(fmt.Sprintf(` %s="%s"`, attr.Name.Local, attr.Value))
				}
				rawContent.WriteString(">")
			case xml.EndElement:
				depth--
				rawContent.WriteString("</")
				rawContent.WriteString(t.Name.Local)
				rawContent.WriteString(">")
			case xml.CharData:
				rawContent.Write(t)
			}
		}

		geom, err := parseKMLGeometry(rawContent.String())
		if err != nil {
			return nil, fmt.Errorf("parsing child of MultiGeometry: %w", err)
		}
		geometries = append(geometries, geom)
	}

	if len(geometries) == 0 {
		return &object.GeometryCollection{}, nil
	}
	return &object.GeometryCollection{Geometries: geometries}, nil
}

// parseKMLCoordinates parses a KML coordinates string of the form "lon,lat[,alt] lon,lat[,alt] ...".
func parseKMLCoordinates(s string) ([]object.Coordinate, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	// Split by whitespace
	tuples := strings.Fields(s)
	coords := make([]object.Coordinate, 0, len(tuples))

	for _, tuple := range tuples {
		parts := strings.Split(tuple, ",")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid KML coordinate tuple: %q", tuple)
		}
		lon, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid longitude in KML coordinate: %w", err)
		}
		lat, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid latitude in KML coordinate: %w", err)
		}
		c := object.Coordinate{X: lon, Y: lat}
		if len(parts) >= 3 {
			alt, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
			if err == nil && alt != 0 {
				c.Z = alt
				c.HasZ = true
			}
		}
		coords = append(coords, c)
	}
	return coords, nil
}
