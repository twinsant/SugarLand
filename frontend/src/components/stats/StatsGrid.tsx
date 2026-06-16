import { useWorld } from '../../features/world/hooks'
import { useCitizens } from '../../features/citizens/hooks'
import { giniCoefficient } from '../../lib/math'

export default function StatsGrid() {
  const { data: world } = useWorld()
  const { data: citizens } = useCitizens()

  const alive = (citizens ?? []).filter((c) => c.alive)
  const avgAge = alive.length > 0
    ? (alive.reduce((s, c) => s + c.age, 0) / alive.length).toFixed(1)
    : '-'
  const gini = alive.length > 0
    ? giniCoefficient(alive.map((c) => c.wealth))
    : NaN
  const giniStr = isNaN(gini) ? '-' : gini.toFixed(3)

  return (
    <div className="section">
      <h2>📊 实时统计</h2>
      <div className="stats">
        <div className="stat">
          <div className="label">人口</div>
          <div className="value gold">{world?.population ?? '-'}</div>
        </div>
        <div className="stat">
          <div className="label">总糖量</div>
          <div className="value green">{world ? world.total_sugar.toFixed(0) : '-'}</div>
        </div>
        <div className="stat">
          <div className="label">基尼系数</div>
          <div className="value red">{giniStr}</div>
        </div>
        <div className="stat">
          <div className="label">平均年龄</div>
          <div className="value blue">{avgAge}</div>
        </div>
      </div>
    </div>
  )
}
