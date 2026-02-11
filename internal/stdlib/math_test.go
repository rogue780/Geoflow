package stdlib

import (
	"math"
	"testing"

	"github.com/rogue780/geoflow/internal/eval"
	"github.com/rogue780/geoflow/internal/lexer"
	"github.com/rogue780/geoflow/internal/object"
	"github.com/rogue780/geoflow/internal/parser"
)

// testEval creates an environment with all builtins and evaluates source code.
func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()
	for name, b := range object.GetBuiltins() {
		env.Set(name, b, false)
	}
	for name, b := range GetMathBuiltins() {
		env.Set(name, b, false)
	}
	return eval.Eval(program, env)
}

func assertFloat(t *testing.T, name string, result object.Object, expected float64) {
	t.Helper()
	f, ok := result.(*object.Float)
	if !ok {
		t.Errorf("%s: expected Float, got %T (%s)", name, result, result.Inspect())
		return
	}
	if math.IsNaN(expected) {
		if !math.IsNaN(f.Value) {
			t.Errorf("%s: expected NaN, got %g", name, f.Value)
		}
		return
	}
	if math.Abs(f.Value-expected) > 1e-9 {
		t.Errorf("%s: expected %g, got %g", name, expected, f.Value)
	}
}

func assertInt(t *testing.T, name string, result object.Object, expected int64) {
	t.Helper()
	i, ok := result.(*object.Integer)
	if !ok {
		t.Errorf("%s: expected Integer, got %T (%s)", name, result, result.Inspect())
		return
	}
	if i.Value != expected {
		t.Errorf("%s: expected %d, got %d", name, expected, i.Value)
	}
}

func assertBool(t *testing.T, name string, result object.Object, expected bool) {
	t.Helper()
	b, ok := result.(*object.Boolean)
	if !ok {
		t.Errorf("%s: expected Boolean, got %T (%s)", name, result, result.Inspect())
		return
	}
	if b.Value != expected {
		t.Errorf("%s: expected %t, got %t", name, expected, b.Value)
	}
}

// ── Core Math ──

func TestMathConstants(t *testing.T) {
	assertFloat(t, "PI", testEval("PI()"), math.Pi)
	assertFloat(t, "E", testEval("E()"), math.E)
	assertFloat(t, "TAU", testEval("TAU()"), 2*math.Pi)
}

func TestRounding(t *testing.T) {
	assertFloat(t, "floor(3.7)", testEval("floor(3.7)"), 3.0)
	assertFloat(t, "floor(-2.3)", testEval("floor(-2.3)"), -3.0)
	assertFloat(t, "ceil(3.2)", testEval("ceil(3.2)"), 4.0)
	assertFloat(t, "ceil(-2.7)", testEval("ceil(-2.7)"), -2.0)
	assertFloat(t, "round(3.5)", testEval("round(3.5)"), 4.0)
	assertFloat(t, "round(3.4)", testEval("round(3.4)"), 3.0)
	assertFloat(t, "trunc(3.7)", testEval("trunc(3.7)"), 3.0)
	assertFloat(t, "trunc(-3.7)", testEval("trunc(-3.7)"), -3.0)
}

func TestTrigonometry(t *testing.T) {
	assertFloat(t, "sin(0)", testEval("sin(0)"), 0.0)
	assertFloat(t, "cos(0)", testEval("cos(0)"), 1.0)
	assertFloat(t, "tan(0)", testEval("tan(0)"), 0.0)
	assertFloat(t, "sin(PI()/2)", testEval("sin(PI() / 2)"), 1.0)
	assertFloat(t, "asin(1)", testEval("asin(1)"), math.Pi/2)
	assertFloat(t, "acos(1)", testEval("acos(1)"), 0.0)
	assertFloat(t, "atan(1)", testEval("atan(1)"), math.Pi/4)
	assertFloat(t, "atan2(1,1)", testEval("atan2(1, 1)"), math.Pi/4)
	assertFloat(t, "sinh(0)", testEval("sinh(0)"), 0.0)
	assertFloat(t, "cosh(0)", testEval("cosh(0)"), 1.0)
	assertFloat(t, "tanh(0)", testEval("tanh(0)"), 0.0)
}

