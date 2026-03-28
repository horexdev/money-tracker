import { api } from './client'
import type { CategoriesResponse, Category } from '../types'

export const categoriesApi = {
  list(type?: string): Promise<CategoriesResponse> {
    const params = type ? `?type=${type}` : ''
    return api.get(`/v1/categories${params}`)
  },
  create(body: { name: string; emoji: string; type: string }): Promise<Category> {
    return api.post('/v1/categories', body)
  },
  update(id: number, body: { name: string; emoji: string; type: string }): Promise<Category> {
    return api.put(`/v1/categories/${id}`, body)
  },
  delete(id: number): Promise<void> {
    return api.delete(`/v1/categories/${id}`)
  },
}
