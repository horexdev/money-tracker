/**
 * Currency badge — shows the typographic symbol for a currency code.
 * Falls back to getCurrencySymbol() which returns the ISO code if no symbol exists.
 */

import { getCurrencySymbol } from './money'

interface CurrencyBadgeProps {
  currency: string
  /** Size variant: 'lg' for hero amount inputs, 'sm' for compact displays */
  size?: 'lg' | 'sm'
  className?: string
}

/**
 * Renders a currency symbol next to an amount input.
 * Always text — no flags, no emojis.
 */
export function CurrencyBadge({ currency, size = 'lg', className = '' }: CurrencyBadgeProps) {
  const symbol = getCurrencySymbol(currency)
  return (
    <span className={`shrink-0 tabular-nums ${
      size === 'lg' ? 'text-3xl font-bold' : 'text-sm font-semibold'
    } ${className}`}>
      {symbol}
    </span>
  )
}
