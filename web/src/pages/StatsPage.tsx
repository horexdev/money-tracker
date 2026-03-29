import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { statsApi } from '../api/stats'
import { formatCents } from '../lib/money'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { SegmentedControl, EmptyState } from '../components/ui'
import type { TransactionType, CategoryStat } from '../types'

const COLORS = [
  '#22c55e', '#06b6d4', '#f59e0b', '#ef4444',
  '#8b5cf6', '#3b82f6', '#ec4899', '#f97316',
]

type Period = 'month' | 'week' | 'today' | 'lastmonth'

export function StatsPage() {
  const { t } = useTranslation()
  const [type, setType]     = useState<TransactionType>('expense')
  const [period, setPeriod] = useState<Period>('month')

  const typeOptions = [
    { value: 'expense', label: t('transactions.expense') },
    { value: 'income',  label: t('transactions.income') },
  ]

  const periodOptions = [
    { value: 'today',     label: t('stats.today') },
    { value: 'week',      label: t('stats.week') },
    { value: 'month',     label: t('stats.month') },
    { value: 'lastmonth', label: t('stats.last_month') },
  ]

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['stats', period],
    queryFn: () => statsApi.get(period),
  })

  const filtered: CategoryStat[] = data?.items.filter((s) => s.type === type) ?? []
  const total = filtered.reduce((sum, s) => sum + s.total_cents, 0)
  const withPercent = filtered.map((s) => ({
    ...s,
    percent: total > 0 ? (s.total_cents / total) * 100 : 0,
  }))

  return (
    <PageTransition>
      <div className="py-4 flex flex-col gap-3 px-4">
        <h1 className="text-xl font-bold text-text">{t('stats.title')}</h1>

        <SegmentedControl
          options={typeOptions}
          value={type}
          onChange={(v) => setType(v as TransactionType)}
        />

        <SegmentedControl
          options={periodOptions}
          value={period}
          onChange={(v) => setPeriod(v as Period)}
          size="sm"
        />

        {isLoading && <div className="flex justify-center py-16"><Spinner /></div>}
        {isError   && <ErrorMessage onRetry={refetch} />}

        {data && (
          <>
            {withPercent.length > 0 ? (
              <>
                {/* Donut chart */}
                <div className="bg-surface rounded-[--radius-card] flex justify-center py-6">
                  <DonutChart data={withPercent} total={total} />
                </div>

                {/* Breakdown list */}
                <div className="bg-surface rounded-[--radius-card] overflow-hidden">
                  {withPercent.map((entry, i) => (
                    <div
                      key={`${entry.category_name}-${i}`}
                      className="flex items-center gap-3 px-4 py-3 border-b border-border last:border-b-0"
                    >
                      <div
                        className="w-2.5 h-2.5 rounded-full shrink-0"
                        style={{ background: COLORS[i % COLORS.length] }}
                      />
                      <span className="text-xl w-7 text-center">{entry.category_emoji}</span>
                      <div className="flex-1 min-w-0">
                        <div className="flex justify-between items-center">
                          <span className="text-sm text-text truncate">{entry.category_name}</span>
                          <span className="text-sm font-semibold tabular-nums text-text ml-2 shrink-0">
                            {formatCents(entry.total_cents, entry.currency_code)}
                          </span>
                        </div>
                        <div className="mt-1.5 h-1.5 rounded-full overflow-hidden bg-border">
                          <div
                            className="h-full rounded-full transition-all duration-500"
                            style={{ width: `${entry.percent}%`, background: COLORS[i % COLORS.length] }}
                          />
                        </div>
                      </div>
                      <span className="text-xs text-muted shrink-0 w-9 text-right">
                        {entry.percent.toFixed(0)}%
                      </span>
                    </div>
                  ))}
                </div>
              </>
            ) : (
              <div className="bg-surface rounded-[--radius-card]">
                <EmptyState icon="📊" title={t('stats.no_stats')} />
              </div>
            )}
          </>
        )}
      </div>
    </PageTransition>
  )
}

function DonutChart({ data, total }: { data: Array<{ percent: number; category_name: string }>; total: number }) {
  const size   = 160
  const cx     = size / 2
  const cy     = size / 2
  const outerR = 68
  const innerR = 46
  const gap    = 0.02

  let startAngle = -Math.PI / 2
  const segments = data.map((d, i) => {
    const sweep = (d.percent / 100) * (2 * Math.PI) - gap
    const s     = startAngle
    startAngle += sweep + gap

    const x1 = cx + outerR * Math.cos(s)
    const y1 = cy + outerR * Math.sin(s)
    const x2 = cx + outerR * Math.cos(s + sweep)
    const y2 = cy + outerR * Math.sin(s + sweep)
    const x3 = cx + innerR * Math.cos(s + sweep)
    const y3 = cy + innerR * Math.sin(s + sweep)
    const x4 = cx + innerR * Math.cos(s)
    const y4 = cy + innerR * Math.sin(s)

    const large = sweep > Math.PI ? 1 : 0
    const path  = [
      `M ${x1} ${y1}`,
      `A ${outerR} ${outerR} 0 ${large} 1 ${x2} ${y2}`,
      `L ${x3} ${y3}`,
      `A ${innerR} ${innerR} 0 ${large} 0 ${x4} ${y4}`,
      'Z',
    ].join(' ')

    return <path key={i} d={path} fill={COLORS[i % COLORS.length]} />
  })

  return (
    <div className="relative" style={{ width: size, height: size }}>
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
        {segments}
      </svg>
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <div className="text-center">
          <p className="text-xs text-muted">Total</p>
          <p className="text-sm font-bold tabular-nums text-text">{formatCents(total)}</p>
        </div>
      </div>
    </div>
  )
}
