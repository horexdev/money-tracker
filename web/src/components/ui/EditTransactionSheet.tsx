import { useState } from 'react'
import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { AnimatePresence } from 'framer-motion'
import { CalendarBlank, Check, CaretDown } from '@phosphor-icons/react'
import { transactionsApi } from '../../api/transactions'
import { categoriesApi } from '../../api/categories'
import { parseCents, getCurrencySymbol } from '../../lib/money'
import { CategoryIcon } from '../../lib/categoryIcons'
import { useCategoryName } from '../../hooks/useCategoryName'
import { useBaseCurrency } from '../../hooks/useBaseCurrency'
import { useHaptic } from '../../hooks/useHaptic'
import { BottomSheet } from './BottomSheet'
import { SingleDateModal, fmtDisplay } from './DatePicker'
import type { Transaction } from '../../types'

function sanitizeAmount(value: string): string {
  let cleaned = value.replace(/[^0-9.]/g, '')
  const dotIndex = cleaned.indexOf('.')
  if (dotIndex !== -1) {
    cleaned = cleaned.slice(0, dotIndex + 1) + cleaned.slice(dotIndex + 1).replace(/\./g, '')
  }
  if (dotIndex !== -1 && cleaned.length - dotIndex > 3) cleaned = cleaned.slice(0, dotIndex + 3)
  if (cleaned.length > 1 && cleaned[0] === '0' && cleaned[1] !== '.') cleaned = cleaned.slice(1)
  return cleaned
}

export function EditTransactionSheet({
  tx,
  onClose,
}: {
  tx: Transaction
  onClose: () => void
}) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const { notification } = useHaptic()
  const tCategory = useCategoryName()
  const { code: baseCurrency } = useBaseCurrency()
  const symbol = getCurrencySymbol(tx.currency_code || baseCurrency)

  const [amount, setAmount] = useState(String(tx.amount_cents / 100))
  const [categoryID, setCategoryID] = useState<number>(tx.category_id)
  const [note, setNote] = useState(tx.note || '')
  const [txDate, setTxDate] = useState(tx.created_at.split('T')[0])
  const [showDatePicker, setShowDatePicker] = useState(false)

  const catsQ = useQuery({ queryKey: ['categories'], queryFn: () => categoriesApi.list() })
  const categories = catsQ.data?.categories ?? []
  const filtered = categories.filter(c => c.type === tx.type || c.type === 'both')

  const updateMut = useMutation({
    mutationFn: () => transactionsApi.update(tx.id, {
      amount_cents: parseCents(amount),
      category_id: categoryID,
      note: note.trim() || undefined,
      created_at: txDate,
    }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['transactions'] })
      qc.invalidateQueries({ queryKey: ['balance'] })
      qc.invalidateQueries({ queryKey: ['stats'] })
      notification('success')
      onClose()
    },
    onError: () => notification('error'),
  })

  const canSubmit = parseCents(amount) > 0 && categoryID > 0 && !updateMut.isPending

  return (
    <BottomSheet onClose={onClose}>
      <div
        className="px-5 space-y-4 overflow-y-auto no-scrollbar"
        style={{ maxHeight: '85dvh', paddingBottom: 'max(1.5rem, env(safe-area-inset-bottom))' }}
      >
        <p className="text-base font-bold text-text">{t('transactions.edit_title')}</p>

        {/* Amount */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('transactions.amount')}
          </label>
          <div className="flex items-baseline gap-1.5 bg-bg rounded-2xl px-4 py-3 focus-within:shadow-[0_0_0_2px_rgba(99,102,241,0.2)] transition-shadow">
            <span className="text-3xl font-bold text-muted/40 tabular-nums">{symbol}</span>
            <input
              inputMode="decimal"
              placeholder="0.00"
              value={amount}
              onChange={e => setAmount(sanitizeAmount(e.target.value))}
              className="flex-1 bg-transparent text-3xl font-bold outline-none text-text placeholder:text-muted/30 tabular-nums min-w-0"
            />
          </div>
        </div>

        {/* Date */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('transactions.date')}
          </label>
          <button
            onClick={() => setShowDatePicker(true)}
            className="w-full bg-bg rounded-2xl px-4 py-3 flex items-center gap-3 active:bg-accent-subtle/30 transition-colors"
          >
            <CalendarBlank size={16} weight="bold" className="text-muted shrink-0" />
            <span className="flex-1 text-sm text-text text-left">{fmtDisplay(txDate)}</span>
            <CaretDown size={14} weight="bold" className="text-muted shrink-0" />
          </button>
        </div>

        {/* Note */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('transactions.note')}
          </label>
          <input
            type="text"
            placeholder={t('transactions.note_placeholder')}
            value={note}
            onChange={e => setNote(e.target.value)}
            maxLength={120}
            className="w-full bg-bg rounded-2xl px-4 py-3 text-sm font-medium outline-none text-text placeholder:text-muted/50 transition-shadow focus:shadow-[0_0_0_2px_rgba(99,102,241,0.2)]"
          />
        </div>

        {/* Category */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('transactions.category')}
          </label>
          <div className="grid grid-cols-4 gap-2">
            {filtered.map(cat => {
              const isActive = categoryID === cat.id
              return (
                <button
                  key={cat.id}
                  onClick={() => setCategoryID(cat.id)}
                  className={`
                    flex flex-col items-center gap-1.5 py-2.5 rounded-2xl text-xs transition-all duration-150 active:scale-95 select-none relative
                    ${isActive ? 'bg-accent/10 shadow-[0_2px_8px_rgba(99,102,241,0.15)]' : 'bg-surface shadow-sm'}
                  `}
                >
                  {isActive && (
                    <div className="absolute top-1.5 right-1.5 w-4 h-4 rounded-full bg-accent flex items-center justify-center">
                      <Check size={10} weight="bold" className="text-white" />
                    </div>
                  )}
                  <div
                    className="w-9 h-9 rounded-2xl flex items-center justify-center"
                    style={{ background: isActive ? 'var(--color-accent)' : (cat.color || 'var(--color-accent)') }}
                  >
                    <CategoryIcon emoji={cat.emoji} size={18} weight="fill" className="text-white" />
                  </div>
                  <span className="truncate w-full text-center px-1 font-medium text-[10px] text-text">
                    {tCategory(cat.name)}
                  </span>
                </button>
              )
            })}
          </div>
        </div>

        {/* Save */}
        <button
          onClick={() => updateMut.mutate()}
          disabled={!canSubmit}
          className={`
            w-full py-4 rounded-2xl text-[15px] font-bold transition-all active:scale-[0.98]
            ${canSubmit
              ? 'bg-accent text-accent-text shadow-[0_4px_16px_rgba(99,102,241,0.3)]'
              : 'bg-border text-muted'
            }
          `}
        >
          {updateMut.isPending ? '...' : t('common.save')}
        </button>

        {updateMut.isError && (
          <p className="text-xs text-destructive text-center">
            {(updateMut.error as Error)?.message}
          </p>
        )}
      </div>

      <AnimatePresence>
        {showDatePicker && (
          <SingleDateModal
            value={txDate}
            onApply={(iso) => setTxDate(iso)}
            onClose={() => setShowDatePicker(false)}
            applyLabel={t('common.done')}
          />
        )}
      </AnimatePresence>
    </BottomSheet>
  )
}
