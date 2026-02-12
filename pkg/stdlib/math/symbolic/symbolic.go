// Package symbolic provides symbolic mathematics functions for the GeoFlow standard library.
package symbolic

import (
	"fmt"
	"math"

	"github.com/rogue780/geoflow/internal/object"
)

// GetExports returns all exported functions for the std.math.symbolic module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		"var":        &object.Builtin{Name: "sym.var", Fn: symVar},
		"const":      &object.Builtin{Name: "sym.const", Fn: symConst},
		"num":        &object.Builtin{Name: "sym.num", Fn: symNum},
		"diff":       &object.Builtin{Name: "sym.diff", Fn: symDiff},
		"simplify":   &object.Builtin{Name: "sym.simplify", Fn: symSimplify},
		"expand":     &object.Builtin{Name: "sym.expand", Fn: symExpand},
		"substitute": &object.Builtin{Name: "sym.substitute", Fn: symSubstitute},
		"realize":    &object.Builtin{Name: "sym.realize", Fn: symRealize},
		"toString":   &object.Builtin{Name: "sym.toString", Fn: symToString},
		"add":        &object.Builtin{Name: "sym.add", Fn: symAdd},
		"sub":        &object.Builtin{Name: "sym.sub", Fn: symSub},
		"mul":        &object.Builtin{Name: "sym.mul", Fn: symMul},
		"div":        &object.Builtin{Name: "sym.div", Fn: symDiv},
		"pow":        &object.Builtin{Name: "sym.pow", Fn: symPow},
		"neg":        &object.Builtin{Name: "sym.neg", Fn: symNeg},
		"sin":        &object.Builtin{Name: "sym.sin", Fn: symSin},
		"cos":        &object.Builtin{Name: "sym.cos", Fn: symCos},
		"exp":        &object.Builtin{Name: "sym.exp", Fn: symExp},
		"log":        &object.Builtin{Name: "sym.log", Fn: symLog},
		"sqrt":       &object.Builtin{Name: "sym.sqrt", Fn: symSqrt},
	}
}

// ── Constructors ──

// symVar creates a symbolic variable: sym.var("x") -> Expr
func symVar(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "sym.var expects 1 argument (name)"}
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "sym.var: argument must be a string"}
	}
	return &object.Expr{Kind: object.ExprVar, Name: name.Value}
}

// symConst creates a named constant: sym.const("pi", 3.14159) -> Expr
func symConst(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "sym.const expects 2 arguments (name, value)"}
	}
	_, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "sym.const: first argument must be a string (name)"}
	}
	val := toFloat(args[1])
	if val == nil {
		return &object.Error{Message: "sym.const: second argument must be numeric"}
	}
	return &object.Expr{Kind: object.ExprNum, Value: *val}
}

// symNum creates a numeric literal expression.
func symNum(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "sym.num expects 1 argument (number)"}
	}
	val := toFloat(args[0])
	if val == nil {
		return &object.Error{Message: "sym.num: argument must be numeric"}
	}
	return &object.Expr{Kind: object.ExprNum, Value: *val}
}

// ── Symbolic functions ──

func symSin(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "sym.sin expects 1 argument"}
	}
	e := toExpr(args[0])
	if e == nil {
		return &object.Error{Message: "sym.sin: argument must be an Expr or number"}
	}
	return &object.Expr{Kind: object.ExprFunc, Name: "sin", Arg: e}
}

func symCos(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "sym.cos expects 1 argument"}
	}
	e := toExpr(args[0])
	if e == nil {
		return &object.Error{Message: "sym.cos: argument must be an Expr or number"}
	}
	return &object.Expr{Kind: object.ExprFunc, Name: "cos", Arg: e}
}

func symExp(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "sym.exp expects 1 argument"}
	}
	e := toExpr(args[0])
	if e == nil {
		return &object.Error{Message: "sym.exp: argument must be an Expr or number"}
	}
	return &object.Expr{Kind: object.ExprFunc, Name: "exp", Arg: e}
}

func symLog(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "sym.log expects 1 argument"}
	}
	e := toExpr(args[0])
	if e == nil {
		return &object.Error{Message: "sym.log: argument must be an Expr or number"}
	}
	return &object.Expr{Kind: object.ExprFunc, Name: "log", Arg: e}
}

