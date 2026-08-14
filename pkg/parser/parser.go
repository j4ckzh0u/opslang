// Package parser 实现 OpsLang 的语法分析器
// 递归下降解析，支持缩进语法
package parser

import (
	"fmt"
	"strconv"
	"strings"

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

// New 创建语法分析器
func New(tokens []lexer.Token) *Parser {
	p := &Parser{tokens: tokens, pos: 0}
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

	p.skipNewlines()

	for !p.isAtEnd() {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.skipNewlines()
	}

	if len(p.errors) > 0 {
		return nil, p.errors[0]
	}
	return program, nil
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
	if p.pos+1 < len(p.tokens) {
		return p.tokens[p.pos+1]
	}
	return lexer.Token{Type: lexer.TOKEN_EOF}
}

func (p *Parser) isAtEnd() bool {
	return p.current.Type == lexer.TOKEN_EOF
}

func (p *Parser) check(t lexer.TokenType) bool {
	return p.current.Type == t
}

func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) expect(t lexer.TokenType) error {
	if p.current.Type != t {
		err := &ParseError{
			Message: fmt.Sprintf("期望 %s，实际 %q", t, p.current.Value),
			Line:    p.current.Line,
			Column:  p.current.Column,
			File:    p.current.File,
		}
		p.errors = append(p.errors, err)
		return err
	}
	return p.advance2()
}

func (p *Parser) advance2() error {
	p.advance()
	return nil
}

func (p *Parser) skipNewlines() {
	for p.check(lexer.TOKEN_NEWLINE) {
		p.advance()
	}
}

func (p *Parser) pos_() ast.Position {
	return ast.Position{Line: p.current.Line, Column: p.current.Column}
}

// --- 语句解析 ---

func (p *Parser) parseStatement() ast.Statement {
	p.skipNewlines()
	if p.isAtEnd() {
		return nil
	}

	switch p.current.Type {
	case lexer.TOKEN_FN:
		return p.parseFnStmt()
	case lexer.TOKEN_IF:
		return p.parseIfStmt()
	case lexer.TOKEN_FOR:
		return p.parseForStmt()
	case lexer.TOKEN_WHILE:
		return p.parseWhileStmt()
	case lexer.TOKEN_RETURN:
		return p.parseReturnStmt()
	case lexer.TOKEN_IMPORT:
		return p.parseImportStmt()
	case lexer.TOKEN_TRY:
		return p.parseTryStmt()
	case lexer.TOKEN_BREAK:
		pos := p.pos_()
		p.advance()
		return &ast.BreakStmt{Position: pos}
	case lexer.TOKEN_CONTINUE:
		pos := p.pos_()
		p.advance()
		return &ast.ContinueStmt{Position: pos}
	default:
		return p.parseExprOrAssign()
	}
}

func (p *Parser) parseFnStmt() ast.Statement {
	pos := p.pos_()
	p.advance() // consume 'fn'

	// 函数名
	name := ""
	if p.check(lexer.TOKEN_IDENT) {
		name = p.current.Value
		p.advance()
	}

	// 参数列表
	p.expect(lexer.TOKEN_LPAREN)
	params := p.parseParamList()
	p.expect(lexer.TOKEN_RPAREN)

	// 函数体（缩进块）
	body := p.parseBlock()

	return &ast.FnStmt{
		Name:     name,
		Params:   params,
		Body:     body,
		Position: pos,
	}
}

func (p *Parser) parseParamList() []ast.Parameter {
	var params []ast.Parameter
	for !p.check(lexer.TOKEN_RPAREN) && !p.isAtEnd() {
		param := ast.Parameter{}
		if p.check(lexer.TOKEN_IDENT) {
			param.Name = p.current.Value
			p.advance()
		}
		// 默认值
		if p.match(lexer.TOKEN_ASSIGN) {
			param.Default = p.parseExpression()
		}
		params = append(params, param)
		if !p.match(lexer.TOKEN_COMMA) {
			break
		}
	}
	return params
}

