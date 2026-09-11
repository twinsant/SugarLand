# SugarLand Godot 本地仿真

- 独立 Godot 4 项目，使用 GDScript 和 Compatibility 渲染器，无插件、外部素材或后端请求。启动时默认使用全屏窗口；启动先显示主菜单，开始游戏后才运行仿真；保持精简，不将 Go/React 启动流程作为前置条件。
- `simulation.gd` 是无场景依赖的单资源经济模型；`world.gd` 连接模型和界面；`main.tscn` 定义界面；`resident.gd` 只绘制和插值移动，不修改经济状态。
- 默认 50×50 环面、80 人、种子 25。每轮先糖再生（+1 至容量），再用模型自有 RNG 随机排序居民。居民在四条正交视线内选择未占用格：糖最多、距离最近、平局随机；允许原地停留，每格只容纳一人。
- 采集保留小数，随后扣代谢并增加年龄；财富 <=0 判饥饿，否则年龄达到寿命判寿终。全部居民行动后移除死者并补足80人；新人带启动财富，当轮不行动。这是开放资源模型，不是人口自然繁殖或封闭财富守恒模型，没有交易、继承或双资源。
- `get_stats()` 的 growth、harvested、consumed、injected_wealth、removed_wealth、released_wealth 均为上一轮预算；removed_wealth 是带符号死者剩余财富，可为负，released_wealth 是其中实际转回地块且受容量限制的正财富。累计死亡和补入单独记录。Gini 只统计当前存活居民。
- 经济回合不依赖动画帧率。空格暂停两者；N 或单步按钮执行一轮、立即显示目的地并保持暂停；R 或重置按钮恢复同一种子和初始统计并开始运行。选中居民死亡后保留原因展示，不转移选择到新居民。
- `simulation.sugar` 是即时经济状态，`world.gd` 的 `display_sugar` 是动画显示状态：每轮先反映生长，已采集地块暂时保留采集前的显示值，居民到达后由 `harvest_ready` 信号切换为采集后的值。这样地块不会在居民移动途中提前变色。
- 地图显示位于可裁剪的 `MapClip/MapLayer` 下；`MapLayer` 与 `Residents` 一起缩放，Sidebar 不缩放。滚轮和触控板捏合缩放，范围为 75%–300%；触控板双指平移反转系统手势 delta，并使用 32 倍 macOS PanGesture 灵敏度补偿，使地图内容跟随手指方向，鼠标中/右键拖动保持直接跟手，`0` 恢复视图。点击先将全局坐标转换到 MapLayer 本地坐标，再计算格子。选中居民接近地图视口边缘两格或移出视口时，MapLayer 以指数平滑移动，让居民回到视口中心。
- 选中居民的 `ThoughtBubble` 显示基于财富、当前糖量和视野内资源的本地行动想法；它是表现层提示，不参与居民决策。气泡接近地图视口上边界时自动翻到居民下方，三角尖端同步翻转。
- 每轮死亡事件记录实际格子，`DeathMarker` 在该位置绘制骷髅，显示 2.5 秒后渐隐并释放；重置时清理所有死亡标记。
- 选中居民时自动开启 `F` 控制的径向视野迷雾：视野圆圈内清晰，边缘外逐渐变黑；再次按 `F` 可关闭，点击空地也会关闭。选中居民和 ThoughtBubble 位于迷雾上层。`E` 切换选中居民的用户接管，方向键提交上下左右移动；接管期间强制开启迷雾，释放后恢复普通迷雾开关。选中居民死亡后按财富最高、财富最低交替选择新的居民。
- `intro.json` 保存游戏介绍页文本；`tutorial_guide.gd` 挂在唯一真实居民节点上提供分页和按钮，并复用该居民的 `ThoughtBubble` 显示教程文本。游戏启动直接进入世界内教程，教程只显示一个居民并暂停仿真，完成教程后才开始当前游戏逻辑。
- 游戏内录像由 `V` 或录像按钮控制，固定 30 FPS 将视口保存为 PNG 序列到 `/tmp/sugarland/session_<timestamp>/`，暂停仿真时仍可录制。PNG 编码和写盘在后台线程执行，避免阻塞主线程。停止录像时自动查找 ffmpeg，先裁剪到偶数宽高，再生成 `~/.sugarland/recordings/session_<timestamp>/sugarland.mp4`；找不到 ffmpeg 或转码失败时保留 PNG。
- 提交项目、场景、脚本及 Godot 生成的 `.gd.uid`，不要提交 `.godot/` 缓存。已验证本机 Godot 4.7.2，其他机器先确认二进制位置与版本。

在本目录执行：

```sh
"/Applications/Godot.app/Contents/MacOS/Godot" --headless --path . --import
"/Applications/Godot.app/Contents/MacOS/Godot" --headless --path . --script res://simulation_test.gd
"/Applications/Godot.app/Contents/MacOS/Godot" --headless --path . --script res://ui_test.gd
"/Applications/Godot.app/Contents/MacOS/Godot" --path . --script res://ui_test.gd
"/Applications/Godot.app/Contents/MacOS/Godot" --path .
```

模型测试检查生长/采集/代谢、择地/占用、死亡/补入、已知 Gini、种子可重复性及长期资源预算。UI 测试通过视口输入验证暂停、单步、选择、死亡显示和重置；注入事件使用视口局部坐标，避免 headless 默认64×64窗口与1160×840逻辑视口的缩放差异。真实窗口测试可用 `-- --capture=/绝对路径/截图.png` 保存画面（先确保父目录存在、不会覆盖已有文件）。修改绘图后还需检查中文、资源颜色、财富配色、跨边界移动及窗口缩放；无头测试不能证明实际画面正常。
