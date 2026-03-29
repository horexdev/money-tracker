import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { AnimatePresence } from 'framer-motion'
import { Plus, X, CheckCircle, ArrowCircleDown, ArrowCircleUp, CalendarBlank, ClockCounterClockwise, Receipt, Target } from '@phosphor-icons/react'
import { fetchGoals, createGoal, updateGoal, depositGoal, withdrawGoal, deleteGoal, fetchGoalHistory } from '../api/goals'
import type { GoalTransaction } from '../api/goals'
import { accountsApi } from '../api/accounts'
import { formatCents, parseCents, formatDate } from '../lib/money'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { useHaptic } from '../hooks/useHaptic'
import { EmptyState, SwipeToDelete, SingleDateModal, fmtDisplay, FAB, BottomSheet } from '../components/ui'
import { useBaseCurrency } from '../hooks/useBaseCurrency'
import type { SavingsGoal } from '../types'

function sanitizeAmount(value: string): string {
  let cleaned = value.replace(/[^0-9.]/g, '')
  const dotIndex = cleaned.indexOf('.')
  if (dotIndex !== -1) {
    cleaned = cleaned.slice(0, dotIndex + 1) + cleaned.slice(dotIndex + 1).replace(/\./g, '')
  }
  if (dotIndex !== -1 && cleaned.length - dotIndex > 3) cleaned = cleaned.slice(0, dotIndex + 3)
  if (cleaned.length > 1 && cleaned[0] === '0' && cleaned[1] !== '.') cleaned = cleaned.slice(1)
  return cleaned
}

/* ─── Goal Row ─── */
function GoalRow({
  goal,
  onDeposit,
  onWithdraw,
  onDelete,
  onEdit,
  onHistory,
}: {
  goal: SavingsGoal
  onDeposit: (id: number) => void
  onWithdraw: (id: number) => void
  onDelete: (id: number) => void
  onEdit: (goal: SavingsGoal) => void
  onHistory: (id: number) => void
}) {
  const { t, i18n } = useTranslation()
  const { code: baseCurrency } = useBaseCurrency()
  const pct = Math.min(goal.progress_percent, 100)
  const color = goal.is_completed ? 'var(--color-income)' : 'var(--color-accent)'

  return (
    <SwipeToDelete onDelete={() => onDelete(goal.id)}>
      <div className="px-4 py-4 flex items-center gap-3">
        {/* Circular progress — tappable to edit */}
        <button
          onClick={() => onEdit(goal)}
          className="relative w-12 h-12 shrink-0 active:opacity-70 transition-opacity"
        >
          <svg viewBox="0 0 36 36" className="w-full h-full -rotate-90">
            <circle cx="18" cy="18" r="15.5" fill="none" stroke="var(--color-border)" strokeWidth="3.5" />
            <circle
              cx="18" cy="18" r="15.5" fill="none"
              stroke={color}
              strokeWidth="3.5"
              strokeDasharray={`${pct * 0.975} 100`}
              strokeLinecap="round"
              className="transition-all duration-700"
            />
          </svg>
          <div className="absolute inset-0 flex items-center justify-center">
            <span className="text-[9px] font-bold tabular-nums" style={{ color }}>{pct.toFixed(0)}%</span>
          </div>
        </button>

        {/* Info — tappable to edit */}
        <button
          onClick={() => onEdit(goal)}
          className="flex-1 min-w-0 text-left active:opacity-70 transition-opacity"
        >
          <div className="flex items-center gap-1.5 mb-0.5">
            <span className="text-[13px] font-bold text-text truncate">{goal.name}</span>
            {goal.is_completed && <CheckCircle size={14} weight="fill" className="text-income shrink-0" />}
          </div>
          <div className="flex items-center gap-1 text-xs text-muted">
            <span className="font-semibold text-text tabular-nums">{formatCents(goal.current_cents, baseCurrency)}</span>
            <span className="text-muted/40">·</span>
            <span className="tabular-nums">{formatCents(goal.target_cents, baseCurrency)}</span>
            {goal.deadline && (
              <>
                <span className="text-muted/40">·</span>
                <span>{formatDate(goal.deadline, i18n.language)}</span>
              </>
            )}
          </div>
        </button>

        {/* Action buttons */}
        <div className="flex gap-1 shrink-0">
          <button
            onClick={() => onWithdraw(goal.id)}
            className="w-10 h-10 rounded-2xl flex flex-col items-center justify-center gap-0.5 text-muted active:text-expense active:bg-expense/10 transition-colors"
          >
            <ArrowCircleUp size={18} weight="fill" />
            <span className="text-[8px] font-bold leading-none">{t('savings.withdraw')}</span>
          </button>
          <button
            onClick={() => onDeposit(goal.id)}
            className="w-10 h-10 rounded-2xl flex flex-col items-center justify-center gap-0.5 text-muted active:text-income active:bg-income/10 transition-colors"
          >
            <ArrowCircleDown size={18} weight="fill" />
            <span className="text-[8px] font-bold leading-none">{t('savings.deposit')}</span>
          </button>
          <button
            onClick={() => onHistory(goal.id)}
            className="w-10 h-10 rounded-2xl flex flex-col items-center justify-center gap-0.5 text-muted active:text-accent active:bg-accent-subtle transition-colors"
          >
            <ClockCounterClockwise size={18} weight="fill" />
            <span className="text-[8px] font-bold leading-none">{t('savings.history')}</span>
          </button>
        </div>
      </div>
    </SwipeToDelete>
  )
}

