import { beforeEach, describe, expect, it, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

vi.mock('../../shared/api/stats', () => ({
  statsApi: {
    get: vi.fn(),
    getRange: vi.fn(),
  },
}))

vi.mock('../../shared/api/accounts', () => ({
  accountsApi: {
    list: vi.fn(),
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
    get: vi.fn(),
    update: vi.fn(),
    resetData: vi.fn(),
  },
}))

vi.mock('framer-motion', async () => {
  const actual = await vi.importActual<Record<string, unknown>>('framer-motion')
  return {
    ...actual,
    useReducedMotion: vi.fn(() => false),
  }
})

import { StatsPage } from './index'
import { statsApi } from '../../shared/api/stats'
import { accountsApi } from '../../shared/api/accounts'
import { settingsApi } from '../../shared/api/settings'
import { renderWithProviders } from '../../test/render'
import type { StatsChartStyle, StatsResponse, UserSettings } from '../../shared/types'

const baseSettings: UserSettings = {
  base_currency: 'USD',
  display_currencies: [],
  language: 'en',
  is_admin: false,
  notify_budget_alerts: false,
  notify_recurring_reminders: false,
  notify_weekly_summary: false,
  notify_goal_milestones: false,
  stats_chart_style: 'donut',
  animate_numbers: false,
}

const sampleStats: StatsResponse = {
  period: 'month',
  items: [
    { category_id: 1, category_name: 'Food',      category_icon: 'fork-knife', category_color: '#f97316', type: 'expense', total_cents: 5000, tx_count: 4, currency_code: 'USD' },
    { category_id: 2, category_name: 'Transport', category_icon: 'car',        category_color: '#8b5cf6', type: 'expense', total_cents: 3000, tx_count: 2, currency_code: 'USD' },
    { category_id: 3, category_name: 'Salary',    category_icon: 'briefcase',  category_color: '#22c55e', type: 'income',  total_cents: 9000, tx_count: 1, currency_code: 'USD' },
  ],
}

beforeEach(() => {
  vi.mocked(statsApi.get).mockResolvedValue(sampleStats)
  vi.mocked(statsApi.getRange).mockResolvedValue(sampleStats)
  vi.mocked(accountsApi.list).mockResolvedValue([
    { id: 1, name: 'Main', is_default: true, currency_code: 'USD', balance_cents: 100000, type: 'cash', icon: 'wallet', color: '#000', include_in_total: true } as never,
  ])
  vi.mocked(settingsApi.get).mockResolvedValue(baseSettings)
  vi.mocked(settingsApi.update).mockReset()
})

function renderWith(style: StatsChartStyle) {
  vi.mocked(settingsApi.get).mockResolvedValue({ ...baseSettings, stats_chart_style: style })
  return renderWithProviders(<StatsPage />)
}

describe('StatsPage chart style', () => {
  it('renders the donut chart by default', async () => {
    renderWith('donut')
    await waitFor(() => {
      expect(document.querySelector('svg')).not.toBeNull()
    })
  })

  it('renders the stacked bar chart when chart_style is stacked_bar', async () => {
    renderWith('stacked_bar')
    await waitFor(() => {
      expect(screen.getByTestId('chart-stacked-bar')).toBeTruthy()
    })
  })

  it('renders the dual bar chart when chart_style is dual_bar', async () => {
    renderWith('dual_bar')
    await waitFor(() => {
      expect(screen.getByTestId('chart-dual-bar')).toBeTruthy()
    })
  })

  it('renders the profit bars chart when chart_style is profit_bars', async () => {
    renderWith('profit_bars')
    await waitFor(() => {
      expect(screen.getByTestId('chart-profit-bars')).toBeTruthy()
    })
  })

  it('hides the expense/income toggle button in dual_bar mode', async () => {
    renderWith('dual_bar')
    await waitFor(() => {
      expect(screen.getByTestId('chart-dual-bar')).toBeTruthy()
    })
    // The labels "Expense" / "Income" still appear inside the chart and breakdown
    // lists, so we specifically assert the togglable buttons are absent.
    expect(screen.queryByRole('button', { name: 'Expense' })).toBeNull()
    expect(screen.queryByRole('button', { name: 'Income' })).toBeNull()
  })

  it('hides the expense/income toggle button in profit_bars mode', async () => {
    renderWith('profit_bars')
    await waitFor(() => {
      expect(screen.getByTestId('chart-profit-bars')).toBeTruthy()
    })
    expect(screen.queryByRole('button', { name: 'Expense' })).toBeNull()
    expect(screen.queryByRole('button', { name: 'Income' })).toBeNull()
  })

  it('shows the toggle in donut and stacked_bar modes', async () => {
    renderWith('stacked_bar')
    await waitFor(() => {
      expect(screen.getByTestId('chart-stacked-bar')).toBeTruthy()
    })
    expect(screen.getByRole('button', { name: 'Expense' })).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Income' })).toBeTruthy()
  })

  it('clicking a switcher pill calls settingsApi.update with the new style', async () => {
    vi.mocked(settingsApi.update).mockResolvedValue({ ...baseSettings, stats_chart_style: 'profit_bars' })
    renderWith('donut')
    await waitFor(() => {
      expect(screen.getByTestId('chart-style-profit_bars')).toBeTruthy()
    })

    const user = userEvent.setup()
    await user.click(screen.getByTestId('chart-style-profit_bars'))

    await waitFor(() => {
      expect(settingsApi.update).toHaveBeenCalledWith({
        ui_preferences: { stats_chart_style: 'profit_bars' },
      })
    })
  })
})
