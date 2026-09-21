shader_type canvas_item;

uniform vec4 ring_color : source_color = vec4(0.27, 0.78, 0.91, 1.0);
uniform vec4 halo_color : source_color = vec4(0.45, 0.88, 1.0, 1.0);
uniform float ring_pos = 0.82;
uniform float ring_width = 0.022;
uniform float pulse_speed = 1.8;

void fragment() {
	float d = distance(UV, vec2(0.5)) * 2.0;
	float breath = pow(0.5 + 0.5 * sin(TIME * pulse_speed), 1.8);
	float w = ring_width * (0.85 + 0.5 * breath);
	float ring = smoothstep(w, 0.0, abs(d - ring_pos));
	float halo_pos = ring_pos + 0.015 + 0.045 * breath;
	float halo = smoothstep(0.10, 0.0, abs(d - halo_pos));
	float ring_a = ring * (0.35 + 0.35 * breath);
	float halo_a = halo * 0.16 * breath;
	vec3 rgb = (ring_color.rgb * ring_a + halo_color.rgb * halo_a) / max(ring_a + halo_a, 0.0001);
	COLOR = vec4(rgb, min(ring_a + halo_a, 1.0));
}
