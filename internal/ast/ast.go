// Package ast defines the abstract syntax tree for the GeoFlow language.
package ast

import (
	"fmt"
	"strings"

	"github.com/rogue780/geoflow/internal/token"
)

// Node is the interface all AST nodes implement.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement nodes implement this interface.
type Statement interface {
	Node
	statementNode()
}

// Expression nodes implement this interface.
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of every AST.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var sb strings.Builder
	for _, s := range p.Statements {
		sb.WriteString(s.String())
		sb.WriteString("\n")
	}
	return sb.String()
}

// ---------- Statements ----------

// LetStatement represents: let [mut] name [: type] = value
type LetStatement struct {
	Token   token.Token // the LET token
	Name    *Identifier
	Mutable bool
	TypeAnn *TypeAnnotation // optional type annotation
	Value   Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var sb strings.Builder
	sb.WriteString("let ")
	if ls.Mutable {
		sb.WriteString("mut ")
	}
	sb.WriteString(ls.Name.String())
	if ls.TypeAnn != nil {
		sb.WriteString(": ")
		sb.WriteString(ls.TypeAnn.String())
	}
	sb.WriteString(" = ")
	sb.WriteString(ls.Value.String())
	return sb.String()
}

// ConstStatement represents: const name = value
type ConstStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (cs *ConstStatement) statementNode()       {}
func (cs *ConstStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ConstStatement) String() string {
	return fmt.Sprintf("const %s = %s", cs.Name.String(), cs.Value.String())
}

// ExpressionStatement wraps an expression as a statement.
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// ReturnStatement represents: return [expr]
type ReturnStatement struct {
	Token token.Token
	Value Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	if rs.Value != nil {
		return fmt.Sprintf("return %s", rs.Value.String())
	}
	return "return"
}

// BreakStatement represents: break
type BreakStatement struct {
	Token token.Token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string       { return "break" }

// ContinueStatement represents: continue
type ContinueStatement struct {
	Token token.Token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) String() string       { return "continue" }

// ImportStatement represents: import path [as alias]
type ImportStatement struct {
	Token token.Token
	Path  []string // e.g. ["std", "math"]
	Alias string   // optional alias
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string {
	path := strings.Join(is.Path, ".")
	if is.Alias != "" {
		return fmt.Sprintf("import %s as %s", path, is.Alias)
	}
	return fmt.Sprintf("import %s", path)
}

// AssignStatement represents: name := value (mutable reassignment)
type AssignStatement struct {
	Token token.Token
	Name  Expression // can be identifier, index expr, or field access
	Value Expression
}

func (as *AssignStatement) statementNode()       {}
func (as *AssignStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignStatement) String() string {
	return fmt.Sprintf("%s := %s", as.Name.String(), as.Value.String())
}

// ---------- Expressions ----------

// Identifier represents a name reference.
type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral represents an integer value.
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral represents a floating-point value.
type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral represents a string value.
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return fmt.Sprintf("%q", sl.Value) }

// BooleanLiteral represents true or false.
type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

// NilLiteral represents the nil value.
type NilLiteral struct {
	Token token.Token
}

func (nl *NilLiteral) expressionNode()      {}
func (nl *NilLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NilLiteral) String() string       { return "nil" }

// WKTLiteral represents a WKT geometry literal #POINT(0 0)#
type WKTLiteral struct {
	Token token.Token
	Value string
}

func (wl *WKTLiteral) expressionNode()      {}
func (wl *WKTLiteral) TokenLiteral() string { return wl.Token.Literal }
func (wl *WKTLiteral) String() string       { return fmt.Sprintf("#%s#", wl.Value) }

// SymbolicLiteral represents a symbolic math expression $x^2 + 1$
type SymbolicLiteral struct {
	Token token.Token
	Value string
}

func (sl *SymbolicLiteral) expressionNode()      {}
func (sl *SymbolicLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *SymbolicLiteral) String() string       { return fmt.Sprintf("$%s$", sl.Value) }

// PrefixExpression represents a prefix operation like -x or !x.
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	return fmt.Sprintf("(%s%s)", pe.Operator, pe.Right.String())
}

// InfixExpression represents a binary operation like x + y.
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", ie.Left.String(), ie.Operator, ie.Right.String())
}

