import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Plus, Trash2, AlertTriangle } from 'lucide-react'
import { fetchBudgets, createBudget, deleteBudget } from '../api/budgets'
import { categoriesApi } from '../api/categories'
import { formatCents } from '../lib/money'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { useHaptic } from '../hooks/useHaptic'
import { Button, EmptyState, SegmentedControl } from '../components/ui'
import type { Budget } from '../types'

function BudgetCard({ budget, onDelete }: { budget: Budget; onDelete: (id: number) => void }) {
  const { t } = useTranslation()
  const pct      = Math.min(budget.usage_percent, 100)
  const barColor = pct >= 100 ? '#ef4444' : pct >= 80 ? '#f59e0b' : '#22c55e'

  return (
    <div className="px-4 py-4 border-b border-border last:border-b-0">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-border flex items-center justify-center text-base">
            {budget.category_emoji}
          </div>
          <span className="text-sm font-semibold text-text">{budget.category_name}</span>
          {budget.is_over_limit && <AlertTriangle size={14} className="text-expense" />}
        </div>
        <button onClick={() => onDelete(budget.id)} className="p-1 text-muted active:text-destructive">
          <Trash2 size={15} />
        </button>
      </div>

      <div className="flex justify-between text-xs text-muted mb-2">
        <span>{t('budgets.spent')}: <span className="font-medium text-text">{formatCents(budget.spent_cents, budget.currency_code)}</span></span>
        <span>{t('budgets.limit')}: <span className="font-medium text-text">{formatCents(budget.limit_cents, budget.currency_code)}</span></span>
      </div>

      <div className="h-2 rounded-full overflow-hidden bg-border">
        <div
          className="h-full rounded-full transition-all duration-500"
          style={{ width: `${pct}%`, background: barColor }}
        />
      </div>

      <div className="flex justify-between items-center mt-1.5">
        <span className="text-[10px] font-semibold text-muted uppercase tracking-wide">{budget.period}</span>
        <span className="text-xs font-bold" style={{ color: barColor }}>
          {budget.usage_percent.toFixed(0)}%
        </span>
      </div>
    </div>
  )
}

export function BudgetsPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { notification, selection } = useHaptic()
  useTgBackButton(() => navigate('/more'))

  const [showForm, setShowForm]     = useState(false)
  const [categoryID, setCategoryID] = useState<number | null>(null)
  const [limitStr, setLimitStr]     = useState('')
  const [period, setPeriod]         = useState('monthly')

  const budgetsQ = useQuery({ queryKey: ['budgets'], queryFn: fetchBudgets })
  const catsQ    = useQuery({ queryKey: ['categories'], queryFn: () => categoriesApi.list('expense') })

  const createMut = useMutation({
    mutationFn: () => {
      const cents = Math.round(parseFloat(limitStr) * 100)
      return createBudget({ category_id: categoryID!, limit_cents: cents, period, currency_code: 'USD' })
    },
    onSuccess: () => {
      notification('success')
      qc.invalidateQueries({ queryKey: ['budgets'] })
      resetForm()
    },
    onError: () => notification('error'),
  })

  const deleteMut = useMutation({
    mutationFn: deleteBudget,
    onSuccess: () => { notification('success'); qc.invalidateQueries({ queryKey: ['budgets'] }) },
  })

  function resetForm() {
    setShowForm(false); setCategoryID(null); setLimitStr(''); setPeriod('monthly')
  }

  const budgets    = budgetsQ.data?.budgets ?? []
  const categories = catsQ.data?.categories ?? []
  const canSubmit  = categoryID !== null && parseFloat(limitStr) > 0 && !createMut.isPending

  if (budgetsQ.isPending) return <div className="flex justify-center py-16"><Spinner /></div>
  if (budgetsQ.isError)   return <ErrorMessage onRetry={() => budgetsQ.refetch()} />

  return (
    <PageTransition>
      <div className="p-4 space-y-4">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-bold">{t('budgets.title')}</h1>
          {!showForm && (
            <Button size="sm" onClick={() => setShowForm(true)}>
              <Plus size={15} className="mr-1" /> {t('budgets.create_new')}
            </Button>
          )}
        </div>

        {showForm && (
          <div className="bg-surface rounded-[--radius-card] p-4 space-y-4">
            <div>
              <label className="block text-xs font-semibold text-muted uppercase tracking-widest mb-2">
                {t('transactions.category')}
              </label>
              <div className="grid grid-cols-4 gap-2">
                {categories.map(cat => (
                  <button
                    key={cat.id}
                    onClick={() => { setCategoryID(cat.id); selection() }}
                    className={`flex flex-col items-center gap-1 py-2.5 rounded-[--radius-sm] text-xs transition-all active:scale-95
                      ${categoryID === cat.id ? 'bg-accent-subtle text-accent ring-2 ring-accent ring-inset' : 'text-text active:bg-border'}`}
                  >
                    <span className="text-xl">{cat.emoji}</span>
                    <span className="truncate w-full text-center px-1 font-medium">{cat.name}</span>
                  </button>
                ))}
              </div>
            </div>

            <div>
              <label className="block text-xs font-semibold text-muted uppercase tracking-widest mb-2">
                {t('budgets.limit')}
              </label>
              <input
                inputMode="decimal"
                placeholder="0.00"
                value={limitStr}
                onChange={e => setLimitStr(e.target.value)}
                className="w-full bg-bg rounded-[--radius-sm] px-3 py-2.5 text-2xl font-bold outline-none focus:ring-2 focus:ring-accent tabular-nums"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-muted uppercase tracking-widest mb-2">
                {t('budgets.period')}
              </label>
              <SegmentedControl
                options={[
                  { value: 'weekly',  label: t('budgets.weekly') },
                  { value: 'monthly', label: t('budgets.monthly') },
                ]}
                value={period}
                onChange={setPeriod}
                size="sm"
              />
            </div>

            <div className="flex gap-2 pt-1">
              <Button size="sm" onClick={() => createMut.mutate()} disabled={!canSubmit}>
                {t('common.create')}
              </Button>
              <Button size="sm" variant="ghost" onClick={resetForm}>{t('common.cancel')}</Button>
            </div>

            {createMut.isError && (
              <p className="text-xs text-destructive">{(createMut.error as Error)?.message}</p>
            )}
          </div>
        )}

        {budgets.length > 0 ? (
          <div className="bg-surface rounded-[--radius-card] overflow-hidden">
            {budgets.map(b => (
              <BudgetCard key={b.id} budget={b} onDelete={id => deleteMut.mutate(id)} />
            ))}
          </div>
        ) : !showForm ? (
          <EmptyState icon="📊" title={t('budgets.no_budgets')} description={t('budgets.set_budget')} />
        ) : null}
      </div>
    </PageTransition>
  )
}
