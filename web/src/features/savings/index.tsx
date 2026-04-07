import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { AnimatePresence } from 'framer-motion'
import { Plus, X, CheckCircle, ArrowCircleDown, ArrowCircleUp, CalendarBlank, ClockCounterClockwise, Receipt, Target } from '@phosphor-icons/react'
import { fetchGoals, createGoal, updateGoal, depositGoal, withdrawGoal, deleteGoal, fetchGoalHistory } from '../../shared/api/goals'
import type { GoalTransaction } from '../../shared/api/goals'
import { accountsApi } from '../../shared/api/accounts'
import { formatCents, parseCents, formatDate, sanitizeAmount } from '../../shared/lib/money'
import { CurrencyBadge } from '../../shared/lib/currencyIcons'
import { friendlyError } from '../../shared/lib/errors'
import { Spinner } from '../../shared/ui/Spinner'
import { ErrorMessage } from '../../shared/ui/ErrorMessage'
import { PageTransition } from '../../shared/ui/PageTransition'
import { useTgBackButton } from '../../shared/hooks/useTelegramApp'
import { useHaptic } from '../../shared/hooks/useHaptic'
import { EmptyState, ActionRow, SingleDateModal, fmtDisplay, FAB, BottomSheet } from '../../shared/ui'
import { useBaseCurrency } from '../../shared/hooks/useBaseCurrency'
import type { SavingsGoal } from '../../shared/types'

/* ─── Goal Card ─── */
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
  const isCompleted = goal.is_completed
  const progressColor = isCompleted ? 'var(--color-income)' : 'var(--color-accent)'

  return (
    <ActionRow onDelete={() => onDelete(goal.id)}>
      <button
        onClick={() => onEdit(goal)}
        className="w-full text-left px-4 pt-4 pb-3 active:opacity-70 transition-opacity"
      >
        {/* Header row: name + completion badge */}
        <div className="flex items-center gap-2 mb-1">
          <span className="text-[14px] font-bold text-text flex-1 min-w-0 truncate">{goal.name}</span>
          {isCompleted && <CheckCircle size={16} weight="fill" className="text-income shrink-0" />}
          {goal.deadline && (
            <span className="text-[10px] font-semibold text-muted shrink-0">
              {formatDate(goal.deadline, i18n.language)}
            </span>
          )}
        </div>

        {/* Amounts */}
        <div className="flex items-baseline gap-1 mb-2.5">
          <span className="text-[18px] font-bold tabular-nums" style={{ color: progressColor }}>
            {formatCents(goal.current_cents, baseCurrency)}
          </span>
          <span className="text-[13px] text-muted font-medium tabular-nums">
            / {formatCents(goal.target_cents, baseCurrency)}
          </span>
          <span className="ml-auto text-[12px] font-bold tabular-nums" style={{ color: progressColor }}>
            {pct.toFixed(0)}%
          </span>
        </div>

        {/* Horizontal progress bar */}
        <div className="h-2 rounded-full bg-border overflow-hidden">
          <div
            className="h-full rounded-full transition-all duration-700"
            style={{ width: `${pct}%`, background: progressColor }}
          />
        </div>
      </button>

      {/* Action buttons row */}
      <div className="flex gap-2 px-4 pb-4">
        <button
          onClick={() => onWithdraw(goal.id)}
          className="flex-1 h-12 rounded-2xl flex items-center justify-center gap-2 bg-expense/8 text-expense active:bg-expense/20 transition-colors"
        >
          <ArrowCircleUp size={20} weight="fill" />
          <span className="text-[13px] font-bold">{t('savings.withdraw')}</span>
        </button>
        <button
          onClick={() => onDeposit(goal.id)}
          className="flex-1 h-12 rounded-2xl flex items-center justify-center gap-2 bg-income/8 text-income active:bg-income/20 transition-colors"
        >
          <ArrowCircleDown size={20} weight="fill" />
          <span className="text-[13px] font-bold">{t('savings.deposit')}</span>
        </button>
        <button
          onClick={() => onHistory(goal.id)}
          className="w-12 h-12 rounded-2xl flex items-center justify-center bg-accent-subtle text-muted active:bg-accent/15 transition-colors shrink-0"
        >
          <ClockCounterClockwise size={20} weight="fill" />
        </button>
      </div>
    </ActionRow>
  )
}

