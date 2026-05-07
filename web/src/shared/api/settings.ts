import { api } from './client'
import type { StatsChartStyle, UserSettings } from '../types'

export interface SettingsUpdatePayload {
  display_currencies?: string[]
  language?: string
  notification_preferences?: {
    notify_budget_alerts?: boolean
    notify_recurring_reminders?: boolean
    notify_weekly_summary?: boolean
    notify_goal_milestones?: boolean
  }
  ui_preferences?: {
    stats_chart_style?: StatsChartStyle
    animate_numbers?: boolean | null
  }
}

export const settingsApi = {
  get(): Promise<UserSettings> {
    return api.get('/v1/settings')
  },
  update(patch: SettingsUpdatePayload): Promise<UserSettings> {
    return api.patch('/v1/settings', patch)
  },
  resetData(): Promise<void> {
    return api.delete('/v1/user/data')
  },
}
