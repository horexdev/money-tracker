import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { settingsApi } from '../api/settings'
import type { UserSettings, ThemePref } from '../types'

export type { ThemePref } from '../types'

const VALID: ReadonlySet<string> = new Set(['system', 'light', 'dark'])

function normalize(value: unknown): ThemePref {
  return typeof value === 'string' && VALID.has(value) ? (value as ThemePref) : 'system'
}

/**
 * Reactive hook reading and writing the user's UI theme preference.
 * Source of truth lives on the server in user_settings; updates use
 * react-query's optimistic-update pattern so the toggle feels instant.
 */
export function useThemePreference(): [ThemePref, (p: ThemePref) => void] {
  const qc = useQueryClient()
  const { data } = useQuery({ queryKey: ['settings'], queryFn: settingsApi.get })
  const pref = normalize(data?.theme)

  const mutation = useMutation({
    mutationFn: (p: ThemePref) => settingsApi.update({ theme: p }),
    onMutate: async (next) => {
      await qc.cancelQueries({ queryKey: ['settings'] })
      const previous = qc.getQueryData<UserSettings>(['settings'])
      if (previous) {
        qc.setQueryData<UserSettings>(['settings'], { ...previous, theme: next })
      }
      return { previous }
    },
    onError: (_err, _next, ctx) => {
      if (ctx?.previous) qc.setQueryData<UserSettings>(['settings'], ctx.previous)
    },
    onSettled: () => qc.invalidateQueries({ queryKey: ['settings'] }),
  })

  return [pref, (p: ThemePref) => mutation.mutate(p)]
}
