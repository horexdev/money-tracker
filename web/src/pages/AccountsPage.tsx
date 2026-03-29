import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { AnimatePresence } from 'framer-motion'
import {
  Bank, PiggyBank, Money, CreditCard, Coins, Star, Plus,
} from '@phosphor-icons/react'
import { accountsApi } from '../api/accounts'

import { formatCents } from '../lib/money'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { useHaptic } from '../hooks/useHaptic'
import { EmptyState, SwipeToDelete, FAB, BottomSheet } from '../components/ui'
import type { Account, AccountType } from '../types'

/* ─── Constants ─── */

const ACCOUNT_TYPE_ICONS: Record<AccountType, React.ComponentType<{ size?: number; weight?: 'fill' | 'regular' | 'bold' | 'duotone' | 'light' | 'thin'; className?: string }>> = {
  checking: Bank,
  savings: PiggyBank,
  cash: Money,
  credit: CreditCard,
  crypto: Coins,
}

const COLOR_SWATCHES = [
  '#6366f1', '#8b5cf6', '#ec4899', '#ef4444',
  '#f97316', '#eab308', '#22c55e', '#10b981',
  '#14b8a6', '#06b6d4', '#3b82f6', '#64748b',
]

const ACCOUNT_TYPES: AccountType[] = ['checking', 'savings', 'cash', 'credit', 'crypto']

const CURRENCY_OPTIONS = [
  'USD', 'EUR', 'GBP', 'RUB', 'UAH', 'BYN', 'KZT', 'UZS',
  'TRY', 'CNY', 'JPY', 'CHF', 'CAD', 'AUD', 'PLN', 'CZK',
]

/* ─── Color Picker ─── */
function ColorPicker({ selected, onSelect }: { selected: string; onSelect: (c: string) => void }) {
  const { selection } = useHaptic()
  return (
    <div className="grid grid-cols-6 gap-2">
      {COLOR_SWATCHES.map((c) => (
        <button
          key={c}
          type="button"
          onClick={() => { onSelect(c); selection() }}
          className="h-10 rounded-2xl transition-all duration-150 active:scale-90 relative flex items-center justify-center"
          style={{ background: c }}
        >
          {selected === c && (
            <span className="w-3 h-3 rounded-full border-2 border-white bg-white/40 block" />
          )}
        </button>
      ))}
    </div>
  )
}

