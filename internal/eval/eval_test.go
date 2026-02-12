package eval

import (
	"testing"

	"github.com/rogue780/geoflow/internal/lexer"
	"github.com/rogue780/geoflow/internal/object"
	"github.com/rogue780/geoflow/internal/parser"
)

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()
	for name, builtin := range object.GetBuiltins() {
		env.Set(name, builtin, false)
	}
	return Eval(program, env)
}

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
		{"2 ^ 10", 1024},
		{"10 % 3", 1},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected, tt.input)
	}
}

func TestEvalFloatExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"3.14", 3.14},
		{"1.0 + 2.0", 3.0},
		{"5 / 2.0", 2.5},
		{"2.0 ^ 3", 8.0},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected, tt.input)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"(1 < 2) == true", true},
		{"!true", false},
		{"!false", true},
		{"!!true", true},
		{"true && true", true},
		{"true && false", false},
		{"false || true", true},
		{"false || false", false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected, tt.input)
	}
}

func TestEvalStringExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"hello" + " " + "world"`, "hello world"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		str, ok := evaluated.(*object.String)
		if !ok {
			t.Fatalf("input=%q: expected String, got %T (%s)", tt.input, evaluated, evaluated.Inspect())
		}
		if str.Value != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, str.Value)
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5\na", 5},
		{"let a = 5 * 5\na", 25},
		{"let a = 5\nlet b = a\nb", 5},
		{"let a = 5\nlet b = a\nlet c = a + b + 5\nc", 15},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected, tt.input)
	}
}

func TestMutableAssignment(t *testing.T) {
	input := `let mut x = 5
x := 10
x`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 10, input)
}

func TestImmutableAssignmentError(t *testing.T) {
	input := `let x = 5
x := 10`
	evaluated := testEval(input)
	errObj, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("expected Error, got %T", evaluated)
	}
	if errObj.Message != "cannot reassign immutable binding 'x'" {
		t.Errorf("unexpected error: %s", errObj.Message)
	}
}

func TestFunctionDefinition(t *testing.T) {
	input := `fn add(a: int, b: int) -> int {
		a + b
	}
	add(3, 4)`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 7, input)
}

func TestLambdaExpression(t *testing.T) {
	input := `let double = \x -> x * 2
double(5)`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 10, input)
}

func TestClosures(t *testing.T) {
	input := `fn makeAdder(x: int) -> auto {
		fn(y: int) -> int { x + y }
	}
	let addFive = makeAdder(5)
	addFive(3)`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 8, input)
}

func TestRecursion(t *testing.T) {
	input := `fn factorial(n: int) -> int {
		if n <= 1 then 1 else n * factorial(n - 1)
	}
	factorial(5)`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 120, input)
}

func TestIfExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if true then 10", int64(10)},
		{"if false then 10", nil},
		{"if 1 < 2 then 10 else 20", int64(10)},
		{"if 1 > 2 then 10 else 20", int64(20)},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if tt.expected == nil {
			if evaluated != object.NIL {
				t.Errorf("input=%q: expected nil, got %s", tt.input, evaluated.Inspect())
			}
		} else {
			testIntegerObject(t, evaluated, tt.expected.(int64), tt.input)
		}
	}
}

func TestListOperations(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[1, 2, 3]", "[1, 2, 3]"},
		{"len([1, 2, 3])", "3"},
		{"[1, 2, 3][0]", "1"},
		{"[1, 2, 3][-1]", "3"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated.Inspect() != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, evaluated.Inspect())
		}
	}
}

func TestListMethods(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[1, 2, 3].length()", "3"},
		{"[1, 2, 3].map(\\x -> x * 2)", "[2, 4, 6]"},
		{"[1, 2, 3, 4].filter(\\x -> x > 2)", "[3, 4]"},
		{"[1, 2, 3].reduce(0, \\acc, x -> acc + x)", "6"},
		{"[1, 2, 3].append(4)", "[1, 2, 3, 4]"},
		{"[3, 1, 2].reverse()", "[2, 1, 3]"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated.Inspect() != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, evaluated.Inspect())
		}
	}
}

func TestPipelineOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"5 |> toString", "5"},
		{"[1, 2, 3] |> sum", "6"},
		{"[1, 2, 3] |> len", "3"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated.Inspect() != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, evaluated.Inspect())
		}
	}
}

func TestPipelineWithFunctions(t *testing.T) {
	input := `let numbers = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
numbers
	|> filter(\x -> x % 2 == 0)
	|> map(\x -> x * x)
	|> sum`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 220, input)
}

func TestForLoop(t *testing.T) {
	input := `let mut total = 0
