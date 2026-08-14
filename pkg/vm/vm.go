// Package vm 实现 OpsLang 的树遍历解释器
// MVP 阶段使用树遍历而非字节码，便于快速实现和调试
package vm

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"

	"github.com/opslang/opslang/pkg/ast"
	"github.com/opslang/opslang/stdlib/net/ssh"
)

// VM 虚拟机（树遍历解释器）
type VM struct {
	globals    map[string]Value
	builtins   map[string]BuiltinFunc
	output     *strings.Builder // 捕获输出（用于测试）
	callDepth  int
	maxDepth   int
	sshPool    *ssh.Pool // SSH 连接池
}

// Value 运行时值
type Value struct {
	Type   ValueType
	Int    int64
	Float  float64
	Str    string
	Bool   bool
	Arr    []Value
	Map    map[string]Value
	Fn     *FuncValue
	Nil    bool
}

// ValueType 值类型
type ValueType int

const (
	TypeNil ValueType = iota
	TypeBool
	TypeInt
	TypeFloat
	TypeString
	TypeArray
	TypeMap
	TypeFunction
)

// FuncValue 函数值
type FuncValue struct {
	Name      string
	Params    []ast.Parameter
	Body      []ast.Statement
	IsBuiltin bool
	Builtin   BuiltinFunc
	Closure   map[string]Value // 闭包变量
}

// BuiltinFunc 内置函数签名
type BuiltinFunc func(args []Value) (Value, error)

// Signal 控制流信号
type Signal int

const (
	SignalNone Signal = iota
	SignalReturn
	SignalBreak
	SignalContinue
)

// RuntimeError 运行时错误
type RuntimeError struct {
	Message string
	Line    int
}

func (e *RuntimeError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("RuntimeError at line %d: %s", e.Line, e.Message)
	}
	return fmt.Sprintf("RuntimeError: %s", e.Message)
}

// New 创建新的虚拟机
func New() *VM {
	vm := &VM{
		globals:  make(map[string]Value),
		builtins: make(map[string]BuiltinFunc),
		maxDepth: 1000,
		sshPool:  ssh.NewPool(),
	}
	vm.registerBuiltins()
	return vm
}

// Close releases resources held by the VM (e.g. SSH pool connections).
func (v *VM) Close() {
	if v.sshPool != nil {
		v.sshPool.Close()
	}
}

// Globals returns the VM's global variable map (for external module registration).
func (v *VM) Globals() map[string]Value {
	return v.globals
}

// SSHPool returns the VM's SSH connection pool (for external module registration).
func (v *VM) SSHPool() *ssh.Pool {
	return v.sshPool
}

// Run 执行程序
func (v *VM) Run(program *ast.Program) error {
	for _, stmt := range program.Statements {
		signal, _, err := v.execStmt(stmt, v.globals)
		if err != nil {
			return err
		}
		if signal == SignalReturn {
			return nil
		}
	}
	return nil
}

// --- 语句执行 ---

func (v *VM) execStmt(stmt ast.Statement, env map[string]Value) (Signal, Value, error) {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		val, err := v.evalExpr(s.Expr, env)
		return SignalNone, val, err

	case *ast.AssignStmt:
		val, err := v.evalExpr(s.Value, env)
		if err != nil {
			return SignalNone, Value{}, err
		}
		// 赋值目标
		switch target := s.Target.(type) {
		case *ast.IdentExpr:
			env[target.Name] = val
		case *ast.MemberExpr:
			obj, _ := v.evalExpr(target.Object, env)
			if obj.Type == TypeMap {
				obj.Map[target.Member] = val
			}
		case *ast.IndexExpr:
			obj, _ := v.evalExpr(target.Object, env)
			idx, _ := v.evalExpr(target.Index, env)
			if obj.Type == TypeArray && idx.Type == TypeInt {
				obj.Arr[idx.Int] = val
			} else if obj.Type == TypeMap && idx.Type == TypeString {
				obj.Map[idx.Str] = val
			}
		}
		return SignalNone, val, nil

	case *ast.FnStmt:
		fn := Value{
			Type: TypeFunction,
			Fn: &FuncValue{
				Name:   s.Name,
				Params: s.Params,
				Body:   s.Body,
			},
		}
		if s.Name != "" {
			env[s.Name] = fn
		}
		return SignalNone, fn, nil

	case *ast.IfStmt:
		cond, err := v.evalExpr(s.Condition, env)
		if err != nil {
			return SignalNone, Value{}, err
		}
		if v.isTruthy(cond) {
			return v.execBlock(s.Body, env)
		}
		// else if
		for _, elseIf := range s.ElseIf {
			cond2, err := v.evalExpr(elseIf.Condition, env)
			if err != nil {
				return SignalNone, Value{}, err
			}
			if v.isTruthy(cond2) {
				return v.execBlock(elseIf.Body, env)
			}
		}
		// else
		if len(s.Else) > 0 {
			return v.execBlock(s.Else, env)
		}
		return SignalNone, Value{Nil: true, Type: TypeNil}, nil

	case *ast.ForStmt:
		iterable, err := v.evalExpr(s.Iterable, env)
		if err != nil {
			return SignalNone, Value{}, err
		}
		return v.execForLoop(s.Variable, iterable, s.Body, env)

	case *ast.WhileStmt:
		return v.execWhile(s.Condition, s.Body, env)

	case *ast.ReturnStmt:
		if s.Value != nil {
			val, err := v.evalExpr(s.Value, env)
			if err != nil {
				return SignalNone, Value{}, err
			}
			return SignalReturn, val, nil
		}
		return SignalReturn, Value{Nil: true, Type: TypeNil}, nil

	case *ast.ImportStmt:
		// MVP: import 暂不实现
		return SignalNone, Value{Nil: true, Type: TypeNil}, nil

	case *ast.TryStmt:
		signal, val, err := v.execBlock(s.Body, env)
		if err != nil {
			if s.CatchVar != "" {
				catchEnv := copyEnv(env)
				catchEnv[s.CatchVar] = Value{Type: TypeString, Str: err.Error()}
				return v.execBlock(s.Catch, catchEnv)
			}
			return SignalNone, Value{}, nil // 吞掉错误
		}
		return signal, val, nil

	case *ast.BreakStmt:
		return SignalBreak, Value{}, nil

	case *ast.ContinueStmt:
		return SignalContinue, Value{}, nil

	default:
		return SignalNone, Value{}, &RuntimeError{Message: fmt.Sprintf("未知语句类型: %T", stmt)}
	}
}

func (v *VM) execBlock(stmts []ast.Statement, env map[string]Value) (Signal, Value, error) {
	for _, stmt := range stmts {
		signal, val, err := v.execStmt(stmt, env)
		if err != nil {
			return SignalNone, Value{}, err
		}
		if signal != SignalNone {
			return signal, val, nil
		}
	}
	return SignalNone, Value{Nil: true, Type: TypeNil}, nil
}

func (v *VM) execForLoop(varName string, iterable Value, body []ast.Statement, env map[string]Value) (Signal, Value, error) {
	// 保存循环变量的旧值（防止泄漏到外层作用域）
	oldVal, hadOld := env[varName]

	switch iterable.Type {
	case TypeArray:
		for _, item := range iterable.Arr {
			env[varName] = item
			signal, _, err := v.execBlock(body, env)
			if err != nil {
				if hadOld {
					env[varName] = oldVal
				} else {
					delete(env, varName)
				}
				return SignalNone, Value{}, err
			}
			if signal == SignalBreak {
				break
			}
			if signal == SignalReturn {
				if hadOld {
					env[varName] = oldVal
				} else {
					delete(env, varName)
				}
				return SignalReturn, Value{}, nil
			}
		}
	case TypeString:
		for _, ch := range iterable.Str {
			env[varName] = Value{Type: TypeString, Str: string(ch)}
			signal, _, err := v.execBlock(body, env)
			if err != nil {
				if hadOld {
					env[varName] = oldVal
				} else {
					delete(env, varName)
				}
				return SignalNone, Value{}, err
			}
			if signal == SignalBreak {
				break
			}
			if signal == SignalReturn {
				if hadOld {
					env[varName] = oldVal
				} else {
					delete(env, varName)
				}
				return SignalReturn, Value{}, nil
			}
		}
	case TypeMap:
		for k := range iterable.Map {
			env[varName] = Value{Type: TypeString, Str: k}
			signal, _, err := v.execBlock(body, env)
			if err != nil {
				if hadOld {
					env[varName] = oldVal
				} else {
					delete(env, varName)
				}
				return SignalNone, Value{}, err
			}
			if signal == SignalBreak {
				break
			}
			if signal == SignalReturn {
				if hadOld {
					env[varName] = oldVal
				} else {
					delete(env, varName)
				}
				return SignalReturn, Value{}, nil
			}
		}
	}

	// 恢复循环变量的旧值
	if hadOld {
		env[varName] = oldVal
	} else {
		delete(env, varName)
	}
	return SignalNone, Value{Nil: true, Type: TypeNil}, nil
}

func (v *VM) execWhile(cond ast.Expression, body []ast.Statement, env map[string]Value) (Signal, Value, error) {
	for {
		condVal, err := v.evalExpr(cond, env)
		if err != nil {
			return SignalNone, Value{}, err
		}
		if !v.isTruthy(condVal) {
			break
		}
		signal, _, err := v.execBlock(body, env)
		if err != nil {
			return SignalNone, Value{}, err
		}
		if signal == SignalBreak {
			break
		}
		if signal == SignalReturn {
			return SignalReturn, Value{}, nil
		}
	}
	return SignalNone, Value{Nil: true, Type: TypeNil}, nil
}