/* ─── Create / Edit Form (bottom sheet) ─── */
function GoalFormSheet({ onClose, editGoal }: { onClose: () => void; editGoal?: SavingsGoal }) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const { notification } = useHaptic()
  const { code: currencyCode, symbol } = useBaseCurrency()

  const isEdit = editGoal !== undefined
  const [name, setName] = useState(editGoal?.name ?? '')
  const [targetStr, setTargetStr] = useState(
    isEdit ? (editGoal!.target_cents / 100).toFixed(2) : ''
  )
  const [deadline, setDeadline] = useState(
    editGoal?.deadline ? editGoal.deadline.split('T')[0] : ''
  )
  const [showDeadlinePicker, setShowDeadlinePicker] = useState(false)
  const [accountId, setAccountId] = useState<number | null>(editGoal?.account_id ?? null)

  const { data: accounts = [] } = useQuery({
    queryKey: ['accounts'],
    queryFn: accountsApi.list,
  })

  const createMut = useMutation({
    mutationFn: () => createGoal({
      name: name.trim(),
      target_cents: parseCents(targetStr),
      currency_code: currencyCode,
      deadline: deadline || undefined,
      account_id: accountId,
    }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['goals'] })
      notification('success')
      onClose()
    },
    onError: () => notification('error'),
  })

  const updateMut = useMutation({
    mutationFn: () => updateGoal(editGoal!.id, {
      name: name.trim(),
      target_cents: parseCents(targetStr),
      deadline: deadline || undefined,
      account_id: accountId,
    }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['goals'] })
      notification('success')
      onClose()
    },
    onError: () => notification('error'),
  })

  const isPending = createMut.isPending || updateMut.isPending
  const isError = createMut.isError || updateMut.isError
  const errorMsg = ((createMut.error || updateMut.error) as Error | null)?.message
  const canSubmit = name.trim() && parseCents(targetStr) > 0 && !isPending

  return (
    <>
      <BottomSheet onClose={onClose}>
        <div className="px-5 pb-safe space-y-4" style={{ paddingBottom: 'max(1.5rem, env(safe-area-inset-bottom))' }}>

          {/* Goal name */}
          <div>
            <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
              {t('categories.name')}
            </label>
            <input
              type="text"
              placeholder={t('savings.create_new')}
              value={name}
              onChange={e => setName(e.target.value)}
              maxLength={50}
              autoFocus
              className="w-full bg-bg rounded-2xl px-4 py-3 text-sm font-medium outline-none text-text placeholder:text-muted/50 transition-shadow focus:shadow-[0_0_0_2px_rgba(99,102,241,0.2)]"
            />
          </div>

          {/* Target amount */}
          <div>
            <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
              {t('savings.target')}
            </label>
            <div className="flex items-baseline gap-2 bg-bg rounded-2xl px-4 py-3 focus-within:shadow-[0_0_0_2px_rgba(99,102,241,0.2)] transition-shadow">
              <span className="text-3xl font-bold text-muted/40 tabular-nums">{symbol}</span>
              <input
                inputMode="decimal"
                placeholder="0.00"
                value={targetStr}
                onChange={e => setTargetStr(sanitizeAmount(e.target.value))}
                className="flex-1 bg-transparent text-3xl font-bold outline-none text-text placeholder:text-muted/20 tabular-nums min-w-0"
              />
            </div>
          </div>

          {/* Deadline */}
          <div>
            <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
              {t('savings.deadline')}
            </label>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setShowDeadlinePicker(true)}
                className="flex-1 bg-bg rounded-2xl px-4 py-3 flex items-center justify-between active:bg-accent/5 transition-colors"
              >
                <span className={`text-sm font-semibold ${deadline ? 'text-text' : 'text-muted/40'}`}>
                  {deadline ? fmtDisplay(deadline) : t('savings.deadline') + '...'}
                </span>
                <CalendarBlank size={18} weight="bold" className="text-muted shrink-0" />
              </button>
              {deadline && (
                <button
                  onClick={() => setDeadline('')}
                  className="w-11 h-11 rounded-2xl bg-bg flex items-center justify-center text-muted active:text-destructive active:bg-destructive/10 transition-colors shrink-0"
                >
                  <X size={18} weight="bold" />
                </button>
              )}
            </div>
          </div>

          {/* Account link */}
          {accounts.length > 0 && (
            <div>
              <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
                {t('accounts.title')}
              </label>
              <div className="flex gap-2 overflow-x-auto no-scrollbar pb-1">
                <button
                  onClick={() => setAccountId(null)}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm font-medium transition-all whitespace-nowrap ${
                    accountId === null
                      ? 'bg-accent text-accent-text shadow-sm'
                      : 'bg-bg text-muted'
                  }`}
                >
                  <span className="text-xs">{t('common.none')}</span>
                </button>
                {accounts.map((acc) => (
                  <button
                    key={acc.id}
                    onClick={() => setAccountId(acc.id)}
                    className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm font-medium transition-all whitespace-nowrap ${
                      accountId === acc.id
                        ? 'text-white shadow-sm'
                        : 'bg-bg text-muted'
                    }`}
                    style={accountId === acc.id ? { backgroundColor: acc.color } : undefined}
                  >
                    <span className="text-xs">{acc.name}</span>
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Submit */}
          <button
            onClick={() => isEdit ? updateMut.mutate() : createMut.mutate()}
            disabled={!canSubmit}
            className={`
              w-full py-4 rounded-2xl text-[15px] font-bold transition-all active:scale-[0.98]
              ${canSubmit
                ? 'bg-accent text-accent-text shadow-[0_4px_16px_rgba(99,102,241,0.3)]'
                : 'bg-border text-muted'
              }
            `}
          >
            {isPending ? '...' : isEdit ? t('common.save') : t('common.create')}
          </button>

          {isError && (
            <p className="text-xs text-destructive text-center">{errorMsg}</p>
          )}
        </div>
      </BottomSheet>

      <AnimatePresence>
        {showDeadlinePicker && (
          <SingleDateModal
            value={deadline || new Date().toISOString().split('T')[0]}
            onApply={(iso) => setDeadline(iso)}
            onClose={() => setShowDeadlinePicker(false)}
            applyLabel={t('stats.apply')}
          />
        )}
      </AnimatePresence>
    </>
  )
}

