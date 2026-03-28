import type { ReactNode } from 'react'

interface CardProps {
  children: ReactNode
  className?: string
  padding?: string
}

export function Card({ children, className = '', padding = 'p-4' }: CardProps) {
  return (
    <div className={`bg-surface rounded-[--radius-card] overflow-hidden ${padding} ${className}`}>
      {children}
    </div>
  )
}
