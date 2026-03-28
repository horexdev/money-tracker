import { api } from './client'
import type { Transaction, ListTransactionsResponse, TransactionType } from '../types'

export interface CreateTransactionPayload {
  category_id: number
  type: TransactionType
  amount_cents: number
  note?: string
  currency_code?: string
}

export const transactionsApi = {
  list(page = 1, pageSize = 20): Promise<ListTransactionsResponse> {
    return api.get(`/v1/transactions?page=${page}&page_size=${pageSize}`)
  },

  create(payload: CreateTransactionPayload): Promise<Transaction> {
    return api.post('/v1/transactions', payload)
  },

  delete(id: number): Promise<void> {
    return api.delete(`/v1/transactions/${id}`)
  },
}
