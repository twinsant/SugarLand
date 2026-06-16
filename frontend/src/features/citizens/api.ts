import { fetcher } from '../../lib/fetcher'
import type { Citizen } from './types'

export const getCitizens = (): Promise<Citizen[]> =>
  fetcher<Citizen[]>('/api/citizens')
