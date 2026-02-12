package parser

import (
	"testing"

	"github.com/rogue780/geoflow/internal/ast"
	"github.com/rogue780/geoflow/internal/lexer"
)

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input       string
		name        string
		mutable     bool
		valueString string
	}{
		{"let x = 5", "x", false, "5"},
		{"let y = 3.14", "y", false, "3.14"},
		{"let name = \"hello\"", "name", false, "\"hello\""},
		{"let flag = true", "flag", false, "true"},
		{"let mut counter = 0", "counter", true, "0"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("input=%q: expected 1 statement, got %d", tt.input, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.LetStatement)
		if !ok {
			t.Fatalf("input=%q: expected LetStatement, got %T", tt.input, program.Statements[0])
		}

		if stmt.Name.Value != tt.name {
			t.Errorf("input=%q: expected name %q, got %q", tt.input, tt.name, stmt.Name.Value)
		}
		if stmt.Mutable != tt.mutable {
			t.Errorf("input=%q: expected mutable=%v, got %v", tt.input, tt.mutable, stmt.Mutable)
		}
	}
}

func TestConstStatements(t *testing.T) {
	input := "const PI = 3.14159"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ConstStatement)
	if !ok {
		t.Fatalf("expected ConstStatement, got %T", program.Statements[0])
	}
	if stmt.Name.Value != "PI" {
		t.Errorf("expected name PI, got %s", stmt.Name.Value)
	}
}

func TestFunctionDefinition(t *testing.T) {
	input := `fn add(a: int, b: int) -> int {
		a + b
	}`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", program.Statements[0])
	}

	fn, ok := stmt.Value.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expected FunctionLiteral, got %T", stmt.Value)
	}

	if fn.Name != "add" {
		t.Errorf("expected function name 'add', got %q", fn.Name)
	}
	if len(fn.Parameters) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(fn.Parameters))
	}
}

func TestDefaultParameters(t *testing.T) {
	input := `fn greet(name, greeting = "Hello") { greeting }`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", program.Statements[0])
	}

	fn, ok := stmt.Value.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expected FunctionLiteral, got %T", stmt.Value)
	}

	if len(fn.Parameters) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(fn.Parameters))
	}

	// First param: no default
	if fn.Parameters[0].Default != nil {
		t.Errorf("expected no default for param 0, got %s", fn.Parameters[0].Default.String())
	}

	// Second param: has default
	if fn.Parameters[1].Default == nil {
		t.Fatal("expected default for param 1, got nil")
	}
	if fn.Parameters[1].Default.String() != `"Hello"` {
		t.Errorf("expected default \"Hello\", got %s", fn.Parameters[1].Default.String())
	}
}

func TestVariadicParameters(t *testing.T) {
	input := `fn sum(...numbers) { numbers }`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", program.Statements[0])
	}

	fn, ok := stmt.Value.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expected FunctionLiteral, got %T", stmt.Value)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(fn.Parameters))
	}

	if !fn.Parameters[0].Variadic {
		t.Error("expected param 0 to be variadic")
	}
	if fn.Parameters[0].Name.Value != "numbers" {
		t.Errorf("expected param name 'numbers', got %q", fn.Parameters[0].Name.Value)
	}
}

func TestVariadicWithRequiredParams(t *testing.T) {
	input := `fn f(a, b, ...rest) { rest }`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.LetStatement)
	fn := stmt.Value.(*ast.FunctionLiteral)

	if len(fn.Parameters) != 3 {
		t.Fatalf("expected 3 parameters, got %d", len(fn.Parameters))
	}

	if fn.Parameters[0].Variadic {
		t.Error("param 0 should not be variadic")
	}
	if fn.Parameters[1].Variadic {
		t.Error("param 1 should not be variadic")
	}
	if !fn.Parameters[2].Variadic {
		t.Error("param 2 should be variadic")
	}
}

func TestPipelineExpression(t *testing.T) {
	input := `x |> f |> g`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
	}

	// Should be a pipeline: (x |> f) |> g
	pipe, ok := stmt.Expression.(*ast.PipelineExpression)
	if !ok {
		t.Fatalf("expected PipelineExpression, got %T", stmt.Expression)
	}

	// Right should be g
	right, ok := pipe.Right.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier for right, got %T", pipe.Right)
	}
	if right.Value != "g" {
		t.Errorf("expected 'g', got %q", right.Value)
	}
}

func TestInfixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{"5 + 5", "+"},
		{"5 - 5", "-"},
		{"5 * 5", "*"},
		{"5 / 5", "/"},
		{"5 % 5", "%"},
		{"5 ^ 2", "^"},
		{"5 > 5", ">"},
		{"5 < 5", "<"},
		{"5 == 5", "=="},
		{"5 != 5", "!="},
		{"true && false", "&&"},
		{"true || false", "||"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("input=%q: expected 1 statement, got %d", tt.input, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("input=%q: expected ExpressionStatement, got %T", tt.input, program.Statements[0])
		}

		infix, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("input=%q: expected InfixExpression, got %T", tt.input, stmt.Expression)
		}
		if infix.Operator != tt.operator {
			t.Errorf("input=%q: expected operator %q, got %q", tt.input, tt.operator, infix.Operator)
		}
	}
}

func TestPrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{"-5", "-"},
		{"!true", "!"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		prefix, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("input=%q: expected PrefixExpression, got %T", tt.input, stmt.Expression)
		}
		if prefix.Operator != tt.operator {
			t.Errorf("input=%q: expected operator %q, got %q", tt.input, tt.operator, prefix.Operator)
		}
	}
}

func TestIfExpression(t *testing.T) {
	input := `if x > 0 then "positive" else "non-positive"`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	ifExpr, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expected IfExpression, got %T", stmt.Expression)
	}
	if ifExpr.Alternative == nil {
		t.Error("expected alternative branch")
	}
}

func TestLambdaExpression(t *testing.T) {
	input := `\x -> x * 2`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	lambda, ok := stmt.Expression.(*ast.LambdaExpression)
	if !ok {
		t.Fatalf("expected LambdaExpression, got %T", stmt.Expression)
	}
	if len(lambda.Parameters) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(lambda.Parameters))
	}
	if lambda.Parameters[0].Value != "x" {
		t.Errorf("expected parameter 'x', got %q", lambda.Parameters[0].Value)
	}
}

func TestListLiteral(t *testing.T) {
	input := `[1, 2, 3, 4, 5]`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	list, ok := stmt.Expression.(*ast.ListLiteral)
	if !ok {
		t.Fatalf("expected ListLiteral, got %T", stmt.Expression)
	}
	if len(list.Elements) != 5 {
		t.Errorf("expected 5 elements, got %d", len(list.Elements))
	}
}

func TestCallExpression(t *testing.T) {
	input := `add(1, 2)`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression, got %T", stmt.Expression)
	}
	if len(call.Arguments) != 2 {
		t.Errorf("expected 2 arguments, got %d", len(call.Arguments))
	}
}

func TestDotExpression(t *testing.T) {
	input := `x.length()`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression, got %T", stmt.Expression)
	}
	dot, ok := call.Function.(*ast.DotExpression)
	if !ok {
		t.Fatalf("expected DotExpression, got %T", call.Function)
	}
	if dot.Field != "length" {
		t.Errorf("expected field 'length', got %q", dot.Field)
	}
}

func TestForExpression(t *testing.T) {
	input := `for x in items { println(x) }`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	forExpr, ok := stmt.Expression.(*ast.ForExpression)
	if !ok {
		t.Fatalf("expected ForExpression, got %T", stmt.Expression)
	}
	ident, ok := forExpr.Pattern.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier pattern, got %T", forExpr.Pattern)
	}
	if ident.Value != "x" {
		t.Errorf("expected pattern 'x', got %q", ident.Value)
	}
}

func TestMatchExpression(t *testing.T) {
	input := `match x {
		1 => "one",
		2 => "two",
		_ => "other"
	}`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	matchExpr, ok := stmt.Expression.(*ast.MatchExpression)
	if !ok {
		t.Fatalf("expected MatchExpression, got %T", stmt.Expression)
	}
	if len(matchExpr.Arms) != 3 {
		t.Errorf("expected 3 arms, got %d", len(matchExpr.Arms))
	}
}

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1 + 2 * 3", "(1 + (2 * 3))"},
		{"1 * 2 + 3", "((1 * 2) + 3)"},
		{"-a * b", "((-a) * b)"},
		{"!true", "(!true)"},
		{"a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f)"},
		{"a ^ b ^ c", "(a ^ (b ^ c))"},  // right-associative
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.Statements[0].(*ast.ExpressionStatement).Expression.String()
		if actual != tt.expected {
			t.Errorf("input=%q: expected=%q, got=%q", tt.input, tt.expected, actual)
		}
	}
}

func TestWhileExpression(t *testing.T) {
	input := `while x > 0 { x := x - 1 }`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	_, ok := stmt.Expression.(*ast.WhileExpression)
	if !ok {
		t.Fatalf("expected WhileExpression, got %T", stmt.Expression)
	}
}

