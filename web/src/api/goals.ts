import { api } from './client'
import type { SavingsGoal, GoalsResponse } from '../types'

export function fetchGoals(): Promise<GoalsResponse> {
  return api.get('/v1/goals')
}

export function createGoal(body: {
  name: string
  target_cents: number
  currency_code: string
  deadline?: string
}): Promise<SavingsGoal> {
  return api.post('/v1/goals', body)
}

export function updateGoal(id: number, body: {
  name?: string
  target_cents?: number
  deadline?: string
}): Promise<SavingsGoal> {
  return api.put(`/v1/goals/${id}`, body)
}

export function depositGoal(id: number, amount_cents: number): Promise<SavingsGoal> {
  return api.post(`/v1/goals/${id}/deposit`, { amount_cents })
}

export function withdrawGoal(id: number, amount_cents: number): Promise<SavingsGoal> {
  return api.post(`/v1/goals/${id}/withdraw`, { amount_cents })
}

export function deleteGoal(id: number): Promise<void> {
  return api.delete(`/v1/goals/${id}`)
}