/* ─── Account Form Sheet ─── */
function AccountFormSheet({
  onClose,
  editAccount,
  accountCount,
}: {
  onClose: () => void
  editAccount?: Account
  accountCount: number
}) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const { notification } = useHaptic()

  const isEdit = editAccount !== undefined
  const defaultColor = editAccount?.color ?? COLOR_SWATCHES[accountCount % COLOR_SWATCHES.length]

  const [name, setName] = useState(editAccount?.name ?? '')
  const [color, setColor] = useState(defaultColor)
  const [accountType, setAccountType] = useState<AccountType>(editAccount?.type ?? 'checking')
  const [currency, setCurrency] = useState(editAccount?.currency_code ?? 'USD')
  const [includeInTotal, setIncludeInTotal] = useState(editAccount?.include_in_total ?? true)

  const createMut = useMutation({
    mutationFn: () => accountsApi.create({
      name: name.trim(),
      icon: accountType,
      color,
      type: accountType,
      currency_code: currency,
      include_in_total: includeInTotal,
    }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['accounts'] })
      notification('success')
      onClose()
    },
    onError: () => notification('error'),
  })

  const updateMut = useMutation({
    mutationFn: () => accountsApi.update(editAccount!.id, {
      name: name.trim(),
      icon: accountType,
      color,
      type: accountType,
      currency_code: currency,
      include_in_total: includeInTotal,
    }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['accounts'] })
      notification('success')
      onClose()
    },
    onError: () => notification('error'),
  })

  const isPending = createMut.isPending || updateMut.isPending
  const isError = createMut.isError || updateMut.isError
  const errorMsg = ((createMut.error || updateMut.error) as Error | null)?.message
  const canSubmit = name.trim().length > 0 && !isPending

  return (
    <BottomSheet onClose={onClose}>
      <div
        className="px-5 space-y-4 overflow-y-auto no-scrollbar"
        style={{ maxHeight: '85dvh', paddingBottom: 'max(1.5rem, env(safe-area-inset-bottom))' }}
      >
        {/* Preview + name */}
        <div className="flex gap-3 items-end">
          <div
            className="w-12 h-12 rounded-2xl flex items-center justify-center shrink-0"
            style={{ background: color, boxShadow: `0 2px 8px ${color}66` }}
          >
            {(() => { const PreviewIcon = ACCOUNT_TYPE_ICONS[accountType]; return <PreviewIcon size={22} weight="fill" className="text-white" /> })()}
          </div>
          <div className="flex-1">
            <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
              {t('categories.name')}
            </label>
            <input
              type="text"
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder={t('newAccount')}
              maxLength={40}
              autoFocus
              className="w-full bg-bg rounded-2xl px-4 py-3 text-sm font-medium outline-none text-text placeholder:text-muted/50 transition-shadow focus:shadow-[0_0_0_2px_rgba(99,102,241,0.2)]"
            />
          </div>
        </div>

        {/* Account type */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('accountType')}
          </label>
          <div className="flex gap-1.5 flex-wrap">
            {ACCOUNT_TYPES.map((at) => {
              const TypeIcon = ACCOUNT_TYPE_ICONS[at]
              const isActive = at === accountType
              return (
                <button
                  key={at}
                  type="button"
                  onClick={() => setAccountType(at)}
                  className={`
                    flex items-center gap-1.5 px-3 py-2 rounded-2xl text-[12px] font-bold transition-all duration-150 active:scale-95
                    ${isActive
                      ? 'bg-accent text-accent-text shadow-[0_2px_8px_rgba(99,102,241,0.3)]'
                      : 'bg-accent-subtle text-muted'
                    }
                  `}
                >
                  <TypeIcon size={14} weight="fill" />
                  {t(`accountTypes.${at}`)}
                </button>
              )
            })}
          </div>
        </div>

        {/* Currency */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('add.select_currency')}
          </label>
          <div className="flex flex-wrap gap-1.5">
            {CURRENCY_OPTIONS.map((c) => (
              <button
                key={c}
                type="button"
                onClick={() => setCurrency(c)}
                className={`
                  px-3 py-1.5 rounded-2xl text-[12px] font-bold transition-all duration-150 active:scale-95
                  ${currency === c
                    ? 'bg-accent text-accent-text shadow-[0_2px_8px_rgba(99,102,241,0.3)]'
                    : 'bg-accent-subtle text-muted'
                  }
                `}
              >
                {c}
              </button>
            ))}
          </div>
        </div>

        {/* Color picker */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('categories.color')}
          </label>
          <ColorPicker selected={color} onSelect={setColor} />
        </div>

        {/* Include in total toggle */}
        <button
          type="button"
          onClick={() => setIncludeInTotal(v => !v)}
          className="w-full flex items-center justify-between bg-bg rounded-2xl px-4 py-3 active:bg-accent/5 transition-colors"
        >
          <span className="text-sm font-semibold text-text">{t('includeInTotal')}</span>
          <div
            className={`w-11 h-6 rounded-full transition-colors duration-200 relative ${includeInTotal ? 'bg-accent' : 'bg-border'}`}
          >
            <div
              className={`absolute top-0.5 w-5 h-5 rounded-full bg-white shadow transition-transform duration-200 ${includeInTotal ? 'translate-x-5' : 'translate-x-0.5'}`}
            />
          </div>
        </button>

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
  )
}

