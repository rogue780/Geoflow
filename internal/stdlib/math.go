// Package stdlib implements the GeoFlow standard library functions.
package stdlib

import (
	"fmt"
	"math"
	"math/cmplx"
	"sort"

	"github.com/rogue780/geoflow/internal/object"
)

// helper to extract a float64 from an Object.
func toFloat(o object.Object) (float64, bool) {
	switch v := o.(type) {
	case *object.Float:
		return v.Value, true
	case *object.Integer:
		return float64(v.Value), true
	default:
		return 0, false
	}
}

// helper to extract a list of float64 from a List object.
func toFloatSlice(o object.Object) ([]float64, *object.Error) {
	list, ok := o.(*object.List)
	if !ok {
		return nil, &object.Error{Message: "expected a list of numbers"}
	}
	result := make([]float64, len(list.Elements))
	for i, elem := range list.Elements {
		v, ok := toFloat(elem)
		if !ok {
			return nil, &object.Error{Message: fmt.Sprintf("element %d is not a number", i)}
		}
		result[i] = v
	}
	return result, nil
}

// errArgs returns an error for wrong argument count.
func errArgs(name string, expected, got int) *object.Error {
	return &object.Error{Message: fmt.Sprintf("%s expects %d argument(s), got %d", name, expected, got)}
}

// errNumeric returns an error for non-numeric arguments.
func errNumeric(name string) *object.Error {
	return &object.Error{Message: fmt.Sprintf("%s expects numeric argument(s)", name)}
}

// unaryFloat creates a builtin that applies f to one numeric argument.
func unaryFloat(name string, f func(float64) float64) *object.Builtin {
	return &object.Builtin{
		Name: name,
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs(name, 1, len(args))
			}
			v, ok := toFloat(args[0])
			if !ok {
				return errNumeric(name)
			}
			return &object.Float{Value: f(v)}
		},
	}
}

// binaryFloat creates a builtin that applies f to two numeric arguments.
func binaryFloat(name string, f func(float64, float64) float64) *object.Builtin {
	return &object.Builtin{
		Name: name,
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return errArgs(name, 2, len(args))
			}
			a, ok1 := toFloat(args[0])
			b, ok2 := toFloat(args[1])
			if !ok1 || !ok2 {
				return errNumeric(name)
			}
			return &object.Float{Value: f(a, b)}
		},
	}
}

