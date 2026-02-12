package object

import (
	"fmt"
	"math"
	"strings"
)

const EXPR_OBJ Type = "Expr"

// ExprType identifies the kind of symbolic expression.
type ExprType int

const (
	ExprNum  ExprType = iota // numeric literal
	ExprVar                  // named variable
	ExprAdd                  // left + right
	ExprSub                  // left - right
	ExprMul                  // left * right
	ExprDiv                  // left / right
	ExprPow                  // left ^ right
	ExprNeg                  // unary negation of Arg
	ExprFunc                 // named function applied to Arg (sin, cos, exp, log, ...)
)

// Expr represents a symbolic mathematical expression.
type Expr struct {
	Kind  ExprType
	Value float64 // for ExprNum
	Name  string  // for ExprVar or ExprFunc
	Left  *Expr   // binary ops
	Right *Expr   // binary ops
	Arg   *Expr   // for ExprFunc, ExprNeg
}

func (e *Expr) Type() Type { return EXPR_OBJ }

func (e *Expr) Inspect() string {
	return exprToString(e, false)
}

// exprToString renders a symbolic expression as a human-readable math string.
// parenthesize indicates whether to wrap lower-precedence subexpressions.
func exprToString(e *Expr, parenthesize bool) string {
	if e == nil {
		return "nil"
	}
	switch e.Kind {
	case ExprNum:
		if e.Value == math.Trunc(e.Value) && !math.IsInf(e.Value, 0) && !math.IsNaN(e.Value) {
			v := int64(e.Value)
			return fmt.Sprintf("%d", v)
		}
		return fmt.Sprintf("%g", e.Value)
	case ExprVar:
		return e.Name
	case ExprNeg:
		inner := exprToString(e.Arg, true)
		// If the arg is a simple atom, no parens needed
		if e.Arg.Kind == ExprNum || e.Arg.Kind == ExprVar {
			return fmt.Sprintf("(-%s)", inner)
		}
		return fmt.Sprintf("(-(%s))", inner)
	case ExprAdd:
		s := fmt.Sprintf("%s + %s", exprToString(e.Left, false), exprToString(e.Right, false))
		if parenthesize {
			return "(" + s + ")"
		}
		return s
	case ExprSub:
		s := fmt.Sprintf("%s - %s", exprToString(e.Left, false), exprToStringSubRight(e.Right))
		if parenthesize {
			return "(" + s + ")"
		}
		return s
	case ExprMul:
		left := exprToStringMulOperand(e.Left)
		right := exprToStringMulOperand(e.Right)
		s := fmt.Sprintf("%s*%s", left, right)
		if parenthesize {
			return "(" + s + ")"
		}
		return s
	case ExprDiv:
		left := exprToStringMulOperand(e.Left)
		right := exprToStringMulOperand(e.Right)
		s := fmt.Sprintf("%s/%s", left, right)
		if parenthesize {
			return "(" + s + ")"
		}
		return s
	case ExprPow:
		left := exprToStringAtom(e.Left)
		right := exprToStringAtom(e.Right)
		return fmt.Sprintf("%s^%s", left, right)
	case ExprFunc:
		return fmt.Sprintf("%s(%s)", e.Name, exprToString(e.Arg, false))
	default:
		return "?"
	}
}

// exprToStringSubRight wraps add/sub in parens on the right of a subtraction.
func exprToStringSubRight(e *Expr) string {
	if e.Kind == ExprAdd || e.Kind == ExprSub {
		return "(" + exprToString(e, false) + ")"
	}
	return exprToString(e, false)
}

// exprToStringMulOperand wraps add/sub operands in parens for mul/div context.
func exprToStringMulOperand(e *Expr) string {
	if e.Kind == ExprAdd || e.Kind == ExprSub {
		return "(" + exprToString(e, false) + ")"
	}
	return exprToString(e, false)
}

// exprToStringAtom wraps anything that is not a simple atom in parens for pow context.
func exprToStringAtom(e *Expr) string {
	if e.Kind == ExprNum || e.Kind == ExprVar || e.Kind == ExprFunc {
		return exprToString(e, false)
	}
	return "(" + exprToString(e, false) + ")"
}

// ── Symbolic helpers available to other packages ──

// ExprEqual checks structural equality of two expressions.
func ExprEqual(a, b *Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case ExprNum:
		return a.Value == b.Value
	case ExprVar:
		return a.Name == b.Name
	case ExprNeg:
		return ExprEqual(a.Arg, b.Arg)
	case ExprFunc:
		return a.Name == b.Name && ExprEqual(a.Arg, b.Arg)
	default:
		return ExprEqual(a.Left, b.Left) && ExprEqual(a.Right, b.Right)
	}
}

