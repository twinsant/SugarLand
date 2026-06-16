import { poster } from '../../lib/fetcher'
import type { WorldState } from '../world/types'

export const stepWorld = (): Promise<WorldState> =>
  poster<WorldState>('/api/world/step')

export const resetWorld = (): Promise<WorldState> =>
  poster<WorldState>('/api/world/reset')
