// Package numeric provides numerical methods for the GeoFlow standard library.
package numeric

import (
	"fmt"
	"math"

	"github.com/rogue780/geoflow/internal/object"
)

// GetExports returns all exported functions for the std.math.numeric module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		"bisect":          &object.Builtin{Name: "num.bisect", Fn: bisect},
		"newton":          &object.Builtin{Name: "num.newton", Fn: newton},
		"quad":            &object.Builtin{Name: "num.quad", Fn: quad},
		"diff":            &object.Builtin{Name: "num.diff", Fn: numDiff},
		"polyfit":         &object.Builtin{Name: "num.polyfit", Fn: polyfit},
		"polyval":         &object.Builtin{Name: "num.polyval", Fn: polyval},
		"interp1d":        &object.Builtin{Name: "num.interp1d", Fn: interp1d},
		"minimize":        &object.Builtin{Name: "num.minimize", Fn: minimize},
		"gradientDescent": &object.Builtin{Name: "num.gradientDescent", Fn: gradientDescent},
		"linspace":        &object.Builtin{Name: "num.linspace", Fn: linspace},
		"arange":          &object.Builtin{Name: "num.arange", Fn: arange},
	}
}

// callFn invokes a callable object with the given arguments.
// Uses object.CallFunction which supports Builtin, Function, and ComposedFunction.
func callFn(fn object.Object, args ...object.Object) object.Object {
	return object.CallFunction(fn, args...)
}

// callFnFloat64 calls fn with a single float64 argument and returns the float64 result.
func callFnFloat64(fn object.Object, x float64) (float64, error) {
	result := callFn(fn, &object.Float{Value: x})
	if err, ok := result.(*object.Error); ok {
		return 0, fmt.Errorf("%s", err.Message)
	}
	switch v := result.(type) {
	case *object.Float:
		return v.Value, nil
	case *object.Integer:
		return float64(v.Value), nil
	default:
		return 0, fmt.Errorf("function returned %s, expected numeric", result.Type())
	}
}

// callFnMultiFloat64 calls fn with multiple float64 arguments and returns the float64 result.
func callFnMultiFloat64(fn object.Object, xs []float64) (float64, error) {
	args := make([]object.Object, len(xs))
	for i, x := range xs {
		args[i] = &object.Float{Value: x}
	}
	result := callFn(fn, args...)
	if err, ok := result.(*object.Error); ok {
		return 0, fmt.Errorf("%s", err.Message)
	}
	switch v := result.(type) {
	case *object.Float:
		return v.Value, nil
	case *object.Integer:
		return float64(v.Value), nil
	default:
		return 0, fmt.Errorf("function returned %s, expected numeric", result.Type())
	}
}

// ── Root Finding: Bisection ──

// bisect implements the bisection method for root finding.
// num.bisect(f, a, b, tol?) -> float
func bisect(args ...object.Object) object.Object {
	if len(args) < 3 || len(args) > 4 {
		return &object.Error{Message: "num.bisect expects 3-4 arguments (f, a, b[, tol])"}
	}

	fn := args[0]
	a, err := getFloat(args[1], "a")
	if err != nil {
		return &object.Error{Message: err.Error()}
	}
	b, err := getFloat(args[2], "b")
	if err != nil {
		return &object.Error{Message: err.Error()}
	}

	tol := 1e-12
	if len(args) == 4 {
		t, err := getFloat(args[3], "tol")
		if err != nil {
			return &object.Error{Message: err.Error()}
		}
		tol = t
	}

	fa, err := callFnFloat64(fn, a)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("num.bisect: %s", err.Error())}
	}
	fb, err := callFnFloat64(fn, b)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("num.bisect: %s", err.Error())}
	}

	if fa*fb > 0 {
		return &object.Error{Message: "num.bisect: f(a) and f(b) must have opposite signs"}
	}

	maxIter := 1000
	for i := 0; i < maxIter; i++ {
		mid := (a + b) / 2.0
		fm, err := callFnFloat64(fn, mid)
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("num.bisect: %s", err.Error())}
		}

		if math.Abs(fm) < tol || (b-a)/2.0 < tol {
			return &object.Float{Value: mid}
		}

		if fa*fm < 0 {
			b = mid
		} else {
			a = mid
			fa = fm
		}
	}

	return &object.Float{Value: (a + b) / 2.0}
}

