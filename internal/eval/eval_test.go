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
