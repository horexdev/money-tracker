import { useState, useEffect, useMemo, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { AnimatePresence, Reorder } from 'framer-motion'
import { Plus, Lightning, DotsSixVertical } from '@phosphor-icons/react'
import { templatesApi } from '../../shared/api/templates'
import { formatCents } from '../../shared/lib/money'
import { friendlyError } from '../../shared/lib/errors'
import { CategoryIcon } from '../../shared/lib/categoryIcons'
import { Spinner } from '../../shared/ui/Spinner'
import { ErrorMessage } from '../../shared/ui/ErrorMessage'
import { PageTransition } from '../../shared/ui/PageTransition'
import { Badge, EmptyState, FAB } from '../../shared/ui'
import { useTgBackButton } from '../../shared/hooks/useTelegramApp'
import { useHaptic } from '../../shared/hooks/useHaptic'
import { useCategoryName } from '../../shared/hooks/useCategoryName'
import { TemplateForm } from './TemplateForm'
import type { TransactionTemplate } from '../../shared/types'

interface TemplateRowProps {
  item: TransactionTemplate
  onEdit: (item: TransactionTemplate) => void
  onDelete: (id: number) => void
}

function TemplateRow({ item, onEdit, onDelete }: TemplateRowProps) {
  const { t } = useTranslation()
  const tCategory = useCategoryName()
  const displayName = item.name || tCategory(item.category_name)

  return (
    <div className="flex items-center gap-3 px-4 py-3.5">
      <DotsSixVertical size={18} weight="bold" className="text-muted/40 shrink-0 cursor-grab" />
      <button
        onClick={() => onEdit(item)}
        className="w-10 h-10 rounded-2xl flex items-center justify-center shrink-0 active:scale-95 transition-transform"
        style={{ background: item.category_color || 'var(--color-accent)' }}
      >
        <CategoryIcon icon={item.category_icon} size={20} weight="fill" className="text-white" />
      </button>
      <button
        onClick={() => onEdit(item)}
        className="flex-1 min-w-0 text-left"
      >
        <div className="flex items-center gap-2 mb-0.5">
          <span className="text-[13px] font-bold text-text truncate">{displayName}</span>
          <Badge variant={item.type === 'income' ? 'income' : 'expense'} className="text-[10px] shrink-0">
            {item.type === 'income' ? t('transactions.income') : t('transactions.expense')}
          </Badge>
        </div>
        <div className="flex items-center gap-1 text-xs text-muted">
          <span className="font-semibold text-text tabular-nums">
            {formatCents(item.amount_cents, item.currency_code)}
          </span>
          {!item.amount_fixed && (
            <>
              <span className="text-muted/40">·</span>
              <span>{t('templates.amount_variable_hint_short')}</span>
            </>
          )}
        </div>
        {item.note && <p className="text-[11px] text-muted/70 mt-0.5 truncate">{item.note}</p>}
      </button>
      <button
        onClick={() => onDelete(item.id)}
        className="w-9 h-9 rounded-xl flex items-center justify-center bg-destructive/10 text-destructive active:bg-destructive/20 transition-colors shrink-0"
        aria-label={t('common.delete')}
      >
        <span className="text-lg leading-none">×</span>
      </button>
    </div>
  )
}

export function TemplatesPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { notification } = useHaptic()
  useTgBackButton(() => navigate('/more'))

  const [editingItem, setEditingItem] = useState<TransactionTemplate | null>(null)
  const [showCreate, setShowCreate] = useState(false)
  const reorderTimerRef = useRef<number | null>(null)

  const templatesQ = useQuery({ queryKey: ['templates'], queryFn: templatesApi.list })

  const items = useMemo(() => templatesQ.data?.templates ?? [], [templatesQ.data])
  const [orderedItems, setOrderedItems] = useState<TransactionTemplate[]>(items)

  // Keep local order in sync with server data when it arrives.
  useEffect(() => {
    setOrderedItems(items)
  }, [items])

  const reorderMut = useMutation({
    mutationFn: (orderedIds: number[]) => templatesApi.reorder(orderedIds),
    onSuccess: (data) => {
      qc.setQueryData(['templates'], data)
    },
  })

  const handleReorder = (next: TransactionTemplate[]) => {
    setOrderedItems(next)
    if (reorderTimerRef.current !== null) {
      window.clearTimeout(reorderTimerRef.current)
    }
    reorderTimerRef.current = window.setTimeout(() => {
      reorderMut.mutate(next.map(t => t.id))
      reorderTimerRef.current = null
    }, 400)
  }

  const deleteMut = useMutation({
    mutationFn: templatesApi.delete,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['templates'] })
      notification('success')
    },
    onError: () => notification('error'),
  })

  if (templatesQ.isPending) return <div className="flex justify-center py-16"><Spinner /></div>
  if (templatesQ.isError) return <ErrorMessage onRetry={() => templatesQ.refetch()} />

  const formOpen = showCreate || editingItem !== null

  return (
    <PageTransition>
      <div className="pt-3 pb-4">
        {orderedItems.length === 0 ? (
          <div className="mx-4 card-elevated mt-2">
            <EmptyState
              icon={Lightning}
              title={t('templates.no_templates')}
              description={t('templates.create_first')}
              action={
                <button
                  onClick={() => setShowCreate(true)}
                  className="flex items-center gap-1.5 px-5 py-2.5 rounded-full bg-accent text-accent-text text-xs font-bold shadow-(--shadow-accent-pill) active:scale-95 transition-transform"
                >
                  <Plus size={14} weight="bold" />
                  {t('templates.create_new')}
                </button>
              }
            />
          </div>
        ) : (
          <Reorder.Group
            axis="y"
            values={orderedItems}
            onReorder={handleReorder}
            className="mx-4 card-elevated divide-y divide-border list-none"
          >
            {orderedItems.map(item => (
              <Reorder.Item
                key={item.id}
                value={item}
                className="bg-surface"
              >
                <TemplateRow
                  item={item}
                  onEdit={setEditingItem}
                  onDelete={id => deleteMut.mutate(id)}
                />
              </Reorder.Item>
            ))}
          </Reorder.Group>
        )}

        {deleteMut.isError && (
          <div className="mx-4 mt-2">
            <p className="text-xs text-destructive text-center bg-expense/10 rounded-2xl py-2 px-3">
              {friendlyError(deleteMut.error, t)}
            </p>
          </div>
        )}
      </div>

      <FAB onClick={() => setShowCreate(true)} label={t('templates.create_new')} />

      <AnimatePresence>
        {formOpen && (
          <TemplateForm
            key={editingItem?.id ?? 'new'}
            editItem={editingItem}
            onClose={() => { setShowCreate(false); setEditingItem(null) }}
          />
        )}
      </AnimatePresence>
    </PageTransition>
  )
}
