# SugarLand 前端

基于 **Vite + React + TypeScript** 的 SugarLand 可视化前端，对应现有 `static/index.html` 单文件实现的迁移骨架。

## 技术栈

| 工具 | 用途 |
|------|------|
| [Vite](https://vitejs.dev/) | 构建工具与开发服务器 |
| [React 18](https://react.dev/) | UI 框架 |
| [TypeScript](https://www.typescriptlang.org/) | 类型安全 |
| [TanStack Query v5](https://tanstack.com/query) | 服务端状态管理（API 数据获取与缓存） |
| [Zustand v5](https://zustand-demo.pmnd.rs/) | 客户端状态管理（UI 状态） |

---

## 安装依赖

```bash
cd frontend
npm install
```

---

## 运行前端（开发模式）

```bash
npm run dev
```

启动后访问 [http://localhost:5173](http://localhost:5173)。

---

## 与 Go 后端联调

前端通过 `vite.config.ts` 中的代理将所有 `/api/*` 请求转发到 `http://localhost:8080`，因此联调步骤如下：

1. 先启动 Go 后端：

   ```bash
   # 在仓库根目录
   go run .
   ```

2. 再启动前端开发服务器：

   ```bash
   cd frontend
   npm run dev
   ```

3. 打开浏览器访问 `http://localhost:5173`，前端会自动向 `localhost:8080` 发送 API 请求。

---

## 构建生产包

```bash
npm run build
```

构建产物在 `frontend/dist/`，可作为 Go 服务的静态文件托管目录（后续集成时使用）。

---

## 目录结构

```
frontend/
├── index.html              # Vite 入口 HTML
├── vite.config.ts          # Vite 配置（含 API 代理）
├── tsconfig*.json          # TypeScript 配置
├── package.json
└── src/
    ├── main.tsx            # React 挂载入口
    ├── App.tsx             # 根组件
    ├── styles.css          # 全局深色赛博风样式
    ├── lib/
    │   ├── fetcher.ts      # 通用 fetch/POST 工具
    │   ├── math.ts         # 数学工具（Gini、直方图、环面坐标）
    │   └── colors.ts       # 颜色映射工具
    ├── store/
    │   └── uiStore.ts      # Zustand UI 状态（播放/速度/热力图）
    ├── features/
    │   ├── world/          # 世界状态（类型、API、hooks）
    │   ├── citizens/       # 公民数据（类型、API、hooks）
    │   ├── cells/          # 格子数据（类型、API）
    │   └── simulation/     # 仿真控制（step/reset API）
    └── components/
        ├── layout/         # AppLayout、MapPanel、InfoPanel
        ├── controls/       # ControlsBar、SpeedControl
        ├── map/            # WorldCanvas（Canvas 渲染）、Tooltip
        └── stats/          # StatsGrid、WealthChart、CitizenTopList、Legend、ConfigInfo
```

---

## 已实现功能

- [x] 左侧地图面板 + 右侧信息面板布局
- [x] 标题 "🌍 SugarLand 元宇宙25号" 与 timestep 显示
- [x] 开始/暂停、单步、重置、热力图开关控件
- [x] 速度滑块（100ms – 3000ms）
- [x] Canvas 渲染：糖热力地形、双峰标记、公民点位、网格线
- [x] 鼠标悬停 Tooltip（格子糖量/容量/污染 + 公民信息）
- [x] 实时统计（人口、总糖量、基尼系数、平均年龄）
- [x] 财富分布直方图
- [x] 图例
- [x] 公民 TOP 10 列表
- [x] 配置信息展示
- [x] API 对接：`GET /api/world`、`GET /api/citizens`、`POST /api/world/step`、`POST /api/world/reset`
- [x] `/api/cells/:x/:y` 格子数据按需拉取并缓存

## 待迁移 / 待完善

- [ ] 格子缓存刷新策略（目前仅初始加载；完整实现应在每 N 步后清空缓存并重新批量拉取）
- [ ] `GET /api/cellspace` 全图快照接口的接入（可替代逐个格子拉取，性能更好）
- [ ] 窗口 resize 时 canvas 尺寸自适应逻辑优化（目前依赖父容器 offsetWidth/Height）
- [ ] 生产环境下将构建产物集成到 Go 静态文件服务
- [ ] 路由支持（多页面/Tab）
- [ ] 更丰富的 AI Agent 控制面板
- [ ] 单元测试 / E2E 测试
