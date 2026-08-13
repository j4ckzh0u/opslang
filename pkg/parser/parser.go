// Package parser 实现 OpsLang 的语法分析器
package parser

import (
	"fmt"

	"github.com/opslang/opslang/pkg/ast"
	"github.com/opslang/opslang/pkg/lexer"
)

// Parser 语法分析器
type Parser struct {
	tokens  []lexer.Token
	pos     int
	current lexer.Token
	errors  []error
}

// New 创建新的语法分析器
func New(tokens []lexer.Token) *Parser {
	p := &Parser{
		tokens: tokens,
		pos:    0,
	}
	if len(tokens) > 0 {
		p.current = tokens[0]
	}
	return p
}

// Parse 解析程序
func (p *Parser) Parse() (*ast.Program, error) {
	program := &ast.Program{
		Position: ast.Position{Line: 1, Column: 1},
	}

	// TODO: 实现完整的语法分析
	// 当前返回空程序

	if len(p.errors) > 0 {
		return nil, p.errors[0]
	}

	return program, nil
}

// ParseError 解析错误
type ParseError struct {
	Message string
	Line    int
	Column  int
	File    string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.File, e.Line, e.Column, e.Message)
}

// --- 辅助方法 ---

func (p *Parser) advance() lexer.Token {
	tok := p.current
	p.pos++
	if p.pos < len(p.tokens) {
		p.current = p.tokens[p.pos]
	}
	return tok
}

func (p *Parser) peek() lexer.Token {
	return p.current
}

func (p *Parser) expect(t lexer.TokenType) error {
	if p.current.Type != t {
		return &ParseError{
			Message: fmt.Sprintf("期望 %d，实际 %s", t, p.current.Value),
			Line:    p.current.Line,
			Column:  p.current.Column,
			File:    p.current.File,
		}
	}
	return nil
}
