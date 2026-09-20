// Package kernel 定义内核暴露给 LPC 脚本的 efun 原语
// 原语全部通用：格子/实体属性按名读写，内核不解释任何游戏字段
package kernel

import (
	"fmt"

	"github.com/twinsant/sugarland/lpc"
)

// lpcToAny 把 LPC 值转为 Go 通用值（int/float64/string/[]any），供属性表存储
func lpcToAny(v lpc.Value) any {
	switch v.Kind {
	case "int":
		return v.IntVal
	case "float":
		return v.FloatVal
	case "string":
		return v.StrVal
	case "array":
		arr := make([]any, len(v.ArrVal))
		for i, e := range v.ArrVal {
			arr[i] = lpcToAny(e)
		}
		return arr
	}
	return 0
}

// anyToLPC 把属性表里的 Go 通用值还原为 LPC 值
func anyToLPC(v any) lpc.Value {
	switch t := v.(type) {
	case int:
		return lpc.IntValue(t)
	case int64:
		return lpc.IntValue(int(t))
	case float64:
		return lpc.FloatValue(t)
	case string:
		return lpc.StringValue(t)
	case bool:
		return boolToInt(t)
	case []any:
		arr := make([]lpc.Value, len(t))
		for i, e := range t {
			arr[i] = anyToLPC(e)
		}
		return lpc.ArrayValue(arr)
	case nil:
		return lpc.IntValue(0)
	}
	return lpc.IntValue(0)
}

// registerCitizenEfuns 注册实体脚本可调用的 efun（以 c 为 this_object）
// 调用方需已持有 w.mu
func (w *World) registerCitizenEfuns(vm *lpc.VM, c *Citizen) {
	vm.RegisterEfun("query_x", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(c.X)
	})
	vm.RegisterEfun("query_y", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(c.Y)
	})
	vm.RegisterEfun("query_timestep", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(w.Timestep)
	})
	// query_attr(string name) - 读取自身属性（任意类型）
	vm.RegisterEfun("query_attr", func(args []lpc.Value) lpc.Value {
		if len(args) < 1 {
			return lpc.IntValue(0)
		}
		return anyToLPC(c.Attr(args[0].String()))
	})
	// set_attr(string name, mixed v) - 写入自身属性（任意类型）
	vm.RegisterEfun("set_attr", func(args []lpc.Value) lpc.Value {
		if len(args) >= 2 {
			c.SetAttr(args[0].String(), lpcToAny(args[1]))
		}
		return lpc.Null()
	})
	// query_cell_here(string name) - 读取本格属性
	vm.RegisterEfun("query_cell_here", func(args []lpc.Value) lpc.Value {
		if len(args) < 1 {
			return lpc.FloatValue(0)
		}
		return lpc.FloatValue(w.GetRoom(c.X, c.Y).Get(args[0].String()))
	})
	// query_cell(int x, int y, string name) - 读取指定格属性
	vm.RegisterEfun("query_cell", func(args []lpc.Value) lpc.Value {
		if len(args) >= 3 {
			return lpc.FloatValue(w.GetRoom(args[0].ToInt(), args[1].ToInt()).Get(args[2].String()))
		}
		return lpc.FloatValue(0)
	})
	// move(int x, int y) - 移动（坐标写入并维护 room contents，按环面归一）
	vm.RegisterEfun("move", func(args []lpc.Value) lpc.Value {
		if len(args) >= 2 {
			w.moveCitizenTo(c, args[0].ToInt(), args[1].ToInt())
		}
		return lpc.Null()
	})
	// walk(string dir) - 沿当前 room 的命名 exit 移动，返回 0/1
	vm.RegisterEfun("walk", func(args []lpc.Value) lpc.Value {
		if len(args) < 1 {
			return lpc.IntValue(0)
		}
		if w.walkCitizen(c, args[0].String()) {
			return lpc.IntValue(1)
		}
		return lpc.IntValue(0)
	})
	// query_env() - 返回当前所在 room 的坐标 ({x, y})
	vm.RegisterEfun("query_env", func(args []lpc.Value) lpc.Value {
		return lpc.ArrayValue([]lpc.Value{lpc.IntValue(c.X), lpc.IntValue(c.Y)})
	})
	// take_cell(int x, int y, string name) - 取走指定格某属性全部存量并记账，返回取走的量
	vm.RegisterEfun("take_cell", func(args []lpc.Value) lpc.Value {
		if len(args) < 3 {
			return lpc.FloatValue(0)
		}
		return lpc.FloatValue(w.takeRoomResource(args[0].ToInt(), args[1].ToInt(), args[2].String()))
	})
	// query_contents(int x, int y) - 返回指定 room 站内实体 ID 数组（占位判定）
	vm.RegisterEfun("query_contents", func(args []lpc.Value) lpc.Value {
		if len(args) < 2 {
			return lpc.ArrayValue([]lpc.Value{})
		}
		r := w.GetRoom(args[0].ToInt(), args[1].ToInt())
		ids := make([]lpc.Value, 0, len(r.Contents))
		for id := range r.Contents {
			ids = append(ids, lpc.IntValue(id))
		}
		return lpc.ArrayValue(ids)
	})
	vm.RegisterEfun("grid_width", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(w.Width())
	})
	vm.RegisterEfun("grid_height", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(w.Height())
	})
	// query_config(string name) - 读取世界配置（任意键，内核不解释）
	vm.RegisterEfun("query_config", func(args []lpc.Value) lpc.Value {
		if len(args) < 1 {
			return lpc.Null()
		}
		switch v := w.Config[args[0].String()].(type) {
		case bool:
			return boolToInt(v)
		case int:
			return lpc.IntValue(v)
		case int64:
			return lpc.IntValue(int(v))
		case float64:
			return lpc.FloatValue(v)
		case string:
			return lpc.StringValue(v)
		}
		return lpc.Null()
	})
	vm.RegisterEfun("random", func(args []lpc.Value) lpc.Value {
		if len(args) > 0 && args[0].ToInt() > 0 {
			return lpc.IntValue(w.rng.Intn(args[0].ToInt()))
		}
		return lpc.IntValue(0)
	})
	vm.RegisterEfun("write", func(args []lpc.Value) lpc.Value {
		if len(args) > 0 {
			fmt.Printf("[LPC Citizen#%d] %s\n", c.ID, args[0].String())
		}
		return lpc.Null()
	})
}

