import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { AmountInput } from './AmountInput'

interface ControlledProps {
  initial?: string
  onChange?: (value: string) => void
  calculator?: boolean
  currency?: string
}

function Controlled({ initial = '', onChange, calculator = true, currency = 'USD' }: ControlledProps) {
  const [value, setValue] = useState(initial)
  return (
    <AmountInput
      value={value}
      onChange={(next) => {
        setValue(next)
        onChange?.(next)
      }}
      currency={currency}
      calculator={calculator}
    />
  )
}

describe('AmountInput', () => {
  it('renders the input with the placeholder and currency badge', () => {
    render(<AmountInput value="" onChange={() => undefined} currency="USD" />)
    expect(screen.getByPlaceholderText('0.00')).toBeInTheDocument()
    expect(screen.getByText('$')).toBeInTheDocument()
  })

  it('strips letters but keeps operators when calculator=true', async () => {
    const user = userEvent.setup()
    const handle = vi.fn()
    render(<Controlled onChange={handle} calculator={true} />)
    await user.type(screen.getByPlaceholderText('0.00'), '1abc+2')
    expect(handle).toHaveBeenLastCalledWith('1+2')
  })

  it('strips operators with calculator=false (strict sanitiser)', async () => {
    const user = userEvent.setup()
    const handle = vi.fn()
    render(<Controlled onChange={handle} calculator={false} />)
    await user.type(screen.getByPlaceholderText('0.00'), '1+2')
    expect(handle).toHaveBeenLastCalledWith('12')
  })

  it('shows a preview when the value is a valid expression', () => {
    render(<AmountInput value="1+2" onChange={() => undefined} currency="USD" />)
    expect(screen.getByText(/= \$3\.00/)).toBeInTheDocument()
  })

  it('does not show a preview for plain numbers', () => {
    render(<AmountInput value="100" onChange={() => undefined} currency="USD" />)
    expect(screen.queryByText(/=/)).not.toBeInTheDocument()
  })

  it('commits the evaluated expression on Enter', async () => {
    const user = userEvent.setup()
    const handle = vi.fn()
    render(<Controlled initial="1+2" onChange={handle} />)
    const input = screen.getByPlaceholderText('0.00') as HTMLInputElement
    input.focus()
    await user.keyboard('{Enter}')
    expect(handle).toHaveBeenLastCalledWith('3')
  })

  it('commits the evaluated expression on blur', () => {
    const handle = vi.fn()
    render(<Controlled initial="1+2" onChange={handle} />)
    const input = screen.getByPlaceholderText('0.00') as HTMLInputElement
    fireEvent.blur(input)
    expect(handle).toHaveBeenLastCalledWith('3')
  })

  it('does not commit on Enter when expression is invalid', async () => {
    const user = userEvent.setup()
    const handle = vi.fn()
    render(<Controlled initial="1+" onChange={handle} />)
    const input = screen.getByPlaceholderText('0.00') as HTMLInputElement
    input.focus()
    await user.keyboard('{Enter}')
    expect(handle).not.toHaveBeenCalled()
  })
})
