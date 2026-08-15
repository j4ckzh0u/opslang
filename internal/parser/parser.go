// Package parser implements a recursive descent parser with Pratt expression
// parsing for the OpsLang language.
package parser

import (
	"fmt"
	"strconv"

	"github.com/opslang/opslang/internal/ast"
	"github.com/opslang/opslang/internal/lexer"
	"github.com/opslang/opslang/internal/lexer/token"
)

// ---------------------------------------------------------------------------
// ParseError
// ---------------------------------------------------------------------------

// ParseError represents a syntax error with source position.
type ParseError struct {
	Pos token.Position
	Msg string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Msg)
}

// ---------------------------------------------------------------------------
// Parser
// ---------------------------------------------------------------------------

// Parser transforms a token stream into an AST.
type Parser struct {
	tokens   []token.Token
	pos      int
	filename string
}

// New creates a Parser for the given OpsLang source.
// It tokenizes the source immediately; lexer errors are returned as parse
// errors from Parse().
func New(source string, filename string) *Parser {
	l := lexer.New(source, filename)
	tokens, err := l.Tokenize()
	if err != nil {
		// Store a minimal token stream with the error surfaced at Parse time.
		tokens = []token.Token{{Type: token.ILLEGAL, Literal: err.Error()}}
	}
	return &Parser{
		tokens:   tokens,
		pos:      0,
		filename: filename,
	}
}

// Parse parses the full source and returns the Program AST.
func (p *Parser) Parse() (*ast.Program, error) {
	if len(p.tokens) > 0 && p.tokens[0].Type == token.ILLEGAL {
		return nil, &ParseError{
			Pos: p.tokens[0].Pos,
			Msg: p.tokens[0].Literal,
		}
	}

	firstPos := p.current().Pos
	prog := &ast.Program{Position: astPos(firstPos)}

	for p.current().Type != token.EOF {
		p.skipNewlines()
		if p.current().Type == token.EOF {
			break
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		prog.Statements = append(prog.Statements, stmt)
		// Consume one trailing newline (statement terminator).
		if p.current().Type == token.NEWLINE {
			p.advance()
		}
	}

	return prog, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func astPos(p token.Position) ast.Position {
	return ast.Position{Line: p.Line, Column: p.Column, File: p.File}
}

func (p *Parser) current() token.Token {
	if p.pos >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1] // EOF sentinel
	}
	return p.tokens[p.pos]
}

func (p *Parser) peek() token.Token {
	next := p.pos + 1
	if next >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[next]
}

