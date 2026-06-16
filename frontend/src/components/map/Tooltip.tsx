import type { Cell } from '../../features/cells/types'
import type { Citizen } from '../../features/citizens/types'

interface TooltipProps {
  gridX: number
  gridY: number
  cell: Cell | null
  citizens: Citizen[]
  screenX: number
  screenY: number
}

export default function Tooltip({ gridX, gridY, cell, citizens, screenX, screenY }: TooltipProps) {
  return (
    <div
      className="tooltip"
      style={{ left: screenX + 12, top: screenY + 12 }}
    >
      <div className="tt-title">格子 ({gridX}, {gridY})</div>
      {cell && (
        <>
          <div className="tt-row">
            <span className="tt-label">糖</span>
            <span>{cell.sugar.toFixed(1)}</span>
          </div>
          <div className="tt-row">
            <span className="tt-label">容量</span>
            <span>{cell.capacity.toFixed(1)}</span>
          </div>
          <div className="tt-row">
            <span className="tt-label">污染</span>
            <span>{cell.pollution.toFixed(1)}</span>
          </div>
        </>
      )}
      {citizens.length > 0 && (
        <div style={{ marginTop: 6, color: '#ff5722', fontWeight: 'bold' }}>
          👤 {citizens.length} 个公民
        </div>
      )}
      {citizens.slice(0, 3).map((c) => (
        <div className="tt-row" key={c.id}>
          <span className="tt-label">ID {c.id}</span>
          <span>💰{c.wealth} 年龄{c.age}</span>
        </div>
      ))}
    </div>
  )
}