// GetMathBuiltins returns all math library built-in functions.
func GetMathBuiltins() map[string]*object.Builtin {
	builtins := map[string]*object.Builtin{
		// ── Constants ──
		"PI":       {Name: "PI", Fn: func(args ...object.Object) object.Object { return &object.Float{Value: math.Pi} }},
		"E":        {Name: "E", Fn: func(args ...object.Object) object.Object { return &object.Float{Value: math.E} }},
		"TAU":      {Name: "TAU", Fn: func(args ...object.Object) object.Object { return &object.Float{Value: 2 * math.Pi} }},
		"INFINITY": {Name: "INFINITY", Fn: func(args ...object.Object) object.Object { return &object.Float{Value: math.Inf(1)} }},
		"NAN":      {Name: "NAN", Fn: func(args ...object.Object) object.Object { return &object.Float{Value: math.NaN()} }},

		// ── Rounding ──
		"floor": unaryFloat("floor", math.Floor),
		"ceil":  unaryFloat("ceil", math.Ceil),
		"round": unaryFloat("round", math.Round),
		"trunc": unaryFloat("trunc", math.Trunc),

		// ── Trigonometric ──
		"sin":  unaryFloat("sin", math.Sin),
		"cos":  unaryFloat("cos", math.Cos),
		"tan":  unaryFloat("tan", math.Tan),
		"asin": unaryFloat("asin", math.Asin),
		"acos": unaryFloat("acos", math.Acos),
		"atan": unaryFloat("atan", math.Atan),
		"atan2": binaryFloat("atan2", math.Atan2),
		"sinh": unaryFloat("sinh", math.Sinh),
		"cosh": unaryFloat("cosh", math.Cosh),
		"tanh": unaryFloat("tanh", math.Tanh),

		// ── Conversion ──
		"toRadians": unaryFloat("toRadians", func(deg float64) float64 { return deg * math.Pi / 180.0 }),
		"toDegrees": unaryFloat("toDegrees", func(rad float64) float64 { return rad * 180.0 / math.Pi }),

		// ── Exponential / Logarithmic ──
		"exp":   unaryFloat("exp", math.Exp),
		"exp2":  unaryFloat("exp2", math.Exp2),
		"log":   unaryFloat("log", math.Log),
		"log2":  unaryFloat("log2", math.Log2),
		"log10": unaryFloat("log10", math.Log10),
		"pow":   binaryFloat("pow", math.Pow),
		"cbrt":  unaryFloat("cbrt", math.Cbrt),
		"hypot": binaryFloat("hypot", math.Hypot),

		// ── Sign / Comparison ──
		"sign": {
			Name: "sign",
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return errArgs("sign", 1, len(args))
				}
				v, ok := toFloat(args[0])
				if !ok {
					return errNumeric("sign")
				}
				switch {
				case v > 0:
					return &object.Integer{Value: 1}
				case v < 0:
					return &object.Integer{Value: -1}
				default:
					return &object.Integer{Value: 0}
				}
			},
		},
		"clamp": {
			Name: "clamp",
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return errArgs("clamp", 3, len(args))
				}
				v, ok1 := toFloat(args[0])
				lo, ok2 := toFloat(args[1])
				hi, ok3 := toFloat(args[2])
				if !ok1 || !ok2 || !ok3 {
					return errNumeric("clamp")
				}
				return &object.Float{Value: math.Max(lo, math.Min(hi, v))}
			},
		},
		"lerp": {
			Name: "lerp",
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return errArgs("lerp", 3, len(args))
				}
				a, ok1 := toFloat(args[0])
				b, ok2 := toFloat(args[1])
				t, ok3 := toFloat(args[2])
				if !ok1 || !ok2 || !ok3 {
					return errNumeric("lerp")
				}
				return &object.Float{Value: a + (b-a)*t}
			},
		},

		// ── Predicates ──
		"isNaN":      {Name: "isNaN", Fn: func(args ...object.Object) object.Object { v, ok := toFloat(args[0]); if !ok { return object.FALSE_OBJ }; return object.NativeBoolToBooleanObject(math.IsNaN(v)) }},
		"isInfinite": {Name: "isInfinite", Fn: func(args ...object.Object) object.Object { v, ok := toFloat(args[0]); if !ok { return object.FALSE_OBJ }; return object.NativeBoolToBooleanObject(math.IsInf(v, 0)) }},
		"isFinite":   {Name: "isFinite", Fn: func(args ...object.Object) object.Object { v, ok := toFloat(args[0]); if !ok { return object.FALSE_OBJ }; return object.NativeBoolToBooleanObject(!math.IsInf(v, 0) && !math.IsNaN(v)) }},

		// ── GCD / LCM ──
		"gcd": {
			Name: "gcd",
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return errArgs("gcd", 2, len(args))
				}
				a, ok1 := args[0].(*object.Integer)
				b, ok2 := args[1].(*object.Integer)
				if !ok1 || !ok2 {
					return &object.Error{Message: "gcd expects integer arguments"}
				}
				return &object.Integer{Value: gcd(absInt(a.Value), absInt(b.Value))}
			},
		},
		"lcm": {
			Name: "lcm",
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return errArgs("lcm", 2, len(args))
				}
				a, ok1 := args[0].(*object.Integer)
				b, ok2 := args[1].(*object.Integer)
				if !ok1 || !ok2 {
					return &object.Error{Message: "lcm expects integer arguments"}
				}
				av, bv := absInt(a.Value), absInt(b.Value)
				if av == 0 || bv == 0 {
					return &object.Integer{Value: 0}
				}
				return &object.Integer{Value: av / gcd(av, bv) * bv}
			},
		},
		"factorial": {
			Name: "factorial",
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return errArgs("factorial", 1, len(args))
				}
				n, ok := args[0].(*object.Integer)
				if !ok || n.Value < 0 {
					return &object.Error{Message: "factorial expects a non-negative integer"}
				}
				result := int64(1)
				for i := int64(2); i <= n.Value; i++ {
					result *= i
				}
				return &object.Integer{Value: result}
			},
		},
	}

	// ── Statistics ──
	addStatisticsBuiltins(builtins)

	// ── Linear Algebra ──
	addLinearAlgebraBuiltins(builtins)

	// ── Complex Numbers ──
	addComplexBuiltins(builtins)

	// ── Numerical Methods ──
	addNumericalBuiltins(builtins)

	return builtins
}