// registerRoomEfuns 注册格子脚本可调用的 efun（以 c 为 this_object，room 语义）
// 调用方需已持有 w.mu
func (w *World) registerRoomEfuns(vm *lpc.VM, c *Room) {
	vm.RegisterEfun("query_x", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(c.X)
	})
	vm.RegisterEfun("query_y", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(c.Y)
	})
	vm.RegisterEfun("query_timestep", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(w.Timestep)
	})
	// query_cell_here(string name) - 读取本格属性
	vm.RegisterEfun("query_cell_here", func(args []lpc.Value) lpc.Value {
		if len(args) < 1 {
			return lpc.FloatValue(0)
		}
		return lpc.FloatValue(c.Get(args[0].String()))
	})
	// set_cell_here(string name, float v) - 写入本格属性
	vm.RegisterEfun("set_cell_here", func(args []lpc.Value) lpc.Value {
		if len(args) >= 2 {
			c.Set(args[0].String(), args[1].ToFloat())
		}
		return lpc.Null()
	})
	// take_cell_here(string name) - 取走本格某属性全部存量并记账
	vm.RegisterEfun("take_cell_here", func(args []lpc.Value) lpc.Value {
		if len(args) < 1 {
			return lpc.FloatValue(0)
		}
		return lpc.FloatValue(w.takeRoomResource(c.X, c.Y, args[0].String()))
	})
	// query_cell(int x, int y, string name) - 读取任意格属性
	vm.RegisterEfun("query_cell", func(args []lpc.Value) lpc.Value {
		if len(args) >= 3 {
			return lpc.FloatValue(w.GetRoom(args[0].ToInt(), args[1].ToInt()).Get(args[2].String()))
		}
		return lpc.FloatValue(0)
	})
	vm.RegisterEfun("random", func(args []lpc.Value) lpc.Value {
		if len(args) > 0 && args[0].ToInt() > 0 {
			return lpc.IntValue(w.rng.Intn(args[0].ToInt()))
		}
		return lpc.IntValue(0)
	})
	vm.RegisterEfun("write", func(args []lpc.Value) lpc.Value {
		if len(args) > 0 {
			fmt.Printf("[LPC Room(%d,%d)] %s\n", c.X, c.Y, args[0].String())
		}
		return lpc.Null()
	})
}

