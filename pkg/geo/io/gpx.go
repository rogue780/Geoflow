package io

import (
	"github.com/rogue780/geoflow/internal/object"
)

// GetGPXExports returns GPX I/O functions (currently placeholders).
func GetGPXExports() map[string]object.Object {
	return map[string]object.Object{
		"fromGPX": &object.Builtin{Name: "geoio.fromGPX", Fn: fromGPX},
		"toGPX":   &object.Builtin{Name: "geoio.toGPX", Fn: toGPX},
	}
}

// fromGPX is a placeholder for GPX parsing.
func fromGPX(args ...object.Object) object.Object {
	return &object.Error{Message: "GPX parsing not yet implemented"}
}

// toGPX is a placeholder for GPX output.
func toGPX(args ...object.Object) object.Object {
	return &object.Error{Message: "GPX output not yet implemented"}
}