/* ─── Create / Edit Form (bottom sheet) ─── */
function GoalFormSheet({ onClose, editGoal }: { onClose: () => void; editGoal?: SavingsGoal }) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const { notification } = useHaptic()
  const { code: currencyCode } = useBaseCurrency()

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
  const errorMsg = friendlyError(createMut.error || updateMut.error, t)
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
              className="w-full bg-bg rounded-2xl px-4 py-3 text-sm font-medium outline-none text-text placeholder:text-muted/50 transition-shadow focus:shadow-(--shadow-focus)"
            />
          </div>

          {/* Target amount */}
          <div>
            <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
              {t('savings.target')}
            </label>
            <div className="flex items-baseline gap-2 bg-bg rounded-2xl px-4 py-3 focus-within:shadow-(--shadow-focus) transition-shadow">
              <CurrencyBadge currency={currencyCode} className="text-muted/40" />
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
                {t('accounts')}
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
                ? 'bg-accent text-accent-text shadow-(--shadow-button)'
                : 'bg-border text-muted'
              }
            `}
          >
            {isPending ? t('common.loading') : isEdit ? t('common.save') : t('common.create')}
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
  const { code: amountCurrency } = useBaseCurrency()
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

        <div className="flex items-baseline gap-2 bg-bg rounded-2xl px-4 py-3 focus-within:shadow-(--shadow-focus) transition-shadow">
          <CurrencyBadge currency={amountCurrency} className="text-muted/40" />
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
                ? 'bg-income text-white shadow-(--shadow-income)'
                : 'bg-expense text-white shadow-(--shadow-expense)'
              : 'bg-border text-muted'
            }
          `}
        >
          {isPending ? t('common.loading') : t('common.confirm')}
        </button>

        {isError && (
          <p className="text-xs text-destructive text-center">{friendlyError(error, t)}</p>
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
      qc.invalidateQueries({ queryKey: ['accounts'] })
      qc.invalidateQueries({ queryKey: ['transactions'] })
      setAmountFor(null)
    },
    onError: () => notification('error'),
  })

  const withdrawMut = useMutation({
    mutationFn: ({ id, cents }: { id: number; cents: number }) => withdrawGoal(id, cents),
    onSuccess: () => {
      notification('success')
      qc.invalidateQueries({ queryKey: ['goals'] })
      qc.invalidateQueries({ queryKey: ['accounts'] })
      qc.invalidateQueries({ queryKey: ['transactions'] })
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
                    className="flex items-center gap-1.5 px-5 py-2.5 rounded-full bg-accent text-accent-text text-xs font-bold shadow-(--shadow-accent-pill) active:scale-95 transition-transform"
                  >
                    <Plus size={14} weight="bold" />
                    {t('savings.create_new')}
                  </button>
                }
              />
            </div>
          ) : (
            <div className="mx-4 space-y-3">
              {goals.map(goal => (
                <div key={goal.id} className="card-elevated overflow-hidden">
                  <GoalRow
                    goal={goal}
                    onEdit={g => { setShowForm(false); setAmountFor(null); setHistoryFor(null); setEditingGoal(g) }}
                    onDeposit={id => { setShowForm(false); setEditingGoal(null); setHistoryFor(null); setAmountFor({ id, action: 'deposit' }) }}
                    onWithdraw={id => { setShowForm(false); setEditingGoal(null); setHistoryFor(null); setAmountFor({ id, action: 'withdraw' }) }}
                    onHistory={id => { setShowForm(false); setEditingGoal(null); setAmountFor(null); setHistoryFor(id) }}
                    onDelete={id => deleteMut.mutate(id)}
                  />
                </div>
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
