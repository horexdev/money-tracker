import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Plus, Trash2, ArrowDownCircle, ArrowUpCircle, CheckCircle2 } from 'lucide-react'
import { fetchGoals, createGoal, depositGoal, withdrawGoal, deleteGoal } from '../api/goals'
import { formatCents, parseCents } from '../lib/money'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { useHaptic } from '../hooks/useHaptic'
import { Button, EmptyState } from '../components/ui'
import type { SavingsGoal } from '../types'

function GoalCard({
  goal,
  onDeposit,
  onWithdraw,
  onDelete,
}: {
  goal: SavingsGoal
  onDeposit: (id: number) => void
  onWithdraw: (id: number) => void
  onDelete: (id: number) => void
}) {
  const { t } = useTranslation()
  const pct   = Math.min(goal.progress_percent, 100)
  const color = goal.is_completed ? 'var(--color-income)' : 'var(--color-accent)'

  return (
    <div className="bg-surface rounded-[--radius-card] p-4 space-y-3">
      <div className="flex items-start justify-between">
        <div>
          <div className="flex items-center gap-2">
            <span className="text-sm font-bold text-text">{goal.name}</span>
            {goal.is_completed && <CheckCircle2 size={16} className="text-income" />}
          </div>
          {goal.deadline && (
            <p className="text-xs text-muted mt-0.5">
              {t('savings.deadline')}: {new Date(goal.deadline).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })}
            </p>
          )}
        </div>
        <button onClick={() => onDelete(goal.id)} className="p-1 text-muted active:text-destructive">
          <Trash2 size={14} />
        </button>
      </div>

      <div className="flex items-center gap-4">
        {/* Circular progress */}
        <div className="relative w-16 h-16 shrink-0">
          <svg viewBox="0 0 36 36" className="w-full h-full -rotate-90">
            <circle cx="18" cy="18" r="15.5" fill="none" stroke="var(--color-border)" strokeWidth="3" />
            <circle
              cx="18" cy="18" r="15.5" fill="none"
              stroke={color}
              strokeWidth="3"
              strokeDasharray={`${pct * 0.975} 100`}
              strokeLinecap="round"
              className="transition-all duration-700"
            />
          </svg>
          <div className="absolute inset-0 flex items-center justify-center">
            <span className="text-xs font-bold" style={{ color }}>{pct.toFixed(0)}%</span>
          </div>
        </div>

        <div className="flex-1 space-y-1 text-xs">
          <div className="flex justify-between">
            <span className="text-muted">{t('savings.current')}</span>
            <span className="font-semibold tabular-nums">{formatCents(goal.current_cents, goal.currency_code)}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted">{t('savings.target')}</span>
            <span className="tabular-nums">{formatCents(goal.target_cents, goal.currency_code)}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted">{t('savings.remaining')}</span>
            <span className="tabular-nums">{formatCents(goal.remaining_cents, goal.currency_code)}</span>
          </div>
        </div>
      </div>

      <div className="flex gap-2 pt-1">
        <Button size="sm" className="flex-1" onClick={() => onDeposit(goal.id)}>
          <ArrowDownCircle size={14} className="mr-1" /> {t('savings.deposit')}
        </Button>
        <Button size="sm" variant="secondary" className="flex-1" onClick={() => onWithdraw(goal.id)}>
          <ArrowUpCircle size={14} className="mr-1" /> {t('savings.withdraw')}
        </Button>
      </div>
    </div>
  )
}

