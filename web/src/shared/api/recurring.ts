import { api } from './client'
import type { RecurringTransaction, RecurringListResponse } from '../types'

export function fetchRecurring(): Promise<RecurringListResponse> {
  return api.get('/v1/recurring')
}

export function createRecurring(body: {
  account_id: number
  type: string
  amount_cents: number
  currency_code: string
  category_id: number
  note: string
  frequency: string
}): Promise<RecurringTransaction> {
  return api.post('/v1/recurring', body)
}

export function updateRecurring(id: number, body: {
  account_id?: number
  type?: string
  amount_cents?: number
  currency_code?: string
  category_id?: number
  note?: string
  frequency?: string
}): Promise<RecurringTransaction> {
  return api.put(`/v1/recurring/${id}`, body)
}

export function toggleRecurring(id: number): Promise<RecurringTransaction> {
  return api.patch(`/v1/recurring/${id}/toggle`, {})
}

export function deleteRecurring(id: number): Promise<void> {
  return api.delete(`/v1/recurring/${id}`)
}
