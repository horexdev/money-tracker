import { useEffect } from 'react'
import { mainButton } from '@tma.js/sdk-react'

interface Options {
  text: string
  onClick: () => void
  enabled?: boolean
  loading?: boolean
}

/** Controls the Telegram native Main Button (bottom blue button). */
export function useTgMainButton({ text, onClick, enabled = true, loading = false }: Options) {
  useEffect(() => {
    mainButton.setText(text)

    if (loading) {
      mainButton.showLoader()
      mainButton.enable()
    } else if (enabled) {
      mainButton.hideLoader()
      mainButton.enable()
    } else {
      mainButton.hideLoader()
      mainButton.disable()
    }

    mainButton.show()
    const off = mainButton.onClick(onClick)

    return () => {
      off()
      mainButton.hide()
    }
  }, [text, onClick, enabled, loading])
}