for x in [1, 2, 3, 4, 5] {
	total := total + x
}
total`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 15, input)
}

func TestWhileLoop(t *testing.T) {
	input := `let mut i = 0
let mut sum = 0
while i < 5 {
	sum := sum + i
	i := i + 1
}
sum`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 10, input)
}

func TestMatchExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`match 1 { 1 => "one", 2 => "two", _ => "other" }`, "one"},
		{`match 2 { 1 => "one", 2 => "two", _ => "other" }`, "two"},
		{`match 3 { 1 => "one", 2 => "two", _ => "other" }`, "other"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		str, ok := evaluated.(*object.String)
		if !ok {
			t.Fatalf("input=%q: expected String, got %T (%s)", tt.input, evaluated, evaluated.Inspect())
		}
		if str.Value != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, str.Value)
		}
	}
}

func TestMatchWithGuards(t *testing.T) {
	input := `let score = 85
match score {
	s if s >= 90 => "A",
	s if s >= 80 => "B",
	s if s >= 70 => "C",
	_ => "F"
}`
	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("expected String, got %T (%s)", evaluated, evaluated.Inspect())
	}
	if str.Value != "B" {
		t.Errorf("expected 'B', got %q", str.Value)
	}
}

func TestStringMethods(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello".length()`, "5"},
		{`"hello".uppercase()`, "HELLO"},
		{`"HELLO".lowercase()`, "hello"},
		{`" hello ".trim()`, "hello"},
		{`"hello world".contains("world")`, "true"},
		{`"hello world".split(" ")`, "[hello, world]"},
		{`"hello".startsWith("hel")`, "true"},
		{`"hello".endsWith("llo")`, "true"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated.Inspect() != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, evaluated.Inspect())
		}
	}
}

func TestBuiltins(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"typeof(42)", "int"},
		{"typeof(3.14)", "float"},
		{`typeof("hello")`, "string"},
		{"typeof(true)", "bool"},
		{"abs(-5)", "5"},
		{"abs(5)", "5"},
		{"min(3, 7)", "3"},
		{"max(3, 7)", "7"},
		{"range(5)", "[0, 1, 2, 3, 4]"},
		{"range(1, 5)", "[1, 2, 3, 4]"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated.Inspect() != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, evaluated.Inspect())
		}
	}
}

func TestOptionType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Some(42)", "Some(42)"},
		{"None()", "None"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated.Inspect() != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, evaluated.Inspect())
		}
	}
}

func TestResultType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Ok(42)", "Ok(42)"},
		{`Err("oops")`, "Err(oops)"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated.Inspect() != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, evaluated.Inspect())
		}
	}
}

func TestBlockReturnValue(t *testing.T) {
	input := `fn test() -> int {
		let x = 5
		let y = 10
		x + y
	}
	test()`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 15, input)
}

func TestBreakContinue(t *testing.T) {
	input := `let mut sum = 0
for x in [1, 2, 3, 4, 5, 6, 7, 8, 9, 10] {
	if x > 5 then break
	sum := sum + x
}
sum`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 15, input)
}

func TestTuple(t *testing.T) {
	input := `let t = (1, "hello", true)
t`
	evaluated := testEval(input)
	tuple, ok := evaluated.(*object.Tuple)
	if !ok {
		t.Fatalf("expected Tuple, got %T", evaluated)
	}
	if len(tuple.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(tuple.Elements))
	}
}

func TestTupleIndexing(t *testing.T) {
	input := `let t = (10, 20, 30)
t[1]`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 20, input)
}

func TestStringInterpolation(t *testing.T) {
	input := `let x = 42
"value is {x}"`
	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("expected String, got %T (%s)", evaluated, evaluated.Inspect())
	}
	if str.Value != "value is 42" {
		t.Errorf("expected 'value is 42', got %q", str.Value)
	}
}

func TestWKTLiteral(t *testing.T) {
	input := `#POINT(0 0)#`
	evaluated := testEval(input)
	pt, ok := evaluated.(*object.Point)
	if !ok {
		t.Fatalf("expected Point, got %T (%s)", evaluated, evaluated.Inspect())
	}
	if pt.Coord.X != 0 || pt.Coord.Y != 0 {
		t.Errorf("expected POINT(0 0), got %s", pt.Inspect())
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"5 + true", "unknown operator: int + bool"},
		{`"hello" - "world"`, "unknown operator: string - string"},
		{"foobar", "identifier not found: foobar"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Fatalf("input=%q: expected Error, got %T (%s)", tt.input, evaluated, evaluated.Inspect())
		}
		if errObj.Message != tt.expected {
			t.Errorf("input=%q: expected error %q, got %q", tt.input, tt.expected, errObj.Message)
		}
	}
}

// Helper functions

