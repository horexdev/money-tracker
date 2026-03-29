import { useState, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { motion, AnimatePresence } from 'framer-motion'
import { Check, MagnifyingGlass, X, Globe, CurrencyDollar, CaretRight, Warning, Trash } from '@phosphor-icons/react'
import { settingsApi } from '../api/settings'
import { balanceApi } from '../api/balance'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { useHaptic } from '../hooks/useHaptic'

const LANGUAGES = [
  { code: 'en', label: 'English', native: 'English', flag: '🇬🇧' },
  { code: 'ru', label: 'Russian', native: 'Русский', flag: '🇷🇺' },
  { code: 'uk', label: 'Ukrainian', native: 'Українська', flag: '🇺🇦' },
  { code: 'be', label: 'Belarusian', native: 'Беларуская', flag: '🇧🇾' },
  { code: 'kk', label: 'Kazakh', native: 'Қазақша', flag: '🇰🇿' },
  { code: 'uz', label: 'Uzbek', native: "O'zbek", flag: '🇺🇿' },
  { code: 'tr', label: 'Turkish', native: 'Türkçe', flag: '🇹🇷' },
  { code: 'ar', label: 'Arabic', native: 'العربية', flag: '🇸🇦' },
  { code: 'es', label: 'Spanish', native: 'Español', flag: '🇪🇸' },
  { code: 'pt', label: 'Portuguese', native: 'Português', flag: '🇧🇷' },
  { code: 'fr', label: 'French', native: 'Français', flag: '🇫🇷' },
  { code: 'de', label: 'German', native: 'Deutsch', flag: '🇩🇪' },
  { code: 'it', label: 'Italian', native: 'Italiano', flag: '🇮🇹' },
  { code: 'nl', label: 'Dutch', native: 'Nederlands', flag: '🇳🇱' },
  { code: 'ko', label: 'Korean', native: '한국어', flag: '🇰🇷' },
  { code: 'ms', label: 'Malay', native: 'Bahasa Melayu', flag: '🇲🇾' },
  { code: 'id', label: 'Indonesian', native: 'Bahasa Indonesia', flag: '🇮🇩' },
]

const POPULAR_CURRENCIES = ['USD', 'EUR', 'GBP', 'UAH', 'RUB', 'TRY', 'KZT', 'UZS', 'BRL', 'JPY']


const ALL_CURRENCIES = [
  'AED','AFN','ALL','AMD','ANG','AOA','ARS','AUD','AWG','AZN',
  'BAM','BBD','BDT','BGN','BHD','BMD','BND','BOB','BRL','BSD',
  'BWP','BYN','BZD','CAD','CDF','CHF','CLP','CNY','COP','CRC',
  'CUP','CVE','CZK','DJF','DKK','DOP','DZD','EGP','ETB','EUR',
  'FJD','GBP','GEL','GHS','GMD','GTQ','GYD','HKD','HNL','HRK',
  'HTG','HUF','IDR','ILS','INR','IQD','IRR','ISK','JMD','JOD',
  'JPY','KES','KGS','KHR','KRW','KWD','KZT','LAK','LBP','LKR',
  'LYD','MAD','MDL','MKD','MMK','MNT','MOP','MRU','MUR','MVR',
  'MWK','MXN','MYR','MZN','NAD','NGN','NIO','NOK','NPR','NZD',
  'OMR','PAB','PEN','PGK','PHP','PKR','PLN','PYG','QAR','RON',
  'RSD','RUB','RWF','SAR','SBD','SCR','SDG','SEK','SGD','SLL',
  'SOS','SRD','SZL','THB','TJS','TMT','TND','TOP','TRY','TTD',
  'TWD','TZS','UAH','UGX','USD','UYU','UZS','VES','VND','VUV',
  'WST','XAF','XCD','XOF','XPF','YER','ZAR','ZMW',
]

type Modal = 'none' | 'language' | 'currency' | 'currency-confirm' | 'reset-confirm'

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
  return (
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
    </AnimatePresence>
  )
}