/* ─── Account Row ─── */
function AccountRow({
  account,
  onEdit,
  onDelete,
  onSetDefault,
  isDeleting,
}: {
  account: Account
  onEdit: (a: Account) => void
  onDelete: (id: number) => void
  onSetDefault: (id: number) => void
  isDeleting: boolean
}) {
  const { t } = useTranslation()
  const TypeIcon = ACCOUNT_TYPE_ICONS[account.type]

  return (
    <div className={`transition-opacity ${isDeleting ? 'opacity-30 pointer-events-none' : ''}`}>
      <SwipeToDelete onDelete={() => onDelete(account.id)}>
        <div className="flex items-center gap-3 px-4 py-3">
          {/* Icon */}
          <div
            className="w-10 h-10 rounded-2xl flex items-center justify-center shrink-0"
            style={{ background: account.color || 'var(--color-accent)' }}
          >
            <TypeIcon size={20} weight="fill" className="text-white" />
          </div>

          {/* Info */}
          <button
            onClick={() => onEdit(account)}
            className="flex-1 min-w-0 text-left active:opacity-70 transition-opacity"
          >
            <div className="flex items-center gap-1.5 mb-0.5">
              <span className="text-[13px] font-bold text-text truncate">{account.name}</span>
              {account.is_default && (
                <span className="text-[9px] font-bold text-accent bg-accent/10 rounded-full px-1.5 py-0.5 leading-none shrink-0">
                  {t('defaultAccount')}
                </span>
              )}
            </div>
            <div className="flex items-center gap-1 text-xs text-muted">
              <TypeIcon size={11} weight="fill" />
              <span className="capitalize">{t(`accountTypes.${account.type}`)}</span>
              <span className="text-muted/40">·</span>
              <span className="font-semibold text-text tabular-nums">
                {formatCents(account.balance_cents, account.currency_code)}
              </span>
            </div>
          </button>

          {/* Set default star */}
          {!account.is_default && (
            <button
              onClick={() => onSetDefault(account.id)}
              className="w-9 h-9 rounded-2xl flex items-center justify-center text-muted active:text-accent active:bg-accent-subtle transition-colors shrink-0"
            >
              <Star size={18} weight="regular" />
            </button>
          )}
        </div>
      </SwipeToDelete>
    </div>
  )
}

/* ─── Main Page ─── */
export function AccountsPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { notification } = useHaptic()
  useTgBackButton(() => navigate('/more'))

  const [editingAccount, setEditingAccount] = useState<Account | null>(null)
  const [showCreate, setShowCreate] = useState(false)
  const [deletingId, setDeletingId] = useState<number | null>(null)

  const { data: accounts = [], isLoading, isError, refetch } = useQuery({
    queryKey: ['accounts'],
    queryFn: () => accountsApi.list(),
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => accountsApi.delete(id),
    onMutate: (id) => setDeletingId(id),
    onSettled: () => setDeletingId(null),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['accounts'] })
      notification('success')
    },
    onError: () => notification('error'),
  })

  const setDefaultMut = useMutation({
    mutationFn: (id: number) => accountsApi.setDefault(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['accounts'] })
      notification('success')
    },
    onError: () => notification('error'),
  })

  const formOpen = showCreate || editingAccount !== null

  if (isLoading) return <div className="flex justify-center py-16"><Spinner /></div>
  if (isError) return <ErrorMessage onRetry={refetch} />

  return (
    <PageTransition>
      <div className="pt-3 pb-4">
          {accounts.length === 0 ? (
            <div className="mx-4 card-elevated mt-2">
              <EmptyState
                icon={Bank}
                title={t('accounts')}
                description={t('newAccount')}
                action={
                  <button
                    onClick={() => { setEditingAccount(null); setShowCreate(true) }}
                    className="flex items-center gap-1.5 px-5 py-2.5 rounded-full bg-accent text-accent-text text-xs font-bold shadow-[0_2px_12px_rgba(99,102,241,0.4)] active:scale-95 transition-transform"
                  >
                    <Plus size={14} weight="bold" />
                    {t('newAccount')}
                  </button>
                }
              />
            </div>
          ) : (
            <div className="mx-4 card-elevated divide-y divide-border">
              {accounts.map((account) => (
                <AccountRow
                  key={account.id}
                  account={account}
                  onEdit={setEditingAccount}
                  onDelete={id => deleteMut.mutate(id)}
                  onSetDefault={id => setDefaultMut.mutate(id)}
                  isDeleting={deletingId === account.id}
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

      <FAB onClick={() => { setEditingAccount(null); setShowCreate(true) }} label={t('newAccount')} />

      <AnimatePresence>
        {formOpen && (
          <AccountFormSheet
            key={editingAccount?.id ?? 'new'}
            editAccount={editingAccount ?? undefined}
            accountCount={accounts.length}
            onClose={() => { setShowCreate(false); setEditingAccount(null) }}
          />
        )}
      </AnimatePresence>
    </PageTransition>
  )
}
