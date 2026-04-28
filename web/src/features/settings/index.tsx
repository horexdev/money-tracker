import { useState } from 'react'
import { createPortal } from 'react-dom'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { motion, AnimatePresence } from 'framer-motion'
import { Check, X, Globe, CaretRight, Trash, Desktop, Sun, Moon, Bell, RepeatOnce, CalendarCheck, Trophy, Sparkle } from '@phosphor-icons/react'
import { settingsApi } from '../../shared/api/settings'
import { LANGUAGES } from '../../shared/lib/constants'
import { Spinner } from '../../shared/ui/Spinner'
import { ErrorMessage } from '../../shared/ui/ErrorMessage'
import { PageTransition } from '../../shared/ui/PageTransition'
import { useTgBackButton, useThemePreference } from '../../shared/hooks/useTelegramApp'
import { useAnimateNumbers } from '../../shared/hooks/useAnimateNumbers'
import { useHaptic } from '../../shared/hooks/useHaptic'
import { FIRST_LAUNCH_KEY } from '../../shared/hooks/useFirstLaunchSetup'
import type { ThemePref } from '../../shared/hooks/useTelegramApp'

type Modal = 'none' | 'language' | 'reset-confirm'

/* ─── Bottom Sheet Modal ─── */
function BottomSheet({
  open,
  onClose,
  title,
  children,
}: {
  open: boolean
  onClose: () => void
  title: string
  children: React.ReactNode
}) {
  return createPortal(
    <AnimatePresence>
      {open && (
        <>
          {/* Backdrop */}
          <motion.div
            className="fixed inset-0 bg-black/40 z-40"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
          />
          {/* Sheet */}
          <motion.div
            className="fixed bottom-0 left-0 right-0 z-[60] bg-surface rounded-t-card flex flex-col"
            style={{ maxHeight: '80dvh' }}
            initial={{ y: '100%' }}
            animate={{ y: 0 }}
            exit={{ y: '100%' }}
            transition={{ type: 'spring', stiffness: 350, damping: 35 }}
            drag="y"
            dragConstraints={{ top: 0 }}
            dragElastic={0.1}
            onDragEnd={(_, info) => {
              if (info.velocity.y > 300 || info.offset.y > 120) onClose()
            }}
          >
            {/* Handle */}
            <div className="flex justify-center pt-3 pb-1 shrink-0">
              <div className="w-10 h-1 rounded-full bg-border" />
            </div>
            {/* Header */}
            <div className="flex items-center justify-between px-5 py-3 shrink-0">
              <span className="text-base font-bold text-text">{title}</span>
              <button
                onClick={onClose}
                className="w-8 h-8 rounded-full bg-accent-subtle flex items-center justify-center text-muted"
              >
                <X size={14} weight="bold" />
              </button>
            </div>
            {/* Content */}
            <div className="flex-1 min-h-0 overflow-y-auto no-scrollbar pb-[var(--tab-bar-h)]">
              {children}
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>,
    document.body,
  )
}

export function SettingsPage() {
  const { t, i18n } = useTranslation()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { selection, notification } = useHaptic()
  const [modal, setModal] = useState<Modal>('none')
  const [themePref, setThemePref] = useThemePreference()
  const [animateNumbers, setAnimateNumbers] = useAnimateNumbers()

  const THEME_OPTIONS: { value: ThemePref; labelKey: string; icon: React.ElementType }[] = [
    { value: 'system', labelKey: 'settings.theme_system', icon: Desktop },
    { value: 'light',  labelKey: 'settings.theme_light',  icon: Sun },
    { value: 'dark',   labelKey: 'settings.theme_dark',   icon: Moon },
  ]

  useTgBackButton(() => navigate('/more'))

  const { data: settings, isLoading, isError, refetch } = useQuery({
    queryKey: ['settings'],
    queryFn: settingsApi.get,
  })

  const languageMutation = useMutation({
    mutationFn: (lang: string) => settingsApi.update({ language: lang }),
    onSuccess: (data) => {
      qc.invalidateQueries({ queryKey: ['settings'] })
      i18n.changeLanguage(data.language)
      notification('success')
    },
  })

  const notifMutation = useMutation({
    mutationFn: (prefs: {
      notify_budget_alerts?: boolean
      notify_recurring_reminders?: boolean
      notify_weekly_summary?: boolean
      notify_goal_milestones?: boolean
    }) => settingsApi.update({ notification_preferences: prefs }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['settings'] })
      selection()
    },
  })

  const resetMutation = useMutation({
    mutationFn: () => settingsApi.resetData(),
    onSuccess: () => {
      localStorage.removeItem(FIRST_LAUNCH_KEY)
      qc.invalidateQueries({ queryKey: ['transactions'] })
      qc.invalidateQueries({ queryKey: ['balance'] })
      qc.invalidateQueries({ queryKey: ['stats'] })
      qc.invalidateQueries({ queryKey: ['budgets'] })
      qc.invalidateQueries({ queryKey: ['recurring'] })
      qc.invalidateQueries({ queryKey: ['goals'] })
      qc.invalidateQueries({ queryKey: ['categories'] })
      notification('success')
      setModal('none')
    },
    onError: () => notification('error'),
  })

  if (isLoading) return <div className="flex justify-center py-16"><Spinner /></div>
  if (isError || !settings) return <ErrorMessage onRetry={refetch} />

  const currentLang = settings.language || i18n.language || 'en'
  const currentLangData = LANGUAGES.find(l => l.code === currentLang)

  function selectLanguage(code: string) {
    selection()
    languageMutation.mutate(code)
    setModal('none')
  }

  function closeModal() {
    setModal('none')
  }

  return (
    <>
      <PageTransition>
        <div className="px-4 pt-3 pb-4 space-y-3">

              {/* Theme picker */}
              <div className="card-elevated p-4 space-y-3">
                <p className="text-[11px] font-bold text-muted uppercase tracking-widest">
                  {t('settings.theme')}
                </p>
                <div className="flex gap-2">
                  {THEME_OPTIONS.map(({ value, labelKey, icon: Icon }) => {
                    const isActive = themePref === value
                    return (
                      <button
                        key={value}
                        onClick={() => { selection(); setThemePref(value) }}
                        className={`
                          flex-1 flex flex-col items-center gap-1.5 py-3 rounded-2xl text-xs font-bold transition-all duration-200 select-none
                          ${isActive
                            ? 'bg-accent text-accent-text shadow-(--shadow-accent-pill)'
                            : 'bg-bg text-muted active:scale-95'
                          }
                        `}
                      >
                        <Icon size={18} weight={isActive ? 'fill' : 'regular'} />
                        {t(labelKey)}
                      </button>
                    )
                  })}
                </div>
              </div>

              {/* Display & Behavior section */}
              <div className="card-elevated">
                <div className="px-4 pt-4 pb-2">
                  <p className="text-[11px] font-bold text-muted uppercase tracking-widest">
                    {t('settings.display')}
                  </p>
                </div>
                <div className="flex items-center gap-4 px-4 py-3.5">
                  <div className="w-9 h-9 rounded-xl bg-accent/10 flex items-center justify-center shrink-0">
                    <Sparkle size={18} weight="fill" className="text-accent" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-[13px] font-semibold text-text leading-tight">
                      {t('settings.animate_numbers')}
                    </p>
                    <p className="text-[11px] text-muted mt-0.5 leading-tight">
                      {t('settings.animate_numbers_desc')}
                    </p>
                  </div>
                  <button
                    role="switch"
                    aria-checked={animateNumbers}
                    onClick={() => { selection(); setAnimateNumbers(!animateNumbers) }}
                    className={`relative shrink-0 w-11 h-6 rounded-full transition-colors duration-200 ${animateNumbers ? 'bg-accent' : 'bg-border'}`}
                  >
                    <span className={`absolute top-0.5 left-0.5 w-5 h-5 rounded-full bg-white shadow-sm transition-transform duration-200 ${animateNumbers ? 'translate-x-5' : 'translate-x-0'}`} />
                  </button>
                </div>
              </div>

              {/* Notifications section */}
              <div className="card-elevated divide-y divide-border">
                {([
                  { key: 'notify_budget_alerts',       icon: Bell,          labelKey: 'settings.notify_budget_alerts',       descKey: 'settings.notify_budget_alerts_desc' },
                  { key: 'notify_recurring_reminders', icon: RepeatOnce,    labelKey: 'settings.notify_recurring_reminders', descKey: 'settings.notify_recurring_reminders_desc' },
                  { key: 'notify_weekly_summary',      icon: CalendarCheck, labelKey: 'settings.notify_weekly_summary',      descKey: 'settings.notify_weekly_summary_desc' },
                  { key: 'notify_goal_milestones',     icon: Trophy,        labelKey: 'settings.notify_goal_milestones',     descKey: 'settings.notify_goal_milestones_desc' },
                ] as const).map(({ key, icon: Icon, labelKey, descKey }) => {
                  const checked = settings[key] ?? false
                  return (
                    <div key={key} className="flex items-center gap-4 px-4 py-3.5">
                      <div className="w-9 h-9 rounded-xl bg-accent/10 flex items-center justify-center shrink-0">
                        <Icon size={18} weight="fill" className="text-accent" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-[13px] font-semibold text-text leading-tight">{t(labelKey)}</p>
                        <p className="text-[11px] text-muted mt-0.5 leading-tight">{t(descKey)}</p>
                      </div>
                      <button
                        role="switch"
                        aria-checked={checked}
                        onClick={() => notifMutation.mutate({ [key]: !checked })}
                        disabled={notifMutation.isPending}
                        className={`relative shrink-0 w-11 h-6 rounded-full transition-colors duration-200 ${checked ? 'bg-accent' : 'bg-border'}`}
                      >
                        <span className={`absolute top-0.5 left-0.5 w-5 h-5 rounded-full bg-white shadow-sm transition-transform duration-200 ${checked ? 'translate-x-5' : 'translate-x-0'}`} />
                      </button>
                    </div>
                  )
                })}
              </div>

              {/* Language row */}
              <button
                onClick={() => setModal('language')}
                className="w-full card-elevated p-4 flex items-center gap-4 active:scale-[0.98] transition-transform text-left"
              >
                <div className="w-11 h-11 rounded-2xl bg-accent/10 flex items-center justify-center shrink-0">
                  <Globe size={22} weight="fill" className="text-accent" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-[11px] font-bold text-muted uppercase tracking-widest">
                    {t('settings.language')}
                  </p>
                  <p className="text-sm font-semibold text-text mt-0.5">
                    {currentLangData ? currentLangData.native : currentLang}
                  </p>
                </div>
                <CaretRight size={16} weight="bold" className="text-muted/40 shrink-0" />
              </button>

              {/* Danger zone */}
              <div className="pt-2">
                <p className="text-[11px] font-bold text-destructive/70 uppercase tracking-widest mb-2 px-1">
                  {t('settings.danger_zone')}
                </p>
                <button
                  onClick={() => setModal('reset-confirm')}
                  className="w-full card-elevated p-4 flex items-center gap-4 active:scale-[0.98] transition-transform text-left border border-destructive/20"
                >
                  <div className="w-11 h-11 rounded-2xl bg-destructive/10 flex items-center justify-center shrink-0">
                    <Trash size={22} weight="fill" className="text-destructive" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-[11px] font-bold text-destructive/70 uppercase tracking-widest">
                      {t('settings.reset_data')}
                    </p>
                    <p className="text-xs text-muted mt-0.5 leading-relaxed">
                      {t('settings.reset_data_desc')}
                    </p>
                  </div>
                </button>
              </div>

        </div>
      </PageTransition>

      {/* Language bottom sheet */}
      <BottomSheet
        open={modal === 'language'}
        onClose={closeModal}
        title={t('settings.language')}
      >
        <div className="divide-y divide-border pb-6">
          {LANGUAGES.map((lang) => {
            const isActive = currentLang === lang.code
            return (
              <button
                key={lang.code}
                onClick={() => selectLanguage(lang.code)}
                className={`w-full flex items-center gap-3 px-5 py-3.5 transition-colors active:bg-border text-left ${
                  isActive ? 'bg-accent-subtle' : ''
                }`}
              >
                <div className={`w-8 h-8 rounded-xl flex items-center justify-center shrink-0 ${isActive ? 'bg-accent/15' : 'bg-border'}`}>
                  <Globe size={16} weight="fill" className={isActive ? 'text-accent' : 'text-muted'} />
                </div>
                <span className="flex-1 text-[13px] font-semibold text-text">{lang.native}</span>
                {isActive && <Check size={16} weight="bold" className="text-accent shrink-0" />}
              </button>
            )
          })}
        </div>
      </BottomSheet>

      {/* Reset data confirmation dialog */}
      {createPortal(
        <AnimatePresence>
          {modal === 'reset-confirm' && (
            <>
              <motion.div
                className="fixed inset-0 bg-black/40 z-40"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                onClick={closeModal}
              />
              <motion.div
                className="fixed bottom-0 left-0 right-0 z-[60] bg-surface rounded-t-card px-5 pt-6 pb-8"
                initial={{ y: '100%' }}
                animate={{ y: 0 }}
                exit={{ y: '100%' }}
                transition={{ type: 'spring', stiffness: 350, damping: 35 }}
              >
                <div className="flex flex-col items-center text-center gap-4">
                  <div className="w-14 h-14 rounded-2xl bg-destructive/10 flex items-center justify-center">
                    <Trash size={28} weight="fill" className="text-destructive" />
                  </div>
                  <div>
                    <p className="text-base font-bold text-text">
                      {t('settings.reset_confirm_title')}
                    </p>
                    <p className="text-sm text-muted mt-2 leading-relaxed">
                      {t('settings.reset_confirm_desc')}
                    </p>
                  </div>
                  <div className="w-full flex flex-col gap-2 mt-2">
                    <button
                      onClick={() => resetMutation.mutate()}
                      disabled={resetMutation.isPending}
                      className="w-full py-3.5 rounded-2xl bg-destructive text-white font-bold text-sm disabled:opacity-50"
                    >
                      {resetMutation.isPending ? t('common.loading') : t('settings.reset_confirm_btn')}
                    </button>
                    <button
                      onClick={closeModal}
                      className="w-full py-3.5 rounded-2xl bg-surface text-muted font-semibold text-sm"
                    >
                      {t('common.cancel')}
                    </button>
                  </div>
                </div>
              </motion.div>
            </>
          )}
        </AnimatePresence>,
        document.body,
      )}
    </>
  )
}