func gcd(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func absInt(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

// ────────────────────────────────────────────────────────────
// Statistics
// ────────────────────────────────────────────────────────────

func addStatisticsBuiltins(builtins map[string]*object.Builtin) {
	builtins["mean"] = &object.Builtin{
		Name: "mean",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("mean", 1, len(args))
			}
			vals, err := toFloatSlice(args[0])
			if err != nil {
				return err
			}
			if len(vals) == 0 {
				return &object.Error{Message: "mean of empty list"}
			}
			sum := 0.0
			for _, v := range vals {
				sum += v
			}
			return &object.Float{Value: sum / float64(len(vals))}
		},
	}

	builtins["median"] = &object.Builtin{
		Name: "median",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("median", 1, len(args))
			}
			vals, err := toFloatSlice(args[0])
			if err != nil {
				return err
			}
			if len(vals) == 0 {
				return &object.Error{Message: "median of empty list"}
			}
			sorted := make([]float64, len(vals))
			copy(sorted, vals)
			sort.Float64s(sorted)
			n := len(sorted)
			if n%2 == 0 {
				return &object.Float{Value: (sorted[n/2-1] + sorted[n/2]) / 2}
			}
			return &object.Float{Value: sorted[n/2]}
		},
	}

	builtins["mode"] = &object.Builtin{
		Name: "mode",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("mode", 1, len(args))
			}
			vals, err := toFloatSlice(args[0])
			if err != nil {
				return err
			}
			if len(vals) == 0 {
				return &object.Error{Message: "mode of empty list"}
			}
			counts := make(map[float64]int)
			for _, v := range vals {
				counts[v]++
			}
			maxCount := 0
			mode := vals[0]
			for v, c := range counts {
				if c > maxCount || (c == maxCount && v < mode) {
					maxCount = c
					mode = v
				}
			}
			return &object.Float{Value: mode}
		},
	}

	builtins["variance"] = &object.Builtin{
		Name: "variance",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("variance", 1, len(args))
			}
			vals, err := toFloatSlice(args[0])
			if err != nil {
				return err
			}
			if len(vals) < 2 {
				return &object.Error{Message: "variance requires at least 2 values"}
			}
			mean := computeMean(vals)
			sumSq := 0.0
			for _, v := range vals {
				d := v - mean
				sumSq += d * d
			}
			return &object.Float{Value: sumSq / float64(len(vals)-1)}
		},
	}

	builtins["stddev"] = &object.Builtin{
		Name: "stddev",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("stddev", 1, len(args))
			}
			vals, err := toFloatSlice(args[0])
			if err != nil {
				return err
			}
			if len(vals) < 2 {
				return &object.Error{Message: "stddev requires at least 2 values"}
			}
			mean := computeMean(vals)
			sumSq := 0.0
			for _, v := range vals {
				d := v - mean
				sumSq += d * d
			}
			return &object.Float{Value: math.Sqrt(sumSq / float64(len(vals)-1))}
		},
	}

	builtins["percentile"] = &object.Builtin{
		Name: "percentile",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return errArgs("percentile", 2, len(args))
			}
			vals, err := toFloatSlice(args[0])
			if err != nil {
				return err
			}
			p, ok := toFloat(args[1])
			if !ok || p < 0 || p > 100 {
				return &object.Error{Message: "percentile expects a number between 0 and 100"}
			}
			if len(vals) == 0 {
				return &object.Error{Message: "percentile of empty list"}
			}
			sorted := make([]float64, len(vals))
			copy(sorted, vals)
			sort.Float64s(sorted)
			rank := (p / 100) * float64(len(sorted)-1)
			lower := int(math.Floor(rank))
			upper := int(math.Ceil(rank))
			if lower == upper {
				return &object.Float{Value: sorted[lower]}
			}
			frac := rank - float64(lower)
			return &object.Float{Value: sorted[lower]*(1-frac) + sorted[upper]*frac}
		},
	}

	builtins["correlation"] = &object.Builtin{
		Name: "correlation",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return errArgs("correlation", 2, len(args))
			}
			xs, err1 := toFloatSlice(args[0])
			ys, err2 := toFloatSlice(args[1])
			if err1 != nil {
				return err1
			}
			if err2 != nil {
				return err2
			}
			if len(xs) != len(ys) {
				return &object.Error{Message: "correlation requires lists of equal length"}
			}
			if len(xs) < 2 {
				return &object.Error{Message: "correlation requires at least 2 data points"}
			}
			mx, my := computeMean(xs), computeMean(ys)
			var sxy, sxx, syy float64
			for i := range xs {
				dx, dy := xs[i]-mx, ys[i]-my
				sxy += dx * dy
				sxx += dx * dx
				syy += dy * dy
			}
			denom := math.Sqrt(sxx * syy)
			if denom == 0 {
				return &object.Float{Value: 0}
			}
			return &object.Float{Value: sxy / denom}
		},
	}

	builtins["covariance"] = &object.Builtin{
		Name: "covariance",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return errArgs("covariance", 2, len(args))
			}
			xs, err1 := toFloatSlice(args[0])
			ys, err2 := toFloatSlice(args[1])
			if err1 != nil {
				return err1
			}
			if err2 != nil {
				return err2
			}
			if len(xs) != len(ys) {
				return &object.Error{Message: "covariance requires lists of equal length"}
			}
			if len(xs) < 2 {
				return &object.Error{Message: "covariance requires at least 2 data points"}
			}
			mx, my := computeMean(xs), computeMean(ys)
			sum := 0.0
			for i := range xs {
				sum += (xs[i] - mx) * (ys[i] - my)
			}
			return &object.Float{Value: sum / float64(len(xs)-1)}
		},
	}

	builtins["linreg"] = &object.Builtin{
		Name: "linreg",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return errArgs("linreg", 2, len(args))
			}
			xs, err1 := toFloatSlice(args[0])
			ys, err2 := toFloatSlice(args[1])
			if err1 != nil {
				return err1
			}
			if err2 != nil {
				return err2
			}
			if len(xs) != len(ys) {
				return &object.Error{Message: "linreg requires lists of equal length"}
			}
			if len(xs) < 2 {
				return &object.Error{Message: "linreg requires at least 2 data points"}
			}
			mx, my := computeMean(xs), computeMean(ys)
			var sxy, sxx float64
			for i := range xs {
				dx := xs[i] - mx
				sxy += dx * (ys[i] - my)
				sxx += dx * dx
			}
			if sxx == 0 {
				return &object.Error{Message: "linreg: all x values are identical"}
			}
			slope := sxy / sxx
			intercept := my - slope*mx
			// Return (slope, intercept, r_squared)
			var ssRes, ssTot float64
			for i := range xs {
				predicted := slope*xs[i] + intercept
				ssRes += (ys[i] - predicted) * (ys[i] - predicted)
				ssTot += (ys[i] - my) * (ys[i] - my)
			}
			rSquared := 0.0
			if ssTot != 0 {
				rSquared = 1 - ssRes/ssTot
			}
			return &object.Tuple{Elements: []object.Object{
				&object.Float{Value: slope},
				&object.Float{Value: intercept},
				&object.Float{Value: rSquared},
			}}
		},
	}
}

