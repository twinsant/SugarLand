# SugarLand Godot 本地仿真

- 独立 Godot 4 项目，使用 GDScript 和 Compatibility 渲染器，无插件、外部素材或后端请求。启动时默认使用全屏窗口；启动先显示主菜单，开始游戏后才运行仿真；保持精简，不将 Go/React 启动流程作为前置条件。
- `simulation.gd` 是无场景依赖的单资源经济模型；`world.gd` 连接模型和界面；`main.tscn` 定义界面；`resident.gd` 只绘制和插值移动，不修改经济状态。
- 默认 50×50 环面、80 人、种子 25。每轮先糖再生（+1 至容量），再用模型自有 RNG 随机排序居民。居民在四条正交视线内选择未占用格：糖最多、距离最近、平局随机；允许原地停留，每格只容纳一人。财富充足（≥ 代谢 × 25）的居民不再追糖，在视野内随机闲逛。
- 采集保留小数，随后扣代谢并增加年龄；移动消耗 `move_cost × 环面距离`（默认系数 0.5，可在设置面板调整），原地停留不扣。财富 <=0 判饥饿，否则年龄达到寿命判寿终。全部居民行动后移除死者并补足80人；新人带启动财富，当轮不行动。这是开放资源模型，不是人口自然繁殖或封闭财富守恒模型，没有交易、继承或双资源。
- `get_stats()` 的 growth、harvested、consumed、move_cost、rent_paid、injected_wealth、removed_wealth、released_wealth 均为上一轮预算；consumed 只含代谢，move_cost 是移动消耗（系数 × 环面距离），rent_paid 是私有土地租金；removed_wealth 是带符号死者剩余财富，可为负，released_wealth 是其中实际转回地块且受容量限制的正财富。累计死亡和补入单独记录。`gini` 是财富基尼，`income_gini` 是收入基尼（基于 `total_harvest` 累计采集），只统计当前存活居民；同时记录峰值 `peak_gini` 和 `peak_income_gini`，在统计面板显示。
- 经济回合不依赖动画帧率。空格暂停两者；N 或单步按钮执行一轮、立即显示目的地并保持暂停；R 或重置按钮恢复同一种子和初始统计并开始运行。选中居民死亡后保留原因展示，不转移选择到新居民。
- `simulation.sugar` 是即时经济状态，`world.gd` 的 `display_sugar` 是动画显示状态：每轮先反映生长，已采集地块暂时保留采集前的显示值，居民到达后由 `harvest_ready` 信号切换为采集后的值。这样地块不会在居民移动途中提前变色。
- 地图显示位于可裁剪的 `MapClip/MapLayer` 下；`MapLayer` 与 `Residents` 一起缩放，Sidebar 不缩放。滚轮和触控板捏合缩放，范围为 75%–300%，缩放时光标位置保持不动；触控板双指平移仅在地图已放大（缩放 > 100%）且光标位于地图区域内时生效，反转系统手势 delta 并使用 32 倍 macOS PanGesture 灵敏度补偿，使地图内容跟随手指方向；鼠标中/右键拖动保持直接跟手，`0` 恢复视图。点击先将全局坐标转换到 MapLayer 本地坐标，再计算格子。平移、缩放和自动跟随都会把 `MapLayer` 限制在 `MapClip` 边界内，不能看到地图以外的空白；选中居民接近地图视口边缘两格或移出视口时，MapLayer 以指数平滑移动，让居民回到视口中心（缩放 100% 时不平移）。
- 选中居民的 `ThoughtBubble` 显示居民最近动作：找糖、闲逛、交租、买地、继承、繁殖、出生等；它是表现层提示，不参与居民决策。气泡接近地图视口上边界时自动翻到居民下方，三角尖端同步翻转。
- 每轮死亡事件记录实际格子，`DeathMarker` 在该位置绘制骷髅，显示 2.5 秒后渐隐并释放；重置时清理所有死亡标记。
- 选中居民时自动开启 `F` 控制的径向视野迷雾：视野圆圈内清晰，边缘外逐渐变黑；再次按 `F` 可关闭，点击空地也会关闭。选中居民和 ThoughtBubble 位于迷雾上层。`E` 切换选中居民的用户接管，方向键提交上下左右移动；接管期间强制开启迷雾，释放后恢复普通迷雾开关。选中居民死亡后按财富最高、财富最低交替选择新的居民。
- `intro.json` 保存游戏介绍页文本；`tutorial_guide.gd` 挂在唯一真实居民节点上提供分页和按钮，并复用该居民的 `ThoughtBubble` 显示教程文本。游戏启动直接进入世界内教程，教程只显示一个居民并暂停仿真，完成教程后才开始当前游戏逻辑。
- 游戏内录像由 `V` 或录像按钮控制，固定 30 FPS 将视口保存为 PNG 序列到 `/tmp/sugarland/session_<timestamp>/`，暂停仿真时仍可录制。PNG 编码和写盘在后台线程执行，避免阻塞主线程。停止录像时自动查找 ffmpeg，先裁剪到偶数宽高，再生成 `~/.sugarland/recordings/session_<timestamp>/sugarland.mp4`；找不到 ffmpeg 或转码失败时保留 PNG。录像期间每轮 `get_stats()` 结果（含 tick、death_events、breeding_events）逐行写入同目录 `stats.jsonl`。Web 构建中 `OS.has_feature("web")` 隐藏录像按钮与状态标签并禁用 `V` 录像，Footer 同时去掉录像说明。
- 经济参数由 `simulation.gd` 的实例变量保存（双峰位置 `peak_a`/`peak_b`、容量 `peak_capacity`、衰减 `peak_falloff`、再生速率 `growth_rate`、人口补充方式 `repopulate_mode`、移动消耗系数 `move_cost`、土地私有开关 `land_privacy_enabled`、租金比例 `rent_rate`），默认值与原硬编码一致。Sidebar 的「设置」按钮打开 `settings_panel.gd` 模态面板，调整这些参数及步进间隔 `step_seconds`；应用时调用 `world.apply_settings()` 重置仿真以重建地图，取消或 ESC 关闭面板并恢复暂停前的运行状态。
- 人口补充方式 `repopulate_mode` 支持两种模式：`"system"`（系统补人，默认）死亡后随机生成新人并注入随机财富，维持 80 人；`"breed"`（繁殖）居民主动繁殖，死亡后不补人，人口动态变化。繁殖条件：两个存活居民在各自正交视野内（双向可见），且双方财富 ≥ 各自代谢 × 5（足够后代存活一小段时间），且双方年龄在 `breed_age_min`～`breed_age_max`（默认 15～65）范围内。繁殖时父母各出 30% 财富给孩子，孩子继承父母平均视野/代谢/寿命并带 ±1/±5 变异，出生在父母中点附近空格。繁殖事件记录到 `breeding_events`（仅含父母和孩子 ID），`world.gd` 在 `sync_residents` 后为每个事件创建 `HeartMarker`：心形连线两端跟踪父母的 Resident 动态位置，心在连线中点，0.8 秒后孩子在该位置显现，心和连线立即消失。繁殖模式下居民死亡时，正财富平分给存活的孩子（通过 `children` 数组记录亲子关系），没有存活孩子才释放到地块。
- 土地私有 `land_privacy_enabled`（默认关闭）开启后，居民所站位置 3×3 范围内的无主土地按单格收购价 `peak_capacity × 2` 购买（2 倍地价），优先购买中心格再按距离由近及远，每次购买后保留 `metabolism × 25` 生存财富，能买几格买几格。私有土地记入 `land_owners` 字典（格子索引 → 地主 ID）。其他居民在私有土地上采集时，80%（`rent_rate`）归地主，20% 归采集者。地主死亡时土地继承给最富有的存活孩子，没有存活孩子才释放为无主土地。地图上私有土地用地主代际颜色的边框标记。
- `system` 模式下所有居民死亡后仿真自动停止，重置后恢复。居民死亡不补人，也不累计 tick。
- 远程模式：启动参数 `-- --server=http://127.0.0.1:8080` 或环境变量 `SUGARLAND_SERVER` 指向 Go 服务（仓库根目录 `go run .`）后，`world.gd` 改用 `remote_simulation.gd`（继承 `simulation.gd`，接口字段一致），本地模式行为完全不变。`backend_client.gd` 是同步 HTTP/JSON 客户端，本机调用按毫秒计、在主线程内阻塞完成。
- 远程模式数据流：`reset()`→`POST /api/world/reset`（body 可带 `{"config": {...}}` 覆盖游戏参数，内核 `ApplyConfig` 先合并再重置，mudlib `has_config` 让显式传入优先于声明默认值），`step()`→`POST /api/world/step` 后 `GET /api/snapshot` 一次取全（糖量/容量扁平数组、居民、私有地块 `owners`、记分板、配置 `config`、本轮死亡/出生/按格采集量）。居民对象按 ID 复用更新以保证选中引用稳定；服务端坐标可能越界（GetCell 查询时才包裹），客户端按环面取模归一。死亡事件 reason 的 "hunger" 映射为 "starvation"。
- 服务端是最小内核 LPC VM：仿真规则（G 生长 → M 移动 → Trade → Mate → R 更替）和世界生成（双峰地形、初始人口、基因分布）全部由 LPC mudlib（`mudlib/world.lpc`，go:embed 内置）实现，Go 只做状态存储、VM、efun 原语、HTTP 与快照。初始化/重置时内核调用世界对象的 `create()`，每步 `POST /api/world/step` 只调用 `heart_beat()`；挂接世界脚本（`POST /api/world/agent`）是整体替换规则而非叠加钩子，卸载（`POST /api/world/agent/detach`）恢复默认 mudlib；`GET /api/world/agent` 返回当前生效源码。
- LPC 对接：Sidebar 的「LPC 脚本」按钮（仅远程可见）打开 `lpc_panel.gd`，可向选中居民挂接 `heart_beat` 脚本（`POST /api/agent/attach`，interval 步间隔生效），或编辑世界规则（面板自动拉取当前生效源码）。每个居民天生绑定 LPC 对象：`spawn_citizen` 自动挂默认蓝图 `mudlib/citizen.lpc`（`spawn_citizen(x,y,src)` 第三参数可换蓝图），attach 等价于替换该对象的程序，detach 恢复蓝图。格子同样是 LPC room 对象：内核初始化时每格绑定 `mudlib/cell.lpc` 蓝图（共享 AST，2500 对象），`bind_cell` 可给特定格子换脚本实现区域差异行为，`call_cell` 驱动。居民 heart_beat 由 mudlib 通过 `call_citizen(id, "heart_beat")` 驱动，脚本负责移动与采集，代谢/衰老/死亡由 mudlib 统一结算。世界脚本 efun 见 `kernel/efuns.go`，全部通用、不含游戏字段：`query_cell(x,y,name)`/`set_cell(x,y,name,v)`/`take_cell(x,y,name)`（格子属性按名存取，take 记账）、`query_citizen(id,field)`/`set_citizen(id,field,v)`（内核字段 x/y/alive 之外全部透传属性表）、`move_citizen`、`spawn_citizen`（空白实体）、`remove_citizen(id,reason)`、`citizen_ids`、`query_neighbors(id,radius)`、`call_citizen`、`set_flag/query_flag`、`emit_stat(name,v)`（游戏指标上报记分板）、`record_consumed(n)`、`query_config(name)`/`has_config(name)`/`set_config(name,v)`（配置读写；游戏默认值由 mudlib create() 用 has_config/set_config 声明补齐，内核 DefaultConfig 只含 width/height）、`bind_cell(x,y,src)`（给格子换 room 对象蓝图，create() 自动调用）、`call_cell(x,y,method)`（驱动格子对象）、room 拓扑 `link_exit/unlink_exit(x,y,dir[,dx,dy])`/`query_exit`/`query_exits`/`query_contents(x,y)`/`walk(id,dir)`（mudlib 用 link_exit 在 create() 里建连接表，默认 mudlib 建四邻环面出口）、`shuffle`、`random`、`push` 等；居民脚本另有 `query_x/query_y`、`query_attr/set_attr`、`query_cell_here(name)`、`move`（写坐标并维护 contents）、`walk(dir)`（沿 exit 移动）、`query_env()`（所在 room 坐标）、`take_cell`；格子脚本另有 `query_x/query_y`、`query_cell_here/set_cell_here/take_cell_here`、`query_cell`。Room 语义：每个 room（原 Cell）持有 `Exits`（方向名→目标坐标，mudlib 维护）与 `Contents`（站内实体 ID 集合，内核在 spawn/move/remove 时同步维护）。LPC 支持 int/float/string/数组、`({...})` 字面量、`break/continue`、字符串转义（`\"` `\n` `\t` `\\`）。
- 远程模式差异：mudlib 已对齐本地规则——`repopulate_mode`（system/breed）、`land_privacy_enabled`、`rent_rate`、`move_cost`、代际/亲子继承、私有地与租金统计全部支持且语义一致；`remote_simulation` 从快照同步 `config`（反写经济参数显示）、`owners`（地主边框）、居民代际/亲子/土地/last_action 和租金统计。设置面板远程分支同样可编辑，`apply_settings` 把差异参数放入 `pending_config` 随 reset 下发。`E` 接管映射为 `POST /api/ai/attach` + `/api/citizens/command`（move/harvest/walk/stay 是纯原语：move 写坐标、walk 沿 exit 移动，采集与代谢由 mudlib 在世界节拍中统一结算）；`queue_manual_move` 在客户端先按占位拒绝（对齐本地一格一人）。服务端网格须为 50×50（GRID_SIZE 常量）。请求失败时 `remote_simulation.last_error` 记录错误，`world.gd` 把新错误弹到 `SystemMessageFeed`（红色警告条），持续断连时状态栏显示「连接失败 · 原因」。
- `remote_smoke.gd` 是远程模式冒烟测试，需 Go 服务已启动：`--script res://remote_smoke.gd -- --server=URL`。
- 字体：`gui/theme/custom_font` 是 `fonts/NotoSansSC-subset.otf`（pyftsubset 子集，覆盖 ASCII/常用标点/CJK 统一表意 U+4E00–9FFF）。`world.gd::_ready` 把根 Control 的 `theme.default_font` 设为 SystemFont（PingFang SC），所以 Control 主题链下全用系统字体；但挂在 Node2D 下的 Control（ThoughtBubble、TutorialGuide 的 Label/Button）主题链断裂，回落到项目默认字体即该子集——子集缺字会在那里显示为方框。新增文字后若出现缺字，用以下命令重建子集（源字体在 ~/Downloads/NotoSansSC-Regular.otf）：

