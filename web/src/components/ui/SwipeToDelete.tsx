import { useRef, useState } from 'react'
import { motion, useMotionValue, useTransform, animate } from 'framer-motion'
import { Trash } from '@phosphor-icons/react'

interface SwipeToDeleteProps {
  onDelete: () => void
  children: React.ReactNode
}

export function SwipeToDelete({ onDelete, children }: SwipeToDeleteProps) {
  const x = useMotionValue(0)
  const [swiped, setSwiped] = useState(false)
  const deleteOpacity = useTransform(x, [-80, -40], [1, 0])
  const deleteScale   = useTransform(x, [-80, -40], [1, 0.8])
  const startX = useRef<number | null>(null)
  const isDragging = useRef(false)

  function handleDragEnd() {
    const val = x.get()
    if (val < -60) {
      animate(x, -72, { type: 'spring', stiffness: 400, damping: 35 })
      setSwiped(true)
    } else {
      animate(x, 0, { type: 'spring', stiffness: 400, damping: 35 })
      setSwiped(false)
    }
  }

  function handleClose() {
    animate(x, 0, { type: 'spring', stiffness: 400, damping: 35 })
    setSwiped(false)
  }

  return (
    <div className="relative overflow-hidden">
      {/* Delete background */}
      <div className="absolute inset-y-0 right-0 w-20 flex items-center justify-center bg-expense/10">
        <motion.div style={{ opacity: deleteOpacity, scale: deleteScale }}>
          <Trash size={20} weight="bold" className="text-destructive" />
        </motion.div>
      </div>

      {/* Row content */}
      <motion.div
        style={{ x }}
        drag="x"
        dragConstraints={{ left: -72, right: 0 }}
        dragElastic={0.05}
        onDragEnd={handleDragEnd}
        onClick={() => { if (swiped) handleClose() }}
        className="relative bg-surface"
      >
        {children}
      </motion.div>

      {/* Tap delete zone when swiped open */}
      {swiped && (
        <button
          className="absolute inset-y-0 right-0 w-20"
          onClick={onDelete}
        />
      )}
    </div>
  )
}