// PipelineExpression represents: expr |> expr
type PipelineExpression struct {
	Token token.Token
	Left  Expression
	Right Expression
}

func (pe *PipelineExpression) expressionNode()      {}
func (pe *PipelineExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PipelineExpression) String() string {
	return fmt.Sprintf("(%s |> %s)", pe.Left.String(), pe.Right.String())
}

// IfExpression represents: if cond then consequence else alternative
type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence Expression
	Alternative Expression // optional
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	s := fmt.Sprintf("if %s then %s", ie.Condition.String(), ie.Consequence.String())
	if ie.Alternative != nil {
		s += fmt.Sprintf(" else %s", ie.Alternative.String())
	}
	return s
}

// BlockExpression represents: { statements; [expr] }
type BlockExpression struct {
	Token      token.Token
	Statements []Statement
	Value      Expression // optional final expression (return value)
}

func (be *BlockExpression) expressionNode()      {}
func (be *BlockExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BlockExpression) String() string {
	var sb strings.Builder
	sb.WriteString("{ ")
	for _, s := range be.Statements {
		sb.WriteString(s.String())
		sb.WriteString("; ")
	}
	if be.Value != nil {
		sb.WriteString(be.Value.String())
	}
	sb.WriteString(" }")
	return sb.String()
}

// FunctionLiteral represents: fn name<T>(params) -> retType { body }
type FunctionLiteral struct {
	Token      token.Token
	Name       string // empty for anonymous
	TypeParams []string
	Parameters []*FunctionParameter
	ReturnType *TypeAnnotation
	Body       *BlockExpression
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var sb strings.Builder
	sb.WriteString("fn")
	if fl.Name != "" {
		sb.WriteString(" ")
		sb.WriteString(fl.Name)
	}
	sb.WriteString("(")
	params := make([]string, len(fl.Parameters))
	for i, p := range fl.Parameters {
		params[i] = p.String()
	}
	sb.WriteString(strings.Join(params, ", "))
	sb.WriteString(")")
	if fl.ReturnType != nil {
		sb.WriteString(" -> ")
		sb.WriteString(fl.ReturnType.String())
	}
	sb.WriteString(" ")
	sb.WriteString(fl.Body.String())
	return sb.String()
}

// FunctionParameter represents a single function parameter.
type FunctionParameter struct {
	Name     *Identifier
	TypeAnn  *TypeAnnotation
	Default  Expression // optional default value
	Variadic bool
}

func (fp *FunctionParameter) String() string {
	var sb strings.Builder
	if fp.Variadic {
		sb.WriteString("...")
	}
	sb.WriteString(fp.Name.Value)
	if fp.TypeAnn != nil {
		sb.WriteString(": ")
		sb.WriteString(fp.TypeAnn.String())
	}
	if fp.Default != nil {
		sb.WriteString(" = ")
		sb.WriteString(fp.Default.String())
	}
	return sb.String()
}

// LambdaExpression represents: \params -> expr
type LambdaExpression struct {
	Token      token.Token
	Parameters []*Identifier
	Body       Expression
}

func (le *LambdaExpression) expressionNode()      {}
func (le *LambdaExpression) TokenLiteral() string { return le.Token.Literal }
func (le *LambdaExpression) String() string {
	params := make([]string, len(le.Parameters))
	for i, p := range le.Parameters {
		params[i] = p.Value
	}
	return fmt.Sprintf("\\%s -> %s", strings.Join(params, ", "), le.Body.String())
}

// CallExpression represents: function(args)
type CallExpression struct {
	Token     token.Token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	args := make([]string, len(ce.Arguments))
	for i, a := range ce.Arguments {
		args[i] = a.String()
	}
	return fmt.Sprintf("%s(%s)", ce.Function.String(), strings.Join(args, ", "))
}

// IndexExpression represents: expr[index]
type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	return fmt.Sprintf("(%s[%s])", ie.Left.String(), ie.Index.String())
}

// DotExpression represents: expr.field
type DotExpression struct {
	Token token.Token
	Left  Expression
	Field string
}

func (de *DotExpression) expressionNode()      {}
func (de *DotExpression) TokenLiteral() string { return de.Token.Literal }
func (de *DotExpression) String() string {
	return fmt.Sprintf("%s.%s", de.Left.String(), de.Field)
}