// --- 表达式求值 ---

func (v *VM) evalExpr(expr ast.Expression, env map[string]Value) (Value, error) {
	switch e := expr.(type) {
	case *ast.IntLitExpr:
		return Value{Type: TypeInt, Int: e.Value}, nil

	case *ast.FloatLitExpr:
		return Value{Type: TypeFloat, Float: e.Value}, nil

	case *ast.StringLitExpr:
		return Value{Type: TypeString, Str: e.Value}, nil

	case *ast.BoolLitExpr:
		return Value{Type: TypeBool, Bool: e.Value}, nil

	case *ast.NilLitExpr:
		return Value{Type: TypeNil, Nil: true}, nil

	case *ast.ArrayLitExpr:
		arr := make([]Value, len(e.Elements))
		for i, elem := range e.Elements {
			val, err := v.evalExpr(elem, env)
			if err != nil {
				return Value{}, err
			}
			arr[i] = val
		}
		return Value{Type: TypeArray, Arr: arr}, nil

	case *ast.MapLitExpr:
		m := make(map[string]Value)
		for i, key := range e.Keys {
			keyVal, err := v.evalExpr(key, env)
			if err != nil {
				return Value{}, err
			}
			valVal, err := v.evalExpr(e.Values[i], env)
			if err != nil {
				return Value{}, err
			}
			m[v.toString(keyVal)] = valVal
		}
		return Value{Type: TypeMap, Map: m}, nil

	case *ast.IdentExpr:
		// 先查局部环境，再查全局
		if val, ok := env[e.Name]; ok {
			return val, nil
		}
		if val, ok := v.globals[e.Name]; ok {
			return val, nil
		}
		return Value{}, &RuntimeError{
			Message: fmt.Sprintf("未定义变量: %s", e.Name),
			Line:    e.Position.Line,
		}

	case *ast.BinaryExpr:
		return v.evalBinary(e, env)

	case *ast.UnaryExpr:
		val, err := v.evalExpr(e.Operand, env)
		if err != nil {
			return Value{}, err
		}
		switch e.Op {
		case "-":
			if val.Type == TypeInt {
				return Value{Type: TypeInt, Int: -val.Int}, nil
			}
			if val.Type == TypeFloat {
				return Value{Type: TypeFloat, Float: -val.Float}, nil
			}
		case "!":
			return Value{Type: TypeBool, Bool: !v.isTruthy(val)}, nil
		}
		return Value{}, &RuntimeError{Message: fmt.Sprintf("不支持的一元运算: %s", e.Op)}

	case *ast.CallExpr:
		return v.evalCall(e, env)

	case *ast.MemberExpr:
		obj, err := v.evalExpr(e.Object, env)
		if err != nil {
			return Value{}, err
		}
		if obj.Type == TypeMap {
			if val, ok := obj.Map[e.Member]; ok {
				return val, nil
			}
			return Value{Type: TypeNil, Nil: true}, nil
		}
		// 字符串方法
		if obj.Type == TypeString {
			if fn, ok := v.getStringMethod(obj.Str, e.Member); ok {
				return fn, nil
			}
			return Value{}, &RuntimeError{Message: fmt.Sprintf("字符串没有方法: %s", e.Member)}
		}
		// 数组方法
		if obj.Type == TypeArray {
			if fn, ok := v.getArrayMethod(obj, e.Member); ok {
				return fn, nil
			}
			return Value{}, &RuntimeError{Message: fmt.Sprintf("数组没有方法: %s", e.Member)}
		}
		return Value{}, &RuntimeError{Message: fmt.Sprintf("不支持成员访问: %s", v.typeName(obj))}

	case *ast.IndexExpr:
		obj, err := v.evalExpr(e.Object, env)
		if err != nil {
			return Value{}, err
		}
		idx, err := v.evalExpr(e.Index, env)
		if err != nil {
			return Value{}, err
		}
		if obj.Type == TypeArray && idx.Type == TypeInt {
			if idx.Int < 0 || idx.Int >= int64(len(obj.Arr)) {
				return Value{}, &RuntimeError{Message: "数组索引越界"}
			}
			return obj.Arr[idx.Int], nil
		}
		if obj.Type == TypeMap {
			key := v.toString(idx)
			if val, ok := obj.Map[key]; ok {
				return val, nil
			}
			return Value{Type: TypeNil, Nil: true}, nil
		}
		if obj.Type == TypeString && idx.Type == TypeInt {
			if idx.Int < 0 || idx.Int >= int64(len(obj.Str)) {
				return Value{}, &RuntimeError{Message: "字符串索引越界"}
			}
			return Value{Type: TypeString, Str: string(obj.Str[idx.Int])}, nil
		}
		return Value{}, &RuntimeError{Message: "不支持索引操作"}

	case *ast.PipeExpr:
		left, err := v.evalExpr(e.Left, env)
		if err != nil {
			return Value{}, err
		}
		// 管道：把左边作为右边函数的第一个参数
		if callExpr, ok := e.Right.(*ast.CallExpr); ok {
			newArgs := append([]ast.Expression{&ast.IdentExpr{Name: "__pipe_val__"}}, callExpr.Args...)
			pipeEnv := copyEnv(env)
			pipeEnv["__pipe_val__"] = left
			newCall := &ast.CallExpr{Callee: callExpr.Callee, Args: newArgs, Position: callExpr.Position}
			return v.evalCall(newCall, pipeEnv)
		}
		return left, nil

	case *ast.InterpolatedStringExpr:
		var sb strings.Builder
		for _, part := range e.Parts {
			val, err := v.evalExpr(part, env)
			if err != nil {
				return Value{}, err
			}
			sb.WriteString(v.toString(val))
		}
		return Value{Type: TypeString, Str: sb.String()}, nil

	case *ast.LambdaExpr:
		fn := &FuncValue{
			Name:   "<lambda>",
			Params: e.Params,
			Body: []ast.Statement{
				&ast.ReturnStmt{Value: e.Body, Position: e.Position},
			},
			Closure: copyEnv(env),
		}
		return Value{Type: TypeFunction, Fn: fn}, nil

	default:
		return Value{}, &RuntimeError{Message: fmt.Sprintf("未知表达式类型: %T", expr)}
	}
}

func (v *VM) evalBinary(e *ast.BinaryExpr, env map[string]Value) (Value, error) {
	left, err := v.evalExpr(e.Left, env)
	if err != nil {
		return Value{}, err
	}
	right, err := v.evalExpr(e.Right, env)
	if err != nil {
		return Value{}, err
	}

	// 字符串拼接
	if e.Op == "+" && (left.Type == TypeString || right.Type == TypeString) {
		return Value{Type: TypeString, Str: v.toString(left) + v.toString(right)}, nil
	}

	switch e.Op {
	case "+":
		if left.Type == TypeInt && right.Type == TypeInt {
			return Value{Type: TypeInt, Int: left.Int + right.Int}, nil
		}
		if left.Type == TypeFloat || right.Type == TypeFloat {
			return Value{Type: TypeFloat, Float: v.toFloat(left) + v.toFloat(right)}, nil
		}
	case "-":
		if left.Type == TypeInt && right.Type == TypeInt {
			return Value{Type: TypeInt, Int: left.Int - right.Int}, nil
		}
		if left.Type == TypeFloat || right.Type == TypeFloat {
			return Value{Type: TypeFloat, Float: v.toFloat(left) - v.toFloat(right)}, nil
		}
	case "*":
		if left.Type == TypeInt && right.Type == TypeInt {
			return Value{Type: TypeInt, Int: left.Int * right.Int}, nil
		}
		if left.Type == TypeFloat || right.Type == TypeFloat {
			return Value{Type: TypeFloat, Float: v.toFloat(left) * v.toFloat(right)}, nil
		}
	case "/":
		if left.Type == TypeInt && right.Type == TypeInt {
			if right.Int == 0 {
				return Value{}, &RuntimeError{Message: "除以零"}
			}
			return Value{Type: TypeInt, Int: left.Int / right.Int}, nil
		}
		if v.toFloat(right) == 0 {
			return Value{}, &RuntimeError{Message: "除以零"}
		}
		return Value{Type: TypeFloat, Float: v.toFloat(left) / v.toFloat(right)}, nil
	case "%":
		if left.Type == TypeInt && right.Type == TypeInt {
			if right.Int == 0 {
				return Value{}, &RuntimeError{Message: "除以零"}
			}
			return Value{Type: TypeInt, Int: left.Int % right.Int}, nil
		}
	case "==":
		return Value{Type: TypeBool, Bool: v.equals(left, right)}, nil
	case "!=":
		return Value{Type: TypeBool, Bool: !v.equals(left, right)}, nil
	case "<":
		return Value{Type: TypeBool, Bool: v.toFloat(left) < v.toFloat(right)}, nil
	case "<=":
		return Value{Type: TypeBool, Bool: v.toFloat(left) <= v.toFloat(right)}, nil
	case ">":
		return Value{Type: TypeBool, Bool: v.toFloat(left) > v.toFloat(right)}, nil
	case ">=":
		return Value{Type: TypeBool, Bool: v.toFloat(left) >= v.toFloat(right)}, nil
	case "&&":
		return Value{Type: TypeBool, Bool: v.isTruthy(left) && v.isTruthy(right)}, nil
	case "||":
		return Value{Type: TypeBool, Bool: v.isTruthy(left) || v.isTruthy(right)}, nil
	}

	return Value{}, &RuntimeError{Message: fmt.Sprintf("不支持的运算: %s %s %s", v.typeName(left), e.Op, v.typeName(right))}
}