func symSqrt(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "sym.sqrt expects 1 argument"}
	}
	e := toExpr(args[0])
	if e == nil {
		return &object.Error{Message: "sym.sqrt: argument must be an Expr or number"}
	}
	// sqrt(x) = x^0.5
	return &object.Expr{Kind: object.ExprPow, Left: e, Right: &object.Expr{Kind: object.ExprNum, Value: 0.5}}
}

// ── Binary constructors ──

func symAdd(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "sym.add expects 2 arguments"}
	}
	l := toExpr(args[0])
	r := toExpr(args[1])
	if l == nil || r == nil {
		return &object.Error{Message: "sym.add: arguments must be Expr or numeric"}
	}
	return &object.Expr{Kind: object.ExprAdd, Left: l, Right: r}
}

func symSub(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "sym.sub expects 2 arguments"}
	}
	l := toExpr(args[0])
	r := toExpr(args[1])
	if l == nil || r == nil {
		return &object.Error{Message: "sym.sub: arguments must be Expr or numeric"}
	}
	return &object.Expr{Kind: object.ExprSub, Left: l, Right: r}
}

func symMul(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "sym.mul expects 2 arguments"}
	}
	l := toExpr(args[0])
	r := toExpr(args[1])
	if l == nil || r == nil {
		return &object.Error{Message: "sym.mul: arguments must be Expr or numeric"}
	}
	return &object.Expr{Kind: object.ExprMul, Left: l, Right: r}
}

func symDiv(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "sym.div expects 2 arguments"}
	}
	l := toExpr(args[0])
	r := toExpr(args[1])
	if l == nil || r == nil {
		return &object.Error{Message: "sym.div: arguments must be Expr or numeric"}
	}
	return &object.Expr{Kind: object.ExprDiv, Left: l, Right: r}
}

func symPow(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "sym.pow expects 2 arguments"}
	}
	l := toExpr(args[0])
	r := toExpr(args[1])
	if l == nil || r == nil {
		return &object.Error{Message: "sym.pow: arguments must be Expr or numeric"}
	}
	return &object.Expr{Kind: object.ExprPow, Left: l, Right: r}
}

func symNeg(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "sym.neg expects 1 argument"}
	}
	e := toExpr(args[0])
	if e == nil {
		return &object.Error{Message: "sym.neg: argument must be Expr or numeric"}
	}
	return &object.Expr{Kind: object.ExprNeg, Arg: e}
}

// ── Differentiation ──

// symDiff differentiates an expression with respect to a variable.
// sym.diff(expr, "x") -> Expr
func symDiff(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "sym.diff expects 2 arguments (expr, variable)"}
	}
	expr, ok := args[0].(*object.Expr)
	if !ok {
		return &object.Error{Message: "sym.diff: first argument must be an Expr"}
	}
	varName, ok := args[1].(*object.String)
	if !ok {
		return &object.Error{Message: "sym.diff: second argument must be a string (variable name)"}
	}
	result := differentiate(expr, varName.Value)
	return simplify(result)
}

