// Package object defines the runtime value types for the GeoFlow interpreter.
package object

import (
	"fmt"
	"math"
	"strings"

	"github.com/rogue780/geoflow/internal/ast"
)

// Type represents the type of a runtime object.
type Type string

const (
	INTEGER_OBJ      Type = "int"
	FLOAT_OBJ        Type = "float"
	BOOLEAN_OBJ      Type = "bool"
	STRING_OBJ       Type = "string"
	NIL_OBJ          Type = "nil"
	LIST_OBJ         Type = "List"
	MAP_OBJ          Type = "Map"
	TUPLE_OBJ        Type = "Tuple"
	FUNCTION_OBJ     Type = "Function"
	BUILTIN_OBJ      Type = "Builtin"
	RETURN_VALUE_OBJ Type = "ReturnValue"
	ERROR_OBJ        Type = "Error"
	BREAK_OBJ        Type = "Break"
	CONTINUE_OBJ     Type = "Continue"
	OPTION_OBJ       Type = "Option"
	RESULT_OBJ       Type = "Result"
)

// Object is the interface all runtime values implement.
type Object interface {
	Type() Type
	Inspect() string
}

// Integer represents a 64-bit signed integer.
type Integer struct {
	Value int64
}

func (i *Integer) Type() Type      { return INTEGER_OBJ }
func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }

// Float represents a 64-bit floating-point number.
type Float struct {
	Value float64
}

func (f *Float) Type() Type { return FLOAT_OBJ }
func (f *Float) Inspect() string {
	if f.Value == math.Trunc(f.Value) && !math.IsInf(f.Value, 0) && !math.IsNaN(f.Value) {
		return fmt.Sprintf("%g.0", f.Value)
	}
	return fmt.Sprintf("%g", f.Value)
}

// Boolean represents a boolean value.
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() Type      { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }

// String represents a UTF-8 string.
type String struct {
	Value string
}

func (s *String) Type() Type      { return STRING_OBJ }
func (s *String) Inspect() string { return s.Value }

// Nil represents the absence of a value.
type Nil struct{}

func (n *Nil) Type() Type      { return NIL_OBJ }
func (n *Nil) Inspect() string { return "nil" }

// List represents an ordered collection.
type List struct {
	Elements []Object
}

func (l *List) Type() Type { return LIST_OBJ }
func (l *List) Inspect() string {
	elems := make([]string, len(l.Elements))
	for i, e := range l.Elements {
		elems[i] = e.Inspect()
	}
	return fmt.Sprintf("[%s]", strings.Join(elems, ", "))
}

// MapPair stores a key-value pair preserving order.
type MapPair struct {
	Key   Object
	Value Object
}

// Map represents a key-value collection.
type Map struct {
	Pairs []MapPair
}

func (m *Map) Type() Type { return MAP_OBJ }
func (m *Map) Inspect() string {
	pairs := make([]string, len(m.Pairs))
	for i, p := range m.Pairs {
		pairs[i] = fmt.Sprintf("%s: %s", p.Key.Inspect(), p.Value.Inspect())
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

// Get retrieves a value by key using string comparison.
func (m *Map) Get(key string) (Object, bool) {
	for _, p := range m.Pairs {
		if s, ok := p.Key.(*String); ok && s.Value == key {
			return p.Value, true
		}
	}
	return nil, false
}

// Tuple represents a fixed-size heterogeneous collection.
type Tuple struct {
	Elements []Object
}

func (t *Tuple) Type() Type { return TUPLE_OBJ }
func (t *Tuple) Inspect() string {
	elems := make([]string, len(t.Elements))
	for i, e := range t.Elements {
		elems[i] = e.Inspect()
	}
	return fmt.Sprintf("(%s)", strings.Join(elems, ", "))
}

// Function represents a user-defined function.
type Function struct {
	Name       string
	Parameters []*ast.FunctionParameter
	Body       *ast.BlockExpression
	Env        *Environment
}

func (f *Function) Type() Type { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	params := make([]string, len(f.Parameters))
	for i, p := range f.Parameters {
		params[i] = p.Name.Value
	}
	name := f.Name
	if name == "" {
		name = "<anonymous>"
	}
	return fmt.Sprintf("fn %s(%s)", name, strings.Join(params, ", "))
}

// BuiltinFunction is the type for built-in function implementations.
type BuiltinFunction func(args ...Object) Object

// Builtin represents a built-in function.
type Builtin struct {
	Name string
	Fn   BuiltinFunction
}

func (b *Builtin) Type() Type      { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string { return fmt.Sprintf("<builtin: %s>", b.Name) }

// ReturnValue wraps a return value to unwind the call stack.
type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() Type      { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string { return rv.Value.Inspect() }

// Error represents a runtime error.
type Error struct {
	Message string
}

func (e *Error) Type() Type      { return ERROR_OBJ }
func (e *Error) Inspect() string { return fmt.Sprintf("Error: %s", e.Message) }

// Break signals a break from a loop.
type Break struct{}

func (b *Break) Type() Type      { return BREAK_OBJ }
func (b *Break) Inspect() string { return "break" }

// Continue signals a continue in a loop.
type Continue struct{}

func (c *Continue) Type() Type      { return CONTINUE_OBJ }
func (c *Continue) Inspect() string { return "continue" }

// Option represents Some(value) or None.
type Option struct {
	Value  Object // nil means None
	IsSome bool
}

func (o *Option) Type() Type { return OPTION_OBJ }
func (o *Option) Inspect() string {
	if o.IsSome {
		return fmt.Sprintf("Some(%s)", o.Value.Inspect())
	}
	return "None"
}

// Result represents Ok(value) or Err(error).
type Result struct {
	Value Object
	IsOk  bool
}

func (r *Result) Type() Type { return RESULT_OBJ }
func (r *Result) Inspect() string {
	if r.IsOk {
		return fmt.Sprintf("Ok(%s)", r.Value.Inspect())
	}
	return fmt.Sprintf("Err(%s)", r.Value.Inspect())
}

// Singleton objects.
var (
	TRUE_OBJ  = &Boolean{Value: true}
	FALSE_OBJ = &Boolean{Value: false}
	NIL       = &Nil{}
	BREAK     = &Break{}
	CONT      = &Continue{}
	NONE      = &Option{IsSome: false}
)

// NativeBoolToBooleanObject converts a Go bool to a GeoFlow Boolean.
func NativeBoolToBooleanObject(input bool) *Boolean {
	if input {
		return TRUE_OBJ
	}
	return FALSE_OBJ
}

// IsTruthy determines if an object is truthy.
func IsTruthy(obj Object) bool {
	switch o := obj.(type) {
	case *Boolean:
		return o.Value
	case *Nil:
		return false
	case *Integer:
		return o.Value != 0
	case *Float:
		return o.Value != 0
	case *String:
		return o.Value != ""
	case *List:
		return len(o.Elements) > 0
	case *Option:
		return o.IsSome
	default:
		return true
	}
}
