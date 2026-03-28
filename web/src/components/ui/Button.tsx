import type { ButtonHTMLAttributes, ReactNode } from 'react'

type Variant = 'primary' | 'secondary' | 'ghost' | 'destructive'
type Size = 'sm' | 'md'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
  size?: Size
  children: ReactNode
}

const variantClasses: Record<Variant, string> = {
  primary: 'bg-accent text-accent-text',
  secondary: 'bg-surface text-text',
  ghost: 'bg-transparent text-accent',
  destructive: 'bg-transparent text-destructive',
}

const sizeClasses: Record<Size, string> = {
  sm: 'px-3 py-1.5 text-xs',
  md: 'px-4 py-2 text-sm',
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
        inline-flex items-center justify-center font-medium
        rounded-[--radius-btn] transition-all duration-200
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
