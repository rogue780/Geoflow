// Package core provides fundamental functional programming utilities for the GeoFlow standard library.
package core

import (
	"fmt"

	"github.com/rogue780/geoflow/internal/object"
)

// GetExports returns all exported functions for the core module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		"typeof":     &object.Builtin{Name: "core.typeof", Fn: typeofFn},
		"instanceof": &object.Builtin{Name: "core.instanceof", Fn: instanceofFn},
		"identity":   &object.Builtin{Name: "core.identity", Fn: identity},
		"compose":    &object.Builtin{Name: "core.compose", Fn: compose},
		"pipe":       &object.Builtin{Name: "core.pipe", Fn: pipe},
		"flip":       &object.Builtin{Name: "core.flip", Fn: flip},
		"curry":      &object.Builtin{Name: "core.curry", Fn: curry},
	}
}

func typeofFn(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "core.typeof expects 1 argument"}
	}
	return &object.String{Value: string(args[0].Type())}
}

func instanceofFn(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "core.instanceof expects 2 arguments (value, typeName)"}
	}
	typeName, ok := args[1].(*object.String)
	if !ok {
		return &object.Error{Message: "core.instanceof: second argument must be a string"}
	}
	return object.NativeBoolToBooleanObject(string(args[0].Type()) == typeName.Value)
}

func identity(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "core.identity expects 1 argument"}
	}
	return args[0]
}

// compose returns a ComposedFunction such that compose(f, g)(x) = f(g(x)).
func compose(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "core.compose expects 2 arguments (f, g)"}
	}
	if !isCallable(args[0]) {
		return &object.Error{Message: "core.compose: first argument must be callable"}
	}
	if !isCallable(args[1]) {
		return &object.Error{Message: "core.compose: second argument must be callable"}
	}
	return &object.ComposedFunction{
		Outer: args[0],
		Inner: args[1],
	}
}

// pipe returns a ComposedFunction such that pipe(f, g)(x) = g(f(x)).
func pipe(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "core.pipe expects 2 arguments (f, g)"}
	}
	if !isCallable(args[0]) {
		return &object.Error{Message: "core.pipe: first argument must be callable"}
	}
	if !isCallable(args[1]) {
		return &object.Error{Message: "core.pipe: second argument must be callable"}
	}
	return &object.ComposedFunction{
		Outer: args[1],
		Inner: args[0],
	}
}

// flip returns a Builtin that swaps the first two arguments before calling the
// underlying function. Since we cannot call GeoFlow functions from Go without
// the evaluator, this creates a placeholder Builtin that wraps only Builtin callables.
func flip(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "core.flip expects 1 argument (function)"}
	}
	if !isCallable(args[0]) {
		return &object.Error{Message: "core.flip: argument must be callable"}
	}
	fn := args[0]
	return &object.Builtin{
		Name: "flipped",
		Fn: func(innerArgs ...object.Object) object.Object {
			if len(innerArgs) < 2 {
				return &object.Error{Message: "flipped function expects at least 2 arguments"}
			}
			// Swap first two arguments.
			swapped := make([]object.Object, len(innerArgs))
			copy(swapped, innerArgs)
			swapped[0], swapped[1] = swapped[1], swapped[0]
			// If the wrapped function is a Builtin, call it directly.
			if b, ok := fn.(*object.Builtin); ok && b.Fn != nil {
				return b.Fn(swapped...)
			}
			// For Function/ComposedFunction, return an error since we need the evaluator.
			return &object.Error{
				Message: fmt.Sprintf("core.flip: cannot directly call %s; use the evaluator", fn.Type()),
			}
		},
	}
}

// curry is a placeholder that returns the function as-is.
// Full currying requires evaluator support.
func curry(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "core.curry expects 1 argument (function)"}
	}
	if !isCallable(args[0]) {
		return &object.Error{Message: "core.curry: argument must be callable"}
	}
	return args[0]
}

// isCallable checks whether an object is a callable type.
func isCallable(o object.Object) bool {
	switch o.Type() {
	case object.FUNCTION_OBJ, object.BUILTIN_OBJ, object.COMPOSED_FUNCTION_OBJ:
		return true
	default:
		return false
	}
}
