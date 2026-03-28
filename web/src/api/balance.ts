import { api } from './client'
import type { BalanceResponse } from '../types'

export const balanceApi = {
  get(): Promise<BalanceResponse> {
    return api.get('/v1/balance')
  },
}
