import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { downloadExport } from './export'

describe('downloadExport', () => {
  let fetchMock: ReturnType<typeof vi.fn>
  let createObjectURLSpy: ReturnType<typeof vi.fn>
  let revokeObjectURLSpy: ReturnType<typeof vi.fn>

  beforeEach(() => {
    fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    createObjectURLSpy = vi.fn(() => 'blob:mock')
    revokeObjectURLSpy = vi.fn()
    URL.createObjectURL = createObjectURLSpy as typeof URL.createObjectURL
    URL.revokeObjectURL = revokeObjectURLSpy as typeof URL.revokeObjectURL
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('triggers a download with the right filename and format', async () => {
    fetchMock.mockResolvedValueOnce(new Response('csv,data\n', { status: 200 }))
    const clickSpy = vi.fn()
    const origCreateElement = document.createElement.bind(document)
    document.createElement = ((tag: string) => {
      const el = origCreateElement(tag) as HTMLElement
      if (tag === 'a') {
        ;(el as HTMLAnchorElement).click = clickSpy
      }
      return el
    }) as typeof document.createElement

    await downloadExport('2026-01-01', '2026-01-31')

    expect(fetchMock).toHaveBeenCalledTimes(1)
    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/api/v1/export?format=csv&from=2026-01-01&to=2026-01-31')
    expect((init as RequestInit).headers).toMatchObject({ 'X-Telegram-Init-Data': 'mock_init_data' })
    expect(clickSpy).toHaveBeenCalled()
    expect(createObjectURLSpy).toHaveBeenCalled()
    expect(revokeObjectURLSpy).toHaveBeenCalledWith('blob:mock')

    document.createElement = origCreateElement
  })

  it('throws on non-OK response', async () => {
    fetchMock.mockResolvedValueOnce(new Response('forbidden', { status: 403 }))
    await expect(downloadExport('2026-01-01', '2026-01-02')).rejects.toThrow('forbidden')
  })

  it('honours custom format argument', async () => {
    fetchMock.mockResolvedValueOnce(new Response('{}', { status: 200 }))
    document.createElement = ((tag: string) => {
      const el = document.createElementNS('http://www.w3.org/1999/xhtml', tag) as HTMLElement
      if (tag === 'a') {
        ;(el as HTMLAnchorElement).click = vi.fn()
      }
      return el
    }) as typeof document.createElement

    await downloadExport('2026-01-01', '2026-01-31', 'json')
    const [url] = fetchMock.mock.calls[0]
    expect(url).toContain('format=json')
  })
})
