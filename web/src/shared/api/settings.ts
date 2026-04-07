import { api } from './client'
import type { UserSettings } from '../types'

export const settingsApi = {
  get(): Promise<UserSettings> {
    return api.get('/v1/settings')
  },
  update(patch: {
    base_currency?: string
    display_currencies?: string[]
    language?: string
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