// ── Root Finding: Newton's Method ──

// newton implements Newton's method using numerical derivative.
// num.newton(f, x0, tol?) -> float
func newton(args ...object.Object) object.Object {
	if len(args) < 2 || len(args) > 3 {
		return &object.Error{Message: "num.newton expects 2-3 arguments (f, x0[, tol])"}
	}

	fn := args[0]
	x, err := getFloat(args[1], "x0")
	if err != nil {
		return &object.Error{Message: err.Error()}
	}

	tol := 1e-12
	if len(args) == 3 {
		t, err := getFloat(args[2], "tol")
		if err != nil {
			return &object.Error{Message: err.Error()}
		}
		tol = t
	}

	h := 1e-8
	maxIter := 1000

	for i := 0; i < maxIter; i++ {
		fx, err := callFnFloat64(fn, x)
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("num.newton: %s", err.Error())}
		}

		if math.Abs(fx) < tol {
			return &object.Float{Value: x}
		}

		// Central difference for derivative
		fxph, err := callFnFloat64(fn, x+h)
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("num.newton: %s", err.Error())}
		}
		fxmh, err := callFnFloat64(fn, x-h)
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("num.newton: %s", err.Error())}
		}

		dfx := (fxph - fxmh) / (2 * h)
		if dfx == 0 {
			return &object.Error{Message: "num.newton: derivative is zero, cannot continue"}
		}

		x = x - fx/dfx
	}

	return &object.Float{Value: x}
}

// ── Numerical Integration: Simpson's Rule ──

// quad performs numerical integration using composite Simpson's 1/3 rule.
// num.quad(f, a, b) -> Tuple<float, float>
func quad(args ...object.Object) object.Object {
	if len(args) < 3 || len(args) > 4 {
		return &object.Error{Message: "num.quad expects 3-4 arguments (f, a, b[, n])"}
	}

	fn := args[0]
	a, err := getFloat(args[1], "a")
	if err != nil {
		return &object.Error{Message: err.Error()}
	}
	b, err := getFloat(args[2], "b")
	if err != nil {
		return &object.Error{Message: err.Error()}
	}

	n := 1000 // number of subintervals (must be even)
	if len(args) == 4 {
		nVal, err := getFloat(args[3], "n")
		if err != nil {
			return &object.Error{Message: err.Error()}
		}
		n = int(nVal)
		if n < 2 {
			n = 2
		}
	}
	// Ensure n is even
	if n%2 != 0 {
		n++
	}

	h := (b - a) / float64(n)

	fa, err := callFnFloat64(fn, a)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("num.quad: %s", err.Error())}
	}
	fb, err := callFnFloat64(fn, b)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("num.quad: %s", err.Error())}
	}

	sum := fa + fb

	for i := 1; i < n; i++ {
		x := a + float64(i)*h
		fx, err := callFnFloat64(fn, x)
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("num.quad: %s", err.Error())}
		}
		if i%2 == 0 {
			sum += 2 * fx
		} else {
			sum += 4 * fx
		}
	}

	result := sum * h / 3.0

	// Estimate error using Richardson extrapolation with n/2
	nHalf := n / 2
	if nHalf%2 != 0 {
		nHalf++
	}
	hHalf := (b - a) / float64(nHalf)
	sumHalf := fa + fb
	for i := 1; i < nHalf; i++ {
		x := a + float64(i)*hHalf
		fx, err := callFnFloat64(fn, x)
		if err != nil {
			// If error estimation fails, return result with 0 error estimate
			return &object.Tuple{Elements: []object.Object{
				&object.Float{Value: result},
				&object.Float{Value: 0},
			}}
		}
		if i%2 == 0 {
			sumHalf += 2 * fx
		} else {
			sumHalf += 4 * fx
		}
	}
	resultHalf := sumHalf * hHalf / 3.0
	errorEstimate := math.Abs(result - resultHalf) / 15.0

	return &object.Tuple{Elements: []object.Object{
		&object.Float{Value: result},
		&object.Float{Value: errorEstimate},
	}}
}

// ── Numerical Differentiation ──