func computeMean(vals []float64) float64 {
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

// ────────────────────────────────────────────────────────────
// Linear Algebra
// ────────────────────────────────────────────────────────────

func addLinearAlgebraBuiltins(builtins map[string]*object.Builtin) {
	builtins["vec"] = &object.Builtin{
		Name: "vec",
		Fn: func(args ...object.Object) object.Object {
			if len(args) == 1 {
				if list, ok := args[0].(*object.List); ok {
					elems := make([]float64, len(list.Elements))
					for i, e := range list.Elements {
						v, ok := toFloat(e)
						if !ok {
							return &object.Error{Message: fmt.Sprintf("vec: element %d is not a number", i)}
						}
						elems[i] = v
					}
					return &object.Vector{Elements: elems}
				}
			}
			elems := make([]float64, len(args))
			for i, arg := range args {
				v, ok := toFloat(arg)
				if !ok {
					return &object.Error{Message: fmt.Sprintf("vec: argument %d is not a number", i)}
				}
				elems[i] = v
			}
			return &object.Vector{Elements: elems}
		},
	}

	builtins["mat"] = &object.Builtin{
		Name: "mat",
		Fn: func(args ...object.Object) object.Object {
			if len(args) == 0 {
				return &object.Error{Message: "mat expects at least 1 argument (rows as lists)"}
			}
			rows := make([][]float64, len(args))
			cols := -1
			for i, arg := range args {
				list, ok := arg.(*object.List)
				if !ok {
					return &object.Error{Message: fmt.Sprintf("mat: argument %d must be a list", i)}
				}
				row := make([]float64, len(list.Elements))
				for j, e := range list.Elements {
					v, ok := toFloat(e)
					if !ok {
						return &object.Error{Message: fmt.Sprintf("mat: element [%d][%d] is not a number", i, j)}
					}
					row[j] = v
				}
				if cols == -1 {
					cols = len(row)
				} else if len(row) != cols {
					return &object.Error{Message: "mat: all rows must have the same length"}
				}
				rows[i] = row
			}
			return &object.Matrix{Rows: len(rows), Cols: cols, Data: rows}
		},
	}

	builtins["dot"] = &object.Builtin{
		Name: "dot",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return errArgs("dot", 2, len(args))
			}
			a, ok1 := args[0].(*object.Vector)
			b, ok2 := args[1].(*object.Vector)
			if !ok1 || !ok2 {
				return &object.Error{Message: "dot expects two vectors"}
			}
			if len(a.Elements) != len(b.Elements) {
				return &object.Error{Message: "dot: vectors must have same length"}
			}
			sum := 0.0
			for i := range a.Elements {
				sum += a.Elements[i] * b.Elements[i]
			}
			return &object.Float{Value: sum}
		},
	}

	builtins["cross"] = &object.Builtin{
		Name: "cross",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return errArgs("cross", 2, len(args))
			}
			a, ok1 := args[0].(*object.Vector)
			b, ok2 := args[1].(*object.Vector)
			if !ok1 || !ok2 {
				return &object.Error{Message: "cross expects two vectors"}
			}
			if len(a.Elements) != 3 || len(b.Elements) != 3 {
				return &object.Error{Message: "cross: vectors must be 3-dimensional"}
			}
			return &object.Vector{Elements: []float64{
				a.Elements[1]*b.Elements[2] - a.Elements[2]*b.Elements[1],
				a.Elements[2]*b.Elements[0] - a.Elements[0]*b.Elements[2],
				a.Elements[0]*b.Elements[1] - a.Elements[1]*b.Elements[0],
			}}
		},
	}

	builtins["magnitude"] = &object.Builtin{
		Name: "magnitude",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("magnitude", 1, len(args))
			}
			v, ok := args[0].(*object.Vector)
			if !ok {
				return &object.Error{Message: "magnitude expects a vector"}
			}
			sum := 0.0
			for _, e := range v.Elements {
				sum += e * e
			}
			return &object.Float{Value: math.Sqrt(sum)}
		},
	}

	builtins["normalize"] = &object.Builtin{
		Name: "normalize",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("normalize", 1, len(args))
			}
			v, ok := args[0].(*object.Vector)
			if !ok {
				return &object.Error{Message: "normalize expects a vector"}
			}
			mag := 0.0
			for _, e := range v.Elements {
				mag += e * e
			}
			mag = math.Sqrt(mag)
			if mag == 0 {
				return &object.Error{Message: "normalize: cannot normalize zero vector"}
			}
			elems := make([]float64, len(v.Elements))
			for i, e := range v.Elements {
				elems[i] = e / mag
			}
			return &object.Vector{Elements: elems}
		},
	}

	builtins["vecAdd"] = vectorBinaryOp("vecAdd", func(a, b float64) float64 { return a + b })
	builtins["vecSub"] = vectorBinaryOp("vecSub", func(a, b float64) float64 { return a - b })
	builtins["vecScale"] = &object.Builtin{
		Name: "vecScale",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return errArgs("vecScale", 2, len(args))
			}
			v, ok := args[0].(*object.Vector)
			if !ok {
				return &object.Error{Message: "vecScale: first argument must be a vector"}
			}
			s, ok := toFloat(args[1])
			if !ok {
				return &object.Error{Message: "vecScale: second argument must be a number"}
			}
			elems := make([]float64, len(v.Elements))
			for i, e := range v.Elements {
				elems[i] = e * s
			}
			return &object.Vector{Elements: elems}
		},
	}

	builtins["matMul"] = &object.Builtin{
		Name: "matMul",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return errArgs("matMul", 2, len(args))
			}
			a, ok1 := args[0].(*object.Matrix)
			b, ok2 := args[1].(*object.Matrix)
			if !ok1 || !ok2 {
				return &object.Error{Message: "matMul expects two matrices"}
			}
			if a.Cols != b.Rows {
				return &object.Error{Message: fmt.Sprintf("matMul: incompatible dimensions %dx%d * %dx%d", a.Rows, a.Cols, b.Rows, b.Cols)}
			}
			result := make([][]float64, a.Rows)
			for i := range result {
				result[i] = make([]float64, b.Cols)
				for j := 0; j < b.Cols; j++ {
					sum := 0.0
					for k := 0; k < a.Cols; k++ {
						sum += a.Data[i][k] * b.Data[k][j]
					}
					result[i][j] = sum
				}
			}
			return &object.Matrix{Rows: a.Rows, Cols: b.Cols, Data: result}
		},
	}

	builtins["matVecMul"] = &object.Builtin{
		Name: "matVecMul",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return errArgs("matVecMul", 2, len(args))
			}
			m, ok1 := args[0].(*object.Matrix)
			v, ok2 := args[1].(*object.Vector)
			if !ok1 || !ok2 {
				return &object.Error{Message: "matVecMul expects a matrix and a vector"}
			}
			if m.Cols != len(v.Elements) {
				return &object.Error{Message: "matVecMul: matrix columns must match vector length"}
			}
			result := make([]float64, m.Rows)
			for i := 0; i < m.Rows; i++ {
				sum := 0.0
				for j := 0; j < m.Cols; j++ {
					sum += m.Data[i][j] * v.Elements[j]
				}
				result[i] = sum
			}
			return &object.Vector{Elements: result}
		},
	}

	builtins["transpose"] = &object.Builtin{
		Name: "transpose",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("transpose", 1, len(args))
			}
			m, ok := args[0].(*object.Matrix)
			if !ok {
				return &object.Error{Message: "transpose expects a matrix"}
			}
			result := make([][]float64, m.Cols)
			for j := 0; j < m.Cols; j++ {
				result[j] = make([]float64, m.Rows)
				for i := 0; i < m.Rows; i++ {
					result[j][i] = m.Data[i][j]
				}
			}
			return &object.Matrix{Rows: m.Cols, Cols: m.Rows, Data: result}
		},
	}

	builtins["determinant"] = &object.Builtin{
		Name: "determinant",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("determinant", 1, len(args))
			}
			m, ok := args[0].(*object.Matrix)
			if !ok {
				return &object.Error{Message: "determinant expects a matrix"}
			}
			if m.Rows != m.Cols {
				return &object.Error{Message: "determinant: matrix must be square"}
			}
			return &object.Float{Value: det(m.Data, m.Rows)}
		},
	}

	builtins["inverse"] = &object.Builtin{
		Name: "inverse",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("inverse", 1, len(args))
			}
			m, ok := args[0].(*object.Matrix)
			if !ok {
				return &object.Error{Message: "inverse expects a matrix"}
			}
			if m.Rows != m.Cols {
				return &object.Error{Message: "inverse: matrix must be square"}
			}
			inv, err := matInverse(m.Data, m.Rows)
			if err != nil {
				return &object.Error{Message: err.Error()}
			}
			return &object.Matrix{Rows: m.Rows, Cols: m.Cols, Data: inv}
		},
	}

	builtins["identity"] = &object.Builtin{
		Name: "identity",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("identity", 1, len(args))
			}
			n, ok := args[0].(*object.Integer)
			if !ok || n.Value <= 0 {
				return &object.Error{Message: "identity expects a positive integer"}
			}
			size := int(n.Value)
			data := make([][]float64, size)
			for i := range data {
				data[i] = make([]float64, size)
				data[i][i] = 1
			}
			return &object.Matrix{Rows: size, Cols: size, Data: data}
		},
	}

	builtins["zeros"] = &object.Builtin{
		Name: "zeros",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return errArgs("zeros", 2, len(args))
			}
			r, ok1 := args[0].(*object.Integer)
			c, ok2 := args[1].(*object.Integer)
			if !ok1 || !ok2 || r.Value <= 0 || c.Value <= 0 {
				return &object.Error{Message: "zeros expects two positive integers (rows, cols)"}
			}
			rows, cols := int(r.Value), int(c.Value)
			data := make([][]float64, rows)
			for i := range data {
				data[i] = make([]float64, cols)
			}
			return &object.Matrix{Rows: rows, Cols: cols, Data: data}
		},
	}
}