func (p *Parser) advance() token.Token {
	tok := p.current()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

func (p *Parser) expect(typ token.TokenType) (token.Token, error) {
	tok := p.current()
	if tok.Type != typ {
		return tok, &ParseError{
			Pos: tok.Pos,
			Msg: fmt.Sprintf("expected %s, got %s", typ, tok.Type),
		}
	}
	p.advance()
	return tok, nil
}

func (p *Parser) skipNewlines() {
	for p.current().Type == token.NEWLINE || p.current().Type == token.SEMICOLON {
		p.advance()
	}
}

func (p *Parser) atStatementEnd() bool {
	switch p.current().Type {
	case token.NEWLINE, token.SEMICOLON, token.EOF, token.RBRACE:
		return true
	}
	return false
}

func (p *Parser) errorf(pos token.Position, format string, args ...interface{}) error {
	return &ParseError{
		Pos: pos,
		Msg: fmt.Sprintf(format, args...),
	}
}

// ---------------------------------------------------------------------------
// Statement parsing
// ---------------------------------------------------------------------------

func (p *Parser) parseStatement() (ast.Statement, error) {
	var stmt ast.Statement
	var err error

	switch p.current().Type {
	case token.LET:
		stmt, err = p.parseLetStatement()
	case token.FN:
		stmt, err = p.parseFnStatement()
	case token.IF:
		stmt, err = p.parseIfStatement()
	case token.FOR:
		stmt, err = p.parseForStatement()
	case token.WHILE:
		stmt, err = p.parseWhileStatement()
	case token.RETURN:
		stmt, err = p.parseReturnStatement()
	case token.TASK:
		stmt, err = p.parseTaskStatement()
	case token.IMPORT:
		stmt, err = p.parseImportStatement()
	case token.REPORT:
		stmt, err = p.parseReportStatement()
	case token.ALERT:
		stmt, err = p.parseAlertStatement()
	case token.METRIC:
		stmt, err = p.parseMetricStatement()
	case token.LOG:
		stmt, err = p.parseLogStatement()
	case token.ENSURE:
		stmt, err = p.parseEnsureStatement()
	case token.PARALLEL:
		stmt, err = p.parseParallelStatement()
	default:
		// Expression statement, possibly an assignment.
		stmt, err = p.parseExpressionOrAssignStatement()
	}

	if err != nil {
		return nil, err
	}

	// Consume trailing newlines (statement terminator).
	for p.current().Type == token.NEWLINE {
		p.advance()
	}

	return stmt, nil
}

// --- Let -------------------------------------------------------------------

func (p *Parser) parseLetStatement() (*ast.LetStatement, error) {
	pos := p.current().Pos
	p.advance() // consume 'let'

	nameTok, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.ASSIGN); err != nil {
		return nil, err
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.LetStatement{
		Position: astPos(pos),
		Name:     &ast.Identifier{Position: astPos(nameTok.Pos), Name: nameTok.Literal},
		Value:    value,
	}, nil
}

// --- Fn --------------------------------------------------------------------

func (p *Parser) parseFnStatement() (*ast.FnStatement, error) {
	pos := p.current().Pos
	p.advance() // consume 'fn'

	nameTok, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.LPAREN); err != nil {
		return nil, err
	}

	params, err := p.parseParameters()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.RPAREN); err != nil {
		return nil, err
	}

	body, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	return &ast.FnStatement{
		Position: astPos(pos),
		Name:     &ast.Identifier{Position: astPos(nameTok.Pos), Name: nameTok.Literal},
		Params:   params,
		Body:     body,
	}, nil
}

func (p *Parser) parseParameters() ([]ast.Parameter, error) {
	var params []ast.Parameter
	for p.current().Type != token.RPAREN && p.current().Type != token.EOF {
		if len(params) > 0 {
			if _, err := p.expect(token.COMMA); err != nil {
				return nil, err
			}
		}
		nameTok, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		param := ast.Parameter{
			Name: &ast.Identifier{Position: astPos(nameTok.Pos), Name: nameTok.Literal},
		}
		// Optional default value.
		if p.current().Type == token.ASSIGN {
			p.advance()
			def, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			param.Default = def
		}
		params = append(params, param)
	}
	return params, nil
}

// --- Block -----------------------------------------------------------------

func (p *Parser) parseBlockStatement() (*ast.BlockStatement, error) {
	pos := p.current().Pos
	if _, err := p.expect(token.LBRACE); err != nil {
		return nil, err
	}
	p.skipNewlines()

	var stmts []ast.Statement
	for p.current().Type != token.RBRACE && p.current().Type != token.EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
		p.skipNewlines()
	}

	if _, err := p.expect(token.RBRACE); err != nil {
		return nil, err
	}

	return &ast.BlockStatement{
		Position:   astPos(pos),
		Statements: stmts,
	}, nil
}

// --- If --------------------------------------------------------------------

func (p *Parser) parseIfStatement() (*ast.IfStatement, error) {
	pos := p.current().Pos
	p.advance() // consume 'if'
	p.skipNewlines()

	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	p.skipNewlines()
	body, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	stmt := &ast.IfStatement{
		Position:  astPos(pos),
		Condition: cond,
		Body:      body,
	}

	// Optional else / else-if.
	p.skipNewlines()
	if p.current().Type == token.ELSE {
		p.advance()
		p.skipNewlines()
		if p.current().Type == token.IF {
			elseIf, err := p.parseIfStatement()
			if err != nil {
				return nil, err
			}
			stmt.ElseClause = elseIf
		} else {
			elseBody, err := p.parseBlockStatement()
			if err != nil {
				return nil, err
			}
			stmt.ElseClause = elseBody
		}
	}

	return stmt, nil
}

