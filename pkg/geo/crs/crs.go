// Package crs provides coordinate reference system constants and functions for GeoFlow.
package crs

import (
	"fmt"

	"github.com/rogue780/geoflow/internal/object"
)

// Well-known CRS definitions.
var (
	wgs84 = &object.CRS{
		EPSGCode:     4326,
		CRSName:      "WGS 84",
		IsGeographic: true,
		IsProjected:  false,
		Units:        "degree",
	}
	webMercator = &object.CRS{
		EPSGCode:     3857,
		CRSName:      "WGS 84 / Pseudo-Mercator",
		IsGeographic: false,
		IsProjected:  true,
		Units:        "metre",
	}
	nad83 = &object.CRS{
		EPSGCode:     4269,
		CRSName:      "NAD83",
		IsGeographic: true,
		IsProjected:  false,
		Units:        "degree",
	}
	nad27 = &object.CRS{
		EPSGCode:     4267,
		CRSName:      "NAD27",
		IsGeographic: true,
		IsProjected:  false,
		Units:        "degree",
	}
	etrs89 = &object.CRS{
		EPSGCode:     4258,
		CRSName:      "ETRS89",
		IsGeographic: true,
		IsProjected:  false,
		Units:        "degree",
	}
)

// knownCRS maps EPSG codes to predefined CRS objects.
var knownCRS = map[int]*object.CRS{
	4326: wgs84,
	3857: webMercator,
	4269: nad83,
	4267: nad27,
	4258: etrs89,
}

// crsMetadata stores additional metadata for well-known CRS definitions.
type crsMetadata struct {
	Datum     string
	Ellipsoid string
	Proj4     string
	WKTDef    string
	BoundsW   float64
	BoundsS   float64
	BoundsE   float64
	BoundsN   float64
}

var knownMetadata = map[int]*crsMetadata{
	4326: {
		Datum:     "World Geodetic System 1984",
		Ellipsoid: "WGS 84",
		Proj4:     "+proj=longlat +datum=WGS84 +no_defs +type=crs",
		WKTDef:    `GEOGCRS["WGS 84",DATUM["World Geodetic System 1984",ELLIPSOID["WGS 84",6378137,298.257223563]],CS[ellipsoidal,2],AXIS["latitude",north],AXIS["longitude",east],UNIT["degree",0.0174532925199433]]`,
		BoundsW:   -180, BoundsS: -90, BoundsE: 180, BoundsN: 90,
	},
	3857: {
		Datum:     "World Geodetic System 1984",
		Ellipsoid: "WGS 84",
		Proj4:     "+proj=merc +a=6378137 +b=6378137 +lat_ts=0 +lon_0=0 +x_0=0 +y_0=0 +k=1 +units=m +no_defs +type=crs",
		WKTDef:    `PROJCRS["WGS 84 / Pseudo-Mercator",BASEGEOGCRS["WGS 84",DATUM["World Geodetic System 1984",ELLIPSOID["WGS 84",6378137,298.257223563]]],CONVERSION["Popular Visualisation Pseudo-Mercator"],CS[Cartesian,2],AXIS["easting",east],AXIS["northing",north],UNIT["metre",1]]`,
		BoundsW:   -20037508.34, BoundsS: -20048966.1, BoundsE: 20037508.34, BoundsN: 20048966.1,
	},
	4269: {
		Datum:     "North American Datum 1983",
		Ellipsoid: "GRS 1980",
		Proj4:     "+proj=longlat +datum=NAD83 +no_defs +type=crs",
		WKTDef:    `GEOGCRS["NAD83",DATUM["North American Datum 1983",ELLIPSOID["GRS 1980",6378137,298.257222101]],CS[ellipsoidal,2],AXIS["latitude",north],AXIS["longitude",east],UNIT["degree",0.0174532925199433]]`,
		BoundsW:   -172.54, BoundsS: 23.81, BoundsE: -47.74, BoundsN: 86.46,
	},
	4267: {
		Datum:     "North American Datum 1927",
		Ellipsoid: "Clarke 1866",
		Proj4:     "+proj=longlat +datum=NAD27 +no_defs +type=crs",
		WKTDef:    `GEOGCRS["NAD27",DATUM["North American Datum 1927",ELLIPSOID["Clarke 1866",6378206.4,294.978698214]],CS[ellipsoidal,2],AXIS["latitude",north],AXIS["longitude",east],UNIT["degree",0.0174532925199433]]`,
		BoundsW:   -172.54, BoundsS: 23.81, BoundsE: -47.74, BoundsN: 86.46,
	},
	4258: {
		Datum:     "European Terrestrial Reference System 1989",
		Ellipsoid: "GRS 1980",
		Proj4:     "+proj=longlat +ellps=GRS80 +no_defs +type=crs",
		WKTDef:    `GEOGCRS["ETRS89",DATUM["European Terrestrial Reference System 1989",ELLIPSOID["GRS 1980",6378137,298.257222101]],CS[ellipsoidal,2],AXIS["latitude",north],AXIS["longitude",east],UNIT["degree",0.0174532925199433]]`,
		BoundsW:   -16.1, BoundsS: 32.88, BoundsE: 40.18, BoundsN: 84.17,
	},
}

