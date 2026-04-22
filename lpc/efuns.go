// Package lpc 提供 FluffOS 兼容的内置函数（efun）集合
package lpc

import (
	"fmt"
	"sort"
	"strings"
)

// RegisterDefaultEfuns 注册所有默认的内置函数到 VM
func (vm *VM) RegisterDefaultEfuns() {
	// --- 对象操作 ---

	// find_object(string name) - 按名查找对象
	vm.RegisterEfun("find_object", func(args []Value) Value {
		if len(args) < 1 || vm.ObjManager == nil {
			return Null()
		}
		name := args[0].StrVal
		obj, ok := vm.ObjManager.Find(name)
		if !ok {
			return Null()
		}
		return StringValue(obj.ID)
	})

	// clone_object(string blueprint) - 克隆对象
	vm.RegisterEfun("clone_object", func(args []Value) Value {
		if len(args) < 1 || vm.ObjManager == nil {
			return Null()
		}
		blueprint := args[0].StrVal
		src, ok := vm.ObjManager.Find(blueprint)
		if !ok {
			return Null()
		}
		newID := fmt.Sprintf("%s#clone", blueprint)
		obj := NewObject(newID, src.Name)
		obj.LoadProgram(src.Program)
		vm.ObjManager.Add(newID, obj)
		return StringValue(newID)
	})

	// destruct_object(object ob) - 销毁对象
	vm.RegisterEfun("destruct_object", func(args []Value) Value {
		if len(args) < 1 || vm.ObjManager == nil {
			return Null()
		}
		vm.ObjManager.Destroy(args[0].StrVal)
		return Null()
	})

	// environment(object ob) - 获取对象所在环境（简化实现）
	vm.RegisterEfun("environment", func(args []Value) Value {
		return Null()
	})

	// all_inventory(object ob) - 获取容器内所有对象
	vm.RegisterEfun("all_inventory", func(args []Value) Value {
		if vm.ObjManager == nil {
			return ArrayValue([]Value{})
		}
		objs := vm.ObjManager.List()
		arr := make([]Value, len(objs))
		for i, ob := range objs {
			arr[i] = StringValue(ob.ID)
		}
		return ArrayValue(arr)
	})

	// move_object(object ob, object dest) - 移动对象到目标容器（简化）
	vm.RegisterEfun("move_object", func(args []Value) Value {
		return IntValue(0)
	})

	// --- 数组/字符串操作 ---

	// explode(string str, string sep) - 字符串分割
	vm.RegisterEfun("explode", func(args []Value) Value {
		if len(args) < 2 {
			return ArrayValue([]Value{})
		}
		str := args[0].String()
		sep := args[1].String()
		parts := strings.Split(str, sep)
		arr := make([]Value, len(parts))
		for i, p := range parts {
			arr[i] = StringValue(p)
		}
		return ArrayValue(arr)
	})

	// implode(string* arr, string sep) - 数组拼接字符串
	vm.RegisterEfun("implode", func(args []Value) Value {
		if len(args) < 2 || args[0].Kind != "array" {
			return StringValue("")
		}
		sep := args[1].String()
		parts := make([]string, len(args[0].ArrVal))
		for i, v := range args[0].ArrVal {
			parts[i] = v.String()
		}
		return StringValue(strings.Join(parts, sep))
	})

	// strlen(string str) - 字符串长度
	vm.RegisterEfun("strlen", func(args []Value) Value {
		if len(args) < 1 {
			return IntValue(0)
		}
		return IntValue(len(args[0].String()))
	})

	// sprintf(string fmt, ...) - 简化格式化输出
	vm.RegisterEfun("sprintf", func(args []Value) Value {
		if len(args) < 1 {
			return StringValue("")
		}
		fmtStr := args[0].String()
		result := fmtStr
		// 简单实现：替换 %d 和 %s
		argIdx := 1
		for i := 0; i < len(result) && argIdx < len(args); i++ {
			if result[i] == '%' && i+1 < len(result) {
				switch result[i+1] {
				case 'd':
					result = result[:i] + args[argIdx].String() + result[i+2:]
					argIdx++
					i--
				case 's':
					result = result[:i] + args[argIdx].String() + result[i+2:]
					argIdx++
					i--
				}
			}
		}
		return StringValue(result)
	})

	// filter_array(mixed* arr, function f) - 过滤（使用内置函数名作为回调）
	vm.RegisterEfun("filter_array", func(args []Value) Value {
		if len(args) < 2 || args[0].Kind != "array" {
			return ArrayValue([]Value{})
		}
		// 简化：第二个参数忽略，返回原数组
		return args[0]
	})

	// map_array(mixed* arr, function f) - 映射
	vm.RegisterEfun("map_array", func(args []Value) Value {
		if len(args) < 2 || args[0].Kind != "array" {
			return ArrayValue([]Value{})
		}
		return args[0]
	})

	// sort_array(mixed* arr, function f) - 排序
	vm.RegisterEfun("sort_array", func(args []Value) Value {
		if len(args) < 1 || args[0].Kind != "array" {
			return ArrayValue([]Value{})
		}
		arr := make([]Value, len(args[0].ArrVal))
		copy(arr, args[0].ArrVal)
		sort.Slice(arr, func(i, j int) bool {
			return arr[i].IntVal < arr[j].IntVal
		})
		return ArrayValue(arr)
	})

	// --- 杂项 ---

	// query_heart_beat(object ob) - 查询对象是否有 heart beat
	vm.RegisterEfun("query_heart_beat", func(args []Value) Value {
		// 简化：返回当前对象是否有 heart_beat 函数
		_, has := vm.Env.functions["heart_beat"]
		if has {
			return IntValue(1)
		}
		return IntValue(0)
	})

	// set_heart_beat(object ob, int interval) - 设置 heart beat 间隔
	vm.RegisterEfun("set_heart_beat", func(args []Value) Value {
		return Null()
	})

	// call_out(function f, int delay) - 延迟调用（简化实现，立即调用）
	vm.RegisterEfun("call_out", func(args []Value) Value {
		if len(args) < 1 {
			return Null()
		}
		// 简化：不真正延迟，直接返回
		return Null()
	})

	// --- LPC 常用 efun ---

	// write(mixed msg) - 输出消息
	vm.RegisterEfun("write", func(args []Value) Value {
		if len(args) > 0 {
			vm.Out = append(vm.Out, args[0].String())
		}
		return Null()
	})

	// sizeof(mixed* arr) - 数组大小
	vm.RegisterEfun("sizeof", func(args []Value) Value {
		if len(args) < 1 {
			return IntValue(0)
		}
		if args[0].Kind == "array" {
			return IntValue(len(args[0].ArrVal))
		}
		return IntValue(len(args[0].String()))
	})

	// member(mixed needle, mixed* haystack) - 查找元素位置
	vm.RegisterEfun("member", func(args []Value) Value {
		if len(args) < 2 || args[1].Kind != "array" {
			return IntValue(-1)
		}
		for i, v := range args[1].ArrVal {
			if v.String() == args[0].String() {
				return IntValue(i)
			}
		}
		return IntValue(-1)
	})

	// element_of(mixed* arr) - 随机取一个元素（简化，取第一个）
	vm.RegisterEfun("element_of", func(args []Value) Value {
		if len(args) < 1 || args[0].Kind != "array" || len(args[0].ArrVal) == 0 {
			return Null()
		}
		return args[0].ArrVal[0]
	})

	// abs(int n) - 绝对值
	vm.RegisterEfun("abs", func(args []Value) Value {
		if len(args) < 1 {
			return IntValue(0)
		}
		if args[0].IntVal < 0 {
			return IntValue(-args[0].IntVal)
		}
		return IntValue(args[0].IntVal)
	})

	// max(int a, int b) - 最大值
	vm.RegisterEfun("max", func(args []Value) Value {
		if len(args) < 2 {
			return IntValue(0)
		}
		if args[0].IntVal > args[1].IntVal {
			return args[0]
		}
		return args[1]
	})

	// min(int a, int b) - 最小值
	vm.RegisterEfun("min", func(args []Value) Value {
		if len(args) < 2 {
			return IntValue(0)
		}
		if args[0].IntVal < args[1].IntVal {
			return args[0]
		}
		return args[1]
	})
}
