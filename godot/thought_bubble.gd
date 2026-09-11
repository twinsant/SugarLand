extends Node2D

const BUBBLE_HEIGHT := 46.0
const BUBBLE_ABOVE_Y := -82.0
const TITLE_HEIGHT := 18.0
const TAIL_HEIGHT := 14.0

const BG_COLOR := Color(0.93, 0.98, 1.0, 0.72)
const BORDER_COLOR := Color(0.27, 0.78, 0.91, 0.9)
const TITLE_BG_COLOR := Color(0.08, 0.22, 0.32, 0.92)
const FONT_COLOR := Color(0.09, 0.2, 0.23, 0.92)
const TITLE_FONT_COLOR := Color(1.0, 1.0, 1.0, 0.95)

var body_label: Label
var title_label: Label
var viewport_rect := Rect2()
var flipped := false
var has_title := false
var width := 116.0
var padding := 6.0
var font_size := 12


func _ready() -> void:
	body_label = Label.new()
	body_label.name = "ThoughtBubble"
	body_label.z_index = 1000
	body_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	body_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	body_label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	body_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	body_label.add_theme_color_override("font_color", FONT_COLOR)
	body_label.add_theme_font_size_override("font_size", font_size)
	add_child(body_label)
	title_label = Label.new()
	title_label.name = "BubbleTitle"
	title_label.z_index = 1000
	title_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	title_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	title_label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	title_label.add_theme_color_override("font_color", TITLE_FONT_COLOR)
	title_label.add_theme_font_size_override("font_size", font_size)
	add_child(title_label)
	_apply_layout()


func set_viewport(rect: Rect2) -> void:
	viewport_rect = rect
	update_placement()


func update_placement() -> void:
	if body_label == null or not body_label.visible or not viewport_rect.has_area():
		return
	var scale_y := get_global_transform().get_scale().y
	var top_y: float = (BUBBLE_ABOVE_Y if not has_title else BUBBLE_ABOVE_Y) * scale_y
	var should_flip := global_position.y + top_y < viewport_rect.position.y
	if should_flip == flipped:
		return
	flipped = should_flip
	_apply_layout()


func show_thought(text: String) -> void:
	has_title = false
	width = 116.0
	padding = 6.0
	font_size = 12
	if body_label != null:
		body_label.text = text
		body_label.visible = not text.is_empty()
	if title_label != null:
		title_label.visible = false
	_apply_layout()


func show_tutorial(title: String, text: String) -> void:
	has_title = not title.is_empty()
	width = 198.0
	padding = 18.0
	font_size = 12
	if body_label != null:
		body_label.text = text
		body_label.visible = not text.is_empty()
	if title_label != null:
		title_label.text = title
		title_label.visible = has_title
	_apply_layout()


func hide_bubble() -> void:
	if body_label != null:
		body_label.visible = false
	if title_label != null:
		title_label.visible = false


func get_bubble_bottom() -> float:
	if body_label == null:
		return 0.0
	return body_label.position.y + body_label.size.y


func _apply_layout() -> void:
	if body_label == null or title_label == null:
		return
	body_label.add_theme_font_size_override("font_size", font_size)
	title_label.add_theme_font_size_override("font_size", font_size)
	var half_w := width * 0.5
	if has_title:
		title_label.position = Vector2(-half_w, BUBBLE_ABOVE_Y)
		title_label.size = Vector2(width, TITLE_HEIGHT)
		title_label.custom_minimum_size = title_label.size
		body_label.position = Vector2(-half_w, BUBBLE_ABOVE_Y + TITLE_HEIGHT)
		body_label.size = Vector2(width, BUBBLE_HEIGHT)
		body_label.custom_minimum_size = body_label.size
		_apply_title_style()
		_apply_body_style(false, true)
	else:
		body_label.position = Vector2(-half_w, BUBBLE_ABOVE_Y)
		body_label.size = Vector2(width, BUBBLE_HEIGHT)
		body_label.custom_minimum_size = body_label.size
		_apply_body_style(true)
		title_label.visible = false


func _apply_body_style(all_rounded: bool = false, bottom_rounded: bool = false) -> void:
	var style := StyleBoxFlat.new()
	style.bg_color = BG_COLOR
	style.border_color = BORDER_COLOR
	style.set_border_width_all(1)
	if all_rounded:
		style.border_width_bottom = 0
		style.set_corner_radius_all(8)
	elif bottom_rounded:
		style.border_width_top = 0
		style.corner_radius_bottom_left = 8
		style.corner_radius_bottom_right = 8
	style.content_margin_left = padding
	style.content_margin_right = padding
	body_label.add_theme_stylebox_override("normal", style)


func _apply_title_style() -> void:
	var style := StyleBoxFlat.new()
	style.bg_color = TITLE_BG_COLOR
	style.corner_radius_top_left = 8
	style.corner_radius_top_right = 8
	style.content_margin_left = padding
	style.content_margin_right = padding
	title_label.add_theme_stylebox_override("normal", style)


func _draw() -> void:
	if body_label == null or not body_label.visible:
		return
	var bubble_bottom: float = get_bubble_bottom()
	var tail := PackedVector2Array([
		Vector2(-7, bubble_bottom - TAIL_HEIGHT), Vector2(0, bubble_bottom), Vector2(8, bubble_bottom - TAIL_HEIGHT),
	]) if flipped else PackedVector2Array([
		Vector2(-7, bubble_bottom), Vector2(0, bubble_bottom + TAIL_HEIGHT), Vector2(8, bubble_bottom),
	])
	draw_colored_polygon(tail, BG_COLOR)
