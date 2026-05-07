import { useMutation, useQueryClient } from '@tanstack/react-query'
import { templatesApi } from '../api/templates'
import { useHaptic } from './useHaptic'

interface ApplyArgs {
  templateId: number
  amountCents?: number
}

/**
 * useApplyTemplate creates a transaction from a template and invalidates the
 * same query keys as a manual transaction would (transactions, balance,
 * accounts, stats), then plays a success haptic.
 */
export function useApplyTemplate() {
  const qc = useQueryClient()
  const { notification } = useHaptic()

  return useMutation({
    mutationFn: ({ templateId, amountCents }: ApplyArgs) =>
      templatesApi.apply(templateId, amountCents),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['transactions'] })
      qc.invalidateQueries({ queryKey: ['balance'] })
      qc.invalidateQueries({ queryKey: ['accounts'] })
      qc.invalidateQueries({ queryKey: ['stats'] })
      notification('success')
    },
    onError: () => notification('error'),
  })
}
