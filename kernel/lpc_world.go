// Package kernel 实现世界级 LPC 对象的加载与驱动
// 服务器是最小内核：世界规则（G/M/Trade/Mate/R）全部在 LPC mudlib 中实现，
// 内核每步只调用世界对象的 heart_beat，规则由脚本编排
package kernel

import (
	"context"
	"fmt"
	"time"

	"github.com/twinsant/sugarland/lpc"
	"github.com/twinsant/sugarland/mudlib"
)

var (
	defaultWorldSource   = mudlib.World
	defaultCitizenSource = mudlib.Citizen
	defaultCellSource    = mudlib.Cell
)

// worldObjectID 是世界对象在 ObjectManager 中的固定 ID
const worldObjectID = "world_daemon"

// worldMaxInstructions 世界脚本的指令预算（mudlib 每步要遍历全图与全部公民）
const worldMaxInstructions = 5000000

// worldCallTimeout 单次世界脚本调用的超时
const worldCallTimeout = 5 * time.Second

// loadDefaultWorld 加载内置 mudlib 世界脚本
// 调用方需已持有 w.mu
func (w *World) loadDefaultWorld() {
	if err := w.attachWorldScript(defaultWorldSource); err != nil {
		// 内置脚本解析失败是编译期错误，必须尽早暴露
		panic(fmt.Sprintf("default mudlib failed to load: %v", err))
	}
	w.WorldSource = defaultWorldSource
	w.WorldCustom = false
}

// AttachWorldScript 用自定义 LPC 脚本替换世界规则（重复调用则替换）
func (w *World) AttachWorldScript(lpcSource string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.attachWorldScript(lpcSource); err != nil {
		return err
	}
	w.WorldSource = lpcSource
	w.WorldCustom = true
	return nil
}

func (w *World) attachWorldScript(lpcSource string) error {
	p := lpc.NewParser(lpcSource)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		return fmt.Errorf("parse errors: %v", p.Errors())
	}
	obj := lpc.NewObject(worldObjectID, "WorldDaemon")
	obj.LoadProgram(prog)
	obj.VM.MaxInstructions = worldMaxInstructions
	w.ObjManager.Add(worldObjectID, obj)
	w.WorldObj = obj
	return nil
}

// DetachWorldScript 卸载自定义世界脚本，恢复默认 mudlib
func (w *World) DetachWorldScript() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.loadDefaultWorld()
}

// GetWorldScript 返回当前世界脚本源码与是否为用户自定义
func (w *World) GetWorldScript() (string, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.WorldSource, w.WorldCustom
}

// lpcWorldHeartBeat 每步调用一次世界对象的 heart_beat
// 世界规则（生长/移动/贸易/繁殖/更替）全部由该脚本编排
// 调用方需已持有 w.mu
func (w *World) lpcWorldHeartBeat() {
	obj := w.WorldObj
	if obj == nil || obj.VM == nil {
		return
	}
	if _, ok := obj.FindMethod("heart_beat"); !ok {
		fmt.Println("[LPC] world_daemon: no heart_beat method, step skipped")
		return
	}
	vm := obj.VM
	vm.ObjManager = w.ObjManager
	w.registerWorldEfuns(vm)

	ctx, cancel := context.WithTimeout(context.Background(), worldCallTimeout)
	defer cancel()
	if _, err := vm.CallFuncWithContext(ctx, "heart_beat", []lpc.Value{}); err != nil {
		fmt.Printf("[LPC] world_daemon heart_beat error: %v\n", err)
	}
}

// callWorldVoid 调用世界对象上的无参方法（若存在）
// 用于世界初始化钩子 create()；调用方需已持有 w.mu
func (w *World) callWorldVoid(name string) {
	obj := w.WorldObj
	if obj == nil || obj.VM == nil {
		return
	}
	if _, ok := obj.FindMethod(name); !ok {
		return
	}
	vm := obj.VM
	vm.ObjManager = w.ObjManager
	w.registerWorldEfuns(vm)
	ctx, cancel := context.WithTimeout(context.Background(), worldCallTimeout)
	defer cancel()
	if _, err := vm.CallFuncWithContext(ctx, name, []lpc.Value{}); err != nil {
		fmt.Printf("[LPC] world_daemon %s error: %v\n", name, err)
	}
}

// CallWorldFunc 调用世界对象上的任意方法（供测试与调试使用）
func (w *World) CallWorldFunc(name string, args []lpc.Value) (lpc.Value, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	obj := w.WorldObj
	if obj == nil || obj.VM == nil {
		return lpc.Null(), fmt.Errorf("no world script loaded")
	}
	vm := obj.VM
	vm.ObjManager = w.ObjManager
	w.registerWorldEfuns(vm)
	ctx, cancel := context.WithTimeout(context.Background(), worldCallTimeout)
	defer cancel()
	return vm.CallFuncWithContext(ctx, name, args)
}

// callRoomMethod 调用指定格子 LPC 对象上的方法
// 返回 1 = 已执行，0 = 无对象/无该方法（调用方可回退）
// 调用方需已持有 w.mu
func (w *World) callRoomMethod(c *Room, method string) int {
	if c.LPCObj == nil || c.LPCObj.VM == nil {
		return 0
	}
	if _, has := c.LPCObj.FindMethod(method); !has {
		return 0
	}
	vm := c.LPCObj.VM
	vm.ObjManager = w.ObjManager
	w.registerRoomEfuns(vm, c)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := vm.CallFuncWithContext(ctx, method, []lpc.Value{}); err != nil {
		fmt.Printf("[LPC] Room(%d,%d) %s error: %v\n", c.X, c.Y, method, err)
	}
	return 1
}

// callCitizenMethod 调用指定公民 LPC 对象上的方法
// 返回 1 = 已执行，0 = 无脚本/无该方法/心跳间隔未到（调用方可回退默认行为）
// 调用方需已持有 w.mu
func (w *World) callCitizenMethod(id int, method string) int {
	c, ok := w.Citizens[id]
	if !ok || c.LPCObj == nil || c.LPCObj.VM == nil {
		return 0
	}
	if _, has := c.LPCObj.FindMethod(method); !has {
		return 0
	}
	objID := fmt.Sprintf("citizen_%d", id)
	if entry := w.HeartBeats.EntryFor(objID); entry != nil {
		if !entry.ShouldCall(w.Timestep) {
			return 0
		}
		entry.MarkCalled(w.Timestep)
	}
	vm := c.LPCObj.VM
	vm.ObjManager = w.ObjManager
	w.registerCitizenEfuns(vm, c)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := vm.CallFuncWithContext(ctx, method, []lpc.Value{}); err != nil {
		fmt.Printf("[LPC] Citizen#%d %s error: %v\n", id, method, err)
	}
	return 1
}
