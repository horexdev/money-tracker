import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { api } from './client'

type FetchMock = ReturnType<typeof vi.fn>

function jsonResponse(body: unknown, init?: ResponseInit): Response {
  return new Response(JSON.stringify(body), {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
}

function lastInit(fetchMock: FetchMock): RequestInit & { headers: Record<string, string> } {
  const [, init] = fetchMock.mock.calls.at(-1) ?? []
  return init as RequestInit & { headers: Record<string, string> }
}

describe('api client', () => {
  let fetchMock: FetchMock

  beforeEach(() => {
    fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.useRealTimers()
  })

  it('GET returns parsed JSON on 200', async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({ value: 42 }))
    const result = await api.get<{ value: number }>('/v1/balance')
    expect(result).toEqual({ value: 42 })
  })

  it('returns undefined on 204 No Content', async () => {
    fetchMock.mockResolvedValueOnce(new Response(null, { status: 204 }))
    const result = await api.delete<void>('/v1/things/1')
    expect(result).toBeUndefined()
  })

  it('throws with response body on non-OK status', async () => {
    fetchMock.mockResolvedValueOnce(new Response('not found', { status: 404 }))
    await expect(api.get('/v1/missing')).rejects.toThrow('not found')
  })

  it('falls back to "HTTP <status>" when body is empty', async () => {
    fetchMock.mockResolvedValueOnce(new Response('', { status: 500 }))
    await expect(api.get('/v1/boom')).rejects.toThrow('HTTP 500')
  })

  it('attaches the X-Telegram-Init-Data header', async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({}))
    await api.get('/v1/anything')
    expect(lastInit(fetchMock).headers['X-Telegram-Init-Data']).toBe('mock_init_data')
  })

  it('sends application/json by default', async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({}))
    await api.get('/v1/anything')
    expect(lastInit(fetchMock).headers['Content-Type']).toBe('application/json')
  })

  it('prefixes paths with /api', async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({}))
    await api.get('/v1/balance')
    expect(fetchMock.mock.calls[0][0]).toBe('/api/v1/balance')
  })

  it('POST serialises body and sets POST method', async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({ id: 1 }))
    await api.post('/v1/transactions', { amount_cents: 100 })
    const init = lastInit(fetchMock)
    expect(init.method).toBe('POST')
    expect(init.body).toBe(JSON.stringify({ amount_cents: 100 }))
  })

  it.each(['put', 'patch'] as const)('%s serialises body with the right method', async (verb) => {
    fetchMock.mockResolvedValueOnce(jsonResponse({}))
    await api[verb]('/v1/things/1', { foo: 'bar' })
    const init = lastInit(fetchMock)
    expect(init.method).toBe(verb.toUpperCase())
    expect(init.body).toBe(JSON.stringify({ foo: 'bar' }))
  })

  it('DELETE uses DELETE method without a body', async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({}))
    await api.delete('/v1/things/1')
    const init = lastInit(fetchMock)
    expect(init.method).toBe('DELETE')
    expect(init.body).toBeUndefined()
  })

  it('aborts the request after 15 seconds', async () => {
    vi.useFakeTimers()
    fetchMock.mockImplementation((_url: string, init?: RequestInit) =>
      new Promise((_resolve, reject) => {
        init?.signal?.addEventListener('abort', () => {
          reject(new DOMException('aborted', 'AbortError'))
        })
      }),
    )

    const promise = api.get('/v1/slow')
    promise.catch(() => undefined)
    await vi.advanceTimersByTimeAsync(15_001)
    await expect(promise).rejects.toThrow(/abort/i)
  })
})
