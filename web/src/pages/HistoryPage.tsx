import { useState, useCallback, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { transactionsApi } from '../api/transactions'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { SectionHeader, TransactionRow, EmptyState, Button } from '../components/ui'
import type { Transaction } from '../types'

const PAGE_SIZE = 20

function groupByDate(transactions: Transaction[], t: (key: string) => string): Map<string, Transaction[]> {
  const groups = new Map<string, Transaction[]>()
  const now       = new Date()
  const today     = now.toDateString()
  const yesterday = new Date(now.getTime() - 86400000).toDateString()

  for (const tx of transactions) {
    const txDate = new Date(tx.created_at).toDateString()
    let label: string
    if (txDate === today) {
      label = t('history.today')
    } else if (txDate === yesterday) {
      label = t('history.yesterday')
    } else {
      label = new Date(tx.created_at).toLocaleDateString(undefined, {
        month: 'long', day: 'numeric', year: 'numeric',
      })
    }
    const existing = groups.get(label)
    if (existing) {
      existing.push(tx)
    } else {
      groups.set(label, [tx])
    }
  }
  return groups
}

export function HistoryPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [deletingId, setDeletingId] = useState<number | null>(null)

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['transactions', page],
    queryFn: () => transactionsApi.list(page, PAGE_SIZE),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => transactionsApi.delete(id),
    onMutate:   (id) => setDeletingId(id),
    onSettled:  () => setDeletingId(null),
    onSuccess:  () => {
      qc.invalidateQueries({ queryKey: ['transactions'] })
      qc.invalidateQueries({ queryKey: ['balance'] })
    },
  })

  const handleDelete = useCallback((id: number) => { deleteMutation.mutate(id) }, [deleteMutation])

  const items = useMemo(() => data?.transactions ?? [], [data?.transactions])
  const totalPages = data?.total_pages ?? 1
  const grouped    = useMemo(() => groupByDate(items, t), [items, t])

  if (isLoading) return <div className="flex justify-center py-16"><Spinner /></div>
  if (isError)   return <ErrorMessage onRetry={refetch} />

  return (
    <PageTransition>
      <div className="py-4">
        <h1 className="text-xl font-bold px-4 mb-4">{t('history.title')}</h1>

        {items.length === 0 ? (
          <div className="mx-4 bg-surface rounded-[--radius-card]">
            <EmptyState icon="📋" title={t('transactions.no_transactions')} description={t('transactions.start_tracking')} />
          </div>
        ) : (
          <div className="flex flex-col gap-4">
            {[...grouped.entries()].map(([date, txs]) => (
              <div key={date}>
                <SectionHeader>{date}</SectionHeader>
                <div className="mx-4 bg-surface rounded-[--radius-card] overflow-hidden">
                  {txs.map((tx) => (
                    <TransactionRow
                      key={tx.id}
                      tx={tx}
                      onDelete={handleDelete}
                      isDeleting={deletingId === tx.id}
                    />
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}

        {totalPages > 1 && (
          <div className="flex items-center justify-center gap-4 mt-4 px-4">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
            >
              <ChevronLeft size={16} /> {t('common.back')}
            </Button>
            <span className="text-sm text-muted font-medium">{page} / {totalPages}</span>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
            >
              {t('common.done')} <ChevronRight size={16} />
            </Button>
          </div>
        )}
      </div>
    </PageTransition>
  )
}
