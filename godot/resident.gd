extends Node2D

const ThoughtBubble = preload("res://thought_bubble.gd")

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
var bubble: ThoughtBubble
var tutorial_bubble := false
var tutorial_show_vision := false
var map_bounds: Rect2


func _ready() -> void:
	bubble = ThoughtBubble.new()
	bubble.name = "ThoughtBubbleRoot"
	add_child(bubble)


func set_bubble_viewport(rect: Rect2) -> void:
	if bubble != null:
		bubble.set_viewport(rect)


func set_thought(text: String) -> void:
	tutorial_bubble = false
	thought_text = text
	z_index = 200 if selected else 0
	if bubble != null:
		if selected and not text.is_empty():
			bubble.show_thought(text)
		else:
			bubble.hide_bubble()
	queue_redraw()


func set_tutorial_thought(title: String, text: String) -> void:
	tutorial_bubble = true
	thought_text = text
	selected = false
	z_index = 0
	if bubble != null:
		bubble.show_tutorial(title, text)
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
	if bubble != null:
		bubble.update_placement()
	queue_redraw()


func _draw() -> void:
	var stride := sin(travel * TAU * 2.0) * 1.8 if travel < 1.0 else 0.0
	if selected or tutorial_show_vision:
		var radius := (vision + 0.5) * tile_size
		draw_arc(Vector2.ZERO, radius, 0.0, TAU, 48, Color(0.27, 0.78, 0.91, 0.55), 1.4, true)
		draw_arc(Vector2.ZERO, radius, 0.0, TAU, 48, Color(0.45, 0.88, 1.0, 0.12), 0.8, true)
	draw_set_transform(Vector2(0, 3), 0.0, Vector2(1.0, 0.4))
	draw_circle(Vector2.ZERO, 5.0, Color(0, 0, 0, 0.3))
	draw_set_transform(Vector2.ZERO)
	draw_line(Vector2(-1.5, 0), Vector2(-2.5 - stride, 4), Color("152a31"), 2.0, true)
	draw_line(Vector2(1.5, 0), Vector2(2.5 + stride, 4), Color("152a31"), 2.0, true)
	draw_circle(Vector2(0, -1.5 - absf(stride) * 0.35), 3.5, tint)
	draw_circle(Vector2(0, -5.0 - absf(stride) * 0.35), 2.5, Color("ffe2b8"))
