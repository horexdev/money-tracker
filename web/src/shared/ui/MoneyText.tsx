import { useTranslation } from 'react-i18next'
import { formatCents } from '../lib/money'
import { useHideAmounts } from '../hooks/useHideAmounts'

interface MoneyTextProps {
  cents: number
  currency?: string
  className?: string
}

/**
 * Renders a formatted money value, or a clickable mask when privacy mode is on.
 * Use it at the leaf of any UI that shows a sum without going through AmountDisplay.
 */
export function MoneyText({ cents, currency = 'USD', className = '' }: MoneyTextProps) {
  const { hidden, toggle } = useHideAmounts()
  const { t } = useTranslation()

  if (hidden) {
    return (
      <span
        role="button"
        tabIndex={0}
        aria-label={t('common.show_amounts')}
        onClick={(e) => { e.stopPropagation(); toggle() }}
        className={`cursor-pointer ${className}`}
      >
        ••••
      </span>
    )
  }

  return <span className={className}>{formatCents(cents, currency)}</span>
}
