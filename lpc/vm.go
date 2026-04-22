package lpc

import (
	"context"
	"fmt"
)

// Efun 是 LPC 内置函数的 Go 实现
type Efun func(args []Value) Value

// Environment 表示执行环境，包含变量作用域和函数表
type Environment struct {
	vars      map[string]Value
	parent    *Environment
	functions map[string]*FunctionDecl
	efuns     map[string]Efun
}

// NewEnvironment 创建新的执行环境
func NewEnvironment() *Environment {
	return &Environment{
		vars:      make(map[string]Value),
		functions: make(map[string]*FunctionDecl),
		efuns:     make(map[string]Efun),
	}
}

// PushScope 创建子作用域
func (e *Environment) PushScope() *Environment {
	return &Environment{
		vars:      make(map[string]Value),
		parent:    e,
		functions: e.functions,
		efuns:     e.efuns,
	}
}

// Get 获取变量值（沿作用域链查找）
func (e *Environment) Get(name string) (Value, bool) {
	if v, ok := e.vars[name]; ok {
		return v, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return Null(), false
}

// Set 设置变量值
func (e *Environment) Set(name string, val Value) {
	e.vars[name] = val
}

// SetGlobal 设置变量值（沿作用域链查找并更新）
func (e *Environment) SetGlobal(name string, val Value) {
	if _, ok := e.vars[name]; ok {
		e.vars[name] = val
		return
	}
	if e.parent != nil {
		e.parent.SetGlobal(name, val)
		return
	}
	// 变量不存在，设置在当前作用域
	e.vars[name] = val
}

// Value 表示 LPC 运行时值
type Value struct {
	Kind   string // "int", "string", "array", "null"
	IntVal int
	StrVal string
	ArrVal []Value
}

// IntValue 创建整数值
func IntValue(v int) Value {
	return Value{Kind: "int", IntVal: v}
}

// ArrayValue 创建数组值
func ArrayValue(v []Value) Value {
	return Value{Kind: "array", ArrVal: v}
}

// StringValue 创建字符串值
func StringValue(v string) Value {
	return Value{Kind: "string", StrVal: v}
}

// Null 创建空值
func Null() Value {
	return Value{Kind: "null"}
}

// IsTrue 判断值是否为真
func (v Value) IsTrue() bool {
	switch v.Kind {
	case "int":
		return v.IntVal != 0
	case "string":
		return v.StrVal != ""
	default:
		return false
	}
}

// String 返回值的字符串表示
func (v Value) String() string {
	switch v.Kind {
	case "int":
		return fmt.Sprintf("%d", v.IntVal)
	case "string":
		return v.StrVal
	case "array":
		return fmt.Sprintf("array(%d)", len(v.ArrVal))
	default:
		return "null"
	}
}

// ReturnValue 表示 return 语句产生的值（用于控制流）
type ReturnValue struct {
	Value Value
}

func (r *ReturnValue) Error() string {
	return "return"
}

const (
	DefaultMaxInstructions = 100000 // 默认最大指令数
	DefaultMaxCallDepth    = 100    // 默认最大调用栈深度
)

// ErrInstructionLimit 指令数超限错误
var ErrInstructionLimit = fmt.Errorf("LPC execution aborted: instruction limit exceeded")

// ErrCallDepthLimit 调用栈深度超限错误
var ErrCallDepthLimit = fmt.Errorf("LPC execution aborted: call depth limit exceeded")

// ErrContextCanceled 执行超时或被取消
var ErrContextCanceled = fmt.Errorf("LPC execution aborted: context canceled or timeout")

// VM 是 LPC 虚拟机，负责执行 AST
type VM struct {
	Env              *Environment
	Out              []string        // write() 输出缓冲
	ObjManager       *ObjectManager  // 对象管理器引用（efun 使用）
	Ctx              context.Context // 执行上下文（超时控制）
	InstructionCount int             // 已执行指令数
	MaxInstructions  int             // 最大指令数限制（0 = 使用默认值）
	CallDepth        int             // 当前调用栈深度
	MaxCallDepth     int             // 最大调用栈深度（0 = 使用默认值）
	abortErr         error           // evalExpr 中产生的致命错误
}

// NewVM 创建新的虚拟机
func NewVM() *VM {
	return &VM{
		Env: NewEnvironment(),
		Out: []string{},
		Ctx: context.Background(),
	}
}

// getMaxInstructions 返回有效的最大指令数
func (vm *VM) getMaxInstructions() int {
	if vm.MaxInstructions > 0 {
		return vm.MaxInstructions
	}
	return DefaultMaxInstructions
}

// getMaxCallDepth 返回有效的最大调用深度
func (vm *VM) getMaxCallDepth() int {
	if vm.MaxCallDepth > 0 {
		return vm.MaxCallDepth
	}
	return DefaultMaxCallDepth
}

// checkLimits 检查执行限制（指令数、超时）
func (vm *VM) checkLimits() error {
	vm.InstructionCount++
	if vm.InstructionCount > vm.getMaxInstructions() {
		return ErrInstructionLimit
	}
	if vm.Ctx != nil {
		select {
		case <-vm.Ctx.Done():
			return ErrContextCanceled
		default:
		}
	}
	return nil
}

// ResetCounters 重置执行计数器（每次 CallFunc 前调用）
func (vm *VM) ResetCounters() {
	vm.InstructionCount = 0
	vm.CallDepth = 0
	vm.abortErr = nil
}

// RegisterEfun 注册内置函数
func (vm *VM) RegisterEfun(name string, fn Efun) {
	vm.Env.efuns[name] = fn
}

// LoadProgram 加载程序到虚拟机
func (vm *VM) LoadProgram(prog *ProgramAST) {
	for name, fn := range prog.Functions {
		vm.Env.functions[name] = fn
	}
	for _, g := range prog.Globals {
		if vd, ok := g.(*VarDecl); ok {
			if vd.Value != nil {
				val := vm.evalExpr(vd.Value)
				vm.Env.vars[vd.Name] = val
			} else {
				vm.Env.vars[vd.Name] = Null()
			}
		}
	}
}

// CallFunc 调用已加载的函数
func (vm *VM) CallFunc(name string, args []Value) (Value, error) {
	vm.ResetCounters()
	return vm.callFunc(name, args)
}

// CallFuncWithContext 带超时上下文调用函数
func (vm *VM) CallFuncWithContext(ctx context.Context, name string, args []Value) (Value, error) {
	vm.Ctx = ctx
	defer func() { vm.Ctx = context.Background() }()
	return vm.CallFunc(name, args)
}

// callFunc 内部函数调用（不重置计数器）
func (vm *VM) callFunc(name string, args []Value) (Value, error) {
	// 检查 efun
	if efun, ok := vm.Env.efuns[name]; ok {
		return efun(args), nil
	}
	// 检查用户函数
	fn, ok := vm.Env.functions[name]
	if !ok {
		return Null(), fmt.Errorf("function %q not found", name)
	}
	return vm.callFunction(fn, args)
}

// callFunction 执行用户定义的函数
func (vm *VM) callFunction(fn *FunctionDecl, args []Value) (Value, error) {
	vm.CallDepth++
	if vm.CallDepth > vm.getMaxCallDepth() {
		return Null(), ErrCallDepthLimit
	}
	defer func() { vm.CallDepth-- }()

	scope := vm.Env.PushScope()
	// 绑定参数
	for i, param := range fn.Params {
		var val Value
		if i < len(args) {
			val = args[i]
		}
		scope.vars[param.Name] = val
	}

	vm.Env = scope
	defer func() { vm.Env = scope.parent }()

	err := vm.execBlock(fn.Body)
	if rv, ok := err.(*ReturnValue); ok {
		return rv.Value, nil
	}
	if err != nil {
		return Null(), err
	}
	return Null(), nil
}

// execBlock 执行代码块
func (vm *VM) execBlock(block *BlockStmt) error {
	for _, stmt := range block.Statements {
		err := vm.execStmt(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

// execStmt 执行单条语句
func (vm *VM) execStmt(stmt Node) error {
	if vm.abortErr != nil {
		return vm.abortErr
	}
	if err := vm.checkLimits(); err != nil {
		return err
	}
	switch s := stmt.(type) {
	case *VarDecl:
		if s.Value != nil {
			val := vm.evalExpr(s.Value)
			vm.Env.Set(s.Name, val)
		} else {
			vm.Env.Set(s.Name, Null())
		}

	case *AssignStmt:
		val := vm.evalExpr(s.Value)
		vm.Env.SetGlobal(s.Name, val)

	case *ExprStmt:
		vm.evalExpr(s.Expr)
		if vm.abortErr != nil {
			return vm.abortErr
		}

	case *IfStmt:
		cond := vm.evalExpr(s.Cond)
		if cond.IsTrue() {
			return vm.execBlock(s.Then)
		} else if s.Else != nil {
			return vm.execBlock(s.Else)
		}

	case *ForStmt:
		// 创建新的作用域
		vm.Env = vm.Env.PushScope()
		defer func() { vm.Env = vm.Env.parent }()

		if s.Init != nil {
			if err := vm.execStmt(s.Init); err != nil {
				return err
			}
		}
		for {
			if s.Cond != nil {
				cond := vm.evalExpr(s.Cond)
				if !cond.IsTrue() {
					break
				}
			}
			if err := vm.execBlock(s.Body); err != nil {
				if _, ok := err.(*ReturnValue); ok {
					return err
				}
				return err
			}
			if s.Post != nil {
				if err := vm.execStmt(s.Post); err != nil {
					return err
				}
			}
		}

	case *WhileStmt:
		for {
			cond := vm.evalExpr(s.Cond)
			if !cond.IsTrue() {
				break
			}
			if err := vm.execBlock(s.Body); err != nil {
				if _, ok := err.(*ReturnValue); ok {
					return err
				}
				return err
			}
		}

	case *ReturnStmt:
		var val Value
		if s.Value != nil {
			val = vm.evalExpr(s.Value)
		}
		return &ReturnValue{Value: val}

	case *BlockStmt:
		vm.Env = vm.Env.PushScope()
		err := vm.execBlock(s)
		vm.Env = vm.Env.parent
		return err
	}

	return nil
}

// evalExpr 求值表达式
func (vm *VM) evalExpr(expr Node) Value {
	switch e := expr.(type) {
	case *IntLiteral:
		return IntValue(e.Value)

	case *StringLiteral:
		return StringValue(e.Value)

	case *IdentExpr:
		if val, ok := vm.Env.Get(e.Name); ok {
			return val
		}
		return Null()

	case *BinaryExpr:
		left := vm.evalExpr(e.Left)
		right := vm.evalExpr(e.Right)
		return vm.evalBinary(e.Op, left, right)

	case *UnaryExpr:
		val := vm.evalExpr(e.Expr)
		if e.Op == "!" {
			if val.IsTrue() {
				return IntValue(0)
			}
			return IntValue(1)
		}
		if e.Op == "-" {
			return IntValue(-val.IntVal)
		}
		return val

	case *CallExpr:
		// 获取函数名
		name := ""
		args := make([]Value, len(e.Args))
		for i, arg := range e.Args {
			args[i] = vm.evalExpr(arg)
		}

		if ident, ok := e.Function.(*IdentExpr); ok {
			name = ident.Name
		} else if member, ok := e.Function.(*MemberExpr); ok {
			name = member.Field
		}

		// 通过内部 callFunc 调用（保持计数器连续）
		result, err := vm.callFunc(name, args)
		if err != nil {
			vm.abortErr = err
			return Null()
		}
		return result

	case *MemberExpr:
		// 简化：暂不处理复杂成员访问
		return Null()

	default:
		return Null()
	}
}

// evalBinary 求值二元运算
func (vm *VM) evalBinary(op string, left, right Value) Value {
	switch op {
	case "+":
		if left.Kind == "string" || right.Kind == "string" {
			return StringValue(left.String() + right.String())
		}
		return IntValue(left.IntVal + right.IntVal)
	case "-":
		return IntValue(left.IntVal - right.IntVal)
	case "*":
		return IntValue(left.IntVal * right.IntVal)
	case "/":
		if right.IntVal == 0 {
			return Null()
		}
		return IntValue(left.IntVal / right.IntVal)
	case "%":
		if right.IntVal == 0 {
			return Null()
		}
		return IntValue(left.IntVal % right.IntVal)
	case "==":
		if left.Kind == "int" && right.Kind == "int" {
			if left.IntVal == right.IntVal {
				return IntValue(1)
			}
			return IntValue(0)
		}
		if left.String() == right.String() {
			return IntValue(1)
		}
		return IntValue(0)
	case "!=":
		if left.String() != right.String() {
			return IntValue(1)
		}
		return IntValue(0)
	case "<":
		return boolVal(left.IntVal < right.IntVal)
	case ">":
		return boolVal(left.IntVal > right.IntVal)
	case "<=":
		return boolVal(left.IntVal <= right.IntVal)
	case ">=":
		return boolVal(left.IntVal >= right.IntVal)
	case "&&":
		if left.IsTrue() && right.IsTrue() {
			return IntValue(1)
		}
		return IntValue(0)
	case "||":
		if left.IsTrue() || right.IsTrue() {
			return IntValue(1)
		}
		return IntValue(0)
	case "++":
		return IntValue(left.IntVal + 1)
	case "--":
		return IntValue(left.IntVal - 1)
	}
	return Null()
}

func boolVal(b bool) Value {
	if b {
		return IntValue(1)
	}
	return IntValue(0)
}
