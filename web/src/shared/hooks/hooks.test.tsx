import type { ReactNode } from 'react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

import {
  useAnimateNumbers,
  setAnimateNumbersPreference,
  ANIMATE_NUMBERS_KEY,
} from './useAnimateNumbers'
import { useHaptic } from './useHaptic'
import { useTgMainButton } from './useMainButton'
import { useFirstLaunchSetup, FIRST_LAUNCH_KEY } from './useFirstLaunchSetup'
import i18n from 'i18next'
import { I18nextProvider, initReactI18next } from 'react-i18next'

vi.mock('framer-motion', async () => {
  const actual = await vi.importActual<Record<string, unknown>>('framer-motion')
  return {
    ...actual,
    useReducedMotion: vi.fn(() => false),
  }
})

import { useReducedMotion } from 'framer-motion'

afterEach(() => {
  localStorage.clear()
})

describe('useAnimateNumbers', () => {
  it('defaults to enabled when no system reduced-motion preference', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false)
    const { result } = renderHook(() => useAnimateNumbers())
    expect(result.current[0]).toBe(true)
  })

  it('honours system reduced-motion when no explicit choice stored', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true)
    const { result } = renderHook(() => useAnimateNumbers())
    expect(result.current[0]).toBe(false)
  })

  it('explicit user choice wins over system preference', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true)
    setAnimateNumbersPreference(true)
    const { result } = renderHook(() => useAnimateNumbers())
    expect(result.current[0]).toBe(true)
  })

  it('setter persists preference and updates state', () => {
    const { result } = renderHook(() => useAnimateNumbers())
    act(() => {
      result.current[1](false)
    })
    expect(result.current[0]).toBe(false)
    expect(localStorage.getItem(ANIMATE_NUMBERS_KEY)).toBe('false')
  })
})

describe('setAnimateNumbersPreference', () => {
  it('stores boolean as string', () => {
    setAnimateNumbersPreference(true)
    expect(localStorage.getItem(ANIMATE_NUMBERS_KEY)).toBe('true')
    setAnimateNumbersPreference(false)
    expect(localStorage.getItem(ANIMATE_NUMBERS_KEY)).toBe('false')
  })
})

describe('useHaptic', () => {
  it('returns three callable methods', () => {
    const { result } = renderHook(() => useHaptic())
    expect(typeof result.current.impact).toBe('function')
    expect(typeof result.current.notification).toBe('function')
    expect(typeof result.current.selection).toBe('function')
  })

  it('callable without throwing', () => {
    const { result } = renderHook(() => useHaptic())
    expect(() => result.current.impact()).not.toThrow()
    expect(() => result.current.impact('heavy')).not.toThrow()
    expect(() => result.current.notification('success')).not.toThrow()
    expect(() => result.current.selection()).not.toThrow()
  })
})

describe('useTgMainButton', () => {
  it('does not throw when SDK button is not mounted', () => {
    const onClick = vi.fn()
    expect(() => {
      renderHook(() => useTgMainButton({ text: 'Save', onClick }))
    }).not.toThrow()
  })

  it('re-runs effect when props change', () => {
    const onClick = vi.fn()
    const { rerender } = renderHook(({ text }) => useTgMainButton({ text, onClick }), {
      initialProps: { text: 'Save' },
    })
    rerender({ text: 'Update' })
    rerender({ text: 'Confirm' })
    // No assertion needed beyond "didn't throw" — SDK mock is unmounted.
  })
})

describe('useFirstLaunchSetup', () => {
  function buildI18n(lang = 'en') {
    const inst = i18n.createInstance()
    inst.use(initReactI18next).init({
      lng: lang,
      fallbackLng: 'en',
      resources: { en: { translation: {} }, ru: { translation: {} } },
      interpolation: { escapeValue: false },
      react: { useSuspense: false },
    })
    return inst
  }

  function Wrapper({ children }: { children: ReactNode }) {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    return (
      <I18nextProvider i18n={buildI18n('en')}>
        <QueryClientProvider client={client}>{children}</QueryClientProvider>
      </I18nextProvider>
    )
  }

  beforeEach(() => {
    localStorage.removeItem(FIRST_LAUNCH_KEY)
  })

  it('does nothing when settings is undefined', () => {
    renderHook(() => useFirstLaunchSetup(undefined), { wrapper: Wrapper })
    expect(localStorage.getItem(FIRST_LAUNCH_KEY)).toBeNull()
  })

  it('marks first launch when settings provided', () => {
    renderHook(
      () => useFirstLaunchSetup({ language: 'en', base_currency: 'USD' } as unknown as never),
      { wrapper: Wrapper },
    )
    expect(localStorage.getItem(FIRST_LAUNCH_KEY)).toBe('1')
  })

  it('does not overwrite existing first-launch flag', () => {
    localStorage.setItem(FIRST_LAUNCH_KEY, 'previous-value')
    renderHook(
      () => useFirstLaunchSetup({ language: 'en', base_currency: 'USD' } as unknown as never),
      { wrapper: Wrapper },
    )
    expect(localStorage.getItem(FIRST_LAUNCH_KEY)).toBe('previous-value')
  })
})
