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
import { transactionsApi, type CreateTransactionPayload, type UpdateTransactionPayload } from './transactions'

const apiMock = api as unknown as Record<'get' | 'post' | 'put' | 'patch' | 'delete', ReturnType<typeof vi.fn>>

beforeEach(() => {
  apiMock.get.mockReset().mockResolvedValue({ items: [], total: 0 })
  apiMock.post.mockReset().mockResolvedValue({ id: 1 })
  apiMock.put.mockReset().mockResolvedValue({ id: 1 })
  apiMock.delete.mockReset().mockResolvedValue(undefined)
})

describe('transactionsApi.list', () => {
  it('uses defaults page=1, page_size=20 when no args given', async () => {
    await transactionsApi.list()
    expect(apiMock.get).toHaveBeenCalledTimes(1)
    expect(apiMock.get).toHaveBeenCalledWith('/v1/transactions?page=1&page_size=20')
  })

  it('passes through custom page and pageSize', async () => {
    await transactionsApi.list(3, 50)
    expect(apiMock.get).toHaveBeenCalledWith('/v1/transactions?page=3&page_size=50')
  })

  it('appends optional accountId, from, to filters when provided', async () => {
    await transactionsApi.list(1, 20, { accountId: 7, from: '2026-04-01', to: '2026-04-30' })
    const url = apiMock.get.mock.calls[0][0]
    expect(url).toContain('account_id=7')
    expect(url).toContain('from=2026-04-01')
    expect(url).toContain('to=2026-04-30')
  })

  it('omits filters that are null or undefined', async () => {
    await transactionsApi.list(1, 20, { accountId: null, from: null, to: null })
    const url = apiMock.get.mock.calls[0][0]
    expect(url).not.toContain('account_id')
    expect(url).not.toContain('from=')
    expect(url).not.toContain('to=')
  })
})

describe('transactionsApi.create', () => {
  it('POSTs the payload to /v1/transactions', async () => {
    const payload: CreateTransactionPayload = {
      category_id: 5,
      type: 'expense',
      amount_cents: 1500,
      account_id: 1,
    }
    await transactionsApi.create(payload)
    expect(apiMock.post).toHaveBeenCalledWith('/v1/transactions', payload)
  })
})

describe('transactionsApi.update', () => {
  it('PUTs to /v1/transactions/:id with payload', async () => {
    const payload: UpdateTransactionPayload = {
      amount_cents: 2500,
      category_id: 5,
      created_at: '2026-04-28T10:00:00Z',
    }
    await transactionsApi.update(42, payload)
    expect(apiMock.put).toHaveBeenCalledWith('/v1/transactions/42', payload)
  })
})

describe('transactionsApi.delete', () => {
  it('DELETEs /v1/transactions/:id', async () => {
    await transactionsApi.delete(42)
    expect(apiMock.delete).toHaveBeenCalledWith('/v1/transactions/42')
  })
})
