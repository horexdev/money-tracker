import { useCallback } from 'react'

type ImpactStyle = 'light' | 'medium' | 'heavy' | 'rigid' | 'soft'
type NotificationType = 'error' | 'success' | 'warning'

let haptic: {
  impactOccurred: (style: ImpactStyle) => void
  notificationOccurred: (type: NotificationType) => void
  selectionChanged: () => void
} | null = null

try {
  const { hapticFeedback } = await import('@tma.js/sdk')
  haptic = hapticFeedback
} catch {
  // SDK not available
}

function safe(fn: () => void) {
  try { fn() } catch { /* not in Telegram context */ }
}

export function useHaptic() {
  const impact = useCallback((style: ImpactStyle = 'light') => {
    if (haptic) safe(() => haptic!.impactOccurred(style))
  }, [])

  const notification = useCallback((type: NotificationType) => {
    if (haptic) safe(() => haptic!.notificationOccurred(type))
  }, [])

  const selection = useCallback(() => {
    if (haptic) safe(() => haptic!.selectionChanged())
  }, [])

  return { impact, notification, selection }
}
