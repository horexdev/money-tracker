import type { ButtonHTMLAttributes, ReactNode } from 'react'

type Variant = 'primary' | 'secondary' | 'ghost' | 'destructive'
type Size = 'sm' | 'md' | 'lg'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
  size?: Size
  children: ReactNode
}

const variantClasses: Record<Variant, string> = {
  primary:     'bg-accent text-accent-text shadow-[0_4px_16px_rgba(67,56,202,0.25)]',
  secondary:   'bg-surface text-text shadow-[0_2px_8px_rgba(0,0,0,0.06)]',
  ghost:       'bg-transparent text-accent',
  destructive: 'bg-expense/10 text-expense',
}

const sizeClasses: Record<Size, string> = {
  sm: 'h-9 px-4 text-xs rounded-[--radius-sm]',
  md: 'h-11 px-5 text-sm rounded-[--radius-btn]',
  lg: 'h-13 px-6 text-base rounded-[--radius-btn]',
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
        inline-flex items-center justify-center font-bold
        transition-all duration-200 select-none
        active:scale-[0.96] disabled:opacity-40 disabled:pointer-events-none
        ${variantClasses[variant]} ${sizeClasses[size]} ${className}
      `}
      disabled={disabled}
      {...rest}
    >
      {children}
    </button>
  )
}
