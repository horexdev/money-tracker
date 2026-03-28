import { useState, useCallback, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { transactionsApi } from '../api/transactions'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { Card, SectionHeader, TransactionRow, EmptyState, Button } from '../components/ui'
import type { Transaction } from '../types'

const PAGE_SIZE = 20

function groupByDate(transactions: Transaction[], t: (key: string) => string): Map<string, Transaction[]> {
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
      label = new Date(tx.created_at).toLocaleDateString(undefined, {
        month: 'long',
        day: 'numeric',
        year: 'numeric',
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
    onMutate: (id) => setDeletingId(id),
    onSettled: () => setDeletingId(null),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['transactions'] })
      qc.invalidateQueries({ queryKey: ['balance'] })
    },
  })

  const handleDelete = useCallback((id: number) => {
    deleteMutation.mutate(id)
  }, [deleteMutation])

  const items = useMemo(() => data?.transactions ?? [], [data?.transactions])
  const totalPages = data?.total_pages ?? 1
  const grouped = useMemo(() => groupByDate(items, t), [items, t])

  if (isLoading) return <div className="flex justify-center py-16"><Spinner /></div>
  if (isError) return <ErrorMessage onRetry={refetch} />

  return (
    <PageTransition>
      <div className="py-4">
        <SectionHeader>{t('history.title')}</SectionHeader>

        {items.length === 0 ? (
          <Card className="mx-4" padding="p-0">
            <EmptyState icon="📋" title={t('transactions.no_transactions')} description={t('transactions.start_tracking')} />
          </Card>
        ) : (
          <div className="flex flex-col gap-4">
            {[...grouped.entries()].map(([date, txs]) => (
              <div key={date}>
                <p className="px-4 mb-1 text-xs font-medium text-muted">{date}</p>
                <Card className="mx-4" padding="p-0">
                  {txs.map((tx) => (
                    <TransactionRow
                      key={tx.id}
                      tx={tx}
                      onDelete={handleDelete}
                      isDeleting={deletingId === tx.id}
                    />
                  ))}
                </Card>
              </div>
            ))}
          </div>
        )}

        {totalPages > 1 && (
          <div className="flex items-center justify-center gap-4 mt-4">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
            >
              &larr; {t('common.back')}
            </Button>
            <span className="text-sm text-muted">
              {page} / {totalPages}
            </span>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
            >
              {t('common.done')} &rarr;
            </Button>
          </div>
        )}
      </div>
    </PageTransition>
  )
}
