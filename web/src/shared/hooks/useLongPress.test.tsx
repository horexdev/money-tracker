import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useLongPress } from './useLongPress'

beforeEach(() => {
  vi.useFakeTimers()
})

afterEach(() => {
  vi.useRealTimers()
})

function makeEvent(x = 0, y = 0): React.PointerEvent {
  return { clientX: x, clientY: y } as unknown as React.PointerEvent
}

describe('useLongPress', () => {
  it('fires onLongPress after the delay', () => {
    const onLongPress = vi.fn()
    const { result } = renderHook(() => useLongPress(onLongPress, { delay: 500 }))

    act(() => result.current.onPointerDown(makeEvent(0, 0)))
    expect(onLongPress).not.toHaveBeenCalled()

    act(() => { vi.advanceTimersByTime(500) })
    expect(onLongPress).toHaveBeenCalledTimes(1)
  })

  it('does not fire onLongPress when pointer up before delay', () => {
    const onLongPress = vi.fn()
    const { result } = renderHook(() => useLongPress(onLongPress, { delay: 500 }))

    act(() => result.current.onPointerDown(makeEvent(0, 0)))
    act(() => { vi.advanceTimersByTime(200) })
    act(() => result.current.onPointerUp(makeEvent(0, 0)))
    act(() => { vi.advanceTimersByTime(500) })

    expect(onLongPress).not.toHaveBeenCalled()
  })

  it('cancels long-press when pointer moves beyond threshold', () => {
    const onLongPress = vi.fn()
    const { result } = renderHook(() => useLongPress(onLongPress, { delay: 500, moveThreshold: 6 }))

    act(() => result.current.onPointerDown(makeEvent(0, 0)))
    act(() => result.current.onPointerMove(makeEvent(10, 10))) // hypot > 6
    act(() => { vi.advanceTimersByTime(500) })

    expect(onLongPress).not.toHaveBeenCalled()
  })

  it('keeps timer when movement is within threshold', () => {
    const onLongPress = vi.fn()
    const { result } = renderHook(() => useLongPress(onLongPress, { delay: 500, moveThreshold: 6 }))

    act(() => result.current.onPointerDown(makeEvent(0, 0)))
    act(() => result.current.onPointerMove(makeEvent(2, 3))) // hypot ~3.6 < 6
    act(() => { vi.advanceTimersByTime(500) })

    expect(onLongPress).toHaveBeenCalledTimes(1)
  })

  it('fires onClick on a quick tap when provided', () => {
    const onLongPress = vi.fn()
    const onClick = vi.fn()
    const { result } = renderHook(() => useLongPress(onLongPress, { delay: 500, onClick }))

    act(() => result.current.onPointerDown(makeEvent(0, 0)))
    act(() => { vi.advanceTimersByTime(100) })
    act(() => result.current.onPointerUp(makeEvent(0, 0)))

    expect(onClick).toHaveBeenCalledTimes(1)
    expect(onLongPress).not.toHaveBeenCalled()
  })

  it('does not fire onClick after long-press has triggered', () => {
    const onLongPress = vi.fn()
    const onClick = vi.fn()
    const { result } = renderHook(() => useLongPress(onLongPress, { delay: 500, onClick }))

    act(() => result.current.onPointerDown(makeEvent(0, 0)))
    act(() => { vi.advanceTimersByTime(500) })
    act(() => result.current.onPointerUp(makeEvent(0, 0)))

    expect(onLongPress).toHaveBeenCalledTimes(1)
    expect(onClick).not.toHaveBeenCalled()
  })

  it('cancels timer on pointer leave', () => {
    const onLongPress = vi.fn()
    const { result } = renderHook(() => useLongPress(onLongPress, { delay: 500 }))

    act(() => result.current.onPointerDown(makeEvent(0, 0)))
    act(() => result.current.onPointerLeave(makeEvent(0, 0)))
    act(() => { vi.advanceTimersByTime(500) })

    expect(onLongPress).not.toHaveBeenCalled()
  })
})
