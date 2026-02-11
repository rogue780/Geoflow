// Package geog provides geography type constructors and geodetic calculations
// on the WGS84 ellipsoid where distances are in meters.
package geog

import (
	"fmt"
	"math"

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

		// Geodetic calculations
		"distanceGreatCircle": &object.Builtin{Name: "geog.distanceGreatCircle", Fn: distanceGreatCircle},
		"bearing":             &object.Builtin{Name: "geog.bearing", Fn: bearing},
		"destination":         &object.Builtin{Name: "geog.destination", Fn: destination},
		"midpoint":            &object.Builtin{Name: "geog.midpoint", Fn: midpoint},
		"areaGeodesic":        &object.Builtin{Name: "geog.areaGeodesic", Fn: areaGeodesic},
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
