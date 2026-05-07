import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { settingsApi, type SettingsUpdatePayload } from '../api/settings'
import type { UserSettings } from '../types'

export const SETTINGS_QUERY_KEY = ['settings'] as const

export function useUserSettings() {
  return useQuery<UserSettings>({
    queryKey: SETTINGS_QUERY_KEY,
    queryFn: settingsApi.get,
    staleTime: 60_000,
  })
}

export function useUpdateSettings() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (patch: SettingsUpdatePayload) => settingsApi.update(patch),
    onSuccess: (data) => {
      qc.setQueryData(SETTINGS_QUERY_KEY, data)
    },
  })
}
