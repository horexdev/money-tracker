import { api } from './client'
import type { AdminUsersResponse, AdminStatsResponse } from '../types'

export const adminApi = {
  getUsers(page = 1, pageSize = 20): Promise<AdminUsersResponse> {
    return api.get(`/v1/admin/users?page=${page}&page_size=${pageSize}`)
  },
  getStats(): Promise<AdminStatsResponse> {
    return api.get('/v1/admin/stats')
  },
}
