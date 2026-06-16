import { useEffect, useRef } from 'react'
import { useCitizens } from '../../features/citizens/hooks'
import { buildHistogram } from '../../lib/math'
import { binToColor } from '../../lib/colors'

const BINS = 20

export default function WealthChart() {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const { data: citizens } = useCitizens()

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const alive = (citizens ?? []).filter((c) => c.alive)
    const wealths = alive.map((c) => c.wealth)
    const histogram = buildHistogram(wealths, BINS)
    const maxCount = Math.max(...histogram, 1)

    const W = canvas.offsetWidth
    const H = canvas.offsetHeight
    canvas.width = W
    canvas.height = H
    ctx.clearRect(0, 0, W, H)

    const barW = W / BINS - 1
    histogram.forEach((count, i) => {
      const barH = (count / maxCount) * (H - 10)
      const x = i * (barW + 1)
      const y = H - barH
      ctx.fillStyle = binToColor(i)
      ctx.fillRect(x, y, barW, barH)
    })
  }, [citizens])

  return (
    <div className="section">
      <h2>📈 财富分布</h2>
      <div className="chart-container">
        <canvas ref={canvasRef} className="chart-canvas" />
      </div>
    </div>
  )
}