// --- For -------------------------------------------------------------------

func (p *Parser) parseForStatement() (*ast.ForStatement, error) {
	pos := p.current().Pos
	p.advance() // consume 'for'

	// Init part (let or expression, possibly assignment).
	init, err := p.parseForPart()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.SEMICOLON); err != nil {
		return nil, err
	}

	// Condition.
	p.skipNewlines()
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// The condition expression stops at the semicolon (it's not an infix op).
	if _, err := p.expect(token.SEMICOLON); err != nil {
		return nil, err
	}

	// Post part.
	p.skipNewlines()
	post, err := p.parseForPart()
	if err != nil {
		return nil, err
	}

	p.skipNewlines()
	body, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	return &ast.ForStatement{
		Position:  astPos(pos),
		Init:      init,
		Condition: cond,
		Post:      post,
		Body:      body,
	}, nil
}

// parseForPart parses a for-loop clause (init or post) as a statement.
// It does NOT consume the trailing semicolon.
func (p *Parser) parseForPart() (ast.Statement, error) {
	stmt, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	// parseStatement already consumed trailing newlines; the current token
	// should now be SEMICOLON (or LBRACE for the post part).
	return stmt, nil
}

// --- While -----------------------------------------------------------------

func (p *Parser) parseWhileStatement() (*ast.WhileStatement, error) {
	pos := p.current().Pos
	p.advance() // consume 'while'
	p.skipNewlines()

	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	p.skipNewlines()
	body, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	return &ast.WhileStatement{
		Position:  astPos(pos),
		Condition: cond,
		Body:      body,
	}, nil
}

// --- Return ----------------------------------------------------------------

func (p *Parser) parseReturnStatement() (*ast.ReturnStatement, error) {
	pos := p.current().Pos
	p.advance() // consume 'return'

	// Bare return when at statement boundary.
	if p.atStatementEnd() {
		return &ast.ReturnStatement{Position: astPos(pos)}, nil
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.ReturnStatement{
		Position: astPos(pos),
		Value:    value,
	}, nil
}

// --- Task ------------------------------------------------------------------

func (p *Parser) parseTaskStatement() (*ast.TaskStatement, error) {
	pos := p.current().Pos
	p.advance() // consume 'task'

	nameTok, err := p.expect(token.STRING)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.ON); err != nil {
		return nil, err
	}

	targets, err := p.parseTargets()
	if err != nil {
		return nil, err
	}

	p.skipNewlines()
	body, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	return &ast.TaskStatement{
		Position: astPos(pos),
		Name:     nameTok.Literal,
		Targets:  targets,
		Body:     body,
	}, nil
}

func (p *Parser) parseTargets() (*ast.TargetClause, error) {
	pos := p.current().Pos

	// If it's a single IDENT (not followed by COMMA), it's a variable ref.
	if p.current().Type == token.IDENT && p.peek().Type != token.COMMA {
		tok := p.advance()
		return &ast.TargetClause{
			Position: astPos(pos),
			Var:      &ast.Identifier{Position: astPos(tok.Pos), Name: tok.Literal},
		}, nil
	}

	// Otherwise, comma-separated expressions.
	var hosts []ast.Expression
	for {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, expr)
		if p.current().Type != token.COMMA {
			break
		}
		p.advance() // consume ','
	}

	return &ast.TargetClause{
		Position: astPos(pos),
		Hosts:    hosts,
	}, nil
}

// --- Import ----------------------------------------------------------------

func (p *Parser) parseImportStatement() (*ast.ImportStatement, error) {
	pos := p.current().Pos
	p.advance() // consume 'import'

	pathTok, err := p.expect(token.STRING)
	if err != nil {
		return nil, err
	}

	return &ast.ImportStatement{
		Position: astPos(pos),
		Path:     pathTok.Literal,
	}, nil
}

// --- Report ----------------------------------------------------------------