func (p *Parser) parseIfStmt() ast.Statement {
	pos := p.pos_()
	p.advance() // consume 'if'

	condition := p.parseExpression()
	body := p.parseBlock()

	var elseIf []*ast.IfStmt
	var elseBody []ast.Statement

	// 检查 else if / else
	for p.check(lexer.TOKEN_NEWLINE) {
		p.advance()
	}

	if p.check(lexer.TOKEN_ELSE) {
		p.advance()
		if p.check(lexer.TOKEN_IF) {
			// else if
			p.advance()
			elseIfCond := p.parseExpression()
			elseIfBody := p.parseBlock()
			elseIf = append(elseIf, &ast.IfStmt{
				Condition: elseIfCond,
				Body:      elseIfBody,
				Position:  pos,
			})
		} else {
			// else
			elseBody = p.parseBlock()
		}
	}

	return &ast.IfStmt{
		Condition: condition,
		Body:      body,
		ElseIf:    elseIf,
		Else:      elseBody,
		Position:  pos,
	}
}

func (p *Parser) parseForStmt() ast.Statement {
	pos := p.pos_()
	p.advance() // consume 'for'

	// for var in expr { ... }
	varName := ""
	if p.check(lexer.TOKEN_IDENT) {
		varName = p.current.Value
		p.advance()
	}
	p.expect(lexer.TOKEN_IN)
	iterable := p.parseExpression()
	body := p.parseBlock()

	return &ast.ForStmt{
		Variable: varName,
		Iterable: iterable,
		Body:     body,
		Position: pos,
	}
}

func (p *Parser) parseWhileStmt() ast.Statement {
	pos := p.pos_()
	p.advance() // consume 'while'

	condition := p.parseExpression()
	body := p.parseBlock()

	return &ast.WhileStmt{
		Condition: condition,
		Body:      body,
		Position:  pos,
	}
}

func (p *Parser) parseReturnStmt() ast.Statement {
	pos := p.pos_()
	p.advance() // consume 'return'

	var value ast.Expression
	if !p.check(lexer.TOKEN_NEWLINE) && !p.isAtEnd() {
		value = p.parseExpression()
	}

	return &ast.ReturnStmt{Value: value, Position: pos}
}

func (p *Parser) parseImportStmt() ast.Statement {
	pos := p.pos_()
	p.advance() // consume 'import'

	path := ""
	if p.check(lexer.TOKEN_STRING) {
		path = p.current.Value
		// 引号已由 lexer 去除
		p.advance()
	} else if p.check(lexer.TOKEN_IDENT) {
		path = p.current.Value
		p.advance()
		// 处理 a.b.c 形式
		for p.match(lexer.TOKEN_DOT) {
			if p.check(lexer.TOKEN_IDENT) {
				path += "." + p.current.Value
				p.advance()
			}
		}
	}

	alias := ""
	if p.match(lexer.TOKEN_AS) {
		if p.check(lexer.TOKEN_IDENT) {
			alias = p.current.Value
			p.advance()
		}
	}

	return &ast.ImportStmt{Path: path, Alias: alias, Position: pos}
}

func (p *Parser) parseTryStmt() ast.Statement {
	pos := p.pos_()
	p.advance() // consume 'try'

	body := p.parseBlock()

	var catchVar string
	var catchBody []ast.Statement

	if p.check(lexer.TOKEN_CATCH) {
		p.advance()
		if p.check(lexer.TOKEN_IDENT) {
			catchVar = p.current.Value
			p.advance()
		}
		catchBody = p.parseBlock()
	}

	return &ast.TryStmt{
		Body:     body,
		CatchVar: catchVar,
		Catch:    catchBody,
		Position: pos,
	}
}

// parseExprOrAssign 解析表达式或赋值语句
func (p *Parser) parseExprOrAssign() ast.Statement {
	pos := p.pos_()
	expr := p.parseExpression()

	// 检查是否是赋值
	if p.match(lexer.TOKEN_ASSIGN) {
		value := p.parseExpression()
		return &ast.AssignStmt{
			Target:   expr,
			Value:    value,
			Position: pos,
		}
	}

	// 复合赋值 += -=
	if p.match(lexer.TOKEN_PLUS_ASSIGN) {
		value := p.parseExpression()
		return &ast.AssignStmt{
			Target: expr,
			Value: &ast.BinaryExpr{
				Op:       "+",
				Left:     expr,
				Right:    value,
				Position: pos,
			},
			Position: pos,
		}
	}
	if p.match(lexer.TOKEN_MINUS_ASSIGN) {
		value := p.parseExpression()
		return &ast.AssignStmt{
			Target: expr,
			Value: &ast.BinaryExpr{
				Op:       "-",
				Left:     expr,
				Right:    value,
				Position: pos,
			},
			Position: pos,
		}
	}

	return &ast.ExprStmt{Expr: expr, Position: pos}
}

