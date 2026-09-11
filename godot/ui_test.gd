extends SceneTree

var failures: Array[String] = []


func _initialize() -> void:
	run.call_deferred()


func check(condition: bool, message: String) -> void:
	if not condition:
		failures.append(message)
		push_error(message)


func click(point: Vector2) -> void:
	var motion := InputEventMouseMotion.new()
	motion.position = point
	root.push_input(motion, true)
	for pressed in [true, false]:
		var event := InputEventMouseButton.new()
		event.position = point
		event.button_index = MOUSE_BUTTON_LEFT
		event.pressed = pressed
		root.push_input(event, true)


func key(code: Key) -> void:
	for pressed in [true, false]:
		var event := InputEventKey.new()
		event.keycode = code
		event.pressed = pressed
		root.push_input(event, true)


func run() -> void:
	create_timer(30.0).timeout.connect(func():
		push_error("UI tests timed out")
		quit(1)
	)
	var scene = load("res://main.tscn").instantiate()
	root.add_child(scene)
	await process_frame
	check(scene.game_started and scene.tutorial_mode and not scene.running, "Game must open directly in the paused world tutorial")
	var guide = scene.tutorial_guide
	var visible_residents := 0
	for resident in scene.residents.values():
		if resident.visible:
			visible_residents += 1
	check(visible_residents == 1, "World tutorial must show exactly one real resident")
	var guide_res = scene.residents[scene.simulation.citizens[0].id]
	var bubble_root = guide_res.get_node("ThoughtBubbleRoot")
	var title_label: Label = bubble_root.get_node("BubbleTitle")
	var body_label: Label = bubble_root.get_node("ThoughtBubble")
	check(guide.get_parent() == guide_res and title_label.text.contains("欢迎") and guide.next_button.visible, "World tutorial must attach its bubble to the real guide resident")
	var intro_text: String = body_label.text
	check(scene.tutorial_action == "frame_resident", "First tutorial page must select the resident")
	check(guide_res.selected, "frame_resident action must select the guide resident")
	click(guide.next_button.get_global_rect().get_center())
	check(body_label.text != intro_text, "World tutorial next button must advance JSON pages")
	check(scene.tutorial_action == "frame_map", "Second tutorial page must frame the map")
	# Advance to a tile-selection page
	while scene.tutorial_action != "select_rich_tile" and scene.tutorial_guide != null:
		click(guide.next_button.get_global_rect().get_center())
		await process_frame
	check(scene.selected_cell.x >= 0, "Tile selection action must select a cell")
	scene.finish_tutorial()
	check(not scene.tutorial_mode and scene.running, "Finishing tutorial must start the current game logic")
	var pause: Button = scene.get_node("%Pause")
	var step_button: Button = scene.get_node("%Step")
	var reset_button: Button = scene.get_node("%Reset")
	var record_button: Button = scene.get_node("%Record")
	var selection: Label = scene.get_node("%Selection")
	var economy: Label = scene.get_node("%Economy")
	var status: Label = scene.get_node("%Status")
	var recording_status: Label = scene.get_node("%RecordingStatus")
	var zoom_status: Label = scene.get_node("%ZoomStatus")
	var initial_economy := economy.text
	check(zoom_status.text == "地图缩放 100%", "Zoom status must show the default scale")
	var first = scene.simulation.citizens[0]
	var initial_cell: Vector2i = first.cell
	var initial_wealth: float = first.wealth
	click(pause.get_global_rect().get_center())
	check(pause.text.begins_with("继续"), "Pause button must pause the economy")
	await create_timer(1.1).timeout
	check(scene.simulation.tick == 0, "Paused economy must not advance")
	var first_view: Node2D = scene.get_node("MapClip/MapLayer/Residents").get_child(0)
	click(first_view.get_global_position())
	check(selection.text.contains("居民 #001"), "Resident click must show its identity")
	check(selection.text.contains("财富") and selection.text.contains("代谢"), "Resident details must show economic attributes")
	check(first_view.get_node("ThoughtBubbleRoot/ThoughtBubble").visible and first_view.get_node("ThoughtBubbleRoot/ThoughtBubble").text.contains("#001") and first_view.get_node("ThoughtBubbleRoot/ThoughtBubble").text.contains("糖"), "Selected resident bubble must show its identity and thought")
	check(scene.fog_enabled and scene.fog_layer.visible, "Selecting a resident must automatically show fog")
	key(KEY_F)
	check(not scene.fog_enabled and not scene.fog_layer.visible, "F must hide the selected resident's fog")
	key(KEY_F)
	check(scene.fog_enabled and scene.fog_layer.visible, "F must show the selected resident's fog")
	key(KEY_E)
	check(scene.manual_controlled and scene.fog_enabled and scene.fog_layer.visible, "E must take over the selected resident and force fog on")
	key(KEY_RIGHT)
	check(scene.simulation.manual_commands.has(first.id), "Arrow key must queue a manual resident move")
	key(KEY_F)
	check(scene.fog_enabled, "Fog must stay on during manual takeover")
	key(KEY_E)
	check(not scene.manual_controlled, "E must release manual takeover")
	click(step_button.get_global_rect().get_center())
	check(scene.simulation.tick == 1 and not scene.running, "Step button must advance exactly one round and pause")
	check(first.age == 1, "A single round must age a resident once")
	check(economy.text != initial_economy, "Economic statistics must refresh after a round")
	check(selection.text.contains("年龄 1 /"), "Selected resident details must refresh after a round")
	var paused_position: Vector2 = first_view.position
	var paused_wealth: float = first.wealth
	await create_timer(0.2).timeout
	check(first_view.position == paused_position and first.wealth == paused_wealth, "Single-step must leave both animation and economy paused")
	key(KEY_N)
	check(scene.simulation.tick == 2, "N must advance exactly one additional round")
	first.wealth = 100.0
	first.max_age = first.age + 1
	key(KEY_N)
	var richest = scene.simulation.citizens[0]
	for citizen in scene.simulation.citizens:
		if citizen.wealth > richest.wealth:
			richest = citizen
	check(scene.selected_citizen == richest, "Selected resident death must choose the richest resident first")
	check(scene.get_node("MapClip/MapLayer/Residents").get_child_count() == 80, "Death replacement must preserve 80 visible residents")
	check(scene.death_markers.size() > 0, "Death must create a skull marker")
	await process_frame
	var fixed_msg: Label = scene.get_node("%SelectedDeathMessage")
	check(fixed_msg.text.contains("寿终") and fixed_msg.text.contains("年龄") and fixed_msg.text.contains("财富"), "Selected resident death must show a fixed message with reason, age and wealth")
	richest.max_age = richest.age + 1
	key(KEY_N)
	var poorest = scene.simulation.citizens[0]
	for citizen in scene.simulation.citizens:
		if citizen.wealth < poorest.wealth:
			poorest = citizen
	check(scene.selected_citizen == poorest, "Next selected resident death must choose the poorest resident")
	click(reset_button.get_global_rect().get_center())
	check(scene.simulation.tick == 0 and scene.running, "Reset button must restart the clock")
	check(economy.text == initial_economy, "Reset must restore initial economic statistics")
	check(scene.simulation.citizens[0].cell == initial_cell and scene.simulation.citizens[0].wealth == initial_wealth, "Reset must restore initial residents")
	click(record_button.get_global_rect().get_center())
	check(scene.recording and record_button.text.begins_with("停止录像"), "Record button must start frame capture")
	await create_timer(0.1).timeout
	check(recording_status.text.contains("录像中"), "Recording status must update while paused")
	if DisplayServer.get_name() != "headless":
		check(scene.recording_frame > 0, "Recording must write PNG frames while paused")
	click(record_button.get_global_rect().get_center())
	check(not scene.recording and recording_status.text.contains("录像结束"), "Record button must stop frame capture")
	var wheel := InputEventMouseButton.new()
	wheel.position = Vector2(350, 400)
	wheel.button_index = MOUSE_BUTTON_WHEEL_UP
	wheel.pressed = true
	root.push_input(wheel, true)
	check(scene.map_layer.scale.x > 1.0 and zoom_status.text != "地图缩放 100%", "Wheel must zoom the map and update the scale")
	var before_pan: Vector2 = scene.map_layer.position
	var pan := InputEventPanGesture.new()
	pan.position = Vector2(350, 400)
	pan.delta = Vector2(18, 12)
	root.push_input(pan, true)
	check(scene.map_layer.position == before_pan - pan.delta * scene.TRACKPAD_PAN_SENSITIVITY, "Trackpad pan must follow Apple's natural content direction with sensitivity compensation")
	var magnify := InputEventMagnifyGesture.new()
	magnify.position = Vector2(350, 400)
	magnify.factor = 1.1
	var before_magnify: float = scene.map_layer.scale.x
	root.push_input(magnify, true)
	check(scene.map_layer.scale.x > before_magnify, "Trackpad pinch must zoom the map")
	for iteration in 20:
		root.push_input(wheel, true)
	check(scene.map_layer.scale.x <= 3.0 and zoom_status.text == "地图缩放 300%", "Map zoom must be capped at 300%")
	key(KEY_0)
	check(scene.map_layer.scale == Vector2.ONE and scene.map_layer.position == Vector2.ZERO and zoom_status.text == "地图缩放 100%", "Zero must reset the map view and scale label")
	var follow_view: Node2D = scene.get_node("MapClip/MapLayer/Residents").get_child(0)
	click(follow_view.get_global_position())
	scene.map_layer.position = Vector2(-600, -600)
	scene.update_selected_follow(10.0)
	check(follow_view.get_global_position().distance_to(scene.map_clip.get_global_rect().get_center()) < 1.0, "Selected resident must recenter smoothly when near the viewport edge")
	key(KEY_0)
	key(KEY_SPACE)
	check(not scene.running, "Space must pause even when Reset has focus")
	click(Vector2(39, 103))
	check(selection.text.contains("(0, 0)"), "Map click must select the correct resource cell")
	key(KEY_SPACE)
	await create_timer(1.1).timeout
	check(scene.simulation.tick >= 1, "Resuming must restart automatic economic rounds")
	check(status.text.contains("运行中"), "Status must show the running state")
	key(KEY_R)
	check(scene.simulation.tick == 0 and economy.text == initial_economy, "R must reset the economy")
	for iteration in 10:
		key(KEY_N)
	check(scene.simulation.tick == 10, "Repeated single-steps must not skip rounds")
	if DisplayServer.get_name() != "headless":
		print("UI checks complete; capturing viewport")
		await RenderingServer.frame_post_draw
		for argument in OS.get_cmdline_user_args():
			if argument.begins_with("--capture="):
				var path := argument.trim_prefix("--capture=")
				check(root.get_texture().get_image().save_png(path) == OK, "Viewport capture must save")
	print("SugarLand UI tests: %d failures" % failures.size())
	quit(0 if failures.is_empty() else 1)
