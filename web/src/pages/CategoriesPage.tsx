import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Plus, Trash2, Pencil, Lock } from 'lucide-react'
import { categoriesApi } from '../api/categories'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { Card, SectionHeader, Button, Badge, EmptyState, SegmentedControl } from '../components/ui'
import type { Category } from '../types'

const TYPE_OPTIONS = [
  { value: 'expense', label: 'Expense' },
  { value: 'income', label: 'Income' },
  { value: 'both', label: 'Both' },
]

export function CategoriesPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const qc = useQueryClient()
  useTgBackButton(() => navigate('/more'))

  const [showForm, setShowForm] = useState(false)
  const [editingCat, setEditingCat] = useState<Category | null>(null)
  const [name, setName] = useState('')
  const [emoji, setEmoji] = useState('')
  const [catType, setCatType] = useState('both')

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  })

  const createMut = useMutation({
    mutationFn: () => categoriesApi.create({ name, emoji, type: catType }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['categories'] })
      resetForm()
    },
  })

  const updateMut = useMutation({
    mutationFn: () => categoriesApi.update(editingCat!.id, { name, emoji, type: catType }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['categories'] })
      resetForm()
    },
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => categoriesApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['categories'] }),
  })

  function resetForm() {
    setShowForm(false)
    setEditingCat(null)
    setName('')
    setEmoji('')
    setCatType('both')
  }

  function startEdit(cat: Category) {
    setEditingCat(cat)
    setName(cat.name)
    setEmoji(cat.emoji)
    setCatType(cat.type || 'both')
    setShowForm(true)
  }

  function handleSubmit() {
    if (!name.trim()) return
    if (editingCat) {
      updateMut.mutate()
    } else {
      createMut.mutate()
    }
  }

  const categories = data?.categories ?? []
  const systemCats = categories.filter(c => c.is_system)
  const customCats = categories.filter(c => !c.is_system)

  if (isLoading) return <div className="flex justify-center py-16"><Spinner /></div>
  if (isError) return <ErrorMessage onRetry={refetch} />

  return (
    <PageTransition>
      <div className="p-4 space-y-4">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-bold">{t('categories.title')}</h1>
          {!showForm && (
            <Button size="sm" onClick={() => setShowForm(true)}>
              <Plus size={16} className="mr-1" /> {t('categories.create_new')}
            </Button>
          )}
        </div>

        {/* Create/Edit form */}
        {showForm && (
          <Card>
            <div className="space-y-3">
              <div className="flex gap-3">
                <div className="w-16">
                  <label className="block text-xs text-muted mb-1">{t('categories.emoji')}</label>
                  <input
                    type="text"
                    value={emoji}
                    onChange={e => setEmoji(e.target.value)}
                    placeholder="🏷"
                    maxLength={4}
                    className="w-full bg-surface rounded-[--radius-sm] px-3 py-2 text-center text-xl outline-none focus:ring-2 focus:ring-accent"
                  />
                </div>
                <div className="flex-1">
                  <label className="block text-xs text-muted mb-1">{t('categories.name')}</label>
                  <input
                    type="text"
                    value={name}
                    onChange={e => setName(e.target.value)}
                    placeholder={t('categories.name')}
                    maxLength={30}
                    className="w-full bg-surface rounded-[--radius-sm] px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-accent"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs text-muted mb-1">{t('categories.type')}</label>
                <SegmentedControl
                  options={TYPE_OPTIONS.map(o => ({ ...o, label: t(`categories.type_${o.value}`) }))}
                  value={catType}
                  onChange={setCatType}
                  size="sm"
                />
              </div>

              <div className="flex gap-2">
                <Button size="sm" onClick={handleSubmit} disabled={!name.trim() || createMut.isPending || updateMut.isPending}>
                  {editingCat ? t('common.save') : t('common.create')}
                </Button>
                <Button size="sm" variant="ghost" onClick={resetForm}>{t('common.cancel')}</Button>
              </div>

              {(createMut.isError || updateMut.isError) && (
                <p className="text-xs text-destructive">
                  {((createMut.error || updateMut.error) as Error)?.message}
                </p>
              )}
            </div>
          </Card>
        )}

        {/* Custom categories */}
        {customCats.length > 0 && (
          <div>
            <SectionHeader>{t('categories.custom')}</SectionHeader>
            <Card padding="p-0">
              <div className="divide-y divide-border">
                {customCats.map(cat => (
                  <div key={cat.id} className="flex items-center gap-3 px-4 py-3">
                    <span className="text-xl w-8 text-center">{cat.emoji}</span>
                    <div className="flex-1 min-w-0">
                      <span className="text-sm font-medium truncate">{cat.name}</span>
                      <Badge variant="default" className="ml-2 text-[10px]">{cat.type}</Badge>
                    </div>
                    <button onClick={() => startEdit(cat)} className="p-1.5 text-muted hover:text-accent">
                      <Pencil size={16} />
                    </button>
                    <button
                      onClick={() => deleteMut.mutate(cat.id)}
                      className="p-1.5 text-muted hover:text-destructive"
                      disabled={deleteMut.isPending}
                    >
                      <Trash2 size={16} />
                    </button>
                  </div>
                ))}
              </div>
            </Card>
          </div>
        )}

        {customCats.length === 0 && !showForm && (
          <EmptyState icon="🏷" title={t('categories.no_categories')} />
        )}

        {/* System categories */}
        {systemCats.length > 0 && (
          <div>
            <SectionHeader>{t('categories.system')}</SectionHeader>
            <Card padding="p-0">
              <div className="divide-y divide-border">
                {systemCats.map(cat => (
                  <div key={cat.id} className="flex items-center gap-3 px-4 py-3">
                    <span className="text-xl w-8 text-center">{cat.emoji}</span>
                    <span className="flex-1 text-sm text-text truncate">{cat.name}</span>
                    <Lock size={14} className="text-muted" />
                  </div>
                ))}
              </div>
            </Card>
          </div>
        )}

        {deleteMut.isError && (
          <p className="text-xs text-destructive text-center">{(deleteMut.error as Error)?.message}</p>
        )}
      </div>
    </PageTransition>
  )
}
