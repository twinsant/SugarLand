// Package mudlib 嵌入默认 LPC 游戏库源码
// 本目录下的 .lpc 文件即游戏本体：换一套 mudlib 就是另一种游戏，
// 内核（kernel/）不需要任何改动
package mudlib

import _ "embed"

//go:embed world.lpc
var World string

//go:embed citizen.lpc
var Citizen string

//go:embed cell.lpc
var Cell string
