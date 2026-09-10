extends Node2D

const Resident = preload("res://resident.gd")

signal finished

var pages: Array = []
var page_index := 0
var bubble: Label
var next_button: Button


func _ready() -> void:
	z_index = 300
	load_pages()
	var resident: Resident = get_parent()
	resident.set_tutorial_thought("")
	bubble = resident.thought_label
	next_button = Button.new()
	next_button.name = "NextButton"
	next_button.text = "下一步"
	next_button.position = Vector2(-45, 20)
	next_button.size = Vector2(90, 28)
	next_button.add_theme_font_size_override("font_size", 13)
	next_button.pressed.connect(next_page)
	add_child(next_button)
	update_page()



func load_pages() -> void:
	var parsed = JSON.parse_string(FileAccess.get_file_as_string("res://intro.json"))
	if parsed is Dictionary and parsed.has("pages"):
		pages = parsed["pages"]


func update_page() -> void:
	if pages.is_empty():
		bubble.text = "教程内容暂时不可用。"
		return
	var page: Dictionary = pages[page_index]
	var text := "%s\n%s" % [page.get("title", ""), page.get("text", "")]
	(get_parent() as Resident).set_tutorial_thought(text)
	bubble = (get_parent() as Resident).thought_label
	next_button.text = "进入糖域" if page_index == pages.size() - 1 else "下一步"


func next_page() -> void:
	if page_index >= pages.size() - 1:
		finished.emit()
		return
	page_index += 1
	update_page()

