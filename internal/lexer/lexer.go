package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/opslang/opslang/internal/lexer/token"
)

// Lexer performs lexical analysis on OpsLang source code.
type Lexer struct {
	source   string
	filename string

	// Current position
	pos  int // byte offset of current char
	line int // current line (1-based)
	col  int // current column (1-based)
}

// New creates a new Lexer for the given source code.
func New(source string, filename string) *Lexer {
	return &Lexer{
		source:   source,
		filename: filename,
		pos:      0,
		line:     1,
		col:      1,
	}
}

// LexError represents a lexer error with position information.
type LexError struct {
	Pos token.Position
	Msg string
}

func (e *LexError) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Msg)
}

// Tokenize scans the source and returns all tokens, ending with EOF.
func (l *Lexer) Tokenize() ([]token.Token, error) {
	var tokens []token.Token

	for {
		tok, err := l.nextToken()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}

	return tokens, nil
}

// nextToken reads and returns the next token from the source.
func (l *Lexer) nextToken() (token.Token, error) {
	l.skipWhitespaceExceptNewline()

	if l.atEnd() {
		return l.makeToken(token.EOF, "", l.pos), nil
	}

	startLine := l.line
	startCol := l.col
	ch := l.peek()

	switch {
	case ch == '\n':
		l.advance()
		return l.makeTokenAt(token.NEWLINE, "\n", startLine, startCol), nil

	case ch == '/' && l.peekAhead(1) == '/':
		// Single-line comment
		l.skipLineComment()
		return l.nextToken()

	case ch == '/' && l.peekAhead(1) == '*':
		if err := l.skipBlockComment(); err != nil {
			return token.Token{}, err
		}
		return l.nextToken()

	case ch == '"':
		return l.scanString(startLine, startCol, false)

	case ch == '`':
		return l.scanString(startLine, startCol, true)

	case ch == '\'':
		return l.scanRuneOrCharLiteral(startLine, startCol)

	case isDigit(ch):
		return l.scanNumber(startLine, startCol)

	case isIdentStart(ch):
		return l.scanIdentOrKeyword(startLine, startCol)

	default:
		return l.scanOperatorOrDelimiter(startLine, startCol)
	}
}

// makeToken creates a token at the current position.
func (l *Lexer) makeToken(typ token.TokenType, literal string, pos int) token.Token {
	return token.Token{
		Type:    typ,
		Literal: literal,
		Pos:     token.Position{Line: l.line, Column: l.col, File: l.filename},
	}
}

// makeTokenAt creates a token at a saved position.
func (l *Lexer) makeTokenAt(typ token.TokenType, literal string, line, col int) token.Token {
	return token.Token{
		Type:    typ,
		Literal: literal,
		Pos:     token.Position{Line: line, Column: col, File: l.filename},
	}
}

// position returns the current Position.
func (l *Lexer) position() token.Position {
	return token.Position{Line: l.line, Column: l.col, File: l.filename}
}

// --- Character helpers ---

func (l *Lexer) atEnd() bool {
	return l.pos >= len(l.source)
}

// peek returns the current character without consuming it.
func (l *Lexer) peek() byte {
	if l.atEnd() {
		return 0
	}
	return l.source[l.pos]
}

// peekAhead returns the character n bytes ahead, or 0 if out of bounds.
func (l *Lexer) peekAhead(n int) byte {
	idx := l.pos + n
	if idx >= len(l.source) {
		return 0
	}
	return l.source[idx]
}