// differentiate computes the symbolic derivative of expr with respect to varName.
func differentiate(e *object.Expr, v string) *object.Expr {
	if e == nil {
		return numExpr(0)
	}
	switch e.Kind {
	case object.ExprNum:
		// d/dx(c) = 0
		return numExpr(0)

	case object.ExprVar:
		if e.Name == v {
			// d/dx(x) = 1
			return numExpr(1)
		}
		// d/dx(y) = 0
		return numExpr(0)

	case object.ExprNeg:
		// d/dx(-f) = -(d/dx(f))
		return &object.Expr{Kind: object.ExprNeg, Arg: differentiate(e.Arg, v)}

	case object.ExprAdd:
		// d/dx(f+g) = f' + g'
		return &object.Expr{
			Kind:  object.ExprAdd,
			Left:  differentiate(e.Left, v),
			Right: differentiate(e.Right, v),
		}

	case object.ExprSub:
		// d/dx(f-g) = f' - g'
		return &object.Expr{
			Kind:  object.ExprSub,
			Left:  differentiate(e.Left, v),
			Right: differentiate(e.Right, v),
		}

	case object.ExprMul:
		// Product rule: d/dx(f*g) = f'*g + f*g'
		return &object.Expr{
			Kind: object.ExprAdd,
			Left: &object.Expr{
				Kind:  object.ExprMul,
				Left:  differentiate(e.Left, v),
				Right: e.Right,
			},
			Right: &object.Expr{
				Kind:  object.ExprMul,
				Left:  e.Left,
				Right: differentiate(e.Right, v),
			},
		}

	case object.ExprDiv:
		// Quotient rule: d/dx(f/g) = (f'*g - f*g') / g^2
		return &object.Expr{
			Kind: object.ExprDiv,
			Left: &object.Expr{
				Kind: object.ExprSub,
				Left: &object.Expr{
					Kind:  object.ExprMul,
					Left:  differentiate(e.Left, v),
					Right: e.Right,
				},
				Right: &object.Expr{
					Kind:  object.ExprMul,
					Left:  e.Left,
					Right: differentiate(e.Right, v),
				},
			},
			Right: &object.Expr{
				Kind:  object.ExprPow,
				Left:  e.Right,
				Right: numExpr(2),
			},
		}

	case object.ExprPow:
		// If exponent is constant (no v), use generalized power rule:
		// d/dx(f^n) = n * f^(n-1) * f'
		if !containsVar(e.Right, v) {
			return &object.Expr{
				Kind: object.ExprMul,
				Left: &object.Expr{
					Kind: object.ExprMul,
					Left: e.Right, // n
					Right: &object.Expr{
						Kind: object.ExprPow,
						Left: e.Left, // f
						Right: &object.Expr{
							Kind:  object.ExprSub,
							Left:  e.Right,
							Right: numExpr(1),
						}, // n-1
					},
				},
				Right: differentiate(e.Left, v), // f'
			}
		}
		// If base is constant (no v) and exponent has v: d/dx(a^g) = a^g * log(a) * g'
		if !containsVar(e.Left, v) {
			return &object.Expr{
				Kind: object.ExprMul,
				Left: &object.Expr{
					Kind: object.ExprMul,
					Left: e, // a^g
					Right: &object.Expr{
						Kind: object.ExprFunc,
						Name: "log",
						Arg:  e.Left, // log(a)
					},
				},
				Right: differentiate(e.Right, v), // g'
			}
		}
		// General case: d/dx(f^g) = f^g * (g'*log(f) + g*f'/f)
		return &object.Expr{
			Kind: object.ExprMul,
			Left: e,
			Right: &object.Expr{
				Kind: object.ExprAdd,
				Left: &object.Expr{
					Kind:  object.ExprMul,
					Left:  differentiate(e.Right, v),
					Right: &object.Expr{Kind: object.ExprFunc, Name: "log", Arg: e.Left},
				},
				Right: &object.Expr{
					Kind: object.ExprMul,
					Left: e.Right,
					Right: &object.Expr{
						Kind:  object.ExprDiv,
						Left:  differentiate(e.Left, v),
						Right: e.Left,
					},
				},
			},
		}

	case object.ExprFunc:
		// Chain rule: d/dx(f(g)) = f'(g) * g'
		inner := e.Arg
		innerDeriv := differentiate(inner, v)

		var outerDeriv *object.Expr
		switch e.Name {
		case "sin":
			// d/dx(sin(g)) = cos(g) * g'
			outerDeriv = &object.Expr{Kind: object.ExprFunc, Name: "cos", Arg: inner}
		case "cos":
			// d/dx(cos(g)) = -sin(g) * g'
			outerDeriv = &object.Expr{
				Kind: object.ExprNeg,
				Arg:  &object.Expr{Kind: object.ExprFunc, Name: "sin", Arg: inner},
			}
		case "tan":
			// d/dx(tan(g)) = (1/cos(g)^2) * g'
			outerDeriv = &object.Expr{
				Kind: object.ExprDiv,
				Left: numExpr(1),
				Right: &object.Expr{
					Kind:  object.ExprPow,
					Left:  &object.Expr{Kind: object.ExprFunc, Name: "cos", Arg: inner},
					Right: numExpr(2),
				},
			}
		case "exp":
			// d/dx(exp(g)) = exp(g) * g'
			outerDeriv = &object.Expr{Kind: object.ExprFunc, Name: "exp", Arg: inner}
		case "log", "ln":
			// d/dx(log(g)) = g'/g
			return &object.Expr{
				Kind:  object.ExprDiv,
				Left:  innerDeriv,
				Right: inner,
			}
		case "sqrt":
			// d/dx(sqrt(g)) = g'/(2*sqrt(g))
			return &object.Expr{
				Kind:  object.ExprDiv,
				Left:  innerDeriv,
				Right: &object.Expr{Kind: object.ExprMul, Left: numExpr(2), Right: &object.Expr{Kind: object.ExprFunc, Name: "sqrt", Arg: inner}},
			}
		case "abs":
			// d/dx(abs(g)) = g/abs(g) * g' (sign function)
			outerDeriv = &object.Expr{
				Kind:  object.ExprDiv,
				Left:  inner,
				Right: &object.Expr{Kind: object.ExprFunc, Name: "abs", Arg: inner},
			}
		case "asin":
			// d/dx(asin(g)) = g'/sqrt(1-g^2)
			return &object.Expr{
				Kind:  object.ExprDiv,
				Left:  innerDeriv,
				Right: &object.Expr{Kind: object.ExprFunc, Name: "sqrt", Arg: &object.Expr{Kind: object.ExprSub, Left: numExpr(1), Right: &object.Expr{Kind: object.ExprPow, Left: inner, Right: numExpr(2)}}},
			}
		case "acos":
			// d/dx(acos(g)) = -g'/sqrt(1-g^2)
			return &object.Expr{
				Kind: object.ExprNeg,
				Arg: &object.Expr{
					Kind:  object.ExprDiv,
					Left:  innerDeriv,
					Right: &object.Expr{Kind: object.ExprFunc, Name: "sqrt", Arg: &object.Expr{Kind: object.ExprSub, Left: numExpr(1), Right: &object.Expr{Kind: object.ExprPow, Left: inner, Right: numExpr(2)}}},
				},
			}
		case "atan":
			// d/dx(atan(g)) = g'/(1+g^2)
			return &object.Expr{
				Kind:  object.ExprDiv,
				Left:  innerDeriv,
				Right: &object.Expr{Kind: object.ExprAdd, Left: numExpr(1), Right: &object.Expr{Kind: object.ExprPow, Left: inner, Right: numExpr(2)}},
			}
		default:
			return numExpr(0) // unknown function, treat as constant
		}

		return &object.Expr{
			Kind:  object.ExprMul,
			Left:  outerDeriv,
			Right: innerDeriv,
		}
	}

	return numExpr(0)
}