func vectorBinaryOp(name string, op func(float64, float64) float64) *object.Builtin {
	return &object.Builtin{
		Name: name,
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return errArgs(name, 2, len(args))
			}
			a, ok1 := args[0].(*object.Vector)
			b, ok2 := args[1].(*object.Vector)
			if !ok1 || !ok2 {
				return &object.Error{Message: fmt.Sprintf("%s expects two vectors", name)}
			}
			if len(a.Elements) != len(b.Elements) {
				return &object.Error{Message: fmt.Sprintf("%s: vectors must have same length", name)}
			}
			elems := make([]float64, len(a.Elements))
			for i := range a.Elements {
				elems[i] = op(a.Elements[i], b.Elements[i])
			}
			return &object.Vector{Elements: elems}
		},
	}
}

// det computes the determinant using cofactor expansion (for small matrices)
// and LU decomposition for larger ones.
func det(data [][]float64, n int) float64 {
	if n == 1 {
		return data[0][0]
	}
	if n == 2 {
		return data[0][0]*data[1][1] - data[0][1]*data[1][0]
	}
	if n == 3 {
		return data[0][0]*(data[1][1]*data[2][2]-data[1][2]*data[2][1]) -
			data[0][1]*(data[1][0]*data[2][2]-data[1][2]*data[2][0]) +
			data[0][2]*(data[1][0]*data[2][1]-data[1][1]*data[2][0])
	}
	// LU decomposition for larger matrices
	lu := make([][]float64, n)
	for i := range lu {
		lu[i] = make([]float64, n)
		copy(lu[i], data[i])
	}
	sign := 1.0
	for i := 0; i < n; i++ {
		// Find pivot
		maxVal := math.Abs(lu[i][i])
		maxRow := i
		for k := i + 1; k < n; k++ {
			if math.Abs(lu[k][i]) > maxVal {
				maxVal = math.Abs(lu[k][i])
				maxRow = k
			}
		}
		if maxVal < 1e-12 {
			return 0
		}
		if maxRow != i {
			lu[i], lu[maxRow] = lu[maxRow], lu[i]
			sign *= -1
		}
		for k := i + 1; k < n; k++ {
			factor := lu[k][i] / lu[i][i]
			for j := i; j < n; j++ {
				lu[k][j] -= factor * lu[i][j]
			}
		}
	}
	result := sign
	for i := 0; i < n; i++ {
		result *= lu[i][i]
	}
	return result
}

