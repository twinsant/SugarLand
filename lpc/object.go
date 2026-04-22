package lpc

import "fmt"

// Object 表示 LPC 运行时对象
type Object struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Properties map[string]Value       `json:"-"`
	Methods    map[string]*FunctionDecl `json:"-"`
	Parent     *Object                `json:"-"` // 继承链
	Program    *ProgramAST            `json:"-"`
	VM         *VM                    `json:"-"`
}

// NewObject 创建新的 LPC 对象
func NewObject(id, name string) *Object {
	return &Object{
		ID:         id,
		Name:       name,
		Properties: make(map[string]Value),
		Methods:    make(map[string]*FunctionDecl),
	}
}

// GetProperty 获取属性值（沿继承链查找）
func (o *Object) GetProperty(name string) (Value, bool) {
	if v, ok := o.Properties[name]; ok {
		return v, true
	}
	if o.Parent != nil {
		return o.Parent.GetProperty(name)
	}
	return Null(), false
}

// SetProperty 设置属性值
func (o *Object) SetProperty(name string, val Value) {
	o.Properties[name] = val
}

// FindMethod 查找方法（沿继承链查找）
func (o *Object) FindMethod(name string) (*FunctionDecl, bool) {
	if fn, ok := o.Methods[name]; ok {
		return fn, true
	}
	if o.Parent != nil {
		return o.Parent.FindMethod(name)
	}
	return nil, false
}

// LoadProgram 加载 LPC 程序到对象
func (o *Object) LoadProgram(prog *ProgramAST) {
	o.Program = prog
	o.VM = NewVM()
	o.VM.LoadProgram(prog)

	// 将函数也注册到对象的 Methods 表
	for name, fn := range prog.Functions {
		o.Methods[name] = fn
	}

	// 注册默认 efun
	o.VM.RegisterEfun("write", func(args []Value) Value {
		if len(args) > 0 {
			o.VM.Out = append(o.VM.Out, args[0].String())
		}
		return Null()
	})
}

// CallMethod 调用对象方法
func (o *Object) CallMethod(name string, args []Value) (Value, error) {
	if o.VM != nil {
		return o.VM.CallFunc(name, args)
	}
	return Null(), fmt.Errorf("no program loaded")
}

// ObjectManager 管理 LPC 对象的生命周期
type ObjectManager struct {
	objects map[string]*Object
}

// NewObjectManager 创建新的对象管理器
func NewObjectManager() *ObjectManager {
	return &ObjectManager{
		objects: make(map[string]*Object),
	}
}

// Create 创建并注册对象
func (m *ObjectManager) Create(id, name string) *Object {
	obj := NewObject(id, name)
	m.objects[id] = obj
	return obj
}

// Add 添加已有对象到管理器
func (m *ObjectManager) Add(id string, obj *Object) {
	m.objects[id] = obj
}

// Destroy 销毁对象
func (m *ObjectManager) Destroy(id string) {
	delete(m.objects, id)
}

// Find 按 ID 查找对象
func (m *ObjectManager) Find(id string) (*Object, bool) {
	obj, ok := m.objects[id]
	return obj, ok
}

// List 返回所有对象
func (m *ObjectManager) List() []*Object {
	list := make([]*Object, 0, len(m.objects))
	for _, obj := range m.objects {
		list = append(list, obj)
	}
	return list
}
