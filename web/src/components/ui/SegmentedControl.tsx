interface Option {
  value: string
  label: string
}

interface SegmentedControlProps {
  options: Option[]
  value: string
  onChange: (value: string) => void
  size?: 'sm' | 'md'
}

export function SegmentedControl({ options, value, onChange, size = 'md' }: SegmentedControlProps) {
  const textSize = size === 'sm' ? 'text-xs' : 'text-sm'
  const padding  = size === 'sm' ? 'py-1.5 px-2' : 'py-2 px-3'

  return (
    <div className="flex bg-surface rounded-[--radius-btn] p-1 gap-0.5">
      {options.map((opt) => (
        <button
          key={opt.value}
          onClick={() => onChange(opt.value)}
          className={`
            flex-1 ${padding} rounded-[--radius-sm] ${textSize} font-semibold
            transition-all duration-150 select-none
            ${opt.value === value
              ? 'bg-accent text-accent-text shadow-sm'
              : 'text-muted active:bg-border'
            }
          `}
        >
          {opt.label}
        </button>
      ))}
    </div>
  )
}
