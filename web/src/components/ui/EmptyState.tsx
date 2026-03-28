import type { ReactNode } from 'react'

interface EmptyStateProps {
  icon?: string
  title: string
  description?: string
  action?: ReactNode
}

export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-2 py-12 text-center">
      {icon && <span className="text-4xl">{icon}</span>}
      <p className="text-sm font-medium text-muted">{title}</p>
      {description && <p className="text-xs text-muted/70">{description}</p>}
      {action && <div className="mt-2">{action}</div>}
    </div>
  )
}
