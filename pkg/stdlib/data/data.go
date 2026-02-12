// Package data provides Array, Series, and DataFrame constructors for the GeoFlow standard library.
package data

import (
	"fmt"
	"sort"

	"github.com/rogue780/geoflow/internal/object"
)

// GetExports returns all exported functions for the data module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		// Array constructors
		"zeros":      &object.Builtin{Name: "data.zeros", Fn: arrayZeros},
		"ones":       &object.Builtin{Name: "data.ones", Fn: arrayOnes},
		"full":       &object.Builtin{Name: "data.full", Fn: arrayFull},
		"eye":        &object.Builtin{Name: "data.eye", Fn: arrayEye},
		"linspace":   &object.Builtin{Name: "data.linspace", Fn: arrayLinspace},
		"arange":     &object.Builtin{Name: "data.arange", Fn: arrayArange},
		"fromList":   &object.Builtin{Name: "data.fromList", Fn: arrayFromList},
		"fromNested": &object.Builtin{Name: "data.fromNested", Fn: arrayFromNested},

		// Series constructors
		"Series":        &object.Builtin{Name: "data.Series", Fn: seriesNew},
		"SeriesFromMap": &object.Builtin{Name: "data.SeriesFromMap", Fn: seriesFromMap},

		// DataFrame constructors
		"DataFrame":         &object.Builtin{Name: "data.DataFrame", Fn: dataFrameNew},
		"DataFrameFromRows": &object.Builtin{Name: "data.DataFrameFromRows", Fn: dataFrameFromRows},
	}
}

// ── Helpers ──

// extractShape converts a List of Integers into a []int shape slice and
// computes the total number of elements implied by the shape.
func extractShape(list *object.List) ([]int, int, *object.Error) {
	if len(list.Elements) == 0 {
		return nil, 0, &object.Error{Message: "shape must be a non-empty list of integers"}
	}
	shape := make([]int, len(list.Elements))
	total := 1
	for i, elem := range list.Elements {
		intVal, ok := elem.(*object.Integer)
		if !ok {
			return nil, 0, &object.Error{Message: fmt.Sprintf("shape element %d must be an integer, got %s", i, elem.Type())}
		}
		if intVal.Value <= 0 {
			return nil, 0, &object.Error{Message: fmt.Sprintf("shape element %d must be positive, got %d", i, intVal.Value)}
		}
		shape[i] = int(intVal.Value)
		total *= shape[i]
	}
	return shape, total, nil
}

// toFloat64 converts an integer or float object to float64.
func toFloat64(obj object.Object) (float64, bool) {
	switch v := obj.(type) {
	case *object.Float:
		return v.Value, true
	case *object.Integer:
		return float64(v.Value), true
	default:
		return 0, false
	}
}

// objectToFloat64 converts an Object to float64 and returns an error object if it fails.
func objectToFloat64(obj object.Object, context string) (float64, *object.Error) {
	v, ok := toFloat64(obj)
	if !ok {
		return 0, &object.Error{Message: fmt.Sprintf("%s: expected a numeric value, got %s", context, obj.Type())}
	}
	return v, nil
}

// ── Array Constructors ──

// arrayZeros creates an Array filled with zeros.
// Usage: zeros(shape) where shape is a List of Integers.
func arrayZeros(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "data.zeros expects 1 argument (shape: List<int>)"}
	}
	shapeList, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "data.zeros: argument must be a List of integers"}
	}
	shape, total, err := extractShape(shapeList)
	if err != nil {
		return err
	}
	data := make([]float64, total)
	return &object.Array{Data: data, Shape: shape}
}

// arrayOnes creates an Array filled with ones.
// Usage: ones(shape) where shape is a List of Integers.
func arrayOnes(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "data.ones expects 1 argument (shape: List<int>)"}
	}
	shapeList, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "data.ones: argument must be a List of integers"}
	}
	shape, total, err := extractShape(shapeList)
	if err != nil {
		return err
	}
	data := make([]float64, total)
	for i := range data {
		data[i] = 1.0
	}
	return &object.Array{Data: data, Shape: shape}
}

// arrayFull creates an Array filled with a given value.
// Usage: full(shape, value) where shape is a List of Integers and value is numeric.
func arrayFull(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "data.full expects 2 arguments (shape: List<int>, value: number)"}
	}
	shapeList, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "data.full: first argument must be a List of integers"}
	}
	fillVal, err := objectToFloat64(args[1], "data.full")
	if err != nil {
		return err
	}
	shape, total, shapeErr := extractShape(shapeList)
	if shapeErr != nil {
		return shapeErr
	}
	data := make([]float64, total)
	for i := range data {
		data[i] = fillVal
	}
	return &object.Array{Data: data, Shape: shape}
}