// parseBlock 解析缩进块
func (p *Parser) parseBlock() []ast.Statement {
	var stmts []ast.Statement

	// 跳过换行
	p.skipNewlines()

	// 检查是否有 INDENT
	if p.match(lexer.TOKEN_INDENT) {
		for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
			p.skipNewlines()
			if p.check(lexer.TOKEN_DEDENT) || p.isAtEnd() {
				break
			}
			stmt := p.parseStatement()
			if stmt != nil {
				stmts = append(stmts, stmt)
			}
			p.skipNewlines()
		}
		p.match(lexer.TOKEN_DEDENT) // consume DEDENT
	} else {
		// 花括号块
		if p.match(lexer.TOKEN_LBRACE) {
			for !p.check(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
				p.skipNewlines()
				if p.check(lexer.TOKEN_RBRACE) {
					break
				}
				stmt := p.parseStatement()
				if stmt != nil {
					stmts = append(stmts, stmt)
				}
				p.skipNewlines()
			}
			p.match(lexer.TOKEN_RBRACE)
		}
	}

	return stmts
}

// --- 表达式解析（优先级从低到高）---

func (p *Parser) parseExpression() ast.Expression {
	return p.parsePipe()
}

func (p *Parser) parsePipe() ast.Expression {
	left := p.parseOr()
	for p.match(lexer.TOKEN_PIPE) {
		pos := p.pos_()
		right := p.parseOr()
		left = &ast.PipeExpr{Left: left, Right: right, Position: pos}
	}
	return left
}

func (p *Parser) parseOr() ast.Expression {
	left := p.parseAnd()
	for p.match(lexer.TOKEN_OR) {
		pos := p.pos_()
		right := p.parseAnd()
		left = &ast.BinaryExpr{Op: "||", Left: left, Right: right, Position: pos}
	}
	return left
}

func (p *Parser) parseAnd() ast.Expression {
	left := p.parseEquality()
	for p.match(lexer.TOKEN_AND) {
		pos := p.pos_()
		right := p.parseEquality()
		left = &ast.BinaryExpr{Op: "&&", Left: left, Right: right, Position: pos}
	}
	return left
}

func (p *Parser) parseEquality() ast.Expression {
	left := p.parseComparison()
	for {
		if p.match(lexer.TOKEN_EQ) {
			pos := p.pos_()
			right := p.parseComparison()
			left = &ast.BinaryExpr{Op: "==", Left: left, Right: right, Position: pos}
		} else if p.match(lexer.TOKEN_NE) {
			pos := p.pos_()
			right := p.parseComparison()
			left = &ast.BinaryExpr{Op: "!=", Left: left, Right: right, Position: pos}
		} else {
			break
		}
	}
	return left
}

func (p *Parser) parseComparison() ast.Expression {
	left := p.parseAddition()
	for {
		if p.match(lexer.TOKEN_LT) {
			pos := p.pos_()
			right := p.parseAddition()
			left = &ast.BinaryExpr{Op: "<", Left: left, Right: right, Position: pos}
		} else if p.match(lexer.TOKEN_LE) {
			pos := p.pos_()
			right := p.parseAddition()
			left = &ast.BinaryExpr{Op: "<=", Left: left, Right: right, Position: pos}
		} else if p.match(lexer.TOKEN_GT) {
			pos := p.pos_()
			right := p.parseAddition()
			left = &ast.BinaryExpr{Op: ">", Left: left, Right: right, Position: pos}
		} else if p.match(lexer.TOKEN_GE) {
			pos := p.pos_()
			right := p.parseAddition()
			left = &ast.BinaryExpr{Op: ">=", Left: left, Right: right, Position: pos}
		} else {
			break
		}
	}
	return left
}