// containsVar checks if an expression contains a particular variable.
func containsVar(e *object.Expr, v string) bool {
	if e == nil {
		return false
	}
	switch e.Kind {
	case object.ExprNum:
		return false
	case object.ExprVar:
		return e.Name == v
	case object.ExprNeg:
		return containsVar(e.Arg, v)
	case object.ExprFunc:
		return containsVar(e.Arg, v)
	default:
		return containsVar(e.Left, v) || containsVar(e.Right, v)
	}
}

// ── Simplification ──

// symSimplify simplifies an expression.
func symSimplify(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "sym.simplify expects 1 argument (expr)"}
	}
	expr, ok := args[0].(*object.Expr)
	if !ok {
		return &object.Error{Message: "sym.simplify: argument must be an Expr"}
	}
	return simplify(expr)
}

// simplify applies basic algebraic simplification rules.
func simplify(e *object.Expr) *object.Expr {
	if e == nil {
		return nil
	}

	switch e.Kind {
	case object.ExprNum, object.ExprVar:
		return e

	case object.ExprNeg:
		arg := simplify(e.Arg)
		// -0 = 0
		if isZero(arg) {
			return numExpr(0)
		}
		// -(-x) = x
		if arg.Kind == object.ExprNeg {
			return arg.Arg
		}
		// -(num) = -num
		if arg.Kind == object.ExprNum {
			return numExpr(-arg.Value)
		}
		return &object.Expr{Kind: object.ExprNeg, Arg: arg}

	case object.ExprAdd:
		left := simplify(e.Left)
		right := simplify(e.Right)
		// 0 + x = x
		if isZero(left) {
			return right
		}
		// x + 0 = x
		if isZero(right) {
			return left
		}
		// num + num = num
		if left.Kind == object.ExprNum && right.Kind == object.ExprNum {
			return numExpr(left.Value + right.Value)
		}
		return &object.Expr{Kind: object.ExprAdd, Left: left, Right: right}

	case object.ExprSub:
		left := simplify(e.Left)
		right := simplify(e.Right)
		// x - 0 = x
		if isZero(right) {
			return left
		}
		// 0 - x = -x
		if isZero(left) {
			return simplify(&object.Expr{Kind: object.ExprNeg, Arg: right})
		}
		// x - x = 0 (structural equality)
		if object.ExprEqual(left, right) {
			return numExpr(0)
		}
		// num - num = num
		if left.Kind == object.ExprNum && right.Kind == object.ExprNum {
			return numExpr(left.Value - right.Value)
		}
		return &object.Expr{Kind: object.ExprSub, Left: left, Right: right}

	case object.ExprMul:
		left := simplify(e.Left)
		right := simplify(e.Right)
		// 0 * x = 0
		if isZero(left) || isZero(right) {
			return numExpr(0)
		}
		// 1 * x = x
		if isOne(left) {
			return right
		}
		// x * 1 = x
		if isOne(right) {
			return left
		}
		// (-1) * x = -x
		if left.Kind == object.ExprNum && left.Value == -1 {
			return simplify(&object.Expr{Kind: object.ExprNeg, Arg: right})
		}
		// num * num = num
		if left.Kind == object.ExprNum && right.Kind == object.ExprNum {
			return numExpr(left.Value * right.Value)
		}
		return &object.Expr{Kind: object.ExprMul, Left: left, Right: right}

	case object.ExprDiv:
		left := simplify(e.Left)
		right := simplify(e.Right)
		// 0 / x = 0
		if isZero(left) {
			return numExpr(0)
		}
		// x / 1 = x
		if isOne(right) {
			return left
		}
		// x / x = 1 (structural equality)
		if object.ExprEqual(left, right) {
			return numExpr(1)
		}
		// num / num = num
		if left.Kind == object.ExprNum && right.Kind == object.ExprNum && right.Value != 0 {
			return numExpr(left.Value / right.Value)
		}
		return &object.Expr{Kind: object.ExprDiv, Left: left, Right: right}

	case object.ExprPow:
		left := simplify(e.Left)
		right := simplify(e.Right)
		// x^0 = 1
		if isZero(right) {
			return numExpr(1)
		}
		// x^1 = x
		if isOne(right) {
			return left
		}
		// 0^n = 0 (for n > 0)
		if isZero(left) && right.Kind == object.ExprNum && right.Value > 0 {
			return numExpr(0)
		}
		// 1^n = 1
		if isOne(left) {
			return numExpr(1)
		}
		// num^num = num
		if left.Kind == object.ExprNum && right.Kind == object.ExprNum {
			return numExpr(math.Pow(left.Value, right.Value))
		}
		return &object.Expr{Kind: object.ExprPow, Left: left, Right: right}

	case object.ExprFunc:
		arg := simplify(e.Arg)
		// If the argument is a constant, evaluate
		if arg.Kind == object.ExprNum {
			val, err := object.ExprRealize(&object.Expr{Kind: object.ExprFunc, Name: e.Name, Arg: arg}, nil)
			if err == nil {
				return numExpr(val)
			}
		}
		return &object.Expr{Kind: object.ExprFunc, Name: e.Name, Arg: arg}
	}

	return e
}

