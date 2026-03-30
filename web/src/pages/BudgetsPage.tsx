import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { AnimatePresence } from 'framer-motion'
import { Plus, Warning, ClockCounterClockwise, ArrowCircleDown, ChartBar, Bell, BellSlash } from '@phosphor-icons/react'
import { fetchBudgets, createBudget, updateBudget, deleteBudget, fetchBudgetTransactions } from '../api/budgets'
import type { BudgetTransaction } from '../api/budgets'
import { categoriesApi } from '../api/categories'
import { formatCents, parseCents } from '../lib/money'
import { friendlyError } from '../lib/errors'
import { CategoryIcon } from '../lib/categoryIcons'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { useHaptic } from '../hooks/useHaptic'
import { EmptyState, SwipeToDelete, FAB, BottomSheet } from '../components/ui'
import { useCategoryName } from '../hooks/useCategoryName'
import { useBaseCurrency } from '../hooks/useBaseCurrency'
import { CurrencyBadge } from '../lib/currencyIcons'
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

/* ─── Budget Transactions Sheet ─── */
function BudgetTransactionsSheet({ budget, onClose }: { budget: Budget; onClose: () => void }) {
  const { t } = useTranslation()
  const tCategory = useCategoryName()
  const { code: baseCurrency } = useBaseCurrency()

  const q = useQuery({
    queryKey: ['budget-transactions', budget.id],
    queryFn: () => fetchBudgetTransactions(budget.id),
  })

  const txs: BudgetTransaction[] = q.data?.transactions ?? []

  function fmtDate(iso: string) {
    const d = new Date(iso)
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
  }

  return (
    <BottomSheet onClose={onClose}>
      <div
        className="px-5 overflow-y-auto no-scrollbar"
        style={{ maxHeight: '80dvh', paddingBottom: 'max(1.5rem, env(safe-area-inset-bottom))' }}
      >
        <div className="flex items-center gap-2 mb-4">
          <div
            className="w-8 h-8 rounded-2xl flex items-center justify-center shrink-0"
            style={{ background: budget.category_color || 'var(--color-accent)' }}
          >
            <CategoryIcon emoji={budget.category_emoji} size={16} weight="fill" className="text-white" />
          </div>
          <h2 className="text-[15px] font-bold text-text">{tCategory(budget.category_name)}</h2>
        </div>

        {q.isPending && <div className="flex justify-center py-8"><Spinner /></div>}
        {q.isError && <ErrorMessage onRetry={() => q.refetch()} />}

        {!q.isPending && !q.isError && txs.length === 0 && (
          <div className="py-8 text-center text-muted text-sm">{t('budgets.no_transactions')}</div>
        )}

        {txs.length > 0 && (
          <div className="space-y-0 divide-y divide-border rounded-2xl overflow-hidden bg-bg">
            {txs.map(tx => (
              <div key={tx.id} className="flex items-center gap-3 px-4 py-3">
                <div
                  className="w-8 h-8 rounded-2xl flex items-center justify-center shrink-0"
                  style={{ background: tx.category_color || 'var(--color-accent)' }}
                >
                  <ArrowCircleDown size={16} weight="fill" className="text-white" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-[13px] font-semibold text-text truncate">
                    {tx.note || tCategory(tx.category_name)}
                  </p>
                  <p className="text-[11px] text-muted">{fmtDate(tx.created_at)}</p>
                </div>
                <span className="text-[13px] font-bold text-expense tabular-nums">
                  -{formatCents(tx.amount_cents, baseCurrency)}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </BottomSheet>
  )
}

/* ─── Budget Card ─── */
function BudgetCard({
  budget,
  onEdit,
  onDelete,
  onHistory,
}: {
  budget: Budget
  onEdit: (b: Budget) => void
  onDelete: (id: number) => void
  onHistory: (b: Budget) => void
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
          <div
            className="w-10 h-10 rounded-2xl flex items-center justify-center shrink-0"
            style={{ background: budget.category_color || 'var(--color-accent)' }}
          >
            <CategoryIcon emoji={budget.category_emoji} size={20} weight="fill" className="text-white" />
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <span className="text-[13px] font-bold text-text truncate">{tCategory(budget.category_name)}</span>
              {budget.is_over_limit && <Warning size={14} weight="fill" className="text-expense shrink-0" />}
            </div>
            <span className="text-[11px] font-semibold text-muted uppercase tracking-wide">{t(`budgets.${budget.period}`)}</span>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={e => { e.stopPropagation(); onHistory(budget) }}
              className="p-1.5 rounded-2xl text-muted active:bg-accent-subtle transition-colors"
              aria-label={t('budgets.view_transactions')}
            >
              <ClockCounterClockwise size={16} weight="bold" />
            </button>
            <span className="text-sm font-bold tabular-nums" style={{ color: barColor }}>
              {budget.usage_percent.toFixed(0)}%
            </span>
          </div>
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
  const { code: currencyCode } = useBaseCurrency()
  const tCategory = useCategoryName()

  const isEdit = editBudget !== null

  const [categoryID, setCategoryID] = useState<number | null>(editBudget?.category_id ?? null)
  const [limitStr, setLimitStr] = useState(editBudget ? String(editBudget.limit_cents / 100) : '')
  const [period, setPeriod] = useState(editBudget?.period ?? 'monthly')
  const [notificationsEnabled, setNotificationsEnabled] = useState(editBudget?.notifications_enabled ?? true)

  const catsQ = useQuery({ queryKey: ['categories'], queryFn: () => categoriesApi.list('expense') })
  const categories = catsQ.data?.categories ?? []

  const createMut = useMutation({
    mutationFn: () => createBudget({
      category_id: categoryID!,
      limit_cents: parseCents(limitStr),
      period,
      currency_code: currencyCode,
      notifications_enabled: notificationsEnabled,
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
      notifications_enabled: notificationsEnabled,
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
                    flex flex-col items-center gap-1.5 py-2.5 rounded-2xl text-xs transition-all duration-150 active:scale-95 select-none
                    ${categoryID === cat.id
                      ? 'bg-accent/10 shadow-[0_2px_8px_rgba(99,102,241,0.15)]'
                      : 'bg-surface shadow-sm'
                    }
                  `}
                >
                  <div
                    className="w-9 h-9 rounded-2xl flex items-center justify-center"
                    style={{ background: categoryID === cat.id ? 'var(--color-accent)' : (cat.color || 'var(--color-accent)') }}
                  >
                    <CategoryIcon emoji={cat.emoji} size={18} weight="fill" className="text-white" />
                  </div>
                  <span className="truncate w-full text-center px-1 font-medium text-[10px] text-text">
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
            <CurrencyBadge currency={currencyCode} className="text-muted/40" />
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

        {/* Notifications toggle */}
        <button
          type="button"
          onClick={() => setNotificationsEnabled(v => !v)}
          className="w-full flex items-center justify-between px-4 py-3 rounded-2xl bg-surface transition-colors active:bg-accent-subtle/30"
        >
          <div className="flex items-center gap-3">
            {notificationsEnabled
              ? <Bell size={18} weight="fill" className="text-accent" />
              : <BellSlash size={18} weight="fill" className="text-muted" />
            }
            <div className="text-left">
              <p className="text-[13px] font-semibold text-text">{t('budgets.notifications')}</p>
              <p className="text-[11px] text-muted">{t('budgets.notify_threshold')}</p>
            </div>
          </div>
          <div className={`relative w-11 h-6 rounded-full transition-colors duration-200 shrink-0 ${notificationsEnabled ? 'bg-accent' : 'bg-border'}`}>
            <div className={`absolute top-1 w-4 h-4 bg-white rounded-full shadow transition-transform duration-200 ${notificationsEnabled ? 'translate-x-6' : 'translate-x-1'}`} />
          </div>
        </button>

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
          {mut.isPending ? t('common.loading') : isEdit ? t('common.save') : t('common.create')}
        </button>

        {mut.isError && (
          <p className="text-xs text-destructive text-center">
            {friendlyError(mut.error, t)}
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
  const [historyFor, setHistoryFor] = useState<Budget | null>(null)

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
      <div className="pt-3 pb-4">
          {budgets.length === 0 ? (
            <div className="mx-4 card-elevated mt-2">
              <EmptyState
                icon={ChartBar}
                title={t('budgets.no_budgets')}
                description={t('budgets.set_budget')}
                action={
                  <button
                    onClick={() => setShowCreate(true)}
                    className="flex items-center gap-1.5 px-5 py-2.5 rounded-full bg-accent text-accent-text text-xs font-bold shadow-[0_2px_12px_rgba(99,102,241,0.4)] active:scale-95 transition-transform"
                  >
                    <Plus size={14} weight="bold" />
                    {t('budgets.create_new')}
                  </button>
                }
              />
            </div>
          ) : (
            <div className="mx-4 card-elevated divide-y divide-border">
              {budgets.map(b => (
                <BudgetCard
                  key={b.id}
                  budget={b}
                  onEdit={setEditingBudget}
                  onDelete={id => deleteMut.mutate(id)}
                  onHistory={setHistoryFor}
                />
              ))}
            </div>
          )}

          {deleteMut.isError && (
            <div className="mx-4 mt-2">
              <p className="text-xs text-destructive text-center bg-expense/10 rounded-2xl py-2 px-3">
                {friendlyError(deleteMut.error, t)}
              </p>
            </div>
          )}
      </div>

      <FAB onClick={() => setShowCreate(true)} label={t('budgets.create_new')} />

      {/* Bottom sheet form */}
      <AnimatePresence>
        {formOpen && (
          <BudgetForm
            key={editingBudget?.id ?? 'new'}
            editBudget={editingBudget}
            onClose={() => { setShowCreate(false); setEditingBudget(null) }}
          />
        )}
        {historyFor && (
          <BudgetTransactionsSheet
            key={`history-${historyFor.id}`}
            budget={historyFor}
            onClose={() => setHistoryFor(null)}
          />
        )}
      </AnimatePresence>
    </PageTransition>
  )
}
