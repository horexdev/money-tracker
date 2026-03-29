import { motion } from 'framer-motion'
import { Plus } from '@phosphor-icons/react'

interface FABProps {
  onClick: () => void
  label?: string
}

export function FAB({ onClick, label }: FABProps) {
  return (
    <motion.button
      onClick={onClick}
      // Skip the mount animation entirely — the position jump caused by
      // --tab-bar-h resolving after first paint looks worse than no animation.
      initial={false}
      whileTap={{ scale: 0.92 }}
      className="fixed z-50 flex items-center gap-2 bg-accent text-accent-text shadow-[0_4px_20px_rgba(99,102,241,0.45)] active:shadow-[0_2px_8px_rgba(99,102,241,0.3)] transition-shadow rounded-full select-none"
      style={{
        right: '1rem',
        bottom: 'calc(var(--tab-bar-h) + 1rem)',
        paddingLeft: label ? '1.125rem' : '1rem',
        paddingRight: label ? '1.125rem' : '1rem',
        paddingTop: '0.875rem',
        paddingBottom: '0.875rem',
      }}
    >
      <Plus size={20} weight="bold" />
      {label && <span className="text-[13px] font-bold">{label}</span>}
    </motion.button>
  )
}
