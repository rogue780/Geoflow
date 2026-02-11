// Package lexer implements the lexer for the GeoFlow language.
package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/rogue780/geoflow/internal/token"
)

// Lexer performs lexical analysis on GeoFlow source code.
type Lexer struct {
	input   string
	pos     int  // current position in input (points to current char)
	readPos int  // current reading position (after current char)
	ch      rune // current character
	line    int
	col     int
}

// New creates a new Lexer for the given input.
func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, col: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	l.pos = l.readPos
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		r, size := utf8.DecodeRuneInString(l.input[l.readPos:])
		l.ch = r
		l.readPos += size
	}
	l.col++
}

func (l *Lexer) peekChar() rune {
	if l.readPos >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPos:])
	return r
}

func (l *Lexer) currentPos() token.Position {
	return token.Position{Line: l.line, Column: l.col, Offset: l.pos}
}

func (l *Lexer) newToken(tokenType token.Type, literal string) token.Token {
	return token.Token{Type: tokenType, Literal: literal, Pos: l.currentPos()}
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() token.Token {
	l.skipWhitespaceAndComments()

	pos := l.currentPos()

	var tok token.Token

	switch l.ch {
	case '+':
		tok = l.newToken(token.PLUS, "+")
	case '*':
		tok = l.newToken(token.ASTERISK, "*")
	case '/':
		tok = l.newToken(token.SLASH, "/")
	case '%':
		tok = l.newToken(token.PERCENT, "%")
	case '^':
		tok = l.newToken(token.CARET, "^")
	case '(':
		tok = l.newToken(token.LPAREN, "(")
	case ')':
		tok = l.newToken(token.RPAREN, ")")
	case '[':
		tok = l.newToken(token.LBRACKET, "[")
	case ']':
		tok = l.newToken(token.RBRACKET, "]")
	case '{':
		// Check for multi-line comment {- ... -}
		if l.peekChar() == '-' {
			l.skipMultiLineComment()
			return l.NextToken()
		}
		tok = l.newToken(token.LBRACE, "{")
	case '}':
		tok = l.newToken(token.RBRACE, "}")
	case ',':
		tok = l.newToken(token.COMMA, ",")
	case ';':
		tok = l.newToken(token.SEMICOLON, ";")
	case '\\':
		tok = l.newToken(token.BACKSLASH, "\\")
	case '?':
		tok = l.newToken(token.QUESTION, "?")
	case ':':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.MUT_ASSIGN, Literal: ":=", Pos: pos}
		} else {
			tok = l.newToken(token.COLON, ":")
		}
	case '.':
		if l.peekChar() == '.' {
			l.readChar()
			if l.peekChar() == '.' {
				l.readChar()
				tok = token.Token{Type: token.ELLIPSIS, Literal: "...", Pos: pos}
			} else {
				tok = token.Token{Type: token.DOTDOT, Literal: "..", Pos: pos}
			}
		} else {
			tok = l.newToken(token.DOT, ".")
		}
	case '-':
		if l.peekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.ARROW, Literal: "->", Pos: pos}
		} else if l.peekChar() == '-' {
			l.skipSingleLineComment()
			return l.NextToken()
		} else {
			tok = l.newToken(token.MINUS, "-")
		}
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: "==", Pos: pos}
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.FAT_ARROW, Literal: "=>", Pos: pos}
		} else {
			tok = l.newToken(token.ASSIGN, "=")
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: "!=", Pos: pos}
		} else {
			tok = l.newToken(token.BANG, "!")
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.LT_EQ, Literal: "<=", Pos: pos}
		} else if l.peekChar() == '|' {
			l.readChar()
			tok = token.Token{Type: token.REVERSE_PIPE, Literal: "<|", Pos: pos}
		} else {
			tok = l.newToken(token.LT, "<")
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.GT_EQ, Literal: ">=", Pos: pos}
		} else {
			tok = l.newToken(token.GT, ">")
		}
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok = token.Token{Type: token.AND, Literal: "&&", Pos: pos}
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.ch), Pos: pos}
		}
	case '|':
		if l.peekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.PIPE, Literal: "|>", Pos: pos}
		} else if l.peekChar() == '|' {
			l.readChar()
			tok = token.Token{Type: token.OR, Literal: "||", Pos: pos}
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.ch), Pos: pos}
		}
	case '"':
		str, err := l.readString()
		if err != "" {
			tok = token.Token{Type: token.ILLEGAL, Literal: err, Pos: pos}
		} else {
			tok = token.Token{Type: token.STRING, Literal: str, Pos: pos}
		}
		return tok
	case '`':
		str := l.readRawString()
		tok = token.Token{Type: token.RAW_STRING, Literal: str, Pos: pos}
		return tok
	case '#':
		wkt := l.readWKTLiteral()
		tok = token.Token{Type: token.WKT, Literal: wkt, Pos: pos}
		return tok
	case '$':
		sym := l.readSymbolicExpression()
		tok = token.Token{Type: token.SYMBOLIC, Literal: sym, Pos: pos}
		return tok
	case 0:
		tok = token.Token{Type: token.EOF, Literal: "", Pos: pos}
		return tok
	default:
		if isLetter(l.ch) {
			ident := l.readIdentifier()
			tokType := token.LookupIdent(ident)
			return token.Token{Type: tokType, Literal: ident, Pos: pos}
		} else if isDigit(l.ch) {
			return l.readNumber(pos)
		}
		tok = token.Token{Type: token.ILLEGAL, Literal: string(l.ch), Pos: pos}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		if l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
			l.readChar()
		} else if l.ch == '\n' {
			l.line++
			l.col = 0
			l.readChar()
		} else {
			break
		}
	}
}

