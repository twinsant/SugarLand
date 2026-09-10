extends Node2D

const FOG_SIZE := 4000.0
const FOG_ORIGIN := -FOG_SIZE * 0.5

var _fog_rect: ColorRect
var _fog_material: ShaderMaterial


func _ready() -> void:
	_fog_rect = ColorRect.new()
	_fog_rect.position = Vector2.ONE * FOG_ORIGIN
	_fog_rect.size = Vector2.ONE * FOG_SIZE
	_fog_rect.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_fog_material = ShaderMaterial.new()
	var shader := Shader.new()
	shader.code = """
shader_type canvas_item;

uniform vec2 center_uv = vec2(0.5);
uniform float radius_uv = 0.08;
uniform float feather_uv = 0.02;
uniform float density = 0.92;

void fragment() {
    float distance_from_center = distance(UV, center_uv);
    float alpha = smoothstep(radius_uv, radius_uv + feather_uv, distance_from_center) * density;
    COLOR = vec4(0.01, 0.02, 0.03, alpha);
}
"""
	_fog_material.shader = shader
	_fog_rect.material = _fog_material
	add_child(_fog_rect)
	visible = false


func set_fog(enabled: bool, center: Vector2, radius: float, feather: float) -> void:
	visible = enabled
	if not enabled or _fog_material == null:
		return
	_fog_material.set_shader_parameter("center_uv", (center - Vector2.ONE * FOG_ORIGIN) / FOG_SIZE)
	_fog_material.set_shader_parameter("radius_uv", radius / FOG_SIZE)
	_fog_material.set_shader_parameter("feather_uv", feather / FOG_SIZE)
