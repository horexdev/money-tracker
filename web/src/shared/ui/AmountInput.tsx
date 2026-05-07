import { useRef } from 'react'
import type { ChangeEvent, KeyboardEvent } from 'react'
import { formatCents, sanitizeAmount } from '../lib/money'
import { evaluate, formatForInput, looksLikeExpression } from '../lib/expression'
import { CurrencyBadge } from '../lib/currencyIcons'

interface AmountInputProps {
  value: string
  onChange: (value: string) => void
  currency: string
  placeholder?: string
  autoFocus?: boolean
  variant?: 'default' | 'hero'
  calculator?: boolean
  className?: string
}

const PERMISSIVE_RE = /[^0-9.,+\-*/×÷−()%\s]/g
const MAX_LEN = 64

function permissiveSanitize(value: string): string {
  return value.replace(PERMISSIVE_RE, '').slice(0, MAX_LEN)
}

const TOOLBAR_OPS = ['+', '−', '×', '÷', '(', ')', '%']

export function AmountInput({
  value,
  onChange,
  currency,
  placeholder = '0.00',
  autoFocus,
  variant = 'default',
  calculator = true,
  className,
}: AmountInputProps) {
  const inputRef = useRef<HTMLInputElement | null>(null)

  const isExpr = calculator && looksLikeExpression(value)
  const evalResult = isExpr ? evaluate(value) : null
  const previewValue = evalResult?.ok ? Math.max(0, evalResult.value) : 0
  const previewString = evalResult?.ok ? formatForInput(previewValue) : ''
  const showPreview = !!evalResult?.ok && previewString !== value

  const commit = () => {
    if (!calculator || !evalResult?.ok) return
    if (previewString !== value) onChange(previewString)
  }

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    onChange(calculator ? permissiveSanitize(e.target.value) : sanitizeAmount(e.target.value))
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      commit()
    }
  }

  const insertOp = (op: string) => {
    const el = inputRef.current
    const start = el?.selectionStart ?? value.length
    const end = el?.selectionEnd ?? value.length
    const next = permissiveSanitize(value.slice(0, start) + op + value.slice(end))
    onChange(next)
    requestAnimationFrame(() => {
      const node = inputRef.current
      if (!node) return
      node.focus()
      const cursor = Math.min(next.length, start + op.length)
      node.setSelectionRange(cursor, cursor)
    })
  }

  const isHero = variant === 'hero'
  const wrapperClass = isHero
    ? 'flex items-baseline gap-1'
    : 'flex items-baseline gap-2 bg-bg rounded-2xl px-4 py-3 focus-within:shadow-(--shadow-focus) transition-shadow'
  const inputClass = isHero
    ? 'flex-1 bg-transparent text-white text-4xl font-extrabold outline-none tabular-nums placeholder:text-white/25 min-w-0'
    : 'flex-1 bg-transparent text-3xl font-bold outline-none text-text placeholder:text-muted/20 tabular-nums min-w-0'
  const badgeClass = isHero ? 'text-white/50' : 'text-muted/40'
  const previewClass = isHero
    ? 'mt-1 text-right text-xs text-white/70 tabular-nums'
    : 'mt-1 text-right text-xs text-muted tabular-nums'

  return (
    <div className={className}>
      <div className={wrapperClass}>
        <CurrencyBadge currency={currency} className={badgeClass} />
        <input
          ref={inputRef}
          inputMode="decimal"
          placeholder={placeholder}
          value={value}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          onBlur={commit}
          autoFocus={autoFocus}
          className={inputClass}
        />
      </div>
      {showPreview && (
        <div className={previewClass}>= {formatCents(Math.round(previewValue * 100), currency)}</div>
      )}
      {calculator && (
        <div className="hidden [@media(pointer:coarse)]:flex gap-1 mt-2">
          {TOOLBAR_OPS.map(op => (
            <button
              key={op}
              type="button"
              onMouseDown={e => e.preventDefault()}
              onClick={() => insertOp(op)}
              className={
                isHero
                  ? 'flex-1 py-1.5 rounded-lg bg-white/15 backdrop-blur-sm text-white font-bold text-sm active:scale-90 transition-transform'
                  : 'flex-1 py-1.5 rounded-lg bg-accent-subtle text-text font-bold text-sm active:scale-90 transition-transform'
              }
            >
              {op}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
