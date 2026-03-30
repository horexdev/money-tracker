import { useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Users, UserPlus, ChartLineUp, CaretLeft, CaretRight, Trash, Warning } from '@phosphor-icons/react'
import { adminApi } from '../api/admin'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { FIRST_LAUNCH_KEY } from '../hooks/useFirstLaunchSetup'
import type { AdminUser } from '../types'

function StatCard({ label, value, icon, gradient }: {
  label: string
  value: string | number
  icon: React.ReactNode
  gradient: string
}) {
  return (
    <div className="card-elevated p-4 relative overflow-hidden">
      <div className="absolute -top-4 -right-4 w-16 h-16 rounded-full bg-accent/[0.04] pointer-events-none" />
      <div className={`w-9 h-9 rounded-xl bg-gradient-to-br ${gradient} flex items-center justify-center text-white mb-2 shadow-sm`}>
        {icon}
      </div>
      <p className="text-2xl font-bold text-text">{value}</p>
      <p className="text-xs text-muted font-medium mt-0.5">{label}</p>
    </div>
  )
}

function RetentionBar({ label, value }: { label: string; value: number }) {
  const pct = Math.min(Math.max(value, 0), 100)
  return (
    <div className="flex items-center gap-3">
      <span className="text-xs text-muted font-medium w-12 shrink-0">{label}</span>
      <div className="flex-1 h-2.5 rounded-full bg-border/50 overflow-hidden">
        <div
          className="h-full rounded-full bg-gradient-to-r from-indigo-500 to-violet-500 transition-all duration-500"
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className="text-xs font-bold text-text w-12 text-right">{value.toFixed(1)}%</span>
    </div>
  )
}

function UserRow({ user, onReset, resetting }: { user: AdminUser; onReset: (id: number) => void; resetting: boolean }) {
  const { t } = useTranslation()
  const name = [user.first_name, user.last_name].filter(Boolean).join(' ') || '—'
  const date = new Date(user.created_at)
  const formattedDate = date.toLocaleDateString(undefined, { day: '2-digit', month: 'short', year: 'numeric' })

  return (
    <div className="flex items-center gap-3 py-3 border-b border-border/30 last:border-0">
      <div className="w-9 h-9 rounded-full bg-gradient-to-br from-indigo-400 to-violet-500 flex items-center justify-center text-white text-xs font-bold shrink-0">
        {(user.first_name?.[0] || user.username?.[0] || '?').toUpperCase()}
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-semibold text-text truncate">{name}</p>
        <p className="text-xs text-muted truncate">
          {user.username ? `@${user.username}` : `ID: ${user.id}`}
          {' · '}{user.currency_code} · {user.language}
        </p>
      </div>
      <p className="text-[11px] text-muted shrink-0">{formattedDate}</p>
      <button
        onClick={() => onReset(user.id)}
        disabled={resetting}
        title={t('admin.reset_user')}
        className="w-7 h-7 rounded-lg bg-red-500/10 flex items-center justify-center text-red-400 disabled:opacity-40 active:scale-90 transition-transform shrink-0"
      >
        {resetting ? <Spinner size="sm" /> : <Trash size={14} weight="bold" />}
      </button>
    </div>
  )
}

