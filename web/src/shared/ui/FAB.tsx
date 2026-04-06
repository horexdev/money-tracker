import { createPortal } from 'react-dom'
import { motion } from 'framer-motion'
import { Plus } from '@phosphor-icons/react'

interface FABProps {
  onClick: () => void
  label?: string
  ariaLabel?: string
}

export function FAB({ onClick, label, ariaLabel }: FABProps) {
  return createPortal(
    <motion.button
      onClick={onClick}
      aria-label={ariaLabel ?? label ?? 'Add'}
      style={{
        right: '1rem',
        bottom: 'calc(72px + env(safe-area-inset-bottom, 0px) + 1rem)',
        paddingLeft: label ? '1.125rem' : '1rem',
        paddingRight: label ? '1.125rem' : '1rem',
        paddingTop: '0.875rem',
        paddingBottom: '0.875rem',
      }}
      initial={{ opacity: 0, scale: 0.8, y: 24 }}
      animate={{ opacity: 1, scale: 1, y: 0 }}
      transition={{ type: 'spring', stiffness: 400, damping: 30, delay: 0.05 }}
      whileTap={{ scale: 0.92 }}
      className="fixed z-50 flex items-center gap-2 bg-accent text-accent-text shadow-(--shadow-fab) active:shadow-(--shadow-fab-active) transition-shadow rounded-full select-none"
    >
      <Plus size={20} weight="bold" />
      {label && <span className="text-[13px] font-bold">{label}</span>}
    </motion.button>,
    document.body,
  )
}
