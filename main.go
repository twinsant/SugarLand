// SugarLand - 基于 Go FluffOS 运行时的 Sugarscape 仿真系统
// Phase 3: FluffOS 兼容层 + LPC 脚本驱动 + 接管模式 + 社会行为 + Scoreboard
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

	config := sim.DefaultConfig()
	world := sim.NewWorld(config)

	fmt.Printf("🌍 SugarLand Sugarscape 仿真系统 (Phase 3)\n")
	fmt.Printf("   网格: %dx%d, 初始人口: %d\n", config.Width, config.Height, config.InitPopulation)
	fmt.Printf("   双峰糖地形: (%d,%d) 和 (%d,%d), 峰值容量: %.1f\n",
		config.PeakX1, config.PeakY1, config.PeakX2, config.PeakY2, config.PeakCapacity)
	fmt.Printf("   环面拓扑: 坐标模运算包裹\n")
	fmt.Printf("   规则: G(生长) → M(移动) → Trade(贸易) → Mate(繁殖) → R(更替)\n")
	fmt.Printf("   贸易: %v, 繁殖: %v, 污染: %v\n",
		config.EnableTrading, config.EnableMating, config.EnablePollution)
	fmt.Printf("   季节间隔: %d 步\n\n", config.SeasonInterval)

	if *autoStep {
		go func() {
			ticker := time.NewTicker(*stepInterval)
			defer ticker.Stop()
			for range ticker.C {
				world.Step()
				state := world.GetState()
				sb := world.Scoreboard
				fmt.Printf("[T=%d] 人口=%d, 总糖=%.1f, 基尼=%.3f, Bull=%d, Bear=%d\n",
					state.Timestep, state.Population, state.TotalSugar,
					sb.GiniCoefficient, sb.BullCount, sb.BearCount)
			}
		}()
	}

	mux := http.NewServeMux()
	handler := api.NewHandler(world)
	handler.RegisterRoutes(mux)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>SugarLand Phase 3</title></head>
<body>
<h1>🌍 SugarLand - Sugarscape 仿真 (Phase 3)</h1>
<h2>API 端点</h2>
<ul>
<li><a href="/api/world">GET /api/world</a> - 世界状态</li>
<li>POST /api/world/step - 推进一步</li>
<li>POST /api/world/reset - 重置世界</li>
<li><a href="/api/citizens">GET /api/citizens</a> - 公民列表</li>
<li><a href="/api/cellspace">GET /api/cellspace</a> - 地图快照</li>
<li>GET /api/cells/:x/:y - 单个格子状态</li>
<li><a href="/api/scoreboard">GET /api/scoreboard</a> - 记分板统计</li>
</ul>
<h2>Agent API（LPC 脚本模式）</h2>
<ul>
<li>POST /api/agent/attach - 绑定 LPC 脚本到 citizen</li>
<li>POST /api/agent/detach - 解绑 LPC 脚本</li>
<li>GET /api/agent/context?id=:id - 获取 agent 上下文</li>
</ul>
<h2>AI 接管模式</h2>
<ul>
<li>POST /api/ai/attach - 绑定 AI agent 到 citizen</li>
<li>POST /api/ai/detach - 解绑 AI agent</li>
<li>POST /api/citizens/command - 给被接管的 citizen 下达指令</li>
</ul>
</body></html>`)
	})

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("🚀 HTTP 服务启动: http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