func (v *VM) evalCall(e *ast.CallExpr, env map[string]Value) (Value, error) {
	callee, err := v.evalExpr(e.Callee, env)
	if err != nil {
		return Value{}, err
	}

	// 求值参数
	args := make([]Value, len(e.Args))
	for i, arg := range e.Args {
		args[i], err = v.evalExpr(arg, env)
		if err != nil {
			return Value{}, err
		}
	}

	if callee.Type != TypeFunction {
		return Value{}, &RuntimeError{
			Message: fmt.Sprintf("不可调用: %s", v.typeName(callee)),
			Line:    e.Position.Line,
		}
	}

	fn := callee.Fn

	// 内置函数
	if fn.IsBuiltin {
		return fn.Builtin(args)
	}

	// 用户函数
	v.callDepth++
	if v.callDepth > v.maxDepth {
		return Value{}, &RuntimeError{Message: "调用栈溢出"}
	}
	defer func() { v.callDepth-- }()

	// 创建函数环境
	fnEnv := make(map[string]Value)
	if fn.Closure != nil {
		for k, val := range fn.Closure {
			fnEnv[k] = val
		}
	}

	// 绑定参数
	for i, param := range fn.Params {
		if i < len(args) {
			fnEnv[param.Name] = args[i]
		} else if param.Default != nil {
			defVal, err := v.evalExpr(param.Default, env)
			if err != nil {
				return Value{}, err
			}
			fnEnv[param.Name] = defVal
		} else {
			fnEnv[param.Name] = Value{Type: TypeNil, Nil: true}
		}
	}

	// 执行函数体
	for _, stmt := range fn.Body {
		signal, val, err := v.execStmt(stmt, fnEnv)
		if err != nil {
			return Value{}, err
		}
		if signal == SignalReturn {
			return val, nil
		}
	}

	return Value{Type: TypeNil, Nil: true}, nil
}

// --- 辅助方法 ---

func (v *VM) isTruthy(val Value) bool {
	switch val.Type {
	case TypeNil:
		return false
	case TypeBool:
		return val.Bool
	case TypeInt:
		return val.Int != 0
	case TypeFloat:
		return val.Float != 0
	case TypeString:
		return val.Str != ""
	case TypeArray:
		return len(val.Arr) > 0
	case TypeMap:
		return len(val.Map) > 0
	}
	return false
}

func (v *VM) equals(a, b Value) bool {
	if a.Type != b.Type {
		return false
	}
	switch a.Type {
	case TypeNil:
		return true
	case TypeBool:
		return a.Bool == b.Bool
	case TypeInt:
		return a.Int == b.Int
	case TypeFloat:
		return a.Float == b.Float
	case TypeString:
		return a.Str == b.Str
	}
	return false
}

func (v *VM) toFloat(val Value) float64 {
	switch val.Type {
	case TypeInt:
		return float64(val.Int)
	case TypeFloat:
		return val.Float
	case TypeString:
		// 尝试解析
		var f float64
		fmt.Sscanf(val.Str, "%f", &f)
		return f
	case TypeBool:
		if val.Bool {
			return 1
		}
		return 0
	}
	return 0
}

