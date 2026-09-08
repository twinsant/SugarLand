extends Control

const Simulation = preload("res://simulation.gd")
const Resident = preload("res://resident.gd")
const GRID_SIZE := Simulation.GRID_SIZE
const TILE_SIZE := 14.0
const MAP_ORIGIN := Vector2(32, 96)
const MAP_SIZE := GRID_SIZE * TILE_SIZE
const STEP_SECONDS := 0.9

var simulation := Simulation.new()
var display_sugar := PackedFloat32Array()
var residents: Dictionary = {}
var selected_citizen: Simulation.Citizen
var selected_cell := Vector2i(-1, -1)
var running := true
var elapsed := 0.0

@onready var status: Label = %Status
@onready var economy: Label = %Economy
@onready var selection: Label = %Selection
@onready var pause_button: Button = %Pause


func _ready() -> void:
	var font := SystemFont.new()
	font.font_names = PackedStringArray(["PingFang SC", "Microsoft YaHei", "Noto Sans CJK SC"])
	theme = Theme.new()
	theme.default_font = font
	theme.default_font_size = 16
	pause_button.pressed.connect(toggle_pause)
	%Reset.pressed.connect(reset)
	%Step.pressed.connect(single_step)
	reset()


func reset() -> void:
	selected_citizen = null
	selected_cell = Vector2i(-1, -1)
	for resident: Resident in residents.values():
		resident.free()
	residents.clear()
	simulation.reset()
	display_sugar = simulation.sugar.duplicate()
	elapsed = 0.0
	sync_residents(false)
	set_running(true)
	queue_redraw()


func _process(delta: float) -> void:
	if not running:
		return
	elapsed += delta
	if elapsed >= STEP_SECONDS:
		elapsed = fmod(elapsed, STEP_SECONDS)
		advance()


func advance() -> void:
	simulation.step()
	display_sugar = simulation.sugar.duplicate()
	for index in simulation.harvested_by_cell:
		display_sugar[index] = simulation.harvested_by_cell[index]
	sync_residents(running)
	if selected_citizen != null:
		selected_cell = selected_citizen.cell
	update_panel()
	queue_redraw()


func sync_residents(animate: bool) -> void:
	var alive := {}
	for citizen in simulation.citizens:
		alive[citizen.id] = true
		var resident: Resident = residents.get(citizen.id)
		if resident == null:
			resident = Resident.new()
			resident.citizen_id = citizen.id
			resident.map_bounds = Rect2(MAP_ORIGIN, Vector2.ONE * MAP_SIZE)
			resident.harvest_ready.connect(_on_harvest_ready)
			$Residents.add_child(resident)
			residents[citizen.id] = resident
			resident.place_at(citizen.cell, cell_center(citizen.cell), citizen.last_harvest)
		elif animate:
			resident.walk_to(citizen.cell, cell_center(citizen.cell), citizen.last_harvest)
		else:
			resident.place_at(citizen.cell, cell_center(citizen.cell), citizen.last_harvest)
		resident.tint = Color("f8f4e8").lerp(Color("d4af37"), clampf(citizen.wealth / 40.0, 0.0, 1.0))
		resident.vision = citizen.vision
		resident.tile_size = TILE_SIZE
		resident.selected = citizen == selected_citizen
		resident.set_process(running)
		resident.queue_redraw()
	for citizen_id: int in residents.keys():
		if not alive.has(citizen_id):
			residents[citizen_id].free()
			residents.erase(citizen_id)


func _on_harvest_ready(cell: Vector2i) -> void:
	var index := cell.y * GRID_SIZE + cell.x
	if index >= 0 and index < simulation.sugar.size():
		display_sugar[index] = simulation.sugar[index]
		queue_redraw()


func toggle_pause() -> void:
	set_running(not running)


func single_step() -> void:
	set_running(false)
	elapsed = 0.0
	advance()


func set_running(value: bool) -> void:
	running = value
	for resident: Resident in residents.values():
		resident.set_process(running)
	pause_button.text = "暂停 [空格]" if running else "继续 [空格]"
	update_panel()


func _input(event: InputEvent) -> void:
	if event is InputEventKey and event.keycode in [KEY_SPACE, KEY_R, KEY_N]:
		get_viewport().set_input_as_handled()
		if event.pressed and not event.echo:
			match event.keycode:
				KEY_SPACE:
					toggle_pause()
				KEY_R:
					reset()
				KEY_N:
					single_step()


