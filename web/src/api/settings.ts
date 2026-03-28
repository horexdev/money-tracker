import { api } from './client'
import type { UserSettings } from '../types'

export const settingsApi = {
  get(): Promise<UserSettings> {
    return api.get('/v1/settings')
  },
  update(patch: { base_currency?: string; display_currencies?: string[]; language?: string }): Promise<UserSettings> {
    return api.patch('/v1/settings', patch)
  },
}
