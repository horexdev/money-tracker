import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { CategoryIcon } from '../../shared/lib/categoryIcons'
import { formatCents } from '../../shared/lib/money'
import { templatesApi } from '../../shared/api/templates'
import { BottomSheet, EmptyState } from '../../shared/ui'
import { Spinner } from '../../shared/ui/Spinner'
import { useCategoryName } from '../../shared/hooks/useCategoryName'
import { Lightning } from '@phosphor-icons/react'
import type { TransactionTemplate } from '../../shared/types'

interface TemplatePickerProps {
  onSelect: (tpl: TransactionTemplate) => void
  onClose: () => void
}

export function TemplatePicker({ onSelect, onClose }: TemplatePickerProps) {
  const { t } = useTranslation()
  const tCategory = useCategoryName()

  const { data, isPending } = useQuery({ queryKey: ['templates'], queryFn: templatesApi.list })
  const items = data?.templates ?? []

  return (
    <BottomSheet onClose={onClose}>
      <div className="px-4 pb-6 max-h-[70dvh] overflow-y-auto no-scrollbar">
        <p className="px-1 pt-2 pb-3 text-sm font-bold text-text">
          {t('templates.from_template')}
        </p>
        {isPending ? (
          <div className="flex justify-center py-8"><Spinner /></div>
        ) : items.length === 0 ? (
          <EmptyState
            icon={Lightning}
            title={t('templates.no_templates')}
            description={t('templates.create_first')}
          />
        ) : (
          <div className="card-elevated divide-y divide-border">
            {items.map(tpl => {
              const displayName = tpl.name || tCategory(tpl.category_name)
              return (
                <button
                  key={tpl.id}
                  onClick={() => { onSelect(tpl); onClose() }}
                  className="w-full flex items-center gap-3 px-4 py-3 active:bg-border text-left"
                >
                  <div
                    className="w-10 h-10 rounded-2xl flex items-center justify-center shrink-0"
                    style={{ background: tpl.category_color || 'var(--color-accent)' }}
                  >
                    <CategoryIcon icon={tpl.category_icon} size={20} weight="fill" className="text-white" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-[13px] font-bold text-text truncate">{displayName}</p>
                    <p className="text-[11px] text-muted tabular-nums truncate">
                      {formatCents(tpl.amount_cents, tpl.currency_code)}
                      {!tpl.amount_fixed && <span className="ml-1">· {t('templates.amount_variable_hint_short')}</span>}
                    </p>
                  </div>
                </button>
              )
            })}
          </div>
        )}
      </div>
    </BottomSheet>
  )
}