// registerWorldEfuns 注册世界脚本可调用的 efun 集合
// 世界脚本通过这些原语编排全部仿真规则
// 调用方需已持有 w.mu
func (w *World) registerWorldEfuns(vm *lpc.VM) {
	// --- 世界与配置查询 ---

	vm.RegisterEfun("query_timestep", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(w.Timestep)
	})
	vm.RegisterEfun("grid_width", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(w.Width())
	})
	vm.RegisterEfun("grid_height", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(w.Height())
	})
	// query_config(string name) - 读取世界配置（任意键，内核不解释）
	vm.RegisterEfun("query_config", func(args []lpc.Value) lpc.Value {
		if len(args) < 1 {
			return lpc.Null()
		}
		switch v := w.Config[args[0].String()].(type) {
		case bool:
			return boolToInt(v)
		case int:
			return lpc.IntValue(v)
		case int64:
			return lpc.IntValue(int(v))
		case float64:
			return lpc.FloatValue(v)
		case string:
			return lpc.StringValue(v)
		}
		return lpc.Null()
	})
	// has_config(string name) - 配置键是否存在（区分缺失与显式 0）
	vm.RegisterEfun("has_config", func(args []lpc.Value) lpc.Value {
		if len(args) < 1 {
			return lpc.IntValue(0)
		}
		_, ok := w.Config[args[0].String()]
		if ok {
			return lpc.IntValue(1)
		}
		return lpc.IntValue(0)
	})
	// set_config(string name, mixed v) - 写入世界配置（mudlib 声明默认参数用）
	vm.RegisterEfun("set_config", func(args []lpc.Value) lpc.Value {
		if len(args) < 2 {
			return lpc.IntValue(0)
		}
		switch args[1].Kind {
		case "int":
			w.Config[args[0].String()] = args[1].IntVal
		case "float":
			w.Config[args[0].String()] = args[1].FloatVal
		case "string":
			w.Config[args[0].String()] = args[1].StrVal
		default:
			return lpc.IntValue(0)
		}
		return lpc.IntValue(1)
	})
	vm.RegisterEfun("query_population", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(len(w.GetAliveCitizens()))
	})

	// --- 格子读写（属性按名存取） ---

	// query_cell(int x, int y, string name) - 读取格子属性
	vm.RegisterEfun("query_cell", func(args []lpc.Value) lpc.Value {
		if len(args) >= 3 {
			return lpc.FloatValue(w.GetRoom(args[0].ToInt(), args[1].ToInt()).Get(args[2].String()))
		}
		return lpc.FloatValue(0)
	})
	// set_cell(int x, int y, string name, float v) - 写入格子属性（纯写入，无夹取）
	vm.RegisterEfun("set_cell", func(args []lpc.Value) lpc.Value {
		if len(args) >= 4 {
			w.GetRoom(args[0].ToInt(), args[1].ToInt()).Set(args[2].String(), args[3].ToFloat())
		}
		return lpc.Null()
	})
	// take_cell(int x, int y, string name) - 取走指定格某属性全部存量并记账，返回取走的量
	vm.RegisterEfun("take_cell", func(args []lpc.Value) lpc.Value {
		if len(args) < 3 {
			return lpc.FloatValue(0)
		}
		return lpc.FloatValue(w.takeRoomResource(args[0].ToInt(), args[1].ToInt(), args[2].String()))
	})

	// --- 实体查询 ---

	vm.RegisterEfun("citizen_ids", func(args []lpc.Value) lpc.Value {
		ids := make([]lpc.Value, 0, len(w.Citizens))
		for _, c := range w.Citizens {
			ids = append(ids, lpc.IntValue(c.ID))
		}
		return lpc.ArrayValue(ids)
	})
	vm.RegisterEfun("citizen_exists", func(args []lpc.Value) lpc.Value {
		if len(args) < 1 {
			return lpc.IntValue(0)
		}
		_, ok := w.Citizens[args[0].ToInt()]
		return boolToInt(ok)
	})
	// query_citizen(int id, string field) - 读取实体字段
	// 内核字段: x, y, alive, agent_controlled, has_script；其余按属性名读取
	vm.RegisterEfun("query_citizen", func(args []lpc.Value) lpc.Value {
		if len(args) < 2 {
			return lpc.Null()
		}
		c, ok := w.Citizens[args[0].ToInt()]
		if !ok {
			return lpc.Null()
		}
		switch args[1].String() {
		case "x":
			return lpc.IntValue(c.X)
		case "y":
			return lpc.IntValue(c.Y)
		case "alive":
			return boolToInt(c.Alive)
		case "agent_controlled":
			return boolToInt(c.IsAgentControlled)
		case "has_script":
			return boolToInt(c.LPCObj != nil)
		}
		return anyToLPC(c.Attr(args[1].String()))
	})

	// --- 实体操作 ---

	// set_citizen(int id, string field, value) - 写入实体字段
	// 内核字段: x, y, alive；其余写入属性表（任意属性名均可）
	vm.RegisterEfun("set_citizen", func(args []lpc.Value) lpc.Value {
		if len(args) < 3 {
			return lpc.IntValue(0)
		}
		c, ok := w.Citizens[args[0].ToInt()]
		if !ok {
			return lpc.IntValue(0)
		}
		switch args[1].String() {
		case "x":
			w.moveCitizenTo(c, args[2].ToInt(), c.Y)
		case "y":
			w.moveCitizenTo(c, c.X, args[2].ToInt())
		case "alive":
			c.Alive = args[2].IsTrue()
		default:
			c.SetAttr(args[1].String(), lpcToAny(args[2]))
		}
		return lpc.IntValue(1)
	})
	// move_citizen(int id, int x, int y) - 移动实体（坐标写入并维护 room contents）
	vm.RegisterEfun("move_citizen", func(args []lpc.Value) lpc.Value {
		if len(args) < 3 {
			return lpc.IntValue(0)
		}
		c, ok := w.Citizens[args[0].ToInt()]
		if !ok {
			return lpc.IntValue(0)
		}
		w.moveCitizenTo(c, args[1].ToInt(), args[2].ToInt())
		return lpc.IntValue(1)
	})
	// spawn_citizen(int x, int y [, string source]) - 生成实体并绑定 LPC 对象，返回 id
	// 属性为空，含义由 mudlib 通过 set_citizen 决定；
	// 第三参数为居民脚本源码，缺省使用默认 citizen.lpc 蓝图
	vm.RegisterEfun("spawn_citizen", func(args []lpc.Value) lpc.Value {
		x := w.rng.Intn(w.Width())
		y := w.rng.Intn(w.Height())
		if len(args) >= 2 {
			x = args[0].ToInt()
			y = args[1].ToInt()
		}
		wx := ((x % w.Width()) + w.Width()) % w.Width()
		wy := ((y % w.Height()) + w.Height()) % w.Height()
		citizen := NewCitizen(w.NextID, wx, wy)
		w.Citizens[citizen.ID] = citizen
		w.NextID++
		w.GetRoom(wx, wy).Contents[citizen.ID] = true
		RecordBirth()
		src := defaultCitizenSource
		if len(args) >= 3 {
			src = args[2].String()
		}
		if err := w.bindCitizenObject(citizen, src); err != nil {
			fmt.Printf("[LPC] spawn_citizen %d blueprint error: %v\n", citizen.ID, err)
		}
		return lpc.IntValue(citizen.ID)
	})
	// remove_citizen(int id, string reason) - 移除实体并记录死亡事件（含属性快照）
	vm.RegisterEfun("remove_citizen", func(args []lpc.Value) lpc.Value {
		if len(args) < 1 {
			return lpc.IntValue(0)
		}
		id := args[0].ToInt()
		c, ok := w.Citizens[id]
		if !ok {
			return lpc.IntValue(0)
		}
		reason := "unknown"
		if len(args) >= 2 {
			reason = args[1].String()
		}
		attrs := make(map[string]any, len(c.Attrs))
		for k, v := range c.Attrs {
			attrs[k] = v
		}
		w.LastDeaths = append(w.LastDeaths, DeathEvent{
			ID:     c.ID,
			X:      c.X,
			Y:      c.Y,
			Reason: reason,
			Attrs:  attrs,
		})
		RecordDeath(reason)
		if c.LPCObj != nil {
			objID := fmt.Sprintf("citizen_%d", id)
			w.HeartBeats.Unregister(objID)
			w.ObjManager.Destroy(objID)
		}
		delete(w.GetRoom(c.X, c.Y).Contents, id)
		delete(w.Citizens, id)
		return lpc.IntValue(1)
	})
	// emit_birth(int a, int b, int child) - 记录一次繁殖事件
	vm.RegisterEfun("emit_birth", func(args []lpc.Value) lpc.Value {
		if len(args) < 3 {
			return lpc.Null()
		}
		w.LastBirths = append(w.LastBirths, MateResult{
			ParentAID: args[0].ToInt(),
			ParentBID: args[1].ToInt(),
			ChildID:   args[2].ToInt(),
			Timestep:  w.Timestep,
		})
		return lpc.Null()
	})

	// --- 空间关系 ---

	// query_neighbors(int id, int radius) - 环面曼哈顿距离 <= radius 的其他存活实体 id 列表
	vm.RegisterEfun("query_neighbors", func(args []lpc.Value) lpc.Value {
		if len(args) < 2 {
			return lpc.ArrayValue([]lpc.Value{})
		}
		a, ok := w.Citizens[args[0].ToInt()]
		if !ok {
			return lpc.ArrayValue([]lpc.Value{})
		}
		radius := float64(args[1].ToInt())
		ids := []lpc.Value{}
		for _, c := range w.Citizens {
			if c.ID == a.ID || !c.Alive {
				continue
			}
			d := torusDist(a.X, a.Y, c.X, c.Y, w.Width(), w.Height())
			if d <= radius {
				ids = append(ids, lpc.IntValue(c.ID))
			}
		}
		return lpc.ArrayValue(ids)
	})

	// --- 实体 scratch flag（每步开始时清空，供 mudlib 记簿使用）---

	vm.RegisterEfun("set_flag", func(args []lpc.Value) lpc.Value {
		if len(args) < 3 {
			return lpc.IntValue(0)
		}
		c, ok := w.Citizens[args[0].ToInt()]
		if !ok {
			return lpc.IntValue(0)
		}
		if c.Flags == nil {
			c.Flags = make(map[string]int)
		}
		c.Flags[args[1].String()] = args[2].ToInt()
		return lpc.IntValue(1)
	})
	vm.RegisterEfun("query_flag", func(args []lpc.Value) lpc.Value {
		if len(args) < 2 {
			return lpc.IntValue(0)
		}
		c, ok := w.Citizens[args[0].ToInt()]
		if !ok || c.Flags == nil {
			return lpc.IntValue(0)
		}
		return lpc.IntValue(c.Flags[args[1].String()])
	})

	// --- 脚本绑定与调用 ---

	vm.RegisterEfun("has_script", func(args []lpc.Value) lpc.Value {
		if len(args) < 1 {
			return lpc.IntValue(0)
		}
		c, ok := w.Citizens[args[0].ToInt()]
		return boolToInt(ok && c.LPCObj != nil)
	})
	// call_citizen(int id, string method) - 调用实体 LPC 方法
	// 返回 1 = 已执行，0 = 无脚本/无方法/心跳间隔未到
	vm.RegisterEfun("call_citizen", func(args []lpc.Value) lpc.Value {
		if len(args) < 2 {
			return lpc.IntValue(0)
		}
		return lpc.IntValue(w.callCitizenMethod(args[0].ToInt(), args[1].String()))
	})
	vm.RegisterEfun("attach_citizen_script", func(args []lpc.Value) lpc.Value {
		if len(args) < 2 {
			return lpc.IntValue(0)
		}
		if err := w.attachAgent(args[0].ToInt(), args[1].String(), 1); err != nil {
			fmt.Printf("[LPC] attach_citizen_script(%d) error: %v\n", args[0].ToInt(), err)
			return lpc.IntValue(0)
		}
		return lpc.IntValue(1)
	})
	vm.RegisterEfun("detach_citizen_script", func(args []lpc.Value) lpc.Value {
		if len(args) < 1 {
			return lpc.IntValue(0)
		}
		if err := w.detachAgent(args[0].ToInt()); err != nil {
			return lpc.IntValue(0)
		}
		return lpc.IntValue(1)
	})

	// bind_cell(int x, int y, string src) - 用 LPC 源码为格子绑定/替换 room 对象
	// 新对象的 create() 会被调用；可用于区域差异行为（特殊地块脚本）
	vm.RegisterEfun("bind_cell", func(args []lpc.Value) lpc.Value {
		if len(args) < 3 {
			return lpc.IntValue(0)
		}
		cell := w.GetRoom(args[0].ToInt(), args[1].ToInt())
		if err := w.bindRoomSource(cell, args[2].String()); err != nil {
			fmt.Printf("[LPC] bind_cell(%d,%d) error: %v\n", args[0].ToInt(), args[1].ToInt(), err)
			return lpc.IntValue(0)
		}
		return lpc.IntValue(1)
	})
	// call_cell(int x, int y, string method) - 调用格子 LPC 对象上的方法
	// 返回 1 = 已执行，0 = 无对象/无方法
	vm.RegisterEfun("call_cell", func(args []lpc.Value) lpc.Value {
		if len(args) < 3 {
			return lpc.IntValue(0)
		}
		cell := w.GetRoom(args[0].ToInt(), args[1].ToInt())
		return lpc.IntValue(w.callRoomMethod(cell, args[2].String()))
	})

	// --- room 拓扑（exits / contents）---

	// link_exit(int x, int y, string dir, int dx, int dy) - 设置 room 命名出口（目标按环面归一）
	vm.RegisterEfun("link_exit", func(args []lpc.Value) lpc.Value {
		if len(args) < 5 {
			return lpc.IntValue(0)
		}
		r := w.GetRoom(args[0].ToInt(), args[1].ToInt())
		dx := ((args[3].ToInt() % w.Width()) + w.Width()) % w.Width()
		dy := ((args[4].ToInt() % w.Height()) + w.Height()) % w.Height()
		r.Exits[args[2].String()] = [2]int{dx, dy}
		return lpc.IntValue(1)
	})
	// unlink_exit(int x, int y, string dir) - 删除 room 的命名出口
	vm.RegisterEfun("unlink_exit", func(args []lpc.Value) lpc.Value {
		if len(args) < 3 {
			return lpc.IntValue(0)
		}
		delete(w.GetRoom(args[0].ToInt(), args[1].ToInt()).Exits, args[2].String())
		return lpc.IntValue(1)
	})
	// query_exit(int x, int y, string dir) - 返回出口目标 ({dx, dy})，无此出口返回 Null
	vm.RegisterEfun("query_exit", func(args []lpc.Value) lpc.Value {
		if len(args) < 3 {
			return lpc.Null()
		}
		if dest, ok := w.GetRoom(args[0].ToInt(), args[1].ToInt()).Exits[args[2].String()]; ok {
			return lpc.ArrayValue([]lpc.Value{lpc.IntValue(dest[0]), lpc.IntValue(dest[1])})
		}
		return lpc.Null()
	})
	// query_exits(int x, int y) - 返回 room 全部出口方向名数组
	vm.RegisterEfun("query_exits", func(args []lpc.Value) lpc.Value {
		if len(args) < 2 {
			return lpc.ArrayValue([]lpc.Value{})
		}
		r := w.GetRoom(args[0].ToInt(), args[1].ToInt())
		dirs := make([]lpc.Value, 0, len(r.Exits))
		for dir := range r.Exits {
			dirs = append(dirs, lpc.StringValue(dir))
		}
		return lpc.ArrayValue(dirs)
	})
	// query_contents(int x, int y) - 返回站内实体 ID 数组（room 的 contents）
	vm.RegisterEfun("query_contents", func(args []lpc.Value) lpc.Value {
		if len(args) < 2 {
			return lpc.ArrayValue([]lpc.Value{})
		}
		r := w.GetRoom(args[0].ToInt(), args[1].ToInt())
		ids := make([]lpc.Value, 0, len(r.Contents))
		for id := range r.Contents {
			ids = append(ids, lpc.IntValue(id))
		}
		return lpc.ArrayValue(ids)
	})
	// walk(int id, string dir) - 沿实体所在 room 的命名 exit 移动，返回 0/1
	vm.RegisterEfun("walk", func(args []lpc.Value) lpc.Value {
		if len(args) < 2 {
			return lpc.IntValue(0)
		}
		c, ok := w.Citizens[args[0].ToInt()]
		if !ok || !w.walkCitizen(c, args[1].String()) {
			return lpc.IntValue(0)
		}
		return lpc.IntValue(1)
	})

	// --- 观测记账 ---

	// emit_stat(string name, float v) - 上报游戏指标，合并入记分板
	vm.RegisterEfun("emit_stat", func(args []lpc.Value) lpc.Value {
		if len(args) >= 2 {
			w.gameStats[args[0].String()] = args[1].ToFloat()
		}
		return lpc.Null()
	})
	// record_consumed(int n) - 记录本轮消耗量（快照统计）
	vm.RegisterEfun("record_consumed", func(args []lpc.Value) lpc.Value {
		if len(args) >= 1 {
			w.stepConsumed += args[0].ToInt()
		}
		return lpc.Null()
	})

	// --- 杂项 ---

	// shuffle(mixed* arr) - 用世界 RNG 洗牌，返回新数组
	vm.RegisterEfun("shuffle", func(args []lpc.Value) lpc.Value {
		if len(args) < 1 || args[0].Kind != "array" {
			return lpc.ArrayValue([]lpc.Value{})
		}
		arr := make([]lpc.Value, len(args[0].ArrVal))
		copy(arr, args[0].ArrVal)
		w.rng.Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})
		return lpc.ArrayValue(arr)
	})
	vm.RegisterEfun("random", func(args []lpc.Value) lpc.Value {
		if len(args) > 0 && args[0].ToInt() > 0 {
			return lpc.IntValue(w.rng.Intn(args[0].ToInt()))
		}
		return lpc.IntValue(0)
	})
	vm.RegisterEfun("write", func(args []lpc.Value) lpc.Value {
		if len(args) > 0 {
			fmt.Printf("[LPC world] %s\n", args[0].String())
		}
		return lpc.Null()
	})
}

func boolToInt(b bool) lpc.Value {
	if b {
		return lpc.IntValue(1)
	}
	return lpc.IntValue(0)
}
