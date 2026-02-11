// Package types defines the type system for the GeoFlow language.
package types

import (
	"fmt"
	"strings"
)

// GeoType represents a type in the GeoFlow type system.
type GeoType interface {
	TypeName() string
	String() string
	Equals(other GeoType) bool
}

// PrimitiveType represents a built-in primitive type.
type PrimitiveType struct {
	Name string
}

func (p *PrimitiveType) TypeName() string { return p.Name }
func (p *PrimitiveType) String() string   { return p.Name }
func (p *PrimitiveType) Equals(other GeoType) bool {
	if o, ok := other.(*PrimitiveType); ok {
		return p.Name == o.Name
	}
	return false
}

// GenericType represents a parameterized type like List<int>.
type GenericType struct {
	Name       string
	TypeParams []GeoType
}

func (g *GenericType) TypeName() string { return g.Name }
func (g *GenericType) String() string {
	if len(g.TypeParams) == 0 {
		return g.Name
	}
	params := make([]string, len(g.TypeParams))
	for i, p := range g.TypeParams {
		params[i] = p.String()
	}
	return fmt.Sprintf("%s<%s>", g.Name, strings.Join(params, ", "))
}
func (g *GenericType) Equals(other GeoType) bool {
	o, ok := other.(*GenericType)
	if !ok || g.Name != o.Name || len(g.TypeParams) != len(o.TypeParams) {
		return false
	}
	for i, p := range g.TypeParams {
		if !p.Equals(o.TypeParams[i]) {
			return false
		}
	}
	return true
}

// FunctionType represents a function type.
type FunctionType struct {
	ParamTypes []GeoType
	ReturnType GeoType
}

func (f *FunctionType) TypeName() string { return "Fn" }
func (f *FunctionType) String() string {
	params := make([]string, len(f.ParamTypes))
	for i, p := range f.ParamTypes {
		params[i] = p.String()
	}
	return fmt.Sprintf("Fn<(%s) -> %s>", strings.Join(params, ", "), f.ReturnType.String())
}
func (f *FunctionType) Equals(other GeoType) bool {
	o, ok := other.(*FunctionType)
	if !ok || len(f.ParamTypes) != len(o.ParamTypes) {
		return false
	}
	for i, p := range f.ParamTypes {
		if !p.Equals(o.ParamTypes[i]) {
			return false
		}
	}
	return f.ReturnType.Equals(o.ReturnType)
}

// AutoType represents an inferred type.
type AutoType struct{}

func (a *AutoType) TypeName() string           { return "auto" }
func (a *AutoType) String() string             { return "auto" }
func (a *AutoType) Equals(other GeoType) bool  { return true } // auto matches anything

// NilType represents the nil type.
type NilType struct{}

func (n *NilType) TypeName() string           { return "nil" }
func (n *NilType) String() string             { return "nil" }
func (n *NilType) Equals(other GeoType) bool  { _, ok := other.(*NilType); return ok }

// TypeVariable represents a type parameter (e.g., T in List<T>).
type TypeVariable struct {
	Name string
}

func (tv *TypeVariable) TypeName() string          { return tv.Name }
func (tv *TypeVariable) String() string            { return tv.Name }
func (tv *TypeVariable) Equals(other GeoType) bool { return true } // Type variables match anything during checking

// Singleton types for primitives.
var (
	IntType    = &PrimitiveType{Name: "int"}
	FloatType  = &PrimitiveType{Name: "float"}
	BoolType   = &PrimitiveType{Name: "bool"}
	StringType = &PrimitiveType{Name: "string"}
	ByteType   = &PrimitiveType{Name: "byte"}
	Nil        = &NilType{}
	Auto       = &AutoType{}
)

// IsNumeric returns true if the type is numeric.
func IsNumeric(t GeoType) bool {
	if p, ok := t.(*PrimitiveType); ok {
		return p.Name == "int" || p.Name == "float" || p.Name == "byte"
	}
	return false
}

// IsCompatible checks if two types are assignment-compatible.
func IsCompatible(target, source GeoType) bool {
	if target.Equals(source) {
		return true
	}
	if _, ok := target.(*AutoType); ok {
		return true
	}
	if _, ok := source.(*NilType); ok {
		// nil is compatible with Option types
		if g, ok := target.(*GenericType); ok && g.Name == "Option" {
			return true
		}
	}
	// int -> float promotion
	if target.Equals(FloatType) && source.Equals(IntType) {
		return true
	}
	return false
}
