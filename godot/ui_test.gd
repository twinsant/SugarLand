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
	var recording_status: Label = scene.get_node("%RecordingStatus")
	var selection: RichTextLabel = scene.get_node("%Selection")
	var economy: Label = scene.get_node("%Economy")
	var status: Label = scene.get_node("%Status")
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
	var settings_button: Button = scene.get_node("%Settings")
	click(settings_button.get_global_rect().get_center())
	check(scene.settings_panel != null and is_instance_valid(scene.settings_panel), "Settings button must open the settings panel")
	check(not scene.running, "Opening settings must pause the economy")
	var panel = scene.settings_panel
	panel._growth_rate.value = 2.5
	panel._peak_capacity.value = 6.0
	panel._apply()
	await process_frame
	check(scene.simulation.growth_rate == 2.5, "Applying settings must update the growth rate")
	check(scene.simulation.peak_capacity == 6.0, "Applying settings must update the peak capacity")
	check(scene.simulation.tick == 0, "Applying settings must reset the simulation")
	check(not is_instance_valid(panel), "Applying settings must close the panel")
	check(scene.running, "Closing settings must restore the running state")
	scene.simulation.growth_rate = 1.0
	scene.simulation.peak_capacity = 4.0
	scene.simulation.peak_a = Vector2i(15, 15)
	scene.simulation.peak_b = Vector2i(35, 35)
	scene.simulation.peak_falloff = 0.16
	# Breed mode: two wealthy residents in mutual vision produce a child
	scene.simulation.repopulate_mode = "breed"
	scene.reset()
	# Place two residents adjacent with high wealth and vision to guarantee breeding
	var a = scene.simulation.citizens[0]
	var b = scene.simulation.citizens[1]
	a.cell = Vector2i(25, 25)
	a.wealth = 20.0
	a.metabolism = 1
	a.vision = 6
	a.age = 20
	a.max_age = 100
	b.cell = Vector2i(26, 25)
	b.wealth = 20.0
	b.metabolism = 1
	b.vision = 6
	b.age = 20
	b.max_age = 100
	# Clear all sugar and set high sugar only at a and b positions so they stay put
	scene.simulation.sugar.fill(0.0)
	scene.simulation.capacities.fill(0.0)
	scene.simulation.capacities[25 * 50 + 25] = 100.0
	scene.simulation.capacities[25 * 50 + 26] = 100.0
	scene.simulation.sugar[25 * 50 + 25] = 100.0
	scene.simulation.sugar[25 * 50 + 26] = 100.0
	# Clear other citizens' wealth to prevent them from breeding
	for i in range(2, scene.simulation.citizens.size()):
		scene.simulation.citizens[i].wealth = 0.1
	scene.advance()
	check(scene.simulation.bred > 0, "Breed mode must produce children when two wealthy residents see each other")
	check(scene.simulation.replacements == 0, "Breed mode must not use system replacement")
	check(not scene.simulation.breeding_events.is_empty(), "Breed event must be recorded")
	var event = scene.simulation.breeding_events[0]
	check(event.has("parent_a") and event.has("parent_b") and event.has("child"), "Breed event must record parent and child ids")
	check(scene.heart_markers.size() > 0, "Heart marker must be created for breeding event")
	var marker = scene.heart_markers[0]
	check(marker.parent_a != null and marker.parent_b != null, "Heart marker must track parent residents")
	# Child inherits average parent vision within mutation range
	var child_id = event["child"]
	var child = null
	for citizen in scene.simulation.citizens:
		if citizen.id == child_id:
			child = citizen
			break
	check(child != null, "Breed child must exist in citizens")
	if child != null:
		var avg_vision = (a.vision + b.vision) / 2
		check(child.vision == clampi(avg_vision - 1, 1, 6) or child.vision == avg_vision or child.vision == clampi(avg_vision + 1, 1, 6), "Breed child must inherit average parent vision within mutation")
		check(child.wealth > 0.0, "Breed child must receive wealth from parents")
		check(child.age == 0, "Breed child must start at age 0")
	# Parents' wealth decreased compared to no-breeding baseline
	var no_breed_wealth := 20.0 + 100.0 - 1.0
	check(a.wealth < no_breed_wealth and b.wealth < no_breed_wealth, "Breed parents must contribute wealth to child")
	# Death inheritance: parent dies, wealth splits among surviving children
	var parent_a_id = a.id
	var child_wealth_before = child.wealth
	# Kill parent a
	a.max_age = a.age + 1
	scene.advance()
	# Find the child again (it may have moved in citizens array)
	var child_after = null
	for citizen in scene.simulation.citizens:
		if citizen.id == child_id:
			child_after = citizen
			break
	if child_after != null:
		check(child_after.wealth > child_wealth_before, "Child must inherit wealth when parent dies in breed mode")
	# Parent a should be dead
	var parent_a_alive = false
	for citizen in scene.simulation.citizens:
		if citizen.id == parent_a_id:
			parent_a_alive = true
			break
	check(not parent_a_alive, "Parent must die after reaching max age")
	# Age limit: residents outside breeding age cannot breed
	scene.simulation.repopulate_mode = "breed"
	scene.simulation.bred = 0
	scene.simulation.breeding_events.clear()
	scene.reset()
	var young = scene.simulation.citizens[0]
	var old = scene.simulation.citizens[1]
	young.cell = Vector2i(25, 25)
	young.wealth = 50.0
	young.metabolism = 1
	young.vision = 6
	young.age = 5
	young.max_age = 100
	old.cell = Vector2i(26, 25)
	old.wealth = 50.0
	old.metabolism = 1
	old.vision = 6
	old.age = 70
	old.max_age = 100
	scene.simulation.sugar.fill(0.0)
	scene.simulation.capacities.fill(0.0)
	scene.simulation.capacities[25 * 50 + 25] = 100.0
	scene.simulation.capacities[25 * 50 + 26] = 100.0
	scene.simulation.sugar[25 * 50 + 25] = 100.0
	scene.simulation.sugar[25 * 50 + 26] = 100.0
	for i in range(2, scene.simulation.citizens.size()):
		scene.simulation.citizens[i].wealth = 0.1
	scene.advance()
	check(scene.simulation.bred == 0, "Residents outside breeding age must not breed")
	# Restore system mode
	scene.simulation.repopulate_mode = "system"
	scene.simulation.bred = 0
	scene.simulation.breed_age_min = 15
	scene.simulation.breed_age_max = 65
	click(record_button.get_global_rect().get_center())
	check(scene.recording and record_button.text == "停止录像 [V]", "Record button must show stop action")
	await create_timer(0.1).timeout
	check(recording_status.text.contains("帧"), "Recording status label must show frame count")
	if DisplayServer.get_name() != "headless":
		check(scene.recording_frame > 0, "Recording must write PNG frames while paused")
	click(record_button.get_global_rect().get_center())
	check(not scene.recording and record_button.text == "录像 [V]", "Record button must show start action")
	check(recording_status.text.contains("完成") or recording_status.text.contains("保留") or recording_status.text.contains("没有可用帧"), "Recording status label must show result")
	var wheel := InputEventMouseButton.new()
	wheel.position = Vector2(350, 400)
	wheel.button_index = MOUSE_BUTTON_WHEEL_UP
	wheel.pressed = true
	root.push_input(wheel, true)
	check(scene.map_layer.scale.x > 1.0 and zoom_status.text != "地图缩放 100%", "Wheel must zoom the map and update the scale")
	var before_pan: Vector2 = scene.map_layer.position
	var pan := InputEventPanGesture.new()
	pan.position = Vector2(350, 400)
	pan.delta = Vector2(1, 1)
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
	var residents_root: Node = scene.get_node("MapClip/MapLayer/Residents")
	var map_center: Vector2 = scene.MAP_ORIGIN + Vector2.ONE * scene.MAP_SIZE * 0.5
	var follow_view: Node2D = residents_root.get_child(0)
	var best_distance := follow_view.position.distance_to(map_center)
	for resident_node in residents_root.get_children():
		var d: float = resident_node.position.distance_to(map_center)
		if d < best_distance:
			best_distance = d
			follow_view = resident_node
	click(follow_view.get_global_position())
	scene.map_layer.position = Vector2(-600, -600)
	scene.update_selected_follow(10.0)
	check(follow_view.get_global_position().distance_to(scene.map_clip.get_global_rect().get_center()) < 1.0, "Selected resident must recenter smoothly when near the viewport edge")
	key(KEY_0)
	check(scene.map_layer.scale == Vector2.ONE and scene.map_layer.position == Vector2.ZERO and zoom_status.text == "地图缩放 100%", "Zero must reset the map view and scale label")
	# At 100% zoom, selected resident should not pan the map
	click(follow_view.get_global_position())
	scene.map_layer.position = Vector2(-100, -100)
	scene.update_selected_follow(10.0)
	check(scene.map_layer.position == Vector2(-100, -100), "Map must not pan for selected resident at 100% zoom")
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
	# Land privatization: wealthy resident acquires land by affordability, others pay rent
	scene.simulation.land_privacy_enabled = true
	scene.simulation.peak_capacity = 4.0
	scene.reset()
	var landowner = scene.simulation.citizens[0]
	landowner.cell = Vector2i(25, 25)
	landowner.wealth = 73.0
	landowner.metabolism = 2
	landowner.vision = 1
	landowner.age = 20
	landowner.max_age = 100
	scene.simulation.sugar.fill(0.0)
	scene.simulation.capacities.fill(0.0)
	for dy in range(-1, 2):
		for dx in range(-1, 2):
			var idx := (25 + dy) * 50 + (25 + dx)
			scene.simulation.capacities[idx] = 4.0
			scene.simulation.sugar[idx] = 4.0
	for i in range(1, scene.simulation.citizens.size()):
		scene.simulation.citizens[i].wealth = 0.1
	scene.advance()
	# 73 wealth, 8 per cell, 50 reserve -> buys 3 cells regardless of move
	check(scene.simulation.land_owners.size() == 3, "Wealthy resident must buy 3 cells by affordability")
	var owner_values: Array = scene.simulation.land_owners.values()
	check(landowner.id in owner_values, "Land owner must be the wealthy resident")
	# Now test rent: landowner moves away, renter harvests on owned land
	scene.simulation.land_owners.clear()
	scene.simulation.land_owners[25 * 50 + 25] = landowner.id
	landowner.cell = Vector2i(26, 25)
	landowner.wealth = 7.0
	landowner.vision = 1
	var renter = scene.simulation.citizens[1]
	renter.cell = Vector2i(25, 25)
	renter.wealth = 10.0
	renter.metabolism = 1
	renter.vision = 1
	renter.age = 20
	renter.max_age = 100
	scene.simulation.sugar.fill(0.0)
	scene.simulation.capacities.fill(0.0)
	scene.simulation.capacities[25 * 50 + 25] = 4.0
	scene.simulation.sugar[25 * 50 + 25] = 4.0
	scene.simulation.capacities[26 * 50 + 25] = 0.0
	var owner_wealth_before: float = landowner.wealth
	scene.advance()
	check(landowner.wealth > owner_wealth_before, "Landowner must receive rent from renter")
	# Land inheritance: when landowner dies, land goes to wealthiest child
	scene.simulation.repopulate_mode = "breed"
	scene.simulation.land_privacy_enabled = true
	scene.reset()
	var parent = scene.simulation.citizens[0]
	parent.cell = Vector2i(25, 25)
	parent.wealth = 73.0
	parent.metabolism = 2
	parent.vision = 1
	parent.age = 20
	parent.max_age = 21  # Will die next step
	var heir = scene.simulation.citizens[1]
	heir.cell = Vector2i(26, 25)
	heir.wealth = 30.0
	heir.metabolism = 1
	heir.vision = 1
	heir.age = 20
	heir.max_age = 100
	parent.children.append(heir.id)
	heir.parents.append(parent.id)
	scene.simulation.sugar.fill(0.0)
	scene.simulation.capacities.fill(0.0)
	for dy in range(-1, 2):
		for dx in range(-1, 2):
			var idx := (25 + dy) * 50 + (25 + dx)
			scene.simulation.capacities[idx] = 4.0
			scene.simulation.sugar[idx] = 4.0
	for i in range(2, scene.simulation.citizens.size()):
		scene.simulation.citizens[i].wealth = 0.1
	scene.advance()
	check(scene.simulation.land_owners.size() > 0, "Wealthy resident must acquire land")
	var land_inherited := false
	for key in scene.simulation.land_owners.keys():
		if scene.simulation.land_owners[key] == heir.id:
			land_inherited = true
			break
	check(land_inherited, "Land must inherit to wealthiest child when parent dies")
	scene.simulation.land_privacy_enabled = false
	scene.simulation.land_owners.clear()
	scene.simulation.repopulate_mode = "system"
	key(KEY_R)
	for iteration in 10:
		key(KEY_N)
	check(scene.simulation.tick == 10, "Repeated single-steps must not skip rounds")
	# Stop on extinction: kill all residents and next step should not advance
	for c in scene.simulation.citizens.duplicate():
		c.death_reason = "starvation"
	scene.advance()
	check(scene.simulation.citizens.is_empty(), "All residents must die")
	var tick_before: int = scene.simulation.tick
	scene.advance()
	check(scene.simulation.tick == tick_before, "Advance must not step when population extinct")
	key(KEY_R)
	check(scene.simulation.tick == 0, "Reset must restore population after extinction")
	if DisplayServer.get_name() != "headless":
		print("UI checks complete; capturing viewport")
		await RenderingServer.frame_post_draw
		for argument in OS.get_cmdline_user_args():
			if argument.begins_with("--capture="):
				var path := argument.trim_prefix("--capture=")
				check(root.get_texture().get_image().save_png(path) == OK, "Viewport capture must save")
	print("SugarLand UI tests: %d failures" % failures.size())
	quit(0 if failures.is_empty() else 1)
