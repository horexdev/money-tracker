import { sanitizeAmount } from '../lib/money'
import { CurrencyBadge } from '../lib/currencyIcons'

interface AmountInputProps {
  value: string
  onChange: (value: string) => void
  currency: string
  placeholder?: string
  autoFocus?: boolean
}

export function AmountInput({ value, onChange, currency, placeholder = '0.00', autoFocus }: AmountInputProps) {
  return (
    <div className="flex items-baseline gap-2 bg-bg rounded-2xl px-4 py-3 focus-within:shadow-(--shadow-focus) transition-shadow">
      <CurrencyBadge currency={currency} className="text-muted/40" />
      <input
        inputMode="decimal"
        placeholder={placeholder}
        value={value}
        onChange={e => onChange(sanitizeAmount(e.target.value))}
        autoFocus={autoFocus}
        className="flex-1 bg-transparent text-3xl font-bold outline-none text-text placeholder:text-muted/20 tabular-nums min-w-0"
      />
    </div>
  )
}
