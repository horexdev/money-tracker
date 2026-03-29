import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { motion, AnimatePresence } from 'framer-motion'
import { Plus, PencilSimple } from '@phosphor-icons/react'
import { categoriesApi } from '../api/categories'
import { CategoryIcon, ICON_CHOICES } from '../lib/categoryIcons'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { useHaptic } from '../hooks/useHaptic'
import { Badge, EmptyState, SwipeToDelete } from '../components/ui'
import { useCategoryName } from '../hooks/useCategoryName'
import type { Category } from '../types'

const TYPE_OPTIONS = [
  { value: 'expense', labelKey: 'categories.type_expense' },
  { value: 'income', labelKey: 'categories.type_income' },
  { value: 'both', labelKey: 'categories.type_both' },
]

/* ─── Bottom Sheet ─── */
function BottomSheet({ onClose, children }: { onClose: () => void; children: React.ReactNode }) {
  return (
    <>
      <motion.div
        className="fixed inset-0 bg-black/40 z-[60]"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        onClick={onClose}
      />
      <motion.div
        className="fixed bottom-0 left-0 right-0 z-[60] bg-surface rounded-t-[24px] overflow-hidden"
        initial={{ y: '100%' }}
        animate={{ y: 0 }}
        exit={{ y: '100%' }}
        transition={{ type: 'spring', damping: 30, stiffness: 300 }}
        drag="y"
        dragConstraints={{ top: 0 }}
        dragElastic={{ top: 0, bottom: 0.3 }}
        onDragEnd={(_, info) => {
          if (info.velocity.y > 500 || info.offset.y > 100) onClose()
        }}
      >
        <div className="pt-3 pb-1 flex justify-center">
          <div className="w-10 h-1 rounded-full bg-border" />
        </div>
        {children}
      </motion.div>
    </>
  )
}

/* ─── Icon Picker ─── */
function IconPicker({
  selected,
  onSelect,
}: {
  selected: string
  onSelect: (id: string) => void
}) {
  const { selection } = useHaptic()

  return (
    <div className="grid grid-cols-8 gap-1.5">
      {ICON_CHOICES.map((choice) => {
        const isActive = selected === choice.id
        return (
          <button
            key={choice.id}
            type="button"
            onClick={() => { onSelect(choice.id); selection() }}
            className={`
              flex items-center justify-center h-10 rounded-xl transition-all duration-150 active:scale-90
              ${isActive
                ? 'bg-accent text-white shadow-[0_2px_8px_rgba(99,102,241,0.4)]'
                : 'bg-accent-subtle text-accent'
              }
            `}
          >
            <choice.Icon size={18} weight="fill" />
          </button>
        )
      })}
    </div>
  )
}

/* ─── Create / Edit Form (bottom sheet) ─── */
function CategoryForm({
  editingCat,
  onClose,
}: {
  editingCat: Category | null
  onClose: () => void
}) {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const { notification } = useHaptic()

  const [name, setName] = useState(editingCat?.name ?? '')
  const [iconId, setIconId] = useState(editingCat?.emoji ?? 'star')
  const [catType, setCatType] = useState(editingCat?.type ?? 'both')

  const createMut = useMutation({
    mutationFn: () => categoriesApi.create({ name, emoji: iconId, type: catType }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['categories'] }); notification('success'); onClose() },
  })

  const updateMut = useMutation({
    mutationFn: () => categoriesApi.update(editingCat!.id, { name, emoji: iconId, type: catType }),
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
          <div className="w-12 h-12 rounded-2xl bg-accent flex items-center justify-center shrink-0 shadow-[0_2px_8px_rgba(99,102,241,0.3)]">
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
              className="w-full bg-bg rounded-2xl px-4 py-3 text-sm font-medium outline-none text-text placeholder:text-muted/50 transition-shadow focus:shadow-[0_0_0_2px_rgba(99,102,241,0.2)]"
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
                    ? 'bg-accent text-accent-text shadow-[0_2px_8px_rgba(99,102,241,0.3)]'
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

        {/* Submit */}
        <button
          onClick={handleSubmit}
          disabled={!name.trim() || isPending}
          className={`
            w-full py-4 rounded-2xl text-[15px] font-bold transition-all active:scale-[0.98]
            ${name.trim() && !isPending
              ? 'bg-accent text-accent-text shadow-[0_4px_16px_rgba(99,102,241,0.3)]'
              : 'bg-border text-muted'
            }
          `}
        >
          {isPending ? '...' : editingCat ? t('common.save') : t('common.create')}
        </button>

        {(createMut.isError || updateMut.isError) && (
          <p className="text-xs text-destructive text-center">
            {((createMut.error || updateMut.error) as Error)?.message}
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
      <SwipeToDelete onDelete={() => onDelete(cat.id)}>
        <div className="flex items-center gap-3 px-4 py-3">
          <div className="w-10 h-10 rounded-xl bg-accent-subtle flex items-center justify-center shrink-0">
            <CategoryIcon emoji={cat.emoji} size={20} weight="fill" className="text-accent" />
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
            className="w-11 h-11 rounded-xl flex items-center justify-center text-muted active:text-accent active:bg-accent-subtle transition-colors shrink-0"
          >
            <PencilSimple size={18} weight="bold" />
          </button>
        </div>
      </SwipeToDelete>
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
      <div className="flex flex-col h-[calc(100dvh-var(--tab-bar-h))]">

        {/* Add button */}
        <div className="shrink-0 px-4 pt-3 pb-2 flex justify-end">
          <button
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-1.5 px-4 py-2 rounded-full bg-accent text-accent-text text-xs font-bold shadow-[0_2px_12px_rgba(99,102,241,0.4)] active:scale-95 transition-transform"
          >
            <Plus size={14} weight="bold" />
            {t('categories.create_new')}
          </button>
        </div>

        {/* Scrollable list */}
        <div className="flex-1 min-h-0 overflow-y-auto no-scrollbar pb-3">
          {categories.length === 0 ? (
            <div className="mx-4 card-elevated mt-2">
              <EmptyState icon="🏷" title={t('categories.no_categories')} />
            </div>
          ) : (
            <div className="mx-4 card-elevated overflow-hidden divide-y divide-border">
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
              <p className="text-xs text-destructive text-center bg-expense/10 rounded-xl py-2 px-3">
                {(deleteMut.error as Error)?.message}
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Bottom sheet form */}
      <AnimatePresence>
        {formOpen && (
          <CategoryForm
            key={editingCat?.id ?? 'new'}
            editingCat={editingCat}
            onClose={() => { setShowCreate(false); setEditingCat(null) }}
          />
        )}
      </AnimatePresence>
    </PageTransition>
  )
}
