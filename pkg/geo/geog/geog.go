// Package geog provides geography type constructors and geodetic calculations
// on the WGS84 ellipsoid where distances are in meters.
package geog

import (
	"fmt"
	"math"
	"strings"

	"github.com/rogue780/geoflow/internal/object"
)

const earthRadiusM = 6371008.8 // mean Earth radius in meters

func deg2rad(d float64) float64 { return d * math.Pi / 180 }
func rad2deg(r float64) float64 { return r * 180 / math.Pi }

// toFloat extracts a float64 from an Integer or Float object.
func toFloat(o object.Object) (float64, bool) {
	switch v := o.(type) {
	case *object.Float:
		return v.Value, true
	case *object.Integer:
		return float64(v.Value), true
	default:
		return 0, false
	}
}

// GetExports returns all geography module functions.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		// Constructors
		"Point":      &object.Builtin{Name: "geog.Point", Fn: geogPoint},
		"LineString": &object.Builtin{Name: "geog.LineString", Fn: geogLineString},
		"fromLatLon": &object.Builtin{Name: "geog.fromLatLon", Fn: fromLatLon},

		// Geodetic calculations
		"distanceGreatCircle": &object.Builtin{Name: "geog.distanceGreatCircle", Fn: distanceGreatCircle},
		"bearing":             &object.Builtin{Name: "geog.bearing", Fn: bearing},
		"destination":         &object.Builtin{Name: "geog.destination", Fn: destination},
		"midpoint":            &object.Builtin{Name: "geog.midpoint", Fn: midpoint},
		"areaGeodesic":        &object.Builtin{Name: "geog.areaGeodesic", Fn: areaGeodesic},

		// Forward geodesic projection
		"project":     &object.Builtin{Name: "geog.project", Fn: project},
		"interpolate": &object.Builtin{Name: "geog.interpolate", Fn: interpolate},
		"buffer":      &object.Builtin{Name: "geog.buffer", Fn: buffer},

		// Spatial predicates (simplified, bounding-box based)
		"intersects": &object.Builtin{Name: "geog.intersects", Fn: intersects},
		"contains":   &object.Builtin{Name: "geog.contains", Fn: contains},
		"within":     &object.Builtin{Name: "geog.within", Fn: within},

		// WKT / Geometry conversion stubs
		"fromWKT":     &object.Builtin{Name: "geog.fromWKT", Fn: fromWKT},
		"toGeometry":  &object.Builtin{Name: "geog.toGeometry", Fn: toGeometry},
	}
}

// ── Constructors ──

// geogPoint creates a Point with SRID=4326 from (lon, lat).
func geogPoint(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "geog.Point expects 2 arguments (lon, lat)"}
	}
	lon, ok1 := toFloat(args[0])
	lat, ok2 := toFloat(args[1])
	if !ok1 || !ok2 {
		return &object.Error{Message: "geog.Point: arguments must be numbers"}
	}
	return &object.Point{
		Coord: object.Coordinate{X: lon, Y: lat},
		SRID:  4326,
	}
}

// fromLatLon creates a Point with SRID=4326 from (lat, lon) — note the reversed order
// compared to geogPoint which takes (lon, lat). This is a convenience for users
// who think in lat/lon order.
func fromLatLon(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "geog.fromLatLon expects 2 arguments (lat, lon)"}
	}
	lat, ok1 := toFloat(args[0])
	lon, ok2 := toFloat(args[1])
	if !ok1 || !ok2 {
		return &object.Error{Message: "geog.fromLatLon: arguments must be numbers"}
	}
	return &object.Point{
		Coord: object.Coordinate{X: lon, Y: lat},
		SRID:  4326,
	}
}

// geogLineString creates a LineString with SRID=4326 from a list of Points.
func geogLineString(args ...object.Object) object.Object {
	if len(args) < 2 {
		return &object.Error{Message: "geog.LineString expects at least 2 point arguments"}
	}
	coords := make([]object.Coordinate, len(args))
	for i, arg := range args {
		p, ok := arg.(*object.Point)
		if !ok {
			return &object.Error{Message: fmt.Sprintf("geog.LineString: argument %d must be a Point", i)}
		}
		coords[i] = p.Coord
	}
	return &object.LineString{
		Coords: coords,
		SRID:   4326,
	}
}

