import { useState } from 'react'
import { createPortal } from 'react-dom'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { motion, AnimatePresence } from 'framer-motion'
import { Check, X, Globe, CaretRight, Trash } from '@phosphor-icons/react'
import { settingsApi } from '../api/settings'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { useHaptic } from '../hooks/useHaptic'
import { FIRST_LAUNCH_KEY } from '../hooks/useFirstLaunchSetup'

const LANGUAGES = [
  { code: 'en', label: 'English', native: 'English' },
  { code: 'ru', label: 'Russian', native: 'Русский' },
  { code: 'uk', label: 'Ukrainian', native: 'Українська' },
  { code: 'be', label: 'Belarusian', native: 'Беларуская' },
  { code: 'kk', label: 'Kazakh', native: 'Қазақша' },
  { code: 'uz', label: 'Uzbek', native: "O'zbek" },
  { code: 'tr', label: 'Turkish', native: 'Türkçe' },
  { code: 'ar', label: 'Arabic', native: 'العربية' },
  { code: 'es', label: 'Spanish', native: 'Español' },
  { code: 'pt', label: 'Portuguese', native: 'Português' },
  { code: 'fr', label: 'French', native: 'Français' },
  { code: 'de', label: 'German', native: 'Deutsch' },
  { code: 'it', label: 'Italian', native: 'Italiano' },
  { code: 'nl', label: 'Dutch', native: 'Nederlands' },
  { code: 'ko', label: 'Korean', native: '한국어' },
  { code: 'ms', label: 'Malay', native: 'Bahasa Melayu' },
  { code: 'id', label: 'Indonesian', native: 'Bahasa Indonesia' },
]

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
