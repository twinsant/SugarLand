extends Control

const Simulation = preload("res://simulation.gd")
const Resident = preload("res://resident.gd")
const GRID_SIZE := Simulation.GRID_SIZE
const TILE_SIZE := 14.0
const MAP_ORIGIN := Vector2(32, 10)
const MAP_SIZE := GRID_SIZE * TILE_SIZE
const MAP_VIEW_RECT := Rect2(0, 86, 750, 710)
const MIN_ZOOM := 0.75
const MAX_ZOOM := 3.0
const ZOOM_STEP := 1.15
const FOLLOW_MARGIN_CELLS := 2.0
const FOLLOW_SMOOTHNESS := 8.0
const STEP_SECONDS := 0.9
const RECORDING_FPS := 30.0
const RECORDING_DIR := "user://recordings"

var simulation := Simulation.new()
var display_sugar := PackedFloat32Array()
var residents: Dictionary = {}
var selected_citizen: Simulation.Citizen
var selected_cell := Vector2i(-1, -1)
var running := true
var elapsed := 0.0
var recording := false
var recording_elapsed := 0.0
var recording_frame := 0
var recording_path := ""
var recording_result := ""
var panning := false
var following_selected := false
var map_pan_target := Vector2.ZERO

@onready var map_clip: Control = $MapClip
@onready var map_layer: Node2D = $MapClip/MapLayer
@onready var residents_root: Node2D = $MapClip/MapLayer/Residents
@onready var status: Label = %Status
@onready var economy: Label = %Economy
@onready var selection: Label = %Selection
@onready var pause_button: Button = %Pause
@onready var record_button: Button = %Record
@onready var recording_status: Label = %RecordingStatus
@onready var zoom_status: Label = %ZoomStatus


func _ready() -> void:
	var font := SystemFont.new()
	font.font_names = PackedStringArray(["PingFang SC", "Microsoft YaHei", "Noto Sans CJK SC"])
	theme = Theme.new()
	theme.default_font = font
	theme.default_font_size = 16
	pause_button.pressed.connect(toggle_pause)
	%Reset.pressed.connect(reset)
	%Step.pressed.connect(single_step)
	record_button.pressed.connect(toggle_recording)
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
	reset_map_view()
	sync_residents(false)
	set_running(true)
	update_map_layer()


func _process(delta: float) -> void:
	if running:
		elapsed += delta
		if elapsed >= STEP_SECONDS:
			elapsed = fmod(elapsed, STEP_SECONDS)
			advance()
	update_selected_follow(delta)
	if recording:
		recording_elapsed += delta
		while recording_elapsed >= 1.0 / RECORDING_FPS:
			recording_elapsed -= 1.0 / RECORDING_FPS
			capture_frame()


func advance() -> void:
	simulation.step()
	display_sugar = simulation.sugar.duplicate()
	for index in simulation.harvested_by_cell:
		display_sugar[index] = simulation.harvested_by_cell[index]
	sync_residents(running)
	if selected_citizen != null:
		selected_cell = selected_citizen.cell
	update_panel()
	update_map_layer()


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
			residents_root.add_child(resident)
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
		resident.set_thought(thought_for(citizen) if resident.selected else "")
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
		update_map_layer()


func update_map_layer() -> void:
	map_layer.set("display_sugar", display_sugar)
	map_layer.set("selected_cell", selected_cell)
	map_layer.queue_redraw()


func reset_map_view() -> void:
	map_layer.position = Vector2.ZERO
	map_layer.scale = Vector2.ONE
	map_pan_target = map_layer.position
	following_selected = false
	update_zoom_status()


func update_zoom_status() -> void:
	zoom_status.text = "地图缩放 %d%%" % roundi(map_layer.scale.x * 100.0)


func update_selected_follow(delta: float) -> void:
	if selected_citizen == null:
		return
	var resident: Resident = residents.get(selected_citizen.id)
	if resident == null:
		return
	var resident_global := map_layer.to_global(resident.position)
	var margin := FOLLOW_MARGIN_CELLS * TILE_SIZE * map_layer.scale.x
	var follow_rect := map_clip.get_global_rect().grow(-margin)
	if not following_selected and follow_rect.has_point(resident_global):
		return
	following_selected = true
	map_pan_target = map_clip.size * 0.5 - resident.position * map_layer.scale.x
	var blend := 1.0 - exp(-FOLLOW_SMOOTHNESS * delta)
	map_layer.position = map_layer.position.lerp(map_pan_target, blend)


func zoom_at(global_point: Vector2, factor: float) -> void:
	var before := map_layer.to_local(global_point)
	var next_zoom := clampf(map_layer.scale.x * factor, MIN_ZOOM, MAX_ZOOM)
	if is_equal_approx(next_zoom, map_layer.scale.x):
		return
	map_layer.scale = Vector2.ONE * next_zoom
	update_zoom_status()
	var after := map_layer.to_local(global_point)
	map_layer.position += (after - before) * next_zoom


func toggle_recording() -> void:
	recording = not recording
	if recording:
		var session_name := "session_%d" % Time.get_ticks_msec()
		recording_path = "%s/%s" % [RECORDING_DIR, session_name]
		DirAccess.make_dir_recursive_absolute(ProjectSettings.globalize_path(recording_path))
		recording_elapsed = 0.0
		recording_frame = 0
		recording_result = ""
		record_button.text = "停止录像 [V]"
		capture_frame()
	else:
		record_button.text = "开始录像 [V]"
		finish_recording()
	update_recording_status()


