// Package vm 实现 OpsLang 的字节码虚拟机
package vm

import (
	"fmt"

	"github.com/opslang/opslang/pkg/ast"
)

// Opcode 操作码
type Opcode byte

const (
	OP_NOP Opcode = iota
	OP_LOAD_CONST    // 加载常量
	OP_LOAD_VAR      // 加载变量
	OP_STORE_VAR     // 存储变量
	OP_POP           // 弹出栈顶
	OP_DUP           // 复制栈顶

	// 算术运算
	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV
	OP_MOD

	// 比较运算
	OP_EQ
	OP_NE
	OP_LT
	OP_LE
	OP_GT
	OP_GE

	// 逻辑运算
	OP_AND
	OP_OR
	OP_NOT

	// 控制流
	OP_JUMP          // 无条件跳转
	OP_JUMP_IF_FALSE // 条件跳转
	OP_CALL          // 函数调用
	OP_RETURN        // 返回

	// 数据结构
	OP_BUILD_ARRAY   // 构建数组
	OP_BUILD_MAP     // 构建字典
	OP_INDEX_GET     // 索引取值
	OP_INDEX_SET     // 索引赋值
	OP_MEMBER_GET    // 成员访问
	OP_MEMBER_SET    // 成员赋值

	// IO
	OP_PRINT         // 打印
	OP_INPUT         // 输入

	// 运维专用
	OP_SSH_RUN       // SSH 执行
	OP_SHELL_RUN     // Shell 执行
	OP_ENSURE        // 声明式保证
	OP_FLEET_EXEC    // 批量执行
)

// VM 虚拟机
type VM struct {
	stack    []Value
	globals  map[string]Value
	code     []Opcode
	consts   []Value
	ip       int // 指令指针
	sp       int // 栈指针
}

// Value 运行时值
type Value struct {
	Type  ValueType
	Int   int64
	Float float64
	Str   string
	Bool  bool
	Nil   bool
	Array []Value
	Map   map[string]Value
	Fn    *Function
}

// ValueType 值类型
type ValueType int

const (
	TYPE_NIL ValueType = iota
	TYPE_BOOL
	TYPE_INT
	TYPE_FLOAT
	TYPE_STRING
	TYPE_ARRAY
	TYPE_MAP
	TYPE_FUNCTION
)

// Function 运行时函数
type Function struct {
	Name      string
	Params    []string
	Body      *ast.FnStmt
	IsBuiltin bool
	Builtin   func(args []Value) (Value, error)
}

// New 创建新的虚拟机
func New() *VM {
	return &VM{
		stack:   make([]Value, 0, 256),
		globals: make(map[string]Value),
	}
}

// Run 执行程序
func (v *VM) Run(program *ast.Program) error {
	// TODO: 实现完整的字节码执行
	// 当前版本仅支持空程序

	// 注册内置函数
	v.registerBuiltins()

	return nil
}

// push 压栈
func (v *VM) push(val Value) {
	v.stack = append(v.stack, val)
	v.sp++
}

// pop 弹栈
func (v *VM) pop() Value {
	if v.sp == 0 {
		panic("VM stack underflow")
	}
	v.sp--
	val := v.stack[v.sp]
	v.stack = v.stack[:v.sp]
	return val
}

// peek 查看栈顶
func (v *VM) peek() Value {
	if v.sp == 0 {
		panic("VM stack underflow")
	}
	return v.stack[v.sp-1]
}

// registerBuiltins 注册内置函数
func (v *VM) registerBuiltins() {
	// print 函数
	v.globals["print"] = Value{
		Type: TYPE_FUNCTION,
		Fn: &Function{
			Name:      "print",
			IsBuiltin: true,
			Builtin: func(args []Value) (Value, error) {
				for i, arg := range args {
					if i > 0 {
						fmt.Print(" ")
					}
					fmt.Print(v.valueToString(arg))
				}
				fmt.Println()
				return Value{Type: TYPE_NIL, Nil: true}, nil
			},
		},
	}
}

// valueToString 值转字符串
func (v *VM) valueToString(val Value) string {
	switch val.Type {
	case TYPE_NIL:
		return "nil"
	case TYPE_BOOL:
		if val.Bool {
			return "true"
		}
		return "false"
	case TYPE_INT:
		return fmt.Sprintf("%d", val.Int)
	case TYPE_FLOAT:
		return fmt.Sprintf("%g", val.Float)
	case TYPE_STRING:
		return val.Str
	case TYPE_ARRAY:
		return fmt.Sprintf("%v", val.Array)
	case TYPE_MAP:
		return fmt.Sprintf("%v", val.Map)
	case TYPE_FUNCTION:
		return fmt.Sprintf("<fn:%s>", val.Fn.Name)
	default:
		return "<unknown>"
	}
}