// arrayEye creates an n x n identity matrix as an Array.
// Usage: eye(n)
func arrayEye(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "data.eye expects 1 argument (n: int)"}
	}
	nObj, ok := args[0].(*object.Integer)
	if !ok {
		return &object.Error{Message: "data.eye: argument must be an integer"}
	}
	n := int(nObj.Value)
	if n <= 0 {
		return &object.Error{Message: "data.eye: n must be positive"}
	}
	data := make([]float64, n*n)
	for i := 0; i < n; i++ {
		data[i*n+i] = 1.0
	}
	return &object.Array{Data: data, Shape: []int{n, n}}
}

// arrayLinspace creates an Array of n evenly spaced values from start to stop (inclusive).
// Usage: linspace(start, stop, n)
func arrayLinspace(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "data.linspace expects 3 arguments (start, stop, n)"}
	}
	start, err := objectToFloat64(args[0], "data.linspace")
	if err != nil {
		return err
	}
	stop, err := objectToFloat64(args[1], "data.linspace")
	if err != nil {
		return err
	}
	nObj, ok := args[2].(*object.Integer)
	if !ok {
		return &object.Error{Message: "data.linspace: third argument (n) must be an integer"}
	}
	n := int(nObj.Value)
	if n <= 0 {
		return &object.Error{Message: "data.linspace: n must be positive"}
	}
	data := make([]float64, n)
	if n == 1 {
		data[0] = start
	} else {
		step := (stop - start) / float64(n-1)
		for i := 0; i < n; i++ {
			data[i] = start + float64(i)*step
		}
	}
	return &object.Array{Data: data, Shape: []int{n}}
}

// arrayArange creates an Array of values from start to stop (exclusive) with a given step.
// Usage: arange(start, stop) or arange(start, stop, step)
func arrayArange(args ...object.Object) object.Object {
	if len(args) < 2 || len(args) > 3 {
		return &object.Error{Message: "data.arange expects 2-3 arguments (start, stop[, step])"}
	}
	start, err := objectToFloat64(args[0], "data.arange")
	if err != nil {
		return err
	}
	stop, err := objectToFloat64(args[1], "data.arange")
	if err != nil {
		return err
	}
	step := 1.0
	if len(args) == 3 {
		step, err = objectToFloat64(args[2], "data.arange")
		if err != nil {
			return err
		}
	}
	if step == 0 {
		return &object.Error{Message: "data.arange: step cannot be zero"}
	}
	var data []float64
	if step > 0 {
		for v := start; v < stop; v += step {
			data = append(data, v)
		}
	} else {
		for v := start; v > stop; v += step {
			data = append(data, v)
		}
	}
	if len(data) == 0 {
		return &object.Array{Data: []float64{}, Shape: []int{0}}
	}
	return &object.Array{Data: data, Shape: []int{len(data)}}
}

// arrayFromList creates a 1D Array from a flat list of numbers.
// Usage: fromList(list)
func arrayFromList(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "data.fromList expects 1 argument (list: List<number>)"}
	}
	list, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "data.fromList: argument must be a List"}
	}
	data := make([]float64, len(list.Elements))
	for i, elem := range list.Elements {
		v, convErr := objectToFloat64(elem, "data.fromList")
		if convErr != nil {
			return convErr
		}
		data[i] = v
	}
	return &object.Array{Data: data, Shape: []int{len(data)}}
}

// arrayFromNested creates a 2D Array from a nested list of numbers (list of lists).
// All inner lists must have the same length.
// Usage: fromNested(nestedList)
func arrayFromNested(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "data.fromNested expects 1 argument (nestedList: List<List<number>>)"}
	}
	outer, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "data.fromNested: argument must be a List of Lists"}
	}
	if len(outer.Elements) == 0 {
		return &object.Array{Data: []float64{}, Shape: []int{0, 0}}
	}

	rows := len(outer.Elements)
	cols := -1
	var data []float64

	for i, rowObj := range outer.Elements {
		rowList, ok := rowObj.(*object.List)
		if !ok {
			return &object.Error{Message: fmt.Sprintf("data.fromNested: element %d must be a List, got %s", i, rowObj.Type())}
		}
		if cols == -1 {
			cols = len(rowList.Elements)
		} else if len(rowList.Elements) != cols {
			return &object.Error{Message: fmt.Sprintf("data.fromNested: row %d has %d elements, expected %d", i, len(rowList.Elements), cols)}
		}
		for j, elem := range rowList.Elements {
			v, convErr := objectToFloat64(elem, fmt.Sprintf("data.fromNested[%d][%d]", i, j))
			if convErr != nil {
				return convErr
			}
			data = append(data, v)
		}
	}

	if cols == -1 {
		cols = 0
	}
	return &object.Array{Data: data, Shape: []int{rows, cols}}
}

