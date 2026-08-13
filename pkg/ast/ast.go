// Package ast 定义 OpsLang 的抽象语法树
package ast

// Node 所有 AST 节点的接口
type Node interface {
	nodeMarker()
	Pos() Position
}

// Position 源码位置
type Position struct {
	File   string
	Line   int
	Column int
}

// Program 程序根节点
type Program struct {
	Statements []Statement
	Position   Position
}

func (p *Program) nodeMarker() {}
func (p *Program) Pos() Position { return p.Position }

// Statement 语句接口
type Statement interface {
	Node
	stmtMarker()
}

// Expression 表达式接口
type Expression interface {
	Node
	exprMarker()
}

// --- 语句类型 ---

// ExprStmt 表达式语句
type ExprStmt struct {
	Expr     Expression
	Position Position
}

func (s *ExprStmt) nodeMarker() {}
func (s *ExprStmt) stmtMarker() {}
func (s *ExprStmt) Pos() Position { return s.Position }

// AssignStmt 赋值语句
type AssignStmt struct {
	Target   Expression
	Value    Expression
	Position Position
}

func (s *AssignStmt) nodeMarker() {}
func (s *AssignStmt) stmtMarker() {}
func (s *AssignStmt) Pos() Position { return s.Position }

// IfStmt if/else 语句
type IfStmt struct {
	Condition Expression
	Body      []Statement
	ElseIf    []*IfStmt  // else if 链
	Else      []Statement
	Position  Position
}

func (s *IfStmt) nodeMarker() {}
func (s *IfStmt) stmtMarker() {}
func (s *IfStmt) Pos() Position { return s.Position }

// ForStmt for 循环
type ForStmt struct {
	Variable string       // 循环变量
	Iterable Expression   // 可迭代对象
	Body     []Statement
	Position Position
}

func (s *ForStmt) nodeMarker() {}
func (s *ForStmt) stmtMarker() {}
func (s *ForStmt) Pos() Position { return s.Position }

// WhileStmt while 循环
type WhileStmt struct {
	Condition Expression
	Body      []Statement
	Position  Position
}

func (s *WhileStmt) nodeMarker() {}
func (s *WhileStmt) stmtMarker() {}
func (s *WhileStmt) Pos() Position { return s.Position }

// FnStmt 函数定义
type FnStmt struct {
	Name     string
	Params   []Parameter
	Body     []Statement
	Position Position
}

func (s *FnStmt) nodeMarker() {}
func (s *FnStmt) stmtMarker() {}
func (s *FnStmt) Pos() Position { return s.Position }

// Parameter 函数参数
type Parameter struct {
	Name    string
	Default Expression // 默认值（可选）
	Type    string     // 类型注解（可选）
}

// ReturnStmt return 语句
type ReturnStmt struct {
	Value    Expression // 返回值（可选）
	Position Position
}

func (s *ReturnStmt) nodeMarker() {}
func (s *ReturnStmt) stmtMarker() {}
func (s *ReturnStmt) Pos() Position { return s.Position }

// ImportStmt import 语句
type ImportStmt struct {
	Path     string
	Alias    string // as 别名
	Position Position
}

func (s *ImportStmt) nodeMarker() {}
func (s *ImportStmt) stmtMarker() {}
func (s *ImportStmt) Pos() Position { return s.Position }

// TryStmt try/catch 语句
type TryStmt struct {
	Body     []Statement
	CatchVar string      // catch 变量名
	Catch    []Statement
	Position Position
}

func (s *TryStmt) nodeMarker() {}
func (s *TryStmt) stmtMarker() {}
func (s *TryStmt) Pos() Position { return s.Position }

// EnsureStmt 声明式资源保证
type EnsureStmt struct {
	Resource string       // 资源类型 (file, service, package, user)
	Name     string       // 资源名称
	Props    map[string]Expression // 属性
	Position Position
}

func (s *EnsureStmt) nodeMarker() {}
func (s *EnsureStmt) stmtMarker() {}
func (s *EnsureStmt) Pos() Position { return s.Position }

// FleetStmt 批量执行语句
type FleetStmt struct {
	Action   string
	Hosts    Expression
	Body     []Statement
	Parallel int
	Position Position
}

func (s *FleetStmt) nodeMarker() {}
func (s *FleetStmt) stmtMarker() {}
func (s *FleetStmt) Pos() Position { return s.Position }

// --- 表达式类型 ---

// IdentExpr 标识符
type IdentExpr struct {
	Name     string
	Position Position
}

func (e *IdentExpr) nodeMarker() {}
func (e *IdentExpr) exprMarker() {}
func (e *IdentExpr) Pos() Position { return e.Position }

// IntLitExpr 整数字面量
type IntLitExpr struct {
	Value    int64
	Position Position
}

func (e *IntLitExpr) nodeMarker() {}
func (e *IntLitExpr) exprMarker() {}
func (e *IntLitExpr) Pos() Position { return e.Position }

// FloatLitExpr 浮点数字面量
type FloatLitExpr struct {
	Value    float64
	Position Position
}

