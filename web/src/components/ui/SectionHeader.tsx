import type { ReactNode } from 'react'

interface SectionHeaderProps {
  children: ReactNode
  action?: ReactNode
  className?: string
}

export function SectionHeader({ children, action, className = '' }: SectionHeaderProps) {
  return (
    <div className={`flex items-center justify-between px-4 mb-2 ${className}`}>
      <span className="text-xs font-semibold uppercase tracking-widest text-muted">{children}</span>
      {action}
    </div>
  )
}
