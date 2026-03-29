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
  const txQ = useQuery({ queryKey: ['transactions', 1], queryFn: () => transactionsApi.list(1, 5) })

  if (balanceQ.isPending) return <div className="flex justify-center pt-20"><Spinner /></div>
  if (balanceQ.isError) return <ErrorMessage onRetry={() => balanceQ.refetch()} />

  const balance = balanceQ.data
  const primary = balance.by_currency[0]

  return (
    <PageTransition>
      <div className="px-4 pt-4 pb-2 space-y-3">

        {/* Bento row 1: Aurora hero (full width) */}
        <div className="aurora-bg rounded-[--radius-card] p-5 relative overflow-hidden">
          {/* Decorative blobs */}
          <div className="absolute -top-8 -right-8 w-32 h-32 rounded-full bg-white/10 blur-2xl pointer-events-none" />
          <div className="absolute -bottom-6 -left-6 w-24 h-24 rounded-full bg-white/10 blur-2xl pointer-events-none" />

          <p className="text-white/70 text-xs font-medium uppercase tracking-widest">
            {t('dashboard.net_balance')}
          </p>
          <p className="text-white text-4xl font-bold mt-1 tabular-nums drop-shadow-sm">
            {primary ? formatCents(primary.net_cents, primary.currency_code) : '$0.00'}
          </p>

          <Link
            to="/add"
            className="absolute top-4 right-4 w-9 h-9 rounded-full bg-white/20 flex items-center justify-center active:scale-95 transition-transform"
          >
            <Plus size={18} className="text-white" />
          </Link>
        </div>

        {/* Bento row 2: income + expense side by side */}
        {primary && (
          <div className="grid grid-cols-2 gap-3">
            <div className="bg-income-subtle rounded-[--radius-card] p-4">
              <div className="flex items-center gap-1.5 mb-2">
                <div className="w-7 h-7 rounded-full bg-income/20 flex items-center justify-center">
                  <TrendingUp size={14} className="text-income" />
                </div>
                <span className="text-xs text-muted font-medium">{t('transactions.income')}</span>
              </div>
              <p className="text-income text-lg font-bold tabular-nums leading-tight">
                {formatCents(primary.income_cents, primary.currency_code)}
              </p>
            </div>

            <div className="bg-expense-subtle rounded-[--radius-card] p-4">
              <div className="flex items-center gap-1.5 mb-2">
                <div className="w-7 h-7 rounded-full bg-expense/20 flex items-center justify-center">
                  <TrendingDown size={14} className="text-expense" />
                </div>
                <span className="text-xs text-muted font-medium">{t('transactions.expense')}</span>
              </div>
              <p className="text-expense text-lg font-bold tabular-nums leading-tight">
                {formatCents(primary.expense_cents, primary.currency_code)}
              </p>
            </div>
          </div>
        )}

        {/* Bento row 3: recent transactions */}
        <div className="bg-surface rounded-[--radius-card] overflow-hidden">
          <div className="flex justify-between items-center px-4 pt-3 pb-1">
            <span className="text-xs font-semibold uppercase tracking-widest text-muted">
              {t('dashboard.recent_transactions')}
            </span>
            <Link to="/history" className="text-xs text-accent flex items-center gap-0.5 font-medium">
              {t('dashboard.view_all')} <ArrowRight size={12} />
            </Link>
          </div>

          {txQ.isPending ? (
            <div className="flex justify-center py-6"><Spinner size="sm" /></div>
          ) : txQ.data?.transactions.length ? (
            <div className="divide-y divide-border">
              {txQ.data.transactions.map(tx => (
                <TransactionRow key={tx.id} tx={tx} />
              ))}
            </div>
          ) : (
            <p className="text-center text-sm text-muted py-8">{t('transactions.no_transactions')}</p>
          )}
        </div>

      </div>
    </PageTransition>
  )
}