```sh
pyftsubset ~/Downloads/NotoSansSC-Regular.otf \
  --output-file=godot/fonts/NotoSansSC-subset.otf \
  --unicodes="U+0020-007E,U+00A0-00FF,U+2000-206F,U+2190-21FF,U+2460-24FF,U+3000-303F,U+4E00-9FFF,U+FF00-FFEF" \
  --layout-features='*' --name-IDs='*' --notdef-outline \
  --recalc-bounds --drop-tables-=DSIG
```
- 提交项目、场景、脚本及 Godot 生成的 `.gd.uid`，不要提交 `.godot/` 缓存。已验证本机 Godot 4.7.2，其他机器先确认二进制位置与版本。

在本目录执行：

```sh
"/Applications/Godot.app/Contents/MacOS/Godot" --headless --path . --import
"/Applications/Godot.app/Contents/MacOS/Godot" --headless --path . --script res://simulation_test.gd
"/Applications/Godot.app/Contents/MacOS/Godot" --headless --path . --script res://ui_test.gd
"/Applications/Godot.app/Contents/MacOS/Godot" --headless --path . --script res://remote_smoke.gd -- --server=http://127.0.0.1:8080
"/Applications/Godot.app/Contents/MacOS/Godot" --path . --script res://ui_test.gd
"/Applications/Godot.app/Contents/MacOS/Godot" --path .
```

模型测试检查生长/采集/代谢、择地/占用、死亡/补入、已知 Gini、种子可重复性及长期资源预算。UI 测试通过视口输入验证暂停、单步、选择、死亡显示和重置；注入事件使用视口局部坐标，避免 headless 默认64×64窗口与1160×840逻辑视口的缩放差异。真实窗口测试可用 `-- --capture=/绝对路径/截图.png` 保存画面（先确保父目录存在、不会覆盖已有文件）。修改绘图后还需检查中文、资源颜色、财富配色、跨边界移动及窗口缩放；无头测试不能证明实际画面正常。