func (e *FloatLitExpr) nodeMarker() {}
func (e *FloatLitExpr) exprMarker() {}
func (e *FloatLitExpr) Pos() Position { return e.Position }

// StringLitExpr 字符串字面量
type StringLitExpr struct {
	Value    string
	Position Position
}

func (e *StringLitExpr) nodeMarker() {}
func (e *StringLitExpr) exprMarker() {}
func (e *StringLitExpr) Pos() Position { return e.Position }

// BoolLitExpr 布尔字面量
type BoolLitExpr struct {
	Value    bool
	Position Position
}

func (e *BoolLitExpr) nodeMarker() {}
func (e *BoolLitExpr) exprMarker() {}
func (e *BoolLitExpr) Pos() Position { return e.Position }

// NilLitExpr nil 字面量
type NilLitExpr struct {
	Position Position
}

func (e *NilLitExpr) nodeMarker() {}
func (e *NilLitExpr) exprMarker() {}
func (e *NilLitExpr) Pos() Position { return e.Position }

// ArrayLitExpr 数组字面量
type ArrayLitExpr struct {
	Elements []Expression
	Position Position
}

func (e *ArrayLitExpr) nodeMarker() {}
func (e *ArrayLitExpr) exprMarker() {}
func (e *ArrayLitExpr) Pos() Position { return e.Position }

// MapLitExpr 字典字面量
type MapLitExpr struct {
	Keys     []Expression
	Values   []Expression
	Position Position
}

func (e *MapLitExpr) nodeMarker() {}
func (e *MapLitExpr) exprMarker() {}
func (e *MapLitExpr) Pos() Position { return e.Position }

// BinaryExpr 二元表达式
type BinaryExpr struct {
	Op       string
	Left     Expression
	Right    Expression
	Position Position
}

func (e *BinaryExpr) nodeMarker() {}
func (e *BinaryExpr) exprMarker() {}
func (e *BinaryExpr) Pos() Position { return e.Position }

// UnaryExpr 一元表达式
type UnaryExpr struct {
	Op       string
	Operand  Expression
	Position Position
}

func (e *UnaryExpr) nodeMarker() {}
func (e *UnaryExpr) exprMarker() {}
func (e *UnaryExpr) Pos() Position { return e.Position }

// CallExpr 函数调用
type CallExpr struct {
	Callee   Expression
	Args     []Expression
	Position Position
}

func (e *CallExpr) nodeMarker() {}
func (e *CallExpr) exprMarker() {}
func (e *CallExpr) Pos() Position { return e.Position }

// MemberExpr 成员访问 (a.b)
type MemberExpr struct {
	Object   Expression
	Member   string
	Position Position
}

func (e *MemberExpr) nodeMarker() {}
func (e *MemberExpr) exprMarker() {}
func (e *MemberExpr) Pos() Position { return e.Position }

// IndexExpr 索引访问 (a[b])
type IndexExpr struct {
	Object   Expression
	Index    Expression
	Position Position
}

func (e *IndexExpr) nodeMarker() {}
func (e *IndexExpr) exprMarker() {}
func (e *IndexExpr) Pos() Position { return e.Position }

// PipeExpr 管道表达式 (a |> b)
type PipeExpr struct {
	Left     Expression
	Right    Expression
	Position Position
}

func (e *PipeExpr) nodeMarker() {}
func (e *PipeExpr) exprMarker() {}
func (e *PipeExpr) Pos() Position { return e.Position }

// LambdaExpr Lambda 表达式 (fn(x) => x + 1)
type LambdaExpr struct {
	Params   []Parameter
	Body     Expression // 单表达式 lambda
	Position Position
}

func (e *LambdaExpr) nodeMarker() {}
func (e *LambdaExpr) exprMarker() {}
func (e *LambdaExpr) Pos() Position { return e.Position }

// InterpolatedStringExpr 插值字符串 "Hello {name}"
type InterpolatedStringExpr struct {
	Parts    []Expression // 字符串片段和表达式交替
	Position Position
}

func (e *InterpolatedStringExpr) nodeMarker() {}
func (e *InterpolatedStringExpr) exprMarker() {}
func (e *InterpolatedStringExpr) Pos() Position { return e.Position }

// BreakStmt break 语句
type BreakStmt struct {
	Position Position
}

func (s *BreakStmt) nodeMarker() {}
func (s *BreakStmt) stmtMarker() {}
func (s *BreakStmt) Pos() Position { return s.Position }

// ContinueStmt continue 语句
type ContinueStmt struct {
	Position Position
}

func (s *ContinueStmt) nodeMarker() {}
func (s *ContinueStmt) stmtMarker() {}
func (s *ContinueStmt) Pos() Position { return s.Position }

// ShExpr Shell 命令执行 sh("ls -la")
type ShExpr struct {
	Command  Expression
	Position Position
}

func (e *ShExpr) nodeMarker() {}
func (e *ShExpr) exprMarker() {}
func (e *ShExpr) Pos() Position { return e.Position }
