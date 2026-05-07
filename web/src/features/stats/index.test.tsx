import { beforeEach, describe, expect, it, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'

const { statsGetMock, statsGetRangeMock, accountsListMock } = vi.hoisted(() => ({
  statsGetMock: vi.fn(),
  statsGetRangeMock: vi.fn(),
  accountsListMock: vi.fn(),
}))

vi.mock('../../shared/api/stats', () => ({
  statsApi: {
    get: statsGetMock,
    getRange: statsGetRangeMock,
  },
}))

vi.mock('../../shared/api/accounts', () => ({
  accountsApi: {
    list: accountsListMock,
    getById: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    setDefault: vi.fn(),
    delete: vi.fn(),
    adjust: vi.fn(),
  },
}))

import { StatsPage, computeNamedPeriodRange } from './index'
import { renderWithProviders } from '../../test/render'

beforeEach(() => {
  vi.clearAllMocks()
  statsGetMock.mockResolvedValue({
    period: 'month',
    items: [
      {
        category_id: 7,
        category_name: 'Food',
        category_icon: 'fork-knife',
        category_color: '#ff5722',
        type: 'expense',
        total_cents: 122000,
        tx_count: 4,
        currency_code: 'USD',
      },
    ],
  })
  statsGetRangeMock.mockResolvedValue({ period: 'custom', items: [] })
  accountsListMock.mockResolvedValue([
    { id: 1, name: 'Main', is_default: true, currency_code: 'USD', balance_cents: 100000, type: 'cash', icon: 'wallet', color: '#000', include_in_total: true },
  ])
})

describe('StatsPage drill-down — UI affordance', () => {
  it('renders each category as an enabled button so users can drill into history', async () => {
    renderWithProviders(<StatsPage />, { route: '/stats' })

    await waitFor(() => expect(statsGetMock).toHaveBeenCalled(), { timeout: 3000 })
    const button = await waitFor(() => {
      const candidates = screen.getAllByText('Food')
      const inRow = candidates.find((el) => el.closest('button'))
      const btn = inRow?.closest('button') as HTMLButtonElement | null
      if (!btn) throw new Error('Food row button not yet rendered')
      return btn
    }, { timeout: 3000 })

    expect(button.tagName).toBe('BUTTON')
    expect(button.disabled).toBe(false)
    expect(button.type).toBe('button')
  })
})

describe('computeNamedPeriodRange', () => {
  // Anchor every assertion to a deterministic Wednesday so day-of-week math is verifiable.
  const wednesday = new Date('2026-04-15T12:00:00Z')

  it('returns today for the today period', () => {
    expect(computeNamedPeriodRange('today', 0, wednesday)).toEqual({
      from: '2026-04-15',
      to: '2026-04-15',
    })
  })

  it('returns Monday-Sunday of the current ISO week for week', () => {
    expect(computeNamedPeriodRange('week', 0, wednesday)).toEqual({
      from: '2026-04-13',
      to: '2026-04-19',
    })
  })

  it('returns the first/last day of the current month for month', () => {
    expect(computeNamedPeriodRange('month', 0, wednesday)).toEqual({
      from: '2026-04-01',
      to: '2026-04-30',
    })
  })

  it('returns the first/last day of an earlier month with a negative offset', () => {
    expect(computeNamedPeriodRange('month', -2, wednesday)).toEqual({
      from: '2026-02-01',
      to: '2026-02-28',
    })
  })

  it('returns the first/last day of the previous month for lastmonth', () => {
    expect(computeNamedPeriodRange('lastmonth', 0, wednesday)).toEqual({
      from: '2026-03-01',
      to: '2026-03-31',
    })
  })

  it('handles Sunday correctly (week starts on the previous Monday)', () => {
    const sunday = new Date('2026-04-19T12:00:00Z')
    expect(computeNamedPeriodRange('week', 0, sunday)).toEqual({
      from: '2026-04-13',
      to: '2026-04-19',
    })
  })
})
