// Package lexer 实现 OpsLang 的词法分析器
//
// 支持 Python 风格的缩进语法、字符串插值、管道操作符等
package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

// TokenType 词法单元类型
type TokenType int

const (
	// 特殊标记
	TOKEN_EOF TokenType = iota
	TOKEN_NEWLINE
	TOKEN_INDENT
	TOKEN_DEDENT

	// 字面量
	TOKEN_IDENT  // 标识符
	TOKEN_INT    // 整数
	TOKEN_FLOAT  // 浮点数
	TOKEN_STRING // 字符串

	// 运算符
	TOKEN_PLUS      // +
	TOKEN_MINUS     // -
	TOKEN_STAR      // *
	TOKEN_SLASH     // /
	TOKEN_PERCENT   // %
	TOKEN_ASSIGN    // =
	TOKEN_EQ        // ==
	TOKEN_NE        // !=
	TOKEN_LT        // <
	TOKEN_LE        // <=
	TOKEN_GT        // >
	TOKEN_GE        // >=
	TOKEN_AND       // &&
	TOKEN_OR        // ||
	TOKEN_NOT       // !
	TOKEN_PIPE      // |>
	TOKEN_ARROW     // =>
	TOKEN_PLUS_ASSIGN  // +=
	TOKEN_MINUS_ASSIGN // -=

	// 分隔符
	TOKEN_LPAREN    // (
	TOKEN_RPAREN    // )
	TOKEN_LBRACKET  // [
	TOKEN_RBRACKET  // ]
	TOKEN_LBRACE    // {
	TOKEN_RBRACE    // }
	TOKEN_COMMA     // ,
	TOKEN_COLON     // :
	TOKEN_DOT       // .
	TOKEN_SEMICOLON // ;

	// 关键字
	TOKEN_FN
	TOKEN_RETURN
	TOKEN_IF
	TOKEN_ELSE
	TOKEN_FOR
	TOKEN_IN
	TOKEN_WHILE
	TOKEN_BREAK
	TOKEN_CONTINUE
	TOKEN_DEFER
	TOKEN_TRY
	TOKEN_CATCH
	TOKEN_IMPORT
	TOKEN_AS
	TOKEN_TRUE
	TOKEN_FALSE
	TOKEN_NIL
	TOKEN_ENSURE
	TOKEN_FLEET
	TOKEN_ON
)

// Token 词法单元
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
	File   string
}

// keywords 关键字映射
var keywords = map[string]TokenType{
	"fn":       TOKEN_FN,
	"return":   TOKEN_RETURN,
	"if":       TOKEN_IF,
	"else":     TOKEN_ELSE,
	"for":      TOKEN_FOR,
	"in":       TOKEN_IN,
	"while":    TOKEN_WHILE,
	"break":    TOKEN_BREAK,
	"continue": TOKEN_CONTINUE,
	"defer":    TOKEN_DEFER,
	"try":      TOKEN_TRY,
	"catch":    TOKEN_CATCH,
	"import":   TOKEN_IMPORT,
	"as":       TOKEN_AS,
	"true":     TOKEN_TRUE,
	"false":    TOKEN_FALSE,
	"nil":      TOKEN_NIL,
	"ensure":   TOKEN_ENSURE,
	"fleet":    TOKEN_FLEET,
	"on":       TOKEN_ON,
}

// Lexer 词法分析器
type Lexer struct {
	source    []rune
	file      string
	pos       int
	line      int
	column    int
	tokens    []Token
	indentStack []int
	lineStart bool
}

// New 创建词法分析器
func New(source, filename string) *Lexer {
	return &Lexer{
		source:      []rune(source),
		file:        filename,
		pos:         0,
		line:        1,
		column:      1,
		indentStack: []int{0},
		lineStart:   true,
	}
}

// Tokenize 执行词法分析
func (l *Lexer) Tokenize() []Token {
	lines := strings.Split(string(l.source), "\n")

	for i, line := range lines {
		l.line = i + 1
		l.column = 1

		// 处理空行和注释行
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			if len(l.tokens) > 0 && l.tokens[len(l.tokens)-1].Type != TOKEN_NEWLINE {
				l.tokens = append(l.tokens, Token{Type: TOKEN_NEWLINE, Line: l.line, Column: 1})
			}
			continue
		}

		// 处理行首缩进
		if l.lineStart {
			indent := l.measureIndent(line)
			l.handleIndent(indent)
			l.lineStart = false
		}

		// 词法分析该行内容
		l.pos = 0
		l.source = []rune(line)
		l.scanLine()

		// 行末添加 NEWLINE
		if len(l.tokens) > 0 && l.tokens[len(l.tokens)-1].Type != TOKEN_NEWLINE {
			l.tokens = append(l.tokens, Token{Type: TOKEN_NEWLINE, Line: l.line, Column: l.column})
		}
		l.lineStart = true
	}

	// 文件结尾：弹出所有剩余缩进
	for len(l.indentStack) > 1 {
		l.indentStack = l.indentStack[:len(l.indentStack)-1]
		l.tokens = append(l.tokens, Token{Type: TOKEN_DEDENT, Line: l.line, Column: 1})
	}

	l.tokens = append(l.tokens, Token{Type: TOKEN_EOF, Line: l.line, Column: 1, File: l.file})
	return l.tokens
}

