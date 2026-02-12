// Package parser implements a recursive descent parser for GeoFlow.
package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rogue780/geoflow/internal/ast"
	"github.com/rogue780/geoflow/internal/lexer"
	"github.com/rogue780/geoflow/internal/token"
)

// Precedence levels for operator parsing.
const (
	_ int = iota
	LOWEST
	PIPELINE    // |>
	COMPOSITION // f g (juxtaposition)
	ASSIGN_PREC // :=
	OR_PREC     // ||
	AND_PREC    // &&
	EQUALS      // == !=
	LESSGREATER // < > <= >=
	SUM         // + -
	PRODUCT     // * / %
	POWER       // ^
	PREFIX      // -x !x
	CALL        // f(x) x.y x[i]
)

var precedences = map[token.Type]int{
	token.PIPE:         PIPELINE,
	token.REVERSE_PIPE: PIPELINE,
	token.OR:           OR_PREC,
	token.AND:          AND_PREC,
	token.EQ:           EQUALS,
	token.NOT_EQ:       EQUALS,
	token.LT:           LESSGREATER,
	token.GT:           LESSGREATER,
	token.LT_EQ:        LESSGREATER,
	token.GT_EQ:        LESSGREATER,
	token.PLUS:         SUM,
	token.MINUS:        SUM,
	token.ASTERISK:     PRODUCT,
	token.SLASH:        PRODUCT,
	token.PERCENT:      PRODUCT,
	token.CARET:        POWER,
	token.LPAREN:       CALL,
	token.DOT:          CALL,
	token.LBRACKET:     CALL,
	token.QUESTION:     CALL,
}

// Parser parses a stream of tokens into an AST.
type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token
}

// New creates a new parser for the given lexer.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	// Read two tokens to populate curToken and peekToken.
	p.nextToken()
	p.nextToken()
	return p
}

// Errors returns any parse errors encountered.
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, fmt.Sprintf("line %d, col %d: %s", p.curToken.Pos.Line, p.curToken.Pos.Column, msg))
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.Type) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.Type) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.Type) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t token.Type) {
	p.addError(fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type))
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// ParseProgram parses the entire program.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.CONST:
		return p.parseConstStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.BREAK:
		return &ast.BreakStatement{Token: p.curToken}
	case token.CONTINUE:
		return &ast.ContinueStatement{Token: p.curToken}
	case token.IMPORT:
		return p.parseImportStatement()
	case token.FN:
		if p.peekTokenIs(token.IDENT) {
			return p.parseFunctionStatement()
		}
		return p.parseExpressionStatement()
	case token.TYPE:
		return p.parseTypeStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	// Check for 'mut' keyword
	if p.peekTokenIs(token.MUT) {
		p.nextToken()
		stmt.Mutable = true
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Optional type annotation
	if p.peekTokenIs(token.COLON) {
		p.nextToken()
		p.nextToken()
		stmt.TypeAnn = p.parseTypeAnnotation()
	}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseConstStatement() *ast.ConstStatement {
	stmt := &ast.ConstStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()

	if !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt.Value = p.parseExpression(LOWEST)
	}
	return stmt
}

func (p *Parser) parseImportStatement() *ast.ImportStatement {
	stmt := &ast.ImportStatement{Token: p.curToken}
	p.nextToken()

	// Parse dotted path: std.math.stats
	stmt.Path = append(stmt.Path, p.curToken.Literal)
	for p.peekTokenIs(token.DOT) {
		p.nextToken() // skip dot
		p.nextToken() // move to next ident
		stmt.Path = append(stmt.Path, p.curToken.Literal)
	}

	// Optional alias: as name
	if p.peekTokenIs(token.AS) {
		p.nextToken()
		p.nextToken()
		stmt.Alias = p.curToken.Literal
	}

	return stmt
}

