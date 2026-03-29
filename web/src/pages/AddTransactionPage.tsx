import { useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Check, CalendarBlank } from '@phosphor-icons/react'
import { AnimatePresence } from 'framer-motion'
import { categoriesApi } from '../api/categories'
import { transactionsApi } from '../api/transactions'
import { balanceApi } from '../api/balance'
import { parseCents, getCurrencySymbol } from '../lib/money'
import { CategoryIcon } from '../lib/categoryIcons'
import { useTgMainButton } from '../hooks/useMainButton'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { useHaptic } from '../hooks/useHaptic'
import { Spinner } from '../components/Spinner'
import { PageTransition } from '../components/PageTransition'
import { useCategoryName } from '../hooks/useCategoryName'
import { SingleDateModal, fmtDisplay } from '../components/ui/DatePicker'
import type { TransactionType } from '../types'

/** Only allow digits and a single dot with up to 2 decimal places */
function sanitizeAmount(value: string): string {
  // Strip everything except digits and dots
  let cleaned = value.replace(/[^0-9.]/g, '')
  // Only allow one dot
  const dotIndex = cleaned.indexOf('.')
  if (dotIndex !== -1) {
    cleaned = cleaned.slice(0, dotIndex + 1) + cleaned.slice(dotIndex + 1).replace(/\./g, '')
  }
  // Limit to 2 decimal places
  if (dotIndex !== -1 && cleaned.length - dotIndex > 3) {
    cleaned = cleaned.slice(0, dotIndex + 3)
  }
  // Prevent leading zeros (except "0." for decimals)
  if (cleaned.length > 1 && cleaned[0] === '0' && cleaned[1] !== '.') {
    cleaned = cleaned.slice(1)
  }
  return cleaned
}

