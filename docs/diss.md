# 技术讨论：Go 实现 FluffOS 运行时 & RESTful API 架构

> 日期：2026-04-21
> 背景：SugarLand（元宇宙25号）项目技术方案讨论

---

## 构想

用 Go 语言实现一个 FluffOS 兼容的运行时（VM），支持 LPC 语言子集，作为 SugarLand 仿真系统的脚本引擎和执行环境。

相关参考项目：
- [ethos](https://github.com/twinsant/ethos)：基于以太坊的 LPmud driver，可参考其 FluffOS 架构

## 为什么选择 LPC？

LPmud 本身就是为"大量异构对象在共享环境里互动"设计的虚拟社会系统，与 Sugarscape 有天然对应关系：

| LPC 概念 | Sugarscape 对应 |
|----------|----------------|
| Object（对象） | Citizen / Cell |
| Room（房间） | Cell（网格单元） |
| Inheritance（继承） | 不同类型的 Citizen 变体 |
| Environment（环境） | Cellspace |
| Heart Beat（心跳） | 仿真时间步 |
| Command（命令） | Citizen 的行为规则 |
| Set/Get Properties | 状态属性（糖量、寿命等） |

## 分阶段实施

- **Phase 1**：实现 LPC 子集（变量、函数、if/for、struct），跑通 Sugarscape 基本规则
- **Phase 2**：加上 Heart Beat 和 Object 系统，模拟时间步推进
- **Phase 3**：FluffOS 兼容层和 efun 扩展

## RESTful API 接口设计

对外暴露 RESTful API，让外部 AI Agent（如 OpenClaw）可以接管虚拟世界中的智能体。

### 架构

```
┌─────────────────────────────────────┐
│  OpenClaw / 其他 AI Agent           │  ← 外部智能体
│  (LLM 驱动的决策)                    │
└──────────────┬──────────────────────┘
               │ RESTful API
               ▼
┌─────────────────────────────────────┐
│  Go FluffOS Runtime (VM)            │  ← 虚拟世界
│  ┌───────────────────────────────┐  │
│  │ LPC 脚本 (基础规则/行为)       │  │
│  │ Citizens / Cells / Cellspace  │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
```

### 世界管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/world` | 世界全局状态（时间步、配置、统计） |
| POST | `/api/world/step` | 推进仿真一步 |
| POST | `/api/world/reset` | 重置世界 |

### Citizen 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/citizens` | 所有公民列表 |
| GET | `/api/citizens/:id` | 某个公民的详细状态 |
| GET | `/api/citizens/:id/perceive` | 获取公民的感知（视野内信息，用于构建 LLM prompt） |
| POST | `/api/citizens/:id/command` | 给公民下达指令（移动、收集、交易等） |

### Cellspace 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/cells/:x/:y` | 某格子的状态（资源、占用情况） |
| GET | `/api/cellspace` | 整个地图快照 |

### Agent 绑定

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/agent/attach` | AI agent 绑定到某个 citizen |
| POST | `/api/agent/detach` | 解绑 |
| GET | `/api/agent/:id/context` | 获取 agent 的完整上下文（用于 LLM prompt） |

## 两种运行模式

### 自主模式（Autonomous）

- Citizen 按 LPC 脚本的硬编码规则行动
- 用于跑大规模宏观模拟，观察涌现现象（贫富分化、贸易、迁徙）
- 无外部干预，纯规则驱动

### 接管模式（Puppet）

- OpenClaw 通过 RESTful API 绑定到某个 Citizen
- 每个时间步：VM 暂停 → 返回感知信息 → AI Agent 做决策 → 下达命令 → VM 执行 → 推进
- 实现"人类/AI 混合社会"模拟

### 混合模式

- 部分 Citizen 由 AI Agent 驱动，部分由 LPC 规则驱动
- 观察 AI 策略与规则行为的交互结果
- 可对比同一场景下 AI 驱动 vs 规则驱动的差异

## 优势

- **可插拔**：不同 AI Agent 可以"扮演"不同角色，随时切换
- **可观测**：所有决策都有 API 日志，方便事后分析
- **可对比**：同场景下规则驱动 vs AI 驱动的行为差异研究
- **可扩展**：未来可以接入多个 AI Agent，模拟更复杂的社会结构

---

*本文档随项目推进持续更新。*