func (l *Lexer) skipSingleLineComment() {
	// Skip the second '-'
	l.readChar()
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

func (l *Lexer) skipMultiLineComment() {
	// We are at '{', peek is '-'
	l.readChar() // consume '-'
	l.readChar() // move past '-'
	depth := 1
	for depth > 0 && l.ch != 0 {
		if l.ch == '{' && l.peekChar() == '-' {
			depth++
			l.readChar()
		} else if l.ch == '-' && l.peekChar() == '}' {
			depth--
			l.readChar()
		} else if l.ch == '\n' {
			l.line++
			l.col = 0
		}
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	start := l.pos
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.pos]
}

func (l *Lexer) readNumber(pos token.Position) token.Token {
	start := l.pos
	isFloat := false

	// Check for hex, binary, octal prefixes
	if l.ch == '0' && l.readPos < len(l.input) {
		next := l.peekChar()
		if next == 'x' || next == 'X' {
			l.readChar() // skip '0'
			l.readChar() // skip 'x'
			for isHexDigit(l.ch) || l.ch == '_' {
				l.readChar()
			}
			return token.Token{Type: token.INT, Literal: l.input[start:l.pos], Pos: pos}
		} else if next == 'b' || next == 'B' {
			l.readChar() // skip '0'
			l.readChar() // skip 'b'
			for l.ch == '0' || l.ch == '1' || l.ch == '_' {
				l.readChar()
			}
			return token.Token{Type: token.INT, Literal: l.input[start:l.pos], Pos: pos}
		} else if next == 'o' || next == 'O' {
			l.readChar() // skip '0'
			l.readChar() // skip 'o'
			for (l.ch >= '0' && l.ch <= '7') || l.ch == '_' {
				l.readChar()
			}
			return token.Token{Type: token.INT, Literal: l.input[start:l.pos], Pos: pos}
		}
	}

	for isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}

	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar() // consume '.'
		for isDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}
	}

	if l.ch == 'e' || l.ch == 'E' {
		isFloat = true
		l.readChar() // consume 'e'
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	literal := l.input[start:l.pos]
	if isFloat {
		return token.Token{Type: token.FLOAT, Literal: literal, Pos: pos}
	}
	return token.Token{Type: token.INT, Literal: literal, Pos: pos}
}

func (l *Lexer) readString() (string, string) {
	l.readChar() // skip opening quote
	var sb strings.Builder
	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			case '\\':
				sb.WriteByte('\\')
			case '"':
				sb.WriteByte('"')
			case '0':
				sb.WriteByte(0)
			default:
				return "", fmt.Sprintf("unknown escape sequence: \\%c", l.ch)
			}
		} else {
			if l.ch == '\n' {
				l.line++
				l.col = 0
			}
			sb.WriteRune(l.ch)
		}
		l.readChar()
	}
	if l.ch == 0 {
		return "", "unterminated string"
	}
	l.readChar() // skip closing quote
	return sb.String(), ""
}

func (l *Lexer) readRawString() string {
	l.readChar() // skip opening backtick
	start := l.pos
	for l.ch != '`' && l.ch != 0 {
		if l.ch == '\n' {
			l.line++
			l.col = 0
		}
		l.readChar()
	}
	str := l.input[start:l.pos]
	if l.ch == '`' {
		l.readChar() // skip closing backtick
	}
	return str
}

func (l *Lexer) readWKTLiteral() string {
	l.readChar() // skip opening '#'
	start := l.pos
	for l.ch != '#' && l.ch != 0 {
		if l.ch == '\n' {
			l.line++
			l.col = 0
		}
		l.readChar()
	}
	wkt := l.input[start:l.pos]
	if l.ch == '#' {
		l.readChar() // skip closing '#'
	}
	return wkt
}

func (l *Lexer) readSymbolicExpression() string {
	l.readChar() // skip opening '$'
	start := l.pos
	for l.ch != '$' && l.ch != 0 {
		if l.ch == '\n' {
			l.line++
			l.col = 0
		}
		l.readChar()
	}
	sym := l.input[start:l.pos]
	if l.ch == '$' {
		l.readChar() // skip closing '$'
	}
	return sym
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch rune) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// Tokenize returns all tokens from the input.
func Tokenize(input string) []token.Token {
	l := New(input)
	var tokens []token.Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}
