extends SceneTree

var _model_script: GDScript
var _failures := 0
var _checks := 0
var _finished := false


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	if not ResourceLoader.exists("res://simulation.gd"):
		printerr("FAIL: 缺少本地经济模型 simulation.gd，无法执行资源生长、收割和代谢")
		quit(1)
		return
	_model_script = load("res://simulation.gd")
	if _model_script == null or not _model_script.can_instantiate():
		printerr("FAIL: 经济模型无法加载")
		quit(1)
		return
	create_timer(30.0).timeout.connect(_watchdog)
	_test_initialization()
	_test_growth_and_harvest()
	_test_vision_and_preferences()
	_test_occupancy_and_shuffle()
	_test_deaths_and_replacements()
	_test_gini()
	_test_determinism_and_reset()
	_test_long_run()
	_finished = true
	print("%s: %d checks, %d failures" % ["PASS" if _failures == 0 else "FAIL", _checks, _failures])
	quit(0 if _failures == 0 else 1)


func _watchdog() -> void:
	if not _finished:
		printerr("FAIL: 测试未完成，可能发生脚本运行错误")
		quit(1)


func _check(condition: bool, message: String) -> void:
	_checks += 1
	if not condition:
		_failures += 1
		printerr("FAIL: " + message)


func _near(actual: float, expected: float, message: String, tolerance: float = 0.0001) -> void:
	_check(absf(actual - expected) <= tolerance, "%s: actual=%s expected=%s" % [message, actual, expected])


func _fixture(count: int = 1, seed_value: int = 25):
	var model = _model_script.new()
	model.reset(seed_value)
	model.sugar.fill(0.0)
	model.capacities.fill(0.0)
	model.citizens.resize(count)
	for index in count:
		var citizen = model.citizens[index]
		citizen.cell = Vector2i(10 + index, 10)
		citizen.wealth = 20.0
		citizen.vision = 1
		citizen.metabolism = 1
		citizen.age = 0
		citizen.max_age = 100
	return model


func _resource(model, cell: Vector2i, capacity: float, amount: float = 0.0) -> void:
	var index := cell.y * 50 + cell.x
	model.capacities[index] = capacity
	model.sugar[index] = amount


func _test_initialization() -> void:
	var model = _model_script.new()
	model.reset()
	_check(model.GRID_SIZE == 50 and model.POPULATION == 80, "固定地图和人口常量")
	_check(model.citizens.size() == 80, "初始人口")
	_check(model.tick == 0 and model.starved == 0 and model.aged_out == 0 and model.replacements == 0, "初始计数清零且初始居民不计补入")
	_check(model.sugar.size() == 2500 and model.capacities.size() == 2500, "资源数组覆盖全图")
	for y in 50:
		for x in 50:
			var dx1 := absi(x - 15)
			var dy1 := absi(y - 15)
			var dx2 := absi(x - 35)
			var dy2 := absi(y - 35)
			var distance := mini(mini(dx1, 50 - dx1) + mini(dy1, 50 - dy1), mini(dx2, 50 - dx2) + mini(dy2, 50 - dy2))
			_near(model.capacities[y * 50 + x], maxf(0.0, 4.0 - 0.16 * distance), "双峰环面容量")
			_near(model.sugar[y * 50 + x], model.capacities[y * 50 + x], "初始资源充满")
	for index in 80:
		_check(model.citizens[index].id == index + 1, "初始递增 ID")
		_check(model.citizens[index].age == 0 and model.citizens[index].last_harvest == 0.0, "初始年龄和收割")
		_check(model.citizens[index].wealth >= 5.0 and model.citizens[index].wealth <= 25.0, "初始财富范围")
	_check_valid(model)
	var stats: Dictionary = model.get_stats()
	for key in ["population", "total_wealth", "mean_wealth", "gini", "land_sugar", "harvested", "consumed"]:
		_check(stats.has(key), "统计字段 " + key)
	for key in ["harvested", "consumed", "growth", "injected_wealth", "removed_wealth"]:
		_near(stats[key], 0.0, "初始预算 " + key)