func (p *Parser) parseFunctionStatement() ast.Statement {
	tok := p.curToken
	p.nextToken() // skip 'fn', now at name

	name := p.curToken.Literal

	// Parse type parameters
	var typeParams []string
	if p.peekTokenIs(token.LT) {
		p.nextToken()
		typeParams = p.parseTypeParamList()
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	params := p.parseFunctionParameters()

	var returnType *ast.TypeAnnotation
	if p.peekTokenIs(token.ARROW) {
		p.nextToken()
		p.nextToken()
		returnType = p.parseTypeAnnotation()
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	body := p.parseBlockExpression()

	fn := &ast.FunctionLiteral{
		Token:      tok,
		Name:       name,
		TypeParams: typeParams,
		Parameters: params,
		ReturnType: returnType,
		Body:       body,
	}

	return &ast.LetStatement{
		Token: tok,
		Name:  &ast.Identifier{Token: token.Token{Type: token.IDENT, Literal: name}, Value: name},
		Value: fn,
	}
}

func (p *Parser) parseTypeStatement() *ast.TypeStatement {
	stmt := &ast.TypeStatement{Token: p.curToken}
	p.nextToken()
	stmt.Name = p.curToken.Literal

	// Optional type parameters
	if p.peekTokenIs(token.LT) {
		p.nextToken()
		stmt.TypeParams = p.parseTypeParamNames()
	}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()

	if p.curTokenIs(token.STRUCT) {
		stmt.TypeExpr = p.parseStructDefinition()
	} else if p.curTokenIs(token.ENUM) {
		stmt.TypeExpr = p.parseEnumDefinition()
	} else {
		// Type alias - parse as type annotation wrapped in identifier
		ta := p.parseTypeAnnotation()
		stmt.TypeExpr = &ast.Identifier{
			Token: token.Token{Type: token.IDENT, Literal: ta.Name},
			Value: ta.String(),
		}
	}

	return stmt
}

func (p *Parser) parseStructDefinition() *ast.StructDefinition {
	sd := &ast.StructDefinition{Token: p.curToken}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		name := p.curToken.Literal
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken()
		ta := p.parseTypeAnnotation()
		sd.Fields = append(sd.Fields, &ast.StructField{Name: name, TypeAnn: ta})

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}
	return sd
}

func (p *Parser) parseEnumDefinition() *ast.EnumDefinition {
	ed := &ast.EnumDefinition{Token: p.curToken}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		variant := &ast.EnumVariant{Name: p.curToken.Literal}
		if p.peekTokenIs(token.LPAREN) {
			p.nextToken()
			p.nextToken()
			for !p.curTokenIs(token.RPAREN) {
				ta := p.parseTypeAnnotation()
				variant.Fields = append(variant.Fields, ta)
				if p.peekTokenIs(token.COMMA) {
					p.nextToken()
				}
				p.nextToken()
			}
		}
		ed.Variants = append(ed.Variants, variant)
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}
	return ed
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	// Handle mutable assignment: name := value
	if p.peekTokenIs(token.MUT_ASSIGN) {
		tok := p.peekToken
		p.nextToken() // consume :=
		p.nextToken() // move to value
		value := p.parseExpression(LOWEST)
		return &ast.ExpressionStatement{
			Token: tok,
			Expression: &ast.InfixExpression{
				Token:    tok,
				Left:     stmt.Expression,
				Operator: ":=",
				Right:    value,
			},
		}
	}

	return stmt
}

// ---------- Expression Parsing (Pratt Parser) ----------

func (p *Parser) parseExpression(precedence int) ast.Expression {
	left := p.parsePrefixExpression()
	if left == nil {
		return nil
	}

	for {
		// Standard Pratt infix loop
		for !p.peekTokenIs(token.EOF) && precedence < p.peekPrecedence() {
			left = p.parseInfixExpressionWith(left)
			if left == nil {
				return nil
			}
		}

		// Juxtaposition composition: f g h (same line, left-to-right)
		// After juxtaposition, loop back to check for infix operators (e.g. |>)
		if precedence < COMPOSITION &&
			p.canStartJuxtaposition() &&
			p.curToken.Pos.Line == p.peekToken.Pos.Line {
			p.nextToken()
			right := p.parseExpression(COMPOSITION)
			if right == nil {
				return nil
			}
			left = &ast.JuxtapositionExpression{Token: p.curToken, Left: left, Right: right}
			continue
		}

		break
	}

	return left
}

