extends RefCounted

const GRID_SIZE := 50
const POPULATION := 80
const DIRECTIONS := [Vector2i.LEFT, Vector2i.RIGHT, Vector2i.UP, Vector2i.DOWN]

class Citizen extends RefCounted:
	var id: int
	var cell: Vector2i
	var wealth: float
	var vision: int
	var metabolism: int
	var age: int = 0
	var max_age: int
	var last_harvest: float = 0.0
	var death_reason: String = ""
	var children: Array[int] = []
	var parents: Array[int] = []
	var generation := 0

var citizens: Array[Citizen] = []
var sugar := PackedFloat32Array()
var capacities := PackedFloat32Array()
var peak_a := Vector2i(15, 15)
var peak_b := Vector2i(35, 35)
var peak_capacity := 4.0
var peak_falloff := 0.16
var growth_rate := 1.0
var repopulate_mode := "system"
var breed_age_min := 15
var breed_age_max := 65
var move_cost := 0.5
var tick: int = 0
var starved: int = 0
var aged_out: int = 0
var replacements: int = 0
var bred: int = 0
var harvested_by_cell: Dictionary = {}
var death_events: Array[Dictionary] = []
var breeding_events: Array[Dictionary] = []
var manual_controlled: Dictionary = {}
var manual_commands: Dictionary = {}

var _rng := RandomNumberGenerator.new()
var _next_id: int = 1
var _harvested: float = 0.0
var _consumed: float = 0.0
var _move_cost_total: float = 0.0
var _growth: float = 0.0
var _injected_wealth: float = 0.0
var _removed_wealth: float = 0.0
var _released_wealth: float = 0.0


func reset(seed_value: int = 25) -> void:
	_rng.seed = seed_value
	_next_id = 1
	tick = 0
	starved = 0
	aged_out = 0
	replacements = 0
	bred = 0
	harvested_by_cell.clear()
	death_events.clear()
	breeding_events.clear()
	manual_controlled.clear()
	manual_commands.clear()
	_clear_budget()
	citizens.clear()
	capacities.resize(GRID_SIZE * GRID_SIZE)
	for y in GRID_SIZE:
		for x in GRID_SIZE:
			var cell := Vector2i(x, y)
			var distance := mini(_distance(cell, peak_a), _distance(cell, peak_b))
			capacities[_index(cell)] = maxf(0.0, peak_capacity - peak_falloff * distance)
	sugar = capacities.duplicate()
	var occupied := {}
	for index in POPULATION:
		_spawn(occupied)


func step() -> void:
	_clear_budget()
	harvested_by_cell.clear()
	death_events.clear()
	for index in sugar.size():
		var before := sugar[index]
		sugar[index] = minf(capacities[index], before + growth_rate)
		_growth += sugar[index] - before
	var occupied := {}
	for citizen in citizens:
		occupied[citizen.cell] = true
	var order: Array[Citizen] = citizens.duplicate()
	for index in range(order.size() - 1, 0, -1):
		var other := _rng.randi_range(0, index)
		var temporary := order[index]
		order[index] = order[other]
		order[other] = temporary
	for citizen in order:
		var previous_cell := citizen.cell
		var destination := _choose_cell(citizen, occupied)
		if manual_controlled.has(citizen.id):
			destination = _choose_manual_cell(citizen, occupied)
		manual_commands.erase(citizen.id)
		occupied.erase(citizen.cell)
		citizen.cell = destination
		occupied[destination] = true
		var index := _index(destination)
		citizen.last_harvest = sugar[index]
		sugar[index] = 0.0
		citizen.wealth += citizen.last_harvest
		citizen.wealth -= citizen.metabolism
		if destination != previous_cell:
			var move_distance := _distance(previous_cell, destination)
			citizen.wealth -= move_distance * move_cost
			_move_cost_total += move_distance * move_cost
		citizen.age += 1
		harvested_by_cell[index] = citizen.last_harvest
		_harvested += citizen.last_harvest
		_consumed += citizen.metabolism
		if citizen.wealth <= 0.0:
			citizen.death_reason = "starvation"
			starved += 1
			death_events.append({"id": citizen.id, "cell": citizen.cell, "reason": citizen.death_reason, "age": citizen.age, "wealth": citizen.wealth, "generation": citizen.generation})
		elif citizen.age >= citizen.max_age:
			citizen.death_reason = "age"
			aged_out += 1
			death_events.append({"id": citizen.id, "cell": citizen.cell, "reason": citizen.death_reason, "age": citizen.age, "wealth": citizen.wealth, "generation": citizen.generation})
	var survivors: Array[Citizen] = []
	for citizen in citizens:
		if citizen.death_reason.is_empty():
			survivors.append(citizen)
		else:
			occupied.erase(citizen.cell)
			_removed_wealth += citizen.wealth
			if citizen.wealth > 0.0:
				if repopulate_mode == "breed" and not citizen.children.is_empty():
					_inherit(citizen)
				else:
					var death_index := _index(citizen.cell)
					var available_capacity := capacities[death_index] - sugar[death_index]
					var released := minf(citizen.wealth, maxf(available_capacity, 0.0))
					sugar[death_index] += released
					_released_wealth += released
	citizens = survivors
	if repopulate_mode == "breed":
		_check_breeding(occupied)
	if repopulate_mode == "system":
		while citizens.size() < POPULATION:
			var newcomer := _spawn(occupied)
			_injected_wealth += newcomer.wealth
			replacements += 1
	tick += 1


