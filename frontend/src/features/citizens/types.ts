// Type matching the Go Citizen struct (only fields used in the frontend)

export interface Citizen {
  id: number
  x: number
  y: number
  age: number
  max_age: number
  wealth: number
  vision: number
  metabolism: number
  alive: boolean
  agent_name: string
  is_agent_controlled: boolean
  behavior_type: string
}
