import { useState, useCallback, useMemo, useRef, useEffect } from 'react'
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { motion, AnimatePresence, useMotionValue, useTransform, animate } from 'framer-motion'
import { MagnifyingGlass, Trash, X, Receipt } from '@phosphor-icons/react'
import { transactionsApi } from '../api/transactions'
import { formatCents, formatDate } from '../lib/money'
import { CategoryIcon } from '../lib/categoryIcons'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { AmountDisplay, EmptyState, EditTransactionSheet } from '../components/ui'
import { useCategoryName } from '../hooks/useCategoryName'
import { useBaseCurrency } from '../hooks/useBaseCurrency'
import type { Transaction, TransactionType } from '../types'

const PAGE_SIZE = 20

/* ─── Swipeable Transaction Row ─── */
function SwipeableRow({
  tx,
  onDelete,
  onEdit,
  isDeleting,
}: {
  tx: Transaction
  onDelete: (id: number) => void
  onEdit: (tx: Transaction) => void
  isDeleting: boolean
}) {
  const x = useMotionValue(0)
  const deleteOpacity = useTransform(x, [-100, -60], [1, 0])
  const deleteScale = useTransform(x, [-100, -60], [1, 0.8])
  const [swiped, setSwiped] = useState(false)
  const isIncome = tx.type === 'income'
  const tCategory = useCategoryName()

  function handleDragEnd() {
    const val = x.get()
    if (val < -60) {
      setSwiped(true)
      animate(x, -80, { type: 'spring', stiffness: 300, damping: 30 })
    } else {
      setSwiped(false)
      animate(x, 0, { type: 'spring', stiffness: 300, damping: 30 })
    }
  }

  function handleClose() {
    setSwiped(false)
    animate(x, 0, { type: 'spring', stiffness: 300, damping: 30 })
  }

  return (
    <div className={`relative overflow-hidden transition-opacity ${isDeleting ? 'opacity-30 pointer-events-none' : ''}`}>
      {/* Delete action behind */}
      <motion.div
        className="absolute inset-y-0 right-0 w-20 flex items-center justify-center"
        style={{ opacity: deleteOpacity, scale: deleteScale }}
      >
        <button
          onClick={() => onDelete(tx.id)}
          className="w-11 h-11 rounded-2xl bg-destructive/10 flex items-center justify-center text-destructive active:scale-90 transition-transform"
        >
          <Trash size={18} weight="bold" />
        </button>
      </motion.div>

      {/* Row content — draggable */}
      <motion.div
        className="relative bg-surface flex items-center gap-3 px-4 py-3"
        style={{ x }}
        drag="x"
        dragConstraints={{ left: -80, right: 0 }}
        dragElastic={0.1}
        onDragEnd={handleDragEnd}
        onClick={() => { if (swiped) { handleClose() } else { onEdit(tx) } }}
      >
        <div
          className="w-10 h-10 rounded-2xl flex items-center justify-center shrink-0"
          style={{ background: tx.category_color || 'var(--color-accent)' }}
        >
          <CategoryIcon emoji={tx.category_emoji} size={20} weight="fill" className="text-white" />
        </div>

        <div className="flex-1 min-w-0">
          <span className="text-[13px] font-semibold text-text truncate block">{tCategory(tx.category_name)}</span>
          {tx.note && (
            <p className="text-[11px] text-muted mt-0.5 truncate">{tx.note}</p>
          )}
        </div>
        <AmountDisplay
          cents={tx.amount_cents}
          currency={tx.currency_code}
          type={isIncome ? 'income' : 'expense'}
          size="sm"
          showSign
        />
      </motion.div>
    </div>
  )
}

