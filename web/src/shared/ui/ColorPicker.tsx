import { useHaptic } from '../hooks/useHaptic'
import { COLOR_SWATCHES } from '../lib/constants'

interface ColorPickerProps {
  selected: string
  onSelect: (color: string) => void
  colors?: string[]
}

export function ColorPicker({ selected, onSelect, colors = COLOR_SWATCHES }: ColorPickerProps) {
  const { selection } = useHaptic()
  return (
    <div className="grid grid-cols-6 gap-2">
      {colors.map((c) => (
        <button
          key={c}
          type="button"
          onClick={() => { onSelect(c); selection() }}
          className="h-10 rounded-2xl transition-all duration-150 active:scale-90 relative flex items-center justify-center"
          style={{ background: c }}
        >
          {selected === c && (
            <span className="w-3 h-3 rounded-full border-2 border-white bg-white/40 block" />
          )}
        </button>
      ))}
    </div>
  )
}
