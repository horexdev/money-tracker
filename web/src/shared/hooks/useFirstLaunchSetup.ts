import { useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import type { UserSettings } from '../types'

export const FIRST_LAUNCH_KEY = 'money_tracker_first_launch_done'

/**
 * On first launch: sync the UI language to the backend-persisted setting.
 * Currency is no longer set here — it is derived from the default account,
 * which is created by ensureUser on the backend with the correct
 * locale-based currency.
 */
export function useFirstLaunchSetup(settings: UserSettings | undefined) {
  const { i18n } = useTranslation()
  const ran = useRef(false)

  useEffect(() => {
    if (ran.current) return
    if (!settings) return

    ran.current = true

    // Always sync the UI language to whatever the backend has persisted.
    if (settings.language && i18n.language?.split('-')[0] !== settings.language) {
      i18n.changeLanguage(settings.language)
    }

    if (!localStorage.getItem(FIRST_LAUNCH_KEY)) {
      localStorage.setItem(FIRST_LAUNCH_KEY, '1')
    }
  }, [settings, i18n])
}