func TestConversion(t *testing.T) {
	assertFloat(t, "toRadians(180)", testEval("toRadians(180)"), math.Pi)
	assertFloat(t, "toDegrees(PI())", testEval("toDegrees(PI())"), 180.0)
}

func TestExponentialLog(t *testing.T) {
	assertFloat(t, "exp(0)", testEval("exp(0)"), 1.0)
	assertFloat(t, "exp(1)", testEval("exp(1)"), math.E)
	assertFloat(t, "log(1)", testEval("log(1)"), 0.0)
	assertFloat(t, "log(E())", testEval("log(E())"), 1.0)
	assertFloat(t, "log2(8)", testEval("log2(8)"), 3.0)
	assertFloat(t, "log10(1000)", testEval("log10(1000)"), 3.0)
	assertFloat(t, "pow(2, 10)", testEval("pow(2, 10)"), 1024.0)
	assertFloat(t, "cbrt(27)", testEval("cbrt(27)"), 3.0)
	assertFloat(t, "hypot(3, 4)", testEval("hypot(3, 4)"), 5.0)
	assertFloat(t, "exp2(3)", testEval("exp2(3)"), 8.0)
}

func TestSignClampLerp(t *testing.T) {
	assertInt(t, "sign(5)", testEval("sign(5)"), 1)
	assertInt(t, "sign(-3)", testEval("sign(-3)"), -1)
	assertInt(t, "sign(0)", testEval("sign(0)"), 0)
	assertFloat(t, "clamp(15, 0, 10)", testEval("clamp(15, 0, 10)"), 10.0)
	assertFloat(t, "clamp(-5, 0, 10)", testEval("clamp(-5, 0, 10)"), 0.0)
	assertFloat(t, "clamp(5, 0, 10)", testEval("clamp(5, 0, 10)"), 5.0)
	assertFloat(t, "lerp(0, 10, 0.5)", testEval("lerp(0, 10, 0.5)"), 5.0)
	assertFloat(t, "lerp(0, 10, 0.0)", testEval("lerp(0, 10, 0.0)"), 0.0)
	assertFloat(t, "lerp(0, 10, 1.0)", testEval("lerp(0, 10, 1.0)"), 10.0)
}

func TestPredicates(t *testing.T) {
	assertBool(t, "isNaN(NAN())", testEval("isNaN(NAN())"), true)
	assertBool(t, "isNaN(1.0)", testEval("isNaN(1.0)"), false)
	assertBool(t, "isInfinite(INFINITY())", testEval("isInfinite(INFINITY())"), true)
	assertBool(t, "isInfinite(1.0)", testEval("isInfinite(1.0)"), false)
	assertBool(t, "isFinite(42.0)", testEval("isFinite(42.0)"), true)
	assertBool(t, "isFinite(INFINITY())", testEval("isFinite(INFINITY())"), false)
}

func TestGCDLCM(t *testing.T) {
	assertInt(t, "gcd(12, 8)", testEval("gcd(12, 8)"), 4)
	assertInt(t, "gcd(7, 13)", testEval("gcd(7, 13)"), 1)
	assertInt(t, "lcm(4, 6)", testEval("lcm(4, 6)"), 12)
	assertInt(t, "lcm(3, 7)", testEval("lcm(3, 7)"), 21)
}

func TestFactorial(t *testing.T) {
	assertInt(t, "factorial(0)", testEval("factorial(0)"), 1)
	assertInt(t, "factorial(1)", testEval("factorial(1)"), 1)
	assertInt(t, "factorial(5)", testEval("factorial(5)"), 120)
	assertInt(t, "factorial(10)", testEval("factorial(10)"), 3628800)
}

// ── Statistics ──

