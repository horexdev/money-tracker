import { useEffect } from 'react'

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

/** Initialise Telegram Mini App: expand, sync theme CSS vars. */
export function useTelegramApp() {
  const tpState = sdkLoaded ? sdkModules!.useSignal(sdkModules!.themeParams.state) : undefined

  useEffect(() => {
    if (!sdkLoaded) return
    try {
      const vp = sdkModules!.viewport
      if (vp.isStable() && !vp.isExpanded()) {
        vp.expand()
      }
    } catch { /* ignore */ }
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

/** Get Telegram initData raw string for API auth. */
export function getInitDataRaw(): string {
  if (!sdkLoaded) return ''
  try {
    return sdkModules!.retrieveRawInitData() ?? ''
  } catch {
    return ''
  }
}
