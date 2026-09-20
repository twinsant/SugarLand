extends Control

signal closed

const CITIZEN_SAMPLE := """void heart_beat() {
    int x = query_x();
    int y = query_y();
    int best_x = x;
    int best_y = y;
    float best_sugar = query_cell_here("sugar");

    float s = query_cell(x, y - 1, "sugar");
    if (s > best_sugar) { best_sugar = s; best_x = x; best_y = y - 1; }
    s = query_cell(x, y + 1, "sugar");
    if (s > best_sugar) { best_sugar = s; best_x = x; best_y = y + 1; }
    s = query_cell(x - 1, y, "sugar");
    if (s > best_sugar) { best_sugar = s; best_x = x - 1; best_y = y; }
    s = query_cell(x + 1, y, "sugar");
    if (s > best_sugar) { best_sugar = s; best_x = x + 1; best_y = y; }

    if (best_sugar == 0.0) {
        int dir = random(4);
        if (dir == 0) { move(x, y - 1); }
        else if (dir == 1) { move(x, y + 1); }
        else if (dir == 2) { move(x - 1, y); }
        else { move(x + 1, y); }
    } else {
        move(best_x, best_y);
    }
    float got = take_cell(query_x(), query_y(), "sugar");
    set_attr("wealth", query_attr("wealth") + to_int(got));
}
"""

const WORLD_SAMPLE := """void heart_beat() {
    // 自定义世界脚本整体替换仿真规则（G/M/Trade/Mate/R 均由本脚本编排）
    // 参考默认 mudlib：grow_all() move_all() do_trading() do_mating() reap()
}
"""

var world
var _panel: PanelContainer
var _target: OptionButton
var _interval: SpinBox
var _source: TextEdit
var _status: Label
var _attach_button: Button
var _detach_button: Button


func _ready() -> void:
	anchor_right = 1.0
	anchor_bottom = 1.0
	offset_right = 0.0
	offset_bottom = 0.0
	z_index = 2000
	mouse_filter = Control.MOUSE_FILTER_STOP
	_build_ui()
	_on_target_changed(0)


func _build_ui() -> void:
	var center := CenterContainer.new()
	center.mouse_filter = Control.MOUSE_FILTER_IGNORE
	center.anchor_right = 1.0
	center.anchor_bottom = 1.0
	center.z_index = 2001
	add_child(center)
	_panel = PanelContainer.new()
	_panel.custom_minimum_size = Vector2(520, 0)
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
	vbox.add_theme_constant_override("separation", 10)
	_panel.add_child(vbox)
	var title := Label.new()
	title.text = "LPC 脚本"
	title.add_theme_font_size_override("font_size", 22)
	title.add_theme_color_override("font_color", Color("dff5e6"))
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	vbox.add_child(title)
	var hint := Label.new()
	hint.text = "脚本在 Go 服务中执行：heart_beat() 每步调用一次"
	hint.add_theme_font_size_override("font_size", 13)
	hint.add_theme_color_override("font_color", Color("9bb9bd"))
	hint.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	vbox.add_child(hint)

	var target_row := HBoxContainer.new()
	target_row.add_theme_constant_override("separation", 12)
	vbox.add_child(target_row)
	var target_label := Label.new()
	target_label.text = "目标"
	target_label.add_theme_font_size_override("font_size", 15)
	target_label.add_theme_color_override("font_color", Color("cfe6e8"))
	target_label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	target_row.add_child(target_label)
	_target = OptionButton.new()
	_target.add_item("选中居民", 0)
	_target.set_item_metadata(0, "citizen")
	_target.add_item("世界规则 (mudlib)", 1)
	_target.set_item_metadata(1, "world")
	_target.custom_minimum_size = Vector2(160, 0)
	_target.item_selected.connect(_on_target_changed)
	target_row.add_child(_target)

	var interval_row := HBoxContainer.new()
	interval_row.add_theme_constant_override("separation", 12)
	vbox.add_child(interval_row)
	var interval_label := Label.new()
	interval_label.text = "调用间隔（步）"
	interval_label.add_theme_font_size_override("font_size", 15)
	interval_label.add_theme_color_override("font_color", Color("cfe6e8"))
	interval_label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	interval_row.add_child(interval_label)
	_interval = SpinBox.new()
	_interval.min_value = 1
	_interval.max_value = 100
	_interval.step = 1
	_interval.value = 1
	_interval.custom_minimum_size = Vector2(160, 0)
	interval_row.add_child(_interval)

	_source = TextEdit.new()
	_source.custom_minimum_size = Vector2(0, 260)
	_source.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_source.add_theme_font_size_override("font_size", 13)
	vbox.add_child(_source)

	_status = Label.new()
	_status.text = ""
	_status.add_theme_font_size_override("font_size", 13)
	_status.add_theme_color_override("font_color", Color("9bb9bd"))
	_status.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	vbox.add_child(_status)

	var buttons := HBoxContainer.new()
	buttons.add_theme_constant_override("separation", 12)
	buttons.alignment = BoxContainer.ALIGNMENT_CENTER
	vbox.add_child(buttons)
	_attach_button = Button.new()
	_attach_button.text = "挂接"
	_attach_button.custom_minimum_size = Vector2(110, 40)
	_attach_button.pressed.connect(_attach)
	buttons.add_child(_attach_button)
	_detach_button = Button.new()
	_detach_button.text = "卸载"
	_detach_button.custom_minimum_size = Vector2(110, 40)
	_detach_button.pressed.connect(_detach)
	buttons.add_child(_detach_button)
	var close_button := Button.new()
	close_button.text = "关闭"
	close_button.custom_minimum_size = Vector2(110, 40)
	close_button.pressed.connect(_close)
	buttons.add_child(close_button)