// ── Expand ──

// symExpand expands products and powers.
func symExpand(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "sym.expand expects 1 argument (expr)"}
	}
	expr, ok := args[0].(*object.Expr)
	if !ok {
		return &object.Error{Message: "sym.expand: argument must be an Expr"}
	}
	return simplify(expand(expr))
}

// expand distributes multiplication over addition.
func expand(e *object.Expr) *object.Expr {
	if e == nil {
		return nil
	}
	switch e.Kind {
	case object.ExprNum, object.ExprVar:
		return e
	case object.ExprNeg:
		return &object.Expr{Kind: object.ExprNeg, Arg: expand(e.Arg)}
	case object.ExprFunc:
		return &object.Expr{Kind: object.ExprFunc, Name: e.Name, Arg: expand(e.Arg)}
	case object.ExprAdd:
		return &object.Expr{Kind: object.ExprAdd, Left: expand(e.Left), Right: expand(e.Right)}
	case object.ExprSub:
		return &object.Expr{Kind: object.ExprSub, Left: expand(e.Left), Right: expand(e.Right)}
	case object.ExprDiv:
		return &object.Expr{Kind: object.ExprDiv, Left: expand(e.Left), Right: expand(e.Right)}
	case object.ExprPow:
		return &object.Expr{Kind: object.ExprPow, Left: expand(e.Left), Right: expand(e.Right)}
	case object.ExprMul:
		left := expand(e.Left)
		right := expand(e.Right)
		return distributeProduct(left, right)
	}
	return e
}

