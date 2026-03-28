import { useEffect } from 'react'
import {
  miniApp,
  themeParams,
  viewport,
  backButton,
  useSignal,
  retrieveRawInitData,
} from '@tma.js/sdk-react'

/** Initialise Telegram Mini App: expand, sync theme CSS vars. */
export function useTelegramApp() {
  const tpState = useSignal(themeParams.state)

  useEffect(() => {
    if (viewport.isStable() && !viewport.isExpanded()) {
      viewport.expand()
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

  return { miniApp, themeParams, viewport }
}

/** Show/hide the Telegram native Back Button and handle clicks. */
export function useTgBackButton(onBack: () => void, enabled = true) {
  useEffect(() => {
    if (!backButton.isAvailable()) return
    try {
      if (!enabled) {
        backButton.hide()
        return
      }
      backButton.show()
      const off = backButton.onClick(onBack)
      return () => {
        off()
        backButton.hide()
      }
    } catch {
      // backButton not supported in this Telegram client
    }
  }, [enabled, onBack])
}

/** Get Telegram initData raw string for API auth. */
export function getInitDataRaw(): string {
  try {
    return retrieveRawInitData() ?? ''
  } catch {
    return ''
  }
}
