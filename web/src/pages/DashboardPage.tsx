import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { AnimatePresence } from 'framer-motion'
import { formatDate, formatCents } from '../lib/money'
import { TrendUp, TrendDown, ArrowRight, Plus, CaretDown } from '@phosphor-icons/react'
import { balanceApi } from '../api/balance'
import { transactionsApi } from '../api/transactions'
import { useBaseCurrency } from '../hooks/useBaseCurrency'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { TransactionRow, EditTransactionSheet } from '../components/ui'
import type { Transaction } from '../types'

export function DashboardPage() {
  const { t, i18n } = useTranslation()
  const navigate = useNavigate()
  const [showCurrencyBreakdown, setShowCurrencyBreakdown] = useState(false)
  const [editingTx, setEditingTx] = useState<Transaction | null>(null)

  const balanceQ = useQuery({ queryKey: ['balance'], queryFn: balanceApi.get })
  const txQ      = useQuery({ queryKey: ['transactions', 1], queryFn: () => transactionsApi.list(1, 5) })
  const { code: baseCurrency } = useBaseCurrency()

  if (balanceQ.isPending) return (
    <div className="flex justify-center items-center pt-24"><Spinner /></div>
  )
  if (balanceQ.isError) return <ErrorMessage onRetry={() => balanceQ.refetch()} />
  const balance = balanceQ.data
  const isMultiCurrency = (balance?.by_currency?.length ?? 0) > 1
  const totalInBase = balance?.total_in_base_cents ?? 0
  // Entry matching base currency for income/expense cards
  const baseEntry = balance?.by_currency?.find(b => b.currency_code === baseCurrency)
    ?? balance?.by_currency?.[0]

  return (
    <PageTransition>
      <div className="flex flex-col h-[calc(100dvh-var(--tab-bar-h))]">

        {/* Fixed top section */}
        <div className="px-4 pt-4 flex flex-col gap-3 shrink-0">

          {/* Hero card */}
          <div className="hero-gradient p-6 relative"
               style={{ boxShadow: 'var(--shadow-hero)' }}>
            <div className="absolute -top-12 -right-12 w-40 h-40 rounded-full bg-white/[0.07] blur-xl pointer-events-none" />
            <div className="absolute -bottom-10 -left-10 w-32 h-32 rounded-full bg-indigo-400/20 blur-2xl pointer-events-none" />
            <div className="absolute top-1/2 right-1/3 w-20 h-20 rounded-full bg-white/[0.04] pointer-events-none" />

            <div className="relative z-10">
              <p className="text-white/60 text-[11px] font-bold uppercase tracking-[0.2em]">
                {t('dashboard.net_balance')}
              </p>
              <div className="flex items-baseline gap-2 mt-1">
                {isMultiCurrency && (
                  <span className="text-white/50 text-2xl font-bold">≈</span>
                )}
                <p className="text-white text-[42px] font-extrabold tabular-nums leading-none tracking-tight">
                  {formatCents(totalInBase, baseCurrency)}
                </p>
              </div>
              {isMultiCurrency ? (
                <button
                  onClick={() => setShowCurrencyBreakdown(v => !v)}
                  className="flex items-center gap-1 text-white/50 text-xs font-medium mt-2"
                >
                  {t('dashboard.converted_from_multiple')}
                  <CaretDown
                    size={10}
                    weight="bold"
                    className={`transition-transform ${showCurrencyBreakdown ? 'rotate-180' : ''}`}
                  />
                </button>
              ) : (
                <p className="text-white/40 text-xs font-medium mt-2">
                  {formatDate(new Date(), i18n.language, { month: 'long', year: 'numeric' })}
                </p>
              )}

              {/* Per-currency breakdown */}
              {isMultiCurrency && showCurrencyBreakdown && (
                <div className="mt-3 pt-3 border-t border-white/10 space-y-1.5">
                  {balance!.by_currency.map((b) => (
                    <div key={b.currency_code} className="flex justify-between text-xs">
                      <span className="text-white/50 font-semibold">{b.currency_code}</span>
                      <span className={`font-bold tabular-nums ${b.net_cents >= 0 ? 'text-white/80' : 'text-expense/80'}`}>
                        {formatCents(b.net_cents, b.currency_code)}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <Link
              to="/add"
              className="absolute top-5 right-5 z-10 w-11 h-11 rounded-2xl bg-white/15 backdrop-blur-sm flex items-center justify-center active:scale-90 transition-transform border border-white/10"
            >
              <Plus size={20} weight="bold" className="text-white" />
            </Link>
          </div>

          {/* Income / Expense bento cards */}
          {baseEntry && (
            <div className="grid grid-cols-2 gap-3">
              <button
                onClick={() => navigate('/stats', { state: { type: 'income' } })}
                className="card-elevated p-4 relative overflow-hidden text-left active:scale-[0.97] transition-transform"
              >
                <div className="absolute top-0 left-0 w-1 h-full rounded-l-[--radius-card] bg-income" />
                <div className="absolute -top-6 -right-6 w-16 h-16 rounded-full bg-income/[0.06] pointer-events-none" />
                <div className="flex items-center gap-2 mb-2">
                  <div className="w-8 h-8 rounded-2xl bg-income/15 flex items-center justify-center">
                    <TrendUp size={16} weight="bold" className="text-income" />
                  </div>
                  <span className="text-xs font-semibold text-muted">{t('transactions.income')}</span>
                  <ArrowRight size={12} weight="bold" className="text-muted/40 ml-auto" />
                </div>
                <p className="text-income text-lg font-bold tabular-nums leading-tight">
                  {formatCents(baseEntry.income_cents, baseCurrency)}
                </p>
              </button>

              <button
                onClick={() => navigate('/stats', { state: { type: 'expense' } })}
                className="card-elevated p-4 relative overflow-hidden text-left active:scale-[0.97] transition-transform"
              >
                <div className="absolute top-0 left-0 w-1 h-full rounded-l-[--radius-card] bg-expense" />
                <div className="absolute -top-6 -right-6 w-16 h-16 rounded-full bg-expense/[0.06] pointer-events-none" />
                <div className="flex items-center gap-2 mb-2">
                  <div className="w-8 h-8 rounded-2xl bg-expense/15 flex items-center justify-center">
                    <TrendDown size={16} weight="bold" className="text-expense" />
                  </div>
                  <span className="text-xs font-semibold text-muted">{t('transactions.expense')}</span>
                  <ArrowRight size={12} weight="bold" className="text-muted/40 ml-auto" />
                </div>
                <p className="text-expense text-lg font-bold tabular-nums leading-tight">
                  {formatCents(baseEntry.expense_cents, baseCurrency)}
                </p>
              </button>
            </div>
          )}
        </div>

        {/* Recent transactions — fills remaining space, scrolls internally */}
        <div className="flex-1 min-h-0 px-4 pt-3 pb-2 flex flex-col">
          <div className="card-elevated overflow-hidden flex-1 min-h-0 flex flex-col">
            <div className="flex justify-between items-center px-5 pt-4 pb-2 shrink-0">
              <span className="text-sm font-bold text-text">
                {t('dashboard.recent_transactions')}
              </span>
              <Link to="/history" className="flex items-center gap-1 text-xs text-accent font-semibold">
                {t('dashboard.view_all')} <ArrowRight size={14} weight="bold" />
              </Link>
            </div>

            {txQ.isPending ? (
              <div className="flex justify-center py-10"><Spinner size="sm" /></div>
            ) : txQ.data?.transactions.length ? (
              <div className="overflow-y-auto no-scrollbar flex-1">
                <div className="divide-y divide-border">
                  {txQ.data.transactions.map(tx => (
                    <TransactionRow key={tx.id} tx={tx} onEdit={setEditingTx} />
                  ))}
                </div>
              </div>
            ) : (
              <p className="text-center text-sm text-muted py-12">
                {t('transactions.no_transactions')}
              </p>
            )}
          </div>
        </div>

      </div>

      <AnimatePresence>
        {editingTx && (
          <EditTransactionSheet
            key={editingTx.id}
            tx={editingTx}
            onClose={() => setEditingTx(null)}
          />
        )}
      </AnimatePresence>
    </PageTransition>
  )
}
