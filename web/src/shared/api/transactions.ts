import { api } from './client'
import type { Transaction, ListTransactionsResponse, TransactionType } from '../types'

export interface CreateTransactionPayload {
  category_id: number
  type: TransactionType
  amount_cents: number
  note?: string
  created_at?: string
  account_id: number
}

export interface UpdateTransactionPayload {
  amount_cents: number
  category_id: number
  note?: string
  created_at: string
}

export const transactionsApi = {
  list(
    page = 1,
    pageSize = 20,
    params?: {
      accountId?: number | null
      from?: string | null
      to?: string | null
      categoryId?: number | null
    },
  ): Promise<ListTransactionsResponse> {
    const qs = new URLSearchParams({ page: String(page), page_size: String(pageSize) })
    if (params?.accountId) qs.set('account_id', String(params.accountId))
    if (params?.from) qs.set('from', params.from)
    if (params?.to) qs.set('to', params.to)
    if (params?.categoryId) qs.set('category_id', String(params.categoryId))
    return api.get(`/v1/transactions?${qs.toString()}`)
  },

  create(payload: CreateTransactionPayload): Promise<Transaction> {
    return api.post('/v1/transactions', payload)
  },

  update(id: number, payload: UpdateTransactionPayload): Promise<Transaction> {
    return api.put(`/v1/transactions/${id}`, payload)
  },

  delete(id: number): Promise<void> {
    return api.delete(`/v1/transactions/${id}`)
  },
}