func (p *Parser) parseAddition() ast.Expression {
	left := p.parseMultiplication()
	for {
		if p.match(lexer.TOKEN_PLUS) {
			pos := p.pos_()
			right := p.parseMultiplication()
			left = &ast.BinaryExpr{Op: "+", Left: left, Right: right, Position: pos}
		} else if p.match(lexer.TOKEN_MINUS) {
			pos := p.pos_()
			right := p.parseMultiplication()
			left = &ast.BinaryExpr{Op: "-", Left: left, Right: right, Position: pos}
		} else {
			break
		}
	}
	return left
}

func (p *Parser) parseMultiplication() ast.Expression {
	left := p.parseUnary()
	for {
		if p.match(lexer.TOKEN_STAR) {
			pos := p.pos_()
			right := p.parseUnary()
			left = &ast.BinaryExpr{Op: "*", Left: left, Right: right, Position: pos}
		} else if p.match(lexer.TOKEN_SLASH) {
			pos := p.pos_()
			right := p.parseUnary()
			left = &ast.BinaryExpr{Op: "/", Left: left, Right: right, Position: pos}
		} else if p.match(lexer.TOKEN_PERCENT) {
			pos := p.pos_()
			right := p.parseUnary()
			left = &ast.BinaryExpr{Op: "%", Left: left, Right: right, Position: pos}
		} else {
			break
		}
	}
	return left
}

func (p *Parser) parseUnary() ast.Expression {
	if p.match(lexer.TOKEN_NOT) {
		pos := p.pos_()
		operand := p.parseUnary()
		return &ast.UnaryExpr{Op: "!", Operand: operand, Position: pos}
	}
	if p.match(lexer.TOKEN_MINUS) {
		pos := p.pos_()
		operand := p.parseUnary()
		return &ast.UnaryExpr{Op: "-", Operand: operand, Position: pos}
	}
	return p.parseCallAndAccess()
}

