import { api } from './client'
import type { Account } from '../types'

export const accountsApi = {
  list: (): Promise<Account[]> =>
    api.get<{ accounts: Account[] }>('/v1/accounts').then(r => r.accounts),

  getById: (id: number): Promise<Account> =>
    api.get<Account>(`/v1/accounts/${id}`),

  create: (data: {
    name: string
    icon: string
    color: string
    type: string
    currency_code: string
    include_in_total: boolean
  }): Promise<Account> =>
    api.post<Account>('/v1/accounts', data),

  update: (id: number, data: Partial<{
    name: string
    icon: string
    color: string
    type: string
    currency_code: string
    include_in_total: boolean
  }>): Promise<Account> =>
    api.put<Account>(`/v1/accounts/${id}`, data),

  setDefault: (id: number): Promise<Account> =>
    api.post<Account>(`/v1/accounts/${id}/set-default`, {}),

  delete: (id: number): Promise<void> =>
    api.delete<void>(`/v1/accounts/${id}`),

  adjust: (id: number, data: { delta_cents: number; note?: string }): Promise<import('../types').Transaction> =>
    api.post<import('../types').Transaction>(`/v1/accounts/${id}/adjust`, data),
}
