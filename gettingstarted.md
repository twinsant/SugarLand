# SugarLand 快速开始

## 项目简介

SugarLand 是一个使用 Go 开发的、面向 AI 的 Sugarscape 仿真平台。它把经典的 Sugarscape 智能体模型、网页可视化、世界状态统计，以及可扩展的 LPC/FluffOS 风格运行时结合在一起，用于模拟一个可演化的虚拟社会环境。

## 这个项目能做什么

- 在 50×50 的环面网格上运行 Sugarscape 仿真
- 让智能体进行移动、采集资源、交易、繁殖和演化
- 统计人口、财富、基尼系数、平均属性等指标
- 通过 HTTP API 提供世界状态与控制能力
- 在浏览器中查看可视化界面
- 为后续接入 AI Agent 和脚本控制预留扩展能力

## 技术栈

- **后端**：Go
- **前端**：HTML
- **核心模块**：仿真引擎 `sim/`
- **脚本运行时**：LPC 子集 `lpc/`
- **界面**：`static/index.html`

## 仓库结构

- `main.go`：程序入口
- `api/`：HTTP API 路由与处理
- `sim/`：Sugarscape 仿真核心逻辑
- `lpc/`：LPC 词法、AST、虚拟机与对象系统
- `static/`：前端静态页面
- `docs/`：设计讨论与技术文档
- `SPEC.md`：详细技术规格说明

## 运行环境

建议准备：

- Go 1.22 或更高版本
- 支持 HTTP 的现代浏览器

## 本地运行

### 1. 克隆仓库

```bash
git clone https://github.com/twinsant/SugarLand.git
cd SugarLand
```

### 2. 启动项目

```bash
go run .
```

### 3. 打开浏览器

在浏览器访问：

```text
http://localhost:8080
```

你会看到 SugarLand 的可视化界面和仿真状态。

## 启动参数

项目支持以下命令行参数：

- `-port`：HTTP 服务端口，默认 `8080`
- `-auto`：开启自动步进模式
- `-interval`：自动步进间隔，默认 `500ms`

示例：

```bash
go run . -port 8080 -auto -interval 200ms
```

## 启动后会发生什么

程序启动后会：

1. 创建默认的 Sugarscape 世界
2. 输出当前网格、人口和规则配置
3. 启动 HTTP 服务
4. 提供前端页面和 API 接口
5. 如果开启 `-auto`，则按固定间隔自动推进仿真

## 推荐先看哪些内容

如果你想快速了解项目，建议按这个顺序看：

1. `README.zh-CN.md`：项目概览
2. `main.go`：启动流程
3. `sim/`：仿真规则和世界状态
4. `lpc/`：脚本运行时实现
5. `SPEC.md`：完整设计细节

## 项目定位

SugarLand 不是一个单纯的演示程序，而是一个研究型仿真平台。它适合用来探索：

- 多智能体系统
- 财富不平等
- 市场与交易机制
- 社会行为演化
- AI Agent 驱动的虚拟世界

## 一句话总结

SugarLand 是一个基于 Go 的 AI 驱动 Sugarscape 仿真平台，提供网页可视化、世界统计和可扩展脚本运行时，用于构建可演化的虚拟社会实验环境。
