import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { AnimatePresence } from 'framer-motion'
import { friendlyError } from '../../shared/lib/errors'
import { COLOR_SWATCHES } from '../../shared/lib/constants'
import { Plus, PencilSimple, Tag } from '@phosphor-icons/react'
import { categoriesApi } from '../../shared/api/categories'
import { CategoryIcon } from '../../shared/lib/categoryIcons'
import { Spinner } from '../../shared/ui/Spinner'
import { ErrorMessage } from '../../shared/ui/ErrorMessage'
import { PageTransition } from '../../shared/ui/PageTransition'
import { useTgBackButton } from '../../shared/hooks/useTelegramApp'
import { useHaptic } from '../../shared/hooks/useHaptic'
import { Badge, EmptyState, ActionRow, FAB, BottomSheet, ColorPicker, IconPicker } from '../../shared/ui'
import { useCategoryName } from '../../shared/hooks/useCategoryName'
import type { Category } from '../../shared/types'

const TYPE_OPTIONS = [
  { value: 'expense', labelKey: 'categories.type_expense' },
  { value: 'income', labelKey: 'categories.type_income' },
  { value: 'both', labelKey: 'categories.type_both' },
]

/* ─── Create / Edit Form (bottom sheet) ─── */
function CategoryForm({
  editingCat,
  categoryCount,
  onClose,
}: {
  editingCat: Category | null
  categoryCount: number
  onClose: () => void
}) {
  const { t } = useTranslation()
  const tCategory = useCategoryName()
  const qc = useQueryClient()
  const { notification } = useHaptic()

  const defaultColor = editingCat?.color ?? COLOR_SWATCHES[categoryCount % COLOR_SWATCHES.length]

  const [name, setName] = useState(editingCat ? tCategory(editingCat.name) : '')
  const [iconId, setIconId] = useState(editingCat?.emoji ?? 'star')
  const [catType, setCatType] = useState(editingCat?.type ?? 'both')
  const [color, setColor] = useState(defaultColor)

  const createMut = useMutation({
    mutationFn: () => categoriesApi.create({ name, emoji: iconId, type: catType, color }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['categories'] }); notification('success'); onClose() },
  })

  const updateMut = useMutation({
    mutationFn: () => categoriesApi.update(editingCat!.id, { name, emoji: iconId, type: catType, color }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['categories'] }); notification('success'); onClose() },
  })

  const isPending = createMut.isPending || updateMut.isPending

  function handleSubmit() {
    if (!name.trim()) return
    if (editingCat) { updateMut.mutate() } else { createMut.mutate() }
  }

  return (
    <BottomSheet onClose={onClose}>
      <div
        className="px-5 space-y-4 overflow-y-auto no-scrollbar"
        style={{ maxHeight: '85dvh', paddingBottom: 'max(1.5rem, env(safe-area-inset-bottom))' }}
      >
        {/* Name + selected icon preview */}
        <div className="flex gap-3 items-end">
          <div
            className="w-12 h-12 rounded-2xl flex items-center justify-center shrink-0"
            style={{ background: color, boxShadow: `0 2px 8px ${color}66` }}
          >
            <CategoryIcon emoji={iconId} size={22} weight="fill" className="text-white" />
          </div>
          <div className="flex-1">
            <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
              {t('categories.name')}
            </label>
            <input
              type="text"
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder={t('categories.name')}
              maxLength={30}
              autoFocus
              className="w-full bg-bg rounded-2xl px-4 py-3 text-sm font-medium outline-none text-text placeholder:text-muted/50 transition-shadow focus:shadow-(--shadow-focus)"
            />
          </div>
        </div>

        {/* Type toggle */}
        <div className="flex gap-1.5">
          {TYPE_OPTIONS.map((opt) => {
            const isActive = opt.value === catType
            return (
              <button
                key={opt.value}
                type="button"
                onClick={() => setCatType(opt.value)}
                className={`
                  flex-1 py-2.5 rounded-2xl text-[13px] font-bold transition-all duration-200 select-none
                  ${isActive
                    ? 'bg-accent text-accent-text shadow-(--shadow-accent-pill)'
                    : 'bg-accent-subtle text-muted'
                  }
                `}
              >
                {t(opt.labelKey)}
              </button>
            )
          })}
        </div>

        {/* Icon picker */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('categories.icon')}
          </label>
          <IconPicker selected={iconId} onSelect={setIconId} />
        </div>

        {/* Color picker */}
        <div>
          <label className="block text-[11px] font-bold text-muted uppercase tracking-widest mb-1.5">
            {t('categories.color')}
          </label>
          <ColorPicker selected={color} onSelect={setColor} />
        </div>

        {/* Submit */}
        <button
          onClick={handleSubmit}
          disabled={!name.trim() || isPending}
          className={`
            w-full py-4 rounded-2xl text-[15px] font-bold transition-all active:scale-[0.98]
            ${name.trim() && !isPending
              ? 'bg-accent text-accent-text shadow-(--shadow-button)'
              : 'bg-border text-muted'
            }
          `}
        >
          {isPending ? t('common.loading') : editingCat ? t('common.save') : t('common.create')}
        </button>

        {(createMut.isError || updateMut.isError) && (
          <p className="text-xs text-destructive text-center">
            {friendlyError(createMut.error || updateMut.error, t)}
          </p>
        )}
      </div>
    </BottomSheet>
  )
}

