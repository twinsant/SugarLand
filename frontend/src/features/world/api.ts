import { fetcher } from '../../lib/fetcher'
import type { WorldState } from './types'

export const getWorld = (): Promise<WorldState> =>
  fetcher<WorldState>('/api/world')
