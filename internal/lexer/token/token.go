package token

import "fmt"

// TokenType represents the type of a token.
type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF
	NEWLINE
	SEMICOLON

	// Literals
	IDENT  // identifiers
	INT    // 123, 456
	FLOAT  // 1.23, 4.56
	STRING // "hello"
	TRUE   // true
	FALSE  // false
	NIL    // nil

	// Keywords (exactly 15)
	LET
	FN
	IF
	ELSE
	FOR
	WHILE
	RETURN
	TASK
	ON
	IMPORT
	REPORT
	ALERT

	// Operators
	PLUS    // +
	MINUS   // -
	STAR    // *
	SLASH   // /
	PERCENT // %
	EQ      // ==
	NEQ     // !=
	LT      // <
	GT      // >
	LTE     // <=
	GTE     // >=
	ASSIGN  // =
	AND     // &&
	OR      // ||
	NOT     // !

	// Delimiters / punctuation
	DOT        // .
	COMMA      // ,
	COLON      // :
	LPAREN     // (
	RPAREN     // )
	LBRACE     // {
	RBRACE     // }
	LBRACKET   // [
	RBRACKET   // ]
)

var tokenNames = map[TokenType]string{
	ILLEGAL:  "ILLEGAL",
	EOF:      "EOF",
	NEWLINE:  "NEWLINE",
	SEMICOLON: "SEMICOLON",
	IDENT:    "IDENT",
	INT:      "INT",
	FLOAT:    "FLOAT",
	STRING:   "STRING",
	TRUE:     "TRUE",
	FALSE:    "FALSE",
	NIL:      "NIL",
	LET:      "LET",
	FN:       "FN",
	IF:       "IF",
	ELSE:     "ELSE",
	FOR:      "FOR",
	WHILE:    "WHILE",
	RETURN:   "RETURN",
	TASK:     "TASK",
	ON:       "ON",
	IMPORT:   "IMPORT",
	REPORT:   "REPORT",
	ALERT:    "ALERT",
	PLUS:     "PLUS",
	MINUS:    "MINUS",
	STAR:     "STAR",
	SLASH:    "SLASH",
	PERCENT:  "PERCENT",
	EQ:       "EQ",
	NEQ:      "NEQ",
	LT:       "LT",
	GT:       "GT",
	LTE:      "LTE",
	GTE:      "GTE",
	ASSIGN:   "ASSIGN",
	AND:      "AND",
	OR:       "OR",
	NOT:      "NOT",
	DOT:      "DOT",
	COMMA:    "COMMA",
	COLON:    "COLON",
	LPAREN:   "LPAREN",
	RPAREN:   "RPAREN",
	LBRACE:   "LBRACE",
	RBRACE:   "RBRACE",
	LBRACKET: "LBRACKET",
	RBRACKET: "RBRACKET",
}

// String returns the name of the token type.
func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("TokenType(%d)", int(t))
}

// Position represents a source location.
type Position struct {
	Line   int
	Column int
	File   string
}

// String returns a human-readable position string.
func (p Position) String() string {
	if p.File != "" {
		return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
	}
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// Token represents a single lexical token.
type Token struct {
	Type    TokenType
	Literal string
	Pos     Position
}

// String returns a human-readable token representation.
func (t Token) String() string {
	return fmt.Sprintf("Token{%s, %q, %s}", t.Type, t.Literal, t.Pos)
}

// keywords maps keyword strings to their token types.
var keywords = map[string]TokenType{
	"let":    LET,
	"fn":     FN,
	"if":     IF,
	"else":   ELSE,
	"for":    FOR,
	"while":  WHILE,
	"return": RETURN,
	"task":   TASK,
	"on":     ON,
	"import": IMPORT,
	"true":   TRUE,
	"false":  FALSE,
	"nil":    NIL,
	"report": REPORT,
	"alert":  ALERT,
}

// LookupKeyword returns the token type for an identifier,
// checking if it is a keyword first.
func LookupKeyword(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// IsKeyword reports whether the given string is an OpsLang keyword.
func IsKeyword(s string) bool {
	_, ok := keywords[s]
	return ok
}
