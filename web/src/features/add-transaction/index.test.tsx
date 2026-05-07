import { beforeEach, describe, expect, it, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

vi.mock('../../shared/api/categories', () => ({
  categoriesApi: {
    list: vi.fn().mockResolvedValue({
      categories: [
        { id: 1, name: 'Food', icon: 'Food', color: '#ff0000', type: 'expense' },
        { id: 2, name: 'Salary', icon: 'Salary', color: '#00ff00', type: 'income' },
      ],
    }),
  },
}))

vi.mock('../../shared/api/transactions', () => ({
  transactionsApi: {
    list: vi.fn(),
    create: vi.fn().mockResolvedValue({ id: 100 }),
    update: vi.fn(),
    delete: vi.fn(),
  },
}))

vi.mock('../../shared/api/transfers', () => ({
  transfersApi: {
    list: vi.fn(),
    create: vi.fn().mockResolvedValue({ id: 200 }),
    delete: vi.fn(),
  },
  exchangeApi: {
    getRate: vi.fn().mockResolvedValue({ rate: 1 }),
  },
}))

vi.mock('../../shared/api/balance', () => ({
  balanceApi: {
    get: vi.fn().mockResolvedValue({
      total_cents: 100000,
      by_currency: [{ currency_code: 'USD', total_cents: 100000 }],
    }),
  },
}))

vi.mock('../../shared/api/accounts', () => ({
  accountsApi: {
    list: vi.fn().mockResolvedValue([
      { id: 1, name: 'Main', currency_code: 'USD', is_default: true, balance_cents: 100000, icon: 'wallet', color: '#000', type: 'cash' },
      { id: 2, name: 'Savings', currency_code: 'USD', is_default: false, balance_cents: 50000, icon: 'piggy', color: '#000', type: 'savings' },
    ]),
  },
}))

vi.mock('../../shared/hooks/useHaptic', () => ({
  useHaptic: () => ({ impact: vi.fn(), notification: vi.fn(), selection: vi.fn() }),
}))

vi.mock('../../shared/hooks/useMainButton', () => ({
  useTgMainButton: vi.fn(),
}))

import { AddTransactionPage } from './index'
import { transactionsApi } from '../../shared/api/transactions'
import { renderWithProviders } from '../../test/render'

describe('AddTransactionPage smoke', () => {
  beforeEach(() => {
    vi.mocked(transactionsApi.create).mockClear()
  })

  it('renders the amount input and the three mode toggles', async () => {
    renderWithProviders(<AddTransactionPage />)

    expect(await screen.findByPlaceholderText('0.00')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /expense/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /income/i })).toBeInTheDocument()
  })

  it('calls transactionsApi.create when amount + category are picked and Save is clicked', async () => {
    const user = userEvent.setup()
    renderWithProviders(<AddTransactionPage />)

    const input = await screen.findByPlaceholderText('0.00')
    await user.type(input, '15')

    const foodCategory = await screen.findByText('Food')
    await user.click(foodCategory)

    const saveButton = (await screen.findAllByRole('button')).find(
      (b) => b.textContent?.trim().toLowerCase() === 'save',
    )
    expect(saveButton).toBeDefined()
    await user.click(saveButton!)

    await waitFor(() => {
      expect(transactionsApi.create).toHaveBeenCalledTimes(1)
    })

    const firstCallPayload = vi.mocked(transactionsApi.create).mock.calls[0][0]
    expect(firstCallPayload).toEqual(
      expect.objectContaining({
        category_id: 1,
        type: 'expense',
        amount_cents: 1500,
        account_id: 1,
      }),
    )
  })
})
