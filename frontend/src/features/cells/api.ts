import { fetcher } from '../../lib/fetcher'
import type { Cell } from './types'

export const getCell = (x: number, y: number): Promise<Cell> =>
  fetcher<Cell>(`/api/cells/${x}/${y}`)
