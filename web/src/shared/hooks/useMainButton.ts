import { useEffect } from 'react'

let mainButtonModule: {
  mainButton: {
    isMounted: () => boolean
    setText: (text: string) => void
    showLoader: () => void
    hideLoader: () => void
    enable: () => void
    disable: () => void
    show: () => void
    hide: () => void
    onClick: (cb: () => void) => () => void
  }
} | null = null

try {
  const mod = await import('@tma.js/sdk-react')
  mainButtonModule = { mainButton: mod.mainButton }
} catch {
  // SDK not available
}

interface Options {
  text: string
  onClick: () => void
  enabled?: boolean
  loading?: boolean
}

/** Controls the Telegram native Main Button (bottom blue button). */
export function useTgMainButton({ text, onClick, enabled = true, loading = false }: Options) {
  useEffect(() => {
    if (!mainButtonModule) return
    const mb = mainButtonModule.mainButton
    if (!mb.isMounted()) return
    try {
      mb.setText(text)

      if (loading) {
        mb.showLoader()
        mb.enable()
      } else if (enabled) {
        mb.hideLoader()
        mb.enable()
      } else {
        mb.hideLoader()
        mb.disable()
      }

      mb.show()
      const off = mb.onClick(onClick)

      return () => {
        off()
        mb.hide()
      }
    } catch {
      // mainButton not supported in this Telegram client
    }
  }, [text, onClick, enabled, loading])
}
