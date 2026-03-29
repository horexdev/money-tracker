import type { ButtonHTMLAttributes, ReactNode } from 'react'

type Variant = 'primary' | 'secondary' | 'ghost' | 'destructive'
type Size = 'sm' | 'md' | 'lg'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
  size?: Size
  children: ReactNode
}

const variantClasses: Record<Variant, string> = {
  primary:     'bg-accent text-accent-text shadow-sm',
  secondary:   'bg-surface text-text border border-border',
  ghost:       'bg-transparent text-accent',
  destructive: 'bg-expense/10 text-expense',
}

const sizeClasses: Record<Size, string> = {
  sm: 'h-8 px-3 text-xs rounded-[--radius-xs]',
  md: 'h-10 px-4 text-sm rounded-[--radius-btn]',
  lg: 'h-12 px-5 text-base rounded-[--radius-btn]',
}

export function Button({
  variant = 'primary',
  size = 'md',
  className = '',
  disabled,
  children,
  ...rest
}: ButtonProps) {
  return (
    <button
      className={`
        inline-flex items-center justify-center font-semibold
        transition-all duration-150 select-none
        active:scale-[0.97] disabled:opacity-40 disabled:pointer-events-none
        ${variantClasses[variant]} ${sizeClasses[size]} ${className}
      `}
      disabled={disabled}
      {...rest}
    >
      {children}
    </button>
  )
}
