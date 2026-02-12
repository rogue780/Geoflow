package object

import (
	"fmt"
	"strings"
	"unicode"
)

// ParseSymbolicExpr parses a symbolic expression string (the content between
// $...$ delimiters) into an *Expr tree. Returns (expr, "") on success or
// (nil, errorMsg) on failure.
//
// Grammar (precedence low→high):
//
//	Expression = Term (('+' | '-') Term)*
//	Term       = Power (('*' | '/') Power)*
//	Power      = Unary ('^' Power)?          // right-associative
//	Unary      = '-' Unary | Atom
//	Atom       = Number | FunctionCall | Variable | '(' Expression ')'
//
// Supported functions: sin, cos, tan, exp, log, ln, sqrt, abs, asin, acos, atan
func ParseSymbolicExpr(input string) (*Expr, string) {
	p := &symParser{input: strings.TrimSpace(input), pos: 0}
	expr := p.parseExpression()
	if p.err != "" {
		return nil, p.err
	}
	p.skipSpaces()
	if p.pos < len(p.input) {
		return nil, fmt.Sprintf("unexpected character '%c' at position %d", p.input[p.pos], p.pos)
	}
	return expr, ""
}

// symParser is a simple recursive descent parser for symbolic math expressions.
type symParser struct {
	input string
	pos   int
	err   string
}

var symFunctions = map[string]bool{
	"sin":  true,
	"cos":  true,
	"tan":  true,
	"exp":  true,
	"log":  true,
	"ln":   true,
	"sqrt": true,
	"abs":  true,
	"asin": true,
	"acos": true,
	"atan": true,
}

func (p *symParser) skipSpaces() {
	for p.pos < len(p.input) && p.input[p.pos] == ' ' {
		p.pos++
	}
}

func (p *symParser) peek() byte {
	p.skipSpaces()
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *symParser) advance() {
	p.pos++
}

// parseExpression: Term (('+' | '-') Term)*
func (p *symParser) parseExpression() *Expr {
	left := p.parseTerm()
	if p.err != "" {
		return nil
	}
	for {
		ch := p.peek()
		if ch != '+' && ch != '-' {
			break
		}
		p.advance()
		right := p.parseTerm()
		if p.err != "" {
			return nil
		}
		if ch == '+' {
			left = &Expr{Kind: ExprAdd, Left: left, Right: right}
		} else {
			left = &Expr{Kind: ExprSub, Left: left, Right: right}
		}
	}
	return left
}

// parseTerm: Power (('*' | '/') Power)*
func (p *symParser) parseTerm() *Expr {
	left := p.parsePower()
	if p.err != "" {
		return nil
	}
	for {
		ch := p.peek()
		if ch != '*' && ch != '/' {
			break
		}
		p.advance()
		right := p.parsePower()
		if p.err != "" {
			return nil
		}
		if ch == '*' {
			left = &Expr{Kind: ExprMul, Left: left, Right: right}
		} else {
			left = &Expr{Kind: ExprDiv, Left: left, Right: right}
		}
	}
	return left
}

// parsePower: Unary ('^' Power)?   — right-associative
func (p *symParser) parsePower() *Expr {
	base := p.parseUnary()
	if p.err != "" {
		return nil
	}
	if p.peek() == '^' {
		p.advance()
		exp := p.parsePower() // right-recursive for right-associativity
		if p.err != "" {
			return nil
		}
		return &Expr{Kind: ExprPow, Left: base, Right: exp}
	}
	return base
}

// parseUnary: '-' Unary | Atom
func (p *symParser) parseUnary() *Expr {
	if p.peek() == '-' {
		p.advance()
		operand := p.parseUnary()
		if p.err != "" {
			return nil
		}
		return &Expr{Kind: ExprNeg, Arg: operand}
	}
	return p.parseAtom()
}

// parseAtom: Number | FunctionCall | Variable | '(' Expression ')'
func (p *symParser) parseAtom() *Expr {
	p.skipSpaces()
	if p.pos >= len(p.input) {
		p.err = "unexpected end of expression"
		return nil
	}

	ch := p.input[p.pos]

	// Parenthesized expression
	if ch == '(' {
		p.advance()
		expr := p.parseExpression()
		if p.err != "" {
			return nil
		}
		if p.peek() != ')' {
			p.err = "expected closing ')'"
			return nil
		}
		p.advance()
		return expr
	}

	// Number: digit or '.'
	if ch >= '0' && ch <= '9' || ch == '.' {
		return p.parseNumber()
	}

	// Identifier: variable or function call
	if ch == '_' || unicode.IsLetter(rune(ch)) {
		return p.parseIdentOrFunc()
	}

	p.err = fmt.Sprintf("unexpected character '%c' at position %d", ch, p.pos)
	return nil
}

// parseNumber parses an integer or floating-point literal.
func (p *symParser) parseNumber() *Expr {
	start := p.pos
	hasDot := false
	hasE := false

	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch >= '0' && ch <= '9' {
			p.pos++
		} else if ch == '.' && !hasDot && !hasE {
			hasDot = true
			p.pos++
		} else if (ch == 'e' || ch == 'E') && !hasE {
			hasE = true
			p.pos++
			// optional sign after exponent
			if p.pos < len(p.input) && (p.input[p.pos] == '+' || p.input[p.pos] == '-') {
				p.pos++
			}
		} else {
			break
		}
	}

	numStr := p.input[start:p.pos]
	var val float64
	_, err := fmt.Sscanf(numStr, "%g", &val)
	if err != nil {
		p.err = fmt.Sprintf("invalid number: %s", numStr)
		return nil
	}
	return &Expr{Kind: ExprNum, Value: val}
}

// parseIdentOrFunc parses a variable name or a function call (name followed by '(').
func (p *symParser) parseIdentOrFunc() *Expr {
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == '_' || unicode.IsLetter(rune(ch)) || (ch >= '0' && ch <= '9') {
			p.pos++
		} else {
			break
		}
	}
	name := p.input[start:p.pos]

	// Check for function call
	if p.peek() == '(' && symFunctions[name] {
		p.advance() // consume '('
		arg := p.parseExpression()
		if p.err != "" {
			return nil
		}
		if p.peek() != ')' {
			p.err = fmt.Sprintf("expected ')' after %s argument", name)
			return nil
		}
		p.advance() // consume ')'
		return &Expr{Kind: ExprFunc, Name: name, Arg: arg}
	}

	return &Expr{Kind: ExprVar, Name: name}
}
