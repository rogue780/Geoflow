package lexer

import (
	"testing"

	"github.com/rogue780/geoflow/internal/token"
)

func TestNextToken_Operators(t *testing.T) {
	input := `+ - * / % ^ ! == != < > <= >= && || |> <| -> => = := . .. ...`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.ASTERISK, "*"},
		{token.SLASH, "/"},
		{token.PERCENT, "%"},
		{token.CARET, "^"},
		{token.BANG, "!"},
		{token.EQ, "=="},
		{token.NOT_EQ, "!="},
		{token.LT, "<"},
		{token.GT, ">"},
		{token.LT_EQ, "<="},
		{token.GT_EQ, ">="},
		{token.AND, "&&"},
		{token.OR, "||"},
		{token.PIPE, "|>"},
		{token.REVERSE_PIPE, "<|"},
		{token.ARROW, "->"},
		{token.FAT_ARROW, "=>"},
		{token.ASSIGN, "="},
		{token.MUT_ASSIGN, ":="},
		{token.DOT, "."},
		{token.DOTDOT, ".."},
		{token.ELLIPSIS, "..."},
		{token.EOF, ""},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Keywords(t *testing.T) {
	input := `let const fn type struct enum if then else match for in while return true false nil auto import export module mut`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.CONST, "const"},
		{token.FN, "fn"},
		{token.TYPE, "type"},
		{token.STRUCT, "struct"},
		{token.ENUM, "enum"},
		{token.IF, "if"},
		{token.THEN, "then"},
		{token.ELSE, "else"},
		{token.MATCH, "match"},
		{token.FOR, "for"},
		{token.IN, "in"},
		{token.WHILE, "while"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.FALSE, "false"},
		{token.NIL, "nil"},
		{token.AUTO, "auto"},
		{token.IMPORT, "import"},
		{token.EXPORT, "export"},
		{token.MODULE, "module"},
		{token.MUT, "mut"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
	}
}

func TestNextToken_Integers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		tokType  token.Type
	}{
		{"42", "42", token.INT},
		{"0xFF", "0xFF", token.INT},
		{"0b1010", "0b1010", token.INT},
		{"0o777", "0o777", token.INT},
		{"1_000_000", "1_000_000", token.INT},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.tokType {
			t.Errorf("input=%q: expected type %s, got %s", tt.input, tt.tokType, tok.Type)
		}
		if tok.Literal != tt.expected {
			t.Errorf("input=%q: expected literal %q, got %q", tt.input, tt.expected, tok.Literal)
		}
	}
}

func TestNextToken_Floats(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"3.14", "3.14"},
		{"2.5e10", "2.5e10"},
		{"1.0E-5", "1.0E-5"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != token.FLOAT {
			t.Errorf("input=%q: expected FLOAT, got %s", tt.input, tok.Type)
		}
		if tok.Literal != tt.expected {
			t.Errorf("input=%q: expected literal %q, got %q", tt.input, tt.expected, tok.Literal)
		}
	}
}

func TestNextToken_Strings(t *testing.T) {
	input := `"hello world" "line1\nline2"`

	l := New(input)
	tok1 := l.NextToken()
	if tok1.Type != token.STRING || tok1.Literal != "hello world" {
		t.Errorf("expected STRING 'hello world', got %s %q", tok1.Type, tok1.Literal)
	}

	tok2 := l.NextToken()
	if tok2.Type != token.STRING || tok2.Literal != "line1\nline2" {
		t.Errorf("expected STRING with newline, got %s %q", tok2.Type, tok2.Literal)
	}
}

func TestNextToken_RawString(t *testing.T) {
	input := "`raw \\n string`"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.RAW_STRING {
		t.Errorf("expected RAW_STRING, got %s", tok.Type)
	}
	if tok.Literal != `raw \n string` {
		t.Errorf("expected raw content, got %q", tok.Literal)
	}
}

func TestNextToken_WKTLiteral(t *testing.T) {
	input := `#POINT(0 0)#`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.WKT {
		t.Errorf("expected WKT, got %s", tok.Type)
	}
	if tok.Literal != "POINT(0 0)" {
		t.Errorf("expected 'POINT(0 0)', got %q", tok.Literal)
	}
}

func TestNextToken_SymbolicLiteral(t *testing.T) {
	input := `$x^2 + 3*x + 2$`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.SYMBOLIC {
		t.Errorf("expected SYMBOLIC, got %s", tok.Type)
	}
	if tok.Literal != "x^2 + 3*x + 2" {
		t.Errorf("expected symbolic expr, got %q", tok.Literal)
	}
}

func TestNextToken_Comments(t *testing.T) {
	input := `42 -- this is a comment
43`
	l := New(input)
	tok1 := l.NextToken()
	if tok1.Type != token.INT || tok1.Literal != "42" {
		t.Errorf("expected INT 42, got %s %q", tok1.Type, tok1.Literal)
	}
	tok2 := l.NextToken()
	if tok2.Type != token.INT || tok2.Literal != "43" {
		t.Errorf("expected INT 43, got %s %q", tok2.Type, tok2.Literal)
	}
}

func TestNextToken_MultiLineComment(t *testing.T) {
	input := `42 {- this is
a multi-line comment -} 43`
	l := New(input)
	tok1 := l.NextToken()
	if tok1.Type != token.INT || tok1.Literal != "42" {
		t.Errorf("expected INT 42, got %s %q", tok1.Type, tok1.Literal)
	}
	tok2 := l.NextToken()
	if tok2.Type != token.INT || tok2.Literal != "43" {
		t.Errorf("expected INT 43, got %s %q", tok2.Type, tok2.Literal)
	}
}

func TestNextToken_CompleteProgram(t *testing.T) {
	input := `let x = 42
let y = 3.14
let name = "GeoFlow"
fn add(a: int, b: int) -> int {
    a + b
}
let result = [1, 2, 3] |> map(\x -> x * 2)`

	l := New(input)
	var tokens []token.Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}

	// Verify we got a reasonable number of tokens
	if len(tokens) < 30 {
		t.Errorf("expected at least 30 tokens, got %d", len(tokens))
	}

	// Check first few tokens
	expected := []token.Type{token.LET, token.IDENT, token.ASSIGN, token.INT}
	for i, exp := range expected {
		if tokens[i].Type != exp {
			t.Errorf("token[%d]: expected %s, got %s", i, exp, tokens[i].Type)
		}
	}
}

func TestNextToken_Delimiters(t *testing.T) {
	input := `( ) [ ] { } , : ; \ ?`
	expected := []token.Type{
		token.LPAREN, token.RPAREN,
		token.LBRACKET, token.RBRACKET,
		token.LBRACE, token.RBRACE,
		token.COMMA, token.COLON, token.SEMICOLON,
		token.BACKSLASH, token.QUESTION,
		token.EOF,
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Errorf("token[%d]: expected %s, got %s (literal=%q)", i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestTokenize(t *testing.T) {
	tokens := Tokenize("1 + 2")
	if len(tokens) != 4 { // INT PLUS INT EOF
		t.Errorf("expected 4 tokens, got %d", len(tokens))
	}
}

func TestNextToken_Position(t *testing.T) {
	input := "let x = 42"
	l := New(input)
	tok := l.NextToken()
	if tok.Pos.Line != 1 {
		t.Errorf("expected line 1, got %d", tok.Pos.Line)
	}
}
