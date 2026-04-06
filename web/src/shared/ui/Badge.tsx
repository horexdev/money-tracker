import type { ReactNode } from 'react'

type BadgeVariant = 'default' | 'accent' | 'income' | 'expense'

interface BadgeProps {
  children: ReactNode
  variant?: BadgeVariant
  className?: string
}

const variantClasses: Record<BadgeVariant, string> = {
  default: 'bg-bg text-muted',
  accent:  'bg-accent-subtle text-accent',
  income:  'bg-income-subtle text-income',
  expense: 'bg-expense-subtle text-expense',
}

export function Badge({ children, variant = 'default', className = '' }: BadgeProps) {
  return (
    <span
      className={`inline-flex items-center px-2.5 py-1 rounded-(--radius-xs) text-[11px] font-bold ${variantClasses[variant]} ${className}`}
    >
      {children}
    </span>
  )
}
