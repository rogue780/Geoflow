// Package collections provides collection utility functions for the GeoFlow standard library.
package collections

import (
	"github.com/rogue780/geoflow/internal/object"
)

// GetSetExports returns all exported functions for the Set sub-module.
func GetSetExports() map[string]object.Object {
	return map[string]object.Object{
		"empty":    &object.Builtin{Name: "Set.empty", Fn: emptySet},
		"of":       &object.Builtin{Name: "Set.of", Fn: of},
		"fromList": &object.Builtin{Name: "Set.fromList", Fn: fromList},
	}
}

func emptySet(args ...object.Object) object.Object {
	if len(args) != 0 {
		return &object.Error{Message: "Set.empty expects 0 arguments"}
	}
	return object.NewSet()
}

func of(args ...object.Object) object.Object {
	s := object.NewSet()
	for _, arg := range args {
		s = s.Add(arg)
	}
	return s
}

func fromList(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "Set.fromList expects 1 argument (list)"}
	}
	list, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "Set.fromList: argument must be a List"}
	}
	s := object.NewSet()
	for _, elem := range list.Elements {
		s = s.Add(elem)
	}
	return s
}