// numDiff computes the numerical derivative using central difference.
// num.diff(f, x, h?) -> float
func numDiff(args ...object.Object) object.Object {
	if len(args) < 2 || len(args) > 3 {
		return &object.Error{Message: "num.diff expects 2-3 arguments (f, x[, h])"}
	}

	fn := args[0]
	x, err := getFloat(args[1], "x")
	if err != nil {
		return &object.Error{Message: err.Error()}
	}

	h := 1e-8
	if len(args) == 3 {
		hVal, err := getFloat(args[2], "h")
		if err != nil {
			return &object.Error{Message: err.Error()}
		}
		h = hVal
	}

	// Central difference: (f(x+h) - f(x-h)) / (2h)
	fxph, err := callFnFloat64(fn, x+h)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("num.diff: %s", err.Error())}
	}
	fxmh, err := callFnFloat64(fn, x-h)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("num.diff: %s", err.Error())}
	}

	return &object.Float{Value: (fxph - fxmh) / (2 * h)}
}

// ── Polynomial Fitting ──

// polyfit fits a polynomial of given degree to (xs, ys) data using normal equations.
// num.polyfit(xs, ys, degree) -> List<float>
func polyfit(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "num.polyfit expects 3 arguments (xs, ys, degree)"}
	}

	xsList, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "num.polyfit: first argument must be a list of x values"}
	}
	ysList, ok := args[1].(*object.List)
	if !ok {
		return &object.Error{Message: "num.polyfit: second argument must be a list of y values"}
	}
	degObj, err := getFloat(args[2], "degree")
	if err != nil {
		return &object.Error{Message: err.Error()}
	}
	degree := int(degObj)

	if len(xsList.Elements) != len(ysList.Elements) {
		return &object.Error{Message: "num.polyfit: xs and ys must have the same length"}
	}
	n := len(xsList.Elements)
	if n < degree+1 {
		return &object.Error{Message: "num.polyfit: not enough data points for the requested degree"}
	}

	xs := make([]float64, n)
	ys := make([]float64, n)
	for i := 0; i < n; i++ {
		x, err := getFloat(xsList.Elements[i], "x")
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("num.polyfit: xs[%d] %s", i, err.Error())}
		}
		y, err := getFloat(ysList.Elements[i], "y")
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("num.polyfit: ys[%d] %s", i, err.Error())}
		}
		xs[i] = x
		ys[i] = y
	}

	m := degree + 1

	// Build Vandermonde matrix X (n x m) where X[i][j] = xs[i]^j
	// Normal equations: (X^T * X) * coeffs = X^T * y

	// Compute X^T * X (m x m)
	xtx := make([][]float64, m)
	for i := 0; i < m; i++ {
		xtx[i] = make([]float64, m)
		for j := 0; j < m; j++ {
			sum := 0.0
			for k := 0; k < n; k++ {
				sum += math.Pow(xs[k], float64(i)) * math.Pow(xs[k], float64(j))
			}
			xtx[i][j] = sum
		}
	}

	// Compute X^T * y (m x 1)
	xty := make([]float64, m)
	for i := 0; i < m; i++ {
		sum := 0.0
		for k := 0; k < n; k++ {
			sum += math.Pow(xs[k], float64(i)) * ys[k]
		}
		xty[i] = sum
	}

	// Solve using Gaussian elimination with partial pivoting
	coeffs, solveErr := solveLinearSystem(xtx, xty)
	if solveErr != nil {
		return &object.Error{Message: fmt.Sprintf("num.polyfit: %s", solveErr.Error())}
	}

	// Return coefficients from highest degree to lowest (standard convention)
	result := make([]object.Object, m)
	for i := 0; i < m; i++ {
		result[i] = &object.Float{Value: coeffs[m-1-i]}
	}
	return &object.List{Elements: result}
}

