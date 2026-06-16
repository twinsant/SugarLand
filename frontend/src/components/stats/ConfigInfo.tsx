import { useWorld } from '../../features/world/hooks'

export default function ConfigInfo() {
  const { data: world } = useWorld()
  const c = world?.config

  return (
    <div className="section">
      <h2>⚙ 配置</h2>
      <div className="config-info">
        {c ? (
          <>
            网格: {c.width}×{c.height}<br />
            双峰: ({c.peak_x1},{c.peak_y1}) &amp; ({c.peak_x2},{c.peak_y2})<br />
            峰值容量: {c.peak_capacity}<br />
            初始人口: {c.init_population}<br />
            贸易: {c.enable_trading ? '✅' : '❌'}<br />
            繁殖: {c.enable_mating ? '✅' : '❌'}<br />
            污染: {c.enable_pollution ? '✅' : '❌'}
          </>
        ) : (
          <span>加载中…</span>
        )}
      </div>
    </div>
  )
}
