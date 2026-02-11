package object

import (
	"fmt"
	"strings"
)

// Geometry type constants.
const (
	POINT_OBJ               Type = "Point"
	LINESTRING_OBJ          Type = "LineString"
	POLYGON_OBJ             Type = "Polygon"
	MULTIPOINT_OBJ          Type = "MultiPoint"
	MULTILINESTRING_OBJ     Type = "MultiLineString"
	MULTIPOLYGON_OBJ        Type = "MultiPolygon"
	GEOMETRYCOLLECTION_OBJ  Type = "GeometryCollection"
	FEATURE_OBJ             Type = "Feature"
	FEATURECOLLECTION_OBJ   Type = "FeatureCollection"
	BBOX_OBJ                Type = "BBox"
)

// Coordinate represents a geographic coordinate (x/lon, y/lat, optional z/elevation).
type Coordinate struct {
	X float64
	Y float64
	Z float64
	HasZ bool
}

func (c Coordinate) String() string {
	if c.HasZ {
		return fmt.Sprintf("%g %g %g", c.X, c.Y, c.Z)
	}
	return fmt.Sprintf("%g %g", c.X, c.Y)
}

func (c Coordinate) Equals(other Coordinate) bool {
	if c.X != other.X || c.Y != other.Y {
		return false
	}
	if c.HasZ && other.HasZ {
		return c.Z == other.Z
	}
	return true
}

// Geometry is the interface all geometry types implement.
type Geometry interface {
	Object
	GeomType() string
	Coordinates() []Coordinate
	ToWKT() string
}

// Point represents a single coordinate position.
type Point struct {
	Coord Coordinate
	SRID  int
}

func (p *Point) Type() Type      { return POINT_OBJ }
func (p *Point) GeomType() string { return "Point" }
func (p *Point) Inspect() string { return p.ToWKT() }
func (p *Point) ToWKT() string {
	if p.Coord.HasZ {
		return fmt.Sprintf("POINT Z (%g %g %g)", p.Coord.X, p.Coord.Y, p.Coord.Z)
	}
	return fmt.Sprintf("POINT (%g %g)", p.Coord.X, p.Coord.Y)
}
func (p *Point) Coordinates() []Coordinate { return []Coordinate{p.Coord} }

// LineString represents an ordered sequence of coordinates forming a line.
type LineString struct {
	Coords []Coordinate
	SRID   int
}

func (l *LineString) Type() Type      { return LINESTRING_OBJ }
func (l *LineString) GeomType() string { return "LineString" }
func (l *LineString) Inspect() string { return l.ToWKT() }
func (l *LineString) ToWKT() string {
	parts := make([]string, len(l.Coords))
	for i, c := range l.Coords {
		parts[i] = c.String()
	}
	return fmt.Sprintf("LINESTRING (%s)", strings.Join(parts, ", "))
}
func (l *LineString) Coordinates() []Coordinate { return l.Coords }

// Polygon represents an area defined by an exterior ring and optional interior rings (holes).
type Polygon struct {
	ExteriorRing []Coordinate
	InteriorRings [][]Coordinate
	SRID         int
}

func (p *Polygon) Type() Type      { return POLYGON_OBJ }
func (p *Polygon) GeomType() string { return "Polygon" }
func (p *Polygon) Inspect() string { return p.ToWKT() }
func (p *Polygon) ToWKT() string {
	rings := make([]string, 0, 1+len(p.InteriorRings))
	rings = append(rings, ringToWKT(p.ExteriorRing))
	for _, ring := range p.InteriorRings {
		rings = append(rings, ringToWKT(ring))
	}
	return fmt.Sprintf("POLYGON (%s)", strings.Join(rings, ", "))
}
func (p *Polygon) Coordinates() []Coordinate {
	all := make([]Coordinate, 0, len(p.ExteriorRing))
	all = append(all, p.ExteriorRing...)
	for _, ring := range p.InteriorRings {
		all = append(all, ring...)
	}
	return all
}

func ringToWKT(coords []Coordinate) string {
	parts := make([]string, len(coords))
	for i, c := range coords {
		parts[i] = c.String()
	}
	return fmt.Sprintf("(%s)", strings.Join(parts, ", "))
}

// MultiPoint represents a collection of points.
type MultiPoint struct {
	Points []*Point
	SRID   int
}

func (mp *MultiPoint) Type() Type      { return MULTIPOINT_OBJ }
func (mp *MultiPoint) GeomType() string { return "MultiPoint" }
func (mp *MultiPoint) Inspect() string { return mp.ToWKT() }
func (mp *MultiPoint) ToWKT() string {
	parts := make([]string, len(mp.Points))
	for i, p := range mp.Points {
		parts[i] = fmt.Sprintf("(%s)", p.Coord.String())
	}
	return fmt.Sprintf("MULTIPOINT (%s)", strings.Join(parts, ", "))
}
func (mp *MultiPoint) Coordinates() []Coordinate {
	coords := make([]Coordinate, len(mp.Points))
	for i, p := range mp.Points {
		coords[i] = p.Coord
	}
	return coords
}

