import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { AnimatePresence } from 'framer-motion'
import { Plus, Pause, Play, ArrowsClockwise } from '@phosphor-icons/react'
import { fetchRecurring, createRecurring, updateRecurring, toggleRecurring, deleteRecurring } from '../../shared/api/recurring'
import { categoriesApi } from '../../shared/api/categories'
import { accountsApi } from '../../shared/api/accounts'
import { formatCents, parseCents, formatDate } from '../../shared/lib/money'
import { friendlyError } from '../../shared/lib/errors'
import { CategoryIcon } from '../../shared/lib/categoryIcons'
import { AmountInput } from '../../shared/ui/AmountInput'
import { Spinner } from '../../shared/ui/Spinner'
import { ErrorMessage } from '../../shared/ui/ErrorMessage'
import { PageTransition } from '../../shared/ui/PageTransition'
import { useTgBackButton } from '../../shared/hooks/useTelegramApp'
import { useHaptic } from '../../shared/hooks/useHaptic'
import { Badge, EmptyState, ActionRow, FAB, BottomSheet } from '../../shared/ui'
import { AccountDropdown } from '../../shared/ui/AccountDropdown'
import { useCategoryName } from '../../shared/hooks/useCategoryName'
import { useBaseCurrency } from '../../shared/hooks/useBaseCurrency'
import type { RecurringTransaction, TransactionType } from '../../shared/types'

const FREQ_OPTIONS = [
  { value: 'daily',   labelKey: 'recurring.daily' },
  { value: 'weekly',  labelKey: 'recurring.weekly' },
  { value: 'monthly', labelKey: 'recurring.monthly' },
  { value: 'yearly',  labelKey: 'recurring.yearly' },
]

/* ─── Recurring Card ─── */
function RecurringCard({
  item,
  onEdit,
  onToggle,
  onDelete,
}: {
  item: RecurringTransaction
  onEdit: (item: RecurringTransaction) => void
  onToggle: (id: number) => void
  onDelete: (id: number) => void
}) {
  const { t, i18n } = useTranslation()
  const tCategory = useCategoryName()
  const { code: baseCurrency } = useBaseCurrency()
  const nextDate = formatDate(item.next_run_at, i18n.language)

  return (
    <ActionRow onDelete={() => onDelete(item.id)}>
      <div className="flex items-center gap-3 px-4 py-3.5">
        <button
          onClick={() => onEdit(item)}
          className={`w-10 h-10 rounded-2xl flex items-center justify-center shrink-0 transition-opacity active:scale-95 ${!item.is_active ? 'opacity-40' : ''}`}
          style={{ background: item.category_color || 'var(--color-accent)' }}
        >
          <CategoryIcon icon={item.category_icon} size={20} weight="fill" className="text-white" />
        </button>
        <button
          onClick={() => onEdit(item)}
          className={`flex-1 min-w-0 text-left transition-opacity ${!item.is_active ? 'opacity-40' : ''}`}
        >
          <div className="flex items-center gap-2 mb-0.5">
            <span className="text-[13px] font-bold text-text truncate">{tCategory(item.category_name)}</span>
            <Badge variant={item.type === 'income' ? 'income' : 'expense'} className="text-[10px] shrink-0">
              {item.type === 'income' ? t('transactions.income') : t('transactions.expense')}
            </Badge>
            {!item.is_active && (
              <span className="text-[10px] font-bold text-muted bg-border px-1.5 py-0.5 rounded-full shrink-0">
                {t('recurring.paused')}
              </span>
            )}
          </div>
          <div className="flex items-center gap-1 text-xs text-muted">
            <span className="font-semibold text-text tabular-nums">{formatCents(item.amount_cents, baseCurrency)}</span>
            <span className="text-muted/40">·</span>
            <span className="capitalize">{t(`recurring.${item.frequency}`)}</span>
            <span className="text-muted/40">·</span>
            <span>{nextDate}</span>
          </div>
          {item.note && <p className="text-[11px] text-muted/70 mt-0.5 truncate">{item.note}</p>}
        </button>
        <button
          onClick={() => onToggle(item.id)}
          className="w-11 h-11 rounded-2xl flex items-center justify-center text-muted active:text-accent active:bg-accent-subtle transition-colors shrink-0"
        >
          {item.is_active ? <Pause size={18} weight="fill" /> : <Play size={18} weight="fill" />}
        </button>
      </div>
    </ActionRow>
  )
}