func (p *Parser) parseReportStatement() (*ast.ReportStatement, error) {
	pos := p.current().Pos
	p.advance() // consume 'report'

	if _, err := p.expect(token.LBRACE); err != nil {
		return nil, err
	}
	p.skipNewlines()

	var fields []ast.ReportField
	for p.current().Type != token.RBRACE && p.current().Type != token.EOF {
		keyTok, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(token.COLON); err != nil {
			return nil, err
		}
		val, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		fields = append(fields, ast.ReportField{
			Key:   keyTok.Literal,
			Value: val,
		})
		if p.current().Type == token.COMMA {
			p.advance()
		}
		p.skipNewlines()
	}

	if _, err := p.expect(token.RBRACE); err != nil {
		return nil, err
	}

	return &ast.ReportStatement{
		Position: astPos(pos),
		Fields:   fields,
	}, nil
}

// --- Alert -----------------------------------------------------------------

func (p *Parser) parseAlertStatement() (*ast.AlertStatement, error) {
	pos := p.current().Pos
	p.advance() // consume 'alert'

	if _, err := p.expect(token.LPAREN); err != nil {
		return nil, err
	}

	msg, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(token.RPAREN); err != nil {
		return nil, err
	}

	return &ast.AlertStatement{
		Position: astPos(pos),
		Message:  msg,
	}, nil
}

// --- Metric ----------------------------------------------------------------

