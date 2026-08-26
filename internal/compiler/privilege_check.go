package compiler

import (
	"fmt"

	"github.com/j4ckzh0u/opslang/internal/ast"
	"github.com/j4ckzh0u/opslang/internal/security"
)

// CheckPrivileges is the AOT compile-time half of privilege enforcement:
// it walks the whole program and rejects statically-resolvable calls to
// mutating functions when the script's declared privilege (default
// read_only) does not allow them. Generate runs it before emitting any
// code; opsctl deploy also calls it directly so a violating script is
// refused on the controller before any host is contacted.
//
// Dynamic dispatch (a function value stored in a variable) is left to the
// runtime checks — builtin SDK calls are always statically named in the
// AST, so in practice every mutating call is visible here.
func CheckPrivileges(prog *ast.Program) error {
	scriptPriv := security.GetScriptPrivilege(prog)
	c := &privilegeChecker{scriptPriv: scriptPriv}
	for _, stmt := range prog.Statements {
		if err := c.checkStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

type privilegeChecker struct {
	scriptPriv ast.PrivilegeLevel
}

// resolveCallName builds the dotted name of a call's function expression
// (e.g. sys.cpu.usage) when it is a static identifier/member chain. Empty
// string for anything dynamic. CodeGenerator.resolveFuncName delegates here.
func resolveCallName(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name
	case *ast.MemberExpression:
		if prefix := resolveCallName(e.Object); prefix != "" {
			return prefix + "." + e.Member.Name
		}
	}
	return ""
}

// checkStatement returns the first privilege violation found in stmt.
func (c *privilegeChecker) checkStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case nil:
		return nil
	case *ast.LetStatement:
		return c.checkExpr(s.Value)
	case *ast.FnStatement:
		for _, p := range s.Params {
			if err := c.checkExpr(p.Default); err != nil {
				return err
			}
		}
		return c.checkBlock(s.Body)
	case *ast.IfStatement:
		if err := c.checkExpr(s.Condition); err != nil {
			return err
		}
		if err := c.checkBlock(s.Body); err != nil {
			return err
		}
		switch e := s.ElseClause.(type) {
		case *ast.BlockStatement:
			return c.checkBlock(e)
		case *ast.IfStatement:
			return c.checkStatement(e)
		}
		return nil
	case *ast.ForStatement:
		if err := c.checkStatement(s.Init); err != nil {
			return err
		}
		if err := c.checkExpr(s.Condition); err != nil {
			return err
		}
		if err := c.checkStatement(s.Post); err != nil {
			return err
		}
		return c.checkBlock(s.Body)
	case *ast.WhileStatement:
		if err := c.checkExpr(s.Condition); err != nil {
			return err
		}
		return c.checkBlock(s.Body)
	case *ast.ReturnStatement:
		return c.checkExpr(s.Value)
	case *ast.TaskStatement:
		return c.checkBlock(s.Body)
	case *ast.ExpressionStatement:
		return c.checkExpr(s.Expr)
	case *ast.AssignStatement:
		if err := c.checkExpr(s.Target); err != nil {
			return err
		}
		return c.checkExpr(s.Value)
	case *ast.ReportStatement:
		for _, f := range s.Fields {
			if err := c.checkExpr(f.Value); err != nil {
				return err
			}
		}
		return nil
	case *ast.AlertStatement:
		return c.checkExpr(s.Message)
	case *ast.EnsureStatement:
		if err := c.checkExpr(s.Condition); err != nil {
			return err
		}
		if err := c.checkBlock(s.Body); err != nil {
			return err
		}
		return c.checkExpr(s.Notify)
	case *ast.MetricStatement:
		if err := c.checkExpr(s.Name); err != nil {
			return err
		}
		if err := c.checkExpr(s.Value); err != nil {
			return err
		}
		return c.checkExpr(s.Labels)
	case *ast.LogStatement:
		return c.checkExpr(s.Message)
	case *ast.ParallelStatement:
		return c.checkBlock(s.Body)
	case *ast.BlockStatement:
		return c.checkBlock(s)
	case *ast.ImportStatement, *ast.PrivilegeStatement:
		return nil
	default:
		return nil
	}
}

func (c *privilegeChecker) checkBlock(block *ast.BlockStatement) error {
	if block == nil {
		return nil
	}
	for _, stmt := range block.Statements {
		if err := c.checkStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

// checkExpr returns the first privilege violation found in expr.
func (c *privilegeChecker) checkExpr(expr ast.Expression) error {
	switch e := expr.(type) {
	case nil:
		return nil
	case *ast.CallExpression:
		if name := resolveCallName(e.Function); name != "" {
			if err := security.CheckFuncPrivilege(c.scriptPriv, name); err != nil {
				pos := e.Pos()
				if pos.File != "" {
					return fmt.Errorf("%s:%d:%d: %w", pos.File, pos.Line, pos.Column, err)
				}
				return fmt.Errorf("%d:%d: %w", pos.Line, pos.Column, err)
			}
		}
		for _, arg := range e.Args {
			if err := c.checkExpr(arg); err != nil {
				return err
			}
		}
		return nil
	case *ast.ListLiteral:
		for _, elem := range e.Elements {
			if err := c.checkExpr(elem); err != nil {
				return err
			}
		}
		return nil
	case *ast.DictLiteral:
		for i := range e.Keys {
			if err := c.checkExpr(e.Keys[i]); err != nil {
				return err
			}
			if err := c.checkExpr(e.Values[i]); err != nil {
				return err
			}
		}
		return nil
	case *ast.BinaryExpression:
		if err := c.checkExpr(e.Left); err != nil {
			return err
		}
		return c.checkExpr(e.Right)
	case *ast.UnaryExpression:
		return c.checkExpr(e.Right)
	case *ast.IndexExpression:
		if err := c.checkExpr(e.Left); err != nil {
			return err
		}
		return c.checkExpr(e.Index)
	case *ast.MemberExpression:
		return c.checkExpr(e.Object)
	case *ast.IfExpression:
		if err := c.checkExpr(e.Condition); err != nil {
			return err
		}
		if err := c.checkExpr(e.Then); err != nil {
			return err
		}
		return c.checkExpr(e.Else)
	default:
		return nil
	}
}