// MultiLineString represents a collection of line strings.
type MultiLineString struct {
	Lines []*LineString
	SRID  int
}

func (ml *MultiLineString) Type() Type      { return MULTILINESTRING_OBJ }
func (ml *MultiLineString) GeomType() string { return "MultiLineString" }
func (ml *MultiLineString) Inspect() string { return ml.ToWKT() }
func (ml *MultiLineString) ToWKT() string {
	parts := make([]string, len(ml.Lines))
	for i, l := range ml.Lines {
		coords := make([]string, len(l.Coords))
		for j, c := range l.Coords {
			coords[j] = c.String()
		}
		parts[i] = fmt.Sprintf("(%s)", strings.Join(coords, ", "))
	}
	return fmt.Sprintf("MULTILINESTRING (%s)", strings.Join(parts, ", "))
}
func (ml *MultiLineString) Coordinates() []Coordinate {
	var all []Coordinate
	for _, l := range ml.Lines {
		all = append(all, l.Coords...)
	}
	return all
}

// MultiPolygon represents a collection of polygons.
type MultiPolygon struct {
	Polygons []*Polygon
	SRID     int
}

func (mp *MultiPolygon) Type() Type      { return MULTIPOLYGON_OBJ }
func (mp *MultiPolygon) GeomType() string { return "MultiPolygon" }
func (mp *MultiPolygon) Inspect() string { return mp.ToWKT() }
func (mp *MultiPolygon) ToWKT() string {
	parts := make([]string, len(mp.Polygons))
	for i, p := range mp.Polygons {
		rings := make([]string, 0, 1+len(p.InteriorRings))
		rings = append(rings, ringToWKT(p.ExteriorRing))
		for _, ring := range p.InteriorRings {
			rings = append(rings, ringToWKT(ring))
		}
		parts[i] = fmt.Sprintf("(%s)", strings.Join(rings, ", "))
	}
	return fmt.Sprintf("MULTIPOLYGON (%s)", strings.Join(parts, ", "))
}
func (mp *MultiPolygon) Coordinates() []Coordinate {
	var all []Coordinate
	for _, p := range mp.Polygons {
		all = append(all, p.Coordinates()...)
	}
	return all
}

// GeometryCollection represents a heterogeneous collection of geometries.
type GeometryCollection struct {
	Geometries []Geometry
	SRID       int
}

func (gc *GeometryCollection) Type() Type      { return GEOMETRYCOLLECTION_OBJ }
func (gc *GeometryCollection) GeomType() string { return "GeometryCollection" }
func (gc *GeometryCollection) Inspect() string { return gc.ToWKT() }
func (gc *GeometryCollection) ToWKT() string {
	parts := make([]string, len(gc.Geometries))
	for i, g := range gc.Geometries {
		parts[i] = g.ToWKT()
	}
	return fmt.Sprintf("GEOMETRYCOLLECTION (%s)", strings.Join(parts, ", "))
}
func (gc *GeometryCollection) Coordinates() []Coordinate {
	var all []Coordinate
	for _, g := range gc.Geometries {
		all = append(all, g.Coordinates()...)
	}
	return all
}

// BBox represents a bounding box (minX, minY, maxX, maxY).
type BBox struct {
	MinX float64
	MinY float64
	MaxX float64
	MaxY float64
}

func (bb *BBox) Type() Type      { return BBOX_OBJ }
func (bb *BBox) Inspect() string {
	return fmt.Sprintf("BBox(%g, %g, %g, %g)", bb.MinX, bb.MinY, bb.MaxX, bb.MaxY)
}

// Feature represents a geometry with associated properties.
type Feature struct {
	Geom       Geometry
	Properties *Map
	ID         Object
}

func (f *Feature) Type() Type      { return FEATURE_OBJ }
func (f *Feature) Inspect() string {
	id := "nil"
	if f.ID != nil {
		id = f.ID.Inspect()
	}
	return fmt.Sprintf("Feature(id=%s, geometry=%s, properties=%s)", id, f.Geom.Inspect(), f.Properties.Inspect())
}

// FeatureCollection represents an ordered collection of Features.
type FeatureCollection struct {
	Features []*Feature
}

func (fc *FeatureCollection) Type() Type      { return FEATURECOLLECTION_OBJ }
func (fc *FeatureCollection) Inspect() string {
	return fmt.Sprintf("FeatureCollection(%d features)", len(fc.Features))
}
