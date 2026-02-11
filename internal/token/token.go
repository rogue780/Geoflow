// Package token defines the token types for the GeoFlow language.
package token

// Type represents the type of a token.
type Type string

// Position tracks source location for error reporting.
type Position struct {
	Line   int
	Column int
	Offset int
}

// Token represents a lexical token with its type, literal value, and position.
type Token struct {
	Type    Type
	Literal string
	Pos     Position
}

const (
	// Special tokens
	ILLEGAL Type = "ILLEGAL"
	EOF     Type = "EOF"

	// Identifiers and literals
	IDENT      Type = "IDENT"
	INT        Type = "INT"
	FLOAT      Type = "FLOAT"
	STRING     Type = "STRING"
	RAW_STRING Type = "RAW_STRING"
	WKT        Type = "WKT"
	SYMBOLIC   Type = "SYMBOLIC"

	// Operators
	PLUS     Type = "+"
	MINUS    Type = "-"
	ASTERISK Type = "*"
	SLASH    Type = "/"
	PERCENT  Type = "%"
	CARET    Type = "^"
	BANG     Type = "!"

	// Comparison
	EQ     Type = "=="
	NOT_EQ Type = "!="
	LT     Type = "<"
	GT     Type = ">"
	LT_EQ  Type = "<="
	GT_EQ  Type = ">="

	// Logical
	AND Type = "&&"
	OR  Type = "||"

	// Assignment
	ASSIGN     Type = "="
	MUT_ASSIGN Type = ":="

	// Delimiters
	COMMA     Type = ","
	COLON     Type = ":"
	SEMICOLON Type = ";"
	DOT       Type = "."
	DOTDOT    Type = ".."
	ELLIPSIS  Type = "..."

	LPAREN   Type = "("
	RPAREN   Type = ")"
	LBRACKET Type = "["
	RBRACKET Type = "]"
	LBRACE   Type = "{"
	RBRACE   Type = "}"

	// Composition
	PIPE         Type = "|>"
	REVERSE_PIPE Type = "<|"
	ARROW        Type = "->"
	FAT_ARROW    Type = "=>"
	BACKSLASH    Type = "\\"
	QUESTION     Type = "?"

	// Keywords
	LET      Type = "let"
	CONST    Type = "const"
	FN       Type = "fn"
	TYPE     Type = "type"
	STRUCT   Type = "struct"
	ENUM     Type = "enum"
	IF       Type = "if"
	THEN     Type = "then"
	ELSE     Type = "else"
	MATCH    Type = "match"
	WITH     Type = "with"
	FOR      Type = "for"
	IN       Type = "in"
	WHILE    Type = "while"
	DO       Type = "do"
	RETURN   Type = "return"
	YIELD    Type = "yield"
	BREAK    Type = "break"
	CONTINUE Type = "continue"
	TRUE     Type = "true"
	FALSE    Type = "false"
	NIL      Type = "nil"
	AUTO     Type = "auto"
	IMPORT   Type = "import"
	EXPORT   Type = "export"
	MODULE   Type = "module"
	MUT      Type = "mut"
	AS       Type = "as"
	TRY      Type = "try"
	CATCH    Type = "catch"
	GEOM     Type = "geom"
	GEOG     Type = "geog"
)

var keywords = map[string]Type{
	"let":      LET,
	"const":    CONST,
	"fn":       FN,
	"type":     TYPE,
	"struct":   STRUCT,
	"enum":     ENUM,
	"if":       IF,
	"then":     THEN,
	"else":     ELSE,
	"match":    MATCH,
	"with":     WITH,
	"for":      FOR,
	"in":       IN,
	"while":    WHILE,
	"do":       DO,
	"return":   RETURN,
	"yield":    YIELD,
	"break":    BREAK,
	"continue": CONTINUE,
	"true":     TRUE,
	"false":    FALSE,
	"nil":      NIL,
	"auto":     AUTO,
	"import":   IMPORT,
	"export":   EXPORT,
	"module":   MODULE,
	"mut":      MUT,
	"as":       AS,
	"try":      TRY,
	"catch":    CATCH,
}

// LookupIdent checks if an identifier is a keyword and returns the appropriate type.
func LookupIdent(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