// solveLinearSystem solves Ax = b using Gaussian elimination with partial pivoting.
func solveLinearSystem(A [][]float64, b []float64) ([]float64, error) {
	n := len(b)

	// Augmented matrix
	aug := make([][]float64, n)
	for i := 0; i < n; i++ {
		aug[i] = make([]float64, n+1)
		copy(aug[i], A[i])
		aug[i][n] = b[i]
	}

	// Forward elimination
	for col := 0; col < n; col++ {
		// Find pivot
		maxRow := col
		maxVal := math.Abs(aug[col][col])
		for row := col + 1; row < n; row++ {
			if math.Abs(aug[row][col]) > maxVal {
				maxVal = math.Abs(aug[row][col])
				maxRow = row
			}
		}

		if maxVal < 1e-15 {
			return nil, fmt.Errorf("singular matrix")
		}

		// Swap rows
		aug[col], aug[maxRow] = aug[maxRow], aug[col]

		// Eliminate
		for row := col + 1; row < n; row++ {
			factor := aug[row][col] / aug[col][col]
			for j := col; j <= n; j++ {
				aug[row][j] -= factor * aug[col][j]
			}
		}
	}

	// Back substitution
	x := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		x[i] = aug[i][n]
		for j := i + 1; j < n; j++ {
			x[i] -= aug[i][j] * x[j]
		}
		x[i] /= aug[i][i]
	}

	return x, nil
}

// ── Polynomial Evaluation ──

// polyval evaluates a polynomial at a given point.
// num.polyval(coeffs, x) -> float
// coeffs are from highest degree to lowest: [a_n, a_{n-1}, ..., a_1, a_0]
func polyval(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "num.polyval expects 2 arguments (coeffs, x)"}
	}

	coeffsList, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "num.polyval: first argument must be a list of coefficients"}
	}
	x, err := getFloat(args[1], "x")
	if err != nil {
		return &object.Error{Message: err.Error()}
	}

	coeffs := make([]float64, len(coeffsList.Elements))
	for i, c := range coeffsList.Elements {
		v, err := getFloat(c, "coeff")
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("num.polyval: coeffs[%d] %s", i, err.Error())}
		}
		coeffs[i] = v
	}

	// Horner's method: a_n * x^n + ... + a_0
	result := 0.0
	for _, c := range coeffs {
		result = result*x + c
	}

	return &object.Float{Value: result}
}

// ── Linear Interpolation ──

// interp1d creates a linear interpolation function.
// num.interp1d(xs, ys) -> Builtin function
func interp1d(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "num.interp1d expects 2 arguments (xs, ys)"}
	}

	xsList, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "num.interp1d: first argument must be a list of x values"}
	}
	ysList, ok := args[1].(*object.List)
	if !ok {
		return &object.Error{Message: "num.interp1d: second argument must be a list of y values"}
	}

	if len(xsList.Elements) != len(ysList.Elements) {
		return &object.Error{Message: "num.interp1d: xs and ys must have the same length"}
	}
	if len(xsList.Elements) < 2 {
		return &object.Error{Message: "num.interp1d: need at least 2 data points"}
	}

	n := len(xsList.Elements)
	xs := make([]float64, n)
	ys := make([]float64, n)
	for i := 0; i < n; i++ {
		x, err := getFloat(xsList.Elements[i], "x")
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("num.interp1d: xs[%d] %s", i, err.Error())}
		}
		y, err := getFloat(ysList.Elements[i], "y")
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("num.interp1d: ys[%d] %s", i, err.Error())}
		}
		xs[i] = x
		ys[i] = y
	}

	return &object.Builtin{
		Name: "interp1d",
		Fn: func(innerArgs ...object.Object) object.Object {
			if len(innerArgs) != 1 {
				return &object.Error{Message: "interp1d function expects 1 argument (x)"}
			}
			x, err := getFloat(innerArgs[0], "x")
			if err != nil {
				return &object.Error{Message: err.Error()}
			}

			// Find the interval
			if x <= xs[0] {
				return &object.Float{Value: ys[0]}
			}
			if x >= xs[n-1] {
				return &object.Float{Value: ys[n-1]}
			}

			for i := 0; i < n-1; i++ {
				if x >= xs[i] && x <= xs[i+1] {
					// Linear interpolation
					t := (x - xs[i]) / (xs[i+1] - xs[i])
					return &object.Float{Value: ys[i] + t*(ys[i+1]-ys[i])}
				}
			}

			return &object.Float{Value: ys[n-1]}
		},
	}
}

// ── Minimization: Golden Section Search ──

