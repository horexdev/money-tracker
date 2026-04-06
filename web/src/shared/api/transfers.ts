import { api } from './client'
import type { Transfer } from '../types'

export const exchangeApi = {
  getRate: (from: string, to: string): Promise<{ rate: number }> =>
    api.get<{ rate: number }>(`/v1/exchange/rate?from=${from}&to=${to}`),
}

export interface TransfersListResponse {
  transfers: Transfer[]
  total: number
  limit: number
  offset: number
}

export const transfersApi = {
  list: (params?: { limit?: number; offset?: number }): Promise<TransfersListResponse> => {
    const q = new URLSearchParams()
    if (params?.limit) q.set('limit', String(params.limit))
    if (params?.offset) q.set('offset', String(params.offset))
    const qs = q.toString()
    return api.get<TransfersListResponse>(`/v1/transfers${qs ? '?' + qs : ''}`)
  },

  create: (data: {
    from_account_id: number
    to_account_id: number
    amount_cents: number
    exchange_rate?: number
    note?: string
  }): Promise<Transfer> =>
    api.post<Transfer>('/v1/transfers', data),

  delete: (id: number): Promise<void> =>
    api.delete<void>(`/v1/transfers/${id}`),
}