export function SavingsPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { notification } = useHaptic()
  useTgBackButton(() => navigate('/more'))

  const [showForm, setShowForm]           = useState(false)
  const [showAmountFor, setShowAmountFor] = useState<{ id: number; action: 'deposit' | 'withdraw' } | null>(null)
  const [amountStr, setAmountStr]         = useState('')
  const [goalName, setGoalName]           = useState('')
  const [targetStr, setTargetStr]         = useState('')
  const [deadline, setDeadline]           = useState('')

  const goalsQ = useQuery({ queryKey: ['goals'], queryFn: fetchGoals })

  const createMut = useMutation({
    mutationFn: () => createGoal({
      name: goalName.trim(), target_cents: parseCents(targetStr),
      currency_code: 'USD', deadline: deadline || undefined,
    }),
    onSuccess: () => {
      notification('success')
      qc.invalidateQueries({ queryKey: ['goals'] })
      setShowForm(false); setGoalName(''); setTargetStr(''); setDeadline('')
    },
    onError: () => notification('error'),
  })

  const depositMut = useMutation({
    mutationFn: ({ id, cents }: { id: number; cents: number }) => depositGoal(id, cents),
    onSuccess: () => {
      notification('success')
      qc.invalidateQueries({ queryKey: ['goals'] })
      setShowAmountFor(null); setAmountStr('')
    },
    onError: () => notification('error'),
  })

  const withdrawMut = useMutation({
    mutationFn: ({ id, cents }: { id: number; cents: number }) => withdrawGoal(id, cents),
    onSuccess: () => {
      notification('success')
      qc.invalidateQueries({ queryKey: ['goals'] })
      setShowAmountFor(null); setAmountStr('')
    },
    onError: () => notification('error'),
  })

  const deleteMut = useMutation({
    mutationFn: deleteGoal,
    onSuccess: () => { notification('success'); qc.invalidateQueries({ queryKey: ['goals'] }) },
  })

  function handleAmountSubmit() {
    if (!showAmountFor) return
    const cents = parseCents(amountStr)
    if (cents <= 0) return
    if (showAmountFor.action === 'deposit') {
      depositMut.mutate({ id: showAmountFor.id, cents })
    } else {
      withdrawMut.mutate({ id: showAmountFor.id, cents })
    }
  }

  const goals          = goalsQ.data?.goals ?? []
  const canCreateGoal  = goalName.trim() && parseCents(targetStr) > 0 && !createMut.isPending

  if (goalsQ.isPending) return <div className="flex justify-center py-16"><Spinner /></div>
  if (goalsQ.isError)   return <ErrorMessage onRetry={() => goalsQ.refetch()} />

  return (
    <PageTransition>
      <div className="p-4 space-y-4">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-bold">{t('savings.title')}</h1>
          {!showForm && (
            <Button size="sm" onClick={() => setShowForm(true)}>
              <Plus size={15} className="mr-1" /> {t('savings.create_new')}
            </Button>
          )}
        </div>

        {/* Create form */}
        {showForm && (
          <div className="bg-surface rounded-[--radius-card] p-4 space-y-4">
            <div>
              <label className="block text-xs font-semibold text-muted uppercase tracking-widest mb-2">
                {t('categories.name')}
              </label>
              <input
                type="text"
                value={goalName}
                onChange={e => setGoalName(e.target.value)}
                placeholder={t('savings.create_new')}
                maxLength={50}
                className="w-full bg-bg rounded-[--radius-sm] px-3 py-2.5 text-sm outline-none focus:ring-2 focus:ring-accent"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-muted uppercase tracking-widest mb-2">
                {t('savings.target')}
              </label>
              <input
                inputMode="decimal"
                placeholder="0.00"
                value={targetStr}
                onChange={e => setTargetStr(e.target.value)}
                className="w-full bg-bg rounded-[--radius-sm] px-3 py-2.5 text-2xl font-bold outline-none focus:ring-2 focus:ring-accent tabular-nums"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-muted uppercase tracking-widest mb-2">
                {t('savings.deadline')} (optional)
              </label>
              <input
                type="date"
                value={deadline}
                onChange={e => setDeadline(e.target.value)}
                className="w-full bg-bg rounded-[--radius-sm] px-3 py-2.5 text-sm outline-none focus:ring-2 focus:ring-accent"
              />
            </div>
            <div className="flex gap-2 pt-1">
              <Button size="sm" onClick={() => createMut.mutate()} disabled={!canCreateGoal}>
                {t('common.create')}
              </Button>
              <Button size="sm" variant="ghost" onClick={() => { setShowForm(false); setGoalName(''); setTargetStr(''); setDeadline('') }}>
                {t('common.cancel')}
              </Button>
            </div>
            {createMut.isError && (
              <p className="text-xs text-destructive">{(createMut.error as Error)?.message}</p>
            )}
          </div>
        )}

        {/* Deposit/Withdraw overlay */}
        {showAmountFor && (
          <div className="bg-surface rounded-[--radius-card] p-4 space-y-4">
            <label className="block text-xs font-semibold text-muted uppercase tracking-widest">
              {showAmountFor.action === 'deposit' ? t('savings.deposit') : t('savings.withdraw')}
            </label>
            <input
              inputMode="decimal"
              placeholder="0.00"
              value={amountStr}
              onChange={e => setAmountStr(e.target.value)}
              autoFocus
              className="w-full bg-bg rounded-[--radius-sm] px-3 py-2.5 text-2xl font-bold outline-none focus:ring-2 focus:ring-accent tabular-nums"
            />
            <div className="flex gap-2">
              <Button size="sm" onClick={handleAmountSubmit} disabled={parseCents(amountStr) <= 0}>
                {t('common.confirm')}
              </Button>
              <Button size="sm" variant="ghost" onClick={() => { setShowAmountFor(null); setAmountStr('') }}>
                {t('common.cancel')}
              </Button>
            </div>
            {(depositMut.isError || withdrawMut.isError) && (
              <p className="text-xs text-destructive">
                {((depositMut.error || withdrawMut.error) as Error)?.message}
              </p>
            )}
          </div>
        )}

        {goals.length > 0 ? (
          <div className="space-y-3">
            {goals.map(goal => (
              <GoalCard
                key={goal.id}
                goal={goal}
                onDeposit={id => { setShowAmountFor({ id, action: 'deposit' }); setAmountStr('') }}
                onWithdraw={id => { setShowAmountFor({ id, action: 'withdraw' }); setAmountStr('') }}
                onDelete={id => deleteMut.mutate(id)}
              />
            ))}
          </div>
        ) : !showForm ? (
          <EmptyState icon="🎯" title={t('savings.no_goals')} description={t('savings.start_saving')} />
        ) : null}
      </div>
    </PageTransition>
  )
}