func (p *Parser) canStartJuxtaposition() bool {
	switch p.peekToken.Type {
	case token.IDENT, token.BACKSLASH, token.FN:
		return true
	}
	return false
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	switch p.curToken.Type {
	case token.IDENT:
		return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	case token.INT:
		return p.parseIntegerLiteral()
	case token.FLOAT:
		return p.parseFloatLiteral()
	case token.STRING:
		return p.parseStringLiteral()
	case token.RAW_STRING:
		return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	case token.TRUE:
		return &ast.BooleanLiteral{Token: p.curToken, Value: true}
	case token.FALSE:
		return &ast.BooleanLiteral{Token: p.curToken, Value: false}
	case token.NIL:
		return &ast.NilLiteral{Token: p.curToken}
	case token.WKT:
		return &ast.WKTLiteral{Token: p.curToken, Value: p.curToken.Literal}
	case token.SYMBOLIC:
		return &ast.SymbolicLiteral{Token: p.curToken, Value: p.curToken.Literal}
	case token.BANG:
		return p.parsePrefixOp()
	case token.MINUS:
		return p.parsePrefixOp()
	case token.LPAREN:
		return p.parseGroupedOrTuple()
	case token.LBRACKET:
		return p.parseListLiteral()
	case token.LBRACE:
		return p.parseMapOrBlock()
	case token.FN:
		return p.parseFunctionLiteral()
	case token.BACKSLASH:
		return p.parseLambdaExpression()
	case token.IF:
		return p.parseIfExpression()
	case token.MATCH:
		return p.parseMatchExpression()
	case token.FOR:
		return p.parseForExpression()
	case token.WHILE:
		return p.parseWhileExpression()
	case token.TRY:
		return p.parseTryCatchExpression()
	case token.BREAK:
		return &ast.Identifier{Token: p.curToken, Value: "break"}
	case token.CONTINUE:
		return &ast.Identifier{Token: p.curToken, Value: "continue"}
	default:
		p.addError(fmt.Sprintf("no prefix parse function for %s", p.curToken.Type))
		return nil
	}
}

func (p *Parser) parseInfixExpressionWith(left ast.Expression) ast.Expression {
	switch p.peekToken.Type {
	case token.PLUS, token.MINUS, token.ASTERISK, token.SLASH, token.PERCENT, token.CARET,
		token.EQ, token.NOT_EQ, token.LT, token.GT, token.LT_EQ, token.GT_EQ,
		token.AND, token.OR:
		p.nextToken()
		return p.parseInfixExpression(left)
	case token.PIPE:
		p.nextToken()
		return p.parsePipelineExpression(left)
	case token.REVERSE_PIPE:
		p.nextToken()
		return p.parseReversePipelineExpression(left)
	case token.QUESTION:
		p.nextToken()
		return &ast.ErrorPropagation{Token: p.curToken, Expression: left}
	case token.LPAREN:
		p.nextToken()
		return p.parseCallExpression(left)
	case token.DOT:
		p.nextToken()
		return p.parseDotExpression(left)
	case token.LBRACKET:
		p.nextToken()
		return p.parseIndexExpression(left)
	default:
		return left
	}
}

func (p *Parser) parsePrefixOp() ast.Expression {
	expr := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expr.Right = p.parseExpression(PREFIX)
	return expr
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expr := &ast.InfixExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}

	precedence := p.curPrecedence()
	// Right-associativity for ^ (power)
	if p.curToken.Type == token.CARET {
		precedence--
	}
	p.nextToken()
	expr.Right = p.parseExpression(precedence)
	return expr
}

func (p *Parser) parsePipelineExpression(left ast.Expression) ast.Expression {
	expr := &ast.PipelineExpression{
		Token: p.curToken,
		Left:  left,
	}
	p.nextToken()
	expr.Right = p.parseExpression(PIPELINE)
	return expr
}

func (p *Parser) parseReversePipelineExpression(left ast.Expression) ast.Expression {
	expr := &ast.PipelineExpression{
		Token:   p.curToken,
		Left:    left,
		Reverse: true,
	}
	p.nextToken()
	expr.Right = p.parseExpression(PIPELINE)
	return expr
}

