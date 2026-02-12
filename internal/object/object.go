// Package object defines the runtime value types for the GeoFlow interpreter.
package object

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/rogue780/geoflow/internal/ast"
)

// Type represents the type of a runtime object.
type Type string

const (
	INTEGER_OBJ           Type = "int"
	FLOAT_OBJ             Type = "float"
	BOOLEAN_OBJ           Type = "bool"
	STRING_OBJ            Type = "string"
	NIL_OBJ               Type = "nil"
	LIST_OBJ              Type = "List"
	MAP_OBJ               Type = "Map"
	SET_OBJ               Type = "Set"
	TUPLE_OBJ             Type = "Tuple"
	FUNCTION_OBJ          Type = "Function"
	BUILTIN_OBJ           Type = "Builtin"
	COMPOSED_FUNCTION_OBJ Type = "ComposedFunction"
	RETURN_VALUE_OBJ      Type = "ReturnValue"
	ERROR_OBJ             Type = "Error"
	BREAK_OBJ             Type = "Break"
	CONTINUE_OBJ          Type = "Continue"
	OPTION_OBJ            Type = "Option"
	RESULT_OBJ            Type = "Result"
	VECTOR_OBJ            Type = "Vector"
	MATRIX_OBJ            Type = "Matrix"
	COMPLEX_OBJ           Type = "Complex"
	MODULE_OBJ            Type = "Module"
	STRUCT_OBJ            Type = "Struct"
	STRUCT_DEF_OBJ        Type = "StructDef"
	ENUM_VARIANT_OBJ      Type = "EnumVariant"
	ENUM_DEF_OBJ          Type = "EnumDef"
	CRS_OBJ               Type = "CRS"
	DATETIME_OBJ          Type = "DateTime"
	DURATION_OBJ          Type = "Duration"
	RTREE_OBJ             Type = "RTree"
)

// Object is the interface all runtime values implement.
type Object interface {
	Type() Type
	Inspect() string
}

// GlobalApplyFunction is a callback set by the evaluator so that stdlib packages
// can invoke user-defined functions (lambdas, closures) without importing eval.
// The evaluator registers applyFunction here during initialization.
var GlobalApplyFunction func(fn Object, args []Object) Object

// CallFunction invokes a callable object (Builtin, Function, ComposedFunction)
// using the global apply function callback. Returns an error object if the
// callback is not registered or the function is not callable.
func CallFunction(fn Object, args ...Object) Object {
	if b, ok := fn.(*Builtin); ok {
		if b.Fn != nil {
			return b.Fn(args...)
		}
	}
	if GlobalApplyFunction != nil {
		return GlobalApplyFunction(fn, args)
	}
	return &Error{Message: fmt.Sprintf("cannot call %s (evaluator not registered)", fn.Type())}
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

// Vector represents a mathematical vector of float64 values.
type Vector struct {
	Elements []float64
}

func (v *Vector) Type() Type { return VECTOR_OBJ }
func (v *Vector) Inspect() string {
	elems := make([]string, len(v.Elements))
	for i, e := range v.Elements {
		if e == math.Trunc(e) && !math.IsInf(e, 0) && !math.IsNaN(e) {
			elems[i] = fmt.Sprintf("%g.0", e)
		} else {
			elems[i] = fmt.Sprintf("%g", e)
		}
	}
	return fmt.Sprintf("vec(%s)", strings.Join(elems, ", "))
}

// Matrix represents a mathematical matrix of float64 values.
type Matrix struct {
	Rows int
	Cols int
	Data [][]float64
}

func (m *Matrix) Type() Type { return MATRIX_OBJ }
func (m *Matrix) Inspect() string {
	rows := make([]string, m.Rows)
	for i, row := range m.Data {
		elems := make([]string, len(row))
		for j, e := range row {
			if e == math.Trunc(e) && !math.IsInf(e, 0) && !math.IsNaN(e) {
				elems[j] = fmt.Sprintf("%g.0", e)
			} else {
				elems[j] = fmt.Sprintf("%g", e)
			}
		}
		rows[i] = fmt.Sprintf("[%s]", strings.Join(elems, ", "))
	}
	return fmt.Sprintf("mat(%s)", strings.Join(rows, ", "))
}

// Complex represents a complex number.
type Complex struct {
	Real float64
	Imag float64
}

func (c *Complex) Type() Type { return COMPLEX_OBJ }
func (c *Complex) Inspect() string {
	if c.Imag >= 0 {
		return fmt.Sprintf("%g+%gi", c.Real, c.Imag)
	}
	return fmt.Sprintf("%g%gi", c.Real, c.Imag)
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
	case *Set:
		return len(o.Elements) > 0
	default:
		return true
	}
}

// ── Module System ──

// Module represents an imported module with named exports.
type Module struct {
	Name    string
	Exports map[string]Object
}

func (m *Module) Type() Type      { return MODULE_OBJ }
func (m *Module) Inspect() string { return fmt.Sprintf("<module: %s>", m.Name) }

// ── Set Type ──