// minimize uses golden section search to find minimum of f on [a, b].
// num.minimize(f, bounds) -> Struct{x, fx, iterations, converged}
func minimize(args ...object.Object) object.Object {
	if len(args) < 2 || len(args) > 3 {
		return &object.Error{Message: "num.minimize expects 2-3 arguments (f, bounds[, tol])"}
	}

	fn := args[0]

	// bounds can be a tuple or list [a, b]
	var a, b float64
	switch bounds := args[1].(type) {
	case *object.Tuple:
		if len(bounds.Elements) != 2 {
			return &object.Error{Message: "num.minimize: bounds must be a tuple/list of 2 elements"}
		}
		var err error
		a, err = getFloat(bounds.Elements[0], "a")
		if err != nil {
			return &object.Error{Message: err.Error()}
		}
		b, err = getFloat(bounds.Elements[1], "b")
		if err != nil {
			return &object.Error{Message: err.Error()}
		}
	case *object.List:
		if len(bounds.Elements) != 2 {
			return &object.Error{Message: "num.minimize: bounds must be a tuple/list of 2 elements"}
		}
		var err error
		a, err = getFloat(bounds.Elements[0], "a")
		if err != nil {
			return &object.Error{Message: err.Error()}
		}
		b, err = getFloat(bounds.Elements[1], "b")
		if err != nil {
			return &object.Error{Message: err.Error()}
		}
	default:
		return &object.Error{Message: "num.minimize: second argument must be a tuple/list [a, b]"}
	}

	tol := 1e-12
	if len(args) == 3 {
		t, err := getFloat(args[2], "tol")
		if err != nil {
			return &object.Error{Message: err.Error()}
		}
		tol = t
	}

	gr := (math.Sqrt(5) + 1) / 2.0 // golden ratio

	c := b - (b-a)/gr
	d := a + (b-a)/gr

	maxIter := 1000
	converged := false
	iterations := 0

	for i := 0; i < maxIter; i++ {
		iterations = i + 1

		fc, err := callFnFloat64(fn, c)
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("num.minimize: %s", err.Error())}
		}
		fd, err := callFnFloat64(fn, d)
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("num.minimize: %s", err.Error())}
		}

		if fc < fd {
			b = d
		} else {
			a = c
		}

		if math.Abs(b-a) < tol {
			converged = true
			break
		}

		c = b - (b-a)/gr
		d = a + (b-a)/gr
	}

	x := (a + b) / 2.0
	fx, err := callFnFloat64(fn, x)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("num.minimize: %s", err.Error())}
	}

	return &object.Struct{
		TypeName: "MinimizeResult",
		Fields: map[string]object.Object{
			"x":          &object.Float{Value: x},
			"fx":         &object.Float{Value: fx},
			"iterations": &object.Integer{Value: int64(iterations)},
			"converged":  object.NativeBoolToBooleanObject(converged),
		},
		Order: []string{"x", "fx", "iterations", "converged"},
	}
}

// ── Gradient Descent ──

// gradientDescent performs gradient descent optimization on a multivariate function.
// num.gradientDescent(f, x0, lr?, maxIter?) -> List<float>
func gradientDescent(args ...object.Object) object.Object {
	if len(args) < 2 || len(args) > 4 {
		return &object.Error{Message: "num.gradientDescent expects 2-4 arguments (f, x0[, lr, maxIter])"}
	}

	fn := args[0]

	// x0 can be a list of floats or a single float
	var x []float64
	switch v := args[1].(type) {
	case *object.List:
		x = make([]float64, len(v.Elements))
		for i, e := range v.Elements {
			val, err := getFloat(e, "x0")
			if err != nil {
				return &object.Error{Message: fmt.Sprintf("num.gradientDescent: x0[%d] %s", i, err.Error())}
			}
			x[i] = val
		}
	case *object.Float:
		x = []float64{v.Value}
	case *object.Integer:
		x = []float64{float64(v.Value)}
	default:
		return &object.Error{Message: "num.gradientDescent: x0 must be a list or number"}
	}

	lr := 0.01
	if len(args) >= 3 {
		l, err := getFloat(args[2], "lr")
		if err != nil {
			return &object.Error{Message: err.Error()}
		}
		lr = l
	}

	maxIter := 1000
	if len(args) >= 4 {
		m, err := getFloat(args[3], "maxIter")
		if err != nil {
			return &object.Error{Message: err.Error()}
		}
		maxIter = int(m)
	}

	h := 1e-8
	n := len(x)

	for iter := 0; iter < maxIter; iter++ {
		// Compute gradient using central differences
		grad := make([]float64, n)
		for i := 0; i < n; i++ {
			xPlus := make([]float64, n)
			xMinus := make([]float64, n)
			copy(xPlus, x)
			copy(xMinus, x)
			xPlus[i] += h
			xMinus[i] -= h

			fPlus, err := callFnMultiFloat64(fn, xPlus)
			if err != nil {
				return &object.Error{Message: fmt.Sprintf("num.gradientDescent: %s", err.Error())}
			}
			fMinus, err := callFnMultiFloat64(fn, xMinus)
			if err != nil {
				return &object.Error{Message: fmt.Sprintf("num.gradientDescent: %s", err.Error())}
			}
			grad[i] = (fPlus - fMinus) / (2 * h)
		}

		// Update x
		gradNorm := 0.0
		for i := 0; i < n; i++ {
			x[i] -= lr * grad[i]
			gradNorm += grad[i] * grad[i]
		}

		// Convergence check
		if math.Sqrt(gradNorm) < 1e-10 {
			break
		}
	}

	result := make([]object.Object, n)
	for i, v := range x {
		result[i] = &object.Float{Value: v}
	}
	return &object.List{Elements: result}
}