func TestImportStatement(t *testing.T) {
	tests := []struct {
		input string
		path  []string
		alias string
	}{
		{"import std.math", []string{"std", "math"}, ""},
		{"import std.math as m", []string{"std", "math"}, "m"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt, ok := program.Statements[0].(*ast.ImportStatement)
		if !ok {
			t.Fatalf("input=%q: expected ImportStatement, got %T", tt.input, program.Statements[0])
		}
		if len(stmt.Path) != len(tt.path) {
			t.Errorf("input=%q: expected path length %d, got %d", tt.input, len(tt.path), len(stmt.Path))
		}
		if stmt.Alias != tt.alias {
			t.Errorf("input=%q: expected alias %q, got %q", tt.input, tt.alias, stmt.Alias)
		}
	}
}

func TestWKTLiteral(t *testing.T) {
	input := `#POINT(0 0)#`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	wkt, ok := stmt.Expression.(*ast.WKTLiteral)
	if !ok {
		t.Fatalf("expected WKTLiteral, got %T", stmt.Expression)
	}
	if wkt.Value != "POINT(0 0)" {
		t.Errorf("expected 'POINT(0 0)', got %q", wkt.Value)
	}
}

func TestJuxtapositionExpression(t *testing.T) {
	// Simple juxtaposition: f g
	input := `f g`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
	}

	jux, ok := stmt.Expression.(*ast.JuxtapositionExpression)
	if !ok {
		t.Fatalf("expected JuxtapositionExpression, got %T", stmt.Expression)
	}

	leftIdent, ok := jux.Left.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier for left, got %T", jux.Left)
	}
	if leftIdent.Value != "f" {
		t.Errorf("expected left='f', got %q", leftIdent.Value)
	}

	rightIdent, ok := jux.Right.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier for right, got %T", jux.Right)
	}
	if rightIdent.Value != "g" {
		t.Errorf("expected right='g', got %q", rightIdent.Value)
	}
}

func TestJuxtapositionLeftAssociative(t *testing.T) {
	// f g h should parse as (f g) h
	input := `f g h`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
	}

	outer, ok := stmt.Expression.(*ast.JuxtapositionExpression)
	if !ok {
		t.Fatalf("expected JuxtapositionExpression, got %T", stmt.Expression)
	}

	// Right should be h
	rightIdent, ok := outer.Right.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier for outer right, got %T", outer.Right)
	}
	if rightIdent.Value != "h" {
		t.Errorf("expected outer right='h', got %q", rightIdent.Value)
	}

	// Left should be (f g)
	inner, ok := outer.Left.(*ast.JuxtapositionExpression)
	if !ok {
		t.Fatalf("expected JuxtapositionExpression for outer left, got %T", outer.Left)
	}
	leftIdent, ok := inner.Left.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier for inner left, got %T", inner.Left)
	}
	if leftIdent.Value != "f" {
		t.Errorf("expected inner left='f', got %q", leftIdent.Value)
	}
}

func TestJuxtapositionWithCallExpression(t *testing.T) {
	// f g(x) should parse as Jux{f, Call{g, [x]}}
	input := `f g(x)`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
	}

	jux, ok := stmt.Expression.(*ast.JuxtapositionExpression)
	if !ok {
		t.Fatalf("expected JuxtapositionExpression, got %T", stmt.Expression)
	}

	_, ok = jux.Right.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression for right, got %T", jux.Right)
	}
}

func TestJuxtapositionMultilineIsNotJuxtaposition(t *testing.T) {
	// f\ng should parse as two separate statements, NOT juxtaposition
	input := "f\ng"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("expected 2 statements (no juxtaposition across lines), got %d", len(program.Statements))
	}
}

func TestJuxtapositionWithPipeline(t *testing.T) {
	// f g |> h should parse as Pipeline{Jux{f, g}, h}
	input := `f g |> h`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
	}

	pipe, ok := stmt.Expression.(*ast.PipelineExpression)
	if !ok {
		t.Fatalf("expected PipelineExpression, got %T", stmt.Expression)
	}

	_, ok = pipe.Left.(*ast.JuxtapositionExpression)
	if !ok {
		t.Fatalf("expected JuxtapositionExpression for pipeline left, got %T", pipe.Left)
	}

	rightIdent, ok := pipe.Right.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier for pipeline right, got %T", pipe.Right)
	}
	if rightIdent.Value != "h" {
		t.Errorf("expected pipeline right='h', got %q", rightIdent.Value)
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	t.Helper()
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}
	t.Errorf("parser has %d errors:", len(errors))
	for _, msg := range errors {
		t.Errorf("  parser error: %s", msg)
	}
	t.FailNow()
}