// ExprSubstitute replaces all occurrences of a variable with another expression.
func ExprSubstitute(e *Expr, varName string, replacement *Expr) *Expr {
	if e == nil {
		return nil
	}
	switch e.Kind {
	case ExprNum:
		return e
	case ExprVar:
		if e.Name == varName {
			return replacement
		}
		return e
	case ExprNeg:
		return &Expr{Kind: ExprNeg, Arg: ExprSubstitute(e.Arg, varName, replacement)}
	case ExprFunc:
		return &Expr{Kind: ExprFunc, Name: e.Name, Arg: ExprSubstitute(e.Arg, varName, replacement)}
	default:
		return &Expr{
			Kind:  e.Kind,
			Left:  ExprSubstitute(e.Left, varName, replacement),
			Right: ExprSubstitute(e.Right, varName, replacement),
		}
	}
}

// ExprRealize evaluates an expression to a float64 given variable bindings.
func ExprRealize(e *Expr, bindings map[string]float64) (float64, error) {
	if e == nil {
		return 0, fmt.Errorf("nil expression")
	}
	switch e.Kind {
	case ExprNum:
		return e.Value, nil
	case ExprVar:
		if v, ok := bindings[e.Name]; ok {
			return v, nil
		}
		return 0, fmt.Errorf("unbound variable: %s", e.Name)
	case ExprNeg:
		v, err := ExprRealize(e.Arg, bindings)
		if err != nil {
			return 0, err
		}
		return -v, nil
	case ExprAdd:
		l, err := ExprRealize(e.Left, bindings)
		if err != nil {
			return 0, err
		}
		r, err := ExprRealize(e.Right, bindings)
		if err != nil {
			return 0, err
		}
		return l + r, nil
	case ExprSub:
		l, err := ExprRealize(e.Left, bindings)
		if err != nil {
			return 0, err
		}
		r, err := ExprRealize(e.Right, bindings)
		if err != nil {
			return 0, err
		}
		return l - r, nil
	case ExprMul:
		l, err := ExprRealize(e.Left, bindings)
		if err != nil {
			return 0, err
		}
		r, err := ExprRealize(e.Right, bindings)
		if err != nil {
			return 0, err
		}
		return l * r, nil
	case ExprDiv:
		l, err := ExprRealize(e.Left, bindings)
		if err != nil {
			return 0, err
		}
		r, err := ExprRealize(e.Right, bindings)
		if err != nil {
			return 0, err
		}
		if r == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return l / r, nil
	case ExprPow:
		l, err := ExprRealize(e.Left, bindings)
		if err != nil {
			return 0, err
		}
		r, err := ExprRealize(e.Right, bindings)
		if err != nil {
			return 0, err
		}
		return math.Pow(l, r), nil
	case ExprFunc:
		arg, err := ExprRealize(e.Arg, bindings)
		if err != nil {
			return 0, err
		}
		switch e.Name {
		case "sin":
			return math.Sin(arg), nil
		case "cos":
			return math.Cos(arg), nil
		case "tan":
			return math.Tan(arg), nil
		case "exp":
			return math.Exp(arg), nil
		case "log", "ln":
			return math.Log(arg), nil
		case "sqrt":
			return math.Sqrt(arg), nil
		case "abs":
			return math.Abs(arg), nil
		case "asin":
			return math.Asin(arg), nil
		case "acos":
			return math.Acos(arg), nil
		case "atan":
			return math.Atan(arg), nil
		default:
			return 0, fmt.Errorf("unknown function: %s", e.Name)
		}
	default:
		return 0, fmt.Errorf("unknown expression kind: %d", e.Kind)
	}
}

// ExprFreeSymbols collects the set of variable names in an expression.
func ExprFreeSymbols(e *Expr) []string {
	seen := make(map[string]bool)
	exprCollectVars(e, seen)
	result := make([]string, 0, len(seen))
	for name := range seen {
		result = append(result, name)
	}
	// Sort for deterministic output.
	sortStrings(result)
	return result
}

func exprCollectVars(e *Expr, seen map[string]bool) {
	if e == nil {
		return
	}
	switch e.Kind {
	case ExprNum:
		// nothing
	case ExprVar:
		seen[e.Name] = true
	case ExprNeg:
		exprCollectVars(e.Arg, seen)
	case ExprFunc:
		exprCollectVars(e.Arg, seen)
	default:
		exprCollectVars(e.Left, seen)
		exprCollectVars(e.Right, seen)
	}
}

// ExprIsConstant returns true if the expression has no free variables.
func ExprIsConstant(e *Expr) bool {
	return len(ExprFreeSymbols(e)) == 0
}

// sortStrings sorts a slice of strings in place.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		key := s[i]
		j := i - 1
		for j >= 0 && strings.Compare(s[j], key) > 0 {
			s[j+1] = s[j]
			j--
		}
		s[j+1] = key
	}
}