func TestMean(t *testing.T) {
	assertFloat(t, "mean([1,2,3,4,5])", testEval("mean([1, 2, 3, 4, 5])"), 3.0)
	assertFloat(t, "mean([10,20])", testEval("mean([10, 20])"), 15.0)
}

func TestMedian(t *testing.T) {
	assertFloat(t, "median([1,2,3,4,5])", testEval("median([1, 2, 3, 4, 5])"), 3.0)
	assertFloat(t, "median([1,2,3,4])", testEval("median([1, 2, 3, 4])"), 2.5)
	assertFloat(t, "median([5,1,3])", testEval("median([5, 1, 3])"), 3.0)
}

func TestVarianceStddev(t *testing.T) {
	result := testEval("variance([2, 4, 4, 4, 5, 5, 7, 9])")
	assertFloat(t, "variance", result, 4.571428571428571)

	result2 := testEval("stddev([2, 4, 4, 4, 5, 5, 7, 9])")
	assertFloat(t, "stddev", result2, math.Sqrt(4.571428571428571))
}

func TestPercentile(t *testing.T) {
	assertFloat(t, "percentile 50", testEval("percentile([1, 2, 3, 4, 5], 50)"), 3.0)
	assertFloat(t, "percentile 0", testEval("percentile([1, 2, 3, 4, 5], 0)"), 1.0)
	assertFloat(t, "percentile 100", testEval("percentile([1, 2, 3, 4, 5], 100)"), 5.0)
}

func TestCorrelation(t *testing.T) {
	// Perfect positive correlation
	assertFloat(t, "correlation perfect", testEval("correlation([1, 2, 3, 4, 5], [2, 4, 6, 8, 10])"), 1.0)
	// Perfect negative correlation
	assertFloat(t, "correlation negative", testEval("correlation([1, 2, 3, 4, 5], [10, 8, 6, 4, 2])"), -1.0)
}

func TestLinreg(t *testing.T) {
	result := testEval("linreg([1, 2, 3, 4, 5], [2, 4, 6, 8, 10])")
	tuple, ok := result.(*object.Tuple)
	if !ok {
		t.Fatalf("linreg: expected Tuple, got %T (%s)", result, result.Inspect())
	}
	if len(tuple.Elements) != 3 {
		t.Fatalf("linreg: expected 3 elements, got %d", len(tuple.Elements))
	}
	slope := tuple.Elements[0].(*object.Float).Value
	intercept := tuple.Elements[1].(*object.Float).Value
	rSquared := tuple.Elements[2].(*object.Float).Value
	if math.Abs(slope-2.0) > 1e-9 {
		t.Errorf("linreg slope: expected 2.0, got %g", slope)
	}
	if math.Abs(intercept-0.0) > 1e-9 {
		t.Errorf("linreg intercept: expected 0.0, got %g", intercept)
	}
	if math.Abs(rSquared-1.0) > 1e-9 {
		t.Errorf("linreg r²: expected 1.0, got %g", rSquared)
	}
}

// ── Linear Algebra ──

func TestVectorCreate(t *testing.T) {
	result := testEval("vec(1, 2, 3)")
	v, ok := result.(*object.Vector)
	if !ok {
		t.Fatalf("vec: expected Vector, got %T (%s)", result, result.Inspect())
	}
	if len(v.Elements) != 3 {
		t.Fatalf("vec: expected 3 elements, got %d", len(v.Elements))
	}
	if v.Elements[0] != 1 || v.Elements[1] != 2 || v.Elements[2] != 3 {
		t.Errorf("vec: expected [1,2,3], got %v", v.Elements)
	}
}

func TestVectorFromList(t *testing.T) {
	result := testEval("vec([4, 5, 6])")
	v, ok := result.(*object.Vector)
	if !ok {
		t.Fatalf("vec from list: expected Vector, got %T (%s)", result, result.Inspect())
	}
	if v.Elements[0] != 4 || v.Elements[1] != 5 || v.Elements[2] != 6 {
		t.Errorf("vec from list: expected [4,5,6], got %v", v.Elements)
	}
}

