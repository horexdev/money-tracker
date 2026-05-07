import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { BottomSheet } from '../../shared/ui'
import { AmountInput } from '../../shared/ui/AmountInput'
import { CategoryIcon } from '../../shared/lib/categoryIcons'
import { parseCents } from '../../shared/lib/money'
import { useCategoryName } from '../../shared/hooks/useCategoryName'
import type { TransactionTemplate } from '../../shared/types'

interface QuickAmountModalProps {
  template: TransactionTemplate
  onSubmit: (amountCents: number) => void
  onClose: () => void
  isPending: boolean
}

export function QuickAmountModal({ template, onSubmit, onClose, isPending }: QuickAmountModalProps) {
  const { t } = useTranslation()
  const tCategory = useCategoryName()
  const [amount, setAmount] = useState(String(template.amount_cents / 100))

  const cents = parseCents(amount)
  const canSubmit = cents > 0 && !isPending
  const displayName = template.name || tCategory(template.category_name)

  return (
    <BottomSheet onClose={onClose}>
      <div
        className="px-5 space-y-4"
        style={{ paddingBottom: 'max(1.5rem, env(safe-area-inset-bottom))' }}
      >
        <div className="flex items-center gap-3">
          <div
            className="w-12 h-12 rounded-2xl flex items-center justify-center shrink-0"
            style={{ background: template.category_color || 'var(--color-accent)' }}
          >
            <CategoryIcon icon={template.category_icon} size={24} weight="fill" className="text-white" />
          </div>
          <div className="min-w-0 flex-1">
            <p className="text-sm font-bold text-text truncate">{displayName}</p>
            <p className="text-[11px] text-muted">{t('templates.amount_variable_hint')}</p>
          </div>
        </div>

        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('transactions.amount')}
          </label>
          <AmountInput value={amount} onChange={setAmount} currency={template.currency_code} />
        </div>

        <button
          onClick={() => onSubmit(cents)}
          disabled={!canSubmit}
          className={`
            w-full py-4 rounded-2xl text-[15px] font-bold transition-all active:scale-[0.98]
            ${canSubmit
              ? 'bg-accent text-accent-text shadow-(--shadow-button)'
              : 'bg-border text-muted'
            }
          `}
        >
          {isPending ? t('common.loading') : t('common.confirm')}
        </button>
      </div>
    </BottomSheet>
  )
}
