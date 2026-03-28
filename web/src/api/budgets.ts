import { api } from './client'
import type { Budget, BudgetsResponse } from '../types'

export function fetchBudgets(): Promise<BudgetsResponse> {
  return api.get('/v1/budgets')
}

export function createBudget(body: {
  category_id: number
  limit_cents: number
  period: string
  currency_code: string
  notify_at_percent?: number
}): Promise<Budget> {
  return api.post('/v1/budgets', body)
}

export function updateBudget(id: number, body: {
  limit_cents?: number
  period?: string
  notify_at_percent?: number
}): Promise<Budget> {
  return api.put(`/v1/budgets/${id}`, body)
}

export function deleteBudget(id: number): Promise<void> {
  return api.delete(`/v1/budgets/${id}`)
}
