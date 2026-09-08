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

var citizens: Array[Citizen] = []
var sugar := PackedFloat32Array()
var capacities := PackedFloat32Array()
var tick: int = 0
var starved: int = 0
var aged_out: int = 0
var replacements: int = 0
var harvested_by_cell: Dictionary = {}

var _rng := RandomNumberGenerator.new()
var _next_id: int = 1
var _harvested: float = 0.0
var _consumed: float = 0.0
var _growth: float = 0.0
var _injected_wealth: float = 0.0
var _removed_wealth: float = 0.0


func reset(seed_value: int = 25) -> void:
	_rng.seed = seed_value
	_next_id = 1
	tick = 0
	starved = 0
	aged_out = 0
	replacements = 0
	harvested_by_cell.clear()
	_clear_budget()
	citizens.clear()
	capacities.resize(GRID_SIZE * GRID_SIZE)
	for y in GRID_SIZE:
		for x in GRID_SIZE:
			var cell := Vector2i(x, y)
			var distance := mini(_distance(cell, Vector2i(15, 15)), _distance(cell, Vector2i(35, 35)))
			capacities[_index(cell)] = maxf(0.0, 4.0 - 0.16 * distance)
	sugar = capacities.duplicate()
	var occupied := {}
	for index in POPULATION:
		_spawn(occupied)


func step() -> void:
	_clear_budget()
	harvested_by_cell.clear()
	for index in sugar.size():
		var before := sugar[index]
		sugar[index] = minf(capacities[index], before + 1.0)
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
		var destination := _choose_cell(citizen, occupied)
		occupied.erase(citizen.cell)
		citizen.cell = destination
		occupied[destination] = true
		var index := _index(destination)
		citizen.last_harvest = sugar[index]
		sugar[index] = 0.0
		citizen.wealth += citizen.last_harvest
		citizen.wealth -= citizen.metabolism
		citizen.age += 1
		harvested_by_cell[index] = citizen.last_harvest
		_harvested += citizen.last_harvest
		_consumed += citizen.metabolism
		if citizen.wealth <= 0.0:
			citizen.death_reason = "starvation"
			starved += 1
		elif citizen.age >= citizen.max_age:
			citizen.death_reason = "age"
			aged_out += 1
	var survivors: Array[Citizen] = []
	for citizen in citizens:
		if citizen.death_reason.is_empty():
			survivors.append(citizen)
		else:
			occupied.erase(citizen.cell)
			_removed_wealth += citizen.wealth
	citizens = survivors
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
		"growth": _growth,
		"injected_wealth": _injected_wealth,
		"removed_wealth": _removed_wealth,
	}


func _choose_cell(citizen: Citizen, occupied: Dictionary) -> Vector2i:
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


func _clear_budget() -> void:
	_harvested = 0.0
	_consumed = 0.0
	_growth = 0.0
	_injected_wealth = 0.0
	_removed_wealth = 0.0


func _index(cell: Vector2i) -> int:
	return cell.y * GRID_SIZE + cell.x


func _distance(a: Vector2i, b: Vector2i) -> int:
	var difference := (a - b).abs()
	return mini(difference.x, GRID_SIZE - difference.x) + mini(difference.y, GRID_SIZE - difference.y)