func _unhandled_input(event: InputEvent) -> void:
	if event is InputEventMouseButton and event.button_index == MOUSE_BUTTON_LEFT and event.pressed:
		var point: Vector2 = make_input_local(event).position
		if not Rect2(MAP_ORIGIN, Vector2.ONE * MAP_SIZE).has_point(point):
			return
		selected_citizen = null
		var nearest := 10.0
		for citizen in simulation.citizens:
			var resident: Resident = residents[citizen.id]
			var distance := resident.position.distance_to(point)
			if distance < nearest:
				nearest = distance
				selected_citizen = citizen
		selected_cell = Vector2i((point - MAP_ORIGIN) / TILE_SIZE)
		if selected_citizen != null:
			selected_cell = selected_citizen.cell
		for resident: Resident in residents.values():
			resident.selected = selected_citizen != null and resident.citizen_id == selected_citizen.id
			resident.queue_redraw()
		update_panel()
		queue_redraw()


func update_panel() -> void:
	var stats := simulation.get_stats()
	status.text = "第 %04d 轮  /  %s" % [simulation.tick, "运行中" if running else "已暂停"]
	economy.text = "人口 %d · 累计补入 %d\n总财富 %.1f · 人均 %.1f\n财富基尼 %.3f · 地图存糖 %.1f\n本轮采集 %.1f · 代谢 %.1f\n累计死亡：饥饿 %d / 寿终 %d" % [
		stats.population, simulation.replacements, stats.total_wealth, stats.mean_wealth,
		stats.gini, stats.land_sugar, stats.harvested, stats.consumed,
		simulation.starved, simulation.aged_out,
	]
	if selected_cell.x < 0:
		selection.text = "点击居民或格子查看详情\n\n居民在视野内寻找最多的糖，\n采集所得用于支付每轮代谢。"
		return
	var index := selected_cell.y * GRID_SIZE + selected_cell.x
	var cell_info := "坐标 (%d, %d) · 糖 %.2f / %.2f" % [
		selected_cell.x, selected_cell.y, simulation.sugar[index], simulation.capacities[index],
	]
	if selected_citizen == null:
		selection.text = "资源格子\n%s\n每轮再生 1，直至达到容量。" % cell_info
	elif not selected_citizen.death_reason.is_empty():
		var reason := "饥饿" if selected_citizen.death_reason == "starvation" else "寿终"
		selection.text = "居民 #%03d · 已%s\n年龄 %d · 最后财富 %.2f\n新居民已补入，可重新选择。" % [
			selected_citizen.id, reason, selected_citizen.age, selected_citizen.wealth,
		]
	else:
		selection.text = "居民 #%03d\n%s\n财富 %.2f · 本轮采集 %.2f\n代谢 %d / 轮 · 视野 %d 格\n年龄 %d / %d 轮" % [
			selected_citizen.id, cell_info, selected_citizen.wealth, selected_citizen.last_harvest,
			selected_citizen.metabolism, selected_citizen.vision, selected_citizen.age, selected_citizen.max_age,
		]


func _draw() -> void:
	if simulation.sugar.is_empty():
		return
	draw_rect(Rect2(MAP_ORIGIN - Vector2.ONE * 2, Vector2.ONE * (MAP_SIZE + 4)), Color("53685e"), false, 2.0)
	for y in GRID_SIZE:
		for x in GRID_SIZE:
			var richness := display_sugar[y * GRID_SIZE + x] / 4.0
			var color := Color("d3b86e").lerp(Color("245b3c"), richness)
			draw_rect(Rect2(MAP_ORIGIN + Vector2(x, y) * TILE_SIZE, Vector2.ONE * (TILE_SIZE - 0.7)), color)
	if selected_cell.x >= 0:
		draw_rect(Rect2(MAP_ORIGIN + Vector2(selected_cell) * TILE_SIZE, Vector2.ONE * TILE_SIZE), Color("fff6dc"), false, 2.0)


func cell_center(cell: Vector2i) -> Vector2:
	return MAP_ORIGIN + (Vector2(cell) + Vector2.ONE * 0.5) * TILE_SIZE
