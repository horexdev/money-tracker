import { Trash2 } from 'lucide-react'
import type { Transaction } from '../../types'
import { AmountDisplay } from './AmountDisplay'

interface TransactionRowProps {
  tx: Transaction
  compact?: boolean
  onDelete?: (id: number) => void
  isDeleting?: boolean
}

export function TransactionRow({ tx, compact = false, onDelete, isDeleting = false }: TransactionRowProps) {
  const isIncome = tx.type === 'income'
  const date = new Date(tx.created_at).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
  })

  return (
    <div
      className={`
        flex items-center gap-3 px-4 py-3 border-b border-border last:border-b-0
        transition-opacity ${isDeleting ? 'opacity-30 pointer-events-none' : ''}
      `}
    >
      <div className="w-9 h-9 rounded-xl bg-border flex items-center justify-center text-xl shrink-0">
        {tx.category_emoji}
      </div>

      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between gap-2">
          <span className="text-sm font-medium text-text truncate">{tx.category_name}</span>
          <AmountDisplay
            cents={tx.amount_cents}
            currency={tx.currency_code}
            type={isIncome ? 'income' : 'expense'}
            size="sm"
            showSign
          />
        </div>
        {!compact && (
          <div className="flex items-center justify-between gap-2 mt-0.5">
            <span className="text-xs text-muted truncate">{tx.note || date}</span>
            {tx.note && <span className="text-xs text-muted shrink-0">{date}</span>}
          </div>
        )}
      </div>

      {onDelete && (
        <button
          onClick={() => onDelete(tx.id)}
          className="shrink-0 w-7 h-7 flex items-center justify-center rounded-full
            text-muted active:text-destructive transition-colors"
          aria-label="Delete"
        >
          <Trash2 size={14} />
        </button>
      )}
    </div>
  )
}
