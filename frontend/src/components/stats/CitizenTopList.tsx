import { useCitizens } from '../../features/citizens/hooks'

export default function CitizenTopList() {
  const { data: citizens } = useCitizens()

  const top10 = [...(citizens ?? [])]
    .filter((c) => c.alive)
    .sort((a, b) => b.wealth - a.wealth)
    .slice(0, 10)

  return (
    <div className="section">
      <h2>👥 公民 TOP 10（财富）</h2>
      <div className="citizen-list">
        <div className="citizen-row header">
          <span>ID</span><span>X</span><span>Y</span><span>年龄</span><span>💰财富</span>
        </div>
        {top10.map((c) => (
          <div className="citizen-row" key={c.id}>
            <span>{c.id}</span>
            <span>{c.x}</span>
            <span>{c.y}</span>
            <span>{c.age}</span>
            <span style={{ color: '#ffd700' }}>{c.wealth}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