/* ─── Amount Sheet (deposit / withdraw) ─── */
function AmountSheet({
  action,
  onConfirm,
  onClose,
  isPending,
  isError,
  error,
}: {
  action: 'deposit' | 'withdraw'
  onConfirm: (cents: number) => void
  onClose: () => void
  isPending: boolean
  isError: boolean
  error: Error | null
}) {
  const { t } = useTranslation()
  const { symbol } = useBaseCurrency()
  const [amountStr, setAmountStr] = useState('')
  const cents = parseCents(amountStr)
  const isDeposit = action === 'deposit'

  return (
    <BottomSheet onClose={onClose}>
      <div className="px-5 space-y-4" style={{ paddingBottom: 'max(1.5rem, env(safe-area-inset-bottom))' }}>
        <div className="flex items-center justify-between">
          <span className="text-base font-bold text-text">
            {isDeposit ? t('savings.deposit') : t('savings.withdraw')}
          </span>
          <button
            onClick={onClose}
            className="w-8 h-8 rounded-full bg-accent-subtle flex items-center justify-center text-muted active:scale-90 transition-transform"
          >
            <X size={14} weight="bold" />
          </button>
        </div>

        <div className="flex items-baseline gap-2 bg-bg rounded-2xl px-4 py-3 focus-within:shadow-[0_0_0_2px_rgba(99,102,241,0.2)] transition-shadow">
          <span className="text-3xl font-bold text-muted/40 tabular-nums">{symbol}</span>
          <input
            inputMode="decimal"
            placeholder="0.00"
            value={amountStr}
            onChange={e => setAmountStr(sanitizeAmount(e.target.value))}
            autoFocus
            className="flex-1 bg-transparent text-3xl font-bold outline-none text-text placeholder:text-muted/20 tabular-nums min-w-0"
          />
        </div>

        <button
          onClick={() => onConfirm(cents)}
          disabled={cents <= 0 || isPending}
          className={`
            w-full py-4 rounded-2xl text-[15px] font-bold transition-all active:scale-[0.98]
            ${cents > 0 && !isPending
              ? isDeposit
                ? 'bg-income text-white shadow-[0_4px_16px_rgba(34,197,94,0.3)]'
                : 'bg-expense text-white shadow-[0_4px_16px_rgba(239,68,68,0.3)]'
              : 'bg-border text-muted'
            }
          `}
        >
          {isPending ? '...' : t('common.confirm')}
        </button>

        {isError && (
          <p className="text-xs text-destructive text-center">{error?.message}</p>
        )}
      </div>
    </BottomSheet>
  )
}