export function AddTransactionPage() {
  const { t } = useTranslation()
  const tCategory = useCategoryName()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { selection, notification } = useHaptic()

  const [type, setType] = useState<TransactionType>('expense')
  const [amount, setAmount] = useState('')
  const [categoryID, setCategoryID] = useState<number | null>(null)
  const [note, setNote] = useState('')
  const [txDate, setTxDate] = useState<string>(new Date().toISOString().split('T')[0])
  const [showDatePicker, setShowDatePicker] = useState(false)

  const { data: catData, isLoading } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  })

  const { data: balanceData } = useQuery({
    queryKey: ['balance'],
    queryFn: balanceApi.get,
  })

  const baseCurrency = balanceData?.by_currency?.[0]?.currency_code ?? 'USD'
  const currencySymbol = getCurrencySymbol(baseCurrency)

  const filtered = (catData?.categories ?? []).filter(
    (c) => c.type === type || c.type === 'both'
  )

  const mutation = useMutation({
    mutationFn: transactionsApi.create,
    onSuccess: () => {
      notification('success')
      qc.invalidateQueries({ queryKey: ['transactions'] })
      qc.invalidateQueries({ queryKey: ['balance'] })
      qc.invalidateQueries({ queryKey: ['stats'] })
      navigate('/')
    },
    onError: () => notification('error'),
  })

  const canSubmit = parseCents(amount) > 0 && categoryID !== null && !mutation.isPending

  const handleSubmit = useCallback(() => {
    if (!canSubmit || categoryID === null) return
    const today = new Date().toISOString().split('T')[0]
    mutation.mutate({
      category_id: categoryID,
      type,
      amount_cents: parseCents(amount),
      note: note.trim() || undefined,
      currency_code: baseCurrency,
      created_at: txDate !== today ? txDate : undefined,
    })
  }, [canSubmit, categoryID, type, amount, note, baseCurrency, mutation])

  const handleAmountChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setAmount(sanitizeAmount(e.target.value))
  }, [])

  useTgBackButton(() => navigate('/'), true)
  useTgMainButton({
    text: mutation.isPending ? t('common.loading') : t('common.save'),
    onClick: handleSubmit,
    enabled: canSubmit,
    loading: mutation.isPending,
  })

  const isExpense = type === 'expense'

  return (
    <PageTransition>
      <div className="flex flex-col h-[calc(100dvh-var(--tab-bar-h))]">

        {/* Hero — type toggle + amount */}
        <div
          className="mx-4 mt-4 rounded-card p-5 pb-6 relative overflow-hidden shrink-0"
          style={{
            background: isExpense
              ? 'linear-gradient(135deg, #7f1d1d 0%, #ef4444 50%, #f87171 100%)'
              : 'linear-gradient(135deg, #14532d 0%, #22c55e 50%, #4ade80 100%)',
            boxShadow: isExpense
              ? '0 8px 32px rgba(239,68,68,0.3), 0 2px 8px rgba(0,0,0,0.1)'
              : '0 8px 32px rgba(34,197,94,0.3), 0 2px 8px rgba(0,0,0,0.1)',
          }}
        >
          <div className="absolute -top-10 -right-10 w-32 h-32 rounded-full bg-white/[0.06] blur-xl pointer-events-none" />
          <div className="absolute -bottom-8 -left-8 w-24 h-24 rounded-full bg-white/10 blur-2xl pointer-events-none" />

          <div className="relative z-10">
            {/* Type toggle — glass pills */}
            <div className="inline-flex bg-white/10 backdrop-blur-sm rounded-2xl p-1 gap-1 border border-white/[0.08]">
              {(['expense', 'income'] as TransactionType[]).map((v) => (
                <button
                  key={v}
                  onClick={() => { setType(v); setCategoryID(null); selection() }}
                  className={`
                    px-5 py-2 rounded-xl text-xs font-bold transition-all duration-200 select-none
                    ${type === v
                      ? 'bg-white/20 text-white shadow-[0_2px_8px_rgba(0,0,0,0.15)]'
                      : 'text-white/50'
                    }
                  `}
                >
                  {v === 'expense' ? t('transactions.expense') : t('transactions.income')}
                </button>
              ))}
            </div>

            {/* Amount input */}
            <div className="mt-4 flex items-baseline gap-1">
              <span className="text-3xl font-bold text-white/50 shrink-0">{currencySymbol}</span>
              <input
                inputMode="decimal"
                placeholder="0.00"
                value={amount}
                onChange={handleAmountChange}
                autoFocus
                className="flex-1 bg-transparent text-white text-4xl font-extrabold outline-none tabular-nums placeholder:text-white/25 min-w-0"
              />
            </div>
          </div>
        </div>

        {/* Note + Date */}
        <div className="mx-4 mt-3 card-elevated shrink-0">
          <div className="px-4 py-3 flex items-center gap-3 border-b border-border">
            <span className="text-[11px] font-bold text-muted uppercase tracking-widest shrink-0">
              {t('transactions.note')}
            </span>
            <input
              type="text"
              placeholder={t('transactions.note_placeholder')}
              value={note}
              onChange={(e) => setNote(e.target.value)}
              maxLength={120}
              className="flex-1 bg-transparent text-sm text-text outline-none min-w-0"
            />
          </div>
          <button
            onClick={() => setShowDatePicker(true)}
            className="w-full px-4 py-3 flex items-center gap-3 active:bg-accent-subtle/30 transition-colors"
          >
            <CalendarBlank size={16} weight="bold" className="text-muted shrink-0" />
            <span className="text-[11px] font-bold text-muted uppercase tracking-widest shrink-0">
              {t('transactions.date')}
            </span>
            <span className="flex-1 text-sm text-text text-right">{fmtDisplay(txDate)}</span>
          </button>
        </div>

        {/* Categories — scrollable grid + inline save button */}
        <div className="flex-1 min-h-0 mt-3 flex flex-col">
          <p className="px-5 mb-2 text-[11px] font-bold text-muted uppercase tracking-widest shrink-0">
            {t('transactions.category')}
          </p>
          <div className="flex-1 min-h-0 overflow-y-auto no-scrollbar px-4 pb-4">
            {isLoading ? (
              <div className="flex justify-center py-8"><Spinner /></div>
            ) : (
              <div className="grid grid-cols-4 gap-2">
                {filtered.map((cat) => {
                  const isSelected = categoryID === cat.id
                  return (
                    <button
                      key={cat.id}
                      onClick={() => { setCategoryID(cat.id); selection() }}
                      className={`
                        flex flex-col items-center justify-center gap-1.5 py-3 rounded-2xl
                        text-[11px] font-semibold transition-all duration-150 active:scale-[0.93] relative
                        ${isSelected
                          ? isExpense
                            ? 'bg-expense/10 text-expense shadow-[0_2px_12px_rgba(239,68,68,0.2)]'
                            : 'bg-income/10 text-income shadow-[0_2px_12px_rgba(34,197,94,0.2)]'
                          : 'bg-surface text-text shadow-sm'
                        }
                      `}
                    >
                      {isSelected && (
                        <div className={`absolute top-1.5 right-1.5 w-4 h-4 rounded-full flex items-center justify-center ${
                          isExpense ? 'bg-expense' : 'bg-income'
                        }`}>
                          <Check size={10} weight="bold" className="text-white" />
                        </div>
                      )}
                      <div
                        className="w-10 h-10 rounded-2xl flex items-center justify-center"
                        style={{ background: isSelected
                          ? (isExpense ? 'var(--color-expense)' : 'var(--color-income)')
                          : (cat.color || 'var(--color-accent)')
                        }}
                      >
                        <CategoryIcon
                          emoji={cat.emoji}
                          size={20}
                          weight="fill"
                          className="text-white"
                        />
                      </div>
                      <span className="text-center leading-tight px-0.5 truncate w-full">
                        {tCategory(cat.name)}
                      </span>
                    </button>
                  )
                })}
              </div>
            )}

            {/* Save button — inline at bottom of scroll area so keyboard never covers it */}
            <button
              onClick={handleSubmit}
              disabled={!canSubmit}
              className={`
                w-full mt-4 py-4 rounded-2xl text-[15px] font-bold transition-all active:scale-[0.98]
                ${canSubmit
                  ? 'bg-accent text-accent-text shadow-[0_4px_16px_rgba(99,102,241,0.35)]'
                  : 'bg-border text-muted'
                }
              `}
            >
              {mutation.isPending ? t('common.loading') : t('common.save')}
            </button>
          </div>
        </div>
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

    </PageTransition>
  )
}
