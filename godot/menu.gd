extends Control

signal start_game
signal start_tutorial

const IntroActor = preload("res://intro_actor.gd")

var pages: Array = []
var page_index := 0
var intro_mode := false
var title_label: Label
var bubble_label: Label
var hint_label: Label
var page_label: Label
var intro_button: Button
var start_button: Button
var next_button: Button
var back_button: Button
var actor: Node2D


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_STOP
	load_pages()
	build_ui()
	resized.connect(layout_ui)
	layout_ui()
	show_home()


func load_pages() -> void:
	var parsed = JSON.parse_string(FileAccess.get_file_as_string("res://intro.json"))
	if parsed is Dictionary and parsed.has("pages"):
		pages = parsed["pages"]


func build_ui() -> void:
	title_label = make_label(30, Color("dff5e6"))
	add_child(title_label)
	bubble_label = make_label(18, Color("18323b"))
	bubble_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	bubble_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	bubble_label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	bubble_label.add_theme_stylebox_override("normal", make_panel(Color(0.93, 0.98, 1.0, 0.9), Color("45c7e8"), 14))
	add_child(bubble_label)
	hint_label = make_label(14, Color("9bb9bd"))
	hint_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	add_child(hint_label)
	page_label = make_label(13, Color("6e9298"))
	page_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	add_child(page_label)
	actor = IntroActor.new()
	add_child(actor)
	intro_button = make_button("游戏介绍")
	intro_button.name = "IntroButton"
	intro_button.pressed.connect(func(): start_tutorial.emit())
	add_child(intro_button)
	start_button = make_button("开始游戏")
	start_button.name = "StartButton"
	start_button.pressed.connect(func(): start_game.emit())
	add_child(start_button)
	back_button = make_button("返回主菜单")
	back_button.name = "BackButton"
	back_button.pressed.connect(show_home)
	add_child(back_button)
	next_button = make_button("下一步")
	next_button.name = "NextButton"
	next_button.pressed.connect(next_page)
	add_child(next_button)


func make_label(font_size: int, color: Color) -> Label:
	var label := Label.new()
	label.add_theme_font_size_override("font_size", font_size)
	label.add_theme_color_override("font_color", color)
	return label


func make_button(text: String) -> Button:
	var button := Button.new()
	button.text = text
	button.custom_minimum_size = Vector2(180, 48)
	button.add_theme_font_size_override("font_size", 17)
	return button


func make_panel(background: Color, border: Color, radius: int) -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = background
	style.border_color = border
	style.set_border_width_all(1)
	style.set_corner_radius_all(radius)
	style.content_margin_left = 20
	style.content_margin_right = 20
	style.content_margin_top = 12
	style.content_margin_bottom = 12
	return style


func layout_ui() -> void:
	var view := size
	var panel := Rect2(view.x * 0.18, view.y * 0.12, view.x * 0.64, view.y * 0.76)
	title_label.position = panel.position + Vector2(0, 28)
	title_label.size = Vector2(panel.size.x, 48)
	bubble_label.position = panel.position + Vector2(panel.size.x * 0.32, 112)
	bubble_label.size = Vector2(panel.size.x * 0.58, 142)
	hint_label.position = panel.position + Vector2(panel.size.x * 0.28, 270)
	hint_label.size = Vector2(panel.size.x * 0.64, 32)
	page_label.position = panel.position + Vector2(0, panel.size.y - 72)
	page_label.size = Vector2(panel.size.x, 28)
	actor.position = panel.position + Vector2(panel.size.x * 0.22, 330)
	intro_button.position = panel.position + Vector2(panel.size.x * 0.2, panel.size.y - 145)
	start_button.position = panel.position + Vector2(panel.size.x * 0.58, panel.size.y - 145)
	back_button.position = intro_button.position
	next_button.position = start_button.position
	queue_redraw()


func show_home() -> void:
	intro_mode = false
	page_index = 0
	bubble_label.text = "糖域正在等待你的观察。\n这是一个关于资源、财富与选择的人工社会。"
	hint_label.text = "选择一个入口开始。"
	page_label.text = ""
	intro_button.visible = true
	start_button.visible = true
	back_button.visible = false
	next_button.visible = false
	actor.visible = true
	queue_redraw()


func show_intro() -> void:
	intro_mode = true
	page_index = 0
	intro_button.visible = false
	start_button.visible = false
	back_button.visible = true
	next_button.visible = true
	actor.visible = true
	update_page()


func update_page() -> void:
	if pages.is_empty():
		bubble_label.text = "介绍内容暂时不可用。"
		hint_label.text = "你仍然可以直接开始游戏。"
		return
	var page: Dictionary = pages[page_index]
	bubble_label.text = "%s\n\n%s" % [page.get("title", ""), page.get("text", "")]
	hint_label.text = page.get("hint", "")
	page_label.text = "%d / %d" % [page_index + 1, pages.size()]
	next_button.text = "开始游戏" if page_index == pages.size() - 1 else "下一步"


func next_page() -> void:
	if page_index >= pages.size() - 1:
		start_game.emit()
		return
	page_index += 1
	update_page()


func _draw() -> void:
	draw_rect(Rect2(Vector2.ZERO, size), Color(0.02, 0.045, 0.065, 0.97))
	var panel := Rect2(size.x * 0.18, size.y * 0.12, size.x * 0.64, size.y * 0.76)
	draw_style_box(make_panel(Color(0.05, 0.1, 0.12, 0.98), Color("315d64"), 18), panel)
