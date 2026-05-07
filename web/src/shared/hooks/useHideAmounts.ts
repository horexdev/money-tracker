import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { settingsApi } from '../api/settings'
import type { UserSettings } from '../types'

interface HideAmountsState {
  hidden: boolean
  toggle: () => void
  setHidden: (hide: boolean) => void
  isLoading: boolean
}

/**
 * Reactive hook for the privacy-mode flag (`hide_amounts`).
 * Server is the source of truth. While settings are still loading we
 * default to `hidden=true` (privacy-first) — the alternative would be
 * to flash the real amounts before the response arrives.
 */
export function useHideAmounts(): HideAmountsState {
  const qc = useQueryClient()
  const { data, isLoading } = useQuery({ queryKey: ['settings'], queryFn: settingsApi.get })
  const hidden = data === undefined ? true : !!data.hide_amounts

  const mutation = useMutation({
    mutationFn: (hide: boolean) => settingsApi.update({ hide_amounts: hide }),
    onMutate: async (next) => {
      await qc.cancelQueries({ queryKey: ['settings'] })
      const previous = qc.getQueryData<UserSettings>(['settings'])
      if (previous) {
        qc.setQueryData<UserSettings>(['settings'], { ...previous, hide_amounts: next })
      }
      return { previous }
    },
    onError: (_err, _next, ctx) => {
      if (ctx?.previous) qc.setQueryData<UserSettings>(['settings'], ctx.previous)
    },
    onSettled: () => qc.invalidateQueries({ queryKey: ['settings'] }),
  })

  return {
    hidden,
    toggle: () => mutation.mutate(!hidden),
    setHidden: (hide: boolean) => mutation.mutate(hide),
    isLoading,
  }
}
