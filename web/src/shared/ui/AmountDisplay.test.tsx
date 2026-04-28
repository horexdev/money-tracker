import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'

import { AmountDisplay } from './AmountDisplay'

describe('AmountDisplay', () => {
  it('renders the formatted amount with USD by default', () => {
    render(<AmountDisplay cents={150050} />)
    expect(screen.getByText(/\$1,500\.50/)).toBeInTheDocument()
  })

  it('uses the absolute value so negative cents render unsigned by default', () => {
    render(<AmountDisplay cents={-150050} />)
    const node = screen.getByText(/\$1,500\.50/)
    expect(node.textContent ?? '').not.toContain('-')
    expect(node.textContent ?? '').not.toContain('−')
  })

  it('renders a + prefix for income when showSign is on', () => {
    render(<AmountDisplay cents={1000} type="income" showSign />)
    const node = screen.getByText(/\$10\.00/)
    expect(node.textContent?.startsWith('+')).toBe(true)
  })

  it('renders the unicode minus for expense when showSign is on', () => {
    render(<AmountDisplay cents={1000} type="expense" showSign />)
    const node = screen.getByText(/\$10\.00/)
    expect(node.textContent?.startsWith('−')).toBe(true)
  })

  it('omits any sign prefix when showSign is off', () => {
    render(<AmountDisplay cents={1000} type="income" showSign={false} />)
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
    render(<AmountDisplay cents={100} type={type} />)
    const node = screen.getByText(/\$1\.00/)
    expect(node.className).toContain(cls)
  })

  it.each([
    ['sm', 'text-sm'],
    ['md', 'text-base'],
    ['lg', 'text-2xl'],
    ['xl', 'text-4xl'],
  ] as const)('applies the size class for size=%s', (size, cls) => {
    render(<AmountDisplay cents={100} size={size} />)
    const node = screen.getByText(/\$1\.00/)
    expect(node.className).toContain(cls)
  })

  it('respects a non-USD currency prop', () => {
    render(<AmountDisplay cents={1000} currency="EUR" />)
    expect(screen.getByText(/€10\.00/)).toBeInTheDocument()
  })

  it('appends a custom className', () => {
    render(<AmountDisplay cents={100} className="custom-class" />)
    const node = screen.getByText(/\$1\.00/)
    expect(node.className).toContain('custom-class')
  })
})
