import { api } from './client'
import type { AdminUsersResponse, AdminStatsResponse } from '../types'

export const adminApi = {
  getUsers(page = 1, pageSize = 20): Promise<AdminUsersResponse> {
    return api.get(`/v1/admin/users?page=${page}&page_size=${pageSize}`)
  },
  getStats(): Promise<AdminStatsResponse> {
    return api.get('/v1/admin/stats')
  },
  resetUser(userID: number): Promise<{ reset: boolean; user_id: number }> {
    return api.delete(`/v1/admin/users/${userID}/data`)
  },
  resetAllUsers(): Promise<{ reset: number; failed: number }> {
    return api.delete('/v1/admin/users/data')
  },
}
