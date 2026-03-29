import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { motion, AnimatePresence } from 'framer-motion'
import { Plus, Warning } from '@phosphor-icons/react'
import { fetchBudgets, createBudget, updateBudget, deleteBudget } from '../api/budgets'
import { categoriesApi } from '../api/categories'
import { formatCents, parseCents } from '../lib/money'
import { CategoryIcon } from '../lib/categoryIcons'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { useHaptic } from '../hooks/useHaptic'
import { EmptyState, SwipeToDelete } from '../components/ui'
import { useCategoryName } from '../hooks/useCategoryName'
import { useBaseCurrency } from '../hooks/useBaseCurrency'
import type { Budget } from '../types'

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

/* ─── Bottom Sheet ─── */
function BottomSheet({ onClose, children }: { onClose: () => void; children: React.ReactNode }) {
  return (
    <>
      <motion.div
        className="fixed inset-0 bg-black/40 z-[60]"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        onClick={onClose}
      />
      <motion.div
        className="fixed bottom-0 left-0 right-0 z-[60] bg-surface rounded-t-[24px] overflow-hidden"
        initial={{ y: '100%' }}
        animate={{ y: 0 }}
        exit={{ y: '100%' }}
        transition={{ type: 'spring', damping: 30, stiffness: 300 }}
        drag="y"
        dragConstraints={{ top: 0 }}
        dragElastic={{ top: 0, bottom: 0.3 }}
        onDragEnd={(_, info) => {
          if (info.velocity.y > 500 || info.offset.y > 100) onClose()
        }}
      >
        <div className="pt-3 pb-1 flex justify-center">
          <div className="w-10 h-1 rounded-full bg-border" />
        </div>
        {children}
      </motion.div>
    </>
  )
}

/* ─── Budget Card ─── */
function BudgetCard({
  budget,
  onEdit,
  onDelete,
}: {
  budget: Budget
  onEdit: (b: Budget) => void
  onDelete: (id: number) => void
}) {
  const { t } = useTranslation()
  const tCategory = useCategoryName()
  const { code: baseCurrency } = useBaseCurrency()
  const pct = Math.min(budget.usage_percent, 100)
  const barColor = pct >= 100 ? '#ef4444' : pct >= 80 ? '#f59e0b' : '#22c55e'

  return (
    <SwipeToDelete onDelete={() => onDelete(budget.id)}>
      <button
        onClick={() => onEdit(budget)}
        className="w-full px-4 py-4 text-left active:bg-accent-subtle/30 transition-colors"
      >
        {/* Header row */}
        <div className="flex items-center gap-3 mb-3">
          <div className="w-10 h-10 rounded-xl bg-accent-subtle flex items-center justify-center shrink-0">
            <CategoryIcon emoji={budget.category_emoji} size={20} weight="fill" className="text-accent" />
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <span className="text-[13px] font-bold text-text truncate">{tCategory(budget.category_name)}</span>
              {budget.is_over_limit && <Warning size={14} weight="fill" className="text-expense shrink-0" />}
            </div>
            <span className="text-[11px] font-semibold text-muted uppercase tracking-wide">{t(`budgets.${budget.period}`)}</span>
          </div>
          <span className="text-sm font-bold tabular-nums" style={{ color: barColor }}>
            {budget.usage_percent.toFixed(0)}%
          </span>
        </div>

        {/* Progress bar */}
        <div className="h-2 rounded-full overflow-hidden bg-border mb-2">
          <div
            className="h-full rounded-full transition-all duration-500"
            style={{ width: `${pct}%`, background: barColor }}
          />
        </div>

        {/* Spent / limit */}
        <div className="flex justify-between text-xs text-muted">
          <span>{t('budgets.spent')}: <span className="font-semibold text-text">{formatCents(budget.spent_cents, baseCurrency)}</span></span>
          <span>{t('budgets.limit')}: <span className="font-semibold text-text">{formatCents(budget.limit_cents, baseCurrency)}</span></span>
        </div>
      </button>
    </SwipeToDelete>
  )
}

