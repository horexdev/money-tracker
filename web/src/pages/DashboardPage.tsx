import { useState, useEffect, useRef } from 'react'
import { useQuery, keepPreviousData } from '@tanstack/react-query'
import { Link, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { motion, AnimatePresence, useSpring, useMotionValueEvent } from 'framer-motion'
import { formatDate, formatCents } from '../lib/money'
import { TrendUp, TrendDown, ArrowRight, Plus, CaretDown, Receipt } from '@phosphor-icons/react'
import { balanceApi } from '../api/balance'
import { transactionsApi } from '../api/transactions'
import { accountsApi } from '../api/accounts'
import { useBaseCurrency } from '../hooks/useBaseCurrency'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { TransactionRow, EditTransactionSheet, AccountDropdown, EmptyState } from '../components/ui'
import type { Transaction } from '../types'

/** Animates a numeric cents value with a spring, formatted via formatCents */
function AnimatedMoney({ cents, currency, className }: { cents: number; currency: string; className?: string }) {
  const spring = useSpring(cents, { stiffness: 70, damping: 18 })
  const formatterRef = useRef((v: number) => formatCents(Math.round(v), currency))
  const [display, setDisplay] = useState(() => formatCents(cents, currency))

  // Update formatter ref when currency changes
  formatterRef.current = (v: number) => formatCents(Math.round(v), currency)

  useEffect(() => {
    spring.set(cents)
  }, [cents, spring])

  // When currency changes, snap immediately (avoid showing wrong currency symbol during animation)
  useEffect(() => {
    setDisplay(formatCents(Math.round(spring.get()), currency))
  }, [currency]) // eslint-disable-line react-hooks/exhaustive-deps

  useMotionValueEvent(spring, 'change', (v) => {
    setDisplay(formatterRef.current(v))
  })

  return <motion.span className={className}>{display}</motion.span>
}

export function DashboardPage() {
  const { t, i18n } = useTranslation()
  const navigate = useNavigate()
  const [showCurrencyBreakdown, setShowCurrencyBreakdown] = useState(false)
  const [editingTx, setEditingTx] = useState<Transaction | null>(null)
  const [selectedAccountId, setSelectedAccountId] = useState<number | null>(null)

  const { data: accounts = [] } = useQuery({
    queryKey: ['accounts'],
    queryFn: accountsApi.list,
  })

  // Pre-select the default account once accounts load
  useEffect(() => {
    if (selectedAccountId === null && accounts.length > 0) {
      const def = accounts.find(a => a.is_default) ?? accounts[0]
      setSelectedAccountId(def.id)
    }
  }, [accounts, selectedAccountId])

  const balanceQ = useQuery({
    queryKey: ['balance', selectedAccountId],
    queryFn: () => balanceApi.get(selectedAccountId),
    placeholderData: keepPreviousData,
  })

  const txQ = useQuery({
    queryKey: ['transactions', 1, selectedAccountId],
    queryFn: () => transactionsApi.list(1, 5, selectedAccountId),
    placeholderData: keepPreviousData,
  })

  const { code: baseCurrency } = useBaseCurrency()

  if (balanceQ.isPending) return (
    <div className="flex justify-center items-center pt-24"><Spinner /></div>
  )
  if (balanceQ.isError) return <ErrorMessage onRetry={() => balanceQ.refetch()} />

  const balance = balanceQ.data
  const isMultiCurrency = (balance?.by_currency?.length ?? 0) > 1
  const totalInBase = balance?.total_in_base_cents ?? 0
  const baseEntry = balance?.by_currency?.find(b => b.currency_code === baseCurrency)
    ?? balance?.by_currency?.[0]

  const selectedAccount = accounts.find(a => a.id === selectedAccountId)
  const heroCents = selectedAccount ? selectedAccount.balance_cents : totalInBase
  const heroCurrency = selectedAccount?.currency_code ?? baseCurrency

  const incomeCents = baseEntry?.income_cents ?? 0
  const expenseCents = baseEntry?.expense_cents ?? 0

  return (
    <PageTransition>
      <div className="flex flex-col min-h-[calc(100dvh-var(--tab-bar-h)-var(--safe-top,0px))]">

        {/* Fixed top section */}
        <div className="px-4 pt-4 flex flex-col gap-3 shrink-0">

          {/* Hero card */}
          <div className="hero-gradient p-6 relative"
               style={{ boxShadow: 'var(--shadow-hero)' }}>
            <div className="absolute -top-4 -right-4 w-40 h-40 rounded-full bg-white/[0.07] blur-xl pointer-events-none" />
            <div className="absolute -bottom-4 -left-4 w-32 h-32 rounded-full bg-indigo-400/20 blur-2xl pointer-events-none" />
            <div className="absolute top-1/2 right-1/3 w-20 h-20 rounded-full bg-white/[0.04] pointer-events-none" />

            <div className="relative z-10">
              {/* Top row: label + add button */}
              <div className="flex items-center justify-between mb-1">
                <p className="text-white/60 text-[11px] font-bold uppercase tracking-[0.2em]">
                  {t('dashboard.net_balance')}
                </p>
                <Link
                  to="/add"
                  className="w-9 h-9 rounded-2xl bg-white/15 backdrop-blur-sm flex items-center justify-center active:scale-90 transition-transform border border-white/10"
                >
                  <Plus size={18} weight="bold" className="text-white" />
                </Link>
              </div>

              <div className="flex items-baseline gap-2 mt-1">
                {!selectedAccount && isMultiCurrency && (
                  <span className="text-white/50 text-2xl font-bold">≈</span>
                )}
                <AnimatedMoney
                  cents={heroCents}
                  currency={heroCurrency}
                  className={`text-white font-extrabold tabular-nums leading-none tracking-tight ${
                    formatCents(heroCents, heroCurrency).length > 14 ? 'text-[22px]' :
                    formatCents(heroCents, heroCurrency).length > 11 ? 'text-[28px]' :
                    formatCents(heroCents, heroCurrency).length > 8  ? 'text-[34px]' :
                    'text-[42px]'
                  }`}
                />
              </div>

              {!selectedAccount && isMultiCurrency ? (
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
              {!selectedAccount && isMultiCurrency && showCurrencyBreakdown && (
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

              {/* Account selector — bottom of card */}
              {accounts.length > 0 && (
                <div className="mt-4 pt-3 border-t border-white/10 flex justify-end">
                  <AccountDropdown
                    accounts={accounts}
                    selectedId={selectedAccountId}
                    onChange={id => { if (id !== null) { setSelectedAccountId(id); setShowCurrencyBreakdown(false) } }}
                    showBalance
                  />
                </div>
              )}
            </div>
          </div>

          {/* Income / Expense bento cards */}
          {baseEntry && (
            <div className="grid grid-cols-2 gap-3">
              <motion.button
                layout
                onClick={() => navigate('/stats', { state: { type: 'income' } })}
                className="card-elevated p-4 relative text-left active:scale-[0.97] transition-transform overflow-hidden"
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
                  <AnimatedMoney cents={incomeCents} currency={heroCurrency} />
                </p>
              </motion.button>

              <motion.button
                layout
                onClick={() => navigate('/stats', { state: { type: 'expense' } })}
                className="card-elevated p-4 relative text-left active:scale-[0.97] transition-transform overflow-hidden"
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
                  <AnimatedMoney cents={expenseCents} currency={heroCurrency} />
                </p>
              </motion.button>
            </div>
          )}
        </div>

        {/* Recent transactions — fade/slide when account changes */}
        <div className="px-4 pt-3 pb-2 flex flex-col">
          <div className="card-elevated flex flex-col">
            <div className="flex justify-between items-center px-5 pt-4 pb-2 shrink-0">
              <span className="text-sm font-bold text-text">
                {t('dashboard.recent_transactions')}
              </span>
              <Link to="/history" className="flex items-center gap-1 text-xs text-accent font-semibold">
                {t('dashboard.view_all')} <ArrowRight size={14} weight="bold" />
              </Link>
            </div>

            <AnimatePresence mode="wait">
              {txQ.isPending ? (
                <motion.div
                  key="loading"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.15 }}
                  className="flex justify-center py-10"
                >
                  <Spinner size="sm" />
                </motion.div>
              ) : txQ.data?.transactions.length ? (
                <motion.div
                  key={`txlist-${selectedAccountId}`}
                  initial={{ opacity: 0, y: 8 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -8 }}
                  transition={{ duration: 0.22, ease: 'easeOut' }}
                  className="overflow-y-auto no-scrollbar"
                >
                  <div className="divide-y divide-border">
                    {txQ.data.transactions.map((tx, i) => (
                      <motion.div
                        key={tx.id}
                        initial={{ opacity: 0, x: -12 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ duration: 0.2, delay: i * 0.04, ease: 'easeOut' }}
                      >
                        <TransactionRow tx={tx} onEdit={setEditingTx} />
                      </motion.div>
                    ))}
                  </div>
                </motion.div>
              ) : (
                <motion.div
                  key={`empty-${selectedAccountId}`}
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.2 }}
                >
                  <EmptyState
                    icon={Receipt}
                    title={t('transactions.no_transactions')}
                    description={t('transactions.start_tracking')}
                  />
                </motion.div>
              )}
            </AnimatePresence>
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