// measureIndent 测量行首缩进（空格数）
func (l *Lexer) measureIndent(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 4 // Tab = 4 spaces
		} else {
			break
		}
	}
	return count
}

// handleIndent 处理缩进变化
func (l *Lexer) handleIndent(indent int) {
	current := l.indentStack[len(l.indentStack)-1]

	if indent > current {
		l.indentStack = append(l.indentStack, indent)
		l.tokens = append(l.tokens, Token{Type: TOKEN_INDENT, Line: l.line, Column: 1})
	} else if indent < current {
		for len(l.indentStack) > 1 && l.indentStack[len(l.indentStack)-1] > indent {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			l.tokens = append(l.tokens, Token{Type: TOKEN_DEDENT, Line: l.line, Column: 1})
		}
	}
}

// scanLine 词法分析单行
func (l *Lexer) scanLine() {
	for l.pos < len(l.source) {
		l.skipWhitespace()
		if l.pos >= len(l.source) {
			break
		}

		ch := l.source[l.pos]

		// 注释
		if ch == '/' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '/' {
			break // 跳过行尾注释
		}

		// 字符串
		if ch == '"' || ch == '\'' {
			l.scanString()
			continue
		}

		// 数字
		if unicode.IsDigit(rune(ch)) {
			l.scanNumber()
			continue
		}

		// 标识符/关键字
		if unicode.IsLetter(rune(ch)) || ch == '_' {
			l.scanIdent()
			continue
		}

		// 运算符和分隔符
		l.scanOperator()
	}
}

// skipWhitespace 跳过空白（不含换行）
func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.source) {
		ch := l.source[l.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.pos++
			l.column++
		} else {
			break
		}
	}
}

// scanString 扫描字符串
func (l *Lexer) scanString() {
	startCol := l.column
	quote := l.source[l.pos]
	l.pos++
	l.column++

	var sb strings.Builder
	sb.WriteRune(quote)

	for l.pos < len(l.source) {
		ch := l.source[l.pos]

		if ch == '\\' && l.pos+1 < len(l.source) {
			// 转义字符
			sb.WriteRune(ch)
			l.pos++
			l.column++
			if l.pos < len(l.source) {
				sb.WriteRune(l.source[l.pos])
				l.pos++
				l.column++
			}
			continue
		}

		if ch == quote {
			sb.WriteRune(ch)
			l.pos++
			l.column++
			break
		}

		sb.WriteRune(ch)
		l.pos++
		l.column++
	}

	l.tokens = append(l.tokens, Token{
		Type:   TOKEN_STRING,
		Value:  sb.String(),
		Line:   l.line,
		Column: startCol,
	})
}

// scanNumber 扫描数字
func (l *Lexer) scanNumber() {
	startCol := l.column
	start := l.pos
	isFloat := false

	for l.pos < len(l.source) {
		ch := l.source[l.pos]
		if unicode.IsDigit(rune(ch)) {
			l.pos++
			l.column++
		} else if ch == '.' && !isFloat {
			// 检查后面是否是数字
			if l.pos+1 < len(l.source) && unicode.IsDigit(l.source[l.pos+1]) {
				isFloat = true
				l.pos++
				l.column++
			} else {
				break
			}
		} else {
			break
		}
	}

	value := string(l.source[start:l.pos])
	tokenType := TOKEN_INT
	if isFloat {
		tokenType = TOKEN_FLOAT
	}

	l.tokens = append(l.tokens, Token{
		Type:   tokenType,
		Value:  value,
		Line:   l.line,
		Column: startCol,
	})
}

// scanIdent 扫描标识符或关键字
func (l *Lexer) scanIdent() {
	startCol := l.column
	start := l.pos

	for l.pos < len(l.source) {
		ch := l.source[l.pos]
		if unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_' {
			l.pos++
			l.column++
		} else {
			break
		}
	}

	value := string(l.source[start:l.pos])

	// 检查是否是关键字
	tokenType := TOKEN_IDENT
	if kw, ok := keywords[value]; ok {
		tokenType = kw
	}

	l.tokens = append(l.tokens, Token{
		Type:   tokenType,
		Value:  value,
		Line:   l.line,
		Column: startCol,
	})
}

