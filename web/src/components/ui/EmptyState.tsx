import type { ReactNode } from 'react'

interface EmptyStateProps {
  icon?: string
  title: string
  description?: string
  action?: ReactNode
}

export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-4 py-16 text-center">
      {icon && (
        <div className="w-16 h-16 rounded-[20px] bg-accent-subtle flex items-center justify-center text-[32px] shadow-[0_2px_12px_rgba(0,0,0,0.04)]">
          {icon}
        </div>
      )}
      <div className="space-y-1.5">
        <p className="text-sm font-bold text-text">{title}</p>
        {description && <p className="text-xs text-muted max-w-[200px] mx-auto">{description}</p>}
      </div>
      {action && <div className="mt-1">{action}</div>}
    </div>
  )
}
