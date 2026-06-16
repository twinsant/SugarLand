import MapPanel from './MapPanel'
import InfoPanel from './InfoPanel'

export default function AppLayout() {
  return (
    <div className="app">
      <MapPanel />
      <InfoPanel />
    </div>
  )
}