export function AdminPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [page, setPage] = useState(1)
  const pageSize = 20
  const [confirmResetAll, setConfirmResetAll] = useState(false)
  const [resettingUserId, setResettingUserId] = useState<number | null>(null)

  useTgBackButton(useCallback(() => navigate('/more'), [navigate]))

  const { data: stats, isLoading: statsLoading, error: statsError } = useQuery({
    queryKey: ['admin-stats'],
    queryFn: adminApi.getStats,
    staleTime: 30_000,
  })

  const { data: usersData, isLoading: usersLoading, error: usersError } = useQuery({
    queryKey: ['admin-users', page],
    queryFn: () => adminApi.getUsers(page, pageSize),
    staleTime: 30_000,
  })

  const resetUserMutation = useMutation({
    mutationFn: (userID: number) => adminApi.resetUser(userID),
    onSuccess: () => {
      localStorage.removeItem(FIRST_LAUNCH_KEY)
      setResettingUserId(null)
      window.location.reload()
    },
    onError: () => {
      setResettingUserId(null)
    },
  })

  const resetAllMutation = useMutation({
    mutationFn: adminApi.resetAllUsers,
    onSuccess: () => {
      localStorage.removeItem(FIRST_LAUNCH_KEY)
      setConfirmResetAll(false)
      window.location.reload()
    },
    onError: () => {
      setConfirmResetAll(false)
    },
  })

  const handleResetUser = (id: number) => {
    if (!window.confirm(t('admin.confirm_reset_user'))) return
    setResettingUserId(id)
    resetUserMutation.mutate(id)
  }

  const totalPages = usersData ? Math.ceil(usersData.total / pageSize) : 1

  if (statsLoading && usersLoading) {
    return <PageTransition><div className="flex justify-center items-center h-48"><Spinner /></div></PageTransition>
  }

  if (statsError) {
    return <PageTransition><div className="px-4 pt-4"><ErrorMessage onRetry={() => {}} /></div></PageTransition>
  }

  return (
    <PageTransition>
      <div className="px-4 pt-3 pb-4 space-y-4">
        {/* Header */}
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-bold text-text">{t('admin.title')}</h1>
          {confirmResetAll ? (
            <div className="flex items-center gap-2">
              <button
                onClick={() => setConfirmResetAll(false)}
                className="text-xs px-3 py-1.5 rounded-xl bg-surface-alt text-muted font-medium active:scale-95 transition-transform"
              >
                {t('cancel')}
              </button>
              <button
                onClick={() => resetAllMutation.mutate()}
                disabled={resetAllMutation.isPending}
                className="text-xs px-3 py-1.5 rounded-xl bg-red-500/20 text-red-400 font-semibold active:scale-95 transition-transform disabled:opacity-50 flex items-center gap-1"
              >
                {resetAllMutation.isPending ? <Spinner size="sm" /> : <Warning size={12} weight="fill" />}
                {t('admin.confirm_reset_all_short')}
              </button>
            </div>
          ) : (
            <button
              onClick={() => setConfirmResetAll(true)}
              className="text-xs px-3 py-1.5 rounded-xl bg-red-500/10 text-red-400 font-medium active:scale-95 transition-transform flex items-center gap-1.5"
            >
              <Trash size={12} weight="bold" />
              {t('admin.reset_all')}
            </button>
          )}
        </div>

        {/* Stats Overview */}
        {stats && (
          <>
            <div className="grid grid-cols-2 gap-3">
              <StatCard
                label={t('admin.total_users')}
                value={stats.total_users}
                icon={<Users size={18} weight="fill" />}
                gradient="from-indigo-400 to-violet-500"
              />
              <StatCard
                label={t('admin.new_today')}
                value={stats.new_today}
                icon={<UserPlus size={18} weight="fill" />}
                gradient="from-emerald-400 to-teal-500"
              />
              <StatCard
                label={t('admin.new_this_week')}
                value={stats.new_this_week}
                icon={<UserPlus size={18} weight="bold" />}
                gradient="from-blue-400 to-indigo-500"
              />
              <StatCard
                label={t('admin.new_this_month')}
                value={stats.new_this_month}
                icon={<ChartLineUp size={18} weight="bold" />}
                gradient="from-amber-400 to-orange-500"
              />
            </div>

            {/* Retention */}
            <div className="card-elevated p-4 space-y-3">
              <p className="text-sm font-bold text-text">{t('admin.retention')}</p>
              <RetentionBar label={t('admin.retention_day1')} value={stats.retention_day1} />
              <RetentionBar label={t('admin.retention_day7')} value={stats.retention_day7} />
              <RetentionBar label={t('admin.retention_day30')} value={stats.retention_day30} />
            </div>
          </>
        )}

        {/* User List */}
        <div className="card-elevated p-4">
          <div className="flex items-center justify-between mb-3">
            <p className="text-sm font-bold text-text">{t('admin.user_list')}</p>
            {usersData && (
              <p className="text-xs text-muted">
                {t('admin.page_of', { page, total: totalPages })}
              </p>
            )}
          </div>

          {usersLoading ? (
            <div className="flex justify-center py-6"><Spinner /></div>
          ) : usersError ? (
            <ErrorMessage onRetry={() => {}} />
          ) : usersData && usersData.users.length === 0 ? (
            <p className="text-sm text-muted text-center py-6">{t('admin.no_users')}</p>
          ) : (
            <>
              <div>
                {usersData?.users.map((user) => (
                  <UserRow
                    key={user.id}
                    user={user}
                    onReset={handleResetUser}
                    resetting={resettingUserId === user.id}
                  />
                ))}
              </div>

              {/* Pagination */}
              {totalPages > 1 && (
                <div className="flex items-center justify-center gap-4 mt-4">
                  <button
                    onClick={() => setPage(p => Math.max(1, p - 1))}
                    disabled={page <= 1}
                    className="w-8 h-8 rounded-xl bg-surface-alt flex items-center justify-center text-muted disabled:opacity-30 active:scale-90 transition-transform"
                  >
                    <CaretLeft size={16} weight="bold" />
                  </button>
                  <span className="text-sm font-semibold text-text">
                    {page} / {totalPages}
                  </span>
                  <button
                    onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                    disabled={page >= totalPages}
                    className="w-8 h-8 rounded-xl bg-surface-alt flex items-center justify-center text-muted disabled:opacity-30 active:scale-90 transition-transform"
                  >
                    <CaretRight size={16} weight="bold" />
                  </button>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </PageTransition>
  )
}