/* ─── Date Group ─── */
function DateGroup({
  label,
  transactions,
  onDelete,
  onEdit,
  deletingId,
  baseCurrency,
}: {
  label: string
  transactions: Transaction[]
  onDelete: (id: number) => void
  onEdit: (tx: Transaction) => void
  deletingId: number | null
  baseCurrency: string
}) {
  const total = transactions.reduce((sum, tx) => {
    return sum + (tx.type === 'income' ? tx.amount_cents : -tx.amount_cents)
  }, 0)

  return (
    <div>
      {/* Date header with day total */}
      <div className="flex items-center justify-between px-5 mb-1.5">
        <span className="text-[11px] font-bold text-muted uppercase tracking-wider">{label}</span>
        <span className={`text-[11px] font-bold tabular-nums ${total >= 0 ? 'text-income' : 'text-expense'}`}>
          {total >= 0 ? '+' : '−'}{formatCents(Math.abs(total), baseCurrency)}
        </span>
      </div>
      {/* Transaction cards */}
      <div className="mx-4 card-elevated divide-y divide-border">
        {transactions.map((tx) => (
          <SwipeableRow
            key={tx.id}
            tx={tx}
            onDelete={onDelete}
            onEdit={onEdit}
            isDeleting={deletingId === tx.id}
          />
        ))}
      </div>
    </div>
  )
}

/* ─── Group by date helper ─── */
function groupByDate(transactions: Transaction[], t: (key: string) => string, lang: string): [string, Transaction[]][] {
  const groups = new Map<string, Transaction[]>()
  const now = new Date()
  const today = now.toDateString()
  const yesterday = new Date(now.getTime() - 86400000).toDateString()

  for (const tx of transactions) {
    const txDate = new Date(tx.created_at).toDateString()
    let label: string
    if (txDate === today) {
      label = t('history.today')
    } else if (txDate === yesterday) {
      label = t('history.yesterday')
    } else {
      label = formatDate(tx.created_at, lang, { month: 'long', day: 'numeric', year: 'numeric' })
    }
    const existing = groups.get(label)
    if (existing) {
      existing.push(tx)
    } else {
      groups.set(label, [tx])
    }
  }
  return [...groups.entries()]
}

/* ─── Filter options ─── */
const FILTER_OPTIONS = [
  { value: 'all', labelKey: 'history.all_types' },
  { value: 'expense', labelKey: 'history.expenses' },
  { value: 'income', labelKey: 'history.incomes' },
]

