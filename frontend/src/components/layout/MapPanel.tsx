import { useWorld } from '../../features/world/hooks'
import ControlsBar from '../controls/ControlsBar'
import SpeedControl from '../controls/SpeedControl'
import WorldCanvas from '../map/WorldCanvas'

export default function MapPanel() {
  const { data: world } = useWorld()
  const timestep = world?.timestep ?? 0

  return (
    <div className="map-panel">
      <div className="map-header">
        <h1>🌍 SugarLand 元宇宙25号</h1>
        <span className="timestep">T = {timestep}</span>
      </div>
      <ControlsBar />
      <SpeedControl />
      <WorldCanvas />
    </div>
  )
}