func (v *VM) toString(val Value) string {
	switch val.Type {
	case TypeNil:
		return "nil"
	case TypeBool:
		if val.Bool {
			return "true"
		}
		return "false"
	case TypeInt:
		return fmt.Sprintf("%d", val.Int)
	case TypeFloat:
		return fmt.Sprintf("%g", val.Float)
	case TypeString:
		return val.Str
	case TypeArray:
		parts := make([]string, len(val.Arr))
		for i, elem := range val.Arr {
			parts[i] = v.toString(elem)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case TypeMap:
		parts := make([]string, 0, len(val.Map))
		for k, val := range val.Map {
			parts = append(parts, fmt.Sprintf("%s: %s", k, v.toString(val)))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case TypeFunction:
		if val.Fn != nil {
			return fmt.Sprintf("<fn:%s>", val.Fn.Name)
		}
		return "<fn>"
	}
	return "<unknown>"
}

func (v *VM) typeName(val Value) string {
	switch val.Type {
	case TypeNil:
		return "nil"
	case TypeBool:
		return "bool"
	case TypeInt:
		return "int"
	case TypeFloat:
		return "float"
	case TypeString:
		return "string"
	case TypeArray:
		return "array"
	case TypeMap:
		return "map"
	case TypeFunction:
		return "function"
	}
	return "unknown"
}

func copyEnv(env map[string]Value) map[string]Value {
	newEnv := make(map[string]Value, len(env))
	for k, v := range env {
		newEnv[k] = v
	}
	return newEnv
}

// --- 内置函数 ---

func (v *VM) registerBuiltins() {
	// print
	v.builtins["print"] = func(args []Value) (Value, error) {
		parts := make([]string, len(args))
		for i, arg := range args {
			parts[i] = v.toString(arg)
		}
		line := strings.Join(parts, " ")
		if v.output != nil {
			v.output.WriteString(line + "\n")
		}
		fmt.Fprintln(os.Stdout, line)
		return Value{Type: TypeNil, Nil: true}, nil
	}

	// len
	v.builtins["len"] = func(args []Value) (Value, error) {
		if len(args) != 1 {
			return Value{}, &RuntimeError{Message: "len() 需要 1 个参数"}
		}
		switch args[0].Type {
		case TypeString:
			return Value{Type: TypeInt, Int: int64(len(args[0].Str))}, nil
		case TypeArray:
			return Value{Type: TypeInt, Int: int64(len(args[0].Arr))}, nil
		case TypeMap:
			return Value{Type: TypeInt, Int: int64(len(args[0].Map))}, nil
		}
		return Value{}, &RuntimeError{Message: fmt.Sprintf("len() 不支持类型: %s", v.typeName(args[0]))}
	}

	// type
	v.builtins["type"] = func(args []Value) (Value, error) {
		if len(args) != 1 {
			return Value{}, &RuntimeError{Message: "type() 需要 1 个参数"}
		}
		return Value{Type: TypeString, Str: v.typeName(args[0])}, nil
	}

	// str
	v.builtins["str"] = func(args []Value) (Value, error) {
		if len(args) != 1 {
			return Value{}, &RuntimeError{Message: "str() 需要 1 个参数"}
		}
		return Value{Type: TypeString, Str: v.toString(args[0])}, nil
	}

	// int
	v.builtins["int"] = func(args []Value) (Value, error) {
		if len(args) != 1 {
			return Value{}, &RuntimeError{Message: "int() 需要 1 个参数"}
		}
		return Value{Type: TypeInt, Int: int64(v.toFloat(args[0]))}, nil
	}

	// range
	v.builtins["range"] = func(args []Value) (Value, error) {
		var start, end, step int64
		switch len(args) {
		case 1:
			start, end, step = 0, args[0].Int, 1
		case 2:
			start, end, step = args[0].Int, args[1].Int, 1
		case 3:
			start, end, step = args[0].Int, args[1].Int, args[2].Int
		default:
			return Value{}, &RuntimeError{Message: "range() 需要 1-3 个参数"}
		}
		if step == 0 {
			return Value{}, &RuntimeError{Message: "range() step 不能为 0"}
		}
		arr := make([]Value, 0)
		for i := start; (step > 0 && i < end) || (step < 0 && i > end); i += step {
			arr = append(arr, Value{Type: TypeInt, Int: i})
		}
		return Value{Type: TypeArray, Arr: arr}, nil
	}

	// append
	v.builtins["append"] = func(args []Value) (Value, error) {
		if len(args) != 2 {
			return Value{}, &RuntimeError{Message: "append() 需要 2 个参数"}
		}
		if args[0].Type != TypeArray {
			return Value{}, &RuntimeError{Message: "append() 第一个参数必须是数组"}
		}
		arr := make([]Value, len(args[0].Arr))
		copy(arr, args[0].Arr)
		arr = append(arr, args[1])
		return Value{Type: TypeArray, Arr: arr}, nil
	}

	// map 函数
	v.builtins["map"] = func(args []Value) (Value, error) {
		if len(args) != 2 {
			return Value{}, &RuntimeError{Message: "map() 需要 2 个参数"}
		}
		if args[0].Type != TypeArray || args[1].Type != TypeFunction {
			return Value{}, &RuntimeError{Message: "map() 参数类型错误"}
		}
		fn := args[1].Fn
		result := make([]Value, len(args[0].Arr))
		for i, item := range args[0].Arr {
			fnEnv := make(map[string]Value)
			if len(fn.Params) > 0 {
				fnEnv[fn.Params[0].Name] = item
			}
			for _, stmt := range fn.Body {
				signal, val, err := v.execStmt(stmt, fnEnv)
				if err != nil {
					return Value{}, err
				}
				if signal == SignalReturn {
					result[i] = val
					break
				}
			}
		}
		return Value{Type: TypeArray, Arr: result}, nil
	}

	// filter 函数
	v.builtins["filter"] = func(args []Value) (Value, error) {
		if len(args) != 2 {
			return Value{}, &RuntimeError{Message: "filter() 需要 2 个参数"}
		}
		if args[0].Type != TypeArray || args[1].Type != TypeFunction {
			return Value{}, &RuntimeError{Message: "filter() 参数类型错误"}
		}
		fn := args[1].Fn
		result := make([]Value, 0)
		for _, item := range args[0].Arr {
			fnEnv := make(map[string]Value)
			if len(fn.Params) > 0 {
				fnEnv[fn.Params[0].Name] = item
			}
			for _, stmt := range fn.Body {
				signal, val, err := v.execStmt(stmt, fnEnv)
				if err != nil {
					return Value{}, err
				}
				if signal == SignalReturn {
					if v.isTruthy(val) {
						result = append(result, item)
					}
					break
				}
			}
		}
		return Value{Type: TypeArray, Arr: result}, nil
	}

	// input
	v.builtins["input"] = func(args []Value) (Value, error) {
		if len(args) > 0 {
			fmt.Print(v.toString(args[0]))
		}
		var line string
		fmt.Scanln(&line)
		return Value{Type: TypeString, Str: line}, nil
	}

	// === 标准库模块 ===

	// file 模块
	v.globals["file"] = Value{
		Type: TypeMap,
		Map: map[string]Value{
			"read": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "file.read",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "file.read(path) 需要 1 个字符串参数"}
					}
					data, err := os.ReadFile(args[0].Str)
					if err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("读取文件失败: %v", err)}
					}
					return Value{Type: TypeString, Str: string(data)}, nil
				},
			}},
			"write": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "file.write",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 2 || args[0].Type != TypeString || args[1].Type != TypeString {
						return Value{}, &RuntimeError{Message: "file.write(path, content) 需要 2 个字符串参数"}
					}
					err := os.WriteFile(args[0].Str, []byte(args[1].Str), 0644)
					if err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("写入文件失败: %v", err)}
					}
					return Value{Type: TypeNil, Nil: true}, nil
				},
			}},
			"exists": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "file.exists",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "file.exists(path) 需要 1 个字符串参数"}
					}
					_, err := os.Stat(args[0].Str)
					return Value{Type: TypeBool, Bool: err == nil}, nil
				},
			}},
			"mkdir": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "file.mkdir",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "file.mkdir(path) 需要 1 个字符串参数"}
					}
					err := os.MkdirAll(args[0].Str, 0755)
					if err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("创建目录失败: %v", err)}
					}
					return Value{Type: TypeNil, Nil: true}, nil
				},
			}},
			"list": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "file.list",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "file.list(dir) 需要 1 个字符串参数"}
					}
					entries, err := os.ReadDir(args[0].Str)
					if err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("列出目录失败: %v", err)}
					}
					arr := make([]Value, len(entries))
					for i, e := range entries {
						arr[i] = Value{Type: TypeString, Str: e.Name()}
					}
					return Value{Type: TypeArray, Arr: arr}, nil
				},
			}},
			"basename": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "file.basename",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "file.basename(path) 需要 1 个字符串参数"}
					}
					return Value{Type: TypeString, Str: filepath.Base(args[0].Str)}, nil
				},
			}},
			"dirname": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "file.dirname",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "file.dirname(path) 需要 1 个字符串参数"}
					}
					return Value{Type: TypeString, Str: filepath.Dir(args[0].Str)}, nil
				},
			}},
		},
	}

	// process 模块
	v.globals["process"] = Value{
		Type: TypeMap,
		Map: map[string]Value{
			"run": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "process.run",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "process.run(cmd, args...) 需要至少 1 个参数"}
					}
					cmdArgs := make([]string, len(args)-1)
					for i := 1; i < len(args); i++ {
						cmdArgs[i-1] = v.toString(args[i])
					}
					cmd := exec.Command(args[0].Str, cmdArgs...)
					output, err := cmd.CombinedOutput()
					exitCode := int64(0)
					if err != nil {
						if exitErr, ok := err.(*exec.ExitError); ok {
							exitCode = int64(exitErr.ExitCode())
						} else {
							return Value{}, &RuntimeError{Message: fmt.Sprintf("执行命令失败: %v", err)}
						}
					}
					return Value{Type: TypeMap, Map: map[string]Value{
						"stdout":   {Type: TypeString, Str: string(output)},
						"exitCode": {Type: TypeInt, Int: exitCode},
						"ok":       {Type: TypeBool, Bool: exitCode == 0},
					}}, nil
				},
			}},
			"shell": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "process.shell",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "process.shell(cmd) 需要 1 个字符串参数"}
					}
					cmd := exec.Command("sh", "-c", args[0].Str)
					output, err := cmd.CombinedOutput()
					exitCode := int64(0)
					if err != nil {
						if exitErr, ok := err.(*exec.ExitError); ok {
							exitCode = int64(exitErr.ExitCode())
						} else {
							return Value{}, &RuntimeError{Message: fmt.Sprintf("执行 Shell 失败: %v", err)}
						}
					}
					return Value{Type: TypeMap, Map: map[string]Value{
						"stdout":   {Type: TypeString, Str: string(output)},
						"exitCode": {Type: TypeInt, Int: exitCode},
						"ok":       {Type: TypeBool, Bool: exitCode == 0},
					}}, nil
				},
			}},
			"env": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "process.env",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "process.env(name) 需要 1 个字符串参数"}
					}
					val := os.Getenv(args[0].Str)
					return Value{Type: TypeString, Str: val}, nil
				},
			}},
			"cwd": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "process.cwd",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					dir, err := os.Getwd()
					if err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("获取当前目录失败: %v", err)}
					}
					return Value{Type: TypeString, Str: dir}, nil
				},
			}},
			"hostname": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "process.hostname",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					name, err := os.Hostname()
					if err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("获取主机名失败: %v", err)}
					}
					return Value{Type: TypeString, Str: name}, nil
				},
			}},
		},
	}

	// ssh 模块（使用 Go 原生 SSH 实现，支持密钥和密码认证，连接池复用）
	// sshExec 内部辅助：通过 SSH 连接池执行远程命令
	sshExec := func(host, command, user, password string) (string, int64, error) {
		cfg := ssh.Config{
			Host:     host,
			User:     user,
			Password: password,
		}
		client, err := v.sshPool.Get(cfg)
		if err != nil {
			return fmt.Sprintf("SSH 连接失败: %v", err), 255, err
		}
		output, exitCode, runErr := client.CombinedOutput(command)
		if runErr != nil || exitCode >= 128 {
			// Connection-level error or signal-killed exit (e.g. SIGPIPE=141):
			// evict from pool to avoid stale state on reused connections.
			// Normal non-zero exits (e.g. grep no-match=1) are NOT connection errors.
			v.sshPool.Remove(cfg)
		}
		return output, int64(exitCode), runErr
	}
	// sftpCopy 内部辅助：通过 SFTP 传输文件
	sftpCopy := func(local, host, remote, user, password string) error {
		cfg := ssh.Config{
			Host:     host,
			User:     user,
			Password: password,
		}
		client, err := v.sshPool.Get(cfg)
		if err != nil {
			return fmt.Errorf("SSH 连接失败: %w", err)
		}
		return client.Upload(local, remote)
	}
	// getStrArg 安全获取字符串参数
	getStrArg := func(args []Value, idx int, def string) string {
		if idx < len(args) && args[idx].Type == TypeString {
			return args[idx].Str
		}
		return def
	}

	v.globals["ssh"] = Value{
		Type: TypeMap,
		Map: map[string]Value{
			"run": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "ssh.run",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 2 || args[0].Type != TypeString || args[1].Type != TypeString {
						return Value{}, &RuntimeError{Message: "ssh.run(host, cmd, [user], [password]) 需要至少 2 个参数"}
					}
					host := args[0].Str
					command := args[1].Str
					user := getStrArg(args, 2, "root")
					password := getStrArg(args, 3, "")
					output, exitCode, connErr := sshExec(host, command, user, password)
					if connErr != nil && exitCode == 255 {
						return Value{Type: TypeMap, Map: map[string]Value{
							"stdout":   {Type: TypeString, Str: fmt.Sprintf("SSH 连接失败: %v", connErr)},
							"exitCode": {Type: TypeInt, Int: 255},
							"ok":       {Type: TypeBool, Bool: false},
							"host":     {Type: TypeString, Str: host},
						}}, nil
					}
					return Value{Type: TypeMap, Map: map[string]Value{
						"stdout":   {Type: TypeString, Str: output},
						"exitCode": {Type: TypeInt, Int: exitCode},
						"ok":       {Type: TypeBool, Bool: exitCode == 0},
						"host":     {Type: TypeString, Str: host},
					}}, nil
				},
			}},
			"copy": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "ssh.copy",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 3 {
						return Value{}, &RuntimeError{Message: "ssh.copy(local, host, remote, [user], [password]) 需要至少 3 个参数"}
					}
					local := args[0].Str
					host := args[1].Str
					remote := args[2].Str
					user := getStrArg(args, 3, "root")
					password := getStrArg(args, 4, "")
					if err := sftpCopy(local, host, remote, user, password); err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("SFTP 复制失败: %v", err)}
					}
					return Value{Type: TypeMap, Map: map[string]Value{
						"stdout":   {Type: TypeString, Str: ""},
						"exitCode": {Type: TypeInt, Int: 0},
						"ok":       {Type: TypeBool, Bool: true},
					}}, nil
				},
			}},
			"ping": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "ssh.ping",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "ssh.ping(host, [user], [password]) 需要至少 1 个参数"}
					}
					host := args[0].Str
					user := getStrArg(args, 1, "root")
					password := getStrArg(args, 2, "")
					_, exitCode, _ := sshExec(host, "echo ok", user, password)
					return Value{Type: TypeBool, Bool: exitCode == 0}, nil
				},
			}},
		},
	}

	// json 模块
	v.globals["json"] = Value{
		Type: TypeMap,
		Map: map[string]Value{
			"parse": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "json.parse",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "json.parse(str) 需要 1 个字符串参数"}
					}
					var raw interface{}
					if err := json.Unmarshal([]byte(args[0].Str), &raw); err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("JSON 解析失败: %v", err)}
					}
					return jsonToValue(raw), nil
				},
			}},
			"dump": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "json.dump",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 {
						return Value{}, &RuntimeError{Message: "json.dump(value) 需要 1 个参数"}
					}
					raw := valueToJson(args[0])
					data, err := json.Marshal(raw)
					if err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("JSON 序列化失败: %v", err)}
					}
					return Value{Type: TypeString, Str: string(data)}, nil
				},
			}},
			"prettify": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "json.prettify",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 {
						return Value{}, &RuntimeError{Message: "json.prettify(value) 需要 1 个参数"}
					}
					raw := valueToJson(args[0])
					data, err := json.MarshalIndent(raw, "", "  ")
					if err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("JSON 格式化失败: %v", err)}
					}
					return Value{Type: TypeString, Str: string(data)}, nil
				},
			}},
		},
	}

	// fleet 模块（批量执行引擎）
	v.globals["fleet"] = Value{
		Type: TypeMap,
		Map: map[string]Value{
			"parallel": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "fleet.parallel",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 2 || args[0].Type != TypeArray {
						return Value{}, &RuntimeError{Message: "fleet.parallel(hosts, fn, [max_concurrent]) 需要至少 2 个参数"}
					}
					hosts := args[0].Arr
					fn := args[1].Fn
					maxConcurrent := int64(10)
					if len(args) >= 3 && args[2].Type == TypeInt {
						maxConcurrent = args[2].Int
					}
					if maxConcurrent <= 0 {
						maxConcurrent = int64(len(hosts))
					}

					// 并发执行
					type result struct {
						host string
						val  Value
						err  error
					}
					results := make([]result, len(hosts))
					sem := make(chan struct{}, maxConcurrent)

					var wg sync.WaitGroup
					for i, h := range hosts {
						wg.Add(1)
						go func(idx int, host Value) {
							defer wg.Done()
							sem <- struct{}{}
							defer func() { <-sem }()

							hostStr := v.toString(host)
							fnEnv := make(map[string]Value)
							if len(fn.Params) > 0 {
								fnEnv[fn.Params[0].Name] = host
							}
							// 复制闭包变量
							if fn.Closure != nil {
								for k, val := range fn.Closure {
									fnEnv[k] = val
								}
							}

							var retVal Value
							var retErr error
							for _, stmt := range fn.Body {
								signal, val, err := v.execStmt(stmt, fnEnv)
								if err != nil {
									retErr = err
									break
								}
								if signal == SignalReturn {
									retVal = val
									break
								}
							}
							results[idx] = result{host: hostStr, val: retVal, err: retErr}
						}(i, h)
					}
					wg.Wait()

					// 收集结果
					arr := make([]Value, len(results))
					for i, r := range results {
						m := map[string]Value{
							"host": {Type: TypeString, Str: r.host},
						}
						if r.err != nil {
							m["error"] = Value{Type: TypeString, Str: r.err.Error()}
							m["ok"] = Value{Type: TypeBool, Bool: false}
						} else {
							m["result"] = r.val
							m["ok"] = Value{Type: TypeBool, Bool: true}
						}
						arr[i] = Value{Type: TypeMap, Map: m}
					}
					return Value{Type: TypeArray, Arr: arr}, nil
				},
			}},
			"serial": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "fleet.serial",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 2 || args[0].Type != TypeArray {
						return Value{}, &RuntimeError{Message: "fleet.serial(hosts, fn) 需要至少 2 个参数"}
					}
					hosts := args[0].Arr
					fn := args[1].Fn

					arr := make([]Value, len(hosts))
					for i, h := range hosts {
						hostStr := v.toString(h)
						fnEnv := make(map[string]Value)
						if len(fn.Params) > 0 {
							fnEnv[fn.Params[0].Name] = h
						}
						if fn.Closure != nil {
							for k, val := range fn.Closure {
								fnEnv[k] = val
							}
						}

						var retVal Value
						var retErr error
						for _, stmt := range fn.Body {
							signal, val, err := v.execStmt(stmt, fnEnv)
							if err != nil {
								retErr = err
								break
							}
							if signal == SignalReturn {
								retVal = val
								break
							}
						}

						m := map[string]Value{
							"host": {Type: TypeString, Str: hostStr},
						}
						if retErr != nil {
							m["error"] = Value{Type: TypeString, Str: retErr.Error()}
							m["ok"] = Value{Type: TypeBool, Bool: false}
						} else {
							m["result"] = retVal
							m["ok"] = Value{Type: TypeBool, Bool: true}
						}
						arr[i] = Value{Type: TypeMap, Map: m}
					}
					return Value{Type: TypeArray, Arr: arr}, nil
				},
			}},
			"exec": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "fleet.exec",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 2 || args[0].Type != TypeArray || args[1].Type != TypeString {
						return Value{}, &RuntimeError{Message: "fleet.exec(hosts, cmd, [user], [password]) 需要至少 2 个参数"}
					}
					hosts := args[0].Arr
					command := args[1].Str
					user := getStrArg(args, 2, "root")
					password := getStrArg(args, 3, "")

					// 使用 SSH 并行执行
					type result struct {
						host string
						out  string
						code int64
					}
					results := make([]result, len(hosts))
					var wg sync.WaitGroup

					for i, h := range hosts {
						wg.Add(1)
						go func(idx int, host Value) {
							defer wg.Done()
							hostStr := v.toString(host)
							output, exitCode, _ := sshExec(hostStr, command, user, password)
							results[idx] = result{host: hostStr, out: output, code: exitCode}
						}(i, h)
					}
					wg.Wait()

					arr := make([]Value, len(results))
					for i, r := range results {
						arr[i] = Value{Type: TypeMap, Map: map[string]Value{
							"host":     {Type: TypeString, Str: r.host},
							"stdout":   {Type: TypeString, Str: r.out},
							"exitCode": {Type: TypeInt, Int: r.code},
							"ok":       {Type: TypeBool, Bool: r.code == 0},
						}}
					}
					return Value{Type: TypeArray, Arr: arr}, nil
				},
			}},
			"summary": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "fleet.summary",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 || args[0].Type != TypeArray {
						return Value{}, &RuntimeError{Message: "fleet.summary(results) 需要 1 个数组参数"}
					}
					total := int64(len(args[0].Arr))
					ok := int64(0)
					fail := int64(0)
					for _, item := range args[0].Arr {
						if item.Type == TypeMap {
							if okVal, exists := item.Map["ok"]; exists && okVal.Bool {
								ok++
							} else {
								fail++
							}
						}
					}
					return Value{Type: TypeMap, Map: map[string]Value{
						"total": {Type: TypeInt, Int: total},
						"ok":    {Type: TypeInt, Int: ok},
						"fail":  {Type: TypeInt, Int: fail},
					}}, nil
				},
			}},
		},
	}

	// str 模块（字符串工具）
	v.globals["strings"] = Value{
		Type: TypeMap,
		Map: map[string]Value{
			"split": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "str.split",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 2 || args[0].Type != TypeString || args[1].Type != TypeString {
						return Value{}, &RuntimeError{Message: "str.split(s, sep) 需要 2 个字符串参数"}
					}
					parts := strings.Split(args[0].Str, args[1].Str)
					arr := make([]Value, len(parts))
					for i, p := range parts {
						arr[i] = Value{Type: TypeString, Str: p}
					}
					return Value{Type: TypeArray, Arr: arr}, nil
				},
			}},
			"join": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "str.join",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 2 || args[0].Type != TypeArray || args[1].Type != TypeString {
						return Value{}, &RuntimeError{Message: "str.join(arr, sep) 需要数组和分隔符"}
					}
					parts := make([]string, len(args[0].Arr))
					for i, item := range args[0].Arr {
						parts[i] = v.toString(item)
					}
					return Value{Type: TypeString, Str: strings.Join(parts, args[1].Str)}, nil
				},
			}},
			"contains": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "str.contains",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 2 || args[0].Type != TypeString || args[1].Type != TypeString {
						return Value{}, &RuntimeError{Message: "str.contains(s, sub) 需要 2 个字符串参数"}
					}
					return Value{Type: TypeBool, Bool: strings.Contains(args[0].Str, args[1].Str)}, nil
				},
			}},
			"replace": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "str.replace",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 3 || args[0].Type != TypeString || args[1].Type != TypeString || args[2].Type != TypeString {
						return Value{}, &RuntimeError{Message: "str.replace(s, old, new) 需要 3 个字符串参数"}
					}
					return Value{Type: TypeString, Str: strings.ReplaceAll(args[0].Str, args[1].Str, args[2].Str)}, nil
				},
			}},
			"trim": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "str.trim",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "str.trim(s) 需要 1 个字符串参数"}
					}
					return Value{Type: TypeString, Str: strings.TrimSpace(args[0].Str)}, nil
				},
			}},
			"upper": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "str.upper",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "str.upper(s) 需要 1 个字符串参数"}
					}
					return Value{Type: TypeString, Str: strings.ToUpper(args[0].Str)}, nil
				},
			}},
			"lower": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "str.lower",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "str.lower(s) 需要 1 个字符串参数"}
					}
					return Value{Type: TypeString, Str: strings.ToLower(args[0].Str)}, nil
				},
			}},
			"has_prefix": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "str.has_prefix",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 2 || args[0].Type != TypeString || args[1].Type != TypeString {
						return Value{}, &RuntimeError{Message: "str.has_prefix(s, prefix) 需要 2 个字符串参数"}
					}
					return Value{Type: TypeBool, Bool: strings.HasPrefix(args[0].Str, args[1].Str)}, nil
				},
			}},
			"has_suffix": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "str.has_suffix",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 2 || args[0].Type != TypeString || args[1].Type != TypeString {
						return Value{}, &RuntimeError{Message: "str.has_suffix(s, suffix) 需要 2 个字符串参数"}
					}
					return Value{Type: TypeBool, Bool: strings.HasSuffix(args[0].Str, args[1].Str)}, nil
				},
			}},
		},
	}

	// math 模块
	v.globals["math"] = Value{
		Type: TypeMap,
		Map: map[string]Value{
			"abs": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "math.abs",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 {
						return Value{}, &RuntimeError{Message: "math.abs(n) 需要 1 个参数"}
					}
					if args[0].Type == TypeInt {
						n := args[0].Int
						if n < 0 {
							n = -n
						}
						return Value{Type: TypeInt, Int: n}, nil
					}
					return Value{}, &RuntimeError{Message: "math.abs 需要数字参数"}
				},
			}},
			"min": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "math.min",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 2 {
						return Value{}, &RuntimeError{Message: "math.min(a, b) 需要 2 个参数"}
					}
					a, b := v.toFloat(args[0]), v.toFloat(args[1])
					if a < b {
						return args[0], nil
					}
					return args[1], nil
				},
			}},
			"max": {Type: TypeFunction, Fn: &FuncValue{
				Name:      "math.max",
				IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 2 {
						return Value{}, &RuntimeError{Message: "math.max(a, b) 需要 2 个参数"}
					}
					a, b := v.toFloat(args[0]), v.toFloat(args[1])
					if a > b {
						return args[0], nil
					}
					return args[1], nil
				},
			}},
		},
	}




	// inventory 模块（主机清单管理）
	v.globals["inventory"] = Value{
		Type: TypeMap,
		Map: map[string]Value{
			"load": {Type: TypeFunction, Fn: &FuncValue{
				Name: "inventory.load", IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "inventory.load(path) 需要文件路径"}
					}
					data, err := os.ReadFile(args[0].Str)
					if err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("读取清单文件失败: %v", err)}
					}
					return parseInventory(string(data)), nil
				},
			}},
			"from_list": {Type: TypeFunction, Fn: &FuncValue{
				Name: "inventory.from_list", IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 1 || args[0].Type != TypeArray {
						return Value{}, &RuntimeError{Message: "inventory.from_list(hosts) 需要数组"}
					}
					hosts := make([]string, len(args[0].Arr))
					for i, h := range args[0].Arr {
						hosts[i] = v.toString(h)
					}
					groups := map[string]Value{
						"all": {Type: TypeArray, Arr: args[0].Arr},
					}
					return Value{Type: TypeMap, Map: map[string]Value{
						"hosts":  {Type: TypeArray, Arr: args[0].Arr},
						"groups": {Type: TypeMap, Map: groups},
					}}, nil
				},
			}},
			"group": {Type: TypeFunction, Fn: &FuncValue{
				Name: "inventory.group", IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 2 || args[0].Type != TypeMap || args[1].Type != TypeString {
						return Value{}, &RuntimeError{Message: "inventory.group(inv, name) 需要清单和组名"}
					}
					inv := args[0]
					groupName := args[1].Str
					if groups, ok := inv.Map["groups"]; ok {
						if g, ok := groups.Map[groupName]; ok {
							return g, nil
						}
					}
					return Value{Type: TypeArray, Arr: []Value{}}, nil
				},
			}},
			"all": {Type: TypeFunction, Fn: &FuncValue{
				Name: "inventory.all", IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 1 || args[0].Type != TypeMap {
						return Value{}, &RuntimeError{Message: "inventory.all(inv) 需要清单"}
					}
					if hosts, ok := args[0].Map["hosts"]; ok {
						return hosts, nil
					}
					return Value{Type: TypeArray, Arr: []Value{}}, nil
				},
			}},
		},
	}
	// ensure 模块（声明式资源管理）
	v.globals["ensure"] = Value{
		Type: TypeMap,
		Map: map[string]Value{
			"file": {Type: TypeFunction, Fn: &FuncValue{
				Name: "ensure.file", IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "ensure.file(path, [content], [mode]) 需要路径参数"}
					}
					path := args[0].Str
					changed := false

					// 检查 content
					if len(args) >= 2 && args[1].Type == TypeString {
						content := args[1].Str
						existing, err := os.ReadFile(path)
						if err != nil || string(existing) != content {
							// 获取目录
							dir := filepath.Dir(path)
							os.MkdirAll(dir, 0755)
							mode := os.FileMode(0644)
							if len(args) >= 3 && args[2].Type == TypeInt {
								mode = os.FileMode(args[2].Int)
							}
							if err := os.WriteFile(path, []byte(content), mode); err != nil {
								return Value{}, &RuntimeError{Message: fmt.Sprintf("写入文件失败: %v", err)}
							}
							changed = true
						}
					} else {
						// 只确保文件存在
						if _, err := os.Stat(path); os.IsNotExist(err) {
							dir := filepath.Dir(path)
							os.MkdirAll(dir, 0755)
							if err := os.WriteFile(path, []byte{}, 0644); err != nil {
								return Value{}, &RuntimeError{Message: fmt.Sprintf("创建文件失败: %v", err)}
							}
							changed = true
						}
					}
					return Value{Type: TypeMap, Map: map[string]Value{
						"path":    {Type: TypeString, Str: path},
						"changed": {Type: TypeBool, Bool: changed},
						"ok":      {Type: TypeBool, Bool: true},
					}}, nil
				},
			}},
			"dir": {Type: TypeFunction, Fn: &FuncValue{
				Name: "ensure.dir", IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "ensure.dir(path, [mode]) 需要路径参数"}
					}
					path := args[0].Str
					mode := os.FileMode(0755)
					if len(args) >= 2 && args[1].Type == TypeInt {
						mode = os.FileMode(args[1].Int)
					}
					changed := false
					if _, err := os.Stat(path); os.IsNotExist(err) {
						if err := os.MkdirAll(path, mode); err != nil {
							return Value{}, &RuntimeError{Message: fmt.Sprintf("创建目录失败: %v", err)}
						}
						changed = true
					}
					return Value{Type: TypeMap, Map: map[string]Value{
						"path":    {Type: TypeString, Str: path},
						"changed": {Type: TypeBool, Bool: changed},
						"ok":      {Type: TypeBool, Bool: true},
					}}, nil
				},
			}},
			"line": {Type: TypeFunction, Fn: &FuncValue{
				Name: "ensure.line", IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 2 || args[0].Type != TypeString || args[1].Type != TypeString {
						return Value{}, &RuntimeError{Message: "ensure.line(path, line) 需要路径和内容参数"}
					}
					path := args[0].Str
					line := args[1].Str
					changed := false

					content := ""
					if data, err := os.ReadFile(path); err == nil {
						content = string(data)
					}

					if !strings.Contains(content, line) {
						if !strings.HasSuffix(content, "\n") && content != "" {
							content += "\n"
						}
						content += line + "\n"
						if err := os.WriteFile(path, []byte(content), 0644); err != nil {
							return Value{}, &RuntimeError{Message: fmt.Sprintf("写入文件失败: %v", err)}
						}
						changed = true
					}

					return Value{Type: TypeMap, Map: map[string]Value{
						"path":    {Type: TypeString, Str: path},
						"line":    {Type: TypeString, Str: line},
						"changed": {Type: TypeBool, Bool: changed},
						"ok":      {Type: TypeBool, Bool: true},
					}}, nil
				},
			}},
			"service": {Type: TypeFunction, Fn: &FuncValue{
				Name: "ensure.service", IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "ensure.service(name, [state], [enabled]) 需要服务名"}
					}
					name := args[0].Str
					state := "running"
					enabled := true
					if len(args) >= 2 && args[1].Type == TypeString {
						state = args[1].Str
					}
					if len(args) >= 3 && args[2].Type == TypeBool {
						enabled = args[2].Bool
					}

					changed := false
					// 检查服务状态
					checkCmd := exec.Command("systemctl", "is-active", name)
					checkOut, _ := checkCmd.Output()
					isActive := strings.TrimSpace(string(checkOut)) == "active"

					// 启动/停止
					if state == "running" && !isActive {
						cmd := exec.Command("systemctl", "start", name)
						if err := cmd.Run(); err != nil {
							return Value{Type: TypeMap, Map: map[string]Value{
								"name":    {Type: TypeString, Str: name},
								"changed": {Type: TypeBool, Bool: false},
								"ok":      {Type: TypeBool, Bool: false},
								"error":   {Type: TypeString, Str: fmt.Sprintf("启动服务失败: %v", err)},
							}}, nil
						}
						changed = true
					} else if state == "stopped" && isActive {
						cmd := exec.Command("systemctl", "stop", name)
						cmd.Run()
						changed = true
					}

					// 启用/禁用
					enCmd := exec.Command("systemctl", "is-enabled", name)
					enOut, _ := enCmd.Output()
					isEnabled := strings.TrimSpace(string(enOut)) == "enabled"

					if enabled && !isEnabled {
						cmd := exec.Command("systemctl", "enable", name)
						cmd.Run()
						changed = true
					} else if !enabled && isEnabled {
						cmd := exec.Command("systemctl", "disable", name)
						cmd.Run()
						changed = true
					}

					return Value{Type: TypeMap, Map: map[string]Value{
						"name":    {Type: TypeString, Str: name},
						"state":   {Type: TypeString, Str: state},
						"changed": {Type: TypeBool, Bool: changed},
						"ok":      {Type: TypeBool, Bool: true},
					}}, nil
				},
			}},
			"package": {Type: TypeFunction, Fn: &FuncValue{
				Name: "ensure.package", IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "ensure.package(name, [state]) 需要包名"}
					}
					name := args[0].Str
					state := "present"
					if len(args) >= 2 && args[1].Type == TypeString {
						state = args[1].Str
					}

					// 检查包是否已安装
					checkCmd := exec.Command("sh", "-c", fmt.Sprintf("which %s 2>/dev/null || dpkg -l %s 2>/dev/null || rpm -q %s 2>/dev/null", name, name, name))
					checkErr := checkCmd.Run()
					isInstalled := checkErr == nil

					changed := false
					if state == "present" && !isInstalled {
						// 尝试安装
						installCmd := exec.Command("sh", "-c",
							fmt.Sprintf("apt-get install -y %s 2>/dev/null || yum install -y %s 2>/dev/null || brew install %s 2>/dev/null", name, name, name))
						if err := installCmd.Run(); err != nil {
							return Value{Type: TypeMap, Map: map[string]Value{
								"name":    {Type: TypeString, Str: name},
								"changed": {Type: TypeBool, Bool: false},
								"ok":      {Type: TypeBool, Bool: false},
								"error":   {Type: TypeString, Str: fmt.Sprintf("安装失败: %v", err)},
							}}, nil
						}
						changed = true
					} else if state == "absent" && isInstalled {
						removeCmd := exec.Command("sh", "-c",
							fmt.Sprintf("apt-get remove -y %s 2>/dev/null || yum remove -y %s 2>/dev/null", name, name))
						removeCmd.Run()
						changed = true
					}

					return Value{Type: TypeMap, Map: map[string]Value{
						"name":    {Type: TypeString, Str: name},
						"state":   {Type: TypeString, Str: state},
						"changed": {Type: TypeBool, Bool: changed},
						"ok":      {Type: TypeBool, Bool: true},
					}}, nil
				},
			}},
			"user": {Type: TypeFunction, Fn: &FuncValue{
				Name: "ensure.user", IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) < 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "ensure.user(name, [shell], [groups]) 需要用户名"}
					}
					name := args[0].Str
					shell := "/bin/bash"
					if len(args) >= 2 && args[1].Type == TypeString {
						shell = args[1].Str
					}

					// 检查用户是否存在
					checkCmd := exec.Command("id", name)
					isExist := checkCmd.Run() == nil

					changed := false
					if !isExist {
						cmdArgs := []string{"-m", "-s", shell}
						if len(args) >= 3 && args[2].Type == TypeString {
							cmdArgs = append(cmdArgs, "-G", args[2].Str)
						}
						cmdArgs = append(cmdArgs, name)
						cmd := exec.Command("useradd", cmdArgs...)
						if err := cmd.Run(); err != nil {
							return Value{Type: TypeMap, Map: map[string]Value{
								"name":    {Type: TypeString, Str: name},
								"changed": {Type: TypeBool, Bool: false},
								"ok":      {Type: TypeBool, Bool: false},
								"error":   {Type: TypeString, Str: fmt.Sprintf("创建用户失败: %v", err)},
							}}, nil
						}
						changed = true
					}

					return Value{Type: TypeMap, Map: map[string]Value{
						"name":    {Type: TypeString, Str: name},
						"changed": {Type: TypeBool, Bool: changed},
						"ok":      {Type: TypeBool, Bool: true},
					}}, nil
				},
			}},
		},
	}
	// yaml 模块
	v.globals["yaml"] = Value{
		Type: TypeMap,
		Map: map[string]Value{
			"parse": {Type: TypeFunction, Fn: &FuncValue{
				Name: "yaml.parse", IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "yaml.parse(str) 需要 1 个字符串参数"}
					}
					var raw interface{}
					if err := yaml.Unmarshal([]byte(args[0].Str), &raw); err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("YAML 解析失败: %v", err)}
					}
					return yamlToValue(raw), nil
				},
			}},
			"dump": {Type: TypeFunction, Fn: &FuncValue{
				Name: "yaml.dump", IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 {
						return Value{}, &RuntimeError{Message: "yaml.dump(value) 需要 1 个参数"}
					}
					raw := valueToYaml(args[0])
					data, err := yaml.Marshal(raw)
					if err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("YAML 序列化失败: %v", err)}
					}
					return Value{Type: TypeString, Str: string(data)}, nil
				},
			}},
			"load_file": {Type: TypeFunction, Fn: &FuncValue{
				Name: "yaml.load_file", IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "yaml.load_file(path) 需要 1 个字符串参数"}
					}
					data, err := os.ReadFile(args[0].Str)
					if err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("读取文件失败: %v", err)}
					}
					var raw interface{}
					if err := yaml.Unmarshal(data, &raw); err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("YAML 解析失败: %v", err)}
					}
					return yamlToValue(raw), nil
				},
			}},
			"save_file": {Type: TypeFunction, Fn: &FuncValue{
				Name: "yaml.save_file", IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 2 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "yaml.save_file(path, value) 需要 2 个参数"}
					}
					raw := valueToYaml(args[1])
					data, err := yaml.Marshal(raw)
					if err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("YAML 序列化失败: %v", err)}
					}
					if err := os.WriteFile(args[0].Str, data, 0644); err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("写入文件失败: %v", err)}
					}
					return Value{Type: TypeNil, Nil: true}, nil
				},
			}},
		},
	}

	// toml 模块
	v.globals["toml"] = Value{
		Type: TypeMap,
		Map: map[string]Value{
			"parse": {Type: TypeFunction, Fn: &FuncValue{
				Name: "toml.parse", IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "toml.parse(str) 需要 1 个字符串参数"}
					}
					var raw map[string]interface{}
					if err := toml.Unmarshal([]byte(args[0].Str), &raw); err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("TOML 解析失败: %v", err)}
					}
					return tomlToValue(raw), nil
				},
			}},
			"load_file": {Type: TypeFunction, Fn: &FuncValue{
				Name: "toml.load_file", IsBuiltin: true,
				Builtin: func(args []Value) (Value, error) {
					if len(args) != 1 || args[0].Type != TypeString {
						return Value{}, &RuntimeError{Message: "toml.load_file(path) 需要 1 个字符串参数"}
					}
					var raw map[string]interface{}
					if _, err := toml.DecodeFile(args[0].Str, &raw); err != nil {
						return Value{}, &RuntimeError{Message: fmt.Sprintf("TOML 文件解析失败: %v", err)}
					}
					return tomlToValue(raw), nil
				},
			}},
		},
	}
	// 将内置函数注册到全局环境（不覆盖已有模块）
	for name, fn := range v.builtins {
		if _, exists := v.globals[name]; !exists {
			v.globals[name] = Value{
				Type: TypeFunction,
				Fn: &FuncValue{
					Name:      name,
					IsBuiltin: true,
					Builtin:   fn,
				},
			}
		}
	}
}

