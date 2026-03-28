import { useTranslation } from 'react-i18next'
import { AlertTriangle } from 'lucide-react'
import { Button } from './ui/Button'

interface Props {
  message?: string
  onRetry?: () => void
}

export function ErrorMessage({ message, onRetry }: Props) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col items-center justify-center gap-3 p-8 text-center">
      <AlertTriangle size={40} className="text-muted" />
      <p className="text-sm text-muted">{message ?? t('common.error')}</p>
      {onRetry && (
        <Button variant="primary" size="md" onClick={onRetry}>
          {t('common.retry')}
        </Button>
      )}
    </div>
  )
}