// GetExports returns all exported constants and functions for the crs module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		"WGS84":       wgs84,
		"WebMercator": webMercator,
		"NAD83":       nad83,
		"NAD27":       nad27,
		"ETRS89":      etrs89,
		"fromEPSG":    &object.Builtin{Name: "crs.fromEPSG", Fn: fromEPSG},
		"UTM":         &object.Builtin{Name: "crs.UTM", Fn: utm},
		"toProj4":     &object.Builtin{Name: "crs.toProj4", Fn: toProj4},
		"toWKT":       &object.Builtin{Name: "crs.toWKT", Fn: toWKT},
		"bounds":      &object.Builtin{Name: "crs.bounds", Fn: bounds},
		"datum":       &object.Builtin{Name: "crs.datum", Fn: datum},
		"ellipsoid":   &object.Builtin{Name: "crs.ellipsoid", Fn: ellipsoid},
	}
}

func fromEPSG(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "crs.fromEPSG expects 1 argument (EPSG code)"}
	}
	code, ok := args[0].(*object.Integer)
	if !ok {
		return &object.Error{Message: "crs.fromEPSG: argument must be an integer"}
	}
	if c, found := knownCRS[int(code.Value)]; found {
		return &object.Result{Value: c, IsOk: true}
	}
	return &object.Result{
		Value: &object.String{Value: fmt.Sprintf("unknown EPSG code: %d", code.Value)},
		IsOk:  false,
	}
}

func utm(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "crs.UTM expects 2 arguments (zone, north)"}
	}
	zone, ok := args[0].(*object.Integer)
	if !ok {
		return &object.Error{Message: "crs.UTM: zone must be an integer"}
	}
	northObj, ok := args[1].(*object.Boolean)
	if !ok {
		return &object.Error{Message: "crs.UTM: north must be a boolean"}
	}

	z := int(zone.Value)
	if z < 1 || z > 60 {
		return &object.Error{Message: fmt.Sprintf("crs.UTM: zone must be between 1 and 60, got %d", z)}
	}

	var epsg int
	var hemisphere string
	if northObj.Value {
		// UTM North zones: EPSG 32601-32660
		epsg = 32600 + z
		hemisphere = "N"
	} else {
		// UTM South zones: EPSG 32701-32760
		epsg = 32700 + z
		hemisphere = "S"
	}

	return &object.CRS{
		EPSGCode:     epsg,
		CRSName:      fmt.Sprintf("WGS 84 / UTM zone %d%s", z, hemisphere),
		IsGeographic: false,
		IsProjected:  true,
		Units:        "metre",
	}
}

