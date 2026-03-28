import { useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { categoriesApi } from '../api/categories'
import { transactionsApi } from '../api/transactions'
import { parseCents } from '../lib/money'
import { useTgMainButton } from '../hooks/useMainButton'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { useHaptic } from '../hooks/useHaptic'
import { Spinner } from '../components/Spinner'
import { PageTransition } from '../components/PageTransition'
import { Card, SegmentedControl } from '../components/ui'
import type { TransactionType, Category } from '../types'

function AmountInput({ value, onChange, label }: { value: string; onChange: (v: string) => void; label: string }) {
  return (
    <Card className="mx-4">
      <label className="block text-xs text-muted mb-1">{label}</label>
      <input
        inputMode="decimal"
        placeholder="0.00"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full bg-transparent text-4xl font-bold text-text outline-none tabular-nums"
      />
    </Card>
  )
}

function CategoryGrid({
  categories,
  selected,
  onSelect,
  label,
}: {
  categories: Category[]
  selected: number | null
  onSelect: (id: number) => void
  label: string
}) {
  return (
    <Card className="mx-4" padding="p-0">
      <p className="px-4 pt-3 pb-2 text-xs uppercase tracking-widest text-muted">
        {label}
      </p>
      <div className="grid grid-cols-4 gap-2 px-3 pb-3">
        {categories.map((cat) => (
          <button
            key={cat.id}
            onClick={() => onSelect(cat.id)}
            className={`
              flex flex-col items-center justify-center gap-1 py-3
              rounded-[--radius-sm] text-xs transition-all duration-150
              active:scale-[0.95]
              ${selected === cat.id
                ? 'bg-accent-subtle text-accent ring-2 ring-accent'
                : 'text-text hover:bg-border'
              }
            `}
          >
            <span className="text-2xl leading-none">{cat.emoji}</span>
            <span className="text-center leading-tight px-1">{cat.name}</span>
          </button>
        ))}
      </div>
    </Card>
  )
}

export function AddTransactionPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const qc = useQueryClient()

  const { selection, notification } = useHaptic()

  const [type, setType] = useState<TransactionType>('expense')
  const [amount, setAmount] = useState('')
  const [categoryID, setCategoryID] = useState<number | null>(null)
  const [note, setNote] = useState('')

  const typeOptions = [
    { value: 'expense', label: t('transactions.expense') },
    { value: 'income', label: t('transactions.income') },
  ]

  const { data: catData, isLoading } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  })

  const mutation = useMutation({
    mutationFn: transactionsApi.create,
    onSuccess: () => {
      notification('success')
      qc.invalidateQueries({ queryKey: ['transactions'] })
      qc.invalidateQueries({ queryKey: ['balance'] })
      navigate('/')
    },
    onError: () => notification('error'),
  })

  const canSubmit = parseCents(amount) > 0 && categoryID !== null && !mutation.isPending

  const handleSubmit = useCallback(() => {
    if (!canSubmit || categoryID === null) return
    mutation.mutate({
      category_id: categoryID,
      type,
      amount_cents: parseCents(amount),
      note: note.trim() || undefined,
    })
  }, [canSubmit, categoryID, type, amount, note, mutation])

  useTgBackButton(() => navigate('/'), true)
  useTgMainButton({
    text: mutation.isPending ? t('common.loading') : t('common.save'),
    onClick: handleSubmit,
    enabled: canSubmit,
    loading: mutation.isPending,
  })

  return (
    <PageTransition>
      <div className="flex flex-col gap-4 py-4">
        <div className="mx-4">
          <SegmentedControl
            options={typeOptions}
            value={type}
            onChange={(v) => { setType(v as TransactionType); setCategoryID(null); selection() }}
          />
        </div>

        <AmountInput value={amount} onChange={setAmount} label={t('transactions.amount')} />

        <Card className="mx-4">
          <label className="block text-xs text-muted mb-1">{t('transactions.note')}</label>
          <input
            type="text"
            placeholder={t('transactions.note_placeholder')}
            value={note}
            onChange={(e) => setNote(e.target.value)}
            maxLength={120}
            className="w-full bg-transparent text-sm text-text outline-none"
          />
        </Card>

        {isLoading ? (
          <div className="flex justify-center py-4"><Spinner /></div>
        ) : (
          <CategoryGrid
            categories={catData?.categories ?? []}
            selected={categoryID}
            onSelect={(id) => { setCategoryID(id); selection() }}
            label={t('transactions.category')}
          />
        )}

        {mutation.isError && (
          <p className="px-4 text-sm text-center text-destructive">
            {(mutation.error as Error).message}
          </p>
        )}
      </div>
    </PageTransition>
  )
}
