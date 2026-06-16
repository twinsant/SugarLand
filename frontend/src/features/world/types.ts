// Types matching the Go WorldState and WorldConfig structs

export interface WorldConfig {
  width: number
  height: number
  peak_x1: number
  peak_y1: number
  peak_x2: number
  peak_y2: number
  peak_capacity: number
  growth_rate: number
  init_population: number
  enable_trading: boolean
  enable_mating: boolean
  enable_pollution: boolean
  season_interval: number
}

export interface WorldState {
  config: WorldConfig
  timestep: number
  population: number
  total_sugar: number
  total_cells: number
}
