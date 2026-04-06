type SpinnerSize = 'sm' | 'md' | 'lg'

const sizeClasses: Record<SpinnerSize, string> = {
  sm: 'w-4 h-4',
  md: 'w-6 h-6',
  lg: 'w-8 h-8',
}

export function Spinner({ size = 'md' }: { size?: SpinnerSize }) {
  return (
    <div
      className={`animate-spin rounded-full border-2 border-current border-t-transparent text-accent ${sizeClasses[size]}`}
      role="status"
      aria-label="Loading"
    />
  )
}