/* ─── Goal History Sheet ─── */
function GoalHistorySheet({ goalId, onClose }: { goalId: number; onClose: () => void }) {
  const { t, i18n } = useTranslation()
  const { code: baseCurrency } = useBaseCurrency()

  const { data, isLoading } = useQuery({
    queryKey: ['goal-history', goalId],
    queryFn: () => fetchGoalHistory(goalId),
  })

  const history = data?.history ?? []

  return (
    <BottomSheet onClose={onClose}>
      <div
        className="px-5 overflow-y-auto no-scrollbar"
        style={{ maxHeight: '75dvh', paddingBottom: 'max(1.5rem, env(safe-area-inset-bottom))' }}
      >
        <h3 className="text-base font-bold text-text mb-3">{t('savings.history_title')}</h3>
        {isLoading ? (
          <div className="flex justify-center py-8"><Spinner /></div>
        ) : history.length === 0 ? (
          <EmptyState icon={Receipt} title={t('savings.no_history')} />
        ) : (
          <div className="space-y-2">
            {history.map((tx: GoalTransaction) => (
              <div key={tx.id} className="flex items-center gap-3 bg-bg rounded-2xl px-3 py-2.5">
                <div className={`w-8 h-8 rounded-2xl flex items-center justify-center shrink-0 ${
                  tx.type === 'deposit' ? 'bg-income/10' : 'bg-expense/10'
                }`}>
                  {tx.type === 'deposit'
                    ? <ArrowCircleDown size={18} weight="fill" className="text-income" />
                    : <ArrowCircleUp size={18} weight="fill" className="text-expense" />
                  }
                </div>
                <div className="flex-1 min-w-0">
                  <span className="text-[12px] font-semibold text-muted capitalize">{tx.type === 'deposit' ? t('savings.deposit') : t('savings.withdraw')}</span>
                  <p className="text-[11px] text-muted/60">{formatDate(tx.created_at, i18n.language)}</p>
                </div>
                <span className={`text-sm font-bold tabular-nums shrink-0 ${
                  tx.type === 'deposit' ? 'text-income' : 'text-expense'
                }`}>
                  {tx.type === 'deposit' ? '+' : '-'}{formatCents(tx.amount_cents, baseCurrency)}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </BottomSheet>
  )
}

/* ─── Main Page ─── */
export function SavingsPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { notification } = useHaptic()
  useTgBackButton(() => navigate('/more'))

  const [showForm, setShowForm] = useState(false)
  const [editingGoal, setEditingGoal] = useState<SavingsGoal | null>(null)
  const [amountFor, setAmountFor] = useState<{ id: number; action: 'deposit' | 'withdraw' } | null>(null)
  const [historyFor, setHistoryFor] = useState<number | null>(null)

  const goalsQ = useQuery({ queryKey: ['goals'], queryFn: fetchGoals })

  const depositMut = useMutation({
    mutationFn: ({ id, cents }: { id: number; cents: number }) => depositGoal(id, cents),
    onSuccess: () => {
      notification('success')
      qc.invalidateQueries({ queryKey: ['goals'] })
      setAmountFor(null)
    },
    onError: () => notification('error'),
  })

  const withdrawMut = useMutation({
    mutationFn: ({ id, cents }: { id: number; cents: number }) => withdrawGoal(id, cents),
    onSuccess: () => {
      notification('success')
      qc.invalidateQueries({ queryKey: ['goals'] })
      setAmountFor(null)
    },
    onError: () => notification('error'),
  })

  const deleteMut = useMutation({
    mutationFn: deleteGoal,
    onSuccess: () => {
      notification('success')
      qc.invalidateQueries({ queryKey: ['goals'] })
    },
    onError: () => notification('error'),
  })

  const goals = goalsQ.data?.goals ?? []
  const activeMut = amountFor?.action === 'deposit' ? depositMut : withdrawMut

  if (goalsQ.isPending) return <div className="flex justify-center py-16"><Spinner /></div>
  if (goalsQ.isError) return <ErrorMessage onRetry={() => goalsQ.refetch()} />

  return (
    <PageTransition>
      <div className="pt-3 pb-4">
          {goals.length === 0 ? (
            <div className="mx-4 card-elevated mt-2">
              <EmptyState
                icon={Target}
                title={t('savings.no_goals')}
                description={t('savings.start_saving')}
                action={
                  <button
                    onClick={() => { setEditingGoal(null); setAmountFor(null); setShowForm(true) }}
                    className="flex items-center gap-1.5 px-5 py-2.5 rounded-full bg-accent text-accent-text text-xs font-bold shadow-[0_2px_12px_rgba(99,102,241,0.4)] active:scale-95 transition-transform"
                  >
                    <Plus size={14} weight="bold" />
                    {t('savings.create_new')}
                  </button>
                }
              />
            </div>
          ) : (
            <div className="mx-4 card-elevated divide-y divide-border">
              {goals.map(goal => (
                <GoalRow
                  key={goal.id}
                  goal={goal}
                  onEdit={g => { setShowForm(false); setAmountFor(null); setHistoryFor(null); setEditingGoal(g) }}
                  onDeposit={id => { setShowForm(false); setEditingGoal(null); setHistoryFor(null); setAmountFor({ id, action: 'deposit' }) }}
                  onWithdraw={id => { setShowForm(false); setEditingGoal(null); setHistoryFor(null); setAmountFor({ id, action: 'withdraw' }) }}
                  onHistory={id => { setShowForm(false); setEditingGoal(null); setAmountFor(null); setHistoryFor(id) }}
                  onDelete={id => deleteMut.mutate(id)}
                />
              ))}
            </div>
          )}

          {deleteMut.isError && (
            <div className="mx-4 mt-2">
              <p className="text-xs text-destructive text-center bg-expense/10 rounded-2xl py-2 px-3">
                {(deleteMut.error as Error)?.message}
              </p>
            </div>
          )}
      </div>

      <FAB onClick={() => { setEditingGoal(null); setAmountFor(null); setShowForm(true) }} label={t('savings.create_new')} />

      {/* Bottom sheet modals */}
      <AnimatePresence>
        {showForm && (
          <GoalFormSheet onClose={() => setShowForm(false)} />
        )}
      </AnimatePresence>

      <AnimatePresence>
        {editingGoal && (
          <GoalFormSheet
            editGoal={editingGoal}
            onClose={() => setEditingGoal(null)}
          />
        )}
      </AnimatePresence>

      <AnimatePresence>
        {amountFor && (
          <AmountSheet
            action={amountFor.action}
            onConfirm={(cents) => {
              if (amountFor.action === 'deposit') {
                depositMut.mutate({ id: amountFor.id, cents })
              } else {
                withdrawMut.mutate({ id: amountFor.id, cents })
              }
            }}
            onClose={() => setAmountFor(null)}
            isPending={activeMut.isPending}
            isError={activeMut.isError}
            error={activeMut.error as Error | null}
          />
        )}
      </AnimatePresence>

      <AnimatePresence>
        {historyFor !== null && (
          <GoalHistorySheet goalId={historyFor} onClose={() => setHistoryFor(null)} />
        )}
      </AnimatePresence>
    </PageTransition>
  )
}
