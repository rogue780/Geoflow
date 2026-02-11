// Package gfassert provides assertion functions for the GeoFlow standard library.
package gfassert

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/rogue780/geoflow/internal/object"
)

// GetExports returns all exported functions for the assert module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		"assert":         &object.Builtin{Name: "assert.assert", Fn: assertFn},
		"equal":          &object.Builtin{Name: "assert.equal", Fn: equal},
		"notEqual":       &object.Builtin{Name: "assert.notEqual", Fn: notEqual},
		"approximately":  &object.Builtin{Name: "assert.approximately", Fn: approximately},
		"greaterThan":    &object.Builtin{Name: "assert.greaterThan", Fn: greaterThan},
		"lessThan":       &object.Builtin{Name: "assert.lessThan", Fn: lessThan},
		"greaterOrEqual": &object.Builtin{Name: "assert.greaterOrEqual", Fn: greaterOrEqual},
		"lessOrEqual":    &object.Builtin{Name: "assert.lessOrEqual", Fn: lessOrEqual},
		"between":        &object.Builtin{Name: "assert.between", Fn: between},
		"isNil":          &object.Builtin{Name: "assert.isNil", Fn: isNil},
		"notNil":         &object.Builtin{Name: "assert.notNil", Fn: notNil},
		"ok":             &object.Builtin{Name: "assert.ok", Fn: ok},
		"err":            &object.Builtin{Name: "assert.err", Fn: errFn},
		"empty":          &object.Builtin{Name: "assert.empty", Fn: empty},
		"notEmpty":       &object.Builtin{Name: "assert.notEmpty", Fn: notEmpty},
		"length":         &object.Builtin{Name: "assert.length", Fn: length},
		"contains":       &object.Builtin{Name: "assert.contains", Fn: contains},
		"startsWith":     &object.Builtin{Name: "assert.startsWith", Fn: startsWith},
		"endsWith":       &object.Builtin{Name: "assert.endsWith", Fn: endsWith},
		"matches":        &object.Builtin{Name: "assert.matches", Fn: matches},
		"throws":         &object.Builtin{Name: "assert.throws", Fn: throws},
	}
}

// fail returns an assertion error.
func fail(msg string) *object.Error {
	return &object.Error{Message: fmt.Sprintf("assertion failed: %s", msg)}
}

// toFloat64 extracts a float64 from a numeric object.
func toFloat64(o object.Object) (float64, bool) {
	switch v := o.(type) {
	case *object.Float:
		return v.Value, true
	case *object.Integer:
		return float64(v.Value), true
	default:
		return 0, false
	}
}

// objectsEqual checks if two objects are equal by value.
func objectsEqual(a, b object.Object) bool {
	if a.Type() != b.Type() {
		// Allow int/float cross-comparison.
		af, aOk := toFloat64(a)
		bf, bOk := toFloat64(b)
		if aOk && bOk {
			return af == bf
		}
		return false
	}
	switch av := a.(type) {
	case *object.Integer:
		return av.Value == b.(*object.Integer).Value
	case *object.Float:
		return av.Value == b.(*object.Float).Value
	case *object.String:
		return av.Value == b.(*object.String).Value
	case *object.Boolean:
		return av.Value == b.(*object.Boolean).Value
	case *object.Nil:
		return true
	default:
		return a.Inspect() == b.Inspect()
	}
}

// compareNumeric compares two numeric objects. Returns -1, 0, 1 or ok=false.
func compareNumeric(a, b object.Object) (int, bool) {
	af, aOk := toFloat64(a)
	bf, bOk := toFloat64(b)
	if !aOk || !bOk {
		return 0, false
	}
	if af < bf {
		return -1, true
	}
	if af > bf {
		return 1, true
	}
	return 0, true
}

func assertFn(args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return &object.Error{Message: "assert.assert expects 1-2 arguments (condition, msg?)"}
	}
	if !object.IsTruthy(args[0]) {
		msg := "condition is not truthy"
		if len(args) == 2 {
			if s, ok := args[1].(*object.String); ok {
				msg = s.Value
			}
		}
		return fail(msg)
	}
	return object.NIL
}

func equal(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "assert.equal expects 2 arguments (actual, expected)"}
	}
	if !objectsEqual(args[0], args[1]) {
		return fail(fmt.Sprintf("expected %s to equal %s", args[0].Inspect(), args[1].Inspect()))
	}
	return object.NIL
}

