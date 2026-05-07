import { api } from './client'
import type { UserSettings, ThemePref } from '../types'

export const settingsApi = {
  get(): Promise<UserSettings> {
    return api.get('/v1/settings')
  },
  update(patch: {
    display_currencies?: string[]
    language?: string
    theme?: ThemePref
    hide_amounts?: boolean
    notification_preferences?: {
      notify_budget_alerts?: boolean
      notify_recurring_reminders?: boolean
      notify_weekly_summary?: boolean
      notify_goal_milestones?: boolean
    }
  }): Promise<UserSettings> {
    return api.patch('/v1/settings', patch)
  },
  resetData(): Promise<void> {
    return api.delete('/v1/user/data')
  },
}
