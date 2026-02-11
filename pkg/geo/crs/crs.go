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

// GetExports returns all exported constants and functions for the crs module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		"WGS84":        wgs84,
		"WebMercator":  webMercator,
		"NAD83":        nad83,
		"NAD27":        nad27,
		"ETRS89":       etrs89,
		"fromEPSG":     &object.Builtin{Name: "crs.fromEPSG", Fn: fromEPSG},
		"UTM":          &object.Builtin{Name: "crs.UTM", Fn: utm},
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
