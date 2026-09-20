extends "res://simulation.gd"

const BackendClient = preload("res://backend_client.gd")

var _client: BackendClient
var _by_id := {}
var _scoreboard := {}
var _world_agent := false


func _init(url := "") -> void:
	is_remote = true
	server_url = url if not url.is_empty() else "http://127.0.0.1:8080"
	_client = BackendClient.new(server_url)


func reset(_seed_value: int = 25) -> void:
	var body := {}
	if not pending_config.is_empty():
		body = {"config": pending_config}
		pending_config = {}
	var err := ""
	if _client.post_json("/api/world/reset", body) == null:
		err = _client.last_error
	_by_id.clear()
	bred = 0
	peak_gini = 0.0
	peak_income_gini = 0.0
	_sync()
	if not err.is_empty() and connected:
		last_error = err


func step() -> void:
	var err := ""
	if _client.post_json("/api/world/step", {}) == null:
		err = _client.last_error
	_sync()
	if not err.is_empty() and connected:
		last_error = err


func get_stats() -> Dictionary:
	var gini: float = _scoreboard.get("gini_coefficient", 0.0)
	var income_gini := float(_scoreboard.get("income_gini", 0.0))
	peak_gini = maxf(peak_gini, gini)
	peak_income_gini = maxf(peak_income_gini, income_gini)
	return {
		"population": citizens.size(),
		"total_wealth": float(_scoreboard.get("avg_wealth", 0.0)) * citizens.size(),
		"mean_wealth": float(_scoreboard.get("avg_wealth", 0.0)),
		"gini": gini,
		"income_gini": income_gini,
		"land_sugar": float(_scoreboard.get("total_sugar", 0.0)),
		"harvested": float(_scoreboard.get("harvested", 0.0)),
		"consumed": float(_scoreboard.get("consumed", 0.0)),
		"move_cost": float(_scoreboard.get("move_cost", 0.0)),
		"rent_paid": float(_scoreboard.get("rent_paid", 0.0)),
		"growth": float(_scoreboard.get("growth", 0.0)),
		"injected_wealth": float(_scoreboard.get("injected_wealth", 0.0)),
		"removed_wealth": float(_scoreboard.get("removed_wealth", 0.0)),
		"released_wealth": float(_scoreboard.get("released_wealth", 0.0)),
		"bred": bred,
		"births": breeding_events.size(),
	}


func get_land_owner(cell: Vector2i) -> int:
	return int(land_owners.get(cell.y * GRID_SIZE + cell.x, -1))


func set_manual_control(citizen_id: int, enabled: bool) -> void:
	if enabled:
		var res: Variant = _client.post_json("/api/ai/attach", {"citizen_id": citizen_id, "agent_name": "godot-user"})
		if res != null:
			manual_controlled[citizen_id] = true
		else:
			last_error = _client.last_error
	else:
		var res: Variant = _client.post_json("/api/ai/detach", {"citizen_id": citizen_id})
		if res == null:
			last_error = _client.last_error
		manual_controlled.erase(citizen_id)
		manual_commands.erase(citizen_id)


func queue_manual_move(citizen_id: int, direction: Vector2i) -> void:
	var citizen: Citizen = _by_id.get(citizen_id)
	if citizen == null or not manual_controlled.has(citizen_id):
		return
	var target := Vector2i(
		posmod(citizen.cell.x + direction.x, GRID_SIZE),
		posmod(citizen.cell.y + direction.y, GRID_SIZE))
	# 与本地 _choose_manual_cell 一致：目标格被占则原地不动
	for other: Citizen in _by_id.values():
		if other.id != citizen_id and other.cell == target:
			return
	var res: Variant = _client.post_json("/api/citizens/command", {
		"citizen_id": citizen_id, "action": "move", "args": {"x": target.x, "y": target.y}})
	if res != null:
		_sync()
	else:
		last_error = _client.last_error


func attach_lpc(citizen_id: int, source: String, interval: int = 1) -> String:
	var res: Variant = _client.post_json("/api/agent/attach", {
		"citizen_id": citizen_id, "lpc_source": source, "interval": interval})
	if res == null:
		return _client.last_error
	_sync()
	return ""


func detach_lpc(citizen_id: int) -> String:
	var res: Variant = _client.post_json("/api/agent/detach", {"citizen_id": citizen_id})
	if res == null:
		return _client.last_error
	_sync()
	return ""


func attach_world_lpc(source: String) -> String:
	var res: Variant = _client.post_json("/api/world/agent", {"lpc_source": source})
	if res == null:
		return _client.last_error
	_world_agent = true
	return ""


func detach_world_lpc() -> String:
	var res: Variant = _client.post_json("/api/world/agent/detach", {})
	if res == null:
		return _client.last_error
	_world_agent = false
	return ""


func world_agent_attached() -> bool:
	return _world_agent


# 返回当前生效的世界 LPC 源码（默认 mudlib 或用户自定义规则）
func get_world_lpc_source() -> String:
	var res: Variant = _client.get_json("/api/world/agent")
	if res == null or typeof(res) != TYPE_DICTIONARY:
		return ""
	return str(res.get("source", ""))


