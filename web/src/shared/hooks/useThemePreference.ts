import type { ThemePref } from '../types'
import { useUserSettings, useUpdateSettings } from './useUserSettings'

const VALID: ReadonlySet<string> = new Set(['system', 'light', 'dark'])

function normalize(value: unknown): ThemePref {
  return typeof value === 'string' && VALID.has(value) ? (value as ThemePref) : 'system'
}

/**
 * Reactive hook reading and writing the user's UI theme preference.
 * Source of truth lives on the server in user_settings; updates use the
 * shared `ui_preferences` envelope.
 */
export function useThemePreference(): [ThemePref, (p: ThemePref) => void] {
  const { data } = useUserSettings()
  const update = useUpdateSettings()
  const pref = normalize(data?.theme)

  function set(next: ThemePref): void {
    if (next === pref) return
    update.mutate({ ui_preferences: { theme: next } })
  }

  return [pref, set]
}

export type { ThemePref } from '../types'