// distributeProduct distributes a*b where a or b may be sums.
// (a+b)*(c+d) = a*c + a*d + b*c + b*d
func distributeProduct(a, b *object.Expr) *object.Expr {
	// If a is a sum: (a1+a2)*b = a1*b + a2*b
	if a.Kind == object.ExprAdd {
		left := distributeProduct(a.Left, b)
		right := distributeProduct(a.Right, b)
		return &object.Expr{Kind: object.ExprAdd, Left: left, Right: right}
	}
	// If a is a difference: (a1-a2)*b = a1*b - a2*b
	if a.Kind == object.ExprSub {
		left := distributeProduct(a.Left, b)
		right := distributeProduct(a.Right, b)
		return &object.Expr{Kind: object.ExprSub, Left: left, Right: right}
	}
	// If b is a sum: a*(b1+b2) = a*b1 + a*b2
	if b.Kind == object.ExprAdd {
		left := distributeProduct(a, b.Left)
		right := distributeProduct(a, b.Right)
		return &object.Expr{Kind: object.ExprAdd, Left: left, Right: right}
	}
	// If b is a difference: a*(b1-b2) = a*b1 - a*b2
	if b.Kind == object.ExprSub {
		left := distributeProduct(a, b.Left)
		right := distributeProduct(a, b.Right)
		return &object.Expr{Kind: object.ExprSub, Left: left, Right: right}
	}
	return &object.Expr{Kind: object.ExprMul, Left: a, Right: b}
}

// ── Substitution ──

// symSubstitute replaces a variable in an expression with a value.
// sym.substitute(expr, "x", 5.0) -> Expr
func symSubstitute(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "sym.substitute expects 3 arguments (expr, varName, value)"}
	}
	expr, ok := args[0].(*object.Expr)
	if !ok {
		return &object.Error{Message: "sym.substitute: first argument must be an Expr"}
	}
	varName, ok := args[1].(*object.String)
	if !ok {
		return &object.Error{Message: "sym.substitute: second argument must be a string"}
	}
	replacement := toExpr(args[2])
	if replacement == nil {
		return &object.Error{Message: "sym.substitute: third argument must be Expr or numeric"}
	}
	return simplify(object.ExprSubstitute(expr, varName.Value, replacement))
}

// ── Realize ──

// symRealize evaluates an expression to a float given variable bindings.
// sym.realize(expr, {"x": 5.0}) -> float
func symRealize(args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return &object.Error{Message: "sym.realize expects 1-2 arguments (expr[, bindings])"}
	}
	expr, ok := args[0].(*object.Expr)
	if !ok {
		return &object.Error{Message: "sym.realize: first argument must be an Expr"}
	}

	bindings := make(map[string]float64)
	if len(args) == 2 {
		m, ok := args[1].(*object.Map)
		if !ok {
			return &object.Error{Message: "sym.realize: second argument must be a Map of bindings"}
		}
		for _, pair := range m.Pairs {
			key, ok := pair.Key.(*object.String)
			if !ok {
				return &object.Error{Message: "sym.realize: binding keys must be strings"}
			}
			val := toFloat(pair.Value)
			if val == nil {
				return &object.Error{Message: fmt.Sprintf("sym.realize: binding value for '%s' must be numeric", key.Value)}
			}
			bindings[key.Value] = *val
		}
	}

	result, err := object.ExprRealize(expr, bindings)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("sym.realize: %s", err.Error())}
	}
	return &object.Float{Value: result}
}

// ── ToString ──

func symToString(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "sym.toString expects 1 argument (expr)"}
	}
	expr, ok := args[0].(*object.Expr)
	if !ok {
		return &object.Error{Message: "sym.toString: argument must be an Expr"}
	}
	return &object.String{Value: expr.Inspect()}
}

// ── Helpers ──

func numExpr(v float64) *object.Expr {
	return &object.Expr{Kind: object.ExprNum, Value: v}
}

func isZero(e *object.Expr) bool {
	return e.Kind == object.ExprNum && e.Value == 0
}

func isOne(e *object.Expr) bool {
	return e.Kind == object.ExprNum && e.Value == 1
}

// toExpr converts an object.Object to an *object.Expr (wrapping numbers).
func toExpr(o object.Object) *object.Expr {
	switch v := o.(type) {
	case *object.Expr:
		return v
	case *object.Integer:
		return &object.Expr{Kind: object.ExprNum, Value: float64(v.Value)}
	case *object.Float:
		return &object.Expr{Kind: object.ExprNum, Value: v.Value}
	default:
		return nil
	}
}

// toFloat converts a numeric object to a *float64.
func toFloat(o object.Object) *float64 {
	switch v := o.(type) {
	case *object.Integer:
		f := float64(v.Value)
		return &f
	case *object.Float:
		return &v.Value
	default:
		return nil
	}
}