// Set represents an unordered collection of unique values.
type Set struct {
	Elements map[string]Object // keyed by Inspect() for dedup
}

func (s *Set) Type() Type { return SET_OBJ }
func (s *Set) Inspect() string {
	elems := make([]string, 0, len(s.Elements))
	keys := make([]string, 0, len(s.Elements))
	for k := range s.Elements {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		elems = append(elems, s.Elements[k].Inspect())
	}
	return fmt.Sprintf("Set{%s}", strings.Join(elems, ", "))
}

// NewSet creates a new empty Set.
func NewSet() *Set {
	return &Set{Elements: make(map[string]Object)}
}

// Add adds an element to the set, returning a new set.
func (s *Set) Add(obj Object) *Set {
	ns := NewSet()
	for k, v := range s.Elements {
		ns.Elements[k] = v
	}
	ns.Elements[obj.Inspect()] = obj
	return ns
}

// Contains checks if the set contains an element.
func (s *Set) Contains(obj Object) bool {
	_, ok := s.Elements[obj.Inspect()]
	return ok
}

// ToList converts a set to a list.
func (s *Set) ToList() *List {
	elems := make([]Object, 0, len(s.Elements))
	keys := make([]string, 0, len(s.Elements))
	for k := range s.Elements {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		elems = append(elems, s.Elements[k])
	}
	return &List{Elements: elems}
}

// ── Struct Type ──

// StructDef describes a struct type (its constructor).
type StructDef struct {
	TypeName   string
	FieldNames []string
}

func (sd *StructDef) Type() Type      { return STRUCT_DEF_OBJ }
func (sd *StructDef) Inspect() string { return fmt.Sprintf("<struct def: %s>", sd.TypeName) }

// Struct represents an instantiated struct value.
type Struct struct {
	TypeName string
	Fields   map[string]Object
	Order    []string // field order
}

func (s *Struct) Type() Type { return STRUCT_OBJ }
func (s *Struct) Inspect() string {
	fields := make([]string, len(s.Order))
	for i, name := range s.Order {
		fields[i] = fmt.Sprintf("%s: %s", name, s.Fields[name].Inspect())
	}
	return fmt.Sprintf("%s{%s}", s.TypeName, strings.Join(fields, ", "))
}

// ── Enum Type ──

// EnumDef describes an enum type.
type EnumDef struct {
	TypeName     string
	VariantNames []string
	VariantArity map[string]int // number of fields per variant; 0 = unit variant
}

func (ed *EnumDef) Type() Type      { return ENUM_DEF_OBJ }
func (ed *EnumDef) Inspect() string { return fmt.Sprintf("<enum def: %s>", ed.TypeName) }

// EnumVariant represents an instantiated enum variant.
type EnumVariant struct {
	TypeName    string
	VariantName string
	Payload     []Object
}

func (ev *EnumVariant) Type() Type { return ENUM_VARIANT_OBJ }
func (ev *EnumVariant) Inspect() string {
	if len(ev.Payload) == 0 {
		return ev.VariantName
	}
	elems := make([]string, len(ev.Payload))
	for i, p := range ev.Payload {
		elems[i] = p.Inspect()
	}
	return fmt.Sprintf("%s(%s)", ev.VariantName, strings.Join(elems, ", "))
}

// ── ComposedFunction ──

// ComposedFunction wraps two callables into a composed function (f . g)(x) = f(g(x)).
type ComposedFunction struct {
	Outer Object // f
	Inner Object // g
}

func (cf *ComposedFunction) Type() Type      { return COMPOSED_FUNCTION_OBJ }
func (cf *ComposedFunction) Inspect() string { return "<composed function>" }

// ── CRS Type ──

// CRS represents a Coordinate Reference System.
type CRS struct {
	EPSGCode     int
	CRSName      string
	IsGeographic bool
	IsProjected  bool
	Units        string
}

func (c *CRS) Type() Type { return CRS_OBJ }
func (c *CRS) Inspect() string {
	return fmt.Sprintf("CRS(EPSG:%d, %s)", c.EPSGCode, c.CRSName)
}

// ── DateTime/Duration Types ──

// DateTime wraps Go's time.Time.
type DateTime struct {
	Value time.Time
}

func (dt *DateTime) Type() Type      { return DATETIME_OBJ }
func (dt *DateTime) Inspect() string { return dt.Value.Format(time.RFC3339) }

// Duration wraps Go's time.Duration.
type Duration struct {
	Value time.Duration
}

func (d *Duration) Type() Type      { return DURATION_OBJ }
func (d *Duration) Inspect() string { return d.Value.String() }

// ── RTree Type ──

// RTreeEntry represents a single entry in an RTree spatial index.
type RTreeEntry struct {
	MinX, MinY, MaxX, MaxY float64
	Item                   Object
}

// RTree represents a spatial index using a flat list of bounding-box entries.
type RTree struct {
	Entries []RTreeEntry
}

func (rt *RTree) Type() Type      { return RTREE_OBJ }
func (rt *RTree) Inspect() string { return fmt.Sprintf("RTree(%d entries)", len(rt.Entries)) }
