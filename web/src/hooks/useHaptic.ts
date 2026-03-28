import { useCallback } from 'react'
import { hapticFeedback } from '@tma.js/sdk'

type ImpactStyle = 'light' | 'medium' | 'heavy' | 'rigid' | 'soft'
type NotificationType = 'error' | 'success' | 'warning'

function safe(fn: () => void) {
  try { fn() } catch { /* not in Telegram context */ }
}

export function useHaptic() {
  const impact = useCallback((style: ImpactStyle = 'light') => {
    safe(() => hapticFeedback.impactOccurred(style))
  }, [])

  const notification = useCallback((type: NotificationType) => {
    safe(() => hapticFeedback.notificationOccurred(type))
  }, [])

  const selection = useCallback(() => {
    safe(() => hapticFeedback.selectionChanged())
  }, [])

  return { impact, notification, selection }
}