// getMetadata returns the metadata for a given CRS, or a generic default.
func getMetadata(c *object.CRS) *crsMetadata {
	if m, ok := knownMetadata[c.EPSGCode]; ok {
		return m
	}
	// Generate generic metadata for unknown CRS (e.g., UTM zones)
	proj4 := fmt.Sprintf("+init=epsg:%d", c.EPSGCode)
	wktDef := fmt.Sprintf(`LOCAL_CS["EPSG:%d"]`, c.EPSGCode)
	d := "Unknown"
	e := "Unknown"
	if c.EPSGCode >= 32601 && c.EPSGCode <= 32660 {
		zone := c.EPSGCode - 32600
		d = "World Geodetic System 1984"
		e = "WGS 84"
		proj4 = fmt.Sprintf("+proj=utm +zone=%d +datum=WGS84 +units=m +no_defs +type=crs", zone)
		wktDef = fmt.Sprintf(`PROJCRS["WGS 84 / UTM zone %dN",BASEGEOGCRS["WGS 84",DATUM["World Geodetic System 1984",ELLIPSOID["WGS 84",6378137,298.257223563]]]]`, zone)
	} else if c.EPSGCode >= 32701 && c.EPSGCode <= 32760 {
		zone := c.EPSGCode - 32700
		d = "World Geodetic System 1984"
		e = "WGS 84"
		proj4 = fmt.Sprintf("+proj=utm +zone=%d +south +datum=WGS84 +units=m +no_defs +type=crs", zone)
		wktDef = fmt.Sprintf(`PROJCRS["WGS 84 / UTM zone %dS",BASEGEOGCRS["WGS 84",DATUM["World Geodetic System 1984",ELLIPSOID["WGS 84",6378137,298.257223563]]]]`, zone)
	}
	return &crsMetadata{
		Datum:     d,
		Ellipsoid: e,
		Proj4:     proj4,
		WKTDef:    wktDef,
		BoundsW:   -180, BoundsS: -90, BoundsE: 180, BoundsN: 90,
	}
}

// toProj4 returns the PROJ4 string representation of a CRS.
func toProj4(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "crs.toProj4 expects 1 argument (CRS)"}
	}
	c, ok := args[0].(*object.CRS)
	if !ok {
		return &object.Error{Message: "crs.toProj4: argument must be a CRS"}
	}
	m := getMetadata(c)
	return &object.String{Value: m.Proj4}
}

// toWKT returns the WKT representation of a CRS.
func toWKT(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "crs.toWKT expects 1 argument (CRS)"}
	}
	c, ok := args[0].(*object.CRS)
	if !ok {
		return &object.Error{Message: "crs.toWKT: argument must be a CRS"}
	}
	m := getMetadata(c)
	return &object.String{Value: m.WKTDef}
}

// bounds returns the bounding box of a CRS as a Map with keys w, s, e, n.
func bounds(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "crs.bounds expects 1 argument (CRS)"}
	}
	c, ok := args[0].(*object.CRS)
	if !ok {
		return &object.Error{Message: "crs.bounds: argument must be a CRS"}
	}
	m := getMetadata(c)
	pairs := []object.MapPair{
		{Key: &object.String{Value: "w"}, Value: &object.Float{Value: m.BoundsW}},
		{Key: &object.String{Value: "s"}, Value: &object.Float{Value: m.BoundsS}},
		{Key: &object.String{Value: "e"}, Value: &object.Float{Value: m.BoundsE}},
		{Key: &object.String{Value: "n"}, Value: &object.Float{Value: m.BoundsN}},
	}
	return &object.Map{Pairs: pairs}
}

// datum returns the datum name of a CRS.
func datum(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "crs.datum expects 1 argument (CRS)"}
	}
	c, ok := args[0].(*object.CRS)
	if !ok {
		return &object.Error{Message: "crs.datum: argument must be a CRS"}
	}
	m := getMetadata(c)
	return &object.String{Value: m.Datum}
}

// ellipsoid returns the ellipsoid name of a CRS.
func ellipsoid(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "crs.ellipsoid expects 1 argument (CRS)"}
	}
	c, ok := args[0].(*object.CRS)
	if !ok {
		return &object.Error{Message: "crs.ellipsoid: argument must be a CRS"}
	}
	m := getMetadata(c)
	return &object.String{Value: m.Ellipsoid}
}
