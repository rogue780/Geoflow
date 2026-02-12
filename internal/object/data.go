package object

import (
	"fmt"
	"sort"
	"strings"
)

// ── Array Type ──

const ARRAY_OBJ Type = "Array"

// Array represents a fixed-size, multidimensional numeric array.
// Data is stored in row-major order as a flat []float64 slice.
type Array struct {
	Data  []float64
	Shape []int
}

func (a *Array) Type() Type { return ARRAY_OBJ }
func (a *Array) Inspect() string {
	// Format data elements.
	elems := make([]string, len(a.Data))
	for i, v := range a.Data {
		elems[i] = formatFloat(v)
	}

	// Format shape.
	shapeParts := make([]string, len(a.Shape))
	for i, s := range a.Shape {
		shapeParts[i] = fmt.Sprintf("%d", s)
	}

	return fmt.Sprintf("Array([%s], shape=[%s])",
		strings.Join(elems, ", "),
		strings.Join(shapeParts, ", "))
}

// Size returns the total number of elements in the array.
func (a *Array) Size() int {
	if len(a.Shape) == 0 {
		return 0
	}
	size := 1
	for _, s := range a.Shape {
		size *= s
	}
	return size
}

// ── Series Type ──

const SERIES_OBJ Type = "Series"

// Series represents a labeled one-dimensional data series.
type Series struct {
	Data  []Object
	Index []string
	Name  string
	DType string
}

func (s *Series) Type() Type { return SERIES_OBJ }
func (s *Series) Inspect() string {
	var b strings.Builder
	b.WriteString("Series(")
	if s.Name != "" {
		b.WriteString(fmt.Sprintf("name=%q, ", s.Name))
	}
	b.WriteString("[")
	for i, d := range s.Data {
		if i > 0 {
			b.WriteString(", ")
		}
		if i < len(s.Index) && s.Index[i] != "" {
			b.WriteString(fmt.Sprintf("%s: %s", s.Index[i], d.Inspect()))
		} else {
			b.WriteString(d.Inspect())
		}
	}
	b.WriteString("])")
	return b.String()
}

// ── DataFrame Type ──

const DATAFRAME_OBJ Type = "DataFrame"

// DataFrame represents a tabular data structure with named columns.
type DataFrame struct {
	Columns  map[string]*Series
	ColOrder []string
	Length   int
}

func (df *DataFrame) Type() Type { return DATAFRAME_OBJ }
func (df *DataFrame) Inspect() string {
	var b strings.Builder
	b.WriteString("DataFrame(")
	b.WriteString(fmt.Sprintf("columns=[%s], ", strings.Join(df.ColOrder, ", ")))
	b.WriteString(fmt.Sprintf("rows=%d)", df.Length))
	return b.String()
}

// formatFloat formats a float64 for display, matching the convention used elsewhere.
func formatFloat(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%g.0", v)
	}
	return fmt.Sprintf("%g", v)
}

// ── Helpers for sorting map keys ──

// SortedKeys returns the keys of a Map in sorted order.
func SortedKeys(m *Map) []string {
	keys := make([]string, 0, len(m.Pairs))
	for _, p := range m.Pairs {
		if s, ok := p.Key.(*String); ok {
			keys = append(keys, s.Value)
		}
	}
	sort.Strings(keys)
	return keys
}
