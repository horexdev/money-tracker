import { useState, useCallback, useMemo, useRef, useEffect } from 'react'
import { useInfiniteQuery, useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { motion, AnimatePresence } from 'framer-motion'
import {
  MagnifyingGlass, X, Receipt, CalendarBlank,
  SortAscending, Funnel, Check,
} from '@phosphor-icons/react'
import { transactionsApi } from '../../shared/api/transactions'
import { accountsApi } from '../../shared/api/accounts'
import { formatCents, formatDate } from '../../shared/lib/money'
import { CategoryIcon } from '../../shared/lib/categoryIcons'
import { Spinner } from '../../shared/ui/Spinner'
import { ErrorMessage } from '../../shared/ui/ErrorMessage'
import { PageTransition } from '../../shared/ui/PageTransition'
import { AmountDisplay, EmptyState, EditTransactionSheet, ActionRow, AccountDropdown, BottomSheet, RangeDateModal, fmtDisplay } from '../../shared/ui'
import { useCategoryName } from '../../shared/hooks/useCategoryName'
import { useBaseCurrency } from '../../shared/hooks/useBaseCurrency'
import type { Transaction, TransactionType } from '../../shared/types'

const PAGE_SIZE = 20

type SortOrder = 'newest' | 'oldest' | 'amount_desc' | 'amount_asc'

/* ─── Transaction Row ─── */
function TransactionRow({
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
  const isIncome = tx.type === 'income'
  const tCategory = useCategoryName()

  return (
    <ActionRow
      onDelete={() => onDelete(tx.id)}
      onEdit={() => onEdit(tx)}
      isDeleting={isDeleting}
    >
      <button
        onClick={() => onEdit(tx)}
        className="flex items-center gap-3 px-4 py-3 w-full text-left active:bg-border/50 transition-colors"
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
      </button>
    </ActionRow>
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
      <div className="flex items-center justify-between px-5 mb-1.5">
        <span className="text-[11px] font-bold text-muted uppercase tracking-wider">{label}</span>
        <span className={`text-[11px] font-bold tabular-nums ${total >= 0 ? 'text-income' : 'text-expense'}`}>
          {total >= 0 ? '+' : '−'}{formatCents(Math.abs(total), baseCurrency)}
        </span>
      </div>
      <div className="mx-4 card-elevated divide-y divide-border">
        {transactions.map((tx) => (
          <TransactionRow
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
const TYPE_FILTER_OPTIONS = [
  { value: 'all', labelKey: 'history.all_types' },
  { value: 'expense', labelKey: 'history.expenses' },
  { value: 'income', labelKey: 'history.incomes' },
]

const SORT_OPTIONS: { value: SortOrder; labelKey: string }[] = [
  { value: 'newest',      labelKey: 'history.sort_newest' },
  { value: 'oldest',      labelKey: 'history.sort_oldest' },
  { value: 'amount_desc', labelKey: 'history.sort_amount_high' },
  { value: 'amount_asc',  labelKey: 'history.sort_amount_low' },
]

/* ─── Main Page ─── */
export function HistoryPage() {
  const { t, i18n } = useTranslation()
  const { code: baseCurrency } = useBaseCurrency()
  const tCategory = useCategoryName()
  const qc = useQueryClient()
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const [editingTx, setEditingTx] = useState<Transaction | null>(null)
  const [typeFilter, setTypeFilter] = useState<'all' | TransactionType>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [showSearch, setShowSearch] = useState(false)
  const [accountFilter, setAccountFilter] = useState<number | null>(null)
  const [dateFrom, setDateFrom] = useState<string | null>(null)
  const [dateTo, setDateTo] = useState<string | null>(null)
  const [sortOrder, setSortOrder] = useState<SortOrder>('newest')
  const [categoryFilter, setCategoryFilter] = useState<number | null>(null)
  const [showDateSheet, setShowDateSheet] = useState(false)
  const [showSortSheet, setShowSortSheet] = useState(false)
  const [showCategorySheet, setShowCategorySheet] = useState(false)
  const searchRef = useRef<HTMLInputElement>(null)
  const listRef = useRef<HTMLDivElement>(null)
  const sentinelRef = useRef<HTMLDivElement>(null)

  const { data: accounts = [] } = useQuery({
    queryKey: ['accounts'],
    queryFn: accountsApi.list,
  })

  const {
    data,
    isLoading,
    isError,
    refetch,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery({
    queryKey: ['transactions', accountFilter, dateFrom, dateTo],
    queryFn: ({ pageParam }) =>
      transactionsApi.list(pageParam, PAGE_SIZE, {
        accountId: accountFilter,
        from: dateFrom,
        to: dateTo,
      }),
    initialPageParam: 1,
    getNextPageParam: (lastPage) =>
      lastPage.current_page < lastPage.total_pages
        ? lastPage.current_page + 1
        : undefined,
  })

  // Scroll to top when backend filters change
  useEffect(() => {
    listRef.current?.scrollTo(0, 0)
  }, [accountFilter, dateFrom, dateTo])

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
    () => (data?.pages.flatMap((p) => p.transactions) ?? []).filter(tx => !tx.is_adjustment),
    [data?.pages]
  )

  // Collect unique categories from loaded transactions (no adjustment)
  const uniqueCategories = useMemo(() => {
    const seen = new Map<number, { id: number; name: string; emoji: string; color: string }>()
    for (const tx of allItems) {
      if (!seen.has(tx.category_id)) {
        seen.set(tx.category_id, {
          id: tx.category_id,
          name: tx.category_name,
          emoji: tx.category_emoji,
          color: tx.category_color,
        })
      }
    }
    return [...seen.values()]
  }, [allItems])

  // Client-side filtering
  const filtered = useMemo(() => {
    let result = allItems
    if (typeFilter !== 'all') {
      result = result.filter(tx => tx.type === typeFilter)
    }
    if (categoryFilter !== null) {
      result = result.filter(tx => tx.category_id === categoryFilter)
    }
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase()
      result = result.filter(tx =>
        tx.note?.toLowerCase().includes(q) ||
        tx.category_name.toLowerCase().includes(q) ||
        tx.account_name?.toLowerCase().includes(q)
      )
    }
    // Sort
    result = [...result].sort((a, b) => {
      switch (sortOrder) {
        case 'oldest':      return new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
        case 'amount_desc': return b.amount_cents - a.amount_cents
        case 'amount_asc':  return a.amount_cents - b.amount_cents
        default:            return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      }
    })
    return result
  }, [allItems, typeFilter, categoryFilter, searchQuery, sortOrder])

  const grouped = useMemo(() => groupByDate(filtered, t, i18n.language), [filtered, t, i18n.language])

  // Active filter count (badge)
  const activeFilterCount = [
    accountFilter !== null,
    dateFrom !== null || dateTo !== null,
    sortOrder !== 'newest',
    categoryFilter !== null,
  ].filter(Boolean).length

  function toggleSearch() {
    if (showSearch) setSearchQuery('')
    setShowSearch(!showSearch)
  }

  function clearAllFilters() {
    setAccountFilter(null)
    setDateFrom(null)
    setDateTo(null)
    setSortOrder('newest')
    setCategoryFilter(null)
    setSearchQuery('')
  }

  if (isLoading) return <div className="flex justify-center py-16"><Spinner /></div>
  if (isError) return <ErrorMessage onRetry={refetch} />

  return (
    <PageTransition>
      <div className="flex flex-col min-h-[calc(100dvh-var(--tab-bar-h)-var(--safe-top,0px))]">

        {/* Toolbar */}
        <div className="shrink-0 px-4 pt-3 pb-3 space-y-2.5">

          {/* Row 1: type pills + search toggle */}
          <div className="flex items-center gap-2">
            <div className="flex gap-2 flex-1 py-1 -my-1">
              {TYPE_FILTER_OPTIONS.map((opt) => {
                const isActive = opt.value === typeFilter
                return (
                  <button
                    key={opt.value}
                    onClick={() => setTypeFilter(opt.value as 'all' | TransactionType)}
                    className={`
                      shrink-0 px-4 py-2 rounded-full text-xs font-bold transition-all duration-200 select-none
                      ${isActive
                        ? 'bg-accent text-accent-text shadow-(--shadow-accent-pill)'
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
                  ? 'bg-accent text-accent-text shadow-(--shadow-accent-pill)'
                  : 'bg-surface text-muted shadow-sm active:scale-95'
              }`}
            >
              {showSearch ? <X size={14} weight="bold" /> : <MagnifyingGlass size={14} weight="bold" />}
            </button>
          </div>

          {/* Row 2: filter/sort buttons */}
          <div className="flex items-center gap-2">
            {/* Account dropdown */}
            <AccountDropdown
              accounts={accounts}
              selectedId={accountFilter}
              onChange={setAccountFilter}
              allLabel={t('accountsAll')}
              variant="surface"
            />

            <div className="flex gap-1.5 ml-auto">
              {/* Date range */}
              <button
                onClick={() => setShowDateSheet(true)}
                className={`flex items-center gap-1 px-3 py-1.5 rounded-full text-xs font-bold transition-all duration-200 ${
                  (dateFrom || dateTo)
                    ? 'bg-accent text-accent-text shadow-(--shadow-accent-pill)'
                    : 'bg-surface text-muted shadow-sm active:scale-95'
                }`}
              >
                <CalendarBlank size={13} weight="bold" />
                {(dateFrom || dateTo) ? (
                  <span>{dateFrom ? fmtDisplay(dateFrom) : '…'} – {dateTo ? fmtDisplay(dateTo) : '…'}</span>
                ) : (
                  <span>{t('history.date_filter')}</span>
                )}
              </button>

              {/* Sort */}
              <button
                onClick={() => setShowSortSheet(true)}
                className={`w-8 h-8 flex items-center justify-center rounded-full transition-all duration-200 ${
                  sortOrder !== 'newest'
                    ? 'bg-accent text-accent-text shadow-(--shadow-accent-pill)'
                    : 'bg-surface text-muted shadow-sm active:scale-95'
                }`}
              >
                <SortAscending size={14} weight="bold" />
              </button>

              {/* Category filter */}
              <button
                onClick={() => setShowCategorySheet(true)}
                className={`w-8 h-8 flex items-center justify-center rounded-full transition-all duration-200 ${
                  categoryFilter !== null
                    ? 'bg-accent text-accent-text shadow-(--shadow-accent-pill)'
                    : 'bg-surface text-muted shadow-sm active:scale-95'
                }`}
              >
                <Funnel size={14} weight="bold" />
              </button>
            </div>
          </div>

          {/* Active filters bar */}
          {activeFilterCount > 0 && (
            <button
              onClick={clearAllFilters}
              className="flex items-center gap-1.5 text-xs font-semibold text-accent"
            >
              <X size={12} weight="bold" />
              {t('history.clear_filters')} ({activeFilterCount})
            </button>
          )}

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
                    className="w-full bg-surface rounded-full pl-9 pr-4 py-2.5 text-xs font-medium outline-none text-text placeholder:text-muted/50 transition-shadow shadow-(--shadow-card) focus:shadow-(--shadow-focus)"
                  />
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>

        {/* Transaction list */}
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

      {/* Date range sheet */}
      <AnimatePresence>
        {showDateSheet && (
          <RangeDateModal
            initialFrom={dateFrom ?? new Date().toISOString().split('T')[0]}
            initialTo={dateTo ?? new Date().toISOString().split('T')[0]}
            onApply={(from, to) => { setDateFrom(from); setDateTo(to) }}
            onClose={() => setShowDateSheet(false)}
            labelFrom={t('stats.date_from')}
            labelTo={t('stats.date_to')}
            applyLabel={t('stats.apply')}
          />
        )}
      </AnimatePresence>

      {/* Sort sheet */}
      <AnimatePresence>
        {showSortSheet && (
          <BottomSheet onClose={() => setShowSortSheet(false)}>
            <div className="px-5 pt-3 pb-1">
              <p className="text-base font-bold text-text mb-2">{t('history.sort')}</p>
            </div>
            <div className="pb-6 divide-y divide-border">
              {SORT_OPTIONS.map(({ value, labelKey }) => (
                <button
                  key={value}
                  onClick={() => { setSortOrder(value); setShowSortSheet(false) }}
                  className={`w-full flex items-center justify-between px-5 py-3.5 transition-colors ${
                    sortOrder === value ? 'bg-accent-subtle' : 'active:bg-border'
                  }`}
                >
                  <span className={`text-[13px] font-semibold ${sortOrder === value ? 'text-accent' : 'text-text'}`}>
                    {t(labelKey)}
                  </span>
                  {sortOrder === value && <Check size={16} weight="bold" className="text-accent" />}
                </button>
              ))}
            </div>
          </BottomSheet>
        )}
      </AnimatePresence>

      {/* Category filter sheet */}
      <AnimatePresence>
        {showCategorySheet && (
          <BottomSheet onClose={() => setShowCategorySheet(false)}>
            <div className="px-5 pt-3 pb-1">
              <p className="text-base font-bold text-text mb-3">{t('history.category_filter')}</p>
            </div>
            <div className="px-4 pb-6">
              <div className="flex flex-wrap gap-2">
                <button
                  onClick={() => { setCategoryFilter(null); setShowCategorySheet(false) }}
                  className={`flex items-center gap-1.5 px-3 py-2 rounded-full text-xs font-bold transition-all ${
                    categoryFilter === null
                      ? 'bg-accent text-accent-text shadow-(--shadow-accent-pill)'
                      : 'bg-surface text-muted shadow-sm'
                  }`}
                >
                  {t('history.all_types')}
                </button>
                {uniqueCategories.map(cat => {
                  const isActive = categoryFilter === cat.id
                  return (
                    <button
                      key={cat.id}
                      onClick={() => { setCategoryFilter(cat.id); setShowCategorySheet(false) }}
                      className={`flex items-center gap-1.5 px-3 py-2 rounded-full text-xs font-bold transition-all ${
                        isActive
                          ? 'bg-accent text-accent-text shadow-(--shadow-accent-pill)'
                          : 'bg-surface text-muted shadow-sm'
                      }`}
                    >
                      <span
                        className="w-4 h-4 rounded-full flex items-center justify-center shrink-0"
                        style={{ background: cat.color }}
                      >
                        <CategoryIcon emoji={cat.emoji} size={10} weight="fill" className={isActive ? 'text-accent-text' : 'text-white'} />
                      </span>
                      <span>{tCategory(cat.name)}</span>
                    </button>
                  )
                })}
              </div>
            </div>
          </BottomSheet>
        )}
      </AnimatePresence>
    </PageTransition>
  )
}