export function SettingsPage() {
  const { t, i18n } = useTranslation()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { selection, notification } = useHaptic()
  const [modal, setModal] = useState<Modal>('none')
  const [currencySearch, setCurrencySearch] = useState('')
  const [pendingCurrency, setPendingCurrency] = useState<string | null>(null)

  useTgBackButton(() => navigate('/more'))

  const { data: settings, isLoading, isError, refetch } = useQuery({
    queryKey: ['settings'],
    queryFn: settingsApi.get,
  })

  const { data: balance } = useQuery({
    queryKey: ['balance'],
    queryFn: balanceApi.get,
  })

  const currencyMutation = useMutation({
    mutationFn: (currency: string) => settingsApi.update({ base_currency: currency }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['settings'] })
      qc.invalidateQueries({ queryKey: ['balance'] })
      notification('success')
    },
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

  const filteredCurrencies = useMemo(() => {
    const q = currencySearch.toUpperCase().trim()
    if (!q) return ALL_CURRENCIES
    return ALL_CURRENCIES.filter((c) => c.includes(q))
  }, [currencySearch])

  if (isLoading) return <div className="flex justify-center py-16"><Spinner /></div>
  if (isError || !settings) return <ErrorMessage onRetry={refetch} />

  const selectedCurrency = settings.base_currency
  const currentLang = settings.language || i18n.language || 'en'
  const currentLangData = LANGUAGES.find(l => l.code === currentLang)

  function selectLanguage(code: string) {
    selection()
    languageMutation.mutate(code)
    setModal('none')
  }

  // Currencies with existing transactions (excluding current base currency)
  const otherCurrencies = (balance?.by_currency ?? []).filter(
    (b) => b.currency_code !== settings?.base_currency
  )

  function selectCurrency(code: string) {
    selection()
    // If user has transactions in other currencies, show confirmation first
    if (otherCurrencies.length > 0 && code !== settings?.base_currency) {
      setPendingCurrency(code)
      setModal('currency-confirm')
    } else {
      currencyMutation.mutate(code)
      setModal('none')
      setCurrencySearch('')
    }
  }

  function confirmCurrencyChange() {
    if (pendingCurrency) {
      currencyMutation.mutate(pendingCurrency)
    }
    setModal('none')
    setCurrencySearch('')
    setPendingCurrency(null)
  }

  function closeModal() {
    setModal('none')
    setCurrencySearch('')
    setPendingCurrency(null)
  }

  return (
    <PageTransition>
      <div className="flex flex-col h-[calc(100dvh-var(--tab-bar-h))]">
        <div className="flex-1 min-h-0 overflow-y-auto no-scrollbar pb-4">
          <div className="px-4 pt-3 space-y-3">

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
                  {currentLangData ? `${currentLangData.flag} ${currentLangData.native}` : currentLang}
                </p>
              </div>
              <CaretRight size={16} weight="bold" className="text-muted/40 shrink-0" />
            </button>

            {/* Currency row */}
            <button
              onClick={() => setModal('currency')}
              className="w-full card-elevated p-4 flex items-center gap-4 active:scale-[0.98] transition-transform text-left"
            >
              <div className="w-11 h-11 rounded-2xl bg-income/10 flex items-center justify-center shrink-0">
                <CurrencyDollar size={22} weight="fill" className="text-income" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-[11px] font-bold text-muted uppercase tracking-widest">
                  {t('settings.base_currency')}
                </p>
                <p className="text-sm font-semibold text-text mt-0.5">
                  {selectedCurrency}
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
        </div>
      </div>

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
                <span className="text-xl">{lang.flag}</span>
                <span className="flex-1 text-[13px] font-semibold text-text">{lang.native}</span>
                {isActive && <Check size={16} weight="bold" className="text-accent shrink-0" />}
              </button>
            )
          })}
        </div>
      </BottomSheet>

      {/* Currency bottom sheet */}
      <BottomSheet
        open={modal === 'currency'}
        onClose={closeModal}
        title={t('settings.base_currency')}
      >
        <div className="px-4 pb-4 space-y-3">
          {/* Popular */}
          <div className="flex flex-wrap gap-2 pt-1 pb-1">
            {POPULAR_CURRENCIES.map((c) => {
              const isActive = selectedCurrency === c
              return (
                <button
                  key={c}
                  onClick={() => selectCurrency(c)}
                  className={`
                    px-4 py-2 rounded-full text-xs font-bold transition-all duration-150 select-none
                    ${isActive
                      ? 'bg-accent text-accent-text shadow-[0_2px_8px_rgba(99,102,241,0.4)]'
                      : 'bg-surface text-muted shadow-sm active:scale-95'
                    }
                  `}
                >
                  {c}
                </button>
              )
            })}
          </div>

          {/* Search */}
          <div className="relative">
            <MagnifyingGlass size={14} weight="bold" className="absolute left-3.5 top-1/2 -translate-y-1/2 text-muted" />
            <input
              type="text"
              value={currencySearch}
              onChange={(e) => setCurrencySearch(e.target.value)}
              placeholder={`${t('common.search')}...`}
              className="w-full bg-surface rounded-2xl pl-9 pr-9 py-2.5 text-xs font-medium outline-none text-text placeholder:text-muted/50 shadow-sm focus:shadow-[0_0_0_2px_rgba(99,102,241,0.2)] transition-shadow"
            />
            {currencySearch && (
              <button
                onClick={() => setCurrencySearch('')}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted"
              >
                <X size={12} weight="bold" />
              </button>
            )}
          </div>

          {/* All currencies list */}
          <div className="card-elevated divide-y divide-border overflow-hidden rounded-2xl">
            {filteredCurrencies.length === 0 ? (
              <p className="py-8 text-center text-sm text-muted">{t('common.no_data')}</p>
            ) : (
              filteredCurrencies.map((c) => {
                const isActive = selectedCurrency === c
                return (
                  <button
                    key={c}
                    onClick={() => selectCurrency(c)}
                    className={`w-full flex items-center justify-between px-4 py-3 text-left transition-colors active:bg-border ${
                      isActive ? 'bg-accent-subtle' : ''
                    }`}
                  >
                    <span className="text-[13px] font-semibold text-text">{c}</span>
                    {isActive && <Check size={14} weight="bold" className="text-accent" />}
                  </button>
                )
              })
            )}
          </div>
        </div>
      </BottomSheet>

      {/* Currency change confirmation dialog */}
      <AnimatePresence>
        {modal === 'currency-confirm' && (
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
                <div className="w-14 h-14 rounded-2xl bg-brand-gold/10 flex items-center justify-center">
                  <Warning size={28} weight="fill" className="text-brand-gold" />
                </div>
                <div>
                  <p className="text-base font-bold text-text">
                    {t('settings.currency_change_title', { currency: pendingCurrency })}
                  </p>
                  <p className="text-sm text-muted mt-2 leading-relaxed">
                    {t('settings.currency_change_desc', {
                      count: otherCurrencies.length,
                      currencies: otherCurrencies.map((b) => b.currency_code).join(', '),
                      newCurrency: pendingCurrency,
                    })}
                  </p>
                </div>
                <div className="w-full flex flex-col gap-2 mt-2">
                  <button
                    onClick={confirmCurrencyChange}
                    className="w-full py-3.5 rounded-2xl bg-accent text-accent-text font-bold text-sm"
                  >
                    {t('settings.currency_change_confirm')}
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
      </AnimatePresence>

      {/* Reset data confirmation dialog */}
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
                    {resetMutation.isPending ? '...' : t('settings.reset_confirm_btn')}
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
      </AnimatePresence>

    </PageTransition>
  )
}
