import type { ReactNode } from 'react'
import type { IconProps } from '@phosphor-icons/react'

type IconComponent = React.ForwardRefExoticComponent<IconProps & React.RefAttributes<SVGSVGElement>>

interface EmptyStateProps {
  icon?: IconComponent
  title: string
  description?: string
  action?: ReactNode
}

export function EmptyState({ icon: IconComponent, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-4 py-16 text-center">
      {IconComponent && (
        <div className="w-16 h-16 rounded-(--radius-btn) bg-accent-subtle flex items-center justify-center shadow-(--shadow-card)">
          <IconComponent size={32} weight="duotone" className="text-accent" />
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
