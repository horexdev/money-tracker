import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('./client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}))

import { api } from './client'
import { templatesApi } from './templates'

const apiMock = api as unknown as Record<'get' | 'post' | 'put' | 'patch' | 'delete', ReturnType<typeof vi.fn>>

beforeEach(() => {
  apiMock.get.mockReset().mockResolvedValue({ templates: [] })
  apiMock.post.mockReset().mockResolvedValue({ id: 1 })
  apiMock.put.mockReset().mockResolvedValue({ id: 1 })
  apiMock.patch.mockReset().mockResolvedValue({ templates: [] })
  apiMock.delete.mockReset().mockResolvedValue(undefined)
})

describe('templatesApi.list', () => {
  it('GETs /v1/templates', async () => {
    await templatesApi.list()
    expect(apiMock.get).toHaveBeenCalledWith('/v1/templates')
  })
})

describe('templatesApi.create', () => {
  it('POSTs body to /v1/templates', async () => {
    await templatesApi.create({
      type: 'expense', amount_cents: 30000, amount_fixed: true,
      category_id: 2, account_id: 7, name: 'Coffee',
    })
    expect(apiMock.post).toHaveBeenCalledWith('/v1/templates', {
      type: 'expense', amount_cents: 30000, amount_fixed: true,
      category_id: 2, account_id: 7, name: 'Coffee',
    })
  })
})

describe('templatesApi.update', () => {
  it('PUTs partial body to /v1/templates/:id', async () => {
    await templatesApi.update(42, { name: 'Renamed' })
    expect(apiMock.put).toHaveBeenCalledWith('/v1/templates/42', { name: 'Renamed' })
  })
})

describe('templatesApi.delete', () => {
  it('DELETEs /v1/templates/:id', async () => {
    await templatesApi.delete(7)
    expect(apiMock.delete).toHaveBeenCalledWith('/v1/templates/7')
  })
})

describe('templatesApi.apply', () => {
  it('POSTs empty body to /v1/templates/:id/apply for fixed amount', async () => {
    await templatesApi.apply(5)
    expect(apiMock.post).toHaveBeenCalledWith('/v1/templates/5/apply', {})
  })

  it('POSTs amount override for variable amount', async () => {
    await templatesApi.apply(5, 75000)
    expect(apiMock.post).toHaveBeenCalledWith('/v1/templates/5/apply', { amount_cents: 75000 })
  })
})

describe('templatesApi.reorder', () => {
  it('PATCHes order array to /v1/templates/reorder', async () => {
    await templatesApi.reorder([3, 1, 2])
    expect(apiMock.patch).toHaveBeenCalledWith('/v1/templates/reorder', { order: [3, 1, 2] })
  })
})