func (p *Parser) parseMetricStatement() (*ast.MetricStatement, error) {
	pos := p.current().Pos
	p.advance() // consume 'metric'

	if _, err := p.expect(token.LPAREN); err != nil {
		return nil, err
	}

	name, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(token.COMMA); err != nil {
		return nil, err
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	var labels ast.Expression
	if p.current().Type == token.COMMA {
		p.advance() // consume ','
		labels, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	if _, err := p.expect(token.RPAREN); err != nil {
		return nil, err
	}

	return &ast.MetricStatement{
		Position: astPos(pos),
		Name:     name,
		Value:    value,
		Labels:   labels,
	}, nil
}

// --- Log -------------------------------------------------------------------

func (p *Parser) parseLogStatement() (*ast.LogStatement, error) {
	pos := p.current().Pos
	p.advance() // consume 'log'

	if _, err := p.expect(token.LPAREN); err != nil {
		return nil, err
	}

	msg, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(token.RPAREN); err != nil {
		return nil, err
	}

	return &ast.LogStatement{
		Position: astPos(pos),
		Message:  msg,
	}, nil
}

// --- Parallel ---------------------------------------------------------------

func (p *Parser) parseParallelStatement() (*ast.ParallelStatement, error) {
	pos := p.current().Pos
	p.advance() // consume 'parallel'
	p.skipNewlines()

	body, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	return &ast.ParallelStatement{
		Position: astPos(pos),
		Body:     body,
	}, nil
}

// --- Ensure ----------------------------------------------------------------

func (p *Parser) parseEnsureStatement() (*ast.EnsureStatement, error) {
	pos := p.current().Pos
	p.advance() // consume 'ensure'
	p.skipNewlines()

	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	p.skipNewlines()
	body, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	stmt := &ast.EnsureStatement{
		Position:  astPos(pos),
		Condition: cond,
		Body:      body,
	}

	// Optional notify clause: check for an IDENT "notify" after the block
	p.skipNewlines()
	if p.current().Type == token.IDENT && p.current().Literal == "notify" {
		p.advance() // consume 'notify'
		p.skipNewlines()
		notifyExpr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		stmt.Notify = notifyExpr
	}

	return stmt, nil
}

// --- Expression / Assignment statement ------------------------------------

func (p *Parser) parseExpressionOrAssignStatement() (ast.Statement, error) {
	pos := p.current().Pos
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Check for assignment.
	if p.current().Type == token.ASSIGN {
		p.advance() // consume '='
		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &ast.AssignStatement{
			Position: astPos(pos),
			Target:   expr,
			Value:    value,
		}, nil
	}

	return &ast.ExpressionStatement{
		Position: astPos(pos),
		Expr:     expr,
	}, nil
}

// ---------------------------------------------------------------------------
// Expression parsing  (Pratt / top-down operator precedence)
//
// Precedence (low to high):
//   1  ||
//   2  &&
//   3  == !=
//   4  < > <= >=
//   5  + -
//   6  * / %
//   7  unary prefix  ! -
//   8  postfix  () [] .
// ---------------------------------------------------------------------------

type precedence int

const (
	precNone      precedence = 0
	precOr        precedence = 1
	precAnd       precedence = 2
	precEquality  precedence = 3
	precCompare   precedence = 4
	precAdditive  precedence = 5
	precMultiplic precedence = 6
)

func infixPrecedence(typ token.TokenType) precedence {
	switch typ {
	case token.OR:
		return precOr
	case token.AND:
		return precAnd
	case token.EQ, token.NEQ:
		return precEquality
	case token.LT, token.GT, token.LTE, token.GTE:
		return precCompare
	case token.PLUS, token.MINUS:
		return precAdditive
	case token.STAR, token.SLASH, token.PERCENT:
		return precMultiplic
	}
	return precNone
}

// parseExpression parses an expression. It does NOT consume trailing newlines.
func (p *Parser) parseExpression() (ast.Expression, error) {
	return p.parsePratt(precNone)
}

func (p *Parser) parsePratt(minPrec precedence) (ast.Expression, error) {
	// --- Prefix ---
	left, err := p.parsePrefix()
	if err != nil {
		return nil, err
	}

	// --- Postfix / infix loop ---
	for {
		// Postfix operators (call, index, member) bind tightest.
		switch p.current().Type {
		case token.LPAREN:
			p.advance()
			args, err := p.parseCallArgs()
			if err != nil {
				return nil, err
			}
			left = &ast.CallExpression{
				Position: left.Pos(),
				Function: left,
				Args:     args,
			}
			continue
		case token.LBRACKET:
			p.advance()
			p.skipNewlines()
			idx, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			p.skipNewlines()
			if _, err := p.expect(token.RBRACKET); err != nil {
				return nil, err
			}
			left = &ast.IndexExpression{
				Position: left.Pos(),
				Left:     left,
				Index:    idx,
			}
			continue
		case token.DOT:
			p.advance()
			memTok, err := p.expect(token.IDENT)
			if err != nil {
				return nil, err
			}
			left = &ast.MemberExpression{
				Position: left.Pos(),
				Object:   left,
				Member:   &ast.Identifier{Position: astPos(memTok.Pos), Name: memTok.Literal},
			}
			continue
		}

		// Binary infix.
		prec := infixPrecedence(p.current().Type)
		if prec == precNone || prec <= minPrec {
			break
		}
		opTok := p.advance()
		right, err := p.parsePratt(prec) // left-associative
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpression{
			Position: astPos(opTok.Pos),
			Left:     left,
			Op:       opTok.Literal,
			Right:    right,
		}
	}

	return left, nil
}

// parsePrefix handles prefix operators and primary expressions.
func (p *Parser) parsePrefix() (ast.Expression, error) {
	switch p.current().Type {
	case token.NOT, token.MINUS:
		opTok := p.advance()
		right, err := p.parsePratt(precMultiplic + 1)
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpression{
			Position: astPos(opTok.Pos),
			Op:       opTok.Literal,
			Right:    right,
		}, nil
	}
	return p.parsePrimary()
}

// ---------------------------------------------------------------------------
// Primary expressions
// ---------------------------------------------------------------------------

func (p *Parser) parsePrimary() (ast.Expression, error) {
	tok := p.current()

	switch tok.Type {
	case token.INT:
		p.advance()
		v, err := strconv.ParseInt(tok.Literal, 10, 64)
		if err != nil {
			return nil, p.errorf(tok.Pos, "invalid integer literal: %s", tok.Literal)
		}
		return &ast.IntegerLiteral{Position: astPos(tok.Pos), Value: v}, nil

	case token.FLOAT:
		p.advance()
		v, err := strconv.ParseFloat(tok.Literal, 64)
		if err != nil {
			return nil, p.errorf(tok.Pos, "invalid float literal: %s", tok.Literal)
		}
		return &ast.FloatLiteral{Position: astPos(tok.Pos), Value: v}, nil

	case token.STRING:
		p.advance()
		return &ast.StringLiteral{Position: astPos(tok.Pos), Value: tok.Literal}, nil

	case token.TRUE:
		p.advance()
		return &ast.BoolLiteral{Position: astPos(tok.Pos), Value: true}, nil

	case token.FALSE:
		p.advance()
		return &ast.BoolLiteral{Position: astPos(tok.Pos), Value: false}, nil

	case token.NIL:
		p.advance()
		return &ast.NilLiteral{Position: astPos(tok.Pos)}, nil

	case token.IDENT:
		p.advance()
		return &ast.Identifier{Position: astPos(tok.Pos), Name: tok.Literal}, nil

	case token.LBRACKET:
		return p.parseListLiteral()

	case token.LBRACE:
		return p.parseDictLiteral()

	case token.LPAREN:
		p.advance()
		p.skipNewlines()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		p.skipNewlines()
		if _, err := p.expect(token.RPAREN); err != nil {
			return nil, err
		}
		return expr, nil

	case token.IF:
		return p.parseIfExpression()

	default:
		return nil, p.errorf(tok.Pos, "unexpected token %s in expression", tok.Type)
	}
}

// --- List literal ----------------------------------------------------------

func (p *Parser) parseListLiteral() (*ast.ListLiteral, error) {
	pos := p.current().Pos
	p.advance() // consume '['
	p.skipNewlines()

	var elems []ast.Expression
	for p.current().Type != token.RBRACKET && p.current().Type != token.EOF {
		elem, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		elems = append(elems, elem)
		p.skipNewlines()
		if p.current().Type == token.COMMA {
			p.advance()
			p.skipNewlines()
		}
	}

	if _, err := p.expect(token.RBRACKET); err != nil {
		return nil, err
	}

	return &ast.ListLiteral{Position: astPos(pos), Elements: elems}, nil
}

// --- Dict literal ----------------------------------------------------------

func (p *Parser) parseDictLiteral() (*ast.DictLiteral, error) {
	pos := p.current().Pos
	p.advance() // consume '{'
	p.skipNewlines()

	var keys, vals []ast.Expression
	for p.current().Type != token.RBRACE && p.current().Type != token.EOF {
		key, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(token.COLON); err != nil {
			return nil, err
		}
		val, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
		vals = append(vals, val)
		p.skipNewlines()
		if p.current().Type == token.COMMA {
			p.advance()
			p.skipNewlines()
		}
	}

	if _, err := p.expect(token.RBRACE); err != nil {
		return nil, err
	}

	return &ast.DictLiteral{
		Position: astPos(pos),
		Keys:     keys,
		Values:   vals,
	}, nil
}

// --- If expression ---------------------------------------------------------

func (p *Parser) parseIfExpression() (*ast.IfExpression, error) {
	pos := p.current().Pos
	p.advance() // consume 'if'

	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(token.LBRACE); err != nil {
		return nil, err
	}
	p.skipNewlines()
	thenExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	p.skipNewlines()
	if _, err := p.expect(token.RBRACE); err != nil {
		return nil, err
	}

	var elseExpr ast.Expression
	p.skipNewlines()
	if p.current().Type == token.ELSE {
		p.advance()
		p.skipNewlines()
		if _, err := p.expect(token.LBRACE); err != nil {
			return nil, err
		}
		p.skipNewlines()
		elseExpr, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
		p.skipNewlines()
		if _, err := p.expect(token.RBRACE); err != nil {
			return nil, err
		}
	}

	return &ast.IfExpression{
		Position:  astPos(pos),
		Condition: cond,
		Then:      thenExpr,
		Else:      elseExpr,
	}, nil
}

// --- Call arguments --------------------------------------------------------

func (p *Parser) parseCallArgs() ([]ast.Expression, error) {
	p.skipNewlines()
	var args []ast.Expression
	for p.current().Type != token.RPAREN && p.current().Type != token.EOF {
		arg, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		p.skipNewlines()
		if p.current().Type == token.COMMA {
			p.advance()
			p.skipNewlines()
		}
	}
	if _, err := p.expect(token.RPAREN); err != nil {
		return nil, err
	}
	return args, nil
}
