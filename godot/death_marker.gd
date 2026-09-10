extends Node2D

const LIFETIME := 2.5

var remaining := LIFETIME


func _ready() -> void:
	z_index = 20
	scale = Vector2.ONE * 0.5
	queue_redraw()


func _process(delta: float) -> void:
	remaining -= delta
	if remaining <= 0.0:
		queue_free()
		return
	queue_redraw()


func _draw() -> void:
	var alpha := clampf(remaining / LIFETIME, 0.0, 1.0)
	var skull := Color(0.92, 0.94, 0.9, alpha * 0.9)
	var shadow := Color(0.05, 0.09, 0.1, alpha * 0.8)
	draw_circle(Vector2(0, -3), 7.0, skull)
	draw_rect(Rect2(-6, 1, 12, 7), skull)
	draw_circle(Vector2(-2.7, -4), 1.8, shadow)
	draw_circle(Vector2(2.7, -4), 1.8, shadow)
	draw_colored_polygon(PackedVector2Array([
		Vector2(-2, 2), Vector2(0, 0), Vector2(2, 2),
	]), shadow)
	for x in [-3.0, 0.0, 3.0]:
		draw_line(Vector2(x, 3), Vector2(x, 7), shadow, 1.2, true)
