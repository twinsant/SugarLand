import { useCallback, useEffect, useRef } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useUIStore } from '../../store/uiStore'
import { stepWorld, resetWorld } from '../../features/simulation/api'
import { WORLD_QUERY_KEY } from '../../features/world/hooks'
import { CITIZENS_QUERY_KEY } from '../../features/citizens/hooks'

export default function ControlsBar() {
  const { playing, showHeat, speed, setPlaying, toggleHeat } = useUIStore()
  const queryClient = useQueryClient()
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const refreshAll = useCallback(async () => {
    await queryClient.invalidateQueries({ queryKey: WORLD_QUERY_KEY })
    await queryClient.invalidateQueries({ queryKey: CITIZENS_QUERY_KEY })
  }, [queryClient])

  const handleStep = useCallback(async () => {
    await stepWorld()
    await refreshAll()
  }, [refreshAll])

  const handleReset = useCallback(async () => {
    await resetWorld()
    await refreshAll()
  }, [refreshAll])

  // Manage auto-play interval
  useEffect(() => {
    if (playing) {
      timerRef.current = setInterval(handleStep, speed)
    } else {
      if (timerRef.current) clearInterval(timerRef.current)
    }
    return () => {
      if (timerRef.current) clearInterval(timerRef.current)
    }
  }, [playing, speed, handleStep])

  return (
    <div className="controls">
      <button
        className={`btn primary${playing ? ' active' : ''}`}
        onClick={() => setPlaying(!playing)}
      >
        {playing ? '⏸ 暂停' : '▶ 开始'}
      </button>
      <button className="btn" onClick={handleStep} disabled={playing}>
        ⏭ 单步
      </button>
      <button className="btn" onClick={handleReset}>
        🔄 重置
      </button>
      <button
        className={`btn${showHeat ? ' active' : ''}`}
        onClick={toggleHeat}
      >
        🌡 糖热力图
      </button>
    </div>
  )
}