// === JSON 转换辅助 ===

// jsonToValue 将 Go 的 JSON 解析结果转换为 OpsLang Value
func jsonToValue(raw interface{}) Value {
	switch v := raw.(type) {
	case nil:
		return Value{Type: TypeNil, Nil: true}
	case bool:
		return Value{Type: TypeBool, Bool: v}
	case float64:
		// JSON 数字都是 float64，尝试转为 int
		if v == float64(int64(v)) {
			return Value{Type: TypeInt, Int: int64(v)}
		}
		return Value{Type: TypeFloat, Float: v}
	case string:
		return Value{Type: TypeString, Str: v}
	case []interface{}:
		arr := make([]Value, len(v))
		for i, item := range v {
			arr[i] = jsonToValue(item)
		}
		return Value{Type: TypeArray, Arr: arr}
	case map[string]interface{}:
		m := make(map[string]Value)
		for k, val := range v {
			m[k] = jsonToValue(val)
		}
		return Value{Type: TypeMap, Map: m}
	default:
		return Value{Type: TypeNil, Nil: true}
	}
}

// valueToJson 将 OpsLang Value 转换为可 JSON 序列化的 Go 值
func valueToJson(val Value) interface{} {
	switch val.Type {
	case TypeNil:
		return nil
	case TypeBool:
		return val.Bool
	case TypeInt:
		return val.Int
	case TypeFloat:
		return val.Float
	case TypeString:
		return val.Str
	case TypeArray:
		arr := make([]interface{}, len(val.Arr))
		for i, item := range val.Arr {
			arr[i] = valueToJson(item)
		}
		return arr
	case TypeMap:
		m := make(map[string]interface{})
		for k, v := range val.Map {
			m[k] = valueToJson(v)
		}
		return m
	default:
		return nil
	}
}