// ── Geodetic Calculations ──

// haversineDistanceM computes the great-circle distance in meters using the haversine formula.
func haversineDistanceM(c1, c2 object.Coordinate) float64 {
	lat1, lon1 := deg2rad(c1.Y), deg2rad(c1.X)
	lat2, lon2 := deg2rad(c2.Y), deg2rad(c2.X)
	dlat := lat2 - lat1
	dlon := lon2 - lon1
	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusM * c
}

// distanceGreatCircle returns the haversine great-circle distance between two Points in meters.
func distanceGreatCircle(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "geog.distanceGreatCircle expects 2 arguments (point, point)"}
	}
	p1, ok1 := args[0].(*object.Point)
	p2, ok2 := args[1].(*object.Point)
	if !ok1 || !ok2 {
		return &object.Error{Message: "geog.distanceGreatCircle: both arguments must be Points"}
	}
	return &object.Float{Value: haversineDistanceM(p1.Coord, p2.Coord)}
}

// bearing computes the initial bearing (forward azimuth) from point 1 to point 2 in degrees (0-360).
func bearing(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "geog.bearing expects 2 arguments (point, point)"}
	}
	p1, ok1 := args[0].(*object.Point)
	p2, ok2 := args[1].(*object.Point)
	if !ok1 || !ok2 {
		return &object.Error{Message: "geog.bearing: both arguments must be Points"}
	}

	lat1 := deg2rad(p1.Coord.Y)
	lon1 := deg2rad(p1.Coord.X)
	lat2 := deg2rad(p2.Coord.Y)
	lon2 := deg2rad(p2.Coord.X)
	dlon := lon2 - lon1

	x := math.Cos(lat2) * math.Sin(dlon)
	y := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(dlon)
	b := rad2deg(math.Atan2(x, y))
	b = math.Mod(b+360, 360)
	return &object.Float{Value: b}
}

// destination computes the destination point given a start point, bearing in degrees,
// and distance in meters.
func destination(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "geog.destination expects 3 arguments (point, bearingDeg, distanceMeters)"}
	}
	p, ok := args[0].(*object.Point)
	if !ok {
		return &object.Error{Message: "geog.destination: first argument must be a Point"}
	}
	bearingDeg, ok2 := toFloat(args[1])
	distM, ok3 := toFloat(args[2])
	if !ok2 || !ok3 {
		return &object.Error{Message: "geog.destination: bearing and distance must be numbers"}
	}

	lat1 := deg2rad(p.Coord.Y)
	lon1 := deg2rad(p.Coord.X)
	br := deg2rad(bearingDeg)
	angDist := distM / earthRadiusM

	lat2 := math.Asin(math.Sin(lat1)*math.Cos(angDist) +
		math.Cos(lat1)*math.Sin(angDist)*math.Cos(br))
	lon2 := lon1 + math.Atan2(
		math.Sin(br)*math.Sin(angDist)*math.Cos(lat1),
		math.Cos(angDist)-math.Sin(lat1)*math.Sin(lat2))

	return &object.Point{
		Coord: object.Coordinate{X: rad2deg(lon2), Y: rad2deg(lat2)},
		SRID:  4326,
	}
}

// midpoint computes the geodesic midpoint between two Points on the sphere.
func midpoint(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "geog.midpoint expects 2 arguments (point, point)"}
	}
	p1, ok1 := args[0].(*object.Point)
	p2, ok2 := args[1].(*object.Point)
	if !ok1 || !ok2 {
		return &object.Error{Message: "geog.midpoint: both arguments must be Points"}
	}

	lat1 := deg2rad(p1.Coord.Y)
	lon1 := deg2rad(p1.Coord.X)
	lat2 := deg2rad(p2.Coord.Y)
	lon2 := deg2rad(p2.Coord.X)
	dlon := lon2 - lon1

	bx := math.Cos(lat2) * math.Cos(dlon)
	by := math.Cos(lat2) * math.Sin(dlon)

	latM := math.Atan2(
		math.Sin(lat1)+math.Sin(lat2),
		math.Sqrt((math.Cos(lat1)+bx)*(math.Cos(lat1)+bx)+by*by))
	lonM := lon1 + math.Atan2(by, math.Cos(lat1)+bx)

	return &object.Point{
		Coord: object.Coordinate{X: rad2deg(lonM), Y: rad2deg(latM)},
		SRID:  4326,
	}
}

