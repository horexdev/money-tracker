import { useTranslation } from 'react-i18next'
import { Warning } from '@phosphor-icons/react'
import { Button } from './ui/Button'

interface Props {
  message?: string
  onRetry?: () => void
}

export function ErrorMessage({ message, onRetry }: Props) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col items-center justify-center gap-4 p-8 text-center">
      <div className="w-16 h-16 rounded-[20px] bg-expense-subtle flex items-center justify-center shadow-[0_2px_12px_rgba(0,0,0,0.04)]">
        <Warning size={28} weight="fill" className="text-expense" />
      </div>
      <p className="text-sm font-medium text-muted">{message ?? t('common.error')}</p>
      {onRetry && (
        <Button variant="primary" size="md" onClick={onRetry}>
          {t('common.retry')}
        </Button>
      )}
    </div>
  )
}