/* ─── Form (create or edit, bottom sheet) ─── */
function RecurringForm({
  editItem,
  onClose,
}: {
  editItem: RecurringTransaction | null
  onClose: () => void
}) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const { notification } = useHaptic()
  const tCategory = useCategoryName()
  const { code: baseCurrency } = useBaseCurrency()

  const isEdit = editItem !== null

  const [type, setType] = useState<TransactionType>(editItem?.type ?? 'expense')
  const [amount, setAmount] = useState(editItem ? String(editItem.amount_cents / 100) : '')
  const [categoryID, setCategoryID] = useState<number | null>(editItem?.category_id ?? null)
  const [note, setNote] = useState(editItem?.note ?? '')
  const [frequency, setFrequency] = useState(editItem?.frequency ?? 'monthly')
  const [accountID, setAccountID] = useState<number | null>(editItem?.account_id ?? null)

  const { data: accounts = [] } = useQuery({ queryKey: ['accounts'], queryFn: accountsApi.list })
  const catsQ = useQuery({ queryKey: ['categories', { order: 'frequency' }], queryFn: () => categoriesApi.list(undefined, 'frequency') })
  const categories = catsQ.data?.categories ?? []
  const filtered = categories.filter(c => c.type === type || c.type === 'both')

  useEffect(() => {
    if (accounts.length === 0) return
    const def = accounts.find(a => a.is_default) ?? accounts[0]
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (accountID === null) setAccountID(def.id)
  }, [accounts]) // eslint-disable-line react-hooks/exhaustive-deps

  const selectedAccount = accounts.find(a => a.id === accountID)
  const currencyCode = selectedAccount?.currency_code ?? baseCurrency

  const createMut = useMutation({
    mutationFn: () => createRecurring({
      account_id: accountID!,
      type,
      amount_cents: parseCents(amount),
      currency_code: currencyCode,
      category_id: categoryID!,
      note: note.trim(),
      frequency,
    }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['recurring'] })
      notification('success')
      onClose()
    },
  })

  const updateMut = useMutation({
    mutationFn: () => updateRecurring(editItem!.id, {
      account_id: accountID!,
      type,
      amount_cents: parseCents(amount),
      currency_code: currencyCode,
      category_id: categoryID!,
      note: note.trim(),
      frequency,
    }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['recurring'] })
      notification('success')
      onClose()
    },
  })

  const mut = isEdit ? updateMut : createMut
  const canSubmit = parseCents(amount) > 0 && categoryID !== null && accountID !== null && !mut.isPending

  return (
    <BottomSheet onClose={onClose}>
      <div
        className="px-5 space-y-4 overflow-y-auto no-scrollbar"
        style={{ maxHeight: '80dvh', paddingBottom: 'max(1.5rem, env(safe-area-inset-bottom))' }}
      >
        {/* Type toggle */}
        <div className="flex gap-1.5">
          {(['expense', 'income'] as TransactionType[]).map((v) => (
            <button
              key={v}
              onClick={() => { setType(v); setCategoryID(null) }}
              className={`
                flex-1 py-2.5 rounded-2xl text-[13px] font-bold transition-all duration-200 select-none
                ${type === v
                  ? 'bg-accent text-accent-text shadow-(--shadow-accent-pill)'
                  : 'bg-accent-subtle text-muted'
                }
              `}
            >
              {v === 'expense' ? t('transactions.expense') : t('transactions.income')}
            </button>
          ))}
        </div>

        {/* Account */}
        {accounts.length > 0 && accountID !== null && (
          <AccountDropdown
            accounts={accounts}
            selectedId={accountID}
            onChange={id => id !== null && setAccountID(id)}
            showBalance
            variant="surface"
          />
        )}

        {/* Amount */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('transactions.amount')}
          </label>
          <AmountInput value={amount} onChange={setAmount} currency={currencyCode} />
        </div>

        {/* Frequency */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('recurring.frequency')}
          </label>
          <div className="grid grid-cols-2 gap-1.5">
            {FREQ_OPTIONS.map((opt) => (
              <button
                key={opt.value}
                onClick={() => setFrequency(opt.value)}
                className={`
                  py-2.5 rounded-2xl text-[13px] font-bold transition-all duration-200 select-none
                  ${frequency === opt.value
                    ? 'bg-accent text-accent-text shadow-(--shadow-accent-pill)'
                    : 'bg-accent-subtle text-muted'
                  }
                `}
              >
                {t(opt.labelKey)}
              </button>
            ))}
          </div>
        </div>

        {/* Category */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('transactions.category')}
          </label>
          <div className="grid grid-cols-4 gap-2">
            {filtered.map(cat => {
              const isActive = categoryID === cat.id
              return (
                <button
                  key={cat.id}
                  onClick={() => setCategoryID(cat.id)}
                  className={`
                    flex flex-col items-center gap-1.5 py-2.5 rounded-2xl text-xs transition-all duration-150 active:scale-95 select-none
                    ${isActive
                      ? 'bg-accent/10 shadow-(--shadow-accent-pill)'
                      : 'bg-surface shadow-sm'
                    }
                  `}
                >
                  <div
                    className="w-9 h-9 rounded-2xl flex items-center justify-center"
                    style={{ background: isActive ? 'var(--color-accent)' : (cat.color || 'var(--color-accent)') }}
                  >
                    <CategoryIcon icon={cat.icon} size={18} weight="fill" className="text-white" />
                  </div>
                  <span className="truncate w-full text-center px-1 font-medium text-[10px] text-text">
                    {tCategory(cat.name)}
                  </span>
                </button>
              )
            })}
          </div>
        </div>

        {/* Note */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('transactions.note')}
          </label>
          <input
            type="text"
            placeholder={t('transactions.note_placeholder')}
            value={note}
            onChange={e => setNote(e.target.value)}
            maxLength={120}
            className="w-full bg-bg rounded-2xl px-4 py-3 text-sm font-medium outline-none text-text placeholder:text-muted/50 transition-shadow focus:shadow-(--shadow-focus)"
          />
        </div>

        {/* Submit */}
        <button
          onClick={() => mut.mutate()}
          disabled={!canSubmit}
          className={`
            w-full py-4 rounded-2xl text-[15px] font-bold transition-all active:scale-[0.98]
            ${canSubmit
              ? 'bg-accent text-accent-text shadow-(--shadow-button)'
              : 'bg-border text-muted'
            }
          `}
        >
          {mut.isPending ? t('common.loading') : isEdit ? t('common.save') : t('common.create')}
        </button>

        {mut.isError && (
          <p className="text-xs text-destructive text-center">
            {friendlyError(mut.error, t)}
          </p>
        )}
      </div>
    </BottomSheet>
  )
}

