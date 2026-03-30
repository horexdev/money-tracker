import { createPortal } from 'react-dom'
import { motion } from 'framer-motion'

interface BottomSheetProps {
  onClose: () => void
  children: React.ReactNode
}

/**
 * Shared bottom sheet — consistent rounded corners, spring animation, drag-to-dismiss.
 * Use inside AnimatePresence for enter/exit transitions.
 */
export function BottomSheet({ onClose, children }: BottomSheetProps) {
  return createPortal(
    <>
      <motion.div
        className="fixed inset-0 bg-black/50 z-[60]"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        onClick={onClose}
      />
      <motion.div
        className="fixed bottom-0 left-0 right-0 z-[60] bg-surface rounded-t-card overflow-hidden"
        style={{ boxShadow: '0 -8px 40px rgba(0,0,0,0.18)' }}
        initial={{ y: '100%' }}
        animate={{ y: 0 }}
        exit={{ y: '100%' }}
        transition={{ type: 'spring', damping: 32, stiffness: 320 }}
        drag="y"
        dragConstraints={{ top: 0 }}
        dragElastic={{ top: 0, bottom: 0.25 }}
        onDragEnd={(_, info) => {
          if (info.velocity.y > 400 || info.offset.y > 100) onClose()
        }}
      >
        <div className="pt-3 pb-1 flex justify-center shrink-0">
          <div className="w-9 h-1 rounded-full bg-border" />
        </div>
        {children}
      </motion.div>
    </>,
    document.body,
  )
}
