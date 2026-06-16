const LEGEND_ITEMS = [
  { color: '#2d1b00', label: '沙漠' },
  { color: '#4a7a20', label: '低糖' },
  { color: '#8bc34a', label: '中糖' },
  { color: '#ffd700', label: '高峰' },
  { color: '#ff5722', label: '公民', circle: true },
]

export default function Legend() {
  return (
    <div className="section">
      <h2>🗺 图例</h2>
      <div className="legend">
        {LEGEND_ITEMS.map(({ color, label, circle }) => (
          <div className="legend-item" key={label}>
            <div
              className="legend-color"
              style={{ background: color, borderRadius: circle ? '50%' : undefined }}
            />
            {label}
          </div>
        ))}
      </div>
    </div>
  )
}