/* ─── Main Page ─── */
export function HistoryPage() {
  const { t, i18n } = useTranslation()
  const { code: baseCurrency } = useBaseCurrency()
  const qc = useQueryClient()
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const [editingTx, setEditingTx] = useState<Transaction | null>(null)
  const [typeFilter, setTypeFilter] = useState<'all' | TransactionType>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [showSearch, setShowSearch] = useState(false)
  const searchRef = useRef<HTMLInputElement>(null)
  const listRef = useRef<HTMLDivElement>(null)
  const sentinelRef = useRef<HTMLDivElement>(null)

  const {
    data,
    isLoading,
    isError,
    refetch,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery({
    queryKey: ['transactions'],
    queryFn: ({ pageParam }) => transactionsApi.list(pageParam, PAGE_SIZE),
    initialPageParam: 1,
    getNextPageParam: (lastPage) =>
      lastPage.current_page < lastPage.total_pages
        ? lastPage.current_page + 1
        : undefined,
  })

  // Intersection Observer for infinite scroll
  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel) return

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasNextPage && !isFetchingNextPage) {
          fetchNextPage()
        }
      },
      { rootMargin: '200px' }
    )

    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [hasNextPage, isFetchingNextPage, fetchNextPage])

  const deleteMutation = useMutation({
    mutationFn: (id: number) => transactionsApi.delete(id),
    onMutate: (id) => setDeletingId(id),
    onSettled: () => setDeletingId(null),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['transactions'] })
      qc.invalidateQueries({ queryKey: ['balance'] })
    },
  })

  const handleDelete = useCallback((id: number) => { deleteMutation.mutate(id) }, [deleteMutation])
  const handleEdit = useCallback((tx: Transaction) => { setEditingTx(tx) }, [])

  const allItems = useMemo(
    () => data?.pages.flatMap((p) => p.transactions) ?? [],
    [data?.pages]
  )

  // Client-side filtering
  const filtered = useMemo(() => {
    let result = allItems
    if (typeFilter !== 'all') {
      result = result.filter(tx => tx.type === typeFilter)
    }
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase()
      result = result.filter(tx =>
        tx.note?.toLowerCase().includes(q) ||
        tx.category_name.toLowerCase().includes(q)
      )
    }
    return result
  }, [allItems, typeFilter, searchQuery])

  const grouped = useMemo(() => groupByDate(filtered, t, i18n.language), [filtered, t, i18n.language])

  function toggleSearch() {
    if (showSearch) {
      setSearchQuery('')
    }
    setShowSearch(!showSearch)
  }

  if (isLoading) return <div className="flex justify-center py-16"><Spinner /></div>
  if (isError) return <ErrorMessage onRetry={refetch} />

  return (
    <PageTransition>
      <div className="flex flex-col min-h-[calc(100dvh-var(--tab-bar-h)-var(--safe-top,0px))]">

        {/* Toolbar — filter pills + search */}
        <div className="shrink-0 px-4 pt-3 pb-3 space-y-2.5">
          <div className="flex items-center gap-2">
            <div className="flex gap-2 flex-1 py-1 -my-1">
              {FILTER_OPTIONS.map((opt) => {
                const isActive = opt.value === typeFilter
                return (
                  <button
                    key={opt.value}
                    onClick={() => setTypeFilter(opt.value as 'all' | TransactionType)}
                    className={`
                      shrink-0 px-4 py-2 rounded-full text-xs font-bold transition-all duration-200 select-none
                      ${isActive
                        ? 'bg-accent text-accent-text shadow-[0_2px_12px_rgba(99,102,241,0.4),0_1px_3px_rgba(0,0,0,0.08)]'
                        : 'bg-surface text-muted shadow-sm active:scale-95'
                      }
                    `}
                  >
                    {t(opt.labelKey)}
                  </button>
                )
              })}
            </div>
            <button
              onClick={toggleSearch}
              className={`w-9 h-9 flex items-center justify-center rounded-full transition-all duration-200 shrink-0 ${
                showSearch
                  ? 'bg-accent text-accent-text shadow-[0_2px_12px_rgba(99,102,241,0.4),0_1px_3px_rgba(0,0,0,0.08)]'
                  : 'bg-surface text-muted shadow-sm active:scale-95'
              }`}
            >
              {showSearch ? <X size={14} weight="bold" /> : <MagnifyingGlass size={14} weight="bold" />}
            </button>
          </div>

          {/* Search bar */}
          <AnimatePresence>
            {showSearch && (
              <motion.div
                initial={{ opacity: 0, y: -8 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -8 }}
                transition={{ duration: 0.15 }}
              >
                <div className="relative">
                  <MagnifyingGlass size={14} weight="bold" className="absolute left-3.5 top-1/2 -translate-y-1/2 text-muted" />
                  <input
                    ref={searchRef}
                    type="text"
                    autoFocus
                    value={searchQuery}
                    onChange={e => setSearchQuery(e.target.value)}
                    placeholder={t('history.search_placeholder')}
                    className="w-full bg-surface rounded-full pl-9 pr-4 py-2.5 text-xs font-medium outline-none text-text placeholder:text-muted/50 transition-shadow shadow-[0_2px_16px_rgba(0,0,0,0.06),0_1px_4px_rgba(0,0,0,0.04)] focus:shadow-[0_2px_16px_rgba(0,0,0,0.06),0_1px_4px_rgba(0,0,0,0.04),0_0_0_2px_rgba(99,102,241,0.2)]"
                  />
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>

        {/* Transaction list — scrollable */}
        <div ref={listRef} className="flex-1 min-h-0 overflow-y-auto no-scrollbar pb-3">
          {filtered.length === 0 ? (
            <div className="mx-4 card-elevated mt-2">
              <EmptyState
                icon={Receipt}
                title={t('transactions.no_transactions')}
                description={searchQuery ? undefined : t('transactions.start_tracking')}
              />
            </div>
          ) : (
            <div className="flex flex-col gap-3">
              {grouped.map(([date, txs]) => (
                <DateGroup
                  key={date}
                  label={date}
                  transactions={txs}
                  onDelete={handleDelete}
                  onEdit={handleEdit}
                  deletingId={deletingId}
                  baseCurrency={baseCurrency}
                />
              ))}
            </div>
          )}

          {/* Infinite scroll sentinel */}
          <div ref={sentinelRef} className="h-1" />
          {isFetchingNextPage && (
            <div className="flex justify-center py-4">
              <Spinner size="sm" />
            </div>
          )}
        </div>
      </div>

      {/* Edit sheet */}
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
