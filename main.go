// SugarLand - 基于 Go FluffOS 运行时的 Sugarscape 仿真系统
// Phase 1: 核心仿真模型 + RESTful API
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/twinsant/sugarland/api"
	"github.com/twinsant/sugarland/sim"
)

func main() {
	port := flag.Int("port", 8080, "HTTP 服务端口")
	autoStep := flag.Bool("auto", false, "启动自动步进模式")
	stepInterval := flag.Duration("interval", 500*time.Millisecond, "自动步进间隔（仅 auto 模式）")
	flag.Parse()

	// 创建默认世界
	config := sim.DefaultConfig()
	world := sim.NewWorld(config)

	fmt.Printf("🌍 SugarLand Sugarscape 仿真系统\n")
	fmt.Printf("   网格: %dx%d, 初始人口: %d\n", config.Width, config.Height, config.InitPopulation)
	fmt.Printf("   双峰糖地形: (%d,%d) 和 (%d,%d), 峰值容量: %.1f\n",
		config.PeakX1, config.PeakY1, config.PeakX2, config.PeakY2, config.PeakCapacity)
	fmt.Printf("   环面拓扑: 坐标模运算包裹\n")
	fmt.Printf("   规则顺序: G(生长) → M(移动) → R(更替)\n\n")

	// 自动步进模式
	if *autoStep {
		go func() {
			ticker := time.NewTicker(*stepInterval)
			defer ticker.Stop()
			for range ticker.C {
				world.Step()
				state := world.GetState()
				fmt.Printf("[T=%d] 人口=%d, 总糖=%.1f\n", state.Timestep, state.Population, state.TotalSugar)
			}
		}()
	}

	// 注册 API 路由
	mux := http.NewServeMux()
	handler := api.NewHandler(world)
	handler.RegisterRoutes(mux)

	// 根路径返回简单欢迎页
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>SugarLand</title></head>
<body>
<h1>🌍 SugarLand - Sugarscape 仿真</h1>
<p>API 端点:</p>
<ul>
<li><a href="/api/world">GET /api/world</a> - 世界状态</li>
<li>POST /api/world/step - 推进一步</li>
<li>POST /api/world/reset - 重置世界</li>
<li><a href="/api/citizens">GET /api/citizens</a> - 公民列表</li>
<li><a href="/api/cellspace">GET /api/cellspace</a> - 地图快照</li>
<li>GET /api/cells/:x/:y - 单个格子状态</li>
</ul>
</body></html>`)
	})

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("🚀 HTTP 服务启动: http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
