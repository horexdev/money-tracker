import { useHaptic } from '../hooks/useHaptic'
import { ICON_CHOICES } from '../lib/categoryIcons'

interface IconPickerProps {
  selected: string
  onSelect: (id: string) => void
}

export function IconPicker({ selected, onSelect }: IconPickerProps) {
  const { selection } = useHaptic()

  return (
    <div className="grid grid-cols-8 gap-1.5">
      {ICON_CHOICES.map((choice) => {
        const isActive = selected === choice.id
        return (
          <button
            key={choice.id}
            type="button"
            onClick={() => { onSelect(choice.id); selection() }}
            className={`
              flex items-center justify-center h-10 rounded-2xl transition-all duration-150 active:scale-90
              ${isActive
                ? 'bg-accent text-white shadow-(--shadow-accent-pill)'
                : 'bg-accent-subtle text-accent'
              }
            `}
          >
            <choice.Icon size={18} weight="fill" />
          </button>
        )
      })}
    </div>
  )
}
