import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { AnimatePresence } from 'framer-motion'
import { ArrowRight, ArrowsHorizontal } from '@phosphor-icons/react'
import { transfersApi } from '../../shared/api/transfers'
import { accountsApi } from '../../shared/api/accounts'
import { formatCents, formatDate, parseCents, sanitizeAmount } from '../../shared/lib/money'
import { friendlyError } from '../../shared/lib/errors'
import { CurrencyBadge } from '../../shared/lib/currencyIcons'
import { Spinner } from '../../shared/ui/Spinner'
import { ErrorMessage } from '../../shared/ui/ErrorMessage'
import { PageTransition } from '../../shared/ui/PageTransition'
import { useTgBackButton } from '../../shared/hooks/useTelegramApp'
import { useHaptic } from '../../shared/hooks/useHaptic'
import { EmptyState, SwipeToDelete, FAB, BottomSheet } from '../../shared/ui'
import type { Transfer, Account } from '../../shared/types'

/* ─── Transfer Row ─── */
function TransferRow({
  transfer,
  onDelete,
  isDeleting,
}: {
  transfer: Transfer
  onDelete: (id: number) => void
  isDeleting: boolean
}) {
  const { i18n } = useTranslation()

  return (
    <div className={`transition-opacity ${isDeleting ? 'opacity-30 pointer-events-none' : ''}`}>
      <SwipeToDelete onDelete={() => onDelete(transfer.id)}>
        <div className="flex items-center gap-3 px-4 py-3">
          {/* Icon */}
          <div className="w-10 h-10 rounded-2xl flex items-center justify-center shrink-0 bg-accent/10">
            <ArrowRight size={20} weight="bold" className="text-accent" />
          </div>

          {/* From → To */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-1.5 mb-0.5 text-[13px] font-bold text-text">
              <span className="truncate max-w-[6rem]">{transfer.from_account_name}</span>
              <ArrowRight size={12} weight="bold" className="text-muted shrink-0" />
              <span className="truncate max-w-[6rem]">{transfer.to_account_name}</span>
            </div>
            <div className="flex items-center gap-1 text-xs text-muted">
              <span>{formatDate(transfer.created_at, i18n.language)}</span>
              {transfer.note && (
                <>
                  <span className="text-muted/40">·</span>
                  <span className="truncate">{transfer.note}</span>
                </>
              )}
            </div>
          </div>

          {/* Amount */}
          <div className="text-right shrink-0">
            <span className="text-[13px] font-bold text-text tabular-nums">
              {formatCents(transfer.amount_cents, transfer.from_currency_code)}
            </span>
            {transfer.from_currency_code !== transfer.to_currency_code && (
              <p className="text-[10px] text-muted">
                ≈ {formatCents(Math.round(transfer.amount_cents * transfer.exchange_rate), transfer.to_currency_code)}
              </p>
            )}
          </div>
        </div>
      </SwipeToDelete>
    </div>
  )
}

/* ─── Create Transfer Sheet ─── */
function TransferFormSheet({
  onClose,
  accounts,
}: {
  onClose: () => void
  accounts: Account[]
}) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const { notification } = useHaptic()

  const [fromId, setFromId] = useState<number>(accounts[0]?.id ?? 0)
  const [toId, setToId] = useState<number>(accounts[1]?.id ?? accounts[0]?.id ?? 0)
  const [amountStr, setAmountStr] = useState('')
  const [exchangeRateStr, setExchangeRateStr] = useState('')
  const [note, setNote] = useState('')

  const fromAccount = accounts.find(a => a.id === fromId)
  const toAccount = accounts.find(a => a.id === toId)
  const diffCurrency = fromAccount?.currency_code !== toAccount?.currency_code

  const amountCents = parseCents(amountStr)
  const exchangeRate = parseFloat(exchangeRateStr) || 1

  const createMut = useMutation({
    mutationFn: () => transfersApi.create({
      from_account_id: fromId,
      to_account_id: toId,
      amount_cents: amountCents,
      exchange_rate: diffCurrency ? exchangeRate : undefined,
      note: note.trim() || undefined,
    }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['transfers'] })
      qc.invalidateQueries({ queryKey: ['accounts'] })
      notification('success')
      onClose()
    },
    onError: () => notification('error'),
  })

  const canSubmit = fromId > 0 && toId > 0 && fromId !== toId && amountCents > 0 && !createMut.isPending

  return (
    <BottomSheet onClose={onClose}>
      <div
        className="px-5 space-y-4 overflow-y-auto no-scrollbar"
        style={{ maxHeight: '85dvh', paddingBottom: 'max(1.5rem, env(safe-area-inset-bottom))' }}
      >
        {/* From account */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('from')}
          </label>
          <div className="flex flex-wrap gap-1.5">
            {accounts.map((a) => (
              <button
                key={a.id}
                type="button"
                onClick={() => setFromId(a.id)}
                className={`
                  px-3 py-2 rounded-2xl text-[12px] font-bold transition-all duration-150 active:scale-95
                  ${fromId === a.id
                    ? 'bg-accent text-accent-text shadow-(--shadow-accent-pill)'
                    : 'bg-accent-subtle text-muted'
                  }
                `}
              >
                {a.name}
              </button>
            ))}
          </div>
        </div>

        {/* To account */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('to')}
          </label>
          <div className="flex flex-wrap gap-1.5">
            {accounts.map((a) => (
              <button
                key={a.id}
                type="button"
                onClick={() => setToId(a.id)}
                disabled={a.id === fromId}
                className={`
                  px-3 py-2 rounded-2xl text-[12px] font-bold transition-all duration-150 active:scale-95
                  ${toId === a.id
                    ? 'bg-accent text-accent-text shadow-(--shadow-accent-pill)'
                    : a.id === fromId
                      ? 'bg-border text-muted/40 cursor-not-allowed'
                      : 'bg-accent-subtle text-muted'
                  }
                `}
              >
                {a.name}
              </button>
            ))}
          </div>
        </div>

        {/* Amount */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('transactions.amount')}
          </label>
          <div className="flex items-baseline gap-2 bg-bg rounded-2xl px-4 py-3 focus-within:shadow-(--shadow-focus) transition-shadow">
            {fromAccount && <CurrencyBadge currency={fromAccount.currency_code} className="text-muted/40" />}
            <input
              inputMode="decimal"
              placeholder="0.00"
              value={amountStr}
              onChange={e => setAmountStr(sanitizeAmount(e.target.value))}
              className="flex-1 bg-transparent text-3xl font-bold outline-none text-text placeholder:text-muted/20 tabular-nums min-w-0"
            />
          </div>
        </div>

        {/* Exchange rate (only when currencies differ) */}
        {diffCurrency && (
          <div>
            <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
              {t('exchangeRate')}
            </label>
            <div className="flex items-center gap-2 bg-bg rounded-2xl px-4 py-3 focus-within:shadow-(--shadow-focus) transition-shadow">
              <span className="text-sm font-semibold text-muted">
                1 {fromAccount?.currency_code} =
              </span>
              <input
                inputMode="decimal"
                placeholder="1.00"
                value={exchangeRateStr}
                onChange={e => setExchangeRateStr(sanitizeAmount(e.target.value))}
                className="flex-1 bg-transparent text-sm font-bold outline-none text-text placeholder:text-muted/20 tabular-nums min-w-0"
              />
              <span className="text-sm font-semibold text-muted">
                {toAccount?.currency_code}
              </span>
            </div>
          </div>
        )}

        {/* Note */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('transactions.note')}
          </label>
          <input
            type="text"
            value={note}
            onChange={e => setNote(e.target.value)}
            placeholder={t('transactions.note_placeholder')}
            maxLength={100}
            className="w-full bg-bg rounded-2xl px-4 py-3 text-sm font-medium outline-none text-text placeholder:text-muted/50 transition-shadow focus:shadow-(--shadow-focus)"
          />
        </div>

        {/* Submit */}
        <button
          onClick={() => createMut.mutate()}
          disabled={!canSubmit}
          className={`
            w-full py-4 rounded-2xl text-[15px] font-bold transition-all active:scale-[0.98]
            ${canSubmit
              ? 'bg-accent text-accent-text shadow-(--shadow-button)'
              : 'bg-border text-muted'
            }
          `}
        >
          {createMut.isPending ? t('common.loading') : t('common.create')}
        </button>

        {createMut.isError && (
          <p className="text-xs text-destructive text-center">
            {friendlyError(createMut.error, t)}
          </p>
        )}
      </div>
    </BottomSheet>
  )
}