// matInverse computes the inverse using Gauss-Jordan elimination.
func matInverse(data [][]float64, n int) ([][]float64, error) {
	aug := make([][]float64, n)
	for i := range aug {
		aug[i] = make([]float64, 2*n)
		copy(aug[i], data[i])
		aug[i][n+i] = 1
	}
	for i := 0; i < n; i++ {
		// Find pivot
		maxVal := math.Abs(aug[i][i])
		maxRow := i
		for k := i + 1; k < n; k++ {
			if math.Abs(aug[k][i]) > maxVal {
				maxVal = math.Abs(aug[k][i])
				maxRow = k
			}
		}
		if maxVal < 1e-12 {
			return nil, fmt.Errorf("inverse: matrix is singular")
		}
		if maxRow != i {
			aug[i], aug[maxRow] = aug[maxRow], aug[i]
		}
		pivot := aug[i][i]
		for j := 0; j < 2*n; j++ {
			aug[i][j] /= pivot
		}
		for k := 0; k < n; k++ {
			if k == i {
				continue
			}
			factor := aug[k][i]
			for j := 0; j < 2*n; j++ {
				aug[k][j] -= factor * aug[i][j]
			}
		}
	}
	result := make([][]float64, n)
	for i := range result {
		result[i] = make([]float64, n)
		copy(result[i], aug[i][n:])
	}
	return result, nil
}

