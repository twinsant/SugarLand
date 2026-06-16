import { create } from 'zustand'

interface UIState {
  /** Whether the simulation auto-play loop is active */
  playing: boolean
  /** Auto-step interval in milliseconds */
  speed: number
  /** Whether to render the sugar heat-map layer */
  showHeat: boolean

  setPlaying: (v: boolean) => void
  setSpeed: (v: number) => void
  toggleHeat: () => void
}

export const useUIStore = create<UIState>((set) => ({
  playing: false,
  speed: 500,
  showHeat: true,

  setPlaying: (v) => set({ playing: v }),
  setSpeed: (v) => set({ speed: v }),
  toggleHeat: () => set((s) => ({ showHeat: !s.showHeat })),
}))
