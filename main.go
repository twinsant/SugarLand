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

	// 静态文件服务（可视化页面）
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "static/index.html")
	})

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("🚀 HTTP 服务启动: http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