func get_stats() -> Dictionary:
	var wealths: Array[float] = []
	var total_wealth := 0.0
	for citizen in citizens:
		wealths.append(citizen.wealth)
		total_wealth += citizen.wealth
	wealths.sort()
	var weighted := 0.0
	for index in wealths.size():
		weighted += (2.0 * (index + 1) - wealths.size() - 1.0) * wealths[index]
	var gini := 0.0
	if total_wealth > 0.0 and not wealths.is_empty():
		gini = weighted / (wealths.size() * total_wealth)
	var land_sugar := 0.0
	for amount in sugar:
		land_sugar += amount
	return {
		"population": citizens.size(),
		"total_wealth": total_wealth,
		"mean_wealth": total_wealth / citizens.size() if not citizens.is_empty() else 0.0,
		"gini": gini,
		"land_sugar": land_sugar,
		"harvested": _harvested,
		"consumed": _consumed,
		"move_cost": _move_cost_total,
		"growth": _growth,
		"injected_wealth": _injected_wealth,
		"removed_wealth": _removed_wealth,
		"released_wealth": _released_wealth,
		"bred": bred,
	}


func set_manual_control(citizen_id: int, enabled: bool) -> void:
	if enabled:
		manual_controlled[citizen_id] = true
	else:
		manual_controlled.erase(citizen_id)
		manual_commands.erase(citizen_id)


func queue_manual_move(citizen_id: int, direction: Vector2i) -> void:
	if manual_controlled.has(citizen_id):
		manual_commands[citizen_id] = direction


func _choose_manual_cell(citizen: Citizen, occupied: Dictionary) -> Vector2i:
	var direction: Vector2i = manual_commands.get(citizen.id, Vector2i.ZERO)
	if direction == Vector2i.ZERO:
		return citizen.cell
	var raw := citizen.cell + direction
	var destination := Vector2i(posmod(raw.x, GRID_SIZE), posmod(raw.y, GRID_SIZE))
	return citizen.cell if occupied.has(destination) else destination


func _choose_cell(citizen: Citizen, occupied: Dictionary) -> Vector2i:
	# Wealthy residents wander randomly instead of seeking sugar
	if citizen.wealth >= citizen.metabolism * 25.0:
		var wander_candidates: Array[Vector2i] = [citizen.cell]
		for direction: Vector2i in DIRECTIONS:
			for distance in range(1, citizen.vision + 1):
				var raw := citizen.cell + direction * distance
				var cell := Vector2i(posmod(raw.x, GRID_SIZE), posmod(raw.y, GRID_SIZE))
				if not occupied.has(cell) and not wander_candidates.has(cell):
					wander_candidates.append(cell)
		return wander_candidates[_rng.randi_range(0, wander_candidates.size() - 1)]
	var candidates: Array[Vector2i] = [citizen.cell]
	var best_sugar := sugar[_index(citizen.cell)]
	var best_distance := 0
	for direction: Vector2i in DIRECTIONS:
		for distance in range(1, citizen.vision + 1):
			var raw := citizen.cell + direction * distance
			var cell := Vector2i(posmod(raw.x, GRID_SIZE), posmod(raw.y, GRID_SIZE))
			if occupied.has(cell):
				continue
			var amount := sugar[_index(cell)]
			if amount > best_sugar or (amount == best_sugar and distance < best_distance):
				candidates.clear()
				candidates.append(cell)
				best_sugar = amount
				best_distance = distance
			elif amount == best_sugar and distance == best_distance:
				candidates.append(cell)
	return candidates[_rng.randi_range(0, candidates.size() - 1)]


func _spawn(occupied: Dictionary) -> Citizen:
	var cell := Vector2i(_rng.randi_range(0, GRID_SIZE - 1), _rng.randi_range(0, GRID_SIZE - 1))
	while occupied.has(cell):
		cell = Vector2i(_rng.randi_range(0, GRID_SIZE - 1), _rng.randi_range(0, GRID_SIZE - 1))
	var citizen := Citizen.new()
	citizen.id = _next_id
	_next_id += 1
	citizen.cell = cell
	citizen.wealth = _rng.randf_range(5.0, 25.0)
	citizen.vision = _rng.randi_range(1, 6)
	citizen.metabolism = _rng.randi_range(1, 4)
	citizen.max_age = _rng.randi_range(60, 100)
	citizens.append(citizen)
	occupied[cell] = true
	return citizen