/* ─── Main Page ─── */
export function TransfersPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { notification } = useHaptic()
  useTgBackButton(() => navigate('/more'))

  const [showCreate, setShowCreate] = useState(false)
  const [deletingId, setDeletingId] = useState<number | null>(null)

  const { data: transfersData, isLoading, isError, refetch } = useQuery({
    queryKey: ['transfers'],
    queryFn: () => transfersApi.list({ limit: 50 }),
  })

  const { data: accounts = [] } = useQuery({
    queryKey: ['accounts'],
    queryFn: () => accountsApi.list(),
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => transfersApi.delete(id),
    onMutate: (id) => setDeletingId(id),
    onSettled: () => setDeletingId(null),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['transfers'] })
      qc.invalidateQueries({ queryKey: ['accounts'] })
      notification('success')
    },
    onError: () => notification('error'),
  })

  const transfers = transfersData?.transfers ?? []

  if (isLoading) return <div className="flex justify-center py-16"><Spinner /></div>
  if (isError) return <ErrorMessage onRetry={refetch} />

  return (
    <PageTransition>
      <div className="pt-3 pb-4">
          {transfers.length === 0 ? (
            <div className="mx-4 card-elevated mt-2">
              <EmptyState
                icon={ArrowsHorizontal}
                title={t('transfers')}
                description={t('newTransfer')}
                action={
                  accounts.length >= 2 ? (
                    <button
                      onClick={() => setShowCreate(true)}
                      className="flex items-center gap-1.5 px-5 py-2.5 rounded-full bg-accent text-accent-text text-xs font-bold shadow-(--shadow-accent-pill) active:scale-95 transition-transform"
                    >
                      {t('newTransfer')}
                    </button>
                  ) : undefined
                }
              />
            </div>
          ) : (
            <div className="mx-4 card-elevated divide-y divide-border">
              {transfers.map((transfer: Transfer) => (
                <TransferRow
                  key={transfer.id}
                  transfer={transfer}
                  onDelete={id => deleteMut.mutate(id)}
                  isDeleting={deletingId === transfer.id}
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

      {accounts.length >= 2 && (
        <FAB onClick={() => setShowCreate(true)} label={t('newTransfer')} />
      )}

      <AnimatePresence>
        {showCreate && (
          <TransferFormSheet
            accounts={accounts}
            onClose={() => setShowCreate(false)}
          />
        )}
      </AnimatePresence>
    </PageTransition>
  )
}