func (p *Parser) parseCallExpression(fn ast.Expression) ast.Expression {
	expr := &ast.CallExpression{Token: p.curToken, Function: fn}
	expr.Arguments = p.parseExpressionList(token.RPAREN)
	return expr
}

func (p *Parser) parseDotExpression(left ast.Expression) ast.Expression {
	dotToken := p.curToken
	p.nextToken() // consume dot

	// If the token after dot is not an identifier, treat as composition operator
	// e.g. (\x -> x * 3) . (\x -> x + 1)
	if !p.curTokenIs(token.IDENT) {
		right := p.parseExpression(CALL)
		if right == nil {
			return nil
		}
		return &ast.InfixExpression{
			Token:    dotToken,
			Left:     left,
			Operator: ".",
			Right:    right,
		}
	}

	expr := &ast.DotExpression{
		Token: p.curToken,
		Left:  left,
		Field: p.curToken.Literal,
	}
	return expr
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	expr := &ast.IndexExpression{Token: p.curToken, Left: left}
	p.nextToken()
	expr.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return expr
}

func (p *Parser) parseGroupedOrTuple() ast.Expression {
	tok := p.curToken
	p.nextToken()

	if p.curTokenIs(token.RPAREN) {
		// Empty tuple ()
		return &ast.TupleLiteral{Token: tok, Elements: []ast.Expression{}}
	}

	first := p.parseExpression(LOWEST)

	if p.peekTokenIs(token.COMMA) {
		// It's a tuple
		elements := []ast.Expression{first}
		for p.peekTokenIs(token.COMMA) {
			p.nextToken() // skip comma
			p.nextToken()
			elements = append(elements, p.parseExpression(LOWEST))
		}
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
		return &ast.TupleLiteral{Token: tok, Elements: elements}
	}

	// Grouped expression
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return first
}

func (p *Parser) parseListLiteral() ast.Expression {
	tok := p.curToken
	elements := p.parseExpressionList(token.RBRACKET)
	return &ast.ListLiteral{Token: tok, Elements: elements}
}

func (p *Parser) parseMapOrBlock() ast.Expression {
	tok := p.curToken

	// Empty braces -> empty map
	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return &ast.MapLiteral{
			Token: tok,
			Pairs: make(map[ast.Expression]ast.Expression),
			Order: []ast.Expression{},
		}
	}

	// Try to determine if this is a map or a block
	// Maps start with string: or ident: (followed by expression)
	if (p.peekTokenIs(token.STRING) || p.peekTokenIs(token.IDENT)) && p.isMapLiteral() {
		return p.parseMapLiteral(tok)
	}

	// It's a block expression
	return p.parseBlockExpression()
}

func (p *Parser) isMapLiteral() bool {
	// Save position and look ahead to check for "key: value" pattern
	// This is a simplified heuristic
	if p.peekTokenIs(token.STRING) {
		return true // String keys always indicate map
	}
	// For identifiers, check if after ident there's a colon
	// We need to be more careful here - could be a block with let statements
	return false
}

func (p *Parser) parseMapLiteral(tok token.Token) ast.Expression {
	ml := &ast.MapLiteral{
		Token: tok,
		Pairs: make(map[ast.Expression]ast.Expression),
		Order: []ast.Expression{},
	}

	p.nextToken()
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken()
		value := p.parseExpression(LOWEST)
		ml.Pairs[key] = value
		ml.Order = append(ml.Order, key)

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}
	return ml
}

