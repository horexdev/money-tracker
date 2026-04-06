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
  const padding  = size === 'sm' ? 'py-2 px-3' : 'py-2.5 px-4'

  return (
    <div className="flex bg-bg rounded-(--radius-btn) p-1.5 gap-1">
      {options.map((opt) => (
        <button
          key={opt.value}
          onClick={() => onChange(opt.value)}
          className={`
            flex-1 ${padding} rounded-(--radius-sm) ${textSize} font-bold
            transition-all duration-200 select-none
            ${opt.value === value
              ? 'bg-surface text-text shadow-card'
              : 'text-muted active:bg-surface/50'
            }
          `}
        >
          {opt.label}
        </button>
      ))}
    </div>
  )
}
