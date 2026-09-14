extends Node2D

const GROW_TIME := 0.8
const HEART_COLOR := Color(0.95, 0.32, 0.42, 0.9)
const LINK_COLOR := Color(0.95, 0.32, 0.42, 0.3)

var parent_a: Node2D
var parent_b: Node2D
var elapsed := 0.0


func _ready() -> void:
	z_index = 0
	set_process(true)


func _process(delta: float) -> void:
	elapsed += delta
	if elapsed >= GROW_TIME:
		queue_free()
		return
	queue_redraw()


func _draw() -> void:
	if parent_a == null or parent_b == null:
		return
	var pa := parent_a.position
	var pb := _unwrap_b(pa, parent_b.position)
	var mid := (pa + pb) * 0.5
	var t := minf(elapsed / GROW_TIME, 1.0)
	var scale_factor := 0.3 + t * 0.9 if t < 1.0 else 1.0
	draw_line(pa, pb, LINK_COLOR, 1.2, true)
	draw_set_transform(mid, 0.0, Vector2.ONE * scale_factor)
	draw_colored_polygon(PackedVector2Array([
		Vector2(0, 5), Vector2(-6, -1), Vector2(-6, -4), Vector2(-3, -6), Vector2(0, -4),
		Vector2(3, -6), Vector2(6, -4), Vector2(6, -1),
	]), HEART_COLOR)
	draw_circle(Vector2(-2.5, -3), 1.5, Color(1.0, 1.0, 1.0, 0.5))
	draw_set_transform(Vector2.ZERO)


func _unwrap_b(pa: Vector2, pb: Vector2) -> Vector2:
	var offset := pb - pa
	if absf(offset.x) > 350.0:
		pb.x -= signf(offset.x) * 700.0
	if absf(offset.y) > 350.0:
		pb.y -= signf(offset.y) * 700.0
	return pb
