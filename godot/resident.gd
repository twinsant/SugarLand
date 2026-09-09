extends Node2D

signal harvest_ready(cell: Vector2i)

var citizen_id: int
var cell: Vector2i
var tint: Color
var selected := false
var vision := 1
var tile_size := 14.0
var start := Vector2.ZERO
var target := Vector2.ZERO
var travel := 1.0
var pending_harvest := 0.0
var thought_text := ""
var thought_label: Label
var map_bounds: Rect2


func _ready() -> void:
	thought_label = Label.new()
	thought_label.name = "ThoughtBubble"
	thought_label.z_index = 1000
	thought_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	thought_label.position = Vector2(-58, -70)
	thought_label.custom_minimum_size = Vector2(116, 34)
	thought_label.size = Vector2(116, 34)
	thought_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	thought_label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	thought_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	thought_label.add_theme_color_override("font_color", Color(0.09, 0.2, 0.23, 0.92))
	thought_label.add_theme_font_size_override("font_size", 12)
	var bubble_style := StyleBoxFlat.new()
	bubble_style.bg_color = Color(0.93, 0.98, 1.0, 0.72)
	bubble_style.border_color = Color(0.27, 0.78, 0.91, 0.9)
	bubble_style.set_border_width_all(1)
	bubble_style.set_corner_radius_all(8)
	bubble_style.content_margin_left = 6
	bubble_style.content_margin_right = 6
	thought_label.add_theme_stylebox_override("normal", bubble_style)
	add_child(thought_label)


func set_thought(text: String) -> void:
	thought_text = text
	z_index = 50 if selected else 0
	if thought_label != null:
		thought_label.text = thought_text
		thought_label.visible = selected and not thought_text.is_empty()
		queue_redraw()


func place_at(next_cell: Vector2i, destination: Vector2, harvested := 0.0) -> void:
	cell = next_cell
	position = destination
	start = destination
	target = destination
	travel = 1.0
	pending_harvest = 0.0
	if harvested > 0.0:
		harvest_ready.emit(cell)
	queue_redraw()


func walk_to(next_cell: Vector2i, destination: Vector2, harvested := 0.0) -> void:
	cell = next_cell
	start = position
	target = destination
	var offset := target - start
	if absf(offset.x) > map_bounds.size.x * 0.5:
		target.x -= signf(offset.x) * map_bounds.size.x
	if absf(offset.y) > map_bounds.size.y * 0.5:
		target.y -= signf(offset.y) * map_bounds.size.y
	travel = 1.0 if start.is_equal_approx(target) else 0.0
	pending_harvest = harvested if travel < 1.0 else 0.0
	if travel >= 1.0 and harvested > 0.0:
		harvest_ready.emit(cell)


func _process(delta: float) -> void:
	var previous_travel := travel
	travel = minf(travel + delta / 0.65, 1.0)
	if pending_harvest > 0.0 and previous_travel < 1.0 and travel >= 1.0:
		harvest_ready.emit(cell)
		pending_harvest = 0.0
	var point := start.lerp(target, smoothstep(0.0, 1.0, travel))
	position = map_bounds.position + Vector2(
		fposmod(point.x - map_bounds.position.x, map_bounds.size.x),
		fposmod(point.y - map_bounds.position.y, map_bounds.size.y),
	)
	queue_redraw()


func _draw() -> void:
	var stride := sin(travel * TAU * 2.0) * 1.8 if travel < 1.0 else 0.0
	if selected:
		var radius := (vision + 0.5) * tile_size
		draw_arc(Vector2.ZERO, radius, 0.0, TAU, 48, Color("45c7e8"), 1.4, true)
		draw_arc(Vector2.ZERO, radius, 0.0, TAU, 48, Color(0.45, 0.88, 1.0, 0.28), 0.8, true)
	if selected and not thought_text.is_empty():
		draw_colored_polygon(PackedVector2Array([
			Vector2(-6, -38), Vector2(0, -30), Vector2(7, -39),
		]), Color(0.93, 0.98, 1.0, 0.72))
	draw_set_transform(Vector2(0, 3), 0.0, Vector2(1.0, 0.4))
	draw_circle(Vector2.ZERO, 5.0, Color(0, 0, 0, 0.3))
	draw_set_transform(Vector2.ZERO)
	draw_line(Vector2(-1.5, 0), Vector2(-2.5 - stride, 4), Color("152a31"), 2.0, true)
	draw_line(Vector2(1.5, 0), Vector2(2.5 + stride, 4), Color("152a31"), 2.0, true)
	draw_circle(Vector2(0, -1.5 - absf(stride) * 0.35), 3.5, tint)
	draw_circle(Vector2(0, -5.0 - absf(stride) * 0.35), 2.5, Color("ffe2b8"))