/* ─── Main Page ─── */
export function RecurringPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { notification } = useHaptic()
  useTgBackButton(() => navigate('/more'))

  const [editingItem, setEditingItem] = useState<RecurringTransaction | null>(null)
  const [showCreate, setShowCreate] = useState(false)

  const recurringQ = useQuery({ queryKey: ['recurring'], queryFn: fetchRecurring })

  const toggleMut = useMutation({
    mutationFn: toggleRecurring,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['recurring'] })
      notification('success')
    },
  })

  const deleteMut = useMutation({
    mutationFn: deleteRecurring,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['recurring'] })
      notification('success')
    },
    onError: () => notification('error'),
  })

  const items = recurringQ.data?.recurring ?? []

  if (recurringQ.isPending) return <div className="flex justify-center py-16"><Spinner /></div>
  if (recurringQ.isError) return <ErrorMessage onRetry={() => recurringQ.refetch()} />

  const formOpen = showCreate || editingItem !== null

  return (
    <PageTransition>
      <div className="pt-3 pb-4">
          {items.length === 0 ? (
            <div className="mx-4 card-elevated mt-2">
              <EmptyState
                icon={ArrowsClockwise}
                title={t('recurring.no_recurring')}
                description={t('recurring.setup_recurring')}
                action={
                  <button
                    onClick={() => setShowCreate(true)}
                    className="flex items-center gap-1.5 px-5 py-2.5 rounded-full bg-accent text-accent-text text-xs font-bold shadow-(--shadow-accent-pill) active:scale-95 transition-transform"
                  >
                    <Plus size={14} weight="bold" />
                    {t('recurring.create_new')}
                  </button>
                }
              />
            </div>
          ) : (
            <div className="mx-4 card-elevated divide-y divide-border">
              {items.map(item => (
                <RecurringCard
                  key={item.id}
                  item={item}
                  onEdit={setEditingItem}
                  onToggle={id => toggleMut.mutate(id)}
                  onDelete={id => deleteMut.mutate(id)}
                />
              ))}
            </div>
          )}

          {deleteMut.isError && (
            <div className="mx-4 mt-2">
              <p className="text-xs text-destructive text-center bg-expense/10 rounded-2xl py-2 px-3">
                {friendlyError(deleteMut.error, t)}
              </p>
            </div>
          )}
      </div>

      <FAB onClick={() => setShowCreate(true)} label={t('recurring.create_new')} />

      {/* Bottom sheet form */}
      <AnimatePresence>
        {formOpen && (
          <RecurringForm
            key={editingItem?.id ?? 'new'}
            editItem={editingItem}
            onClose={() => { setShowCreate(false); setEditingItem(null) }}
          />
        )}
      </AnimatePresence>
    </PageTransition>
  )
}