func TestVectorArithmetic(t *testing.T) {
	result := testEval("vec(1, 2, 3) + vec(4, 5, 6)")
	v, ok := result.(*object.Vector)
	if !ok {
		t.Fatalf("vec add: expected Vector, got %T (%s)", result, result.Inspect())
	}
	if v.Elements[0] != 5 || v.Elements[1] != 7 || v.Elements[2] != 9 {
		t.Errorf("vec add: expected [5,7,9], got %v", v.Elements)
	}

	result2 := testEval("vec(4, 5, 6) - vec(1, 2, 3)")
	v2, ok := result2.(*object.Vector)
	if !ok {
		t.Fatalf("vec sub: expected Vector, got %T (%s)", result2, result2.Inspect())
	}
	if v2.Elements[0] != 3 || v2.Elements[1] != 3 || v2.Elements[2] != 3 {
		t.Errorf("vec sub: expected [3,3,3], got %v", v2.Elements)
	}

	// Scalar multiply
	result3 := testEval("vec(1, 2, 3) * 2")
	v3, ok := result3.(*object.Vector)
	if !ok {
		t.Fatalf("vec scale: expected Vector, got %T (%s)", result3, result3.Inspect())
	}
	if v3.Elements[0] != 2 || v3.Elements[1] != 4 || v3.Elements[2] != 6 {
		t.Errorf("vec scale: expected [2,4,6], got %v", v3.Elements)
	}
}

func TestDotProduct(t *testing.T) {
	assertFloat(t, "dot", testEval("dot(vec(1, 2, 3), vec(4, 5, 6))"), 32.0)
}

func TestCrossProduct(t *testing.T) {
	result := testEval("cross(vec(1, 0, 0), vec(0, 1, 0))")
	v, ok := result.(*object.Vector)
	if !ok {
		t.Fatalf("cross: expected Vector, got %T (%s)", result, result.Inspect())
	}
	if v.Elements[0] != 0 || v.Elements[1] != 0 || v.Elements[2] != 1 {
		t.Errorf("cross: expected [0,0,1], got %v", v.Elements)
	}
}

func TestMagnitude(t *testing.T) {
	assertFloat(t, "magnitude", testEval("magnitude(vec(3, 4))"), 5.0)
}

func TestNormalize(t *testing.T) {
	result := testEval("normalize(vec(3, 4))")
	v, ok := result.(*object.Vector)
	if !ok {
		t.Fatalf("normalize: expected Vector, got %T (%s)", result, result.Inspect())
	}
	if math.Abs(v.Elements[0]-0.6) > 1e-9 || math.Abs(v.Elements[1]-0.8) > 1e-9 {
		t.Errorf("normalize: expected [0.6, 0.8], got %v", v.Elements)
	}
}

func TestVectorMethods(t *testing.T) {
	assertFloat(t, "vec.x", testEval("vec(3, 4, 5).x"), 3.0)
	assertFloat(t, "vec.y", testEval("vec(3, 4, 5).y"), 4.0)
	assertFloat(t, "vec.z", testEval("vec(3, 4, 5).z"), 5.0)
	assertFloat(t, "vec.magnitude()", testEval("vec(3, 4).magnitude()"), 5.0)
}

func TestVectorNegate(t *testing.T) {
	result := testEval("-vec(1, 2, 3)")
	v, ok := result.(*object.Vector)
	if !ok {
		t.Fatalf("negate: expected Vector, got %T (%s)", result, result.Inspect())
	}
	if v.Elements[0] != -1 || v.Elements[1] != -2 || v.Elements[2] != -3 {
		t.Errorf("negate: expected [-1,-2,-3], got %v", v.Elements)
	}
}

// ── Matrix ──

func TestMatrixCreate(t *testing.T) {
	result := testEval("mat([1, 2], [3, 4])")
	m, ok := result.(*object.Matrix)
	if !ok {
		t.Fatalf("mat: expected Matrix, got %T (%s)", result, result.Inspect())
	}
	if m.Rows != 2 || m.Cols != 2 {
		t.Errorf("mat: expected 2x2, got %dx%d", m.Rows, m.Cols)
	}
	if m.Data[0][0] != 1 || m.Data[0][1] != 2 || m.Data[1][0] != 3 || m.Data[1][1] != 4 {
		t.Errorf("mat: unexpected data %v", m.Data)
	}
}

