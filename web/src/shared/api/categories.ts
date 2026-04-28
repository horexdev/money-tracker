import { api } from './client'
import type { CategoriesResponse, Category } from '../types'

export const categoriesApi = {
  list(type?: string, order?: string): Promise<CategoriesResponse> {
    const params = new URLSearchParams()
    if (type) params.set('type', type)
    if (order) params.set('order', order)
    const qs = params.toString()
    return api.get(`/v1/categories${qs ? `?${qs}` : ''}`)
  },
  create(body: { name: string; icon: string; type: string; color: string }): Promise<Category> {
    return api.post('/v1/categories', body)
  },
  update(id: number, body: { name: string; icon: string; type: string; color: string }): Promise<Category> {
    return api.put(`/v1/categories/${id}`, body)
  },
  delete(id: number): Promise<void> {
    return api.delete(`/v1/categories/${id}`)
  },
}
