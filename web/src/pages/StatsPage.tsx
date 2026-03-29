import { useState, useMemo, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { motion, AnimatePresence, useSpring, useMotionValueEvent } from 'framer-motion'
import { CalendarDots, ChartBar } from '@phosphor-icons/react'
import { statsApi } from '../api/stats'
import { formatCents } from '../lib/money'
import { CategoryIcon } from '../lib/categoryIcons'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { EmptyState, RangeDateModal } from '../components/ui'
import { useCategoryName } from '../hooks/useCategoryName'
import { useBaseCurrency } from '../hooks/useBaseCurrency'
import type { TransactionType, CategoryStat } from '../types'

const COLORS = [
  '#6366f1', '#22c55e', '#06b6d4', '#f59e0b',
  '#ef4444', '#8b5cf6', '#ec4899', '#f97316',
]

type Period = 'month' | 'week' | 'today' | 'lastmonth'

interface CustomRange {
  from: string
  to: string
}

/* ─── Animated Number Counter ─── */
function AnimatedNumber({ value, formatter }: { value: number; formatter?: (v: number) => string }) {
  const spring = useSpring(0, { stiffness: 80, damping: 20 })
  const [display, setDisplay] = useState(() =>
    formatter ? formatter(value) : value.toString()
  )

  useEffect(() => { spring.set(value) }, [value, spring])

  useMotionValueEvent(spring, 'change', (v) => {
    setDisplay(formatter ? formatter(Math.round(v)) : Math.round(v).toString())
  })

  return <motion.span>{display}</motion.span>
}

/* ─── Donut Chart (stroke-based, animated) ─── */
function DonutChart({
  data,
  total,
  animationKey,
  currency,
}: {
  data: Array<{ percent: number; category_name: string; category_color: string }>
  total: number
  animationKey: string
  currency: string
}) {
  const size = 130
  const cx = size / 2
  const cy = size / 2
  const r = 50
  const strokeWidth = 18
  const circumference = 2 * Math.PI * r

  let accumulated = 0
  const segments = data.map((d, i) => {
    const segLen = (d.percent / 100) * circumference
    const offset = circumference - accumulated
    accumulated += segLen

    return (
      <motion.circle
        key={`${animationKey}-${i}`}
        cx={cx}
        cy={cy}
        r={r}
        fill="none"
        stroke={d.category_color || COLORS[i % COLORS.length]}
        strokeWidth={strokeWidth}
        strokeLinecap="round"
        strokeDasharray={`${segLen - 2} ${circumference - segLen + 2}`}
        strokeDashoffset={offset}
        initial={{ opacity: 0, strokeDasharray: `0 ${circumference}` }}
        animate={{ opacity: 1, strokeDasharray: `${segLen - 2} ${circumference - segLen + 2}` }}
        transition={{ duration: 0.6, delay: i * 0.08, ease: 'easeOut' }}
      />
    )
  })

  return (
    <div className="relative" style={{ width: size, height: size }}>
      <svg
        width={size}
        height={size}
        viewBox={`0 0 ${size} ${size}`}
        style={{ transform: 'rotate(-90deg)' }}
      >
        {/* Background ring */}
        <circle
          cx={cx}
          cy={cy}
          r={r}
          fill="none"
          stroke="var(--color-border)"
          strokeWidth={strokeWidth}
        />
        {segments}
      </svg>
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <div className="text-center px-1">
          <p className="text-[9px] font-semibold text-muted uppercase tracking-wide">{currency}</p>
          <p className="text-[11px] font-bold tabular-nums text-text leading-tight">
            <AnimatedNumber
              value={total}
              formatter={(v) =>
                new Intl.NumberFormat('en-US', {
                  style: 'decimal',
                  minimumFractionDigits: 0,
                  maximumFractionDigits: 0,
                  notation: total >= 100000_00 ? 'compact' : 'standard',
                }).format(v / 100)
              }
            />
          </p>
        </div>
      </div>
    </div>
  )
}

/* ─── Category Row ─── */
function CategoryRow({
  entry,
  index,
  color,
  currency,
}: {
  entry: CategoryStat & { percent: number }
  index: number
  color: string
  currency: string
}) {
  const tCategory = useCategoryName()
  return (
    <motion.div
      className="flex items-center gap-3 px-4 py-3 border-b border-border last:border-b-0"
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3, delay: index * 0.05 }}
    >
      <div
        className="w-2.5 h-2.5 rounded-full shrink-0"
        style={{ background: color }}
      />
      <div
        className="w-9 h-9 rounded-2xl flex items-center justify-center shrink-0"
        style={{ background: entry.category_color || color }}
      >
        <CategoryIcon emoji={entry.category_emoji} size={18} weight="fill" className="text-white" />
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex justify-between items-center">
          <span className="text-sm font-semibold text-text truncate">{tCategory(entry.category_name)}</span>
          <span className="text-sm font-bold tabular-nums text-text ml-2 shrink-0">
            {formatCents(entry.total_cents, currency)}
          </span>
        </div>
        <div className="mt-1.5 h-1.5 rounded-full overflow-hidden bg-bg">
          <motion.div
            className="h-full rounded-full"
            style={{ background: color }}
            initial={{ width: 0 }}
            animate={{ width: `${entry.percent}%` }}
            transition={{ duration: 0.5, delay: 0.15 + index * 0.05, ease: 'easeOut' }}
          />
        </div>
      </div>
      <span className="text-xs font-bold text-muted shrink-0 w-9 text-right">
        {entry.percent.toFixed(0)}%
      </span>
    </motion.div>
  )
}