// areaGeodesic computes the geodesic area of a Polygon in square meters
// using the spherical excess formula.
func areaGeodesic(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "geog.areaGeodesic expects 1 argument (polygon)"}
	}
	poly, ok := args[0].(*object.Polygon)
	if !ok {
		return &object.Error{Message: "geog.areaGeodesic: argument must be a Polygon"}
	}

	area := geodesicAreaM2(poly.ExteriorRing)
	for _, ring := range poly.InteriorRings {
		area -= geodesicAreaM2(ring)
	}
	return &object.Float{Value: area}
}

// geodesicAreaM2 computes the area of a coordinate ring in square meters
// using the spherical excess formula.
func geodesicAreaM2(ring []object.Coordinate) float64 {
	n := len(ring)
	if n < 3 {
		return 0
	}
	sum := 0.0
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		k := (i + 2) % n
		sum += (deg2rad(ring[k].X) - deg2rad(ring[i].X)) * math.Sin(deg2rad(ring[j].Y))
	}
	return math.Abs(sum) * earthRadiusM * earthRadiusM / 2
}

// ── Forward Geodesic / Projection ──

// project computes the forward geodesic problem: given a start point, bearing (degrees),
// and distance (meters), returns the destination point. This is an alias for destination.
func project(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "geog.project expects 3 arguments (point, bearingDeg, distanceMeters)"}
	}
	return destination(args...)
}

// interpolate computes an intermediate point along the great circle between two points.
// fraction is in [0, 1] where 0 = p1, 1 = p2.
func interpolate(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "geog.interpolate expects 3 arguments (point1, point2, fraction)"}
	}
	p1, ok1 := args[0].(*object.Point)
	p2, ok2 := args[1].(*object.Point)
	if !ok1 || !ok2 {
		return &object.Error{Message: "geog.interpolate: first two arguments must be Points"}
	}
	frac, ok := toFloat(args[2])
	if !ok {
		return &object.Error{Message: "geog.interpolate: fraction must be a number"}
	}
	if frac < 0 || frac > 1 {
		return &object.Error{Message: "geog.interpolate: fraction must be between 0 and 1"}
	}

	lat1 := deg2rad(p1.Coord.Y)
	lon1 := deg2rad(p1.Coord.X)
	lat2 := deg2rad(p2.Coord.Y)
	lon2 := deg2rad(p2.Coord.X)

	// Angular distance between the two points
	dlat := lat2 - lat1
	dlon := lon2 - lon1
	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlon/2)*math.Sin(dlon/2)
	delta := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	if delta < 1e-12 {
		return &object.Point{
			Coord: object.Coordinate{X: p1.Coord.X, Y: p1.Coord.Y},
			SRID:  4326,
		}
	}

	sinDelta := math.Sin(delta)
	af := math.Sin((1-frac)*delta) / sinDelta
	bf := math.Sin(frac*delta) / sinDelta

	x := af*math.Cos(lat1)*math.Cos(lon1) + bf*math.Cos(lat2)*math.Cos(lon2)
	y := af*math.Cos(lat1)*math.Sin(lon1) + bf*math.Cos(lat2)*math.Sin(lon2)
	z := af*math.Sin(lat1) + bf*math.Sin(lat2)

	latI := math.Atan2(z, math.Sqrt(x*x+y*y))
	lonI := math.Atan2(y, x)

	return &object.Point{
		Coord: object.Coordinate{X: rad2deg(lonI), Y: rad2deg(latI)},
		SRID:  4326,
	}
}