// ── Series Constructors ──

// seriesNew creates a Series from a List of data values.
// Usage: Series(data) or Series(data, name) or Series(data, index, name)
func seriesNew(args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 3 {
		return &object.Error{Message: "data.Series expects 1-3 arguments (data[, index][, name])"}
	}

	dataList, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "data.Series: first argument must be a List"}
	}

	seriesData := make([]object.Object, len(dataList.Elements))
	copy(seriesData, dataList.Elements)

	var index []string
	name := ""
	dtype := inferDType(seriesData)

	switch len(args) {
	case 1:
		// Series(data)
	case 2:
		// Series(data, name) where name is a String
		// or Series(data, index) where index is a List
		switch v := args[1].(type) {
		case *object.String:
			name = v.Value
		case *object.List:
			idx, err := extractIndex(v, len(seriesData))
			if err != nil {
				return err
			}
			index = idx
		default:
			return &object.Error{Message: "data.Series: second argument must be a String (name) or List (index)"}
		}
	case 3:
		// Series(data, index, name)
		idxList, ok := args[1].(*object.List)
		if !ok {
			return &object.Error{Message: "data.Series: second argument must be a List (index)"}
		}
		idx, err := extractIndex(idxList, len(seriesData))
		if err != nil {
			return err
		}
		index = idx
		nameStr, ok := args[2].(*object.String)
		if !ok {
			return &object.Error{Message: "data.Series: third argument must be a String (name)"}
		}
		name = nameStr.Value
	}

	return &object.Series{
		Data:  seriesData,
		Index: index,
		Name:  name,
		DType: dtype,
	}
}

// seriesFromMap creates a Series from a Map.
// The map keys become the index labels and values become the data.
// Usage: SeriesFromMap(map)
func seriesFromMap(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "data.SeriesFromMap expects 1 argument (map: Map<string, value>)"}
	}
	m, ok := args[0].(*object.Map)
	if !ok {
		return &object.Error{Message: "data.SeriesFromMap: argument must be a Map"}
	}

	// Collect keys in sorted order for deterministic output.
	type kv struct {
		key   string
		value object.Object
	}
	pairs := make([]kv, 0, len(m.Pairs))
	for _, p := range m.Pairs {
		keyStr, ok := p.Key.(*object.String)
		if !ok {
			return &object.Error{Message: "data.SeriesFromMap: map keys must be strings"}
		}
		pairs = append(pairs, kv{key: keyStr.Value, value: p.Value})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].key < pairs[j].key
	})

	index := make([]string, len(pairs))
	data := make([]object.Object, len(pairs))
	for i, p := range pairs {
		index[i] = p.key
		data[i] = p.value
	}

	return &object.Series{
		Data:  data,
		Index: index,
		Name:  "",
		DType: inferDType(data),
	}
}

// extractIndex converts a List of Strings into a []string index slice.
func extractIndex(list *object.List, expectedLen int) ([]string, *object.Error) {
	if len(list.Elements) != expectedLen {
		return nil, &object.Error{Message: fmt.Sprintf("index length (%d) must match data length (%d)", len(list.Elements), expectedLen)}
	}
	index := make([]string, len(list.Elements))
	for i, elem := range list.Elements {
		s, ok := elem.(*object.String)
		if !ok {
			return nil, &object.Error{Message: fmt.Sprintf("index element %d must be a string, got %s", i, elem.Type())}
		}
		index[i] = s.Value
	}
	return index, nil
}

// inferDType determines the dominant type of the data elements.
func inferDType(data []object.Object) string {
	if len(data) == 0 {
		return "empty"
	}
	hasFloat := false
	hasInt := false
	hasString := false
	for _, d := range data {
		switch d.(type) {
		case *object.Float:
			hasFloat = true
		case *object.Integer:
			hasInt = true
		case *object.String:
			hasString = true
		}
	}
	if hasString {
		return "string"
	}
	if hasFloat {
		return "float"
	}
	if hasInt {
		return "int"
	}
	return "mixed"
}

// ── DataFrame Constructors ──