func _test_growth_and_harvest() -> void:
	var model = _fixture()
	var citizen = model.citizens[0]
	citizen.metabolism = 2
	citizen.age = 7
	_resource(model, citizen.cell, 3.25, 2.5)
	_resource(model, Vector2i(30, 30), 4.0, 1.25)
	_resource(model, Vector2i(31, 30), 3.25, 3.0)
	_resource(model, Vector2i(32, 30), 4.0, 4.0)
	_step_with_budget(model)
	_near(model.sugar[30 * 50 + 30], 2.25, "每轮增长一单位")
	_near(model.sugar[30 * 50 + 31], 3.25, "生长受小数容量限制")
	_near(model.sugar[30 * 50 + 32], 4.0, "已满不增长")
	_near(model.sugar[0], 0.0, "零容量不增长")
	_near(citizen.last_harvest, 3.25, "先增长后采集且保留小数")
	_near(citizen.wealth, 21.25, "采集后扣除代谢")
	_near(model.sugar[10 * 50 + 10], 0.0, "收割全部糖")
	_check(citizen.age == 8 and citizen.cell == Vector2i(10, 10), "年龄增加一次且可以留在原地")
	_near(model.get_stats().harvested, 3.25, "上一轮总采集")
	_near(model.get_stats().consumed, 2.0, "新人当轮不代谢")


func _test_vision_and_preferences() -> void:
	var model = _fixture()
	var citizen = model.citizens[0]
	citizen.vision = 2
	_resource(model, Vector2i(11, 11), 4.0, 4.0)
	_resource(model, Vector2i(13, 10), 4.0, 4.0)
	_resource(model, Vector2i(10, 12), 3.0, 3.0)
	model.step()
	_check(citizen.cell == Vector2i(10, 12), "只看正交射线，排除对角及视野外")
	model = _fixture()
	citizen = model.citizens[0]
	citizen.vision = 6
	_resource(model, Vector2i(11, 10), 2.0, 2.0)
	_resource(model, Vector2i(16, 10), 3.5, 3.5)
	model.step()
	_check(citizen.cell == Vector2i(16, 10), "资源优先于距离且包含视野末端")
	model = _fixture()
	citizen = model.citizens[0]
	citizen.vision = 3
	_resource(model, Vector2i(11, 10), 3.0, 3.0)
	_resource(model, Vector2i(13, 10), 3.0, 3.0)
	model.step()
	_check(citizen.cell == Vector2i(11, 10), "资源相等取最近")
	model = _fixture()
	citizen = model.citizens[0]
	_resource(model, citizen.cell, 3.0, 3.0)
	_resource(model, Vector2i(11, 10), 3.0, 3.0)
	model.step()
	_check(citizen.cell == Vector2i(10, 10), "原地距离零优先")
	for pair in [[Vector2i(0, 10), Vector2i(49, 10)], [Vector2i(49, 10), Vector2i(0, 10)], [Vector2i(10, 0), Vector2i(10, 49)], [Vector2i(10, 49), Vector2i(10, 0)]]:
		model = _fixture()
		citizen = model.citizens[0]
		citizen.cell = pair[0]
		_resource(model, pair[1], 4.0, 4.0)
		model.step()
		_check(citizen.cell == pair[1], "四方向环面移动")
	var chosen := {}
	for seed_value in 32:
		model = _fixture(1, seed_value)
		citizen = model.citizens[0]
		for direction in [Vector2i.LEFT, Vector2i.RIGHT, Vector2i.UP, Vector2i.DOWN]:
			_resource(model, citizen.cell + direction, 4.0, 4.0)
		model.step()
		chosen[citizen.cell] = true
	_check(chosen.size() == 4, "等价候选由 seed 随机选择而非固定方向")


func _test_occupancy_and_shuffle() -> void:
	var winners := {}
	var follower_outcomes := {}
	for seed_value in 32:
		var model = _fixture(2, seed_value)
		var left = model.citizens[0]
		var right = model.citizens[1]
		left.cell = Vector2i(9, 10)
		right.cell = Vector2i(11, 10)
		_resource(model, Vector2i(10, 10), 4.0, 4.0)
		model.step()
		_check(left.cell != right.cell, "竞争资源后占用即时更新")
		_near(left.last_harvest + right.last_harvest, 4.0, "竞争格仅收割一次")
		winners[left.id if left.cell == Vector2i(10, 10) else right.id] = true
		model = _fixture(2, seed_value)
		left = model.citizens[0]
		right = model.citizens[1]
		_resource(model, Vector2i(11, 10), 3.0, 3.0)
		_resource(model, Vector2i(12, 10), 4.0, 4.0)
		model.step()
		_check(right.cell == Vector2i(12, 10), "先占用的居民可以离开")
		_check(left.cell in [Vector2i(10, 10), Vector2i(11, 10)], "禁止进入尚未腾空的占用格")
		follower_outcomes[left.cell] = true
	_check(winners.size() == 2, "居民执行顺序随 seed 洗牌")
	_check(follower_outcomes.size() == 2, "腾空格可被当轮后续居民使用")
	var model = _fixture(2)
	var seeker = model.citizens[0]
	var blocker = model.citizens[1]
	seeker.vision = 2
	blocker.cell = Vector2i(11, 10)
	_resource(model, blocker.cell, 4.0, 4.0)
	_resource(model, Vector2i(12, 10), 3.0, 3.0)
	model.step()
	_check(blocker.cell == Vector2i(11, 10) and seeker.cell == Vector2i(12, 10), "不可占用他人格子但视野不被居民遮挡")