// === YAML/TOML 转换辅助 ===

func yamlToValue(raw interface{}) Value {
	switch v := raw.(type) {
	case nil:
		return Value{Type: TypeNil, Nil: true}
	case bool:
		return Value{Type: TypeBool, Bool: v}
	case int:
		return Value{Type: TypeInt, Int: int64(v)}
	case int64:
		return Value{Type: TypeInt, Int: v}
	case float64:
		if v == float64(int64(v)) {
			return Value{Type: TypeInt, Int: int64(v)}
		}
		return Value{Type: TypeFloat, Float: v}
	case string:
		return Value{Type: TypeString, Str: v}
	case []interface{}:
		arr := make([]Value, len(v))
		for i, item := range v {
			arr[i] = yamlToValue(item)
		}
		return Value{Type: TypeArray, Arr: arr}
	case map[string]interface{}:
		m := make(map[string]Value)
		for k, val := range v {
			m[k] = yamlToValue(val)
		}
		return Value{Type: TypeMap, Map: m}
	case map[interface{}]interface{}:
		m := make(map[string]Value)
		for k, val := range v {
			m[fmt.Sprintf("%v", k)] = yamlToValue(val)
		}
		return Value{Type: TypeMap, Map: m}
	default:
		return Value{Type: TypeString, Str: fmt.Sprintf("%v", v)}
	}
}

