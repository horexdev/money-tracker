import { api } from './client'
import type { BalanceResponse } from '../types'

export const balanceApi = {
  get(accountId?: number | null): Promise<BalanceResponse> {
    const qs = accountId ? `?account_id=${accountId}` : ''
    return api.get(`/v1/balance${qs}`)
  },
}
