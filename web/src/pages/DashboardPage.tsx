import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { TrendingUp, TrendingDown, ArrowRight, Plus } from 'lucide-react'
import { balanceApi } from '../api/balance'
import { transactionsApi } from '../api/transactions'
import { formatCents } from '../lib/money'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { TransactionRow } from '../components/ui'

export function DashboardPage() {
  const { t } = useTranslation()

  const balanceQ = useQuery({ queryKey: ['balance'], queryFn: balanceApi.get })
  const txQ      = useQuery({ queryKey: ['transactions', 1], queryFn: () => transactionsApi.list(1, 5) })

  if (balanceQ.isPending) return (
    <div className="flex justify-center items-center pt-24"><Spinner /></div>
  )
  if (balanceQ.isError) return <ErrorMessage onRetry={() => balanceQ.refetch()} />

  const balance = balanceQ.data
  const primary = balance?.by_currency?.[0]

  return (
    <PageTransition>
      <div className="px-4 pt-4 pb-2 flex flex-col gap-3">

        {/* Hero card — aurora */}
        <div className="aurora-bg rounded-[--radius-card] p-5 relative overflow-hidden">
          <div className="absolute -top-10 -right-10 w-36 h-36 rounded-full bg-white/10 blur-2xl pointer-events-none" />
          <div className="absolute -bottom-8 -left-8 w-28 h-28 rounded-full bg-white/10 blur-2xl pointer-events-none" />

          <p className="text-white/70 text-xs font-semibold uppercase tracking-widest">
            {t('dashboard.net_balance')}
          </p>
          <p className="text-white text-4xl font-bold mt-1.5 tabular-nums leading-none">
            {primary ? formatCents(primary.net_cents, primary.currency_code) : '$0.00'}
          </p>
          <p className="text-white/50 text-xs mt-1">
            {new Date().toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}
          </p>

          <Link
            to="/add"
            className="absolute top-4 right-4 w-9 h-9 rounded-2xl bg-white/20 flex items-center justify-center active:scale-90 transition-transform"
          >
            <Plus size={18} className="text-white" />
          </Link>
        </div>

        {/* Income / Expense row */}
        {primary && (
          <div className="grid grid-cols-2 gap-3">
            <div className="bg-income-subtle rounded-[--radius-card] p-4">
              <div className="flex items-center gap-2 mb-2.5">
                <div className="w-7 h-7 rounded-full bg-income/20 flex items-center justify-center">
                  <TrendingUp size={14} className="text-income" />
                </div>
                <span className="text-xs font-medium text-muted">{t('transactions.income')}</span>
              </div>
              <p className="text-income text-lg font-bold tabular-nums leading-tight">
                {formatCents(primary.income_cents, primary.currency_code)}
              </p>
            </div>

            <div className="bg-expense-subtle rounded-[--radius-card] p-4">
              <div className="flex items-center gap-2 mb-2.5">
                <div className="w-7 h-7 rounded-full bg-expense/20 flex items-center justify-center">
                  <TrendingDown size={14} className="text-expense" />
                </div>
                <span className="text-xs font-medium text-muted">{t('transactions.expense')}</span>
              </div>
              <p className="text-expense text-lg font-bold tabular-nums leading-tight">
                {formatCents(primary.expense_cents, primary.currency_code)}
              </p>
            </div>
          </div>
        )}

        {/* Recent transactions */}
        <div className="bg-surface rounded-[--radius-card] overflow-hidden">
          <div className="flex justify-between items-center px-4 pt-3 pb-1">
            <span className="text-xs font-semibold uppercase tracking-widest text-muted">
              {t('dashboard.recent_transactions')}
            </span>
            <Link to="/history" className="flex items-center gap-0.5 text-xs text-accent font-medium">
              {t('dashboard.view_all')} <ArrowRight size={12} />
            </Link>
          </div>

          {txQ.isPending ? (
            <div className="flex justify-center py-8"><Spinner size="sm" /></div>
          ) : txQ.data?.transactions.length ? (
            <div className="divide-y divide-border">
              {txQ.data.transactions.map(tx => (
                <TransactionRow key={tx.id} tx={tx} />
              ))}
            </div>
          ) : (
            <p className="text-center text-sm text-muted py-10">
              {t('transactions.no_transactions')}
            </p>
          )}
        </div>

      </div>
    </PageTransition>
  )
}