func _check_breeding(occupied: Dictionary) -> void:
	if citizens.size() >= POPULATION:
		return
	var paired := {}
	for i in citizens.size():
		var a: Citizen = citizens[i]
		if a.death_reason != "" or paired.has(a.id) or a.wealth < a.metabolism * 5.0:
			continue
		if a.age < breed_age_min or a.age > breed_age_max:
			continue
		for j in range(i + 1, citizens.size()):
			var b: Citizen = citizens[j]
			if b.death_reason != "" or paired.has(b.id) or b.wealth < b.metabolism * 5.0:
				continue
			if b.age < breed_age_min or b.age > breed_age_max:
				continue
			if not (_in_vision(a, b) and _in_vision(b, a)):
				continue
			var child_cell := _find_birth_cell_between(a, b, occupied)
			if child_cell == Vector2i(-1, -1):
				continue
			var child := _create_child(a, b, child_cell, occupied)
			breeding_events.append({
				"parent_a": a.id, "parent_b": b.id, "child": child.id,
			})
			paired[a.id] = true
			paired[b.id] = true
			bred += 1
			if citizens.size() >= POPULATION:
				return


func _in_vision(a: Citizen, b: Citizen) -> bool:
	if a.cell == b.cell:
		return false
	for direction: Vector2i in DIRECTIONS:
		for distance in range(1, a.vision + 1):
			var raw := a.cell + direction * distance
			var cell := Vector2i(posmod(raw.x, GRID_SIZE), posmod(raw.y, GRID_SIZE))
			if cell == b.cell:
				return true
	return false


func _find_birth_cell_between(a: Citizen, b: Citizen, occupied: Dictionary) -> Vector2i:
	var candidates: Array[Vector2i] = []
	for direction: Vector2i in DIRECTIONS:
		var raw_a := a.cell + direction
		var cell_a := Vector2i(posmod(raw_a.x, GRID_SIZE), posmod(raw_a.y, GRID_SIZE))
		if not occupied.has(cell_a):
			candidates.append(cell_a)
		var raw_b := b.cell + direction
		var cell_b := Vector2i(posmod(raw_b.x, GRID_SIZE), posmod(raw_b.y, GRID_SIZE))
		if not occupied.has(cell_b) and not candidates.has(cell_b):
			candidates.append(cell_b)
	if candidates.is_empty():
		return Vector2i(-1, -1)
	var mid := _torus_midpoint(a.cell, b.cell)
	var best := candidates[0]
	var best_dist := _distance(best, mid)
	for cell in candidates:
		var d := _distance(cell, mid)
		if d < best_dist:
			best = cell
			best_dist = d
	return best


func _torus_midpoint(a: Vector2i, b: Vector2i) -> Vector2i:
	var dx := b.x - a.x
	if absi(dx) > GRID_SIZE / 2:
		dx -= signi(dx) * GRID_SIZE
	var dy := b.y - a.y
	if absi(dy) > GRID_SIZE / 2:
		dy -= signi(dy) * GRID_SIZE
	return Vector2i(posmod(a.x + dx / 2, GRID_SIZE), posmod(a.y + dy / 2, GRID_SIZE))


func _create_child(a: Citizen, b: Citizen, cell: Vector2i, occupied: Dictionary) -> Citizen:
	var contribution_a := a.wealth * 0.3
	var contribution_b := b.wealth * 0.3
	a.wealth -= contribution_a
	b.wealth -= contribution_b
	var citizen := Citizen.new()
	citizen.id = _next_id
	_next_id += 1
	citizen.cell = cell
	citizen.wealth = contribution_a + contribution_b
	citizen.vision = clampi((a.vision + b.vision) / 2 + _rng.randi_range(-1, 1), 1, 6)
	citizen.metabolism = clampi((a.metabolism + b.metabolism) / 2 + _rng.randi_range(-1, 1), 1, 4)
	citizen.max_age = clampi((a.max_age + b.max_age) / 2 + _rng.randi_range(-5, 5), 60, 100)
	citizen.generation = maxi(a.generation, b.generation) + 1
	a.children.append(citizen.id)
	b.children.append(citizen.id)
	citizen.parents = [a.id, b.id]
	citizens.append(citizen)
	occupied[cell] = true
	return citizen


func _inherit(deceased: Citizen) -> void:
	var alive_children: Array[Citizen] = []
	for child_id in deceased.children:
		for citizen in citizens:
			if citizen.id == child_id and citizen.death_reason.is_empty():
				alive_children.append(citizen)
				break
	if alive_children.is_empty():
		var death_index := _index(deceased.cell)
		var available_capacity := capacities[death_index] - sugar[death_index]
		var released := minf(deceased.wealth, maxf(available_capacity, 0.0))
		sugar[death_index] += released
		_released_wealth += released
		return
	var share := deceased.wealth / alive_children.size()
	for child in alive_children:
		child.wealth += share


func _clear_budget() -> void:
	_harvested = 0.0
	_consumed = 0.0
	_move_cost_total = 0.0
	_growth = 0.0
	_injected_wealth = 0.0
	_removed_wealth = 0.0
	_released_wealth = 0.0


func _index(cell: Vector2i) -> int:
	return cell.y * GRID_SIZE + cell.x


func _distance(a: Vector2i, b: Vector2i) -> int:
	var difference := (a - b).abs()
	return mini(difference.x, GRID_SIZE - difference.x) + mini(difference.y, GRID_SIZE - difference.y)
