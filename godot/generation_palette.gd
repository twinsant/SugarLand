extends HBoxContainer

signal swatch_clicked(generation: int)

const COLORS := [
	Color("e67e22"), Color("f1c40f"), Color("2ecc71"),
	Color("1abc9c"), Color("3498db"), Color("9b59b6"), Color("a188a6"),
	Color("607d8b"), Color("795548"), Color("00bcd4"),
]

var _swatches: Array[Panel] = []
var _count_labels: Array[Label] = []


func _ready() -> void:
	set_anchors_preset(Control.PRESET_BOTTOM_LEFT)
	custom_minimum_size = Vector2(COLORS.size() * 28.0, 46.0)
	mouse_filter = Control.MOUSE_FILTER_STOP
	for i in COLORS.size():
		var column := VBoxContainer.new()
		column.custom_minimum_size = Vector2(26, 0)
		column.mouse_filter = Control.MOUSE_FILTER_STOP
		column.add_theme_constant_override("separation", 2)
		var swatch := Panel.new()
		swatch.custom_minimum_size = Vector2(26, 26)
		swatch.mouse_filter = Control.MOUSE_FILTER_STOP
		swatch.tooltip_text = "第 %d 代 · #%02x%02x%02x" % [i, int(COLORS[i].r * 255), int(COLORS[i].g * 255), int(COLORS[i].b * 255)]
		var style := StyleBoxFlat.new()
		style.bg_color = COLORS[i]
		style.border_color = Color(0.1, 0.1, 0.1, 0.6)
		style.set_border_width_all(1)
		swatch.add_theme_stylebox_override("panel", style)
		swatch.gui_input.connect(_on_swatch_input.bind(i))
		column.add_child(swatch)
		_swatches.append(swatch)
		var count_label := Label.new()
		count_label.text = "0"
		count_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
		count_label.add_theme_font_size_override("font_size", 11)
		count_label.add_theme_color_override("font_color", Color(0.85, 0.9, 0.85, 1))
		column.add_child(count_label)
		_count_labels.append(count_label)
		add_child(column)


func update_counts(generation_counts: Array[int], generation_list: Array[Array]) -> void:
	for i in _count_labels.size():
		var count := 0
		if i < generation_counts.size():
			count = generation_counts[i]
		_count_labels[i].text = str(count)
		var hex := "#%02x%02x%02x" % [int(COLORS[i].r * 255), int(COLORS[i].g * 255), int(COLORS[i].b * 255)]
		if i < generation_list.size() and not generation_list[i].is_empty():
			var gens := generation_list[i]
			var gen_text := ", ".join(gens.map(func(g): return str(g)))
			_swatches[i].tooltip_text = "第 %s 代 · %s" % [gen_text, hex]
		else:
			_swatches[i].tooltip_text = "第 %d 代 · %s" % [i, hex]


func _on_swatch_input(event: InputEvent, generation: int) -> void:
	if event is InputEventMouseButton and event.pressed and event.button_index == MOUSE_BUTTON_LEFT:
		swatch_clicked.emit(generation)