/* ─── Main Page ─── */
export function StatsPage() {
  const { t } = useTranslation()
  const tCategory = useCategoryName()
  const { code: baseCurrency } = useBaseCurrency()
  const location = useLocation()
  const initialType = (location.state as { type?: TransactionType } | null)?.type ?? 'expense'
  const [type, setType] = useState<TransactionType>(initialType)
  const [period, setPeriod] = useState<Period>('month')
  const [customRange, setCustomRange] = useState<CustomRange | null>(null)
  const [showDatePicker, setShowDatePicker] = useState(false)

  const isCustom = customRange !== null

  const periodOptions: { value: Period | 'custom'; label: string; icon?: boolean }[] = [
    { value: 'today',     label: t('stats.today') },
    { value: 'week',      label: t('stats.week') },
    { value: 'month',     label: t('stats.month') },
    { value: 'lastmonth', label: t('stats.last_month') },
    { value: 'custom',    label: t('stats.custom'), icon: true },
  ]

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: customRange
      ? ['stats', 'custom', customRange.from, customRange.to]
      : ['stats', period],
    queryFn: customRange
      ? () => statsApi.getRange(customRange.from, customRange.to)
      : () => statsApi.get(period),
  })

  const filtered: CategoryStat[] = data?.items.filter((s) => s.type === type) ?? []
  const total = filtered.reduce((sum, s) => sum + s.total_cents, 0)
  const txCount = filtered.reduce((sum, s) => sum + s.tx_count, 0)
  const withPercent = useMemo(
    () =>
      filtered
        .map((s) => ({
          ...s,
          percent: total > 0 ? (s.total_cents / total) * 100 : 0,
        }))
        .sort((a, b) => b.total_cents - a.total_cents),
    [filtered, total],
  )

  const topCategory = withPercent[0] ?? null
  const avgPerTx = txCount > 0 ? Math.round(total / txCount) : 0
  const animationKey = `${type}-${period}-${customRange?.from ?? ''}-${customRange?.to ?? ''}`

  function handlePeriodChange(v: Period | 'custom') {
    if (v === 'custom') {
      setShowDatePicker(true)
    } else {
      setCustomRange(null)
      setPeriod(v)
    }
  }

  function handleCustomApply(from: string, to: string) {
    setCustomRange({ from, to })
  }

  return (
    <PageTransition>
      <div className="flex flex-col h-[calc(100dvh-var(--tab-bar-h))]">

        {/* Hero block with type toggle */}
        <div className="mx-4 mt-4 hero-gradient rounded-card px-5 pt-5 pb-4 relative overflow-hidden shrink-0"
             style={{ boxShadow: 'var(--shadow-hero)' }}>
          <div className="absolute -top-10 -right-10 w-36 h-36 rounded-full bg-white/[0.06] blur-xl pointer-events-none" />
          <div className="absolute -bottom-8 -left-8 w-28 h-28 rounded-full bg-indigo-400/15 blur-2xl pointer-events-none" />

          <div className="relative z-10">
            {/* Glass type toggle */}
            <div className="inline-flex bg-white/10 backdrop-blur-sm rounded-2xl p-1 gap-1 border border-white/[0.08]">
              {(['expense', 'income'] as TransactionType[]).map((v) => (
                <button
                  key={v}
                  onClick={() => setType(v)}
                  className={`
                    px-5 py-2 rounded-xl text-xs font-bold transition-all duration-200 select-none
                    ${type === v
                      ? 'bg-white/20 text-white shadow-[0_2px_8px_rgba(0,0,0,0.15)]'
                      : 'text-white/50'
                    }
                  `}
                >
                  {v === 'expense' ? t('transactions.expense') : t('transactions.income')}
                </button>
              ))}
            </div>

            {/* Animated total + count */}
            <div className="mt-3 flex items-end gap-3">
              <p className="text-white text-3xl font-extrabold tabular-nums leading-none tracking-tight">
                <AnimatedNumber value={total} formatter={(v) => formatCents(v, baseCurrency)} />
              </p>
              <p className="text-white/40 text-xs font-medium pb-0.5">
                <AnimatedNumber value={txCount} /> {t('stats.transactions_count_other', { count: txCount }).replace(/^\d+\s*/, '')}
              </p>
            </div>
          </div>
        </div>

        {/* Period pills */}
        <div className="shrink-0 px-4 pt-3 pb-2">
          <div className="flex gap-2 overflow-x-auto no-scrollbar -mx-4 px-4">
            {periodOptions.map((opt) => {
              const isActive = opt.value === 'custom' ? isCustom : (!isCustom && opt.value === period)
              return (
                <button
                  key={opt.value}
                  onClick={() => handlePeriodChange(opt.value)}
                  className={`
                    shrink-0 px-4 py-2 rounded-full text-xs font-bold transition-all duration-200 select-none
                    flex items-center gap-1.5
                    ${isActive
                      ? 'bg-accent text-accent-text shadow-[0_2px_8px_rgba(99,102,241,0.3)]'
                      : 'bg-surface text-muted shadow-sm active:scale-95'
                    }
                  `}
                >
                  {opt.icon && <CalendarDots size={14} weight="bold" />}
                  {opt.label}
                </button>
              )
            })}
          </div>
        </div>

        {/* Content area */}
        <div className="flex-1 min-h-0 px-4 pb-3 flex flex-col gap-3">
          {isLoading && <div className="flex justify-center py-12"><Spinner /></div>}
          {isError && <ErrorMessage onRetry={refetch} />}

          <AnimatePresence mode="wait">
            {data && (
              <motion.div
                key={animationKey}
                className="flex flex-col gap-3 flex-1 min-h-0"
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                transition={{ duration: 0.2 }}
              >
                {withPercent.length > 0 ? (
                  <>
                    {/* Chart card — 3-column layout */}
                    <div className="card-elevated overflow-hidden shrink-0">
                      <div className="flex items-center px-4 py-4">
                        {/* Left — top category */}
                        <div className="flex-1 min-w-0">
                          <p className="text-[10px] font-semibold text-muted uppercase tracking-wide">{t('stats.top_category')}</p>
                          {topCategory && (
                            <div className="mt-1.5">
                              <div
                                className="w-10 h-10 rounded-2xl flex items-center justify-center mb-1.5"
                                style={{ background: topCategory.category_color || COLORS[0] }}
                              >
                                <CategoryIcon emoji={topCategory.category_emoji} size={20} weight="fill" className="text-white" />
                              </div>
                              <p className="text-xs font-bold text-text truncate">{tCategory(topCategory.category_name)}</p>
                              <p className="text-lg font-extrabold tabular-nums" style={{ color: topCategory.category_color || COLORS[0] }}>
                                {topCategory.percent.toFixed(0)}%
                              </p>
                            </div>
                          )}
                        </div>

                        {/* Center — donut */}
                        <div className="shrink-0 mx-2">
                          <DonutChart data={withPercent} total={total} animationKey={animationKey} currency={baseCurrency} />
                        </div>

                        {/* Right — stats */}
                        <div className="flex-1 min-w-0 text-right">
                          <div className="mb-3">
                            <p className="text-[10px] font-semibold text-muted uppercase tracking-wide">
                              {t('stats.categories_count', { count: withPercent.length }).replace(/^\d+\s*/, '')}
                            </p>
                            <p className="text-lg font-extrabold text-text tabular-nums">
                              <AnimatedNumber value={withPercent.length} />
                            </p>
                          </div>
                          <div>
                            <p className="text-[10px] font-semibold text-muted uppercase tracking-wide">{t('stats.avg_per_tx')}</p>
                            <p className="text-sm font-bold text-text tabular-nums">
                              <AnimatedNumber value={avgPerTx} formatter={(v) => formatCents(v, baseCurrency)} />
                            </p>
                          </div>
                        </div>
                      </div>
                    </div>

                    {/* Category breakdown — scrollable */}
                    <div className="card-elevated overflow-hidden flex-1 min-h-0 flex flex-col">
                      <div className="overflow-y-auto no-scrollbar flex-1">
                        {withPercent.map((entry, i) => (
                          <CategoryRow
                            key={`${animationKey}-${entry.category_name}`}
                            entry={entry}
                            index={i}
                            color={entry.category_color || COLORS[i % COLORS.length]}
                            currency={baseCurrency}
                          />
                        ))}
                      </div>
                    </div>
                  </>
                ) : (
                  <div className="card-elevated">
                    <EmptyState icon={ChartBar} title={t('stats.no_stats')} />
                  </div>
                )}
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </div>

      {/* Date range modal */}
      <AnimatePresence>
        {showDatePicker && (
          <RangeDateModal
            initialFrom={customRange?.from ?? new Date(Date.now() - 30 * 86400000).toISOString().split('T')[0]}
            initialTo={customRange?.to ?? new Date().toISOString().split('T')[0]}
            onApply={handleCustomApply}
            onClose={() => setShowDatePicker(false)}
            labelFrom={t('stats.date_from')}
            labelTo={t('stats.date_to')}
            applyLabel={t('stats.apply')}
          />
        )}
      </AnimatePresence>
    </PageTransition>
  )
}
