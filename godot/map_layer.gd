extends Node2D

var grid_size := 50
var tile_size := 14.0
var map_origin := Vector2(32, 10)
var display_sugar := PackedFloat32Array()
var selected_cell := Vector2i(-1, -1)


func _draw() -> void:
	if display_sugar.is_empty():
		return
	var map_size := grid_size * tile_size
	draw_rect(Rect2(map_origin - Vector2.ONE * 2, Vector2.ONE * (map_size + 4)), Color("53685e"), false, 2.0)
	for y in grid_size:
		for x in grid_size:
			var richness := display_sugar[y * grid_size + x] / 4.0
			var color := Color("d3b86e").lerp(Color("245b3c"), richness)
			draw_rect(Rect2(map_origin + Vector2(x, y) * tile_size, Vector2.ONE * (tile_size - 0.7)), color)
	if selected_cell.x >= 0:
		draw_rect(Rect2(map_origin + Vector2(selected_cell) * tile_size, Vector2.ONE * tile_size), Color("fff6dc"), false, 2.0)
