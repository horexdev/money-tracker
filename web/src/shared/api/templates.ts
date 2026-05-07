import { api } from './client'
import type {
  Transaction,
  TransactionTemplate,
  TemplatesListResponse,
  CreateTemplatePayload,
  UpdateTemplatePayload,
} from '../types'

export const templatesApi = {
  list(): Promise<TemplatesListResponse> {
    return api.get('/v1/templates')
  },

  create(body: CreateTemplatePayload): Promise<TransactionTemplate> {
    return api.post('/v1/templates', body)
  },

  update(id: number, body: UpdateTemplatePayload): Promise<TransactionTemplate> {
    return api.put(`/v1/templates/${id}`, body)
  },

  delete(id: number): Promise<void> {
    return api.delete(`/v1/templates/${id}`)
  },

  apply(id: number, amountCents?: number): Promise<Transaction> {
    return api.post(`/v1/templates/${id}/apply`, amountCents != null ? { amount_cents: amountCents } : {})
  },

  reorder(orderedIds: number[]): Promise<TemplatesListResponse> {
    return api.patch('/v1/templates/reorder', { order: orderedIds })
  },
}
