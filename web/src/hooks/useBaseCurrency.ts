import { useQuery } from '@tanstack/react-query'
import { settingsApi } from '../api/settings'
import { getCurrencySymbol } from '../lib/money'

export function useBaseCurrency() {
  const { data } = useQuery({
    queryKey: ['settings'],
    queryFn: settingsApi.get,
    staleTime: 60_000,
  })
  const code = data?.base_currency ?? 'USD'
  return { code, symbol: getCurrencySymbol(code) }
}