func valueToYaml(val Value) interface{} {
	switch val.Type {
	case TypeNil:
		return nil
	case TypeBool:
		return val.Bool
	case TypeInt:
		return val.Int
	case TypeFloat:
		return val.Float
	case TypeString:
		return val.Str
	case TypeArray:
		arr := make([]interface{}, len(val.Arr))
		for i, item := range val.Arr {
			arr[i] = valueToYaml(item)
		}
		return arr
	case TypeMap:
		m := make(map[string]interface{})
		for k, v := range val.Map {
			m[k] = valueToYaml(v)
		}
		return m
	default:
		return nil
	}
}

func tomlToValue(raw map[string]interface{}) Value {
	m := make(map[string]Value)
	for k, v := range raw {
		m[k] = yamlToValue(v) // reuse yamlToValue since TOML types are similar
	}
	return Value{Type: TypeMap, Map: m}
}

// === 字符串方法 ===

func (v *VM) getStringMethod(s string, method string) (Value, bool) {
	switch method {
	case "upper":
		return Value{Type: TypeFunction, Fn: &FuncValue{
			Name: "upper", IsBuiltin: true,
			Builtin: func(args []Value) (Value, error) {
				return Value{Type: TypeString, Str: strings.ToUpper(s)}, nil
			},
		}}, true
	case "lower":
		return Value{Type: TypeFunction, Fn: &FuncValue{
			Name: "lower", IsBuiltin: true,
			Builtin: func(args []Value) (Value, error) {
				return Value{Type: TypeString, Str: strings.ToLower(s)}, nil
			},
		}}, true
	case "trim":
		return Value{Type: TypeFunction, Fn: &FuncValue{
			Name: "trim", IsBuiltin: true,
			Builtin: func(args []Value) (Value, error) {
				return Value{Type: TypeString, Str: strings.TrimSpace(s)}, nil
			},
		}}, true
	case "split":
		return Value{Type: TypeFunction, Fn: &FuncValue{
			Name: "split", IsBuiltin: true,
			Builtin: func(args []Value) (Value, error) {
				sep := " "
				if len(args) > 0 && args[0].Type == TypeString {
					sep = args[0].Str
				}
				parts := strings.Split(s, sep)
				arr := make([]Value, len(parts))
				for i, p := range parts {
					arr[i] = Value{Type: TypeString, Str: p}
				}
				return Value{Type: TypeArray, Arr: arr}, nil
			},
		}}, true
	case "contains":
		return Value{Type: TypeFunction, Fn: &FuncValue{
			Name: "contains", IsBuiltin: true,
			Builtin: func(args []Value) (Value, error) {
				if len(args) != 1 || args[0].Type != TypeString {
					return Value{}, &RuntimeError{Message: "contains() 需要 1 个字符串参数"}
				}
				return Value{Type: TypeBool, Bool: strings.Contains(s, args[0].Str)}, nil
			},
		}}, true
	case "replace":
		return Value{Type: TypeFunction, Fn: &FuncValue{
			Name: "replace", IsBuiltin: true,
			Builtin: func(args []Value) (Value, error) {
				if len(args) != 2 || args[0].Type != TypeString || args[1].Type != TypeString {
					return Value{}, &RuntimeError{Message: "replace(old, new) 需要 2 个字符串参数"}
				}
				return Value{Type: TypeString, Str: strings.ReplaceAll(s, args[0].Str, args[1].Str)}, nil
			},
		}}, true
	case "starts_with":
		return Value{Type: TypeFunction, Fn: &FuncValue{
			Name: "starts_with", IsBuiltin: true,
			Builtin: func(args []Value) (Value, error) {
				if len(args) != 1 || args[0].Type != TypeString {
					return Value{}, &RuntimeError{Message: "starts_with() 需要 1 个字符串参数"}
				}
				return Value{Type: TypeBool, Bool: strings.HasPrefix(s, args[0].Str)}, nil
			},
		}}, true
	case "ends_with":
		return Value{Type: TypeFunction, Fn: &FuncValue{
			Name: "ends_with", IsBuiltin: true,
			Builtin: func(args []Value) (Value, error) {
				if len(args) != 1 || args[0].Type != TypeString {
					return Value{}, &RuntimeError{Message: "ends_with() 需要 1 个字符串参数"}
				}
				return Value{Type: TypeBool, Bool: strings.HasSuffix(s, args[0].Str)}, nil
			},
		}}, true
	case "len":
		return Value{Type: TypeFunction, Fn: &FuncValue{
			Name: "len", IsBuiltin: true,
			Builtin: func(args []Value) (Value, error) {
				return Value{Type: TypeInt, Int: int64(len(s))}, nil
			},
		}}, true
	case "to_int":
		return Value{Type: TypeFunction, Fn: &FuncValue{
			Name: "to_int", IsBuiltin: true,
			Builtin: func(args []Value) (Value, error) {
				var n int64
				fmt.Sscanf(s, "%d", &n)
				return Value{Type: TypeInt, Int: n}, nil
			},
		}}, true
	}
	return Value{}, false
}