func _sync() -> void:
	var snap: Variant = _client.get_json("/api/snapshot")
	if snap == null or not (snap is Dictionary):
		connected = false
		last_error = _client.last_error
		return
	connected = true
	last_error = ""
	tick = int(snap.get("timestep", 0))
	sugar = PackedFloat32Array(snap.get("sugar", []))
	capacities = PackedFloat32Array(snap.get("capacity", []))
	_world_agent = bool(snap.get("world_agent", false))

	# 服务端配置回灌到本地字段（设置面板据此显示当前值）
	var cfg: Dictionary = snap.get("config", {})
	growth_rate = float(cfg.get("growth_rate", growth_rate))
	peak_a = Vector2i(int(cfg.get("peak_x1", peak_a.x)), int(cfg.get("peak_y1", peak_a.y)))
	peak_b = Vector2i(int(cfg.get("peak_x2", peak_b.x)), int(cfg.get("peak_y2", peak_b.y)))
	peak_capacity = float(cfg.get("peak_capacity", peak_capacity))
	peak_falloff = float(cfg.get("peak_falloff", peak_falloff))
	repopulate_mode = str(cfg.get("repopulate_mode", repopulate_mode))
	breed_age_min = int(cfg.get("breed_age_min", breed_age_min))
	breed_age_max = int(cfg.get("breed_age_max", breed_age_max))
	move_cost = float(cfg.get("move_cost", move_cost))
	var lp: Variant = cfg.get("land_privacy_enabled", false)
	land_privacy_enabled = (lp is bool and lp) or (lp is float and lp != 0.0) or (lp is int and lp != 0)
	rent_rate = float(cfg.get("rent_rate", rent_rate))

	# 土地私有制：room 属性 owner 平铺为数组（0=无主），转为 index→地主
	land_owners.clear()
	var owners: Array = snap.get("owner", [])
	for index in owners.size():
		var oid := int(owners[index])
		if oid != 0:
			land_owners[index] = oid

	var harvests: Dictionary = snap.get("harvests", {})
	harvested_by_cell.clear()
	for key in harvests:
		harvested_by_cell[int(key)] = float(harvests[key])

	var list: Array[Citizen] = []
	for item in snap.get("citizens", []):
		var citizen: Citizen = _by_id.get(int(item["id"]))
		if citizen == null:
			citizen = Citizen.new()
			citizen.id = int(item["id"])
			_by_id[citizen.id] = citizen
		# 服务端坐标可能越界（GetCell 查询时才做环面包裹），客户端归一化
		citizen.cell = Vector2i(posmod(int(item["x"]), GRID_SIZE), posmod(int(item["y"]), GRID_SIZE))
		citizen.wealth = float(item["wealth"])
		citizen.vision = int(item["vision"])
		citizen.metabolism = int(item["metabolism"])
		citizen.age = int(item["age"])
		citizen.max_age = int(item["max_age"])
		citizen.death_reason = ""
		citizen.agent_name = str(item.get("agent_name", ""))
		citizen.is_agent_controlled = bool(item.get("is_agent_controlled", false))
		citizen.has_lpc = bool(item.get("has_lpc", false))
		citizen.generation = int(item.get("generation", 0))
		citizen.last_action = str(item.get("last_action", ""))
		citizen.last_rent = float(item.get("last_rent", 0.0))
		citizen.total_harvest = float(item.get("total_harvest", 0.0))
		citizen.parents.clear()
		for p in item.get("parents", []):
			citizen.parents.append(int(p))
		citizen.children.clear()
		for ch in item.get("children", []):
			citizen.children.append(int(ch))
		var index := citizen.cell.y * GRID_SIZE + citizen.cell.x
		citizen.last_harvest = float(harvested_by_cell.get(index, 0.0))
		list.append(citizen)
	citizens = list

	death_events.clear()
	for d in snap.get("deaths", []):
		var id := int(d["id"])
		var reason := "age" if str(d.get("reason")) == "age" else "starvation"
		var generation := 0
		if _by_id.has(id):
			var dead: Citizen = _by_id[id]
			dead.death_reason = reason
			dead.age = int(d["age"])
			dead.wealth = float(d["wealth"])
			generation = dead.generation
		death_events.append({
			"id": id, "cell": Vector2i(posmod(int(d["x"]), GRID_SIZE), posmod(int(d["y"]), GRID_SIZE)), "reason": reason,
			"age": int(d["age"]), "wealth": float(d["wealth"]), "generation": generation,
		})

	breeding_events.clear()
	for b in snap.get("births", []):
		breeding_events.append({
			"parent_a": int(b["parent_a_id"]),
			"parent_b": int(b["parent_b_id"]),
			"child": int(b["child_id"]),
		})
	bred += breeding_events.size()

	_scoreboard = snap.get("scoreboard", {})
	_scoreboard["harvested"] = snap.get("harvested", 0.0)
	_scoreboard["consumed"] = snap.get("consumed", 0)
	starved = int(_scoreboard.get("death_by_hunger", 0))
	aged_out = int(_scoreboard.get("death_by_age", 0))
	replacements = int(_scoreboard.get("birth_count", 0)) - bred


func _gini(values: Array[float]) -> float:
	if values.is_empty():
		return 0.0
	var sorted := values.duplicate()
	sorted.sort()
	var weighted := 0.0
	var total := 0.0
	for index in sorted.size():
		weighted += (2.0 * (index + 1) - sorted.size() - 1.0) * sorted[index]
		total += sorted[index]
	if total <= 0.0:
		return 0.0
	return weighted / (sorted.size() * total)