// ── Linspace / Arange ──

// linspace creates n evenly spaced points between start and end (inclusive).
func linspace(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "num.linspace expects 3 arguments (start, end, n)"}
	}

	start, err := getFloat(args[0], "start")
	if err != nil {
		return &object.Error{Message: err.Error()}
	}
	end, err := getFloat(args[1], "end")
	if err != nil {
		return &object.Error{Message: err.Error()}
	}
	nVal, err := getFloat(args[2], "n")
	if err != nil {
		return &object.Error{Message: err.Error()}
	}
	n := int(nVal)
	if n < 1 {
		return &object.Error{Message: "num.linspace: n must be >= 1"}
	}

	elems := make([]object.Object, n)
	if n == 1 {
		elems[0] = &object.Float{Value: start}
	} else {
		step := (end - start) / float64(n-1)
		for i := 0; i < n; i++ {
			elems[i] = &object.Float{Value: start + float64(i)*step}
		}
	}

	return &object.List{Elements: elems}
}

// arange creates a list of floats from start to end (exclusive) with given step.
func arange(args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 3 {
		return &object.Error{Message: "num.arange expects 1-3 arguments ([start,] end[, step])"}
	}

	var start, end, step float64
	step = 1.0

	switch len(args) {
	case 1:
		e, err := getFloat(args[0], "end")
		if err != nil {
			return &object.Error{Message: err.Error()}
		}
		end = e
	case 2:
		s, err := getFloat(args[0], "start")
		if err != nil {
			return &object.Error{Message: err.Error()}
		}
		e, err := getFloat(args[1], "end")
		if err != nil {
			return &object.Error{Message: err.Error()}
		}
		start = s
		end = e
	case 3:
		s, err := getFloat(args[0], "start")
		if err != nil {
			return &object.Error{Message: err.Error()}
		}
		e, err := getFloat(args[1], "end")
		if err != nil {
			return &object.Error{Message: err.Error()}
		}
		st, err := getFloat(args[2], "step")
		if err != nil {
			return &object.Error{Message: err.Error()}
		}
		start = s
		end = e
		step = st
	}

	if step == 0 {
		return &object.Error{Message: "num.arange: step cannot be 0"}
	}

	var elems []object.Object
	if step > 0 {
		for x := start; x < end; x += step {
			elems = append(elems, &object.Float{Value: x})
		}
	} else {
		for x := start; x > end; x += step {
			elems = append(elems, &object.Float{Value: x})
		}
	}

	return &object.List{Elements: elems}
}

// ── Helpers ──

func getFloat(obj object.Object, name string) (float64, error) {
	switch v := obj.(type) {
	case *object.Float:
		return v.Value, nil
	case *object.Integer:
		return float64(v.Value), nil
	default:
		return 0, fmt.Errorf("num: '%s' must be numeric, got %s", name, obj.Type())
	}
}
