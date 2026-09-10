extends Node2D

var elapsed := 0.0


func _process(delta: float) -> void:
	elapsed += delta
	queue_redraw()


func _draw() -> void:
	var bob := sin(elapsed * 2.4) * 2.0
	draw_circle(Vector2(0, 18), 18, Color(0, 0, 0, 0.25))
	draw_circle(Vector2(0, -12 + bob), 16, Color("ffe2b8"))
	draw_circle(Vector2(-5, -15 + bob), 2, Color("18323b"))
	draw_circle(Vector2(5, -15 + bob), 2, Color("18323b"))
	draw_circle(Vector2(0, -5 + bob), 3, Color("e9938c"))
	draw_rect(Rect2(-17, 5 + bob, 34, 39), Color("d4af37"), true)
	draw_line(Vector2(-11, 44 + bob), Vector2(-15, 59 + bob), Color("18323b"), 5, true)
	draw_line(Vector2(11, 44 + bob), Vector2(15, 59 + bob), Color("18323b"), 5, true)
