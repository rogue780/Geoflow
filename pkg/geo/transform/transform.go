// Package transform provides coordinate transformation functions for GeoFlow.
// It supports transforming geometries between coordinate reference systems.
package transform

import (
	"fmt"
	"math"

	"github.com/rogue780/geoflow/internal/object"
)

// GetExports returns all exported functions for the std.geo.transform module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		"transform":       &object.Builtin{Name: "transform.transform", Fn: transformGeom},
		"toWebMercator":   &object.Builtin{Name: "transform.toWebMercator", Fn: toWebMercator},
		"fromWebMercator": &object.Builtin{Name: "transform.fromWebMercator", Fn: fromWebMercator},
		"toUTM":           &object.Builtin{Name: "transform.toUTM", Fn: toUTMFunc},
		"transformPoint":  &object.Builtin{Name: "transform.transformPoint", Fn: transformPoint},
	}
}

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

// Semi-major axis of WGS84 ellipsoid in metres
const a = 6378137.0

// lonLatToWebMercator converts a WGS84 (lon, lat) coordinate to Web Mercator (x, y).
func lonLatToWebMercator(lon, lat float64) (float64, float64) {
	x := a * lon * math.Pi / 180
	y := a * math.Log(math.Tan(math.Pi/4+lat*math.Pi/360))
	return x, y
}

// webMercatorToLonLat converts Web Mercator (x, y) to WGS84 (lon, lat).
func webMercatorToLonLat(x, y float64) (float64, float64) {
	lon := x / a * 180 / math.Pi
	lat := (2*math.Atan(math.Exp(y/a)) - math.Pi/2) * 180 / math.Pi
	return lon, lat
}

// transformCoord transforms a single coordinate between two CRS.
// Currently supports: WGS84 <-> WebMercator conversions.
// For same-datum geographic CRS (NAD83, ETRS89), coordinates pass through since
// they share the same (or very close) ellipsoid. For unsupported transforms, an
// error string is returned.
func transformCoord(c object.Coordinate, from, to *object.CRS) (object.Coordinate, string) {
	if from.EPSGCode == to.EPSGCode {
		return c, ""
	}

	// Normalize to WGS84 first if needed
	lon, lat := c.X, c.Y
	switch from.EPSGCode {
	case 4326, 4269, 4258: // Geographic CRS on same/similar ellipsoid
		// lon, lat already in degrees
	case 3857:
		lon, lat = webMercatorToLonLat(c.X, c.Y)
	default:
		if from.EPSGCode >= 32601 && from.EPSGCode <= 32660 || from.EPSGCode >= 32701 && from.EPSGCode <= 32760 {
			return c, fmt.Sprintf("transform from UTM (EPSG:%d) not yet supported", from.EPSGCode)
		}
		return c, fmt.Sprintf("transform from EPSG:%d not supported", from.EPSGCode)
	}

	// Then convert from WGS84 to target
	switch to.EPSGCode {
	case 4326, 4269, 4258:
		return object.Coordinate{X: lon, Y: lat, Z: c.Z, HasZ: c.HasZ}, ""
	case 3857:
		x, y := lonLatToWebMercator(lon, lat)
		return object.Coordinate{X: x, Y: y, Z: c.Z, HasZ: c.HasZ}, ""
	default:
		return c, fmt.Sprintf("transform to EPSG:%d not supported", to.EPSGCode)
	}
}

// transformGeomCoords transforms all coordinates of a geometry from one CRS to another.
func transformGeomCoords(geom object.Object, from, to *object.CRS) object.Object {
	switch g := geom.(type) {
	case *object.Point:
		nc, errMsg := transformCoord(g.Coord, from, to)
		if errMsg != "" {
			return &object.Error{Message: errMsg}
		}
		return &object.Point{Coord: nc, SRID: to.EPSGCode}

	case *object.LineString:
		coords := make([]object.Coordinate, len(g.Coords))
		for i, c := range g.Coords {
			nc, errMsg := transformCoord(c, from, to)
			if errMsg != "" {
				return &object.Error{Message: errMsg}
			}
			coords[i] = nc
		}
		return &object.LineString{Coords: coords, SRID: to.EPSGCode}

	case *object.Polygon:
		ext := make([]object.Coordinate, len(g.ExteriorRing))
		for i, c := range g.ExteriorRing {
			nc, errMsg := transformCoord(c, from, to)
			if errMsg != "" {
				return &object.Error{Message: errMsg}
			}
			ext[i] = nc
		}
		holes := make([][]object.Coordinate, len(g.InteriorRings))
		for r, ring := range g.InteriorRings {
			holes[r] = make([]object.Coordinate, len(ring))
			for i, c := range ring {
				nc, errMsg := transformCoord(c, from, to)
				if errMsg != "" {
					return &object.Error{Message: errMsg}
				}
				holes[r][i] = nc
			}
		}
		return &object.Polygon{ExteriorRing: ext, InteriorRings: holes, SRID: to.EPSGCode}

	default:
		return &object.Error{Message: fmt.Sprintf("transform: unsupported geometry type %s", geom.Type())}
	}
}