func _test_deaths_and_replacements() -> void:
	var model = _fixture(3)
	var starving = model.citizens[0]
	var aging = model.citizens[1]
	var negative = model.citizens[2]
	starving.wealth = 1.0
	starving.age = 59
	starving.max_age = 60
	aging.age = 59
	aging.max_age = 60
	negative.wealth = 0.5
	negative.metabolism = 4
	_step_with_budget(model)
	_check(starving.death_reason == "starvation" and starving.age == 60, "零财富且到寿命时饥饿优先")
	_check(aging.death_reason == "age" and aging.age == 60, "达到寿命即老死")
	_check(negative.death_reason == "starvation", "负财富饥饿")
	_near(negative.wealth, -3.5, "死者保留真实负财富用于预算")
	_check(model.starved == 2 and model.aged_out == 1 and model.replacements == 80, "累计死亡及补入统计")
	_check(not model.citizens.has(starving) and not model.citizens.has(aging) and not model.citizens.has(negative), "移除所有死者但外部引用可报告原因")
	_near(model.get_stats().removed_wealth, 15.5, "移除财富包含负数")
	_near(model.get_stats().consumed, 6.0, "仅旧居民代谢")
	for citizen in model.citizens:
		_check(citizen.id >= 81, "补入 ID 不重用")
		_check(citizen.age == 0 and citizen.last_harvest == 0.0, "新人本轮不行动")
		_check(citizen.wealth >= 5.0 and citizen.wealth <= 25.0, "新人财富范围")
	_check_valid(model)
	for seed_value in 16:
		model = _fixture(2, seed_value)
		var survivor = model.citizens[0]
		var dying = model.citizens[1]
		dying.wealth = -10.0
		_resource(model, dying.cell, 4.0, 4.0)
		model.step()
		_check(survivor.cell == Vector2i(10, 10), "死者在全部行动结束之前仍占格")


func _test_gini() -> void:
	var model = _fixture(4)
	var distribution := [6.0, 0.0, 2.0, 0.0]
	for index in 4:
		model.citizens[index].wealth = distribution[index]
	var stats: Dictionary = model.get_stats()
	_near(stats.total_wealth, 8.0, "总财富")
	_near(stats.mean_wealth, 2.0, "平均财富")
	_near(stats.gini, 0.625, "已知非排序财富分布 Gini")
	_check(model.citizens[0].wealth == 6.0, "统计不重排居民")
	for citizen in model.citizens:
		citizen.wealth = 5.0
	_near(model.get_stats().gini, 0.0, "均等财富 Gini")
	for citizen in model.citizens:
		citizen.wealth = 0.0
	_near(model.get_stats().gini, 0.0, "零财富 Gini")
	model.citizens.clear()
	_near(model.get_stats().gini, 0.0, "空人口 Gini")
	_near(model.get_stats().mean_wealth, 0.0, "空人口均值")


func _snapshot(model) -> Array:
	var result: Array = [model.tick, model.starved, model.aged_out, model.replacements, model.sugar.duplicate(), model.capacities.duplicate(), model.get_stats()]
	for citizen in model.citizens:
		result.append([citizen.id, citizen.cell, citizen.wealth, citizen.vision, citizen.metabolism, citizen.age, citizen.max_age, citizen.last_harvest, citizen.death_reason])
	return result


func _test_determinism_and_reset() -> void:
	var first = _model_script.new()
	var second = _model_script.new()
	first.reset(913)
	second.reset(913)
	var initial := _snapshot(first)
	_check(initial == _snapshot(second), "同 seed 初始布局相同")
	for round_index in 160:
		seed(round_index)
		first.step()
		seed(round_index + 10000)
		second.step()
		_check(_snapshot(first) == _snapshot(second), "长期确定性不依赖全局 RNG，第 %d 轮" % round_index)
	first.reset(913)
	_check(_snapshot(first) == initial, "重置恢复 seed、ID、布局、统计和预算")
	second.reset(913)
	for round_index in 120:
		first.step()
		second.step()
		_check(_snapshot(first) == _snapshot(second), "重置后长期轨迹一致")
	first.reset()
	second.reset(25)
	_check(_snapshot(first) == _snapshot(second), "默认 seed 为 25")
	second.reset(26)
	_check(_snapshot(first) != _snapshot(second), "不同 seed 产生不同布局")


