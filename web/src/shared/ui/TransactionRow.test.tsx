import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { I18nextProvider, initReactI18next } from 'react-i18next'
import i18n from 'i18next'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

import { TransactionRow } from './TransactionRow'
import type { Transaction, UserSettings } from '../types'

const SETTINGS_VISIBLE: UserSettings = {
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

function withI18n(ui: React.ReactElement, lang = 'en') {
  const inst = i18n.createInstance()
  inst.use(initReactI18next).init({
    lng: lang,
    fallbackLng: 'en',
    resources: {
      en: {
        translation: {
          common: { delete: 'Delete', show_amounts: 'Show amounts' },
          categories: { names: { Food: 'Food' } },
        },
      },
    },
    interpolation: { escapeValue: false },
    react: { useSuspense: false },
  })
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0, staleTime: 0 } },
  })
  client.setQueryData<UserSettings>(['settings'], SETTINGS_VISIBLE)
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <I18nextProvider i18n={inst}>
        <QueryClientProvider client={client}>{children}</QueryClientProvider>
      </I18nextProvider>
    )
  }
  return render(ui, { wrapper: Wrapper })
}

const baseTx: Transaction = {
  id: 1,
  type: 'expense',
  amount_cents: 1500,
  category_id: 1,
  category_name: 'Food',
  category_icon: 'fork-knife',
  category_color: '#f97316',
  note: '',
  currency_code: 'USD',
  account_id: 1,
  created_at: '2026-04-15T10:00:00Z',
} as unknown as Transaction

describe('TransactionRow', () => {
  it('renders the category name and amount', () => {
    withI18n(<TransactionRow tx={baseTx} />)
    expect(screen.getByText('Food')).toBeInTheDocument()
    expect(screen.getByText(/\$15\.00/)).toBeInTheDocument()
  })

  it('renders income with positive sign', () => {
    const income = { ...baseTx, type: 'income' as const, amount_cents: 5000 }
    withI18n(<TransactionRow tx={income} />)
    const amount = screen.getByText(/\$50\.00/)
    expect(amount.textContent?.startsWith('+')).toBe(true)
  })

  it('renders expense with unicode minus prefix', () => {
    withI18n(<TransactionRow tx={baseTx} />)
    const amount = screen.getByText(/\$15\.00/)
    expect(amount.textContent?.startsWith('−')).toBe(true)
  })

  it('shows note instead of date when note is present (compact mode)', () => {
    const tx = { ...baseTx, note: 'Lunch with team' }
    withI18n(<TransactionRow tx={tx} />)
    expect(screen.getByText('Lunch with team')).toBeInTheDocument()
  })

  it('triggers onEdit when row is clicked', async () => {
    const user = userEvent.setup()
    const onEdit = vi.fn()
    withI18n(<TransactionRow tx={baseTx} onEdit={onEdit} />)
    await user.click(screen.getByText('Food'))
    expect(onEdit).toHaveBeenCalledWith(baseTx)
  })

  it('triggers onDelete via the delete button without firing onEdit', async () => {
    const user = userEvent.setup()
    const onEdit = vi.fn()
    const onDelete = vi.fn()
    withI18n(<TransactionRow tx={baseTx} onEdit={onEdit} onDelete={onDelete} />)
    const btn = screen.getByRole('button', { name: 'Delete' })
    await user.click(btn)
    expect(onDelete).toHaveBeenCalledWith(1)
    expect(onEdit).not.toHaveBeenCalled()
  })

  it('reduces opacity while deleting', () => {
    const { container } = withI18n(<TransactionRow tx={baseTx} isDeleting />)
    const root = container.firstChild as HTMLElement
    expect(root.className).toContain('opacity-30')
    expect(root.className).toContain('pointer-events-none')
  })
})
