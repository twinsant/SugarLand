extends SceneTree

const RemoteSimulation = preload("res://remote_simulation.gd")

const CITIZEN_LPC := """void heart_beat() {
    int x = query_x();
    int y = query_y();
    move(x + 1, y);
    float got = take_cell(x + 1, y, "sugar");
    set_attr("wealth", query_attr("wealth") + to_int(got));
}
"""

# 自定义世界脚本整体替换仿真规则；空 heart_beat 意味着居民不再行动
const WORLD_LPC := """void heart_beat() {
}
"""

var failures := 0


func check(condition: bool, label: String) -> void:
	if condition:
		print("PASS ", label)
	else:
		failures += 1
		print("FAIL ", label)


func _init() -> void:
	var url := "http://127.0.0.1:8080"
	var args := OS.get_cmdline_user_args()
	for i in args.size():
		if args[i] == "--server" and i + 1 < args.size():
			url = args[i + 1]
		elif args[i].begins_with("--server="):
			url = args[i].substr("--server=".length())

	var sim := RemoteSimulation.new(url)
	sim.reset()
	check(sim.connected, "连接服务 %s" % url)
	if not sim.connected:
		print("错误: ", sim.last_error)
		quit(1)
		return
	check(sim.sugar.size() == 2500, "糖量数组 2500 格")
	check(sim.capacities.size() == 2500, "容量数组 2500 格")
	check(not sim.citizens.is_empty(), "居民列表非空 (%d)" % sim.citizens.size())

	var stats := sim.get_stats()
	check(stats["population"] == sim.citizens.size(), "统计人口一致")

	# 默认 mudlib 驱动：推进 2 步，居民应有采集与代谢记账
	var first = sim.citizens[0]
	var tick_before := sim.tick
	for i in 2:
		sim.step()
	check(sim.tick == tick_before + 2, "mudlib 推进 2 步 (%d -> %d)" % [tick_before, sim.tick])
	stats = sim.get_stats()
	check(stats.get("consumed", 0) > 0, "mudlib 代谢记账 (consumed=%s)" % stats.get("consumed", 0))
	check(first.age >= 2, "mudlib 衰老记账 (age=%d)" % first.age)

	# 居民 LPC 由世界脚本 call_citizen 驱动
	# 注意：mudlib 每步有死亡+补员，attach 目标须从推进后的新鲜列表里挑
	var target = null
	var err := "no citizen"
	for cand in sim.citizens:
		err = sim.attach_lpc(cand.id, CITIZEN_LPC)
		if err.is_empty():
			target = cand
			break
	check(target != null, "挂接居民 LPC 脚本%s" % ("" if target != null else "（%s）" % err))
	var x_before: int = target.cell.x
	sim.step()
	check(target.cell.x != x_before, "居民脚本驱动移动 (x %d -> %d)" % [x_before, target.cell.x])
	check(target.has_lpc, "居民 has_lpc 标记")

	# 自定义世界脚本替换全部规则：空 heart_beat → 居民冻结
	err = sim.attach_world_lpc(WORLD_LPC)
	check(err.is_empty(), "挂接世界 LPC 脚本%s" % ("" if err.is_empty() else "（%s）" % err))
	sim.step()
	stats = sim.get_stats()
	check(stats.get("consumed", -1) == 0, "自定义世界脚本接管规则（居民冻结，consumed=%s）" % stats.get("consumed", -1))
	check(sim.world_agent_attached(), "世界自定义脚本已挂接")

	# 卸载恢复默认 mudlib，规则恢复（用 consumed 判定，不依赖单个居民存活）
	err = sim.detach_world_lpc()
	check(err.is_empty(), "卸载世界 LPC 脚本")
	sim.step()
	stats = sim.get_stats()
	check(stats.get("consumed", 0) > 0, "卸载后默认 mudlib 恢复规则")
	check(not sim.world_agent_attached(), "世界脚本恢复默认")

	# 目标可能在恢复 mudlib 后的步骤中死亡；必要时换新鲜居民验证 detach
	#（attach 后立即 detach，中间不推进世界，无死亡窗口）
	if not sim.citizens.has(target):
		for cand in sim.citizens:
			if sim.attach_lpc(cand.id, CITIZEN_LPC).is_empty():
				target = cand
				break
	err = sim.detach_lpc(target.id)
	check(err.is_empty(), "卸载居民 LPC 脚本")
	check(sim.citizens.has(target), "被挂接居民仍存活")

	if failures == 0:
		print("全部通过")
		quit(0)
	else:
		print("失败 %d 项" % failures)
		quit(1)