// dataFrameNew creates a DataFrame from a Map of column name -> List.
// Usage: DataFrame(columns: Map<string, List>)
func dataFrameNew(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "data.DataFrame expects 1 argument (columns: Map<string, List>)"}
	}
	m, ok := args[0].(*object.Map)
	if !ok {
		return &object.Error{Message: "data.DataFrame: argument must be a Map<string, List>"}
	}

	columns := make(map[string]*object.Series)
	colOrder := make([]string, 0, len(m.Pairs))
	dfLength := -1

	for _, p := range m.Pairs {
		keyStr, ok := p.Key.(*object.String)
		if !ok {
			return &object.Error{Message: "data.DataFrame: column names must be strings"}
		}
		colName := keyStr.Value
		colList, ok := p.Value.(*object.List)
		if !ok {
			return &object.Error{Message: fmt.Sprintf("data.DataFrame: column %q must be a List", colName)}
		}

		if dfLength == -1 {
			dfLength = len(colList.Elements)
		} else if len(colList.Elements) != dfLength {
			return &object.Error{Message: fmt.Sprintf("data.DataFrame: column %q has %d elements, expected %d", colName, len(colList.Elements), dfLength)}
		}

		seriesData := make([]object.Object, len(colList.Elements))
		copy(seriesData, colList.Elements)

		columns[colName] = &object.Series{
			Data:  seriesData,
			Name:  colName,
			DType: inferDType(seriesData),
		}
		colOrder = append(colOrder, colName)
	}

	if dfLength == -1 {
		dfLength = 0
	}

	return &object.DataFrame{
		Columns:  columns,
		ColOrder: colOrder,
		Length:   dfLength,
	}
}

// dataFrameFromRows creates a DataFrame from a List of Maps (row-oriented data).
// Each Map represents one row with string keys as column names.
// Usage: DataFrameFromRows(rows: List<Map>)
func dataFrameFromRows(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "data.DataFrameFromRows expects 1 argument (rows: List<Map>)"}
	}
	rowList, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "data.DataFrameFromRows: argument must be a List of Maps"}
	}

	if len(rowList.Elements) == 0 {
		return &object.DataFrame{
			Columns:  make(map[string]*object.Series),
			ColOrder: []string{},
			Length:   0,
		}
	}

	// Discover column names from the first row, preserving order.
	firstRow, ok := rowList.Elements[0].(*object.Map)
	if !ok {
		return &object.Error{Message: "data.DataFrameFromRows: each row must be a Map"}
	}

	colOrder := make([]string, 0, len(firstRow.Pairs))
	colSet := make(map[string]bool)
	for _, p := range firstRow.Pairs {
		keyStr, ok := p.Key.(*object.String)
		if !ok {
			return &object.Error{Message: "data.DataFrameFromRows: row keys must be strings"}
		}
		colOrder = append(colOrder, keyStr.Value)
		colSet[keyStr.Value] = true
	}

	// Also check subsequent rows for additional columns.
	for i := 1; i < len(rowList.Elements); i++ {
		rowMap, ok := rowList.Elements[i].(*object.Map)
		if !ok {
			return &object.Error{Message: fmt.Sprintf("data.DataFrameFromRows: element %d must be a Map", i)}
		}
		for _, p := range rowMap.Pairs {
			keyStr, ok := p.Key.(*object.String)
			if !ok {
				return &object.Error{Message: "data.DataFrameFromRows: row keys must be strings"}
			}
			if !colSet[keyStr.Value] {
				colOrder = append(colOrder, keyStr.Value)
				colSet[keyStr.Value] = true
			}
		}
	}

	// Build column data.
	numRows := len(rowList.Elements)
	colData := make(map[string][]object.Object)
	for _, col := range colOrder {
		colData[col] = make([]object.Object, numRows)
	}

	for i, rowObj := range rowList.Elements {
		rowMap, ok := rowObj.(*object.Map)
		if !ok {
			return &object.Error{Message: fmt.Sprintf("data.DataFrameFromRows: element %d must be a Map", i)}
		}
		// Set values from this row.
		present := make(map[string]bool)
		for _, p := range rowMap.Pairs {
			keyStr, _ := p.Key.(*object.String) // already validated above
			colData[keyStr.Value][i] = p.Value
			present[keyStr.Value] = true
		}
		// Fill missing columns with nil.
		for _, col := range colOrder {
			if !present[col] {
				colData[col][i] = object.NIL
			}
		}
	}

	// Create Series for each column.
	columns := make(map[string]*object.Series)
	for _, col := range colOrder {
		columns[col] = &object.Series{
			Data:  colData[col],
			Name:  col,
			DType: inferDType(colData[col]),
		}
	}

	return &object.DataFrame{
		Columns:  columns,
		ColOrder: colOrder,
		Length:   numRows,
	}
}
