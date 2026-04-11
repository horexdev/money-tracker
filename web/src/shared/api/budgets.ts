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
  notifications_enabled?: boolean
}): Promise<Budget> {
  return api.post('/v1/budgets', body)
}

export function updateBudget(id: number, body: {
  category_id?: number
  limit_cents?: number
  period?: string
  notify_at_percent?: number
  notifications_enabled?: boolean
}): Promise<Budget> {
  return api.put(`/v1/budgets/${id}`, body)
}

export function deleteBudget(id: number): Promise<void> {
  return api.delete(`/v1/budgets/${id}`)
}

export interface BudgetTransaction {
  id: number
  amount_cents: number
  category_name: string
  category_icon: string
  category_color: string
  note: string
  currency_code: string
  created_at: string
}

export interface BudgetTransactionsResponse {
  transactions: BudgetTransaction[]
}

export function fetchBudgetTransactions(id: number): Promise<BudgetTransactionsResponse> {
  return api.get(`/v1/budgets/${id}/transactions`)
}