func notEqual(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "assert.notEqual expects 2 arguments (actual, expected)"}
	}
	if objectsEqual(args[0], args[1]) {
		return fail(fmt.Sprintf("expected %s to not equal %s", args[0].Inspect(), args[1].Inspect()))
	}
	return object.NIL
}

func approximately(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "assert.approximately expects 3 arguments (actual, expected, epsilon)"}
	}
	actual, aOk := toFloat64(args[0])
	expected, eOk := toFloat64(args[1])
	epsilon, epOk := toFloat64(args[2])
	if !aOk || !eOk || !epOk {
		return &object.Error{Message: "assert.approximately: all arguments must be numeric"}
	}
	if math.Abs(actual-expected) > epsilon {
		return fail(fmt.Sprintf("expected %g to be approximately %g (epsilon %g)", actual, expected, epsilon))
	}
	return object.NIL
}

func greaterThan(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "assert.greaterThan expects 2 arguments (a, b)"}
	}
	cmp, ok := compareNumeric(args[0], args[1])
	if !ok {
		return &object.Error{Message: "assert.greaterThan: both arguments must be numeric"}
	}
	if cmp <= 0 {
		return fail(fmt.Sprintf("expected %s > %s", args[0].Inspect(), args[1].Inspect()))
	}
	return object.NIL
}

func lessThan(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "assert.lessThan expects 2 arguments (a, b)"}
	}
	cmp, ok := compareNumeric(args[0], args[1])
	if !ok {
		return &object.Error{Message: "assert.lessThan: both arguments must be numeric"}
	}
	if cmp >= 0 {
		return fail(fmt.Sprintf("expected %s < %s", args[0].Inspect(), args[1].Inspect()))
	}
	return object.NIL
}

func greaterOrEqual(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "assert.greaterOrEqual expects 2 arguments (a, b)"}
	}
	cmp, ok := compareNumeric(args[0], args[1])
	if !ok {
		return &object.Error{Message: "assert.greaterOrEqual: both arguments must be numeric"}
	}
	if cmp < 0 {
		return fail(fmt.Sprintf("expected %s >= %s", args[0].Inspect(), args[1].Inspect()))
	}
	return object.NIL
}

func lessOrEqual(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "assert.lessOrEqual expects 2 arguments (a, b)"}
	}
	cmp, ok := compareNumeric(args[0], args[1])
	if !ok {
		return &object.Error{Message: "assert.lessOrEqual: both arguments must be numeric"}
	}
	if cmp > 0 {
		return fail(fmt.Sprintf("expected %s <= %s", args[0].Inspect(), args[1].Inspect()))
	}
	return object.NIL
}

func between(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "assert.between expects 3 arguments (value, low, high)"}
	}
	val, vOk := toFloat64(args[0])
	low, lOk := toFloat64(args[1])
	high, hOk := toFloat64(args[2])
	if !vOk || !lOk || !hOk {
		return &object.Error{Message: "assert.between: all arguments must be numeric"}
	}
	if val < low || val > high {
		return fail(fmt.Sprintf("expected %g to be between %g and %g", val, low, high))
	}
	return object.NIL
}

func isNil(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "assert.isNil expects 1 argument"}
	}
	if _, ok := args[0].(*object.Nil); !ok {
		return fail(fmt.Sprintf("expected nil, got %s (%s)", args[0].Inspect(), args[0].Type()))
	}
	return object.NIL
}

func notNil(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "assert.notNil expects 1 argument"}
	}
	if _, ok := args[0].(*object.Nil); ok {
		return fail("expected non-nil value, got nil")
	}
	return object.NIL
}

func ok(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "assert.ok expects 1 argument (Result)"}
	}
	result, ok := args[0].(*object.Result)
	if !ok {
		return &object.Error{Message: "assert.ok: argument must be a Result"}
	}
	if !result.IsOk {
		return fail(fmt.Sprintf("expected Ok, got Err(%s)", result.Value.Inspect()))
	}
	return object.NIL
}

func errFn(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "assert.err expects 1 argument (Result)"}
	}
	result, ok := args[0].(*object.Result)
	if !ok {
		return &object.Error{Message: "assert.err: argument must be a Result"}
	}
	if result.IsOk {
		return fail(fmt.Sprintf("expected Err, got Ok(%s)", result.Value.Inspect()))
	}
	return object.NIL
}

// collectionLen returns the length of a collection object, or -1 if not a collection.
func collectionLen(o object.Object) int {
	switch v := o.(type) {
	case *object.List:
		return len(v.Elements)
	case *object.Map:
		return len(v.Pairs)
	case *object.Set:
		return len(v.Elements)
	case *object.Tuple:
		return len(v.Elements)
	case *object.String:
		return len(v.Value)
	default:
		return -1
	}
}

