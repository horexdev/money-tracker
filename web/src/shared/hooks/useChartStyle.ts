import type { StatsChartStyle } from '../types'
import { useUserSettings, useUpdateSettings } from './useUserSettings'

const DEFAULT_STYLE: StatsChartStyle = 'donut'

export function useChartStyle(): [StatsChartStyle, (next: StatsChartStyle) => void] {
  const { data } = useUserSettings()
  const update = useUpdateSettings()

  const style = data?.stats_chart_style ?? DEFAULT_STYLE

  function setStyle(next: StatsChartStyle): void {
    if (next === style) return
    update.mutate({ ui_preferences: { stats_chart_style: next } })
  }

  return [style, setStyle]
}
