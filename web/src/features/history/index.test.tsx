import { beforeEach, describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'

const { listMock } = vi.hoisted(() => ({ listMock: vi.fn() }))

vi.mock('../../shared/api/transactions', () => ({
  transactionsApi: {
    list: listMock,
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
  },
}))

vi.mock('../../shared/api/accounts', () => ({
  accountsApi: {
    list: vi.fn().mockResolvedValue([]),
    getById: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    setDefault: vi.fn(),
    delete: vi.fn(),
    adjust: vi.fn(),
  },
}))

vi.mock('../../shared/api/settings', () => ({
  settingsApi: {
    get: vi.fn().mockResolvedValue({
      base_currency: 'USD',
      display_currencies: [],
      language: 'en',
      is_admin: false,
      notify_budget_alerts: true,
      notify_recurring_reminders: true,
      notify_weekly_summary: false,
      notify_goal_milestones: true,
    }),
    update: vi.fn(),
    resetData: vi.fn(),
  },
}))

import { HistoryPage } from './index'
import { renderWithProviders } from '../../test/render'

beforeEach(() => {
  vi.clearAllMocks()
  listMock.mockResolvedValue({ transactions: [], total_pages: 0, current_page: 1 })
})

describe('HistoryPage URL hydration', () => {
  it('passes category_id from the URL to the transactions list query', async () => {
    renderWithProviders(<HistoryPage />, {
      route: '/history?category_id=5&from=2026-04-01&to=2026-04-30',
    })

    await waitFor(() => {
      expect(listMock).toHaveBeenCalled()
    })

    const lastCall = listMock.mock.calls.at(-1)!
    expect(lastCall[2]).toMatchObject({
      categoryId: 5,
      from: '2026-04-01',
      to: '2026-04-30',
    })
  })

  it('omits category_id when the URL has none', async () => {
    renderWithProviders(<HistoryPage />, { route: '/history' })

    await waitFor(() => {
      expect(listMock).toHaveBeenCalled()
    })

    const lastCall = listMock.mock.calls.at(-1)!
    expect(lastCall[2]).toMatchObject({ categoryId: null })
  })

  it('does not crash when the URL carries an unknown category_id (falls back to empty list)', async () => {
    renderWithProviders(<HistoryPage />, { route: '/history?category_id=9999' })

    await waitFor(() => {
      expect(listMock).toHaveBeenCalled()
    })

    const lastCall = listMock.mock.calls.at(-1)!
    expect(lastCall[2]).toMatchObject({ categoryId: 9999 })
  })

  it('ignores garbage category_id values', async () => {
    renderWithProviders(<HistoryPage />, { route: '/history?category_id=abc' })

    await waitFor(() => {
      expect(listMock).toHaveBeenCalled()
    })

    const lastCall = listMock.mock.calls.at(-1)!
    expect(lastCall[2]).toMatchObject({ categoryId: null })
  })
})
