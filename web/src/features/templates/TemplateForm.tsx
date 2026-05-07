import { useState, useEffect } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { templatesApi } from '../../shared/api/templates'
import { categoriesApi } from '../../shared/api/categories'
import { accountsApi } from '../../shared/api/accounts'
import { parseCents } from '../../shared/lib/money'
import { friendlyError } from '../../shared/lib/errors'
import { CategoryIcon } from '../../shared/lib/categoryIcons'
import { AmountInput } from '../../shared/ui/AmountInput'
import { BottomSheet } from '../../shared/ui'
import { AccountDropdown } from '../../shared/ui/AccountDropdown'
import { useHaptic } from '../../shared/hooks/useHaptic'
import { useCategoryName } from '../../shared/hooks/useCategoryName'
import { useBaseCurrency } from '../../shared/hooks/useBaseCurrency'
import type { TransactionTemplate, TransactionType } from '../../shared/types'

interface TemplateFormProps {
  editItem: TransactionTemplate | null
  initialState?: PrefillState
  onClose: () => void
  onSaved?: (tpl: TransactionTemplate) => void
}

// PrefillState lets callers (e.g. AddTransaction's "Save as template") pass
// a starting set of values without committing to "edit" semantics.
export interface PrefillState {
  type?: TransactionType
  amountCents?: number
  categoryID?: number | null
  accountID?: number | null
  note?: string
}

export function TemplateForm({ editItem, initialState, onClose, onSaved }: TemplateFormProps) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const { notification } = useHaptic()
  const tCategory = useCategoryName()
  const { code: baseCurrency } = useBaseCurrency()

  const isEdit = editItem !== null

  const [name, setName] = useState(editItem?.name ?? '')
  const [type, setType] = useState<TransactionType>(editItem?.type ?? initialState?.type ?? 'expense')
  const [amount, setAmount] = useState(
    editItem ? String(editItem.amount_cents / 100) :
    initialState?.amountCents ? String(initialState.amountCents / 100) : ''
  )
  const [amountFixed, setAmountFixed] = useState(editItem?.amount_fixed ?? true)
  const [categoryID, setCategoryID] = useState<number | null>(editItem?.category_id ?? initialState?.categoryID ?? null)
  const [note, setNote] = useState(editItem?.note ?? initialState?.note ?? '')
  const [accountID, setAccountID] = useState<number | null>(editItem?.account_id ?? initialState?.accountID ?? null)

  const { data: accounts = [] } = useQuery({ queryKey: ['accounts'], queryFn: accountsApi.list })
  const catsQ = useQuery({ queryKey: ['categories', { order: 'frequency' }], queryFn: () => categoriesApi.list(undefined, 'frequency') })
  const categories = catsQ.data?.categories ?? []
  const filtered = categories.filter(c => c.type === type || c.type === 'both')

  useEffect(() => {
    if (accounts.length === 0) return
    if (accountID !== null) return
    const def = accounts.find(a => a.is_default) ?? accounts[0]
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setAccountID(def.id)
  }, [accounts]) // eslint-disable-line react-hooks/exhaustive-deps

  const selectedAccount = accounts.find(a => a.id === accountID)
  const currencyCode = selectedAccount?.currency_code ?? baseCurrency

  const createMut = useMutation({
    mutationFn: () => templatesApi.create({
      name: name.trim(),
      type,
      amount_cents: parseCents(amount),
      amount_fixed: amountFixed,
      category_id: categoryID!,
      account_id: accountID!,
      currency_code: currencyCode,
      note: note.trim(),
    }),
    onSuccess: (tpl) => {
      qc.invalidateQueries({ queryKey: ['templates'] })
      notification('success')
      onSaved?.(tpl)
      onClose()
    },
  })

  const updateMut = useMutation({
    mutationFn: () => templatesApi.update(editItem!.id, {
      name: name.trim(),
      type,
      amount_cents: parseCents(amount),
      amount_fixed: amountFixed,
      category_id: categoryID!,
      account_id: accountID!,
      currency_code: currencyCode,
      note: note.trim(),
    }),
    onSuccess: (tpl) => {
      qc.invalidateQueries({ queryKey: ['templates'] })
      notification('success')
      onSaved?.(tpl)
      onClose()
    },
  })

  const mut = isEdit ? updateMut : createMut
  const canSubmit = parseCents(amount) > 0 && categoryID !== null && accountID !== null && !mut.isPending

  return (
    <BottomSheet onClose={onClose}>
      <div
        className="px-5 space-y-4 overflow-y-auto no-scrollbar"
        style={{ maxHeight: '85dvh', paddingBottom: 'max(1.5rem, env(safe-area-inset-bottom))' }}
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

        {/* Amount fixed toggle */}
        <button
          onClick={() => setAmountFixed(v => !v)}
          className="w-full flex items-center justify-between bg-bg rounded-2xl px-4 py-3 active:scale-[0.99] transition-transform"
        >
          <div className="text-left">
            <p className="text-sm font-semibold text-text">{t('templates.amount_fixed')}</p>
            <p className="text-[11px] text-muted mt-0.5">
              {amountFixed ? t('templates.amount_fixed_hint') : t('templates.amount_variable_hint')}
            </p>
          </div>
          <div className={`w-10 h-6 rounded-full p-0.5 transition-colors ${amountFixed ? 'bg-accent' : 'bg-border'}`}>
            <div className={`w-5 h-5 rounded-full bg-white shadow transition-transform ${amountFixed ? 'translate-x-4' : ''}`} />
          </div>
        </button>

        {/* Name */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('templates.name')}
          </label>
          <input
            type="text"
            placeholder={t('templates.name_placeholder')}
            value={name}
            onChange={e => setName(e.target.value)}
            maxLength={50}
            className="w-full bg-bg rounded-2xl px-4 py-3 text-sm font-medium outline-none text-text placeholder:text-muted/50 transition-shadow focus:shadow-(--shadow-focus)"
          />
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