func empty(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "assert.empty expects 1 argument (collection)"}
	}
	n := collectionLen(args[0])
	if n < 0 {
		return &object.Error{Message: fmt.Sprintf("assert.empty: argument of type %s is not a collection", args[0].Type())}
	}
	if n != 0 {
		return fail(fmt.Sprintf("expected empty collection, got length %d", n))
	}
	return object.NIL
}

func notEmpty(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "assert.notEmpty expects 1 argument (collection)"}
	}
	n := collectionLen(args[0])
	if n < 0 {
		return &object.Error{Message: fmt.Sprintf("assert.notEmpty: argument of type %s is not a collection", args[0].Type())}
	}
	if n == 0 {
		return fail("expected non-empty collection, got empty")
	}
	return object.NIL
}

func length(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "assert.length expects 2 arguments (collection, expectedLength)"}
	}
	n := collectionLen(args[0])
	if n < 0 {
		return &object.Error{Message: fmt.Sprintf("assert.length: first argument of type %s is not a collection", args[0].Type())}
	}
	expected, ok := args[1].(*object.Integer)
	if !ok {
		return &object.Error{Message: "assert.length: second argument must be an integer"}
	}
	if int64(n) != expected.Value {
		return fail(fmt.Sprintf("expected length %d, got %d", expected.Value, n))
	}
	return object.NIL
}

func contains(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "assert.contains expects 2 arguments (collection, item)"}
	}
	switch coll := args[0].(type) {
	case *object.List:
		for _, elem := range coll.Elements {
			if objectsEqual(elem, args[1]) {
				return object.NIL
			}
		}
		return fail(fmt.Sprintf("list does not contain %s", args[1].Inspect()))
	case *object.Set:
		if coll.Contains(args[1]) {
			return object.NIL
		}
		return fail(fmt.Sprintf("set does not contain %s", args[1].Inspect()))
	case *object.String:
		sub, ok := args[1].(*object.String)
		if !ok {
			return &object.Error{Message: "assert.contains: when collection is a string, item must be a string"}
		}
		if strings.Contains(coll.Value, sub.Value) {
			return object.NIL
		}
		return fail(fmt.Sprintf("string %q does not contain %q", coll.Value, sub.Value))
	default:
		return &object.Error{Message: fmt.Sprintf("assert.contains: first argument of type %s is not a supported collection", args[0].Type())}
	}
}

func startsWith(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "assert.startsWith expects 2 arguments (string, prefix)"}
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "assert.startsWith: first argument must be a string"}
	}
	prefix, ok := args[1].(*object.String)
	if !ok {
		return &object.Error{Message: "assert.startsWith: second argument must be a string"}
	}
	if !strings.HasPrefix(s.Value, prefix.Value) {
		return fail(fmt.Sprintf("expected %q to start with %q", s.Value, prefix.Value))
	}
	return object.NIL
}

func endsWith(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "assert.endsWith expects 2 arguments (string, suffix)"}
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "assert.endsWith: first argument must be a string"}
	}
	suffix, ok := args[1].(*object.String)
	if !ok {
		return &object.Error{Message: "assert.endsWith: second argument must be a string"}
	}
	if !strings.HasSuffix(s.Value, suffix.Value) {
		return fail(fmt.Sprintf("expected %q to end with %q", s.Value, suffix.Value))
	}
	return object.NIL
}

func matches(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "assert.matches expects 2 arguments (string, pattern)"}
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "assert.matches: first argument must be a string"}
	}
	pattern, ok := args[1].(*object.String)
	if !ok {
		return &object.Error{Message: "assert.matches: second argument must be a string (regex pattern)"}
	}
	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("assert.matches: invalid regex %q: %s", pattern.Value, err.Error())}
	}
	if !re.MatchString(s.Value) {
		return fail(fmt.Sprintf("expected %q to match pattern %q", s.Value, pattern.Value))
	}
	return object.NIL
}

func throws(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "assert.throws expects 1 argument"}
	}
	// Since we cannot call GeoFlow functions from Go without the evaluator,
	// we check if the argument is already an Error object.
	if args[0].Type() == object.ERROR_OBJ {
		return object.NIL
	}
	return fail(fmt.Sprintf("expected an error, got %s (%s)", args[0].Inspect(), args[0].Type()))
}