func (p *Parser) parseBlockExpression() *ast.BlockExpression {
	block := &ast.BlockExpression{Token: p.curToken}
	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		// Try to parse as statement
		stmt := p.parseStatement()
		if stmt != nil {
			// Check if next token indicates this is the last expression (the return value)
			if p.peekTokenIs(token.RBRACE) {
				// If the statement is an ExpressionStatement, it's the block's return value
				if es, ok := stmt.(*ast.ExpressionStatement); ok {
					block.Value = es.Expression
				} else {
					block.Statements = append(block.Statements, stmt)
				}
			} else {
				block.Statements = append(block.Statements, stmt)
			}
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	tok := p.curToken

	// Check for named function: fn name(...)
	name := ""
	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		name = p.curToken.Literal
	}

	// Parse type parameters
	var typeParams []string
	if p.peekTokenIs(token.LT) {
		p.nextToken()
		typeParams = p.parseTypeParamList()
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	params := p.parseFunctionParameters()

	var returnType *ast.TypeAnnotation
	if p.peekTokenIs(token.ARROW) {
		p.nextToken()
		p.nextToken()
		returnType = p.parseTypeAnnotation()
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	body := p.parseBlockExpression()

	return &ast.FunctionLiteral{
		Token:      tok,
		Name:       name,
		TypeParams: typeParams,
		Parameters: params,
		ReturnType: returnType,
		Body:       body,
	}
}

func (p *Parser) parseFunctionParameters() []*ast.FunctionParameter {
	var params []*ast.FunctionParameter

	p.nextToken()
	if p.curTokenIs(token.RPAREN) {
		return params
	}

	for {
		param := &ast.FunctionParameter{}

		// Check for variadic
		if p.curTokenIs(token.ELLIPSIS) {
			param.Variadic = true
			p.nextToken()
		}

		param.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		// Optional type annotation
		if p.peekTokenIs(token.COLON) {
			p.nextToken()
			p.nextToken()
			param.TypeAnn = p.parseTypeAnnotation()
		}

		// Optional default value
		if p.peekTokenIs(token.ASSIGN) {
			p.nextToken()
			p.nextToken()
			param.Default = p.parseExpression(LOWEST)
		}

		params = append(params, param)

		if !p.peekTokenIs(token.COMMA) {
			break
		}
		p.nextToken() // skip comma
		p.nextToken() // move to next param
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return params
}

func (p *Parser) parseLambdaExpression() ast.Expression {
	tok := p.curToken
	p.nextToken()

	// Parse parameter names
	var params []*ast.Identifier
	params = append(params, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		params = append(params, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
	}

	if !p.expectPeek(token.ARROW) {
		return nil
	}
	p.nextToken()

	body := p.parseExpression(LOWEST)

	return &ast.LambdaExpression{
		Token:      tok,
		Parameters: params,
		Body:       body,
	}
}

func (p *Parser) parseIfExpression() ast.Expression {
	expr := &ast.IfExpression{Token: p.curToken}
	p.nextToken()

	expr.Condition = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.THEN) {
		p.nextToken()
	}
	p.nextToken()

	// Parse consequence - could be a block or a single expression
	if p.curTokenIs(token.LBRACE) {
		expr.Consequence = p.parseBlockExpression()
	} else {
		expr.Consequence = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		p.nextToken()

		if p.curTokenIs(token.IF) {
			expr.Alternative = p.parseIfExpression()
		} else if p.curTokenIs(token.LBRACE) {
			expr.Alternative = p.parseBlockExpression()
		} else {
			expr.Alternative = p.parseExpression(LOWEST)
		}
	}

	return expr
}

func (p *Parser) parseMatchExpression() ast.Expression {
	expr := &ast.MatchExpression{Token: p.curToken}
	p.nextToken()
	expr.Subject = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		arm := p.parseMatchArm()
		if arm != nil {
			expr.Arms = append(expr.Arms, arm)
		}
		// Skip comma if present
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}

	return expr
}

func (p *Parser) parseMatchArm() *ast.MatchArm {
	arm := &ast.MatchArm{}

	// Parse pattern
	if p.curToken.Literal == "_" {
		arm.Pattern = &ast.Wildcard{Token: p.curToken}
	} else {
		arm.Pattern = p.parseExpression(LOWEST)
	}

	// Optional guard
	if p.peekTokenIs(token.IF) {
		p.nextToken()
		p.nextToken()
		arm.Guard = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(token.FAT_ARROW) {
		return nil
	}
	p.nextToken()

	arm.Body = p.parseExpression(LOWEST)

	return arm
}

func (p *Parser) parseForExpression() ast.Expression {
	expr := &ast.ForExpression{Token: p.curToken}
	p.nextToken()

	// Parse pattern (simple identifier or tuple destructuring)
	if p.curTokenIs(token.LPAREN) {
		expr.Pattern = p.parseGroupedOrTuple()
	} else {
		expr.Pattern = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	if !p.expectPeek(token.IN) {
		return nil
	}
	p.nextToken()
	expr.Iterable = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expr.Body = p.parseBlockExpression()

	return expr
}

func (p *Parser) parseWhileExpression() ast.Expression {
	expr := &ast.WhileExpression{Token: p.curToken}
	p.nextToken()
	expr.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expr.Body = p.parseBlockExpression()

	return expr
}

func (p *Parser) parseTryCatchExpression() ast.Expression {
	expr := &ast.TryCatchExpression{Token: p.curToken}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expr.TryBody = p.parseBlockExpression()

	if !p.expectPeek(token.CATCH) {
		return nil
	}

	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		expr.CatchParam = p.curToken.Literal
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expr.CatchBody = p.parseBlockExpression()

	return expr
}

func (p *Parser) parseExpressionList(end token.Type) []ast.Expression {
	var list []ast.Expression

	p.nextToken()
	if p.curTokenIs(end) {
		return list
	}

	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}
	return list
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	s := p.curToken.Literal
	s = strings.ReplaceAll(s, "_", "")

	var value int64
	var err error
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		value, err = strconv.ParseInt(s[2:], 16, 64)
	} else if strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B") {
		value, err = strconv.ParseInt(s[2:], 2, 64)
	} else if strings.HasPrefix(s, "0o") || strings.HasPrefix(s, "0O") {
		value, err = strconv.ParseInt(s[2:], 8, 64)
	} else {
		value, err = strconv.ParseInt(s, 10, 64)
	}

	if err != nil {
		p.addError(fmt.Sprintf("could not parse %q as integer", p.curToken.Literal))
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}
	s := strings.ReplaceAll(p.curToken.Literal, "_", "")
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		p.addError(fmt.Sprintf("could not parse %q as float", p.curToken.Literal))
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	value := p.curToken.Literal

	// Check for string interpolation
	if strings.Contains(value, "{") && strings.Contains(value, "}") {
		return p.parseStringInterpolation(value)
	}

	return &ast.StringLiteral{Token: p.curToken, Value: value}
}

func (p *Parser) parseStringInterpolation(value string) ast.Expression {
	si := &ast.StringInterpolation{Token: p.curToken}
	i := 0
	for i < len(value) {
		if value[i] == '{' {
			// Find matching }
			j := i + 1
			depth := 1
			for j < len(value) && depth > 0 {
				if value[j] == '{' {
					depth++
				} else if value[j] == '}' {
					depth--
				}
				if depth > 0 {
					j++
				}
			}
			exprStr := value[i+1 : j]
			// Parse the expression
			l := lexer.New(exprStr)
			parser := New(l)
			expr := parser.parseExpression(LOWEST)
			if expr != nil {
				si.Parts = append(si.Parts, expr)
			}
			i = j + 1
		} else {
			// Find next { or end
			j := i
			for j < len(value) && value[j] != '{' {
				j++
			}
			si.Parts = append(si.Parts, &ast.StringLiteral{
				Token: p.curToken,
				Value: value[i:j],
			})
			i = j
		}
	}
	return si
}

func (p *Parser) parseTypeAnnotation() *ast.TypeAnnotation {
	ta := &ast.TypeAnnotation{
		Token: p.curToken,
		Name:  p.curToken.Literal,
	}

	// Check for generic params: Type<A, B>
	if p.peekTokenIs(token.LT) {
		p.nextToken()
		p.nextToken()
		for !p.curTokenIs(token.GT) && !p.curTokenIs(token.EOF) {
			param := p.parseTypeAnnotation()
			ta.TypeParams = append(ta.TypeParams, param)
			if p.peekTokenIs(token.COMMA) {
				p.nextToken()
			}
			p.nextToken()
		}
	}

	return ta
}

func (p *Parser) parseTypeParamList() []string {
	var params []string
	p.nextToken()
	for !p.curTokenIs(token.GT) && !p.curTokenIs(token.EOF) {
		params = append(params, p.curToken.Literal)
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}
	return params
}

func (p *Parser) parseTypeParamNames() []string {
	var params []string
	p.nextToken()
	for !p.curTokenIs(token.GT) && !p.curTokenIs(token.EOF) {
		params = append(params, p.curToken.Literal)
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}
	return params
}
