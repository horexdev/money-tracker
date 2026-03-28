import type { ReactNode } from 'react'

type BadgeVariant = 'default' | 'accent' | 'income' | 'expense'

interface BadgeProps {
  children: ReactNode
  variant?: BadgeVariant
  className?: string
}

const variantClasses: Record<BadgeVariant, string> = {
  default: 'bg-surface text-text',
  accent: 'bg-accent-subtle text-accent',
  income: 'bg-income-subtle text-income',
  expense: 'bg-expense-subtle text-expense',
}

export function Badge({ children, variant = 'default', className = '' }: BadgeProps) {
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded-[--radius-sm] text-xs font-medium ${variantClasses[variant]} ${className}`}
    >
      {children}
    </span>
  )
}
