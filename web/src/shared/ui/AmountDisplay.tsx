import { useTranslation } from 'react-i18next'
import { formatCents } from '../lib/money'
import { useHideAmounts } from '../hooks/useHideAmounts'

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
  const { hidden, toggle } = useHideAmounts()
  const { t } = useTranslation()

  if (hidden) {
    return (
      <span
        role="button"
        tabIndex={0}
        aria-label={t('common.show_amounts')}
        onClick={(e) => { e.stopPropagation(); toggle() }}
        className={`font-semibold tabular-nums cursor-pointer text-text ${sizeClasses[size]} ${className}`}
      >
        ••••
      </span>
    )
  }

  const sign = showSign ? (type === 'income' ? '+' : type === 'expense' ? '−' : '') : ''

  return (
    <span className={`font-semibold tabular-nums ${sizeClasses[size]} ${typeClasses[type]} ${className}`}>
      {sign}{formatCents(Math.abs(cents), currency)}
    </span>
  )
}