/* ─── Budget Form (create or edit, bottom sheet) ─── */
function BudgetForm({
  editBudget,
  onClose,
}: {
  editBudget: Budget | null
  onClose: () => void
}) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const { notification, selection } = useHaptic()
  const { code: currencyCode, symbol } = useBaseCurrency()
  const tCategory = useCategoryName()

  const isEdit = editBudget !== null

  const [categoryID, setCategoryID] = useState<number | null>(editBudget?.category_id ?? null)
  const [limitStr, setLimitStr] = useState(editBudget ? String(editBudget.limit_cents / 100) : '')
  const [period, setPeriod] = useState(editBudget?.period ?? 'monthly')

  const catsQ = useQuery({ queryKey: ['categories'], queryFn: () => categoriesApi.list('expense') })
  const categories = catsQ.data?.categories ?? []

  const createMut = useMutation({
    mutationFn: () => createBudget({
      category_id: categoryID!,
      limit_cents: parseCents(limitStr),
      period,
      currency_code: currencyCode,
    }),
    onSuccess: () => {
      notification('success')
      qc.invalidateQueries({ queryKey: ['budgets'] })
      onClose()
    },
    onError: () => notification('error'),
  })

  const updateMut = useMutation({
    mutationFn: () => updateBudget(editBudget!.id, {
      category_id: categoryID ?? undefined,
      limit_cents: parseCents(limitStr),
      period,
    }),
    onSuccess: () => {
      notification('success')
      qc.invalidateQueries({ queryKey: ['budgets'] })
      onClose()
    },
    onError: () => notification('error'),
  })

  const mut = isEdit ? updateMut : createMut
  const canSubmit = categoryID !== null && parseCents(limitStr) > 0 && !mut.isPending

  return (
    <BottomSheet onClose={onClose}>
      <div
        className="px-5 space-y-4 overflow-y-auto no-scrollbar"
        style={{ maxHeight: '80dvh', paddingBottom: 'max(1.5rem, env(safe-area-inset-bottom))' }}
      >
        {/* Category picker */}
        <div>
            <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
              {t('transactions.category')}
            </label>
            <div className="grid grid-cols-4 gap-2">
              {categories.map(cat => (
                <button
                  key={cat.id}
                  onClick={() => { setCategoryID(cat.id); selection() }}
                  className={`
                    flex flex-col items-center gap-1 py-2.5 rounded-2xl text-xs transition-all duration-150 active:scale-95 select-none
                    ${categoryID === cat.id
                      ? 'bg-accent text-accent-text shadow-[0_2px_8px_rgba(99,102,241,0.3)]'
                      : 'bg-accent-subtle text-accent'
                    }
                  `}
                >
                  <CategoryIcon
                    emoji={cat.emoji}
                    size={18}
                    weight="fill"
                    className={categoryID === cat.id ? 'text-white' : 'text-accent'}
                  />
                  <span className="truncate w-full text-center px-1 font-medium text-[10px]">
                    {tCategory(cat.name)}
                  </span>
                </button>
              ))}
            </div>
          </div>

        {/* Limit amount */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('budgets.limit')}
          </label>
          <div className="flex items-baseline gap-1.5 bg-bg rounded-2xl px-4 py-3 focus-within:shadow-[0_0_0_2px_rgba(99,102,241,0.2)] transition-shadow">
            <span className="text-3xl font-bold text-muted/40 tabular-nums">{symbol}</span>
            <input
              inputMode="decimal"
              placeholder="0.00"
              value={limitStr}
              onChange={e => setLimitStr(sanitizeAmount(e.target.value))}
              className="flex-1 bg-transparent text-3xl font-bold outline-none text-text placeholder:text-muted/30 tabular-nums min-w-0"
            />
          </div>
        </div>

        {/* Period */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('budgets.period')}
          </label>
          <div className="flex gap-1.5">
            {(['weekly', 'monthly'] as const).map(p => (
              <button
                key={p}
                onClick={() => setPeriod(p)}
                className={`
                  flex-1 py-2.5 rounded-2xl text-[13px] font-bold transition-all duration-200 select-none
                  ${period === p
                    ? 'bg-accent text-accent-text shadow-[0_2px_8px_rgba(99,102,241,0.3)]'
                    : 'bg-accent-subtle text-muted'
                  }
                `}
              >
                {t(`budgets.${p}`)}
              </button>
            ))}
          </div>
        </div>

        {/* Submit */}
        <button
          onClick={() => mut.mutate()}
          disabled={!canSubmit}
          className={`
            w-full py-4 rounded-2xl text-[15px] font-bold transition-all active:scale-[0.98]
            ${canSubmit
              ? 'bg-accent text-accent-text shadow-[0_4px_16px_rgba(99,102,241,0.3)]'
              : 'bg-border text-muted'
            }
          `}
        >
          {mut.isPending ? '...' : isEdit ? t('common.save') : t('common.create')}
        </button>

        {mut.isError && (
          <p className="text-xs text-destructive text-center">
            {(mut.error as Error)?.message}
          </p>
        )}
      </div>
    </BottomSheet>
  )
}

