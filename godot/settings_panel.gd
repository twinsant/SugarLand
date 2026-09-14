extends Control

signal closed

var world

var _panel: PanelContainer
var _growth_rate: SpinBox
var _peak_a_x: SpinBox
var _peak_a_y: SpinBox
var _peak_b_x: SpinBox
var _peak_b_y: SpinBox
var _peak_capacity: SpinBox
var _peak_falloff: SpinBox
var _step_seconds: SpinBox
var _repopulate_mode: OptionButton
var _breed_age_min: SpinBox
var _breed_age_max: SpinBox
var _move_cost: SpinBox


func _ready() -> void:
	anchor_right = 1.0
	anchor_bottom = 1.0
	offset_right = 0.0
	offset_bottom = 0.0
	z_index = 2000
	mouse_filter = Control.MOUSE_FILTER_STOP
	_build_ui()
	_load_values()


func _build_ui() -> void:
	var center := CenterContainer.new()
	center.mouse_filter = Control.MOUSE_FILTER_IGNORE
	center.anchor_right = 1.0
	center.anchor_bottom = 1.0
	center.z_index = 2001
	add_child(center)
	_panel = PanelContainer.new()
	_panel.custom_minimum_size = Vector2(440, 0)
	var style := StyleBoxFlat.new()
	style.bg_color = Color(0.05, 0.1, 0.12, 0.98)
	style.border_color = Color("315d64")
	style.set_border_width_all(1)
	style.set_corner_radius_all(14)
	style.content_margin_left = 24
	style.content_margin_right = 24
	style.content_margin_top = 20
	style.content_margin_bottom = 20
	_panel.add_theme_stylebox_override("panel", style)
	center.add_child(_panel)
	var vbox := VBoxContainer.new()
	vbox.add_theme_constant_override("separation", 12)
	_panel.add_child(vbox)
	var title := Label.new()
	title.text = "经济参数设置"
	title.add_theme_font_size_override("font_size", 22)
	title.add_theme_color_override("font_color", Color("dff5e6"))
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	vbox.add_child(title)
	var hint := Label.new()
	hint.text = "修改地形参数后会重置仿真以重建地图"
	hint.add_theme_font_size_override("font_size", 13)
	hint.add_theme_color_override("font_color", Color("9bb9bd"))
	hint.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	vbox.add_child(hint)
	_growth_rate = _add_row(vbox, "糖再生速率", 0.0, 3.0, 0.1)
	_peak_capacity = _add_row(vbox, "双峰容量", 1.0, 8.0, 0.1)
	_peak_falloff = _add_row(vbox, "容量衰减", 0.05, 0.5, 0.01)
	_peak_a_x = _add_row(vbox, "糖峰 A 横坐标", 0, 49, 1)
	_peak_a_y = _add_row(vbox, "糖峰 A 纵坐标", 0, 49, 1)
	_peak_b_x = _add_row(vbox, "糖峰 B 横坐标", 0, 49, 1)
	_peak_b_y = _add_row(vbox, "糖峰 B 纵坐标", 0, 49, 1)
	_step_seconds = _add_row(vbox, "步进间隔（秒）", 0.1, 3.0, 0.1)
	_repopulate_mode = _add_option_row(vbox, "人口补充方式", ["系统补人", "繁殖"], ["system", "breed"])
	_breed_age_min = _add_row(vbox, "繁殖年龄下限", 0, 100, 1)
	_breed_age_max = _add_row(vbox, "繁殖年龄上限", 0, 100, 1)
	_move_cost = _add_row(vbox, "移动消耗系数", 0.0, 2.0, 0.1)
	var buttons := HBoxContainer.new()
	buttons.add_theme_constant_override("separation", 12)
	buttons.alignment = BoxContainer.ALIGNMENT_CENTER
	vbox.add_child(buttons)
	var apply_button := Button.new()
	apply_button.name = "Apply"
	apply_button.text = "应用并重置"
	apply_button.custom_minimum_size = Vector2(150, 44)
	apply_button.pressed.connect(_apply)
	buttons.add_child(apply_button)
	var cancel_button := Button.new()
	cancel_button.name = "Cancel"
	cancel_button.text = "取消"
	cancel_button.custom_minimum_size = Vector2(150, 44)
	cancel_button.pressed.connect(_close)
	buttons.add_child(cancel_button)


func _add_row(parent: Control, text: String, min_value: float, max_value: float, step: float) -> SpinBox:
	var row := HBoxContainer.new()
	row.add_theme_constant_override("separation", 12)
	parent.add_child(row)
	var label := Label.new()
	label.text = text
	label.add_theme_font_size_override("font_size", 15)
	label.add_theme_color_override("font_color", Color("cfe6e8"))
	label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(label)
	var spin := SpinBox.new()
	spin.min_value = min_value
	spin.max_value = max_value
	spin.step = step
	spin.allow_greater = false
	spin.allow_lesser = false
	spin.custom_minimum_size = Vector2(120, 0)
	row.add_child(spin)
	return spin


func _add_option_row(parent: Control, text: String, labels: Array, values: Array) -> OptionButton:
	var row := HBoxContainer.new()
	row.add_theme_constant_override("separation", 12)
	parent.add_child(row)
	var label := Label.new()
	label.text = text
	label.add_theme_font_size_override("font_size", 15)
	label.add_theme_color_override("font_color", Color("cfe6e8"))
	label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(label)
	var option := OptionButton.new()
	for index in labels.size():
		option.add_item(labels[index], index)
		option.set_item_metadata(index, values[index])
	option.custom_minimum_size = Vector2(120, 0)
	row.add_child(option)
	return option


func _load_values() -> void:
	var sim = world.simulation
	_growth_rate.value = sim.growth_rate
	_peak_a_x.value = sim.peak_a.x
	_peak_a_y.value = sim.peak_a.y
	_peak_b_x.value = sim.peak_b.x
	_peak_b_y.value = sim.peak_b.y
	_peak_capacity.value = sim.peak_capacity
	_peak_falloff.value = sim.peak_falloff
	_step_seconds.value = world.step_seconds
	_breed_age_min.value = sim.breed_age_min
	_breed_age_max.value = sim.breed_age_max
	_move_cost.value = sim.move_cost
	for index in _repopulate_mode.item_count:
		if _repopulate_mode.get_item_metadata(index) == sim.repopulate_mode:
			_repopulate_mode.select(index)
			break


func _apply() -> void:
	var sim = world.simulation
	sim.growth_rate = _growth_rate.value
	sim.peak_a = Vector2i(int(_peak_a_x.value), int(_peak_a_y.value))
	sim.peak_b = Vector2i(int(_peak_b_x.value), int(_peak_b_y.value))
	sim.peak_capacity = _peak_capacity.value
	sim.peak_falloff = _peak_falloff.value
	sim.repopulate_mode = _repopulate_mode.get_selected_metadata()
	sim.breed_age_min = int(_breed_age_min.value)
	sim.breed_age_max = int(_breed_age_max.value)
	sim.move_cost = _move_cost.value
	world.step_seconds = _step_seconds.value
	world.apply_settings()
	_close()


func _close() -> void:
	closed.emit()
	queue_free()


func _draw() -> void:
	draw_rect(Rect2(Vector2.ZERO, size), Color(0.0, 0.0, 0.0, 0.55))


func _input(event: InputEvent) -> void:
	if event is InputEventKey and event.pressed and event.keycode == KEY_ESCAPE:
		_close()
		get_viewport().set_input_as_handled()
