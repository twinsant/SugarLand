import StatsGrid from '../stats/StatsGrid'
import WealthChart from '../stats/WealthChart'
import Legend from '../stats/Legend'
import CitizenTopList from '../stats/CitizenTopList'
import ConfigInfo from '../stats/ConfigInfo'

export default function InfoPanel() {
  return (
    <div className="info-panel">
      <StatsGrid />
      <WealthChart />
      <Legend />
      <CitizenTopList />
      <ConfigInfo />
    </div>
  )
}
