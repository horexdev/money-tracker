import { describe, expect, it, vi } from 'vitest'
import { screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient } from '@tanstack/react-query'

import { AmountDisplay } from './AmountDisplay'
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

function renderAmount(
  ui: React.ReactElement,
  settings: Partial<UserSettings> = {},
) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0, staleTime: 0 }, mutations: { retry: false } },
  })
  queryClient.setQueryData<UserSettings>(['settings'], { ...baseSettings, ...settings })
  return renderWithProviders(ui, { queryClient })
}

describe('AmountDisplay', () => {
  it('renders the formatted amount with USD by default', () => {
    renderAmount(<AmountDisplay cents={150050} />)
    expect(screen.getByText(/\$1,500\.50/)).toBeInTheDocument()
  })

  it('uses the absolute value so negative cents render unsigned by default', () => {
    renderAmount(<AmountDisplay cents={-150050} />)
    const node = screen.getByText(/\$1,500\.50/)
    expect(node.textContent ?? '').not.toContain('-')
    expect(node.textContent ?? '').not.toContain('−')
  })

  it('renders a + prefix for income when showSign is on', () => {
    renderAmount(<AmountDisplay cents={1000} type="income" showSign />)
    const node = screen.getByText(/\$10\.00/)
    expect(node.textContent?.startsWith('+')).toBe(true)
  })

  it('renders the unicode minus for expense when showSign is on', () => {
    renderAmount(<AmountDisplay cents={1000} type="expense" showSign />)
    const node = screen.getByText(/\$10\.00/)
    expect(node.textContent?.startsWith('−')).toBe(true)
  })

  it('omits any sign prefix when showSign is off', () => {
    renderAmount(<AmountDisplay cents={1000} type="income" showSign={false} />)
    const node = screen.getByText(/\$10\.00/)
    expect(node.textContent?.startsWith('+')).toBe(false)
    expect(node.textContent?.startsWith('-')).toBe(false)
    expect(node.textContent?.startsWith('−')).toBe(false)
  })

  it.each([
    ['income', 'text-income'],
    ['expense', 'text-expense'],
    ['neutral', 'text-text'],
  ] as const)('applies the colour class for type=%s', (type, cls) => {
    renderAmount(<AmountDisplay cents={100} type={type} />)
    const node = screen.getByText(/\$1\.00/)
    expect(node.className).toContain(cls)
  })

  it.each([
    ['sm', 'text-sm'],
    ['md', 'text-base'],
    ['lg', 'text-2xl'],
    ['xl', 'text-4xl'],
  ] as const)('applies the size class for size=%s', (size, cls) => {
    renderAmount(<AmountDisplay cents={100} size={size} />)
    const node = screen.getByText(/\$1\.00/)
    expect(node.className).toContain(cls)
  })

  it('respects a non-USD currency prop', () => {
    renderAmount(<AmountDisplay cents={1000} currency="EUR" />)
    expect(screen.getByText(/€10\.00/)).toBeInTheDocument()
  })

  it('appends a custom className', () => {
    renderAmount(<AmountDisplay cents={100} className="custom-class" />)
    const node = screen.getByText(/\$1\.00/)
    expect(node.className).toContain('custom-class')
  })

  describe('privacy mode', () => {
    it('renders a •••• mask when hide_amounts=true', () => {
      renderAmount(<AmountDisplay cents={1000} />, { hide_amounts: true })
      expect(screen.getByText('••••')).toBeInTheDocument()
      expect(screen.queryByText(/\$10\.00/)).toBeNull()
    })

    it('mask is rendered as a button with an aria-label', () => {
      renderAmount(<AmountDisplay cents={1000} />, { hide_amounts: true })
      const node = screen.getByRole('button')
      expect(node.textContent).toBe('••••')
      expect(node.getAttribute('aria-label')).toBeTruthy()
    })

    it('clicking the mask invokes settingsApi.update with hide_amounts:false and stops propagation', async () => {
      const updateSpy = vi.spyOn(settingsApi, 'update').mockResolvedValue({ ...baseSettings })
      const parentClick = vi.fn()
      renderAmount(
        <div onClick={parentClick}>
          <AmountDisplay cents={1000} />
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
  })
})
