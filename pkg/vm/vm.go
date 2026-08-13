// Package vm 实现 OpsLang 的树遍历解释器
// MVP 阶段使用树遍历而非字节码，便于快速实现和调试
package vm

import (
	"fmt"
	"os"
	"strings"

	"github.com/opslang/opslang/pkg/ast"
)

// VM 虚拟机（树遍历解释器）
type VM struct {
	globals    map[string]Value
	builtins   map[string]BuiltinFunc
	output     *strings.Builder // 捕获输出（用于测试）
	callDepth  int
	maxDepth   int
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
	}
	vm.registerBuiltins()
	return vm
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
	switch iterable.Type {
	case TypeArray:
		for _, item := range iterable.Arr {
			env[varName] = item
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
	case TypeString:
		for _, ch := range iterable.Str {
			env[varName] = Value{Type: TypeString, Str: string(ch)}
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
	case TypeMap:
		for k := range iterable.Map {
			env[varName] = Value{Type: TypeString, Str: k}
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

	// 将内置函数注册到全局环境
	for name, fn := range v.builtins {
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