func (p *Parser) parseCallAndAccess() ast.Expression {
	expr := p.parsePrimary()

	for {
		if p.check(lexer.TOKEN_LPAREN) {
			// 函数调用
			pos := p.pos_()
			p.advance()
			args := p.parseArgList()
			p.expect(lexer.TOKEN_RPAREN)
			expr = &ast.CallExpr{Callee: expr, Args: args, Position: pos}
		} else if p.match(lexer.TOKEN_DOT) {
			// 成员访问
			pos := p.pos_()
			if p.check(lexer.TOKEN_IDENT) {
				member := p.current.Value
				p.advance()
				expr = &ast.MemberExpr{Object: expr, Member: member, Position: pos}
			}
		} else if p.match(lexer.TOKEN_LBRACKET) {
			// 索引访问
			pos := p.pos_()
			index := p.parseExpression()
			p.expect(lexer.TOKEN_RBRACKET)
			expr = &ast.IndexExpr{Object: expr, Index: index, Position: pos}
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) parseArgList() []ast.Expression {
	var args []ast.Expression
	for !p.check(lexer.TOKEN_RPAREN) && !p.isAtEnd() {
		args = append(args, p.parseExpression())
		if !p.match(lexer.TOKEN_COMMA) {
			break
		}
	}
	return args
}

func (p *Parser) parsePrimary() ast.Expression {
	pos := p.pos_()

	switch p.current.Type {
	case lexer.TOKEN_INT:
		val, _ := strconv.ParseInt(p.current.Value, 10, 64)
		p.advance()
		return &ast.IntLitExpr{Value: val, Position: pos}

	case lexer.TOKEN_FLOAT:
		val, _ := strconv.ParseFloat(p.current.Value, 64)
		p.advance()
		return &ast.FloatLitExpr{Value: val, Position: pos}

	case lexer.TOKEN_STRING:
		val := p.current.Value
		quote := p.current.Quote
		// 引号已由 lexer 去除
		p.advance()
		// 仅双引号字符串支持插值
		if quote == '"' && strings.Contains(val, "{") {
			return p.parseInterpolatedString(val, pos)
		}
		return &ast.StringLitExpr{Value: val, Position: pos}

	case lexer.TOKEN_TRUE:
		p.advance()
		return &ast.BoolLitExpr{Value: true, Position: pos}

	case lexer.TOKEN_FALSE:
		p.advance()
		return &ast.BoolLitExpr{Value: false, Position: pos}

	case lexer.TOKEN_NIL:
		p.advance()
		return &ast.NilLitExpr{Position: pos}

	case lexer.TOKEN_IDENT:
		name := p.current.Value
		p.advance()
		return &ast.IdentExpr{Name: name, Position: pos}

	case lexer.TOKEN_LPAREN:
		p.advance()
		expr := p.parseExpression()
		p.expect(lexer.TOKEN_RPAREN)
		return expr

	case lexer.TOKEN_LBRACKET:
		return p.parseArrayLiteral()

	case lexer.TOKEN_LBRACE:
		return p.parseMapLiteral()

	case lexer.TOKEN_FN:
		return p.parseLambdaExpr()

	default:
		// 错误恢复
		p.errors = append(p.errors, &ParseError{
			Message: fmt.Sprintf("意外的 token: %q", p.current.Value),
			Line:    p.current.Line,
			Column:  p.current.Column,
		})
		p.advance()
		return &ast.NilLitExpr{Position: pos}
	}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	pos := p.pos_()
	p.advance() // consume '['
	var elements []ast.Expression
	for !p.check(lexer.TOKEN_RBRACKET) && !p.isAtEnd() {
		elements = append(elements, p.parseExpression())
		if !p.match(lexer.TOKEN_COMMA) {
			break
		}
	}
	p.expect(lexer.TOKEN_RBRACKET)
	return &ast.ArrayLitExpr{Elements: elements, Position: pos}
}

func (p *Parser) parseMapLiteral() ast.Expression {
	pos := p.pos_()
	p.advance() // consume '{'
	var keys, values []ast.Expression
	for !p.check(lexer.TOKEN_RBRACE) && !p.isAtEnd() {
		key := p.parseExpression()
		p.expect(lexer.TOKEN_COLON)
		value := p.parseExpression()
		keys = append(keys, key)
		values = append(values, value)
		if !p.match(lexer.TOKEN_COMMA) {
			break
		}
	}
	p.expect(lexer.TOKEN_RBRACE)
	return &ast.MapLitExpr{Keys: keys, Values: values, Position: pos}
}

func (p *Parser) parseLambdaExpr() ast.Expression {
	pos := p.pos_()
	p.advance() // consume 'fn'

	p.expect(lexer.TOKEN_LPAREN)
	params := p.parseParamList()
	p.expect(lexer.TOKEN_RPAREN)
	p.expect(lexer.TOKEN_ARROW)

	body := p.parseExpression()
	return &ast.LambdaExpr{Params: params, Body: body, Position: pos}
}

func (p *Parser) parseInterpolatedString(raw string, pos ast.Position) ast.Expression {
	parts := parseInterpolation(raw)
	var exprs []ast.Expression
	for _, part := range parts {
		if part.isExpr {
			// 对表达式文本进行子解析
			subLexer := lexer.New(part.text, "<interpolation>")
			tokens := subLexer.Tokenize()
			subParser := New(tokens)
			expr := subParser.parseExpression()
			if len(subParser.errors) > 0 {
				// 解析失败时回退为标识符
				exprs = append(exprs, &ast.IdentExpr{Name: part.text, Position: pos})
			} else {
				exprs = append(exprs, expr)
			}
		} else {
			exprs = append(exprs, &ast.StringLitExpr{Value: part.text, Position: pos})
		}
	}
	return &ast.InterpolatedStringExpr{Parts: exprs, Position: pos}
}

type interpPart struct {
	text   string
	isExpr bool
}

func parseInterpolation(s string) []interpPart {
	var parts []interpPart
	var current strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '{' {
			if current.Len() > 0 {
				parts = append(parts, interpPart{text: current.String()})
				current.Reset()
			}
			// 找到对应的 }
			j := i + 1
			for j < len(s) && s[j] != '}' {
				j++
			}
			if j < len(s) {
				parts = append(parts, interpPart{text: s[i+1 : j], isExpr: true})
				i = j + 1
			} else {
				current.WriteByte(s[i])
				i++
			}
		} else {
			current.WriteByte(s[i])
			i++
		}
	}
	if current.Len() > 0 {
		parts = append(parts, interpPart{text: current.String()})
	}
	return parts
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
