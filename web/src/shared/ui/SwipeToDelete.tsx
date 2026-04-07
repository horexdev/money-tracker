import { ActionRow } from './ActionRow'

interface SwipeToDeleteProps {
  onDelete: () => void
  children: React.ReactNode
}

export function SwipeToDelete({ onDelete, children }: SwipeToDeleteProps) {
  return <ActionRow onDelete={onDelete}>{children}</ActionRow>
}