// ────────────────────────────────────────────────────────────
// Complex Numbers
// ────────────────────────────────────────────────────────────

func addComplexBuiltins(builtins map[string]*object.Builtin) {
	builtins["complex"] = &object.Builtin{
		Name: "complex",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return errArgs("complex", 2, len(args))
			}
			r, ok1 := toFloat(args[0])
			i, ok2 := toFloat(args[1])
			if !ok1 || !ok2 {
				return errNumeric("complex")
			}
			return &object.Complex{Real: r, Imag: i}
		},
	}

	builtins["real"] = &object.Builtin{
		Name: "real",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("real", 1, len(args))
			}
			c, ok := args[0].(*object.Complex)
			if !ok {
				return &object.Error{Message: "real expects a complex number"}
			}
			return &object.Float{Value: c.Real}
		},
	}

	builtins["imag"] = &object.Builtin{
		Name: "imag",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("imag", 1, len(args))
			}
			c, ok := args[0].(*object.Complex)
			if !ok {
				return &object.Error{Message: "imag expects a complex number"}
			}
			return &object.Float{Value: c.Imag}
		},
	}

	builtins["conjugate"] = &object.Builtin{
		Name: "conjugate",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("conjugate", 1, len(args))
			}
			c, ok := args[0].(*object.Complex)
			if !ok {
				return &object.Error{Message: "conjugate expects a complex number"}
			}
			return &object.Complex{Real: c.Real, Imag: -c.Imag}
		},
	}

	builtins["complexMag"] = &object.Builtin{
		Name: "complexMag",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("complexMag", 1, len(args))
			}
			c, ok := args[0].(*object.Complex)
			if !ok {
				return &object.Error{Message: "complexMag expects a complex number"}
			}
			return &object.Float{Value: cmplx.Abs(complex(c.Real, c.Imag))}
		},
	}

	builtins["complexPhase"] = &object.Builtin{
		Name: "complexPhase",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("complexPhase", 1, len(args))
			}
			c, ok := args[0].(*object.Complex)
			if !ok {
				return &object.Error{Message: "complexPhase expects a complex number"}
			}
			return &object.Float{Value: cmplx.Phase(complex(c.Real, c.Imag))}
		},
	}

	builtins["complexAdd"] = complexBinaryOp("complexAdd", func(a, b complex128) complex128 { return a + b })
	builtins["complexSub"] = complexBinaryOp("complexSub", func(a, b complex128) complex128 { return a - b })
	builtins["complexMul"] = complexBinaryOp("complexMul", func(a, b complex128) complex128 { return a * b })
	builtins["complexDiv"] = complexBinaryOp("complexDiv", func(a, b complex128) complex128 { return a / b })

	builtins["complexSqrt"] = &object.Builtin{
		Name: "complexSqrt",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("complexSqrt", 1, len(args))
			}
			c, ok := args[0].(*object.Complex)
			if !ok {
				return &object.Error{Message: "complexSqrt expects a complex number"}
			}
			result := cmplx.Sqrt(complex(c.Real, c.Imag))
			return &object.Complex{Real: real(result), Imag: imag(result)}
		},
	}

	builtins["complexExp"] = &object.Builtin{
		Name: "complexExp",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return errArgs("complexExp", 1, len(args))
			}
			c, ok := args[0].(*object.Complex)
			if !ok {
				return &object.Error{Message: "complexExp expects a complex number"}
			}
			result := cmplx.Exp(complex(c.Real, c.Imag))
			return &object.Complex{Real: real(result), Imag: imag(result)}
		},
	}
}

