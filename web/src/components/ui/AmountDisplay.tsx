import { formatCents } from '../../lib/money'

type AmountSize = 'sm' | 'md' | 'lg' | 'xl'
type AmountType = 'income' | 'expense' | 'neutral'

interface AmountDisplayProps {
  cents: number
  currency?: string
  type?: AmountType
  size?: AmountSize
  showSign?: boolean
  className?: string
}

const sizeClasses: Record<AmountSize, string> = {
  sm: 'text-sm',
  md: 'text-base',
  lg: 'text-2xl font-bold',
  xl: 'text-4xl font-bold',
}

const typeClasses: Record<AmountType, string> = {
  income:  'text-income',
  expense: 'text-expense',
  neutral: 'text-text',
}

export function AmountDisplay({
  cents,
  currency = 'USD',
  type = 'neutral',
  size = 'md',
  showSign = false,
  className = '',
}: AmountDisplayProps) {
  const sign = showSign ? (type === 'income' ? '+' : type === 'expense' ? '−' : '') : ''

  return (
    <span className={`font-semibold tabular-nums ${sizeClasses[size]} ${typeClasses[type]} ${className}`}>
      {sign}{formatCents(Math.abs(cents), currency)}
    </span>
  )
}
