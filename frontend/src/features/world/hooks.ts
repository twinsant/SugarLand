import { useQuery } from '@tanstack/react-query'
import { getWorld } from './api'
import type { WorldState } from './types'

export const WORLD_QUERY_KEY = ['world'] as const

export function useWorld() {
  return useQuery<WorldState>({
    queryKey: WORLD_QUERY_KEY,
    queryFn: getWorld,
    // Polling is handled manually in the simulation loop; disable auto-refetch
    refetchInterval: false,
  })
}
