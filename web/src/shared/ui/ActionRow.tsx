import { Trash, PencilSimple } from '@phosphor-icons/react'

interface ActionRowProps {
  onDelete: () => void
  onEdit?: () => void
  isDeleting?: boolean
  children: React.ReactNode
}

export function ActionRow({ onDelete, onEdit, isDeleting, children }: ActionRowProps) {
  return (
    <div className={`flex items-center ${isDeleting ? 'opacity-30 pointer-events-none' : ''}`}>
      <div className="flex-1 min-w-0">{children}</div>
      <div className="flex items-center gap-1 px-2 shrink-0">
        {onEdit && (
          <button
            onClick={onEdit}
            className="w-8 h-8 rounded-xl flex items-center justify-center bg-accent/10 text-accent active:bg-accent/20 transition-colors"
          >
            <PencilSimple size={15} weight="bold" />
          </button>
        )}
        <button
          onClick={onDelete}
          className="w-8 h-8 rounded-xl flex items-center justify-center bg-destructive/10 text-destructive active:bg-destructive/20 transition-colors"
        >
          <Trash size={15} weight="bold" />
        </button>
      </div>
    </div>
  )
}