// transformGeom(geometry, fromCRS, toCRS) -> geometry
func transformGeom(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "transform.transform expects 3 arguments (geometry, fromCRS, toCRS)"}
	}
	from, ok := args[1].(*object.CRS)
	if !ok {
		return &object.Error{Message: "transform.transform: second argument must be a CRS"}
	}
	to, ok := args[2].(*object.CRS)
	if !ok {
		return &object.Error{Message: "transform.transform: third argument must be a CRS"}
	}
	return transformGeomCoords(args[0], from, to)
}

// toWebMercator(geometry) -> geometry
// Assumes input is WGS84 (EPSG:4326) by default.
func toWebMercator(args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return &object.Error{Message: "transform.toWebMercator expects 1-2 arguments (geometry[, fromCRS])"}
	}
	from := &object.CRS{EPSGCode: 4326, CRSName: "WGS 84", IsGeographic: true, Units: "degree"}
	if len(args) == 2 {
		c, ok := args[1].(*object.CRS)
		if !ok {
			return &object.Error{Message: "transform.toWebMercator: second argument must be a CRS"}
		}
		from = c
	}
	to := &object.CRS{EPSGCode: 3857, CRSName: "WGS 84 / Pseudo-Mercator", IsProjected: true, Units: "metre"}
	return transformGeomCoords(args[0], from, to)
}

// fromWebMercator(geometry) -> geometry
// Converts from Web Mercator to WGS84 by default.
func fromWebMercator(args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return &object.Error{Message: "transform.fromWebMercator expects 1-2 arguments (geometry[, toCRS])"}
	}
	from := &object.CRS{EPSGCode: 3857, CRSName: "WGS 84 / Pseudo-Mercator", IsProjected: true, Units: "metre"}
	to := &object.CRS{EPSGCode: 4326, CRSName: "WGS 84", IsGeographic: true, Units: "degree"}
	if len(args) == 2 {
		c, ok := args[1].(*object.CRS)
		if !ok {
			return &object.Error{Message: "transform.fromWebMercator: second argument must be a CRS"}
		}
		to = c
	}
	return transformGeomCoords(args[0], from, to)
}

// toUTMFunc(geometry) -> (geometry, CRS)
// Auto-detects UTM zone from the geometry centroid. Assumes WGS84 input.
func toUTMFunc(args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return &object.Error{Message: "transform.toUTM expects 1-2 arguments (geometry[, fromCRS])"}
	}
	// Determine centroid to detect UTM zone
	var lon, lat float64
	switch g := args[0].(type) {
	case *object.Point:
		lon, lat = g.Coord.X, g.Coord.Y
	case *object.LineString:
		if len(g.Coords) == 0 {
			return &object.Error{Message: "transform.toUTM: empty LineString"}
		}
		for _, c := range g.Coords {
			lon += c.X
			lat += c.Y
		}
		lon /= float64(len(g.Coords))
		lat /= float64(len(g.Coords))
	case *object.Polygon:
		if len(g.ExteriorRing) == 0 {
			return &object.Error{Message: "transform.toUTM: empty Polygon"}
		}
		for _, c := range g.ExteriorRing {
			lon += c.X
			lat += c.Y
		}
		lon /= float64(len(g.ExteriorRing))
		lat /= float64(len(g.ExteriorRing))
	default:
		return &object.Error{Message: fmt.Sprintf("transform.toUTM: unsupported geometry type %s", args[0].Type())}
	}

	zone := int(math.Floor((lon+180)/6)) + 1
	if zone > 60 {
		zone = 60
	}
	north := lat >= 0

	var epsg int
	var hemisphere string
	if north {
		epsg = 32600 + zone
		hemisphere = "N"
	} else {
		epsg = 32700 + zone
		hemisphere = "S"
	}

	utmCRS := &object.CRS{
		EPSGCode:     epsg,
		CRSName:      fmt.Sprintf("WGS 84 / UTM zone %d%s", zone, hemisphere),
		IsGeographic: false,
		IsProjected:  true,
		Units:        "metre",
	}

	// For now, return a tuple of (geometry, CRS) where the geometry keeps its coords
	// (actual UTM projection would need full Transverse Mercator equations)
	return &object.Tuple{Elements: []object.Object{args[0], utmCRS}}
}

// transformPoint(x, y, fromCRS, toCRS) -> Tuple<float, float>
func transformPoint(args ...object.Object) object.Object {
	if len(args) != 4 {
		return &object.Error{Message: "transform.transformPoint expects 4 arguments (x, y, fromCRS, toCRS)"}
	}
	x, ok1 := toFloat(args[0])
	y, ok2 := toFloat(args[1])
	if !ok1 || !ok2 {
		return &object.Error{Message: "transform.transformPoint: x and y must be numbers"}
	}
	from, ok := args[2].(*object.CRS)
	if !ok {
		return &object.Error{Message: "transform.transformPoint: third argument must be a CRS"}
	}
	to, ok := args[3].(*object.CRS)
	if !ok {
		return &object.Error{Message: "transform.transformPoint: fourth argument must be a CRS"}
	}
	coord := object.Coordinate{X: x, Y: y}
	nc, errMsg := transformCoord(coord, from, to)
	if errMsg != "" {
		return &object.Error{Message: errMsg}
	}
	return &object.Tuple{Elements: []object.Object{
		&object.Float{Value: nc.X},
		&object.Float{Value: nc.Y},
	}}
}
