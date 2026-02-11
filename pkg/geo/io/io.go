// Package io provides additional geospatial I/O functions (WKB, KML, etc.) for GeoFlow.
package io

import (
	"github.com/rogue780/geoflow/internal/object"
)

// GetExports returns all exported functions for the geo.io module.
func GetExports() map[string]object.Object {
	exports := make(map[string]object.Object)
	for k, v := range GetKMLExports() {
		exports[k] = v
	}
	for k, v := range GetWKBExports() {
		exports[k] = v
	}
	return exports
}