func testIntegerObject(t *testing.T, obj object.Object, expected int64, input string) {
	t.Helper()
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Fatalf("input=%q: expected Integer, got %T (%s)", input, obj, obj.Inspect())
	}
	if result.Value != expected {
		t.Errorf("input=%q: expected %d, got %d", input, expected, result.Value)
	}
}

func testFloatObject(t *testing.T, obj object.Object, expected float64, input string) {
	t.Helper()
	result, ok := obj.(*object.Float)
	if !ok {
		t.Fatalf("input=%q: expected Float, got %T (%s)", input, obj, obj.Inspect())
	}
	if result.Value != expected {
		t.Errorf("input=%q: expected %f, got %f", input, expected, result.Value)
	}
}

func TestDotComposition(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		// f . g → f(g(x))
		{`fn double(x) { x * 2 }
fn inc(x) { x + 1 }
let f = inc . double
f(5)`, 11}, // inc(double(5)) = inc(10) = 11
		// Chaining: h . g . f → h(g(f(x)))
		{`let a = \x -> x + 1
let b = \x -> x * 2
let c = \x -> x * 3
let f = c . b . a
f(1)`, 12}, // c(b(a(1))) = c(b(2)) = c(4) = 12
		// Dot composition with lambdas (non-identifier right side)
		{`let f = (\x -> x * 3) . (\x -> x + 1) . (\x -> x * 2)
f(2)`, 15}, // 3 * ((2*2) + 1) = 15
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected, tt.input)
	}
}

func TestDotCompositionDoesNotBreakFieldAccess(t *testing.T) {
	// Dot on lists/strings should still work as field access
	tests := []struct {
		input    string
		expected int64
	}{
		{`len([1, 2, 3])`, 3},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected, tt.input)
	}
}

func TestJuxtapositionComposition(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		// f g → g(f(x)) (left-to-right)
		{`fn double(x) { x * 2 }
fn inc(x) { x + 1 }
let f = double inc
f(5)`, 11}, // inc(double(5)) = 11
		// Chain: a b c → c(b(a(x)))
		{`let a = \x -> x + 1
let b = \x -> x * 2
let c = \x -> x * 3
let f = a b c
f(1)`, 12}, // c(b(a(1))) = c(b(2)) = c(4) = 12
		// Value application: 5 double → double(5)
		{`fn double(x) { x * 2 }
5 double`, 10},
		// Mixed: 5 inc double → double(inc(5))
		{`fn double(x) { x * 2 }
fn inc(x) { x + 1 }
5 inc double`, 12}, // double(inc(5)) = double(6) = 12
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected, tt.input)
	}
}

func TestJuxtapositionWithPartialApplication(t *testing.T) {
	// inc triple → CF{Outer: triple, Inner: inc} → triple(inc(x))
	// f(5) = triple(inc(5)) = mul(3, 6) = 18
	input := `fn mul(a, b) { a * b }
let triple = mul(3)
fn inc(x) { x + 1 }
let f = inc triple
f(5)`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 18, input)
}

func TestDefaultParameters(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		// Default used
		{`fn add(a, b = 10) { a + b }
add(1)`, 11},
		// Default overridden
		{`fn add(a, b = 10) { a + b }
add(1, 2)`, 3},
		// Multiple defaults, all used
		{`fn add(a, b = 10, c = 20) { a + b + c }
add(1)`, 31},
		// Multiple defaults, one overridden
		{`fn add(a, b = 10, c = 20) { a + b + c }
add(1, 2)`, 23},
		// Multiple defaults, all overridden
		{`fn add(a, b = 10, c = 20) { a + b + c }
add(1, 2, 3)`, 6},
		// Partial application with default params
		{`fn add(a, b, c = 10) { a + b + c }
let addOne = add(1)
addOne(2)`, 13},
		// Partial application overriding default
		{`fn add(a, b, c = 10) { a + b + c }
let addOne = add(1)
addOne(2, 3)`, 6},
		// Default referencing earlier expression
		{`fn f(a, b = 5) { a * b }
f(3)`, 15},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected, tt.input)
	}
}

