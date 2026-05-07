import { describe, expect, it, vi } from 'vitest'
import { screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient } from '@tanstack/react-query'

import { MoneyText } from './MoneyText'
import { renderWithProviders } from '../../test/render'
import { settingsApi } from '../api/settings'
import type { UserSettings } from '../types'

const baseSettings: UserSettings = {
  base_currency: 'USD',
  display_currencies: [],
  language: 'en',
  is_admin: false,
  notify_budget_alerts: false,
  notify_recurring_reminders: false,
  notify_weekly_summary: false,
  notify_goal_milestones: false,
  theme: 'system',
  hide_amounts: false,
}

function renderMoney(ui: React.ReactElement, settings: Partial<UserSettings> = {}) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0, staleTime: 0 }, mutations: { retry: false } },
  })
  queryClient.setQueryData<UserSettings>(['settings'], { ...baseSettings, ...settings })
  return renderWithProviders(ui, { queryClient })
}

describe('MoneyText', () => {
  it('renders formatted money when not hidden', () => {
    renderMoney(<MoneyText cents={150050} currency="USD" />)
    expect(screen.getByText(/\$1,500\.50/)).toBeInTheDocument()
  })

  it('renders a •••• mask when hide_amounts is true', () => {
    renderMoney(<MoneyText cents={150050} currency="USD" />, { hide_amounts: true })
    expect(screen.getByText('••••')).toBeInTheDocument()
  })

  it('clicking the mask toggles privacy and stops propagation', async () => {
    const updateSpy = vi.spyOn(settingsApi, 'update').mockResolvedValue({ ...baseSettings })
    const parentClick = vi.fn()
    renderMoney(
      <div onClick={parentClick}>
        <MoneyText cents={5000} currency="USD" />
      </div>,
      { hide_amounts: true },
    )
    fireEvent.click(screen.getByRole('button'))
    await waitFor(() => {
      expect(updateSpy).toHaveBeenCalledWith({ hide_amounts: false })
    })
    expect(parentClick).not.toHaveBeenCalled()
    updateSpy.mockRestore()
  })

  it('passes through className when not hidden', () => {
    renderMoney(<MoneyText cents={1000} currency="USD" className="custom" />)
    const node = screen.getByText(/\$10\.00/)
    expect(node.className).toContain('custom')
  })

  it('passes through className when hidden', () => {
    renderMoney(<MoneyText cents={1000} currency="USD" className="custom" />, { hide_amounts: true })
    const node = screen.getByText('••••')
    expect(node.className).toContain('custom')
    expect(node.className).toContain('cursor-pointer')
  })
})
