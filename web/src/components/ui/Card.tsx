import type { ReactNode } from 'react'

interface CardProps {
  children: ReactNode
  className?: string
  padding?: string
}

export function Card({ children, className = '', padding = 'p-5' }: CardProps) {
  return (
    <div className={`card-elevated ${padding} ${className}`}>
      {children}
    </div>
  )
}