func TestVariadicFunctions(t *testing.T) {
	intTests := []struct {
		input    string
		expected int64
	}{
		// Basic variadic: collect all args into list, reduce
		{`fn sum(...numbers) {
	numbers.reduce(0, \acc, n -> acc + n)
}
sum(1, 2, 3)`, 6},
		// Variadic with single arg
		{`fn sum(...numbers) {
	numbers.reduce(0, \acc, n -> acc + n)
}
sum(42)`, 42},
		// Variadic with zero args
		{`fn sum(...numbers) {
	numbers.reduce(0, \acc, n -> acc + n)
}
sum()`, 0},
		// Variadic with preceding required params
		{`fn addTo(base, ...numbers) {
	base + numbers.reduce(0, \acc, n -> acc + n)
}
addTo(100, 1, 2, 3)`, 106},
		// Variadic with preceding required params, zero variadic args
		{`fn addTo(base, ...numbers) {
	base + numbers.reduce(0, \acc, n -> acc + n)
}
addTo(100)`, 100},
		// Variadic length check
		{`fn count(...items) { len(items) }
count(1, 2, 3, 4, 5)`, 5},
		// Variadic with preceding required and default params
		{`fn f(a, b = 10, ...rest) {
	a + b + len(rest)
}
f(1)`, 11},
		// Variadic with preceding required and default params, default overridden
		{`fn f(a, b = 10, ...rest) {
	a + b + len(rest)
}
f(1, 2, 3, 4)`, 5},
	}

	for _, tt := range intTests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected, tt.input)
	}

	// Test that variadic param is a list
	input := `fn first(...args) { args[0] }
first(10, 20, 30)`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 10, input)
}

func TestGenericFunctions(t *testing.T) {
	// Generic type parameters are parsed but ignored at runtime (dynamically typed)
	intTests := []struct {
		input    string
		expected int64
	}{
		// Generic identity function
		{`fn identity<T>(x: T) -> T { x }
identity(42)`, 42},
		// Generic function with multiple type params
		{`fn first<A, B>(a: A, b: B) -> A { a }
first(10, "hello")`, 10},
		// Generic function works with different types
		{`fn apply<T, U>(f: Fn<(T) -> U>, x: T) -> U { f(x) }
apply(\x -> x * 2, 21)`, 42},
	}

	for _, tt := range intTests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected, tt.input)
	}

	// Generic function returning string
	input := `fn identity<T>(x: T) -> T { x }
identity("hello")`
	evaluated := testEval(input)
	if s, ok := evaluated.(*object.String); !ok || s.Value != "hello" {
		t.Fatalf("expected string 'hello', got %T (%s)", evaluated, evaluated.Inspect())
	}

	// Type annotations with generics on parameters
	input2 := `fn len2<T>(list: List<T>) -> int { len(list) }
len2([1, 2, 3])`
	evaluated2 := testEval(input2)
	testIntegerObject(t, evaluated2, 3, input2)
}

