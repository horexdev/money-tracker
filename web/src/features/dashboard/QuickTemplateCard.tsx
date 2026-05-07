import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { CategoryIcon } from '../../shared/lib/categoryIcons'
import { formatCents } from '../../shared/lib/money'
import { useLongPress } from '../../shared/hooks/useLongPress'
import { useCategoryName } from '../../shared/hooks/useCategoryName'
import type { TransactionTemplate } from '../../shared/types'

interface QuickTemplateCardProps {
  template: TransactionTemplate
  onTap: (template: TransactionTemplate) => void
}

export function QuickTemplateCard({ template, onTap }: QuickTemplateCardProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const tCategory = useCategoryName()

  const handlers = useLongPress(
    () => navigate('/add', { state: { fromTemplate: template } }),
    {
      delay: 500,
      onClick: () => onTap(template),
    },
  )

  const displayName = template.name || tCategory(template.category_name)

  return (
    <button
      type="button"
      {...handlers}
      className="shrink-0 w-[112px] card-elevated p-3 flex flex-col gap-2 active:scale-[0.97] transition-transform"
      style={{ touchAction: 'manipulation', userSelect: 'none' }}
      aria-label={t('templates.apply')}
    >
      <div
        className="w-10 h-10 rounded-2xl flex items-center justify-center"
        style={{ background: template.category_color || 'var(--color-accent)' }}
      >
        <CategoryIcon icon={template.category_icon} size={20} weight="fill" className="text-white" />
      </div>
      <p className="text-[12px] font-bold text-text truncate text-left">{displayName}</p>
      <p className="text-[11px] text-muted tabular-nums truncate text-left">
        {formatCents(template.amount_cents, template.currency_code)}
      </p>
    </button>
  )
}