// advance consumes and returns the current character, updating line/col.
func (l *Lexer) advance() byte {
	if l.atEnd() {
		return 0
	}
	ch := l.source[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

// advanceRune consumes a full UTF-8 rune and returns it.
func (l *Lexer) advanceRune() rune {
	if l.atEnd() {
		return 0
	}
	r, size := utf8.DecodeRuneInString(l.source[l.pos:])
	l.pos += size
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r
}


// --- Whitespace and comment skipping ---

func (l *Lexer) skipWhitespaceExceptNewline() {
	for !l.atEnd() {
		ch := l.source[l.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *Lexer) skipLineComment() {
	// Skip until end of line (don't consume the newline).
	for !l.atEnd() && l.source[l.pos] != '\n' {
		l.advance()
	}
}

func (l *Lexer) skipBlockComment() error {
	startLine := l.line
	startCol := l.col
	// Skip the opening /*
	l.advance() // /
	l.advance() // *

	for {
		if l.atEnd() {
			return &LexError{
				Pos: token.Position{Line: startLine, Column: startCol, File: l.filename},
				Msg: "unterminated block comment",
			}
		}
		if l.source[l.pos] == '*' && l.peekAhead(1) == '/' {
			l.advance() // *
			l.advance() // /
			return nil
		}
		l.advance()
	}
}

// --- String scanning ---

func (l *Lexer) scanString(startLine, startCol int, raw bool) (token.Token, error) {
	// Consume opening quote
	l.advance()

	var buf strings.Builder
	for {
		if l.atEnd() {
			return token.Token{}, &LexError{
				Pos: token.Position{Line: startLine, Column: startCol, File: l.filename},
				Msg: "unterminated string literal",
			}
		}

		ch := l.source[l.pos]

		if raw {
			// Backtick string: only backtick terminates.
			if ch == '`' {
				l.advance()
				return l.makeTokenAt(token.STRING, buf.String(), startLine, startCol), nil
			}
			buf.WriteByte(ch)
			l.advance()
			continue
		}

		// Regular double-quoted string.
		if ch == '"' {
			l.advance()
			return l.makeTokenAt(token.STRING, buf.String(), startLine, startCol), nil
		}

		if ch == '\n' {
			return token.Token{}, &LexError{
				Pos: token.Position{Line: l.line, Column: l.col, File: l.filename},
				Msg: "newline in string literal",
			}
		}

		if ch == '\\' {
			l.advance() // consume backslash
			if l.atEnd() {
				return token.Token{}, &LexError{
					Pos: token.Position{Line: l.line, Column: l.col, File: l.filename},
					Msg: "unterminated escape sequence in string",
				}
			}
			escaped := l.advance()
			switch escaped {
			case 'n':
				buf.WriteByte('\n')
			case 't':
				buf.WriteByte('\t')
			case 'r':
				buf.WriteByte('\r')
			case '\\':
				buf.WriteByte('\\')
			case '"':
				buf.WriteByte('"')
			case '0':
				buf.WriteByte(0)
			default:
				return token.Token{}, &LexError{
					Pos: token.Position{Line: l.line, Column: l.col - 1, File: l.filename},
					Msg: fmt.Sprintf("unknown escape sequence '\\%c'", escaped),
				}
			}
		} else {
			buf.WriteByte(ch)
			l.advance()
		}
	}
}

// scanRuneOrCharLiteral handles 'x' style character literals as strings.
func (l *Lexer) scanRuneOrCharLiteral(startLine, startCol int) (token.Token, error) {
	// Consume opening quote
	l.advance()

	var buf strings.Builder
	for {
		if l.atEnd() {
			return token.Token{}, &LexError{
				Pos: token.Position{Line: startLine, Column: startCol, File: l.filename},
				Msg: "unterminated character literal",
			}
		}

		ch := l.source[l.pos]

		if ch == '\'' {
			l.advance()
			return l.makeTokenAt(token.STRING, buf.String(), startLine, startCol), nil
		}

		if ch == '\n' {
			return token.Token{}, &LexError{
				Pos: token.Position{Line: l.line, Column: l.col, File: l.filename},
				Msg: "newline in character literal",
			}
		}

		if ch == '\\' {
			l.advance()
			if l.atEnd() {
				return token.Token{}, &LexError{
					Pos: token.Position{Line: l.line, Column: l.col, File: l.filename},
					Msg: "unterminated escape in character literal",
				}
			}
			escaped := l.advance()
			switch escaped {
			case 'n':
				buf.WriteByte('\n')
			case 't':
				buf.WriteByte('\t')
			case 'r':
				buf.WriteByte('\r')
			case '\\':
				buf.WriteByte('\\')
			case '\'':
				buf.WriteByte('\'')
			default:
				return token.Token{}, &LexError{
					Pos: token.Position{Line: l.line, Column: l.col - 1, File: l.filename},
					Msg: fmt.Sprintf("unknown escape '\\%c'", escaped),
				}
			}
		} else {
			buf.WriteByte(ch)
			l.advance()
		}
	}
}

// --- Number scanning ---

func (l *Lexer) scanNumber(startLine, startCol int) (token.Token, error) {
	start := l.pos
	isFloat := false

	for !l.atEnd() && isDigit(l.source[l.pos]) {
		l.advance()
	}

	// Check for decimal point (but not ".." range operator).
	if !l.atEnd() && l.source[l.pos] == '.' {
		if l.peekAhead(1) != 0 && isDigit(l.peekAhead(1)) {
			isFloat = true
			l.advance() // consume '.'
			for !l.atEnd() && isDigit(l.source[l.pos]) {
				l.advance()
			}
		}
	}

	// Check for exponent.
	if !l.atEnd() && (l.source[l.pos] == 'e' || l.source[l.pos] == 'E') {
		isFloat = true
		l.advance()
		if !l.atEnd() && (l.source[l.pos] == '+' || l.source[l.pos] == '-') {
			l.advance()
		}
		if l.atEnd() || !isDigit(l.source[l.pos]) {
			return token.Token{}, &LexError{
				Pos: token.Position{Line: l.line, Column: l.col, File: l.filename},
				Msg: "invalid number: expected digit after exponent",
			}
		}
		for !l.atEnd() && isDigit(l.source[l.pos]) {
			l.advance()
		}
	}

	literal := l.source[start:l.pos]
	typ := token.INT
	if isFloat {
		typ = token.FLOAT
	}
	return l.makeTokenAt(typ, literal, startLine, startCol), nil
}

// --- Identifier / keyword scanning ---

func (l *Lexer) scanIdentOrKeyword(startLine, startCol int) (token.Token, error) {
	start := l.pos
	for !l.atEnd() && isIdentPart(l.source[l.pos]) {
		l.advance()
	}

	literal := l.source[start:l.pos]
	typ := token.LookupKeyword(literal)
	return l.makeTokenAt(typ, literal, startLine, startCol), nil
}

// --- Operator / delimiter scanning ---

func (l *Lexer) scanOperatorOrDelimiter(startLine, startCol int) (token.Token, error) {
	ch := l.advance()

	switch ch {
	case '+':
		return l.makeTokenAt(token.PLUS, "+", startLine, startCol), nil
	case '-':
		return l.makeTokenAt(token.MINUS, "-", startLine, startCol), nil
	case '*':
		return l.makeTokenAt(token.STAR, "*", startLine, startCol), nil
	case '/':
		return l.makeTokenAt(token.SLASH, "/", startLine, startCol), nil
	case '%':
		return l.makeTokenAt(token.PERCENT, "%", startLine, startCol), nil

	case '=':
		if !l.atEnd() && l.source[l.pos] == '=' {
			l.advance()
			return l.makeTokenAt(token.EQ, "==", startLine, startCol), nil
		}
		return l.makeTokenAt(token.ASSIGN, "=", startLine, startCol), nil

	case '!':
		if !l.atEnd() && l.source[l.pos] == '=' {
			l.advance()
			return l.makeTokenAt(token.NEQ, "!=", startLine, startCol), nil
		}
		return l.makeTokenAt(token.NOT, "!", startLine, startCol), nil

	case '<':
		if !l.atEnd() && l.source[l.pos] == '=' {
			l.advance()
			return l.makeTokenAt(token.LTE, "<=", startLine, startCol), nil
		}
		return l.makeTokenAt(token.LT, "<", startLine, startCol), nil

	case '>':
		if !l.atEnd() && l.source[l.pos] == '=' {
			l.advance()
			return l.makeTokenAt(token.GTE, ">=", startLine, startCol), nil
		}
		return l.makeTokenAt(token.GT, ">", startLine, startCol), nil

	case '&':
		if !l.atEnd() && l.source[l.pos] == '&' {
			l.advance()
			return l.makeTokenAt(token.AND, "&&", startLine, startCol), nil
		}
		return token.Token{}, &LexError{
			Pos: token.Position{Line: startLine, Column: startCol, File: l.filename},
			Msg: "unexpected character '&', did you mean '&&'?",
		}

	case '|':
		if !l.atEnd() && l.source[l.pos] == '|' {
			l.advance()
			return l.makeTokenAt(token.OR, "||", startLine, startCol), nil
		}
		return token.Token{}, &LexError{
			Pos: token.Position{Line: startLine, Column: startCol, File: l.filename},
			Msg: "unexpected character '|', did you mean '||'?",
		}

	case '.':
		return l.makeTokenAt(token.DOT, ".", startLine, startCol), nil
	case ',':
		return l.makeTokenAt(token.COMMA, ",", startLine, startCol), nil
	case ':':
		return l.makeTokenAt(token.COLON, ":", startLine, startCol), nil
	case ';':
		return l.makeTokenAt(token.SEMICOLON, ";", startLine, startCol), nil

	case '(':
		return l.makeTokenAt(token.LPAREN, "(", startLine, startCol), nil
	case ')':
		return l.makeTokenAt(token.RPAREN, ")", startLine, startCol), nil
	case '{':
		return l.makeTokenAt(token.LBRACE, "{", startLine, startCol), nil
	case '}':
		return l.makeTokenAt(token.RBRACE, "}", startLine, startCol), nil
	case '[':
		return l.makeTokenAt(token.LBRACKET, "[", startLine, startCol), nil
	case ']':
		return l.makeTokenAt(token.RBRACKET, "]", startLine, startCol), nil

	default:
		return token.Token{}, &LexError{
			Pos: token.Position{Line: startLine, Column: startCol, File: l.filename},
			Msg: fmt.Sprintf("unexpected character %q (U+%04X)", ch, ch),
		}
	}
}

// --- Character classification helpers ---

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentPart(ch byte) bool {
	return isIdentStart(ch) || isDigit(ch)
}

// isIdentStartRune checks if a rune can start an identifier (Unicode-aware).
func isIdentStartRune(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

// isIdentPartRune checks if a rune can continue an identifier (Unicode-aware).
func isIdentPartRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