func _on_target_changed(_index: int) -> void:
	if _target == null:
		return
	if _target.get_selected_metadata() == "world":
		# 世界脚本承载全部仿真规则：拉取当前生效源码供编辑
		var live: String = world.simulation.get_world_lpc_source()
		_source.text = live if not live.is_empty() else WORLD_SAMPLE
		_interval.editable = false
		_detach_button.visible = true
	else:
		_source.text = CITIZEN_SAMPLE
		_interval.editable = true
		_detach_button.visible = true
	_update_target_status()


func _update_target_status() -> void:
	if _target.get_selected_metadata() == "world":
		_status.text = "世界规则：%s" % ("自定义 LPC 脚本" if world.simulation.world_agent_attached() else "默认 mudlib")
	else:
		if world.selected_citizen == null:
			_status.text = "未选中居民：先点击地图上的居民"
		elif world.selected_citizen.has_lpc:
			_status.text = "居民 #%03d：已挂接 LPC 脚本" % world.selected_citizen.id
		else:
			_status.text = "居民 #%03d：未挂接脚本" % world.selected_citizen.id


func _attach() -> void:
	var err := ""
	if _target.get_selected_metadata() == "world":
		err = world.simulation.attach_world_lpc(_source.text)
	else:
		if world.selected_citizen == null:
			_status.text = "未选中居民：先点击地图上的居民"
			return
		err = world.simulation.attach_lpc(world.selected_citizen.id, _source.text, int(_interval.value))
	_status.text = err if not err.is_empty() else "挂接成功"
	if err.is_empty():
		_update_target_status()
		world.update_panel()


func _detach() -> void:
	var err := ""
	if _target.get_selected_metadata() == "world":
		err = world.simulation.detach_world_lpc()
	else:
		if world.selected_citizen == null:
			_status.text = "未选中居民：先点击地图上的居民"
			return
		err = world.simulation.detach_lpc(world.selected_citizen.id)
	_status.text = err if not err.is_empty() else "已卸载"
	if err.is_empty():
		_update_target_status()
		world.update_panel()


func _close() -> void:
	closed.emit()
	queue_free()


func _draw() -> void:
	draw_rect(Rect2(Vector2.ZERO, size), Color(0.0, 0.0, 0.0, 0.55))


func _input(event: InputEvent) -> void:
	if event is InputEventKey and event.pressed and event.keycode == KEY_ESCAPE:
		_close()
		get_viewport().set_input_as_handled()
