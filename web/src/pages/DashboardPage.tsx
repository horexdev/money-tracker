import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { TrendingUp, TrendingDown, ArrowRight } from 'lucide-react'
import { balanceApi } from '../api/balance'
import { transactionsApi } from '../api/transactions'
import { formatCents } from '../lib/money'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { Card, SectionHeader, TransactionRow, EmptyState } from '../components/ui'

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
      <div className="p-4 space-y-4">
        {/* Hero balance card */}
        <Card className="bg-gradient-to-br from-accent/10 to-accent/5">
          <p className="text-xs text-muted uppercase tracking-wide">{t('dashboard.net_balance')}</p>
          <p className="text-3xl font-bold mt-1">
            {primary ? formatCents(primary.net_cents, primary.currency_code) : '$0.00'}
          </p>
          {primary && (
            <div className="flex gap-4 mt-3 text-sm">
              <span className="flex items-center gap-1 text-income">
                <TrendingUp size={14} />
                {formatCents(primary.income_cents, primary.currency_code)}
              </span>
              <span className="flex items-center gap-1 text-expense">
                <TrendingDown size={14} />
                {formatCents(primary.expense_cents, primary.currency_code)}
              </span>
            </div>
          )}
        </Card>

        {/* Recent transactions */}
        <div>
          <div className="flex justify-between items-center mb-2">
            <SectionHeader>{t('dashboard.recent_transactions')}</SectionHeader>
            <Link to="/history" className="text-xs text-accent flex items-center gap-0.5">
              {t('dashboard.view_all')} <ArrowRight size={12} />
            </Link>
          </div>
          <Card padding="p-0">
            {txQ.isPending ? (
              <div className="flex justify-center py-6"><Spinner size="sm" /></div>
            ) : txQ.data?.transactions.length ? (
              <div className="divide-y divide-border">
                {txQ.data.transactions.map(tx => (
                  <TransactionRow key={tx.id} tx={tx} />
                ))}
              </div>
            ) : (
              <EmptyState title={t('transactions.no_transactions')} />
            )}
          </Card>
        </div>
      </div>
    </PageTransition>
  )
}