// buffer creates a simplified circular buffer around a point.
// Returns a Polygon approximating a circle with the given radius in meters.
// Uses 32 segments by default.
func buffer(args ...object.Object) object.Object {
	if len(args) < 2 || len(args) > 3 {
		return &object.Error{Message: "geog.buffer expects 2-3 arguments (point, radiusMeters[, segments])"}
	}
	p, ok := args[0].(*object.Point)
	if !ok {
		return &object.Error{Message: "geog.buffer: first argument must be a Point"}
	}
	radiusM, ok2 := toFloat(args[1])
	if !ok2 {
		return &object.Error{Message: "geog.buffer: radius must be a number"}
	}
	segments := 32
	if len(args) == 3 {
		s, ok := args[2].(*object.Integer)
		if !ok || s.Value < 4 {
			return &object.Error{Message: "geog.buffer: segments must be an integer >= 4"}
		}
		segments = int(s.Value)
	}

	coords := make([]object.Coordinate, segments+1)
	for i := 0; i < segments; i++ {
		angle := float64(i) * 360.0 / float64(segments)
		// Use the destination function to compute each point on the circle
		lat1 := deg2rad(p.Coord.Y)
		lon1 := deg2rad(p.Coord.X)
		br := deg2rad(angle)
		angDist := radiusM / earthRadiusM

		lat2 := math.Asin(math.Sin(lat1)*math.Cos(angDist) +
			math.Cos(lat1)*math.Sin(angDist)*math.Cos(br))
		lon2 := lon1 + math.Atan2(
			math.Sin(br)*math.Sin(angDist)*math.Cos(lat1),
			math.Cos(angDist)-math.Sin(lat1)*math.Sin(lat2))

		coords[i] = object.Coordinate{X: rad2deg(lon2), Y: rad2deg(lat2)}
	}
	// Close the ring
	coords[segments] = coords[0]

	return &object.Polygon{
		ExteriorRing: coords,
		SRID:         4326,
	}
}

// ── Spatial Predicates (simplified, bounding-box based) ──

// bboxOf returns the bounding box (minX, minY, maxX, maxY) for a geometry.
func bboxOf(obj object.Object) (float64, float64, float64, float64, bool) {
	switch g := obj.(type) {
	case *object.Point:
		return g.Coord.X, g.Coord.Y, g.Coord.X, g.Coord.Y, true
	case *object.LineString:
		if len(g.Coords) == 0 {
			return 0, 0, 0, 0, false
		}
		minX, minY := g.Coords[0].X, g.Coords[0].Y
		maxX, maxY := minX, minY
		for _, c := range g.Coords[1:] {
			if c.X < minX {
				minX = c.X
			}
			if c.Y < minY {
				minY = c.Y
			}
			if c.X > maxX {
				maxX = c.X
			}
			if c.Y > maxY {
				maxY = c.Y
			}
		}
		return minX, minY, maxX, maxY, true
	case *object.Polygon:
		if len(g.ExteriorRing) == 0 {
			return 0, 0, 0, 0, false
		}
		minX, minY := g.ExteriorRing[0].X, g.ExteriorRing[0].Y
		maxX, maxY := minX, minY
		for _, c := range g.ExteriorRing[1:] {
			if c.X < minX {
				minX = c.X
			}
			if c.Y < minY {
				minY = c.Y
			}
			if c.X > maxX {
				maxX = c.X
			}
			if c.Y > maxY {
				maxY = c.Y
			}
		}
		return minX, minY, maxX, maxY, true
	default:
		return 0, 0, 0, 0, false
	}
}

// bboxIntersects tests if two bounding boxes overlap.
func bboxIntersects(ax1, ay1, ax2, ay2, bx1, by1, bx2, by2 float64) bool {
	return ax1 <= bx2 && ax2 >= bx1 && ay1 <= by2 && ay2 >= by1
}

// bboxContains tests if bbox A fully contains bbox B.
func bboxContains(ax1, ay1, ax2, ay2, bx1, by1, bx2, by2 float64) bool {
	return ax1 <= bx1 && ay1 <= by1 && ax2 >= bx2 && ay2 >= by2
}

// intersects tests whether two geometries' bounding boxes overlap (simplified spatial predicate).
func intersects(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "geog.intersects expects 2 arguments (geometry, geometry)"}
	}
	ax1, ay1, ax2, ay2, ok1 := bboxOf(args[0])
	bx1, by1, bx2, by2, ok2 := bboxOf(args[1])
	if !ok1 || !ok2 {
		return &object.Error{Message: "geog.intersects: both arguments must be geometry types (Point, LineString, Polygon)"}
	}
	return object.NativeBoolToBooleanObject(bboxIntersects(ax1, ay1, ax2, ay2, bx1, by1, bx2, by2))
}

