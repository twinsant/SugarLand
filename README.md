# SugarLand - Go FluffOS Runtime (Phase 1)

基于 Go 语言实现的 FluffOS 兼容运行时 + LPC 语言子集 + Sugarscape 仿真系统。

## 架构

```
┌──────────────────────────────┐
│  RESTful API (HTTP)          │  ← 外部 AI Agent / 前端
└──────────┬───────────────────┘
           │
┌──────────▼───────────────────┐
│  Go FluffOS Runtime (VM)     │
│  ┌─────────────────────────┐ │
│  │ Cellspace (50x50 环面)  │ │
│  │ Citizens (智能体)        │ │
│  │ Rules: G → M → R        │ │
│  └─────────────────────────┘ │
└──────────────────────────────┘
```

## Phase 1 功能

- **细胞空间**：50x50 网格，环面拓扑，双峰糖资源地形
- **公民智能体**：视觉、代谢、年龄、财富属性，均匀分布初始化
- **核心规则**：生长(G) → 移动(M) → 更替(R)，随机顺序执行
- **RESTful API**：世界管理、公民查询、地图快照

## 构建和运行

```bash
# 构建
go build -o sugarland .

# 运行（默认端口 8080）
./sugarland

# 指定端口
./sugarland -port 9090

# 自动步进模式（每 500ms 推进一步）
./sugarland -auto -interval 500ms
```

## API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/world` | 世界状态（时间步、人口、配置） |
| POST | `/api/world/step` | 推进仿真一步 |
| POST | `/api/world/reset` | 重置世界 |
| GET | `/api/citizens` | 公民列表 |
| GET | `/api/citizens/:id` | 公民详情 |
| GET | `/api/cellspace` | 地图快照（所有格子） |
| GET | `/api/cells/:x/:y` | 单个格子状态 |

### 示例

```bash
# 查看世界状态
curl http://localhost:8080/api/world

# 推进一步
curl -X POST http://localhost:8080/api/world/step

# 获取第 10 个公民的信息
curl http://localhost:8080/api/citizens/10

# 获取坐标 (25,25) 的格子状态
curl http://localhost:8080/api/cells/25/25
```

## 项目结构

```
├── main.go           # 主入口，HTTP server
├── sim/
│   ├── cell.go       # Cell 结构（糖资源、容量、污染）
│   ├── cellspace.go  # (内嵌在 world.go 中)
│   ├── citizen.go    # Citizen 结构（视觉、代谢、年龄、财富）
│   ├── world.go      # World（初始化、配置、状态）
│   └── rules.go      # G/M/R 规则实现
├── api/
│   └── handlers.go   # RESTful API 处理器
├── lpc/
│   └── ast.go        # LPC 语言 AST（Phase 1 占位）
├── go.mod
└── README.md
```

## Phase 2 计划

- 完整的 LPC lexer + parser + interpreter
- Heart Beat 机制（仿真时间步心跳）
- Object 继承系统
- Agent 绑定接口（AI Agent 接管 Citizen）