func TestDataStructures(t *testing.T) {
	// Array constructors
	tests := []struct {
		input    string
		expected string
	}{
		{`import std.data
let a = data.zeros([3])
a.size()`, "3"},
		{`import std.data
let a = data.ones([2, 3])
a.ndim()`, "2"},
		{`import std.data
let a = data.eye(3)
a.shape()`, "[3, 3]"},
		{`import std.data
let a = data.linspace(0, 10, 5)
a.size()`, "5"},
		{`import std.data
let a = data.fromList([1, 2, 3, 4])
a.sum()`, "10.0"},
		{`import std.data
let a = data.fromList([1, 2, 3, 4])
a.mean()`, "2.5"},
		{`import std.data
let a = data.fromNested([[1, 2], [3, 4]])
a.shape()`, "[2, 2]"},
		// Series
		{`import std.data
let s = data.Series([10, 20, 30])
s.length()`, "3"},
		{`import std.data
let s = data.Series([10, 20, 30], "values")
s.name`, "values"},
		{`import std.data
let s = data.Series([1, 2, 3])
s.sum()`, "6.0"},
		// DataFrame
		{`import std.data
let df = data.DataFrame({"a": [1, 2, 3], "b": [4, 5, 6]})
df.length()`, "3"},
		{`import std.data
let df = data.DataFrame({"x": [10, 20], "y": [30, 40]})
df.shape()`, "(2, 2)"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated == nil {
			t.Fatalf("input=%q: got nil", tt.input)
		}
		actual := evaluated.Inspect()
		if actual != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

func TestSymbolicMath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Symbolic variable creation
		{`import std.math.symbolic as sym
let x = sym.var("x")
x`, "x"},
		// Symbolic arithmetic
		{`import std.math.symbolic as sym
let x = sym.var("x")
let expr = sym.add(x, sym.num(1))
expr`, "x + 1"},
		// Symbolic differentiation (d/dx of x^2 = 2*x)
		{`import std.math.symbolic as sym
let x = sym.var("x")
let expr = sym.pow(x, sym.num(2))
sym.diff(expr, "x")`, "2*x"},
		// Symbolic substitution and realize
		{`import std.math.symbolic as sym
let x = sym.var("x")
let expr = sym.add(x, sym.num(3))
expr.realize({"x": 5.0})`, "8.0"},
		// Symbolic simplify (0 + x = x)
		{`import std.math.symbolic as sym
let x = sym.var("x")
sym.simplify(sym.add(sym.num(0), x))`, "x"},
		// Symbolic toString
		{`import std.math.symbolic as sym
let x = sym.var("x")
let expr = sym.mul(sym.num(2), x)
expr.toString()`, "2*x"},
		// Expr infix operator overloading
		{`import std.math.symbolic as sym
let x = sym.var("x")
let expr = x + sym.num(1)
expr`, "x + 1"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated == nil {
			t.Fatalf("input=%q: got nil", tt.input)
		}
		actual := evaluated.Inspect()
		if actual != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

func TestNumericalMethods(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Bisection root finding: sqrt(2) ≈ 1.414...
		{`import std.math.numeric as num
let f = \x -> x * x - 2
let root = num.bisect(f, 1.0, 2.0)
root > 1.41 && root < 1.42`, "true"},
		// Newton's method
		{`import std.math.numeric as num
let f = \x -> x * x - 4
let root = num.newton(f, 3.0)
root > 1.99 && root < 2.01`, "true"},
		// Numerical differentiation
		{`import std.math.numeric as num
let f = \x -> x * x
let deriv = num.diff(f, 3.0)
deriv > 5.99 && deriv < 6.01`, "true"},
		// Polynomial evaluation
		{`import std.math.numeric as num
num.polyval([1.0, 0.0, -1.0], 2.0)`, "3.0"},
		// Linspace
		{`import std.math.numeric as num
let xs = num.linspace(0.0, 1.0, 3)
len(xs)`, "3"},
		// Arange
		{`import std.math.numeric as num
let xs = num.arange(0.0, 5.0, 1.0)
len(xs)`, "5"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated == nil {
			t.Fatalf("input=%q: got nil", tt.input)
		}
		actual := evaluated.Inspect()
		if actual != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

func TestRasterOperations(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Raster empty construction
		{`import std.geo.raster as raster
let r = raster.empty(10, 10, {"minX": 0.0, "minY": 0.0, "maxX": 1.0, "maxY": 1.0})
r`, "Raster(10x10)"},
		// Raster from array
		{`import std.geo.raster as raster
let data = [1.0, 2.0, 3.0, 4.0]
let r = raster.fromArray(data, 2, 2)
r`, "Raster(2x2)"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated == nil {
			t.Fatalf("input=%q: got nil", tt.input)
		}
		actual := evaluated.Inspect()
		if actual != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

func TestH3S2Indexing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// H3 index creation
		{`import std.geo.h3 as h3
let idx = h3.fromLatLon(37.7749, -122.4194, 5)
h3.isValid(idx)`, "true"},
		// H3 resolution
		{`import std.geo.h3 as h3
let idx = h3.fromLatLon(37.7749, -122.4194, 7)
h3.resolution(idx)`, "7"},
		// H3 kRing returns a list
		{`import std.geo.h3 as h3
let idx = h3.fromLatLon(37.7749, -122.4194, 5)
let ring = h3.kRing(idx, 1)
len(ring) > 0`, "true"},
		// H3 distance is non-negative
		{`import std.geo.h3 as h3
let a = h3.fromLatLon(37.7749, -122.4194, 5)
let b = h3.fromLatLon(37.8, -122.4, 5)
h3.distance(a, b) >= 0`, "true"},
		// S2 cell creation
		{`import std.geo.s2 as s2
let cell = s2.fromLatLon(37.7749, -122.4194, 10)
s2.isValid(cell)`, "true"},
		// S2 level
		{`import std.geo.s2 as s2
let cell = s2.fromLatLon(37.7749, -122.4194, 12)
s2.level(cell)`, "12"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated == nil {
			t.Fatalf("input=%q: got nil", tt.input)
		}
		actual := evaluated.Inspect()
		if actual != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

func TestSpatialAnalysis(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Nearest neighbor
		{`import std.geo.analysis as analysis
let target = #POINT(0 0)#
let candidates = [#POINT(1 0)#, #POINT(0 2)#, #POINT(3 3)#]
let nearest = analysis.nearestNeighbor(target, candidates)
nearest`, "POINT (1 0)"},
		// DBSCAN clustering returns a list
		{`import std.geo.analysis as analysis
let pts = [#POINT(0 0)#, #POINT(0.1 0.1)#, #POINT(5 5)#, #POINT(5.1 5.1)#]
let clusters = analysis.dbscan(pts, 1.0, 1)
len(clusters) >= 1`, "true"},
		// K-means clustering
		{`import std.geo.analysis as analysis
let pts = [#POINT(0 0)#, #POINT(1 0)#, #POINT(10 10)#, #POINT(11 10)#]
let clusters = analysis.kmeans(pts, 2)
len(clusters)`, "2"},
		// IDW interpolation
		{`import std.geo.analysis as analysis
let pts = [#POINT(0 0)#, #POINT(1 0)#, #POINT(0 1)#]
let vals = [10.0, 20.0, 30.0]
let target = #POINT(0.5 0.5)#
let result = analysis.idw(pts, vals, target)
result > 0`, "true"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated == nil {
			t.Fatalf("input=%q: got nil", tt.input)
		}
		actual := evaluated.Inspect()
		if actual != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

// Task 10: CRS/Transform tests
func TestCRSTransform(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// CRS construction via fromEPSG (returns Result, unwrap with ?)
		{`import std.geo.crs as crs
let c = crs.fromEPSG(4326)?
c.name`, "WGS 84"},
		// CRS isGeographic (property)
		{`import std.geo.crs as crs
let c = crs.fromEPSG(4326)?
c.isGeographic`, "true"},
		// CRS isProjected for Web Mercator (property)
		{`import std.geo.crs as crs
let c = crs.fromEPSG(3857)?
c.isProjected`, "true"},
		// CRS toProj4 (property)
		{`import std.geo.crs as crs
let c = crs.fromEPSG(4326)?
c.toProj4 != ""`, "true"},
		// CRS datum (property)
		{`import std.geo.crs as crs
let c = crs.fromEPSG(4326)?
c.datum != ""`, "true"},
		// Transform toWebMercator returns a point
		{`import std.geo.transform as t
let p = #POINT(0 0)#
let wm = t.toWebMercator(p)
wm != nil`, "true"},
		// Transform fromWebMercator returns a point
		{`import std.geo.transform as t
let wm = t.toWebMercator(#POINT(0 0)#)
let back = t.fromWebMercator(wm)
back != nil`, "true"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated == nil {
			t.Fatalf("input=%q: got nil", tt.input)
		}
		actual := evaluated.Inspect()
		if actual != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

// Task 11: Linear Algebra tests
func TestLinearAlgebra(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Vector construction
		{`import std.math.linalg as la
let v = la.vector([1.0, 2.0, 3.0])
v`, "vec(1.0, 2.0, 3.0)"},
		// Identity matrix trace
		{`import std.math.linalg as la
let m = la.eye(2)
la.trace(m)`, "2.0"},
		// Zeros matrix norm
		{`import std.math.linalg as la
let m = la.zeros(2, 3)
la.matrixNorm(m)`, "0.0"},
		// Determinant of 2x2
		{`import std.math.linalg as la
let m = la.matrix([1.0, 2.0], [3.0, 4.0])
la.det(m)`, "-2.0"},
		// Solve Ax=b
		{`import std.math.linalg as la
let A = la.matrix([2.0, 1.0], [1.0, 3.0])
let b = la.vector([5.0, 10.0])
let x = la.solve(A, b)
x`, "vec(1.0, 3.0)"},
		// Matrix rank
		{`import std.math.linalg as la
let m = la.eye(3)
la.rank(m)`, "3"},
		// Hilbert 2x2: H[0][0]=1, H[1][1]=1/3, trace = 4/3 = 1.333...
		{`import std.math.linalg as la
let h = la.hilbert(2)
la.det(h) > 0`, "true"},
		// Kronecker product trace
		{`import std.math.linalg as la
let a = la.eye(2)
let b = la.eye(3)
let k = la.kronecker(a, b)
la.trace(k)`, "6.0"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated == nil {
			t.Fatalf("input=%q: got nil", tt.input)
		}
		actual := evaluated.Inspect()
		if actual != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

// Task 12: Geography tests
func TestGeography(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Geodesic distance (great circle)
		{`import std.geo.geog as geog
let d = geog.distanceGreatCircle(#POINT(0 0)#, #POINT(0 1)#)
d > 100000`, "true"},
		// Bearing between points
		{`import std.geo.geog as geog
let b = geog.bearing(#POINT(0 0)#, #POINT(1 0)#)
b >= 0.0`, "true"},
		// Buffer creates a polygon
		{`import std.geo.geog as geog
let p = #POINT(0 0)#
let poly = geog.buffer(p, 1000.0)
poly != nil`, "true"},
		// fromLatLon
		{`import std.geo.geog as geog
let p = geog.fromLatLon(51.5, -0.1)
p`, "POINT (-0.1 51.5)"},
		// Interpolate between points
		{`import std.geo.geog as geog
let p = geog.interpolate(#POINT(0 0)#, #POINT(10 0)#, 0.5)
p != nil`, "true"},
		// Destination point
		{`import std.geo.geog as geog
let p = geog.destination(#POINT(0 0)#, 90.0, 100000.0)
p != nil`, "true"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated == nil {
			t.Fatalf("input=%q: got nil", tt.input)
		}
		actual := evaluated.Inspect()
		if actual != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

// Task 13: Assertions tests (assertions return nil on success, Error on failure)
func TestAssertions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// assert.equal (returns nil on success)
		{`import std.assert as assert
assert.equal(1, 1)
"pass"`, "pass"},
		// assert.notEqual
		{`import std.assert as assert
assert.notEqual(1, 2)
"pass"`, "pass"},
		// assert.approximately
		{`import std.assert as assert
assert.approximately(3.14159, 3.14160, 0.001)
"pass"`, "pass"},
		// assert.greaterThan
		{`import std.assert as assert
assert.greaterThan(5, 3)
"pass"`, "pass"},
		// assert.lessThan
		{`import std.assert as assert
assert.lessThan(3, 5)
"pass"`, "pass"},
		// assert.contains for list
		{`import std.assert as assert
assert.contains([1, 2, 3], 2)
"pass"`, "pass"},
		// assert.empty
		{`import std.assert as assert
assert.empty([])
"pass"`, "pass"},
		// assert.positive
		{`import std.assert as assert
assert.positive(5.0)
"pass"`, "pass"},
		// assert.negative
		{`import std.assert as assert
assert.negative(-3.0)
"pass"`, "pass"},
		// assert.startsWith
		{`import std.assert as assert
assert.startsWith("hello world", "hello")
"pass"`, "pass"},
		// assert.endsWith
		{`import std.assert as assert
assert.endsWith("hello world", "world")
"pass"`, "pass"},
		// assert.blank for empty string
		{`import std.assert as assert
assert.blank("")
"pass"`, "pass"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated == nil {
			t.Fatalf("input=%q: got nil", tt.input)
		}
		actual := evaluated.Inspect()
		if actual != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

// Task 14: Logging tests (log functions return nil on success)
func TestLogging(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// setLevel returns nil on success
		{`import std.log as log
log.setLevel("debug")
"pass"`, "pass"},
		// withContext returns a map
		{`import std.log as log
let ctx = log.withContext({"module": "test"})
ctx != nil`, "true"},
		// setFormat returns nil on success
		{`import std.log as log
log.setFormat("%level - %message")
"pass"`, "pass"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated == nil {
			t.Fatalf("input=%q: got nil", tt.input)
		}
		actual := evaluated.Inspect()
		if actual != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

// Task 15: Statistics tests
func TestStatistics(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Existing stats (floats return with .0)
		{`import std.math.stats as stats
stats.mean([1.0, 2.0, 3.0, 4.0, 5.0])`, "3.0"},
		// Skewness
		{`import std.math.stats as stats
let s = stats.skewness([1.0, 2.0, 3.0, 4.0, 5.0])
s >= -1.0 && s <= 1.0`, "true"},
		// Kurtosis
		{`import std.math.stats as stats
let k = stats.kurtosis([1.0, 2.0, 3.0, 4.0, 5.0])
k >= -10.0 && k <= 10.0`, "true"},
		// Weighted mean
		{`import std.math.stats as stats
let wm = stats.weightedMean([1.0, 2.0, 3.0], [1.0, 1.0, 1.0])
wm`, "2.0"},
		// Linear regression returns struct
		{`import std.math.stats as stats
let model = stats.linearRegression([1.0, 2.0, 3.0, 4.0], [2.0, 4.0, 6.0, 8.0])
model.slope`, "2.0"},
		// Min-max scale
		{`import std.math.stats as stats
let scaled = stats.minMaxScale([1.0, 2.0, 3.0, 4.0, 5.0])
len(scaled)`, "5"},
		// Standardize
		{`import std.math.stats as stats
let z = stats.standardize([1.0, 2.0, 3.0, 4.0, 5.0])
len(z)`, "5"},
		// Product
		{`import std.math.stats as stats
stats.product([2.0, 3.0, 4.0])`, "24.0"},
		// EWMA
		{`import std.math.stats as stats
let e = stats.ewma([1.0, 2.0, 3.0, 4.0], 0.5)
len(e)`, "4"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated == nil {
			t.Fatalf("input=%q: got nil", tt.input)
		}
		actual := evaluated.Inspect()
		if actual != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

// Task 16: Extended string methods tests
func TestStringMethodsExtended(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// titleCase
		{`"hello world".titleCase()`, "Hello World"},
		// strip
		{`"***hello***".strip("*")`, "hello"},
		// slice
		{`"hello world".slice(0, 5)`, "hello"},
		// insert
		{`"helo".insert(3, "l")`, "hello"},
		// remove
		{`"hello world".remove(5, 11)`, "hello"},
		// matches
		{`"hello123".matches("[0-9]+")`, "true"},
		// findAll
		{`let results = "cat bat hat".findAll("[a-z]at")
len(results)`, "3"},
		// replaceRegex
		{`"foo123bar456".replaceRegex("[0-9]+", "NUM")`, "fooNUMbarNUM"},
		// parseInt (returns Result)
		{`"42".parseInt()`, "Ok(42)"},
		// parseFloat (returns Result)
		{`"3.14".parseFloat()`, "Ok(3.14)"},
		// parseBool (returns Result)
		{`"true".parseBool()`, "Ok(true)"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated == nil {
			t.Fatalf("input=%q: got nil", tt.input)
		}
		actual := evaluated.Inspect()
		if actual != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

// Task 18: Geometry methods tests
func TestGeometryMethods(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Point equals
		{`#POINT(1 2)#.equals(#POINT(1 2)#)`, "true"},
		// Point coordinateDimension (property, not method)
		{`#POINT(1 2)#.coordinateDimension`, "2"},
		// LineString equals
		{`#LINESTRING(0 0, 1 1)#.equals(#LINESTRING(0 0, 1 1)#)`, "true"},
		// Polygon covers
		{`let poly = #POLYGON((0 0, 10 0, 10 10, 0 10, 0 0))#
let pt = #POINT(5 5)#
poly.covers(pt)`, "true"},
		// isSimple for point (property, not method)
		{`#POINT(1 2)#.isSimple`, "true"},
		// translate point
		{`let p = #POINT(1 2)#
p.translate(3.0, 4.0)`, "POINT (4 6)"},
		// scale point
		{`let p = #POINT(2 3)#
p.scale(2.0)`, "POINT (4 6)"},
		// rotate point (90 degrees = pi/2)
		{`let p = #POINT(1 0)#
let r = p.rotate(1.5707963267948966)
r != nil`, "true"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated == nil {
			t.Fatalf("input=%q: got nil", tt.input)
		}
		actual := evaluated.Inspect()
		if actual != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

// Task 19: DateTime methods tests
func TestDateTimeMethods(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// DateTime.now exists
		{`import std.time as time
let dt = time.now()
dt != nil`, "true"},
		// fromUnixMillis
		{`import std.time as time
let dt = time.fromUnixMillis(0)
dt != nil`, "true"},
		// instant
		{`import std.time as time
let i = time.instant()
i != nil`, "true"},
		// DateTime quarter (property)
		{`import std.time as time
let dt = time.parse("2024-03-15T12:00:00Z", "RFC3339")?
dt.quarter`, "1"},
		// DateTime isBefore (method)
		{`import std.time as time
let a = time.parse("2024-01-01T00:00:00Z", "RFC3339")?
let b = time.parse("2024-06-01T00:00:00Z", "RFC3339")?
a.isBefore(b)`, "true"},
		// DateTime isAfter (method)
		{`import std.time as time
let a = time.parse("2024-06-01T00:00:00Z", "RFC3339")?
let b = time.parse("2024-01-01T00:00:00Z", "RFC3339")?
a.isAfter(b)`, "true"},
		// DateTime startOfDay (property) then .hour (property)
		{`import std.time as time
let dt = time.parse("2024-03-15T14:30:00Z", "RFC3339")?
dt.startOfDay.hour`, "0"},
		// DateTime startOfMonth (property) then .day (property)
		{`import std.time as time
let dt = time.parse("2024-03-15T14:30:00Z", "RFC3339")?
dt.startOfMonth.day`, "1"},
		// DateTime withYear (method) then .year (property)
		{`import std.time as time
let dt = time.parse("2024-03-15T14:30:00Z", "RFC3339")?
dt.withYear(2025).year`, "2025"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated == nil {
			t.Fatalf("input=%q: got nil", tt.input)
		}
		actual := evaluated.Inspect()
		if actual != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

// Task 20: I/O tests
func TestIOExtensions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// path.relative returns a Result
		{`import std.io as io
let r = io.relative("/home/user", "/home/user/docs/file.txt")?
r`, "docs/file.txt"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated == nil {
			t.Fatalf("input=%q: got nil", tt.input)
		}
		actual := evaluated.Inspect()
		if actual != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

// Task 21: Miscellaneous tests
func TestByteBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`byte(65)`, "65"},
		{`byte(0)`, "0"},
		{`byte(255)`, "255"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if evaluated == nil {
			t.Fatalf("input=%q: got nil", tt.input)
		}
		actual := evaluated.Inspect()
		if actual != tt.expected {
			t.Errorf("input=%q: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool, input string) {
	t.Helper()
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Fatalf("input=%q: expected Boolean, got %T (%s)", input, obj, obj.Inspect())
	}
	if result.Value != expected {
		t.Errorf("input=%q: expected %v, got %v", input, expected, result.Value)
	}
}