// contains tests whether the first geometry's bounding box fully contains the second's.
func contains(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "geog.contains expects 2 arguments (geometry, geometry)"}
	}
	ax1, ay1, ax2, ay2, ok1 := bboxOf(args[0])
	bx1, by1, bx2, by2, ok2 := bboxOf(args[1])
	if !ok1 || !ok2 {
		return &object.Error{Message: "geog.contains: both arguments must be geometry types (Point, LineString, Polygon)"}
	}
	return object.NativeBoolToBooleanObject(bboxContains(ax1, ay1, ax2, ay2, bx1, by1, bx2, by2))
}

// within tests whether the first geometry's bounding box is fully within the second's.
func within(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "geog.within expects 2 arguments (geometry, geometry)"}
	}
	ax1, ay1, ax2, ay2, ok1 := bboxOf(args[0])
	bx1, by1, bx2, by2, ok2 := bboxOf(args[1])
	if !ok1 || !ok2 {
		return &object.Error{Message: "geog.within: both arguments must be geometry types (Point, LineString, Polygon)"}
	}
	// A is within B means B contains A
	return object.NativeBoolToBooleanObject(bboxContains(bx1, by1, bx2, by2, ax1, ay1, ax2, ay2))
}

// ── WKT / Geometry Conversion Stubs ──

// fromWKT parses a simplified WKT string into a geography Point.
// Only supports POINT geometry for now.
func fromWKT(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "geog.fromWKT expects 1 argument (WKT string)"}
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "geog.fromWKT: argument must be a string"}
	}
	wkt := strings.TrimSpace(s.Value)

	// Simple POINT parser
	if strings.HasPrefix(strings.ToUpper(wkt), "POINT") {
		// Extract coordinates from POINT (x y)
		start := strings.Index(wkt, "(")
		end := strings.LastIndex(wkt, ")")
		if start == -1 || end == -1 || end <= start {
			return &object.Error{Message: "geog.fromWKT: malformed POINT WKT"}
		}
		coordStr := strings.TrimSpace(wkt[start+1 : end])
		parts := strings.Fields(coordStr)
		if len(parts) < 2 {
			return &object.Error{Message: "geog.fromWKT: POINT requires at least 2 coordinates"}
		}
		var x, y float64
		_, err := fmt.Sscanf(parts[0], "%f", &x)
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("geog.fromWKT: invalid x coordinate: %s", parts[0])}
		}
		_, err = fmt.Sscanf(parts[1], "%f", &y)
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("geog.fromWKT: invalid y coordinate: %s", parts[1])}
		}
		return &object.Point{
			Coord: object.Coordinate{X: x, Y: y},
			SRID:  4326,
		}
	}

	return &object.Error{Message: fmt.Sprintf("geog.fromWKT: unsupported geometry type in WKT: %s", wkt)}
}

// toGeometry converts a geography object (SRID=4326 point) to a plain geometry point.
// This is a stub that simply copies the coordinates, clearing the SRID context to 0
// to indicate a "geometry" (planar) context.
func toGeometry(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "geog.toGeometry expects 1 argument (geography point/linestring/polygon)"}
	}
	switch g := args[0].(type) {
	case *object.Point:
		return &object.Point{
			Coord: g.Coord,
			SRID:  0,
		}
	case *object.LineString:
		coords := make([]object.Coordinate, len(g.Coords))
		copy(coords, g.Coords)
		return &object.LineString{
			Coords: coords,
			SRID:   0,
		}
	case *object.Polygon:
		ext := make([]object.Coordinate, len(g.ExteriorRing))
		copy(ext, g.ExteriorRing)
		holes := make([][]object.Coordinate, len(g.InteriorRings))
		for i, ring := range g.InteriorRings {
			holes[i] = make([]object.Coordinate, len(ring))
			copy(holes[i], ring)
		}
		return &object.Polygon{
			ExteriorRing:  ext,
			InteriorRings: holes,
			SRID:          0,
		}
	default:
		return &object.Error{Message: fmt.Sprintf("geog.toGeometry: unsupported type %s", args[0].Type())}
	}
}
