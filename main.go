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
	"github.com/twinsant/sugarland/kernel"
)

func main() {
	port := flag.Int("port", 8080, "HTTP 服务端口")
	autoStep := flag.Bool("auto", false, "启动自动步进模式")
	stepInterval := flag.Duration("interval", 500*time.Millisecond, "自动步进间隔（仅 auto 模式）")
	flag.Parse()

	config := kernel.DefaultConfig()
	world := kernel.NewWorld(config)

	fmt.Printf("🌍 SugarLand 最小内核 LPC VM 服务器\n")
	fmt.Printf("   网格: %dx%d（环面拓扑）, 世界规则: mudlib/world.lpc\n", config.Int("width"), config.Int("height"))
	fmt.Printf("   内核只管状态存储 / LPC VM / efun 原语 / 快照；全部游戏规则在 mudlib\n\n")

	if *autoStep {
		go func() {
			ticker := time.NewTicker(*stepInterval)
			defer ticker.Stop()
			for range ticker.C {
				world.Step()
				state := world.GetState()
				sb := world.Scoreboard
				fmt.Printf("[T=%d] 人口=%d, 出生=%v, 死亡=%v\n",
					state.Timestep, state.Population, sb["birth_count"], sb["death_count"])
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
	log.Fatal(http.ListenAndServe(addr, logMiddleware(corsMiddleware(mux))))
}

// statusWriter 捕获响应状态码供日志使用
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// logMiddleware 记录每个 HTTP 请求的方法/路径/状态码/耗时
func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, lw.status, time.Since(start).Round(time.Microsecond))
	})
}

// corsMiddleware 允许跨域访问（Godot Web 导出端可用）
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