func TestMatMul(t *testing.T) {
	result := testEval("matMul(mat([1, 2], [3, 4]), mat([5, 6], [7, 8]))")
	m, ok := result.(*object.Matrix)
	if !ok {
		t.Fatalf("matMul: expected Matrix, got %T (%s)", result, result.Inspect())
	}
	// [1*5+2*7, 1*6+2*8] = [19, 22]
	// [3*5+4*7, 3*6+4*8] = [43, 50]
	if m.Data[0][0] != 19 || m.Data[0][1] != 22 || m.Data[1][0] != 43 || m.Data[1][1] != 50 {
		t.Errorf("matMul: expected [[19,22],[43,50]], got %v", m.Data)
	}
}

func TestTranspose(t *testing.T) {
	result := testEval("transpose(mat([1, 2, 3], [4, 5, 6]))")
	m, ok := result.(*object.Matrix)
	if !ok {
		t.Fatalf("transpose: expected Matrix, got %T (%s)", result, result.Inspect())
	}
	if m.Rows != 3 || m.Cols != 2 {
		t.Errorf("transpose: expected 3x2, got %dx%d", m.Rows, m.Cols)
	}
	if m.Data[0][0] != 1 || m.Data[0][1] != 4 {
		t.Errorf("transpose: unexpected first row %v", m.Data[0])
	}
}

func TestDeterminant(t *testing.T) {
	assertFloat(t, "det 2x2", testEval("determinant(mat([1, 2], [3, 4]))"), -2.0)
	assertFloat(t, "det identity", testEval("determinant(identity(3))"), 1.0)
}

func TestIdentity(t *testing.T) {
	result := testEval("identity(3)")
	m, ok := result.(*object.Matrix)
	if !ok {
		t.Fatalf("identity: expected Matrix, got %T (%s)", result, result.Inspect())
	}
	if m.Rows != 3 || m.Cols != 3 {
		t.Errorf("identity: expected 3x3, got %dx%d", m.Rows, m.Cols)
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			expected := 0.0
			if i == j {
				expected = 1.0
			}
			if m.Data[i][j] != expected {
				t.Errorf("identity[%d][%d]: expected %g, got %g", i, j, expected, m.Data[i][j])
			}
		}
	}
}

func TestMatVecMul(t *testing.T) {
	result := testEval("matVecMul(mat([1, 2], [3, 4]), vec(5, 6))")
	v, ok := result.(*object.Vector)
	if !ok {
		t.Fatalf("matVecMul: expected Vector, got %T (%s)", result, result.Inspect())
	}
	// [1*5+2*6, 3*5+4*6] = [17, 39]
	if v.Elements[0] != 17 || v.Elements[1] != 39 {
		t.Errorf("matVecMul: expected [17, 39], got %v", v.Elements)
	}
}

func TestMatrixMethods(t *testing.T) {
	result := testEval("mat([1, 2], [3, 4]).rows")
	assertInt(t, "rows", result, 2)

	result2 := testEval("mat([1, 2], [3, 4]).cols")
	assertInt(t, "cols", result2, 2)
}

// ── Complex Numbers ──

func TestComplexCreate(t *testing.T) {
	result := testEval("complex(3, 4)")
	c, ok := result.(*object.Complex)
	if !ok {
		t.Fatalf("complex: expected Complex, got %T (%s)", result, result.Inspect())
	}
	if c.Real != 3 || c.Imag != 4 {
		t.Errorf("complex: expected 3+4i, got %g+%gi", c.Real, c.Imag)
	}
}

