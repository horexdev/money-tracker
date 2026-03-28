import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Plus, Trash2, Pause, Play } from 'lucide-react'
import { fetchRecurring, createRecurring, toggleRecurring, deleteRecurring } from '../api/recurring'
import { categoriesApi } from '../api/categories'
import { formatCents, parseCents } from '../lib/money'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { Card, Button, Badge, EmptyState, SegmentedControl } from '../components/ui'
import type { RecurringTransaction, TransactionType } from '../types'

function RecurringCard({
  item,
  onToggle,
  onDelete,
}: {
  item: RecurringTransaction
  onToggle: (id: number) => void
  onDelete: (id: number) => void
}) {
  const { t } = useTranslation()
  const nextDate = new Date(item.next_run_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })

  return (
    <div className={`px-4 py-3 border-b border-border last:border-b-0 ${!item.is_active ? 'opacity-50' : ''}`}>
      <div className="flex items-center gap-3">
        <span className="text-xl w-8 text-center">{item.category_emoji}</span>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium truncate">{item.category_name}</span>
            <Badge variant={item.type === 'income' ? 'income' : 'expense'}>
              {item.type === 'income' ? t('transactions.income') : t('transactions.expense')}
            </Badge>
          </div>
          <div className="flex items-center gap-2 mt-0.5 text-xs text-muted">
            <span>{formatCents(item.amount_cents, item.currency_code)}</span>
            <span>·</span>
            <span className="capitalize">{t(`recurring.${item.frequency}`)}</span>
            <span>·</span>
            <span>{t('recurring.next_run')}: {nextDate}</span>
          </div>
          {item.note && <p className="text-xs text-muted mt-0.5 truncate">{item.note}</p>}
        </div>
        <button onClick={() => onToggle(item.id)} className="p-1.5 text-muted hover:text-accent">
          {item.is_active ? <Pause size={16} /> : <Play size={16} />}
        </button>
        <button onClick={() => onDelete(item.id)} className="p-1.5 text-muted hover:text-destructive">
          <Trash2 size={16} />
        </button>
      </div>
    </div>
  )
}

const FREQ_OPTIONS = [
  { value: 'daily', labelKey: 'recurring.daily' },
  { value: 'weekly', labelKey: 'recurring.weekly' },
  { value: 'monthly', labelKey: 'recurring.monthly' },
  { value: 'yearly', labelKey: 'recurring.yearly' },
]

export function RecurringPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const qc = useQueryClient()
  useTgBackButton(() => navigate('/more'))

  const [showForm, setShowForm] = useState(false)
  const [type, setType] = useState<TransactionType>('expense')
  const [amount, setAmount] = useState('')
  const [categoryID, setCategoryID] = useState<number | null>(null)
  const [note, setNote] = useState('')
  const [frequency, setFrequency] = useState('monthly')

  const recurringQ = useQuery({ queryKey: ['recurring'], queryFn: fetchRecurring })
  const catsQ = useQuery({ queryKey: ['categories'], queryFn: () => categoriesApi.list() })

  const createMut = useMutation({
    mutationFn: () => createRecurring({
      type,
      amount_cents: parseCents(amount),
      currency_code: 'USD',
      category_id: categoryID!,
      note: note.trim(),
      frequency,
    }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['recurring'] })
      resetForm()
    },
  })

  const toggleMut = useMutation({
    mutationFn: toggleRecurring,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recurring'] }),
  })

  const deleteMut = useMutation({
    mutationFn: deleteRecurring,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recurring'] }),
  })

  function resetForm() {
    setShowForm(false)
    setType('expense')
    setAmount('')
    setCategoryID(null)
    setNote('')
    setFrequency('monthly')
  }

  const items = recurringQ.data?.recurring ?? []
  const categories = catsQ.data?.categories ?? []
  const canSubmit = parseCents(amount) > 0 && categoryID !== null && !createMut.isPending

  if (recurringQ.isPending) return <div className="flex justify-center py-16"><Spinner /></div>
  if (recurringQ.isError) return <ErrorMessage onRetry={() => recurringQ.refetch()} />

  return (
    <PageTransition>
      <div className="p-4 space-y-4">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-bold">{t('recurring.title')}</h1>
          {!showForm && (
            <Button size="sm" onClick={() => setShowForm(true)}>
              <Plus size={16} className="mr-1" /> {t('recurring.create_new')}
            </Button>
          )}
        </div>

        {showForm && (
          <Card>
            <div className="space-y-3">
              <SegmentedControl
                options={[
                  { value: 'expense', label: t('transactions.expense') },
                  { value: 'income', label: t('transactions.income') },
                ]}
                value={type}
                onChange={v => { setType(v as TransactionType); setCategoryID(null) }}
                size="sm"
              />

              <div>
                <label className="block text-xs text-muted mb-1">{t('transactions.amount')}</label>
                <input
                  inputMode="decimal"
                  placeholder="0.00"
                  value={amount}
                  onChange={e => setAmount(e.target.value)}
                  className="w-full bg-surface rounded-[--radius-sm] px-3 py-2 text-2xl font-bold outline-none focus:ring-2 focus:ring-accent tabular-nums"
                />
              </div>

              <div>
                <label className="block text-xs text-muted mb-1">{t('recurring.frequency')}</label>
                <SegmentedControl
                  options={FREQ_OPTIONS.map(o => ({ value: o.value, label: t(o.labelKey) }))}
                  value={frequency}
                  onChange={setFrequency}
                  size="sm"
                />
              </div>

              <div>
                <label className="block text-xs text-muted mb-1">{t('transactions.category')}</label>
                <div className="grid grid-cols-4 gap-2">
                  {categories.map(cat => (
                    <button
                      key={cat.id}
                      onClick={() => setCategoryID(cat.id)}
                      className={`flex flex-col items-center gap-1 py-2 rounded-[--radius-sm] text-xs transition-all
                        ${categoryID === cat.id ? 'bg-accent-subtle text-accent ring-2 ring-accent' : 'text-text hover:bg-border'}`}
                    >
                      <span className="text-lg">{cat.emoji}</span>
                      <span className="truncate w-full text-center px-1">{cat.name}</span>
                    </button>
                  ))}
                </div>
              </div>

              <div>
                <label className="block text-xs text-muted mb-1">{t('transactions.note')}</label>
                <input
                  type="text"
                  placeholder={t('transactions.note_placeholder')}
                  value={note}
                  onChange={e => setNote(e.target.value)}
                  maxLength={120}
                  className="w-full bg-surface rounded-[--radius-sm] px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-accent"
                />
              </div>

              <div className="flex gap-2">
                <Button size="sm" onClick={() => createMut.mutate()} disabled={!canSubmit}>
                  {t('common.create')}
                </Button>
                <Button size="sm" variant="ghost" onClick={resetForm}>{t('common.cancel')}</Button>
              </div>

              {createMut.isError && (
                <p className="text-xs text-destructive">{(createMut.error as Error)?.message}</p>
              )}
            </div>
          </Card>
        )}

        {items.length > 0 ? (
          <Card padding="p-0">
            {items.map(item => (
              <RecurringCard
                key={item.id}
                item={item}
                onToggle={id => toggleMut.mutate(id)}
                onDelete={id => deleteMut.mutate(id)}
              />
            ))}
          </Card>
        ) : !showForm ? (
          <EmptyState icon="🔄" title={t('recurring.no_recurring')} description={t('recurring.setup_recurring')} />
        ) : null}
      </div>
    </PageTransition>
  )
}