/* ─── Main Page ─── */
export function BudgetsPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { notification } = useHaptic()
  useTgBackButton(() => navigate('/more'))

  const [editingBudget, setEditingBudget] = useState<Budget | null>(null)
  const [showCreate, setShowCreate] = useState(false)

  const budgetsQ = useQuery({ queryKey: ['budgets'], queryFn: fetchBudgets })

  const deleteMut = useMutation({
    mutationFn: deleteBudget,
    onSuccess: () => { notification('success'); qc.invalidateQueries({ queryKey: ['budgets'] }) },
    onError: () => notification('error'),
  })

  const budgets = budgetsQ.data?.budgets ?? []
  const formOpen = showCreate || editingBudget !== null

  if (budgetsQ.isPending) return <div className="flex justify-center py-16"><Spinner /></div>
  if (budgetsQ.isError) return <ErrorMessage onRetry={() => budgetsQ.refetch()} />

  return (
    <PageTransition>
      <div className="flex flex-col h-[calc(100dvh-var(--tab-bar-h))]">

        {/* Add button */}
        <div className="shrink-0 px-4 pt-3 pb-2 flex justify-end">
          <button
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-1.5 px-4 py-2 rounded-full bg-accent text-accent-text text-xs font-bold shadow-[0_2px_12px_rgba(99,102,241,0.4)] active:scale-95 transition-transform"
          >
            <Plus size={14} weight="bold" />
            {t('budgets.create_new')}
          </button>
        </div>

        {/* Scrollable list */}
        <div className="flex-1 min-h-0 overflow-y-auto no-scrollbar pb-3">
          {budgets.length === 0 ? (
            <div className="mx-4 card-elevated mt-2">
              <EmptyState icon="📊" title={t('budgets.no_budgets')} description={t('budgets.set_budget')} />
            </div>
          ) : (
            <div className="mx-4 card-elevated overflow-hidden divide-y divide-border">
              {budgets.map(b => (
                <BudgetCard
                  key={b.id}
                  budget={b}
                  onEdit={setEditingBudget}
                  onDelete={id => deleteMut.mutate(id)}
                />
              ))}
            </div>
          )}

          {deleteMut.isError && (
            <div className="mx-4 mt-2">
              <p className="text-xs text-destructive text-center bg-expense/10 rounded-xl py-2 px-3">
                {(deleteMut.error as Error)?.message}
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Bottom sheet form */}
      <AnimatePresence>
        {formOpen && (
          <BudgetForm
            key={editingBudget?.id ?? 'new'}
            editBudget={editingBudget}
            onClose={() => { setShowCreate(false); setEditingBudget(null) }}
          />
        )}
      </AnimatePresence>
    </PageTransition>
  )
}