/* ─── Category Row ─── */
function CategoryRow({
  cat,
  onEdit,
  onDelete,
  isDeleting,
}: {
  cat: Category
  onEdit: (cat: Category) => void
  onDelete: (id: number) => void
  isDeleting: boolean
}) {
  const { t } = useTranslation()
  const tCategory = useCategoryName()
  return (
    <div className={`transition-opacity ${isDeleting ? 'opacity-30 pointer-events-none' : ''}`}>
      <ActionRow onDelete={() => onDelete(cat.id)}>
        <div className="flex items-center gap-3 px-4 py-3">
          <div
            className="w-10 h-10 rounded-2xl flex items-center justify-center shrink-0"
            style={{ background: cat.color || 'var(--color-accent)' }}
          >
            <CategoryIcon emoji={cat.emoji} size={20} weight="fill" className="text-white" />
          </div>
          <button
            onClick={() => onEdit(cat)}
            className="flex-1 min-w-0 flex items-center gap-2 text-left"
          >
            <span className="text-[13px] font-semibold text-text truncate">{tCategory(cat.name)}</span>
            <Badge variant="default" className="text-[10px] shrink-0 capitalize">
              {cat.type === 'both' ? t('categories.type_both') : t(`categories.type_${cat.type}`)}
            </Badge>
          </button>
          <button
            onClick={() => onEdit(cat)}
            className="w-11 h-11 rounded-2xl flex items-center justify-center text-muted active:text-accent active:bg-accent-subtle transition-colors shrink-0"
          >
            <PencilSimple size={18} weight="bold" />
          </button>
        </div>
      </ActionRow>
    </div>
  )
}

/* ─── Main Page ─── */
export function CategoriesPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { notification } = useHaptic()
  useTgBackButton(() => navigate('/more'))

  const [editingCat, setEditingCat] = useState<Category | null>(null)
  const [showCreate, setShowCreate] = useState(false)
  const [deletingId, setDeletingId] = useState<number | null>(null)

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => categoriesApi.delete(id),
    onMutate: (id) => setDeletingId(id),
    onSettled: () => setDeletingId(null),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['categories'] })
      notification('success')
    },
    onError: () => notification('error'),
  })

  const categories = data?.categories ?? []
  const formOpen = showCreate || editingCat !== null

  if (isLoading) return <div className="flex justify-center py-16"><Spinner /></div>
  if (isError) return <ErrorMessage onRetry={refetch} />

  return (
    <PageTransition>
      <div className="pt-3 pb-4">
          {categories.length === 0 ? (
            <div className="mx-4 card-elevated mt-2">
              <EmptyState
                icon={Tag}
                title={t('categories.no_categories')}
                action={
                  <button
                    onClick={() => setShowCreate(true)}
                    className="flex items-center gap-1.5 px-5 py-2.5 rounded-full bg-accent text-accent-text text-xs font-bold shadow-(--shadow-accent-pill) active:scale-95 transition-transform"
                  >
                    <Plus size={14} weight="bold" />
                    {t('categories.create_new')}
                  </button>
                }
              />
            </div>
          ) : (
            <div className="mx-4 card-elevated divide-y divide-border">
              {categories.map((cat) => (
                <CategoryRow
                  key={cat.id}
                  cat={cat}
                  onEdit={setEditingCat}
                  onDelete={id => deleteMut.mutate(id)}
                  isDeleting={deletingId === cat.id}
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

      <FAB onClick={() => setShowCreate(true)} label={t('categories.create_new')} />

      {/* Bottom sheet form */}
      <AnimatePresence>
        {formOpen && (
          <CategoryForm
            key={editingCat?.id ?? 'new'}
            editingCat={editingCat}
            categoryCount={categories.length}
            onClose={() => { setShowCreate(false); setEditingCat(null) }}
          />
        )}
      </AnimatePresence>
    </PageTransition>
  )
}