// === 数组方法 ===

func (v *VM) getArrayMethod(arr Value, method string) (Value, bool) {
	switch method {
	case "len":
		return Value{Type: TypeFunction, Fn: &FuncValue{
			Name: "len", IsBuiltin: true,
			Builtin: func(args []Value) (Value, error) {
				return Value{Type: TypeInt, Int: int64(len(arr.Arr))}, nil
			},
		}}, true
	case "append":
		return Value{Type: TypeFunction, Fn: &FuncValue{
			Name: "append", IsBuiltin: true,
			Builtin: func(args []Value) (Value, error) {
				newArr := make([]Value, len(arr.Arr))
				copy(newArr, arr.Arr)
				if len(args) > 0 {
					newArr = append(newArr, args[0])
				}
				return Value{Type: TypeArray, Arr: newArr}, nil
			},
		}}, true
	case "join":
		return Value{Type: TypeFunction, Fn: &FuncValue{
			Name: "join", IsBuiltin: true,
			Builtin: func(args []Value) (Value, error) {
				sep := ""
				if len(args) > 0 && args[0].Type == TypeString {
					sep = args[0].Str
				}
				parts := make([]string, len(arr.Arr))
				for i, item := range arr.Arr {
					parts[i] = v.toString(item)
				}
				return Value{Type: TypeString, Str: strings.Join(parts, sep)}, nil
			},
		}}, true
	case "reverse":
		return Value{Type: TypeFunction, Fn: &FuncValue{
			Name: "reverse", IsBuiltin: true,
			Builtin: func(args []Value) (Value, error) {
				n := len(arr.Arr)
				newArr := make([]Value, n)
				for i := 0; i < n; i++ {
					newArr[i] = arr.Arr[n-1-i]
				}
				return Value{Type: TypeArray, Arr: newArr}, nil
			},
		}}, true
	}
	return Value{}, false
}

// === Inventory 解析 ===

// parseInventory 解析 INI 格式的主机清单文件
// 格式：
// [group_name]
// host1 ansible_host=10.0.0.1
// host2 ansible_host=10.0.0.2
func parseInventory(content string) Value {
	groups := make(map[string]Value)
	allHosts := make([]Value, 0)
	currentGroup := "all"
	groupHosts := make(map[string][]Value)
	groupHosts["all"] = make([]Value, 0)

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// 组头 [group_name]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentGroup = strings.TrimSpace(line[1 : len(line)-1])
			if _, ok := groupHosts[currentGroup]; !ok {
				groupHosts[currentGroup] = make([]Value, 0)
			}
			continue
		}

		// 主机行: hostname [key=value ...]
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		hostname := parts[0]
		hostVars := make(map[string]Value)
		hostVars["name"] = Value{Type: TypeString, Str: hostname}

		// 解析变量
		for _, p := range parts[1:] {
			if idx := strings.Index(p, "="); idx > 0 {
				key := p[:idx]
				val := p[idx+1:]
				hostVars[key] = Value{Type: TypeString, Str: val}
			}
		}

		hostVal := Value{Type: TypeMap, Map: hostVars}
		allHosts = append(allHosts, hostVal)
		groupHosts["all"] = append(groupHosts["all"], hostVal)
		if currentGroup != "all" {
			groupHosts[currentGroup] = append(groupHosts[currentGroup], hostVal)
		}
	}

	// 构建 groups map
	for groupName, hosts := range groupHosts {
		groups[groupName] = Value{Type: TypeArray, Arr: hosts}
	}

	return Value{Type: TypeMap, Map: map[string]Value{
		"hosts":  {Type: TypeArray, Arr: allHosts},
		"groups": {Type: TypeMap, Map: groups},
	}}
}
