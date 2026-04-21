// Package lpc 提供 LPC 语言子集的解析和执行能力
// Phase 1: 使用硬编码的 Go 结构代替 LPC parser
// Phase 2: 将实现完整的 lexer + parser + interpreter
package lpc

// LPC 节点类型常量
const (
	NodeInt    = "int"
	NodeString = "string"
	NodeIdent  = "ident"
)

// Expr 表示 LPC 表达式节点（Phase 1 占位）
type Expr struct {
	Type  string
	Value interface{}
}

// Stmt 表示 LPC 语句节点（Phase 1 占位）
type Stmt struct {
	Type string
	Body interface{}
}

// Function 表示 LPC 函数定义（Phase 1 占位）
type Function struct {
	Name   string
	Params []string
	Body   []Stmt
}

// Program 表示一个 LPC 程序（Phase 1 占位）
type Program struct {
	Functions map[string]Function
	Globals   map[string]Expr
}

// NewProgram 创建一个新的空程序
func NewProgram() *Program {
	return &Program{
		Functions: make(map[string]Function),
		Globals:   make(map[string]Expr),
	}
}
