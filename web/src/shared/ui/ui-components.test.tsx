import { describe, expect, it, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { I18nextProvider, initReactI18next } from 'react-i18next'
import i18n from 'i18next'

import { Badge } from './Badge'
import { Button } from './Button'
import { Card } from './Card'
import { EmptyState } from './EmptyState'
import { ErrorBoundary } from './ErrorBoundary'
import { ErrorMessage } from './ErrorMessage'
import { FAB } from './FAB'
import { SectionHeader } from './SectionHeader'
import { SegmentedControl } from './SegmentedControl'
import { Spinner } from './Spinner'

function renderI18n(ui: React.ReactElement) {
  const inst = i18n.createInstance()
  inst.use(initReactI18next).init({
    lng: 'en',
    fallbackLng: 'en',
    resources: {
      en: { translation: { common: { error: 'Something went wrong', retry: 'Retry' } } },
    },
    interpolation: { escapeValue: false },
    react: { useSuspense: false },
  })
  function Wrapper({ children }: { children: ReactNode }) {
    return <I18nextProvider i18n={inst}>{children}</I18nextProvider>
  }
  return render(ui, { wrapper: Wrapper })
}

describe('Badge', () => {
  it('renders children with default variant class', () => {
    render(<Badge>New</Badge>)
    const node = screen.getByText('New')
    expect(node.className).toContain('bg-bg')
  })

  it.each(['accent', 'income', 'expense'] as const)('applies variant=%s class', (variant) => {
    render(<Badge variant={variant}>x</Badge>)
    const node = screen.getByText('x')
    expect(node.className).toContain(variant === 'accent' ? 'bg-accent-subtle' : `bg-${variant}-subtle`)
  })
})

describe('Button', () => {
  it('renders children and is clickable', async () => {
    const user = userEvent.setup()
    const handle = vi.fn()
    render(<Button onClick={handle}>Save</Button>)
    await user.click(screen.getByRole('button', { name: 'Save' }))
    expect(handle).toHaveBeenCalledTimes(1)
  })

  it('respects the disabled attribute', async () => {
    const user = userEvent.setup()
    const handle = vi.fn()
    render(
      <Button onClick={handle} disabled>
        Disabled
      </Button>,
    )
    const btn = screen.getByRole('button', { name: 'Disabled' })
    expect(btn).toBeDisabled()
    await user.click(btn)
    expect(handle).not.toHaveBeenCalled()
  })

  it.each(['primary', 'secondary', 'ghost', 'destructive'] as const)('applies variant=%s class', (variant) => {
    render(<Button variant={variant}>x</Button>)
    const btn = screen.getByRole('button', { name: 'x' })
    if (variant === 'primary') expect(btn.className).toContain('bg-accent')
    if (variant === 'secondary') expect(btn.className).toContain('bg-surface')
    if (variant === 'ghost') expect(btn.className).toContain('bg-transparent')
    if (variant === 'destructive') expect(btn.className).toContain('text-expense')
  })

  it.each(['sm', 'md', 'lg'] as const)('applies size=%s height class', (size) => {
    render(<Button size={size}>x</Button>)
    const btn = screen.getByRole('button', { name: 'x' })
    if (size === 'sm') expect(btn.className).toContain('h-9')
    if (size === 'md') expect(btn.className).toContain('h-11')
    if (size === 'lg') expect(btn.className).toContain('h-13')
  })
})

describe('Card', () => {
  it('renders children with card-elevated', () => {
    const { container } = render(<Card>hi</Card>)
    const node = container.firstChild as HTMLElement
    expect(node.className).toContain('card-elevated')
  })

  it('applies custom padding', () => {
    const { container } = render(<Card padding="p-3">hi</Card>)
    const node = container.firstChild as HTMLElement
    expect(node.className).toContain('p-3')
  })
})

describe('EmptyState', () => {
  it('renders title and description', () => {
    render(<EmptyState title="No data" description="Add something" />)
    expect(screen.getByText('No data')).toBeInTheDocument()
    expect(screen.getByText('Add something')).toBeInTheDocument()
  })

  it('renders an action when provided', () => {
    render(
      <EmptyState
        title="empty"
        action={<button type="button">do it</button>}
      />,
    )
    expect(screen.getByRole('button', { name: 'do it' })).toBeInTheDocument()
  })
})

describe('ErrorBoundary', () => {
  it('renders children when no error', () => {
    render(<ErrorBoundary>ok</ErrorBoundary>)
    expect(screen.getByText('ok')).toBeInTheDocument()
  })

  it('renders fallback when child throws', () => {
    const Bomb = () => {
      throw new Error('boom')
    }
    // Suppress expected console.error noise from React.
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined)
    try {
      render(
        <ErrorBoundary>
          <Bomb />
        </ErrorBoundary>,
      )
      expect(screen.getByText('Something went wrong')).toBeInTheDocument()
      expect(screen.getByText('boom')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Reload' })).toBeInTheDocument()
    } finally {
      errSpy.mockRestore()
    }
  })
})

describe('ErrorMessage', () => {
  it('renders the default i18n message when no override given', () => {
    renderI18n(<ErrorMessage />)
    expect(screen.getByText('Something went wrong')).toBeInTheDocument()
  })

  it('renders an override message', () => {
    renderI18n(<ErrorMessage message="Custom error" />)
    expect(screen.getByText('Custom error')).toBeInTheDocument()
  })

  it('shows a retry button when onRetry is provided', async () => {
    const user = userEvent.setup()
    const onRetry = vi.fn()
    renderI18n(<ErrorMessage onRetry={onRetry} />)
    const btn = screen.getByRole('button', { name: 'Retry' })
    await user.click(btn)
    expect(onRetry).toHaveBeenCalledTimes(1)
  })

  it('omits the retry button when no onRetry is given', () => {
    renderI18n(<ErrorMessage />)
    expect(screen.queryByRole('button', { name: 'Retry' })).not.toBeInTheDocument()
  })
})

describe('FAB', () => {
  it('renders into document.body and triggers onClick', () => {
    const handle = vi.fn()
    render(<FAB onClick={handle} label="Add" ariaLabel="add" />)
    const btn = screen.getByRole('button', { name: 'add' })
    fireEvent.click(btn)
    expect(handle).toHaveBeenCalledTimes(1)
  })

  it('uses label as default aria-label when ariaLabel is omitted', () => {
    render(<FAB onClick={() => undefined} label="Create" />)
    expect(screen.getByRole('button', { name: 'Create' })).toBeInTheDocument()
  })

  it('falls back to "Add" aria-label without label or ariaLabel', () => {
    render(<FAB onClick={() => undefined} />)
    expect(screen.getByRole('button', { name: 'Add' })).toBeInTheDocument()
  })
})

describe('SectionHeader', () => {
  it('renders heading text', () => {
    render(<SectionHeader>Title</SectionHeader>)
    expect(screen.getByText('Title')).toBeInTheDocument()
  })

  it('renders an action slot', () => {
    render(<SectionHeader action={<span data-testid="act">act</span>}>x</SectionHeader>)
    expect(screen.getByTestId('act')).toBeInTheDocument()
  })
})

describe('SegmentedControl', () => {
  const options = [
    { value: 'a', label: 'A' },
    { value: 'b', label: 'B' },
    { value: 'c', label: 'C' },
  ]

  it('marks active option with shadow-card and inactive with text-muted', () => {
    render(<SegmentedControl options={options} value="b" onChange={() => undefined} />)
    const a = screen.getByRole('button', { name: 'A' })
    const b = screen.getByRole('button', { name: 'B' })
    expect(b.className).toContain('shadow-card')
    expect(a.className).toContain('text-muted')
    expect(a.className).not.toContain('shadow-card')
  })

  it('calls onChange with the selected value', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<SegmentedControl options={options} value="a" onChange={onChange} />)
    await user.click(screen.getByRole('button', { name: 'C' }))
    expect(onChange).toHaveBeenCalledWith('c')
  })

  it('honours size=sm', () => {
    render(<SegmentedControl options={options} value="a" onChange={() => undefined} size="sm" />)
    const a = screen.getByRole('button', { name: 'A' })
    expect(a.className).toContain('text-xs')
  })
})

describe('Spinner', () => {
  it('renders default size md', () => {
    const { container } = render(<Spinner />)
    expect(container.firstChild).toHaveProperty('className')
    const cls = (container.firstChild as HTMLElement).className
    expect(cls).toContain('w-6')
    expect(cls).toContain('h-6')
  })

  it.each(['sm', 'lg'] as const)('respects size=%s', (size) => {
    const { container } = render(<Spinner size={size} />)
    const cls = (container.firstChild as HTMLElement).className
    if (size === 'sm') expect(cls).toContain('w-4')
    if (size === 'lg') expect(cls).toContain('w-8')
  })
})
