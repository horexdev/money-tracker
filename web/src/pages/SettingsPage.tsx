import { useState, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Search } from 'lucide-react'
import { settingsApi } from '../api/settings'
import { Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { PageTransition } from '../components/PageTransition'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { Card, SectionHeader, Badge } from '../components/ui'

const POPULAR_CURRENCIES = ['USD', 'EUR', 'GBP', 'UAH', 'RUB', 'JPY', 'CNY', 'CAD', 'AUD', 'CHF']

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

export function SettingsPage() {
  const { t, i18n } = useTranslation()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const [search, setSearch] = useState('')

  useTgBackButton(() => navigate('/more'))

  const { data: settings, isLoading, isError, refetch } = useQuery({
    queryKey: ['settings'],
    queryFn: settingsApi.get,
  })

  const currencyMutation = useMutation({
    mutationFn: (currency: string) => settingsApi.update({ base_currency: currency }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['settings'] }),
  })

  const languageMutation = useMutation({
    mutationFn: (lang: string) => settingsApi.update({ language: lang }),
    onSuccess: (data) => {
      qc.invalidateQueries({ queryKey: ['settings'] })
      i18n.changeLanguage(data.language)
    },
  })

  const filtered = useMemo(() => {
    const q = search.toUpperCase().trim()
    if (!q) return ALL_CURRENCIES
    return ALL_CURRENCIES.filter((c) => c.includes(q))
  }, [search])

  if (isLoading) return <div className="flex justify-center py-16"><Spinner /></div>
  if (isError || !settings) return <ErrorMessage onRetry={refetch} />

  const selected = settings.base_currency
  const currentLang = settings.language || 'en'

  return (
    <PageTransition>
      <div className="py-4 flex flex-col gap-4">
        {/* Language section */}
        <div>
          <SectionHeader>{t('settings.language')}</SectionHeader>
          <Card className="mx-4" padding="p-0">
            {[
              { code: 'en', label: t('settings.english'), flag: '🇬🇧' },
              { code: 'ru', label: t('settings.russian'), flag: '🇷🇺' },
            ].map(lang => (
              <button
                key={lang.code}
                onClick={() => languageMutation.mutate(lang.code)}
                className={`
                  w-full flex items-center justify-between px-4 py-3 text-sm
                  border-b border-border last:border-b-0 transition-colors
                  ${currentLang === lang.code ? 'bg-accent-subtle' : 'active:bg-border'}
                `}
              >
                <span className="flex items-center gap-2">
                  <span>{lang.flag}</span>
                  <span className="text-text">{lang.label}</span>
                </span>
                {currentLang === lang.code && <span className="text-accent font-medium">✓</span>}
              </button>
            ))}
          </Card>
        </div>

        {/* Base currency section */}
        <div>
          <SectionHeader>{t('settings.base_currency')}</SectionHeader>
          <div className="flex flex-wrap gap-2 px-4 mb-3">
            {POPULAR_CURRENCIES.map((c) => (
              <button key={c} onClick={() => currencyMutation.mutate(c)}>
                <Badge variant={selected === c ? 'accent' : 'default'} className="cursor-pointer">
                  {c}
                </Badge>
              </button>
            ))}
          </div>

          <div className="px-4 mb-3">
            <div className="flex items-center gap-2 px-3 py-2 bg-surface rounded-[--radius-btn] focus-within:ring-2 focus-within:ring-accent transition-all">
              <Search size={16} className="text-muted shrink-0" />
              <input
                type="text"
                placeholder={t('common.search') + '...'}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="flex-1 bg-transparent text-sm text-text outline-none"
              />
              {search && (
                <button onClick={() => setSearch('')} className="text-muted text-sm">✕</button>
              )}
            </div>
          </div>

          <Card className="mx-4" padding="p-0">
            {filtered.length === 0 ? (
              <p className="py-8 text-center text-sm text-muted">{t('common.no_data')}</p>
            ) : (
              <div className="max-h-72 overflow-y-auto no-scrollbar">
                {filtered.map((c) => (
                  <button
                    key={c}
                    onClick={() => currencyMutation.mutate(c)}
                    className={`
                      w-full flex items-center justify-between px-4 py-3 text-sm text-left
                      text-text border-b border-border last:border-b-0 transition-colors
                      ${selected === c ? 'bg-accent-subtle' : 'active:bg-border'}
                    `}
                  >
                    <span>{c}</span>
                    {selected === c && <span className="text-accent font-medium">✓</span>}
                  </button>
                ))}
              </div>
            )}
          </Card>
        </div>

        {(currencyMutation.isPending || languageMutation.isPending) && (
          <div className="flex justify-center"><Spinner size="sm" /></div>
        )}

        {/* Display currencies info */}
        {settings.display_currencies?.length > 0 && (
          <div>
            <SectionHeader>{t('settings.display_currencies')}</SectionHeader>
            <Card className="mx-4">
              <div className="flex gap-2 flex-wrap">
                {settings.display_currencies.map((c) => (
                  <Badge key={c} variant="accent">{c}</Badge>
                ))}
              </div>
            </Card>
          </div>
        )}
      </div>
    </PageTransition>
  )
}
