import { useState } from 'react'
import { useReducedMotion } from 'framer-motion'

export const ANIMATE_NUMBERS_KEY = 'app_animate_numbers'

function getStored(): boolean | null {
  try {
    const v = localStorage.getItem(ANIMATE_NUMBERS_KEY)
    if (v === 'true') return true
    if (v === 'false') return false
  } catch { /* ignore */ }
  return null
}

export function setAnimateNumbersPreference(enabled: boolean): void {
  try {
    localStorage.setItem(ANIMATE_NUMBERS_KEY, String(enabled))
  } catch { /* ignore */ }
}

/** Reactive hook: returns [enabled, setter].
 *  Default: honor system `prefers-reduced-motion` when no explicit choice stored.
 *  Explicit user choice always wins over OS preference.
 */
export function useAnimateNumbers(): [boolean, (b: boolean) => void] {
  const systemReduced = useReducedMotion()
  const [stored, setStored] = useState<boolean | null>(getStored)

  const enabled = stored ?? !(systemReduced ?? false)

  const set = (b: boolean) => {
    setStored(b)
    setAnimateNumbersPreference(b)
  }

  return [enabled, set]
}