// scanOperator 扫描运算符和分隔符
func (l *Lexer) scanOperator() {
	startCol := l.column
	ch := l.source[l.pos]
	l.pos++
	l.column++

	// 双字符运算符
	if l.pos < len(l.source) {
		next := l.source[l.pos]
		twoChar := string([]rune{ch, next})

		switch twoChar {
		case "==":
			l.tokens = append(l.tokens, Token{Type: TOKEN_EQ, Value: "==", Line: l.line, Column: startCol})
			l.pos++
			l.column++
			return
		case "!=":
			l.tokens = append(l.tokens, Token{Type: TOKEN_NE, Value: "!=", Line: l.line, Column: startCol})
			l.pos++
			l.column++
			return
		case "<=":
			l.tokens = append(l.tokens, Token{Type: TOKEN_LE, Value: "<=", Line: l.line, Column: startCol})
			l.pos++
			l.column++
			return
		case ">=":
			l.tokens = append(l.tokens, Token{Type: TOKEN_GE, Value: ">=", Line: l.line, Column: startCol})
			l.pos++
			l.column++
			return
		case "&&":
			l.tokens = append(l.tokens, Token{Type: TOKEN_AND, Value: "&&", Line: l.line, Column: startCol})
			l.pos++
			l.column++
			return
		case "||":
			l.tokens = append(l.tokens, Token{Type: TOKEN_OR, Value: "||", Line: l.line, Column: startCol})
			l.pos++
			l.column++
			return
		case "=>":
			l.tokens = append(l.tokens, Token{Type: TOKEN_ARROW, Value: "=>", Line: l.line, Column: startCol})
			l.pos++
			l.column++
			return
		case "|>":
			l.tokens = append(l.tokens, Token{Type: TOKEN_PIPE, Value: "|>", Line: l.line, Column: startCol})
			l.pos++
			l.column++
			return
		case "+=":
			l.tokens = append(l.tokens, Token{Type: TOKEN_PLUS_ASSIGN, Value: "+=", Line: l.line, Column: startCol})
			l.pos++
			l.column++
			return
		case "-=":
			l.tokens = append(l.tokens, Token{Type: TOKEN_MINUS_ASSIGN, Value: "-=", Line: l.line, Column: startCol})
			l.pos++
			l.column++
			return
		}
	}

	// 单字符运算符
	var tokenType TokenType
	switch ch {
	case '+':
		tokenType = TOKEN_PLUS
	case '-':
		tokenType = TOKEN_MINUS
	case '*':
		tokenType = TOKEN_STAR
	case '/':
		tokenType = TOKEN_SLASH
	case '%':
		tokenType = TOKEN_PERCENT
	case '=':
		tokenType = TOKEN_ASSIGN
	case '<':
		tokenType = TOKEN_LT
	case '>':
		tokenType = TOKEN_GT
	case '!':
		tokenType = TOKEN_NOT
	case '(':
		tokenType = TOKEN_LPAREN
	case ')':
		tokenType = TOKEN_RPAREN
	case '[':
		tokenType = TOKEN_LBRACKET
	case ']':
		tokenType = TOKEN_RBRACKET
	case '{':
		tokenType = TOKEN_LBRACE
	case '}':
		tokenType = TOKEN_RBRACE
	case ',':
		tokenType = TOKEN_COMMA
	case ':':
		tokenType = TOKEN_COLON
	case '.':
		tokenType = TOKEN_DOT
	case ';':
		tokenType = TOKEN_SEMICOLON
	default:
		// 未知字符，跳过
		return
	}

	l.tokens = append(l.tokens, Token{
		Type:   tokenType,
		Value:  string(ch),
		Line:   l.line,
		Column: startCol,
	})
}

// String 返回 Token 类型的可读名称
func (t TokenType) String() string {
	names := map[TokenType]string{
		TOKEN_EOF:     "EOF",
		TOKEN_NEWLINE: "NEWLINE",
		TOKEN_INDENT:  "INDENT",
		TOKEN_DEDENT:  "DEDENT",
		TOKEN_IDENT:   "IDENT",
		TOKEN_INT:     "INT",
		TOKEN_FLOAT:   "FLOAT",
		TOKEN_STRING:  "STRING",
		TOKEN_PLUS:    "+",
		TOKEN_MINUS:   "-",
		TOKEN_STAR:    "*",
		TOKEN_SLASH:   "/",
		TOKEN_ASSIGN:  "=",
		TOKEN_EQ:      "==",
		TOKEN_NE:      "!=",
		TOKEN_LT:      "<",
		TOKEN_LE:      "<=",
		TOKEN_GT:      ">",
		TOKEN_GE:      ">=",
		TOKEN_LPAREN:  "(",
		TOKEN_RPAREN:  ")",
		TOKEN_LBRACE:  "{",
		TOKEN_RBRACE:  "}",
		TOKEN_COMMA:   ",",
		TOKEN_COLON:   ":",
		TOKEN_DOT:     ".",
		TOKEN_FN:      "fn",
		TOKEN_IF:      "if",
		TOKEN_ELSE:    "else",
		TOKEN_FOR:     "for",
		TOKEN_IN:      "in",
		TOKEN_RETURN:  "return",
	}
	if name, ok := names[t]; ok {
		return name
	}
	return fmt.Sprintf("TOKEN_%d", int(t))
}
