import { useUIStore } from '../../store/uiStore'

export default function SpeedControl() {
  const { speed, setSpeed } = useUIStore()

  return (
    <div className="speed-control">
      <label>速度</label>
      <input
        type="range"
        min={100}
        max={3000}
        value={speed}
        onChange={(e) => setSpeed(Number(e.target.value))}
      />
      <span>{speed}ms</span>
    </div>
  )
}
