import { useEffect, useState } from 'react'

let sdkLoaded = false
let sdkModules: {
  miniApp: unknown
  themeParams: { state: unknown }
  viewport: { isStable: () => boolean; isExpanded: () => boolean; expand: () => void }
  backButton: { isSupported: () => boolean; show: () => void; hide: () => void; onClick: (cb: () => void) => () => void }
  useSignal: (signal: unknown) => Record<string, string | undefined> | undefined
  retrieveRawInitData: () => string | undefined
} | null = null

try {
  const mod = await import('@tma.js/sdk-react')
  sdkModules = {
    miniApp: mod.miniApp,
    themeParams: mod.themeParams,
    viewport: mod.viewport,
    backButton: mod.backButton,
    useSignal: mod.useSignal as (signal: unknown) => Record<string, string | undefined> | undefined,
    retrieveRawInitData: mod.retrieveRawInitData,
  }
  sdkLoaded = true
} catch {
  // SDK not available outside Telegram
}

export type ThemePref = 'system' | 'light' | 'dark'
export const THEME_KEY = 'app_theme'

function getStoredTheme(): ThemePref {
  try {
    const v = localStorage.getItem(THEME_KEY)
    if (v === 'light' || v === 'dark' || v === 'system') return v
  } catch { /* ignore */ }
  return 'system'
}

/** Applies the theme to document root, respecting localStorage override. */
function applyThemeGlobal(tgScheme: string) {
  const pref = getStoredTheme()
  if (pref === 'light' || pref === 'dark') {
    document.documentElement.setAttribute('data-theme', pref)
  } else {
    document.documentElement.setAttribute('data-theme', tgScheme)
  }
}

/** Write theme preference to localStorage and immediately apply it. */
export function setThemePreference(pref: ThemePref, tgScheme = 'light') {
  try {
    localStorage.setItem(THEME_KEY, pref)
  } catch { /* ignore */ }
  applyThemeGlobal(tgScheme)
}

/** Reactive hook: returns [current pref, setter]. */
export function useThemePreference(): [ThemePref, (p: ThemePref) => void] {
  const [pref, setPref] = useState<ThemePref>(getStoredTheme)

  const set = (p: ThemePref) => {
    setPref(p)
    setThemePreference(p)
  }

  return [pref, set]
}

/** Initialise Telegram Mini App: expand, sync theme CSS vars. */
export function useTelegramApp() {
  const tpState = sdkLoaded ? sdkModules!.useSignal(sdkModules!.themeParams.state) : undefined

  useEffect(() => {
    type TgWebApp = {
      colorScheme?: string
      contentSafeAreaInset?: { top?: number }
      safeAreaInset?: { top?: number }
      onEvent?: (event: string, cb: () => void) => void
      offEvent?: (event: string, cb: () => void) => void
    }
    const tg = (((window as unknown) as Record<string, unknown>).Telegram as Record<string, unknown>)
      ?.WebApp as TgWebApp | undefined

    // Apply color scheme from Telegram, respecting localStorage override
    function applyTheme() {
      applyThemeGlobal(tg?.colorScheme ?? 'light')
    }

    // Re-apply safe area in case insets changed after initial render
    function applySafeTop() {
      try {
        const top =
          (tg?.contentSafeAreaInset?.top ?? 0) +
          (tg?.safeAreaInset?.top ?? 0)
        document.documentElement.style.setProperty('--safe-top', top > 0 ? `${top}px` : '0px')
      } catch { /* ignore */ }
    }

    applyTheme()
    applySafeTop()
    tg?.onEvent?.('themeChanged', applyTheme)
    tg?.onEvent?.('safeAreaChanged', applySafeTop)
    tg?.onEvent?.('contentSafeAreaChanged', applySafeTop)
    return () => {
      tg?.offEvent?.('themeChanged', applyTheme)
      tg?.offEvent?.('safeAreaChanged', applySafeTop)
      tg?.offEvent?.('contentSafeAreaChanged', applySafeTop)
    }
  }, [])

  useEffect(() => {
    if (!tpState) return
    const root = document.documentElement
    const map: Record<string, string | undefined> = {
      '--tg-theme-bg-color':                  tpState.bgColor,
      '--tg-theme-secondary-bg-color':        tpState.secondaryBgColor,
      '--tg-theme-text-color':                tpState.textColor,
      '--tg-theme-hint-color':                tpState.hintColor,
      '--tg-theme-link-color':                tpState.linkColor,
      '--tg-theme-button-color':              tpState.buttonColor,
      '--tg-theme-button-text-color':         tpState.buttonTextColor,
      '--tg-theme-header-bg-color':           tpState.headerBgColor,
      '--tg-theme-accent-text-color':         tpState.accentTextColor,
      '--tg-theme-section-bg-color':          tpState.sectionBgColor,
      '--tg-theme-section-header-text-color': tpState.sectionHeaderTextColor,
      '--tg-theme-subtitle-text-color':       tpState.subtitleTextColor,
      '--tg-theme-destructive-text-color':    tpState.destructiveTextColor,
    }
    for (const [prop, value] of Object.entries(map)) {
      if (value) root.style.setProperty(prop, value)
    }
  }, [tpState])

  return { miniApp: sdkModules?.miniApp, themeParams: sdkModules?.themeParams, viewport: sdkModules?.viewport }
}

/** Show/hide the Telegram native Back Button and handle clicks. */
export function useTgBackButton(onBack: () => void, enabled = true) {
  useEffect(() => {
    if (!sdkLoaded) return
    const bb = sdkModules!.backButton
    if (!bb.isSupported()) return
    try {
      if (!enabled) {
        bb.hide()
        return
      }
      bb.show()
      const off = bb.onClick(onBack)
      return () => {
        off()
        bb.hide()
      }
    } catch {
      // backButton not supported in this Telegram client
    }
  }, [enabled, onBack])
}

/** Get Telegram initData raw string for API auth.
 *
 * Outside Telegram in dev mode, returns a "dev:<user_id>" token accepted by
 * the backend when DEV_MODE=true, bypassing HMAC validation.
 */
export function getInitDataRaw(): string {
  if (sdkLoaded) {
    try {
      const raw = sdkModules!.retrieveRawInitData() ?? ''
      if (raw) return raw
    } catch {
      // fall through to dev bypass below
    }
  }
  if (import.meta.env.DEV) {
    const devUserID = import.meta.env.VITE_DEV_USER_ID ?? '6554524765'
    return `dev:${devUserID}`
  }
  return ''
}