func _check_valid(model) -> void:
	_check(model.citizens.size() == 80, "始终保持人口 80")
	var occupied := {}
	var ids := {}
	for citizen in model.citizens:
		_check(citizen.cell.x >= 0 and citizen.cell.x < 50 and citizen.cell.y >= 0 and citizen.cell.y < 50, "坐标处于环面范围")
		_check(not occupied.has(citizen.cell), "一格一人")
		_check(not ids.has(citizen.id), "ID 唯一")
		occupied[citizen.cell] = true
		ids[citizen.id] = true
		_check(citizen.wealth > 0.0 and is_finite(citizen.wealth), "活人财富正且有限")
		_check(citizen.age >= 0 and citizen.age < citizen.max_age and citizen.death_reason.is_empty(), "活人年龄及死亡标记有效")
		_check(citizen.vision >= 1 and citizen.vision <= 6 and citizen.metabolism >= 1 and citizen.metabolism <= 4, "视野及代谢范围")
		_check(citizen.max_age >= 60 and citizen.max_age <= 100, "寿命范围")
	for index in 2500:
		_check(model.sugar[index] >= 0.0 and model.sugar[index] <= model.capacities[index] and is_finite(model.sugar[index]), "土地资源有效且不超过容量")


func _step_with_budget(model) -> void:
	var before: Dictionary = model.get_stats()
	var old_citizens: Array = model.citizens.duplicate()
	var expected_growth := 0.0
	var expected_consumed := 0.0
	for index in 2500:
		expected_growth += minf(model.capacities[index], model.sugar[index] + 1.0) - model.sugar[index]
	for citizen in old_citizens:
		expected_consumed += citizen.metabolism
	model.step()
	var after: Dictionary = model.get_stats()
	var injected := 0.0
	var removed := 0.0
	var harvested := 0.0
	for citizen in old_citizens:
		harvested += citizen.last_harvest
		if not model.citizens.has(citizen):
			removed += citizen.wealth
	for citizen in model.citizens:
		if not old_citizens.has(citizen):
			injected += citizen.wealth
	_near(after.growth, expected_growth, "独立计算增长量", 0.001)
	_near(after.consumed, expected_consumed, "独立计算代谢量")
	_near(after.harvested, harvested, "独立计算收割量")
	_near(after.injected_wealth, injected, "独立计算新人财富注入")
	_near(after.removed_wealth, removed, "独立计算死者财富移除")
	_near(after.land_sugar - before.land_sugar, expected_growth - harvested, "土地预算", 0.001)
	_near(after.total_wealth - before.total_wealth, harvested - expected_consumed + injected - removed, "居民预算", 0.001)
	_near(after.land_sugar + after.total_wealth - before.land_sugar - before.total_wealth, expected_growth - expected_consumed + injected - removed, "总资源及财富守恒预算", 0.001)


func _test_long_run() -> void:
	for seed_value in [25, 726]:
		var model = _model_script.new()
		model.reset(seed_value)
		var highest_id := 80
		var previous_ids := {}
		for citizen in model.citizens:
			previous_ids[citizen.id] = true
		for round_index in 200:
			_step_with_budget(model)
			_check_valid(model)
			_check(model.tick == round_index + 1, "tick 单步增长")
			_check(model.replacements == model.starved + model.aged_out, "正常人口下累计补入等于累计死亡")
			var next_ids := {}
			var next_highest := highest_id
			for citizen in model.citizens:
				if not previous_ids.has(citizen.id):
					_check(citizen.id > highest_id and citizen.age == 0 and citizen.last_harvest == 0.0, "长期新人 ID 递增且当轮不行动")
				next_highest = maxi(next_highest, citizen.id)
				next_ids[citizen.id] = true
			highest_id = next_highest
			previous_ids = next_ids
			var stats: Dictionary = model.get_stats()
			_check(stats.population == 80 and stats.gini >= 0.0 and stats.gini <= 1.0, "长期人口统计及 Gini 范围")
		_check(model.replacements > 0 and model.starved > 0 and model.aged_out > 0, "长期实际经历饥饿和寿终补入")
