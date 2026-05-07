import { useReducedMotion } from 'framer-motion'
import { useUpdateSettings, useUserSettings } from './useUserSettings'

/**
 * Reactive hook: returns [enabled, setter].
 * The preference lives on the backend (user_settings.animate_numbers); a null
 * value means "no explicit choice" and the UI falls back to the OS
 * prefers-reduced-motion setting. An explicit user choice always wins.
 */
export function useAnimateNumbers(): [boolean, (b: boolean) => void] {
  const systemReduced = useReducedMotion()
  const { data } = useUserSettings()
  const update = useUpdateSettings()

  const stored = data?.animate_numbers ?? null
  const enabled = stored ?? !(systemReduced ?? false)

  function set(b: boolean): void {
    update.mutate({ ui_preferences: { animate_numbers: b } })
  }

  return [enabled, set]
}
