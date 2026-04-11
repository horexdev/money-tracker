import { Trash } from '@phosphor-icons/react'
import { useTranslation } from 'react-i18next'
import type { Transaction } from '../types'
import { CategoryIcon } from '../lib/categoryIcons'
import { AmountDisplay } from './AmountDisplay'
import { useCategoryName } from '../hooks/useCategoryName'
import { formatDate } from '../lib/money'

interface TransactionRowProps {
  tx: Transaction
  compact?: boolean
  onDelete?: (id: number) => void
  onEdit?: (tx: Transaction) => void
  isDeleting?: boolean
}

export function TransactionRow({ tx, compact = false, onDelete, onEdit, isDeleting = false }: TransactionRowProps) {
  const isIncome = tx.type === 'income'
  const tCategory = useCategoryName()
  const { t, i18n } = useTranslation()
  const date = formatDate(tx.created_at, i18n.language)

  return (
    <div
      className={`
        flex items-center gap-3.5 px-5 py-3.5 border-b border-border last:border-b-0
        transition-opacity ${isDeleting ? 'opacity-30 pointer-events-none' : ''}
        ${onEdit ? 'active:bg-accent-subtle/20 cursor-pointer' : ''}
      `}
      onClick={onEdit ? () => onEdit(tx) : undefined}
    >
      <div
        className="w-11 h-11 rounded-(--radius-btn) flex items-center justify-center shrink-0"
        style={{ background: tx.category_color || 'var(--color-accent)' }}
      >
        <CategoryIcon icon={tx.category_icon} size={22} weight="fill" className="text-white" />
      </div>

      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between gap-2">
          <span className="text-sm font-semibold text-text truncate">{tCategory(tx.category_name)}</span>
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
          onClick={(e) => { e.stopPropagation(); onDelete(tx.id) }}
          className="shrink-0 w-8 h-8 flex items-center justify-center rounded-(--radius-xs)
            text-muted hover:text-destructive hover:bg-expense-subtle transition-all"
          aria-label={t('common.delete')}
        >
          <Trash size={16} weight="bold" />
        </button>
      )}
    </div>
  )
}