// ListLiteral represents: [elem1, elem2, ...]
type ListLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (ll *ListLiteral) expressionNode()      {}
func (ll *ListLiteral) TokenLiteral() string { return ll.Token.Literal }
func (ll *ListLiteral) String() string {
	elems := make([]string, len(ll.Elements))
	for i, e := range ll.Elements {
		elems[i] = e.String()
	}
	return fmt.Sprintf("[%s]", strings.Join(elems, ", "))
}

// MapLiteral represents: {key: value, ...}
type MapLiteral struct {
	Token token.Token
	Pairs map[Expression]Expression
	Order []Expression // preserve insertion order
}

func (ml *MapLiteral) expressionNode()      {}
func (ml *MapLiteral) TokenLiteral() string { return ml.Token.Literal }
func (ml *MapLiteral) String() string {
	pairs := make([]string, 0, len(ml.Pairs))
	for _, k := range ml.Order {
		v := ml.Pairs[k]
		pairs = append(pairs, fmt.Sprintf("%s: %s", k.String(), v.String()))
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

// TupleLiteral represents: (expr1, expr2, ...)
type TupleLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (tl *TupleLiteral) expressionNode()      {}
func (tl *TupleLiteral) TokenLiteral() string { return tl.Token.Literal }
func (tl *TupleLiteral) String() string {
	elems := make([]string, len(tl.Elements))
	for i, e := range tl.Elements {
		elems[i] = e.String()
	}
	return fmt.Sprintf("(%s)", strings.Join(elems, ", "))
}

// MatchExpression represents: match expr { arms }
type MatchExpression struct {
	Token   token.Token
	Subject Expression
	Arms    []*MatchArm
}

func (me *MatchExpression) expressionNode()      {}
func (me *MatchExpression) TokenLiteral() string { return me.Token.Literal }
func (me *MatchExpression) String() string {
	var sb strings.Builder
	sb.WriteString("match ")
	sb.WriteString(me.Subject.String())
	sb.WriteString(" { ")
	for _, arm := range me.Arms {
		sb.WriteString(arm.String())
		sb.WriteString(", ")
	}
	sb.WriteString("}")
	return sb.String()
}

// MatchArm represents one arm of a match expression.
type MatchArm struct {
	Pattern Expression
	Guard   Expression // optional: if condition
	Body    Expression
}

func (ma *MatchArm) String() string {
	s := ma.Pattern.String()
	if ma.Guard != nil {
		s += fmt.Sprintf(" if %s", ma.Guard.String())
	}
	s += fmt.Sprintf(" => %s", ma.Body.String())
	return s
}

// ForExpression represents: for pattern in iterable { body }
type ForExpression struct {
	Token    token.Token
	Pattern  Expression // Identifier or TupleLiteral for destructuring
	Iterable Expression
	Body     *BlockExpression
}

func (fe *ForExpression) expressionNode()      {}
func (fe *ForExpression) TokenLiteral() string { return fe.Token.Literal }
func (fe *ForExpression) String() string {
	return fmt.Sprintf("for %s in %s %s", fe.Pattern.String(), fe.Iterable.String(), fe.Body.String())
}

// WhileExpression represents: while condition { body }
type WhileExpression struct {
	Token     token.Token
	Condition Expression
	Body      *BlockExpression
}

func (we *WhileExpression) expressionNode()      {}
func (we *WhileExpression) TokenLiteral() string { return we.Token.Literal }
func (we *WhileExpression) String() string {
	return fmt.Sprintf("while %s %s", we.Condition.String(), we.Body.String())
}

// ErrorPropagation represents: expr?
type ErrorPropagation struct {
	Token      token.Token
	Expression Expression
}

func (ep *ErrorPropagation) expressionNode()      {}
func (ep *ErrorPropagation) TokenLiteral() string { return ep.Token.Literal }
func (ep *ErrorPropagation) String() string {
	return fmt.Sprintf("%s?", ep.Expression.String())
}

// ConstructorExpression represents: Type(args) e.g. Some(42), Ok(value), Err("msg")
type ConstructorExpression struct {
	Token     token.Token
	Name      string
	Arguments []Expression
}

func (ce *ConstructorExpression) expressionNode()      {}
func (ce *ConstructorExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *ConstructorExpression) String() string {
	args := make([]string, len(ce.Arguments))
	for i, a := range ce.Arguments {
		args[i] = a.String()
	}
	return fmt.Sprintf("%s(%s)", ce.Name, strings.Join(args, ", "))
}

// Wildcard represents the _ pattern in match expressions.
type Wildcard struct {
	Token token.Token
}

func (w *Wildcard) expressionNode()      {}
func (w *Wildcard) TokenLiteral() string { return w.Token.Literal }
func (w *Wildcard) String() string       { return "_" }

// ---------- Type Annotations ----------

// TypeAnnotation represents a type expression in source code.
type TypeAnnotation struct {
	Token      token.Token
	Name       string
	TypeParams []*TypeAnnotation // generic type parameters
}

func (ta *TypeAnnotation) String() string {
	if len(ta.TypeParams) == 0 {
		return ta.Name
	}
	params := make([]string, len(ta.TypeParams))
	for i, p := range ta.TypeParams {
		params[i] = p.String()
	}
	return fmt.Sprintf("%s<%s>", ta.Name, strings.Join(params, ", "))
}

// ---------- Type Definitions ----------

// TypeStatement represents: type Name = typeExpr
type TypeStatement struct {
	Token      token.Token
	Name       string
	TypeParams []string
	TypeExpr   Expression // StructLiteral, EnumDef, or TypeAnnotation alias
}

func (ts *TypeStatement) statementNode()       {}
func (ts *TypeStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *TypeStatement) String() string {
	return fmt.Sprintf("type %s = %s", ts.Name, ts.TypeExpr.String())
}

// StructDefinition represents: struct { field: type, ... }
type StructDefinition struct {
	Token  token.Token
	Fields []*StructField
}

func (sd *StructDefinition) expressionNode()      {}
func (sd *StructDefinition) TokenLiteral() string { return sd.Token.Literal }
func (sd *StructDefinition) String() string {
	fields := make([]string, len(sd.Fields))
	for i, f := range sd.Fields {
		fields[i] = fmt.Sprintf("%s: %s", f.Name, f.TypeAnn.String())
	}
	return fmt.Sprintf("struct { %s }", strings.Join(fields, ", "))
}

// StructField represents a field in a struct definition.
type StructField struct {
	Name    string
	TypeAnn *TypeAnnotation
}

// EnumDefinition represents: enum { Variant1, Variant2(type), ... }
type EnumDefinition struct {
	Token    token.Token
	Variants []*EnumVariant
}

func (ed *EnumDefinition) expressionNode()      {}
func (ed *EnumDefinition) TokenLiteral() string { return ed.Token.Literal }
func (ed *EnumDefinition) String() string {
	variants := make([]string, len(ed.Variants))
	for i, v := range ed.Variants {
		variants[i] = v.String()
	}
	return fmt.Sprintf("enum { %s }", strings.Join(variants, ", "))
}

// EnumVariant represents a single variant in an enum.
type EnumVariant struct {
	Name   string
	Fields []*TypeAnnotation // optional payload types
}

func (ev *EnumVariant) String() string {
	if len(ev.Fields) == 0 {
		return ev.Name
	}
	fields := make([]string, len(ev.Fields))
	for i, f := range ev.Fields {
		fields[i] = f.String()
	}
	return fmt.Sprintf("%s(%s)", ev.Name, strings.Join(fields, ", "))
}

// StringInterpolation represents a string with embedded expressions.
type StringInterpolation struct {
	Token token.Token
	Parts []Expression // alternating StringLiteral and Expression
}

func (si *StringInterpolation) expressionNode()      {}
func (si *StringInterpolation) TokenLiteral() string { return si.Token.Literal }
func (si *StringInterpolation) String() string {
	var sb strings.Builder
	sb.WriteString("\"")
	for _, p := range si.Parts {
		if sl, ok := p.(*StringLiteral); ok {
			sb.WriteString(sl.Value)
		} else {
			sb.WriteString("{")
			sb.WriteString(p.String())
			sb.WriteString("}")
		}
	}
	sb.WriteString("\"")
	return sb.String()
}
