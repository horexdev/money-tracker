import { useUserSettings, useUpdateSettings } from './useUserSettings'

interface HideAmountsState {
  hidden: boolean
  toggle: () => void
  setHidden: (hide: boolean) => void
  isLoading: boolean
}

/**
 * Reactive hook for the privacy-mode flag (`hide_amounts`).
 * Server is the source of truth. While settings are still loading we default
 * to `hidden=true` (privacy-first) — the alternative would be to flash the
 * real amounts before the response arrives.
 */
export function useHideAmounts(): HideAmountsState {
  const { data, isLoading } = useUserSettings()
  const update = useUpdateSettings()
  const hidden = data === undefined ? true : !!data.hide_amounts

  function setHidden(next: boolean): void {
    update.mutate({ ui_preferences: { hide_amounts: next } })
  }

  return {
    hidden,
    toggle: () => setHidden(!hidden),
    setHidden,
    isLoading,
  }
}