func TestComplexArithmetic(t *testing.T) {
	// (3+4i) + (1+2i) = (4+6i)
	result := testEval("complex(3, 4) + complex(1, 2)")
	c, ok := result.(*object.Complex)
	if !ok {
		t.Fatalf("complex add: expected Complex, got %T (%s)", result, result.Inspect())
	}
	if c.Real != 4 || c.Imag != 6 {
		t.Errorf("complex add: expected 4+6i, got %g+%gi", c.Real, c.Imag)
	}

	// (3+4i) * (1+2i) = (3*1-4*2) + (3*2+4*1)i = -5 + 10i
	result2 := testEval("complex(3, 4) * complex(1, 2)")
	c2 := result2.(*object.Complex)
	if c2.Real != -5 || c2.Imag != 10 {
		t.Errorf("complex mul: expected -5+10i, got %g+%gi", c2.Real, c2.Imag)
	}
}

func TestComplexMethods(t *testing.T) {
	assertFloat(t, "real", testEval("complex(3, 4).real"), 3.0)
	assertFloat(t, "imag", testEval("complex(3, 4).imag"), 4.0)
	assertFloat(t, "magnitude", testEval("complexMag(complex(3, 4))"), 5.0)

	result := testEval("conjugate(complex(3, 4))")
	c, ok := result.(*object.Complex)
	if !ok {
		t.Fatalf("conjugate: expected Complex, got %T (%s)", result, result.Inspect())
	}
	if c.Real != 3 || c.Imag != -4 {
		t.Errorf("conjugate: expected 3-4i, got %g+%gi", c.Real, c.Imag)
	}
}

func TestComplexNegate(t *testing.T) {
	result := testEval("-complex(3, 4)")
	c, ok := result.(*object.Complex)
	if !ok {
		t.Fatalf("negate: expected Complex, got %T (%s)", result, result.Inspect())
	}
	if c.Real != -3 || c.Imag != -4 {
		t.Errorf("negate: expected -3-4i, got %g+%gi", c.Real, c.Imag)
	}
}

func TestComplexScalarArithmetic(t *testing.T) {
	// (3+4i) + 5 = (8+4i)
	result := testEval("complex(3, 4) + 5")
	c, ok := result.(*object.Complex)
	if !ok {
		t.Fatalf("complex+scalar: expected Complex, got %T (%s)", result, result.Inspect())
	}
	if c.Real != 8 || c.Imag != 4 {
		t.Errorf("complex+scalar: expected 8+4i, got %g+%gi", c.Real, c.Imag)
	}
}

// ── Numerical Utilities ──

func TestLinspace(t *testing.T) {
	result := testEval("linspace(0, 10, 5)")
	list, ok := result.(*object.List)
	if !ok {
		t.Fatalf("linspace: expected List, got %T (%s)", result, result.Inspect())
	}
	if len(list.Elements) != 5 {
		t.Fatalf("linspace: expected 5 elements, got %d", len(list.Elements))
	}
	expected := []float64{0, 2.5, 5, 7.5, 10}
	for i, e := range expected {
		f := list.Elements[i].(*object.Float)
		if math.Abs(f.Value-e) > 1e-9 {
			t.Errorf("linspace[%d]: expected %g, got %g", i, e, f.Value)
		}
	}
}

func TestArange(t *testing.T) {
	result := testEval("arange(0, 5, 1)")
	list, ok := result.(*object.List)
	if !ok {
		t.Fatalf("arange: expected List, got %T (%s)", result, result.Inspect())
	}
	if len(list.Elements) != 5 {
		t.Fatalf("arange: expected 5 elements, got %d", len(list.Elements))
	}
}

// ── Integration: Pipeline with math ──

func TestMathPipeline(t *testing.T) {
	// Sum of squares of evens using math functions
	result := testEval(`
[1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
  |> filter(\x -> x % 2 == 0)
  |> map(\x -> x ^ 2)
  |> mean()
`)
	// evens: [2,4,6,8,10], squares: [4,16,36,64,100], mean = 220/5 = 44
	assertFloat(t, "pipeline+mean", result, 44.0)
}

// Ensure unused import doesn't cause issues
var _ = parser.New
var _ = lexer.New
