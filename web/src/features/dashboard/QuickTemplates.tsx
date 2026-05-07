import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { templatesApi } from '../../shared/api/templates'
import { useApplyTemplate } from '../../shared/hooks/useApplyTemplate'
import { QuickTemplateCard } from './QuickTemplateCard'
import { QuickAmountModal } from './QuickAmountModal'
import type { TransactionTemplate } from '../../shared/types'

const MAX_VISIBLE = 7

/**
 * QuickTemplates renders a horizontal carousel of the user's first 7 templates
 * (by sort_order). Tap → apply (instant for fixed-amount, modal for variable).
 * Long-press → navigate to /add with the template prefilled.
 */
export function QuickTemplates() {
  const { t } = useTranslation()
  const [pending, setPending] = useState<TransactionTemplate | null>(null)

  const { data, isPending, isError } = useQuery({
    queryKey: ['templates'],
    queryFn: templatesApi.list,
  })

  const applyMut = useApplyTemplate()

  if (isPending || isError) return null
  const items = (data?.templates ?? []).slice(0, MAX_VISIBLE)
  if (items.length === 0) return null

  const handleTap = (tpl: TransactionTemplate) => {
    if (tpl.amount_fixed) {
      applyMut.mutate({ templateId: tpl.id })
    } else {
      setPending(tpl)
    }
  }

  return (
    <div className="px-4 pt-3">
      <p className="text-[11px] font-bold text-muted uppercase tracking-widest mb-2 px-1">
        {t('templates.quick')}
      </p>
      <div className="flex gap-2 overflow-x-auto no-scrollbar -mx-4 px-4 pb-1">
        {items.map(tpl => (
          <QuickTemplateCard key={tpl.id} template={tpl} onTap={handleTap} />
        ))}
      </div>

      {pending && (
        <QuickAmountModal
          template={pending}
          isPending={applyMut.isPending}
          onClose={() => setPending(null)}
          onSubmit={(amountCents) => {
            applyMut.mutate(
              { templateId: pending.id, amountCents },
              { onSuccess: () => setPending(null) },
            )
          }}
        />
      )}
    </div>
  )
}