func finish_recording() -> void:
	if recording_frame == 0:
		recording_result = "录像结束：没有可用帧"
		return
	var ffmpeg_path := find_ffmpeg()
	var input_pattern := ProjectSettings.globalize_path(recording_path) + "/frame_" + "%06d" + ".png"
	var output_path := ProjectSettings.globalize_path(recording_path + "/sugarland.mp4")
	var arguments := PackedStringArray([
		"-y", "-framerate", str(RECORDING_FPS), "-i", input_pattern,
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2",
		"-c:v", "libx264", "-pix_fmt", "yuv420p", "-movflags", "+faststart", output_path,
	])
	var output: Array[String] = []
	var exit_code := OS.execute(ffmpeg_path, arguments, output, true)
	if exit_code == 0:
		recording_result = "录像完成：%d 帧\n%s" % [recording_frame, recording_path + "/sugarland.mp4"]
	else:
		recording_result = "转码失败（%d），PNG 已保留\n%s" % [exit_code, recording_path]


func find_ffmpeg() -> String:
	for candidate in ["/opt/homebrew/bin/ffmpeg", "/usr/local/bin/ffmpeg"]:
		if FileAccess.file_exists(candidate):
			return candidate
	return "ffmpeg"


func capture_frame() -> void:
	if not recording or DisplayServer.get_name() == "headless":
		return
	var viewport_texture := get_viewport().get_texture()
	if viewport_texture == null:
		update_recording_status()
		return
	var image := viewport_texture.get_image()
	if image == null:
		update_recording_status()
		return
	var frame_path := "%s/frame_%06d.png" % [recording_path, recording_frame]
	if image.save_png(frame_path) == OK:
		recording_frame += 1
	update_recording_status()


func update_recording_status() -> void:
	if recording:
		recording_status.text = "录像中 · %d 帧\n%s" % [recording_frame, recording_path]
	elif not recording_result.is_empty():
		recording_status.text = recording_result
	else:
		recording_status.text = "录像：未开始"


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
	if event is InputEventKey and event.keycode in [KEY_SPACE, KEY_R, KEY_N, KEY_V, KEY_0]:
		get_viewport().set_input_as_handled()
		if event.pressed and not event.echo:
			match event.keycode:
				KEY_SPACE:
					toggle_pause()
				KEY_R:
					reset()
				KEY_N:
					single_step()
				KEY_V:
					toggle_recording()
				KEY_0:
					reset_map_view()


func _unhandled_input(event: InputEvent) -> void:
	if event is InputEventMagnifyGesture:
		if map_clip.get_global_rect().has_point(event.position):
			zoom_at(event.position, event.factor)
			get_viewport().set_input_as_handled()
		return
	if event is InputEventPanGesture:
		pan_map(-event.delta)
		get_viewport().set_input_as_handled()
		return
	if event is InputEventMouseButton:
		if event.button_index in [MOUSE_BUTTON_MIDDLE, MOUSE_BUTTON_RIGHT]:
			panning = event.pressed
			get_viewport().set_input_as_handled()
			return
		if event.button_index in [MOUSE_BUTTON_WHEEL_UP, MOUSE_BUTTON_WHEEL_DOWN] and event.pressed:
			if map_clip.get_global_rect().has_point(event.position):
				var factor := ZOOM_STEP if event.button_index == MOUSE_BUTTON_WHEEL_UP else 1.0 / ZOOM_STEP
				zoom_at(event.position, factor)
				get_viewport().set_input_as_handled()
			return
		if event.button_index == MOUSE_BUTTON_LEFT and event.pressed:
			select_at(event.position)
			return
	if event is InputEventMouseMotion and panning:
		pan_map(event.relative)
		get_viewport().set_input_as_handled()


func pan_map(delta: Vector2) -> void:
	map_layer.position += delta
	map_pan_target = map_layer.position
	following_selected = false


func thought_for(citizen: Simulation.Citizen) -> String:
	var current_sugar := simulation.sugar[citizen.cell.y * GRID_SIZE + citizen.cell.x]
	var best_sugar := current_sugar
	for direction: Vector2i in Simulation.DIRECTIONS:
		for distance in range(1, citizen.vision + 1):
			var raw := citizen.cell + direction * distance
			var cell := Vector2i(posmod(raw.x, GRID_SIZE), posmod(raw.y, GRID_SIZE))
			best_sugar = maxf(best_sugar, simulation.sugar[cell.y * GRID_SIZE + cell.x])
	if citizen.wealth <= citizen.metabolism * 2:
		return "我需要糖！"
	if best_sugar > current_sugar + 0.1:
		return "去找更肥的糖田"
	if current_sugar > 0.1:
		return "这里还有糖，先采集"
	return "附近没有糖，继续寻找"


func select_at(global_point: Vector2) -> void:
	var point := map_layer.to_local(global_point)
	if not Rect2(MAP_ORIGIN, Vector2.ONE * MAP_SIZE).has_point(point):
		return
	selected_citizen = null
	following_selected = false
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
	for citizen in simulation.citizens:
		var resident: Resident = residents[citizen.id]
		resident.selected = selected_citizen != null and resident.citizen_id == selected_citizen.id
		resident.set_thought(thought_for(citizen) if resident.selected else "")
		resident.queue_redraw()
	update_panel()
	update_map_layer()


func update_panel() -> void:
	var stats := simulation.get_stats()
	status.text = "糖域纪年 %03d 年  /  %s" % [simulation.tick, "运行中" if running else "已暂停"]
	update_recording_status()
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


func cell_center(cell: Vector2i) -> Vector2:
	return MAP_ORIGIN + (Vector2(cell) + Vector2.ONE * 0.5) * TILE_SIZE
