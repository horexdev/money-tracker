import { useCallback, useRef } from 'react'

interface UseLongPressOptions {
  /** Long-press threshold in milliseconds. Default: 500. */
  delay?: number
  /** Movement threshold in pixels that cancels the long-press. Default: 6. */
  moveThreshold?: number
  /** Called on a normal click (no long-press). */
  onClick?: () => void
}

interface PointerHandlers {
  onPointerDown: (e: React.PointerEvent) => void
  onPointerMove: (e: React.PointerEvent) => void
  onPointerUp: (e: React.PointerEvent) => void
  onPointerLeave: (e: React.PointerEvent) => void
  onPointerCancel: (e: React.PointerEvent) => void
}

/**
 * useLongPress detects a long-press gesture on touch + mouse pointers.
 * Returns handlers to spread on the target element. Movement beyond
 * `moveThreshold` cancels the gesture so swipes/scrolls don't fire.
 */
export function useLongPress(onLongPress: () => void, opts: UseLongPressOptions = {}): PointerHandlers {
  const delay = opts.delay ?? 500
  const moveThreshold = opts.moveThreshold ?? 6
  const onClick = opts.onClick

  const timerRef = useRef<number | null>(null)
  const startRef = useRef<{ x: number; y: number } | null>(null)
  const triggeredRef = useRef(false)

  const clear = useCallback(() => {
    if (timerRef.current !== null) {
      window.clearTimeout(timerRef.current)
      timerRef.current = null
    }
    startRef.current = null
  }, [])

  const onPointerDown = useCallback((e: React.PointerEvent) => {
    triggeredRef.current = false
    startRef.current = { x: e.clientX, y: e.clientY }
    timerRef.current = window.setTimeout(() => {
      triggeredRef.current = true
      timerRef.current = null
      onLongPress()
    }, delay)
  }, [delay, onLongPress])

  const onPointerMove = useCallback((e: React.PointerEvent) => {
    if (!startRef.current || timerRef.current === null) return
    const dx = e.clientX - startRef.current.x
    const dy = e.clientY - startRef.current.y
    if (Math.hypot(dx, dy) > moveThreshold) {
      clear()
    }
  }, [clear, moveThreshold])

  const onPointerUp = useCallback(() => {
    const wasTimer = timerRef.current !== null
    clear()
    if (wasTimer && !triggeredRef.current && onClick) {
      onClick()
    }
  }, [clear, onClick])

  const onPointerLeave = useCallback(() => {
    clear()
  }, [clear])

  const onPointerCancel = useCallback(() => {
    clear()
  }, [clear])

  return { onPointerDown, onPointerMove, onPointerUp, onPointerLeave, onPointerCancel }
}
