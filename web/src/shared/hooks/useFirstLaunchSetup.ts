import { useEffect, useRef } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { settingsApi } from '../api/settings'
import type { UserSettings } from '../types'

// Default currency for each supported language — used only on first launch
// to set the initial base currency (which determines the default account currency).
// Language and currency are independent after first launch.
const LANG_DEFAULT_CURRENCY: Record<string, string> = {
  en: 'USD', ru: 'RUB', uk: 'UAH', be: 'BYN',
  kk: 'KZT', uz: 'UZS', tr: 'TRY', ar: 'SAR',
  es: 'EUR', pt: 'BRL', fr: 'EUR', de: 'EUR',
  it: 'EUR', nl: 'EUR', ko: 'KRW', ms: 'MYR', id: 'IDR',
}

export const FIRST_LAUNCH_KEY = 'money_tracker_first_launch_done'

/**
 * On first launch: detect language from Telegram / browser and persist
 * language + a sensible default currency to the backend.
 * The currency is only used to create the initial default account — it is
 * never changed automatically again. Language is changed only by the user
 * via Settings. Runs once per device — guarded by localStorage flag.
 */
export function useFirstLaunchSetup(settings: UserSettings | undefined) {
  const { i18n } = useTranslation()
  const qc = useQueryClient()
  const ran = useRef(false)

  useEffect(() => {
    if (ran.current) return
    if (!settings) return

    ran.current = true

    // Always sync the UI language to whatever the backend has persisted.
    // This handles the case where DEV_LANG or Telegram language differs from browser language.
    if (settings.language && i18n.language?.split('-')[0] !== settings.language) {
      i18n.changeLanguage(settings.language)
    }

    if (localStorage.getItem(FIRST_LAUNCH_KEY)) return
    localStorage.setItem(FIRST_LAUNCH_KEY, '1')

    const effectiveLang = settings.language || i18n.language?.split('-')[0] || 'en'
    const defaultCurrency = LANG_DEFAULT_CURRENCY[effectiveLang] ?? 'USD'
    const needsCurrencyUpdate = settings.base_currency !== defaultCurrency

    if (!needsCurrencyUpdate) return

    settingsApi.update({ base_currency: defaultCurrency }).then((updated) => {
      qc.setQueryData(['settings'], updated)
    })
  }, [settings, i18n, qc])
}
