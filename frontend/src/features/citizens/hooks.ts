import { useQuery } from '@tanstack/react-query'
import { getCitizens } from './api'
import type { Citizen } from './types'

export const CITIZENS_QUERY_KEY = ['citizens'] as const

export function useCitizens() {
  return useQuery<Citizen[]>({
    queryKey: CITIZENS_QUERY_KEY,
    queryFn: getCitizens,
    refetchInterval: false,
  })
}
