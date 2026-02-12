package object

import (
	"testing"
)

func TestParseSymbolicExpr_Numbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"42", "42"},
		{"3.14", "3.14"},
		{"0.5", "0.5"},
		{"100", "100"},
		{"1e3", "1000"},
		{"2.5e2", "250"},
	}
	for _, tt := range tests {
		expr, errMsg := ParseSymbolicExpr(tt.input)
		if errMsg != "" {
			t.Errorf("ParseSymbolicExpr(%q) error: %s", tt.input, errMsg)
			continue
		}
		if got := expr.Inspect(); got != tt.expected {
			t.Errorf("ParseSymbolicExpr(%q).Inspect() = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseSymbolicExpr_Variables(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x", "x"},
		{"y", "y"},
		{"foo", "foo"},
		{"x1", "x1"},
		{"_a", "_a"},
	}
	for _, tt := range tests {
		expr, errMsg := ParseSymbolicExpr(tt.input)
		if errMsg != "" {
			t.Errorf("ParseSymbolicExpr(%q) error: %s", tt.input, errMsg)
			continue
		}
		if got := expr.Inspect(); got != tt.expected {
			t.Errorf("ParseSymbolicExpr(%q).Inspect() = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseSymbolicExpr_BinaryOps(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x + y", "x + y"},
		{"x - y", "x - y"},
		{"x * y", "x*y"},
		{"x / y", "x/y"},
		{"x ^ y", "x^y"},
	}
	for _, tt := range tests {
		expr, errMsg := ParseSymbolicExpr(tt.input)
		if errMsg != "" {
			t.Errorf("ParseSymbolicExpr(%q) error: %s", tt.input, errMsg)
			continue
		}
		if got := expr.Inspect(); got != tt.expected {
			t.Errorf("ParseSymbolicExpr(%q).Inspect() = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseSymbolicExpr_Precedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// * binds tighter than +
		{"x + y * z", "x + y*z"},
		// ^ binds tighter than *
		{"x * y ^ 2", "x*y^2"},
		// Full precedence chain
		{"a + b * c ^ d", "a + b*c^d"},
		// Left-to-right for same-precedence + and -
		{"a + b - c", "a + b - c"},
		// Left-to-right for same-precedence * and /
		{"a * b / c", "a*b/c"},
	}
	for _, tt := range tests {
		expr, errMsg := ParseSymbolicExpr(tt.input)
		if errMsg != "" {
			t.Errorf("ParseSymbolicExpr(%q) error: %s", tt.input, errMsg)
			continue
		}
		if got := expr.Inspect(); got != tt.expected {
			t.Errorf("ParseSymbolicExpr(%q).Inspect() = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseSymbolicExpr_RightAssociativePow(t *testing.T) {
	// x ^ y ^ z should parse as x ^ (y ^ z)
	expr, errMsg := ParseSymbolicExpr("x ^ y ^ z")
	if errMsg != "" {
		t.Fatalf("ParseSymbolicExpr error: %s", errMsg)
	}
	// Verify structure: ExprPow(x, ExprPow(y, z))
	if expr.Kind != ExprPow {
		t.Fatalf("expected ExprPow, got %d", expr.Kind)
	}
	if expr.Left.Kind != ExprVar || expr.Left.Name != "x" {
		t.Fatalf("expected left=x, got %v", expr.Left)
	}
	if expr.Right.Kind != ExprPow {
		t.Fatalf("expected right=ExprPow, got %d", expr.Right.Kind)
	}
	if expr.Right.Left.Name != "y" || expr.Right.Right.Name != "z" {
		t.Fatalf("expected right=y^z, got %s^%s", expr.Right.Left.Name, expr.Right.Right.Name)
	}
}

func TestParseSymbolicExpr_UnaryNegation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-x", "(-x)"},
		{"-3", "(-3)"},
		{"x + -y", "x + (-y)"},
		{"-x + y", "(-x) + y"},
		{"--x", "(-((-x)))"},
	}
	for _, tt := range tests {
		expr, errMsg := ParseSymbolicExpr(tt.input)
		if errMsg != "" {
			t.Errorf("ParseSymbolicExpr(%q) error: %s", tt.input, errMsg)
			continue
		}
		if got := expr.Inspect(); got != tt.expected {
			t.Errorf("ParseSymbolicExpr(%q).Inspect() = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseSymbolicExpr_Parentheses(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"(x + y) * z", "(x + y)*z"},
		{"x * (y + z)", "x*(y + z)"},
		{"(a + b) * (c + d)", "(a + b)*(c + d)"},
		{"((x))", "x"},
	}
	for _, tt := range tests {
		expr, errMsg := ParseSymbolicExpr(tt.input)
		if errMsg != "" {
			t.Errorf("ParseSymbolicExpr(%q) error: %s", tt.input, errMsg)
			continue
		}
		if got := expr.Inspect(); got != tt.expected {
			t.Errorf("ParseSymbolicExpr(%q).Inspect() = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseSymbolicExpr_Functions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sin(x)", "sin(x)"},
		{"cos(x)", "cos(x)"},
		{"tan(x)", "tan(x)"},
		{"exp(x)", "exp(x)"},
		{"log(x)", "log(x)"},
		{"ln(x)", "ln(x)"},
		{"sqrt(x)", "sqrt(x)"},
		{"abs(x)", "abs(x)"},
		{"asin(x)", "asin(x)"},
		{"acos(x)", "acos(x)"},
		{"atan(x)", "atan(x)"},
	}
	for _, tt := range tests {
		expr, errMsg := ParseSymbolicExpr(tt.input)
		if errMsg != "" {
			t.Errorf("ParseSymbolicExpr(%q) error: %s", tt.input, errMsg)
			continue
		}
		if got := expr.Inspect(); got != tt.expected {
			t.Errorf("ParseSymbolicExpr(%q).Inspect() = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseSymbolicExpr_NestedFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sin(cos(x))", "sin(cos(x))"},
		{"exp(log(x))", "exp(log(x))"},
		{"sqrt(x^2 + y^2)", "sqrt(x^2 + y^2)"},
	}
	for _, tt := range tests {
		expr, errMsg := ParseSymbolicExpr(tt.input)
		if errMsg != "" {
			t.Errorf("ParseSymbolicExpr(%q) error: %s", tt.input, errMsg)
			continue
		}
		if got := expr.Inspect(); got != tt.expected {
			t.Errorf("ParseSymbolicExpr(%q).Inspect() = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseSymbolicExpr_ComplexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x^2 + 3*x + 2", "x^2 + 3*x + 2"},
		{"2*x^2 - 5*x + 3", "2*x^2 - 5*x + 3"},
		{"sin(x) * cos(y)", "sin(x)*cos(y)"},
		{"a*x^2 + b*x + c", "a*x^2 + b*x + c"},
		{"(x + 1) * (x - 1)", "(x + 1)*(x - 1)"},
	}
	for _, tt := range tests {
		expr, errMsg := ParseSymbolicExpr(tt.input)
		if errMsg != "" {
			t.Errorf("ParseSymbolicExpr(%q) error: %s", tt.input, errMsg)
			continue
		}
		if got := expr.Inspect(); got != tt.expected {
			t.Errorf("ParseSymbolicExpr(%q).Inspect() = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseSymbolicExpr_Errors(t *testing.T) {
	tests := []struct {
		input       string
		expectError bool
	}{
		{"", true},          // empty
		{"(x + y", true},   // unclosed paren
		{"* x", true},      // leading operator
		{"x +", true},      // trailing operator
		{"sin(", true},     // unclosed function
		{"sin(x", true},    // unclosed function paren
		{"x y", true},      // missing operator
	}
	for _, tt := range tests {
		_, errMsg := ParseSymbolicExpr(tt.input)
		if tt.expectError && errMsg == "" {
			t.Errorf("ParseSymbolicExpr(%q) expected error but got none", tt.input)
		}
	}
}

func TestParseSymbolicExpr_Whitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  x + y  ", "x + y"},
		{"x+y", "x + y"},
		{"  sin( x )  ", "sin(x)"},
		{" x ^ 2 + 3 * x + 2 ", "x^2 + 3*x + 2"},
	}
	for _, tt := range tests {
		expr, errMsg := ParseSymbolicExpr(tt.input)
		if errMsg != "" {
			t.Errorf("ParseSymbolicExpr(%q) error: %s", tt.input, errMsg)
			continue
		}
		if got := expr.Inspect(); got != tt.expected {
			t.Errorf("ParseSymbolicExpr(%q).Inspect() = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseSymbolicExpr_Realize(t *testing.T) {
	// Parse and then realize to verify structural correctness
	expr, errMsg := ParseSymbolicExpr("x^2 + 3*x + 2")
	if errMsg != "" {
		t.Fatalf("parse error: %s", errMsg)
	}
	// f(x) = x^2 + 3x + 2, f(4) = 16 + 12 + 2 = 30
	val, err := ExprRealize(expr, map[string]float64{"x": 4})
	if err != nil {
		t.Fatalf("realize error: %v", err)
	}
	if val != 30 {
		t.Errorf("expected 30, got %g", val)
	}
}
