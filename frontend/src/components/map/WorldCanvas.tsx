import { useEffect, useRef, useCallback, useState } from 'react'
import { useWorld } from '../../features/world/hooks'
import { useCitizens } from '../../features/citizens/hooks'
import { useUIStore } from '../../store/uiStore'
import { getCell } from '../../features/cells/api'
import { sugarToColor, estimatedCellColor, wealthToColor } from '../../lib/colors'
import { torusWrap } from '../../lib/math'
import type { Cell } from '../../features/cells/types'
import type { Citizen } from '../../features/citizens/types'
import Tooltip from './Tooltip'

interface TooltipState {
  x: number
  y: number
  cell: Cell | null
  citizens: Citizen[]
  screenX: number
  screenY: number
}

export default function WorldCanvas() {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const { data: world } = useWorld()
  const { data: rawCitizens } = useCitizens()
  const { showHeat } = useUIStore()

  const cellCacheRef = useRef<Record<string, Cell>>({})
  const [tooltip, setTooltip] = useState<TooltipState | null>(null)

  const config = world?.config ?? {
    width: 50, height: 50,
    peak_x1: 15, peak_y1: 15,
    peak_x2: 35, peak_y2: 35,
    peak_capacity: 4, growth_rate: 1, init_population: 400,
    enable_trading: true, enable_mating: true,
    enable_pollution: false, season_interval: 100,
  }

  // Normalize citizen coordinates with torus wrapping
  const citizens: Citizen[] = (rawCitizens ?? [])
    .filter((c) => c.alive)
    .map((c) => ({
      ...c,
      x: torusWrap(c.x, config.width),
      y: torusWrap(c.y, config.height),
    }))

  // Batch-fetch cell samples and populate cache
  const fetchCellsBatch = useCallback(async () => {
    const promises: Promise<void>[] = []
    for (let y = 0; y < config.height; y += 2) {
      for (let x = 0; x < config.width; x += 2) {
        const key = `${x},${y}`
        if (!cellCacheRef.current[key]) {
          promises.push(
            getCell(x, y).then((cell) => {
              cellCacheRef.current[key] = cell
            }).catch(() => undefined),
          )
        }
      }
    }
    await Promise.all(promises)
  }, [config.width, config.height])

  // Render world onto canvas
  const render = useCallback(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const W = canvas.width
    const H = canvas.height
    const cellW = W / config.width
    const cellH = H / config.height

    ctx.fillStyle = '#0a0a1a'
    ctx.fillRect(0, 0, W, H)

    // Sugar terrain
    if (showHeat) {
      for (let y = 0; y < config.height; y++) {
        for (let x = 0; x < config.width; x++) {
          const cell = cellCacheRef.current[`${x},${y}`]
          ctx.fillStyle = cell
            ? sugarToColor(cell.sugar, cell.capacity)
            : estimatedCellColor(x, y, config.peak_x1, config.peak_y1, config.peak_x2, config.peak_y2)
          ctx.fillRect(x * cellW, y * cellH, cellW + 0.5, cellH + 0.5)
        }
      }
    }

    // Peak markers
    ctx.strokeStyle = '#ff660044'
    ctx.lineWidth = 1
    const peaks = [
      { px: config.peak_x1, py: config.peak_y1 },
      { px: config.peak_x2, py: config.peak_y2 },
    ]
    for (const { px, py } of peaks) {
      ctx.beginPath()
      ctx.arc(px * cellW + cellW / 2, py * cellH + cellH / 2, 15, 0, Math.PI * 2)
      ctx.stroke()
    }

    // Citizens
    for (const c of citizens) {
      const cx = c.x * cellW + cellW / 2
      const cy = c.y * cellH + cellH / 2
      const size = Math.max(2, Math.min(6, c.wealth / 15))

      ctx.beginPath()
      ctx.arc(cx, cy, size + 2, 0, Math.PI * 2)
      ctx.fillStyle = 'rgba(255, 87, 34, 0.2)'
      ctx.fill()

      ctx.beginPath()
      ctx.arc(cx, cy, size, 0, Math.PI * 2)
      ctx.fillStyle = wealthToColor(c.wealth)
      ctx.fill()
    }

    // Subtle grid lines
    if (cellW > 4) {
      ctx.strokeStyle = 'rgba(255,255,255,0.03)'
      ctx.lineWidth = 0.5
      for (let x = 0; x <= config.width; x += 5) {
        ctx.beginPath()
        ctx.moveTo(x * cellW, 0)
        ctx.lineTo(x * cellW, H)
        ctx.stroke()
      }
      for (let y = 0; y <= config.height; y += 5) {
        ctx.beginPath()
        ctx.moveTo(0, y * cellH)
        ctx.lineTo(W, y * cellH)
        ctx.stroke()
      }
    }
  }, [config, showHeat, citizens])

  // Resize canvas to fill available space
  const resize = useCallback(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const panel = canvas.parentElement
    if (!panel) return
    const availW = panel.offsetWidth - 24
    const availH = panel.offsetHeight - 24
    const size = Math.min(availW, availH)
    canvas.width = size
    canvas.height = size
    render()
  }, [render])

  // Initial fetch + render
  useEffect(() => {
    fetchCellsBatch().then(() => render())
  }, [fetchCellsBatch, render])

  useEffect(() => {
    render()
  }, [render])

  useEffect(() => {
    resize()
    window.addEventListener('resize', resize)
    return () => window.removeEventListener('resize', resize)
  }, [resize])

  // Mouse tooltip
  const handleMouseMove = useCallback(
    (e: React.MouseEvent<HTMLCanvasElement>) => {
      const canvas = canvasRef.current
      if (!canvas) return
      const rect = canvas.getBoundingClientRect()
      const gx = Math.floor((e.clientX - rect.left) / (rect.width / config.width))
      const gy = Math.floor((e.clientY - rect.top) / (rect.height / config.height))
      const cell = cellCacheRef.current[`${gx},${gy}`] ?? null
      const citizensHere = citizens.filter((c) => c.x === gx && c.y === gy)
      if (cell || citizensHere.length > 0) {
        setTooltip({ x: gx, y: gy, cell, citizens: citizensHere, screenX: e.clientX, screenY: e.clientY })
      } else {
        setTooltip(null)
      }
    },
    [config.width, config.height, citizens],
  )

  return (
    <>
      <canvas
        ref={canvasRef}
        className="map-canvas"
        onMouseMove={handleMouseMove}
        onMouseLeave={() => setTooltip(null)}
      />
      {tooltip && (
        <Tooltip
          gridX={tooltip.x}
          gridY={tooltip.y}
          cell={tooltip.cell}
          citizens={tooltip.citizens}
          screenX={tooltip.screenX}
          screenY={tooltip.screenY}
        />
      )}
    </>
  )
}