func complexBinaryOp(name string, op func(complex128, complex128) complex128) *object.Builtin {
	return &object.Builtin{
		Name: name,
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return errArgs(name, 2, len(args))
			}
			a, ok1 := args[0].(*object.Complex)
			b, ok2 := args[1].(*object.Complex)
			if !ok1 || !ok2 {
				return &object.Error{Message: fmt.Sprintf("%s expects two complex numbers", name)}
			}
			result := op(complex(a.Real, a.Imag), complex(b.Real, b.Imag))
			return &object.Complex{Real: real(result), Imag: imag(result)}
		},
	}
}

// ────────────────────────────────────────────────────────────
// Numerical Methods
// ────────────────────────────────────────────────────────────

func addNumericalBuiltins(builtins map[string]*object.Builtin) {
	// These are placeholder builtins that work with GeoFlow function objects.
	// They need access to the evaluator, so they use a callback pattern.
	// The actual implementations are registered by the evaluator at startup.

	builtins["linspace"] = &object.Builtin{
		Name: "linspace",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return errArgs("linspace", 3, len(args))
			}
			start, ok1 := toFloat(args[0])
			end, ok2 := toFloat(args[1])
			n, ok3 := args[2].(*object.Integer)
			if !ok1 || !ok2 || !ok3 || n.Value < 2 {
				return &object.Error{Message: "linspace expects (start, end, n) where n >= 2"}
			}
			count := int(n.Value)
			step := (end - start) / float64(count-1)
			elems := make([]object.Object, count)
			for i := 0; i < count; i++ {
				elems[i] = &object.Float{Value: start + float64(i)*step}
			}
			return &object.List{Elements: elems}
		},
	}

	builtins["arange"] = &object.Builtin{
		Name: "arange",
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 2 || len(args) > 3 {
				return &object.Error{Message: "arange expects 2-3 arguments: start, end[, step]"}
			}
			start, ok1 := toFloat(args[0])
			end, ok2 := toFloat(args[1])
			if !ok1 || !ok2 {
				return errNumeric("arange")
			}
			step := 1.0
			if len(args) == 3 {
				s, ok := toFloat(args[2])
				if !ok {
					return errNumeric("arange")
				}
				step = s
			}
			if step == 0 {
				return &object.Error{Message: "arange: step cannot be 0"}
			}
			var elems []object.Object
			if step > 0 {
				for v := start; v < end; v += step {
					elems = append(elems, &object.Float{Value: v})
				}
			} else {
				for v := start; v > end; v += step {
					elems = append(elems, &object.Float{Value: v})
				}
			}
			return &object.List{Elements: elems}
		},
	}
}
