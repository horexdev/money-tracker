import { api } from './client'
import type { StatsResponse } from '../types'

export const statsApi = {
  get(period: 'month' | 'week' | 'today' | 'lastmonth' = 'month', accountId?: number | null): Promise<StatsResponse> {
    const qs = new URLSearchParams({ period })
    if (accountId) qs.set('account_id', String(accountId))
    return api.get(`/v1/stats?${qs}`)
  },
  getRange(from: string, to: string, accountId?: number | null): Promise<StatsResponse> {
    const qs = new URLSearchParams({ from, to })
    if (accountId) qs.set('account_id', String(accountId))
    return api.get(`/v1/stats?${qs}`)
  },
}
