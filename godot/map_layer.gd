extends Node2D

var grid_size := 50
var tile_size := 14.0
var map_origin := Vector2(32, 10)
var display_sugar := PackedFloat32Array()
var land_owners := {}
var owner_colors := {}
var selected_cell := Vector2i(-1, -1)
var tutorial_action := "none"
var tutorial_resident_pos := Vector2.ZERO
const TUTORIAL_COLOR := Color("ffd23f")


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
			var owner_id: int = land_owners.get(y * grid_size + x, -1)
			if owner_id >= 0:
				var owner_color: Color = owner_colors.get(owner_id, Color(1, 1, 1, 0.5))
				draw_rect(Rect2(map_origin + Vector2(x, y) * tile_size, Vector2.ONE * tile_size), owner_color, false, 1.5)
	if selected_cell.x >= 0:
		draw_rect(Rect2(map_origin + Vector2(selected_cell) * tile_size, Vector2.ONE * tile_size), Color("fff6dc"), false, 2.0)
	if tutorial_action == "frame_map":
		draw_rect(Rect2(map_origin - Vector2.ONE * 4, Vector2.ONE * (map_size + 8)), TUTORIAL_COLOR, false, 3.0)
