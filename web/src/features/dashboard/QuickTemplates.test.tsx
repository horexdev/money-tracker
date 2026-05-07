import { beforeEach, describe, expect, it, vi } from 'vitest'
import { screen, waitFor, fireEvent } from '@testing-library/react'

const { applyMock, listMock } = vi.hoisted(() => ({
  applyMock: vi.fn().mockResolvedValue({ id: 100 }),
  listMock: vi.fn(),
}))

vi.mock('../../shared/api/templates', () => ({
  templatesApi: {
    list: () => listMock(),
    apply: applyMock,
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    reorder: vi.fn(),
  },
}))

vi.mock('../../shared/hooks/useHaptic', () => ({
  useHaptic: () => ({ impact: vi.fn(), notification: vi.fn(), selection: vi.fn() }),
}))

import { QuickTemplates } from './QuickTemplates'
import { renderWithProviders } from '../../test/render'

beforeEach(() => {
  vi.clearAllMocks()
  applyMock.mockClear()
  listMock.mockReset()
})

describe('QuickTemplates', () => {
  it('renders nothing when no templates', async () => {
    listMock.mockResolvedValue({ templates: [] })
    const { container } = renderWithProviders(<QuickTemplates />)
    await waitFor(() => {
      expect(listMock).toHaveBeenCalled()
    })
    expect(container.textContent ?? '').not.toContain('Coffee')
  })

  it('renders cards and applies a fixed-amount template on tap', async () => {
    listMock.mockResolvedValue({
      templates: [
        {
          id: 5, name: 'Coffee', type: 'expense',
          amount_cents: 30000, amount_fixed: true, currency_code: 'USD',
          category_id: 2, category_name: 'Food', category_icon: 'coffee', category_color: '#f59e0b',
          account_id: 7, note: '', sort_order: 0, created_at: '',
        },
      ],
    })

    renderWithProviders(<QuickTemplates />)
    const btn = await screen.findByRole('button', { name: /apply/i })
    // Simulate click via pointerdown + quick pointerup (long-press hook treats short tap as click).
    fireEvent.pointerDown(btn, { clientX: 0, clientY: 0 })
    fireEvent.pointerUp(btn, { clientX: 0, clientY: 0 })

    await waitFor(() => {
      expect(applyMock).toHaveBeenCalledWith(5, undefined)
    })
  })

  it('opens the amount modal for a variable-amount template', async () => {
    listMock.mockResolvedValue({
      templates: [
        {
          id: 6, name: 'Lunch', type: 'expense',
          amount_cents: 50000, amount_fixed: false, currency_code: 'USD',
          category_id: 2, category_name: 'Food', category_icon: 'fork', category_color: '#f59e0b',
          account_id: 7, note: '', sort_order: 0, created_at: '',
        },
      ],
    })

    renderWithProviders(<QuickTemplates />)
    const btn = await screen.findByRole('button', { name: /apply/i })
    fireEvent.pointerDown(btn, { clientX: 0, clientY: 0 })
    fireEvent.pointerUp(btn, { clientX: 0, clientY: 0 })

    // Modal opens — Confirm button is rendered, apply should NOT have been called yet.
    await screen.findByRole('button', { name: /confirm/i })
    expect(applyMock).not.toHaveBeenCalled()
  })
})
