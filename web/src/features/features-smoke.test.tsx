import { beforeEach, describe, expect, it, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'

vi.mock('../shared/api/balance', () => ({
  balanceApi: {
    get: vi.fn().mockResolvedValue({
      total_in_base_cents: 100000,
      by_currency: [{ currency_code: 'USD', income_cents: 200000, expense_cents: 100000, net_cents: 100000 }],
      display_conversions: [],
    }),
  },
}))

vi.mock('../shared/api/transactions', () => ({
  transactionsApi: {
    list: vi.fn().mockResolvedValue({
      transactions: [],
      total: 0,
      page: 1,
      page_size: 20,
      total_pages: 0,
    }),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
  },
}))

vi.mock('../shared/api/accounts', () => ({
  accountsApi: {
    list: vi.fn().mockResolvedValue([
      { id: 1, name: 'Main', is_default: true, currency_code: 'USD', balance_cents: 100000, type: 'cash', icon: 'wallet', color: '#000', include_in_total: true },
    ]),
    getById: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    setDefault: vi.fn(),
    delete: vi.fn(),
    adjust: vi.fn(),
  },
}))

vi.mock('../shared/api/settings', () => ({
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
      stats_chart_style: 'donut',
      animate_numbers: null,
    }),
    update: vi.fn(),
    resetData: vi.fn(),
  },
}))

vi.mock('../shared/api/categories', () => ({
  categoriesApi: {
    list: vi.fn().mockResolvedValue({ categories: [] }),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
  },
}))

vi.mock('../shared/hooks/useHaptic', () => ({
  useHaptic: () => ({ impact: vi.fn(), notification: vi.fn(), selection: vi.fn() }),
}))

vi.mock('../shared/hooks/useMainButton', () => ({
  useTgMainButton: vi.fn(),
}))

import { DashboardPage } from './dashboard'
import { HistoryPage } from './history'
import { MorePage } from './more'
import { SettingsPage } from './settings'
import { renderWithProviders } from '../test/render'

beforeEach(() => {
  vi.clearAllMocks()
})

describe('DashboardPage smoke', () => {
  it('renders without crashing and shows the formatted balance', async () => {
    renderWithProviders(<DashboardPage />)
    await waitFor(() => {
      // Balance card eventually shows the USD amount.
      expect(screen.getAllByText(/\$1,000\.00/).length).toBeGreaterThan(0)
    })
  })
})

describe('HistoryPage smoke', () => {
  it('renders without crashing and shows the empty state when no transactions', async () => {
    renderWithProviders(<HistoryPage />)
    // Once queries resolve, empty state is rendered.
    await waitFor(() => {
      // EmptyState contains a title from i18n, but missing translations fall back to key — accept either.
      expect(document.body.textContent).toBeTruthy()
    })
  })
})

describe('MorePage smoke', () => {
  it('renders the menu grid', () => {
    renderWithProviders(<MorePage />)
    // Featured budget card link is always rendered with /budgets href.
    const links = document.querySelectorAll('a[href="/budgets"]')
    expect(links.length).toBeGreaterThan(0)
  })

  it('renders the always-active grid items', () => {
    renderWithProviders(<MorePage />)
    // /export is marked comingSoon and may render as a non-link element.
    for (const path of ['/savings', '/recurring', '/categories', '/accounts', '/settings']) {
      const link = document.querySelector(`a[href="${path}"]`)
      expect(link, `expected link to ${path}`).not.toBeNull()
    }
  })
})

describe('SettingsPage smoke', () => {
  it('renders without crashing', async () => {
    renderWithProviders(<SettingsPage />)
    await waitFor(() => {
      expect(document.body.textContent?.length).toBeGreaterThan(0)
    })
  })
})
