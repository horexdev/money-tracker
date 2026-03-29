import { api } from './client'
import type { StatsResponse } from '../types'

export const statsApi = {
  get(period: 'month' | 'week' | 'today' | 'lastmonth' = 'month'): Promise<StatsResponse> {
    return api.get(`/v1/stats?period=${period}`)
  },
  getRange(from: string, to: string): Promise<StatsResponse> {
    return api.get(`/v1/stats?from=${from}&to=${to}`)
  },
}
