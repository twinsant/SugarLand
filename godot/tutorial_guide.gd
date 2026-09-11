extends Node2D

const Resident = preload("res://resident.gd")

signal finished
signal action_changed(action: String)

var pages: Array = []
var page_index := 0
var next_button: Button


func _ready() -> void:
	z_index = 300
	load_pages()
	var resident: Resident = get_parent()
	resident.set_tutorial_thought("", "")
	next_button = Button.new()
	next_button.name = "NextButton"
	next_button.text = "下一步"
	next_button.position = Vector2(-45, 20)
	next_button.size = Vector2(90, 28)
	next_button.add_theme_font_size_override("font_size", 13)
	var btn_style := StyleBoxFlat.new()
	btn_style.bg_color = Color(0.27, 0.78, 0.91, 1.0)
	btn_style.border_color = Color(0.15, 0.55, 0.68, 1.0)
	btn_style.set_border_width_all(1)
	btn_style.set_corner_radius_all(6)
	btn_style.content_margin_left = 8
	btn_style.content_margin_right = 8
	next_button.add_theme_stylebox_override("normal", btn_style)
	var btn_hover := StyleBoxFlat.new()
	btn_hover.bg_color = Color(0.35, 0.85, 0.98, 1.0)
	btn_hover.border_color = Color(0.15, 0.55, 0.68, 1.0)
	btn_hover.set_border_width_all(1)
	btn_hover.set_corner_radius_all(6)
	btn_hover.content_margin_left = 8
	btn_hover.content_margin_right = 8
	next_button.add_theme_stylebox_override("hover", btn_hover)
	next_button.add_theme_color_override("font_color", Color(1.0, 1.0, 1.0, 1.0))
	next_button.add_theme_color_override("font_hover_color", Color(1.0, 1.0, 1.0, 1.0))
	next_button.pressed.connect(next_page)
	add_child(next_button)
	update_page()



func load_pages() -> void:
	var parsed = JSON.parse_string(FileAccess.get_file_as_string("res://intro.json"))
	if parsed is Dictionary and parsed.has("pages"):
		pages = parsed["pages"]


func update_page() -> void:
	if pages.is_empty():
		(get_parent() as Resident).set_tutorial_thought("教程", "教程内容暂时不可用。")
		return
	var page: Dictionary = pages[page_index]
	(get_parent() as Resident).set_tutorial_thought(page.get("title", ""), page.get("text", ""))
	next_button.text = "进入糖域" if page_index == pages.size() - 1 else "下一步"
	action_changed.emit(page.get("action", "none"))


func next_page() -> void:
	if page_index >= pages.size() - 1:
		finished.emit()
		return
	page_index += 1
	update_page()
