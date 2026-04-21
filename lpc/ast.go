// Package lpc 提供 LPC 语言子集的词法分析、语法解析和执行能力
package lpc

// NodeType 表示 AST 节点类型
type NodeType string

const (
	NodeProgram      NodeType = "Program"
	NodeFunctionDecl NodeType = "FunctionDecl"
	NodeVarDecl      NodeType = "VarDecl"
	NodeAssignStmt   NodeType = "AssignStmt"
	NodeExprStmt     NodeType = "ExprStmt"
	NodeIfStmt       NodeType = "IfStmt"
	NodeForStmt      NodeType = "ForStmt"
	NodeWhileStmt    NodeType = "WhileStmt"
	NodeReturnStmt   NodeType = "ReturnStmt"
	NodeBlockStmt    NodeType = "BlockStmt"

	NodeBinaryExpr  NodeType = "BinaryExpr"
	NodeUnaryExpr   NodeType = "UnaryExpr"
	NodeCallExpr    NodeType = "CallExpr"
	NodeMemberExpr  NodeType = "MemberExpr"
	NodeIndexExpr   NodeType = "IndexExpr"
	NodeIntLiteral  NodeType = "IntLiteral"
	NodeStringLiteral NodeType = "StringLiteral"
	NodeIdentExpr   NodeType = "IdentExpr"
)

// Node 是 AST 节点的接口
type Node interface {
	nodeType() NodeType
}

// --- 程序 ---

// Program 表示一个 LPC 程序
type ProgramAST struct {
	Globals   []Node // 全局变量声明
	Functions map[string]*FunctionDecl
}

func (p *ProgramAST) nodeType() NodeType { return NodeProgram }

// --- 声明 ---

// Param 表示函数参数
type Param struct {
	Type string
	Name string
}

// FunctionDecl 表示函数定义
type FunctionDecl struct {
	ReturnType string
	Name       string
	Params     []Param
	Body       *BlockStmt
}

func (f *FunctionDecl) nodeType() NodeType { return NodeFunctionDecl }

// VarDecl 表示变量声明
type VarDecl struct {
	Type  string
	Name  string
	Value Node // 可选的初始值
}

func (v *VarDecl) nodeType() NodeType { return NodeVarDecl }

// --- 语句 ---

// BlockStmt 表示代码块 { ... }
type BlockStmt struct {
	Statements []Node
}

func (b *BlockStmt) nodeType() NodeType { return NodeBlockStmt }

// AssignStmt 表示赋值语句
type AssignStmt struct {
	Name  string
	Value Node
}

func (a *AssignStmt) nodeType() NodeType { return NodeAssignStmt }

// ExprStmt 表示表达式语句
type ExprStmt struct {
	Expr Node
}

func (e *ExprStmt) nodeType() NodeType { return NodeExprStmt }

// IfStmt 表示 if/else 语句
type IfStmt struct {
	Cond Node
	Then *BlockStmt
	Else *BlockStmt
}

func (i *IfStmt) nodeType() NodeType { return NodeIfStmt }

// ForStmt 表示 for 循环
type ForStmt struct {
	Init Node // VarDecl 或 AssignStmt
	Cond Node // 表达式
	Post Node // AssignStmt 或 ExprStmt
	Body *BlockStmt
}

func (f *ForStmt) nodeType() NodeType { return NodeForStmt }

// WhileStmt 表示 while 循环
type WhileStmt struct {
	Cond Node
	Body *BlockStmt
}

func (w *WhileStmt) nodeType() NodeType { return NodeWhileStmt }

// ReturnStmt 表示 return 语句
type ReturnStmt struct {
	Value Node // 可选
}

func (r *ReturnStmt) nodeType() NodeType { return NodeReturnStmt }

// --- 表达式 ---

// BinaryExpr 表示二元表达式
type BinaryExpr struct {
	Op    string
	Left  Node
	Right Node
}

func (b *BinaryExpr) nodeType() NodeType { return NodeBinaryExpr }

// UnaryExpr 表示一元表达式
type UnaryExpr struct {
	Op   string
	Expr Node
}

func (u *UnaryExpr) nodeType() NodeType { return NodeUnaryExpr }

// CallExpr 表示函数调用
type CallExpr struct {
	Function Node // IdentExpr 或 MemberExpr
	Args     []Node
}

func (c *CallExpr) nodeType() NodeType { return NodeCallExpr }

// MemberExpr 表示属性访问：obj.field 或 obj->method()
type MemberExpr struct {
	Object Node
	Field  string
	Arrow  bool // true = ->, false = .
}

func (m *MemberExpr) nodeType() NodeType { return NodeMemberExpr }

// IndexExpr 表示数组下标访问
type IndexExpr struct {
	Object Node
	Index  Node
}

func (i *IndexExpr) nodeType() NodeType { return NodeIndexExpr }

// IntLiteral 表示整数字面量
type IntLiteral struct {
	Value int
}

func (i *IntLiteral) nodeType() NodeType { return NodeIntLiteral }

// StringLiteral 表示字符串字面量
type StringLiteral struct {
	Value string
}

func (s *StringLiteral) nodeType() NodeType { return NodeStringLiteral }

// IdentExpr 表示标识符
type IdentExpr struct {
	Name string
}

func (i *IdentExpr) nodeType() NodeType { return NodeIdentExpr }
