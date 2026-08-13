// Package lexer 实现 OpsLang 的词法分析器
package lexer

// TokenType 词法单元类型
type TokenType int

const (
	// 特殊标记
	TOKEN_EOF TokenType = iota
	TOKEN_NEWLINE
	TOKEN_INDENT
	TOKEN_DEDENT

	// 字面量
	TOKEN_IDENT   // 标识符
	TOKEN_INT     // 整数 42
	TOKEN_FLOAT   // 浮点 3.14
	TOKEN_STRING  // 字符串 "hello"
	TOKEN_TRUE    // true
	TOKEN_FALSE   // false
	TOKEN_NIL     // nil

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

	// 分隔符
	TOKEN_LPAREN   // (
	TOKEN_RPAREN   // )
	TOKEN_LBRACKET // [
	TOKEN_RBRACKET // ]
	TOKEN_LBRACE   // {
	TOKEN_RBRACE   // }
	TOKEN_COMMA    // ,
	TOKEN_COLON    // :
	TOKEN_DOT      // .
	TOKEN_SEMICOLON // ;

	// 关键字
	TOKEN_FN       // fn
	TOKEN_RETURN   // return
	TOKEN_IF       // if
	TOKEN_ELSE     // else
	TOKEN_FOR      // for
	TOKEN_IN       // in
	TOKEN_WHILE    // while
	TOKEN_BREAK    // break
	TOKEN_CONTINUE // continue
	TOKEN_DEFER    // defer
	TOKEN_TRY      // try
	TOKEN_CATCH    // catch
	TOKEN_IMPORT   // import
	TOKEN_AS       // as

	// 内置关键字（运维专用）
	TOKEN_ENSURE   // ensure
	TOKEN_FLEET    // fleet
	TOKEN_ON       // on (事件处理)
)

// Token 词法单元
type Token struct {
	Type    TokenType
	Value   string
	Line    int
	Column  int
	File    string
}

// Lexer 词法分析器
type Lexer struct {
	source  []rune
	file    string
	pos     int
	line    int
	column  int
	tokens  []Token
	indent  []int // 缩进栈
}

// New 创建新的词法分析器
func New(source, filename string) *Lexer {
	return &Lexer{
		source: []rune(source),
		file:   filename,
		pos:    0,
		line:   1,
		column: 1,
		indent: []int{0},
	}
}

// Tokenize 执行词法分析，返回所有 Token
func (l *Lexer) Tokenize() []Token {
	// TODO: 实现完整的词法分析
	// 当前返回空列表，后续逐步实现
	return []Token{{Type: TOKEN_EOF, Line: 1, Column: 1, File: l.file}}
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
