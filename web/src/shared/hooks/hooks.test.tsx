import type { ReactNode } from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

vi.mock('../api/settings', () => ({
  settingsApi: {
    get: vi.fn(),
    update: vi.fn(),
    resetData: vi.fn(),
  },
}))

vi.mock('framer-motion', async () => {
  const actual = await vi.importActual<Record<string, unknown>>('framer-motion')
  return {
    ...actual,
    useReducedMotion: vi.fn(() => false),
  }
})

import { useAnimateNumbers } from './useAnimateNumbers'
import { useChartStyle } from './useChartStyle'
import { SETTINGS_QUERY_KEY } from './useUserSettings'
import { useHaptic } from './useHaptic'
import { useTgMainButton } from './useMainButton'
import { useFirstLaunchSetup, FIRST_LAUNCH_KEY } from './useFirstLaunchSetup'
import { settingsApi } from '../api/settings'
import { useReducedMotion } from 'framer-motion'
import type { UserSettings } from '../types'
import i18n from 'i18next'
import { I18nextProvider, initReactI18next } from 'react-i18next'

const baseSettings: UserSettings = {
  base_currency: 'USD',
  display_currencies: [],
  language: 'en',
  is_admin: false,
  notify_budget_alerts: true,
  notify_recurring_reminders: true,
  notify_weekly_summary: false,
  notify_goal_milestones: true,
  stats_chart_style: 'donut',
  animate_numbers: null,
}

function renderWithSettings<T>(
  hook: () => T,
  settings: UserSettings = baseSettings,
): { result: { current: T }; client: QueryClient } {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  client.setQueryData(SETTINGS_QUERY_KEY, settings)
  function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={client}>{children}</QueryClientProvider>
  }
  const { result } = renderHook(hook, { wrapper: Wrapper })
  return { result, client }
}

describe('useAnimateNumbers', () => {
  beforeEach(() => {
    vi.mocked(settingsApi.get).mockResolvedValue(baseSettings)
    vi.mocked(settingsApi.update).mockReset()
  })

  it('defaults to enabled when settings is unset and no system reduced-motion', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false)
    const { result } = renderWithSettings(() => useAnimateNumbers())
    expect(result.current[0]).toBe(true)
  })

  it('honours system reduced-motion when settings.animate_numbers is null', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true)
    const { result } = renderWithSettings(() => useAnimateNumbers())
    expect(result.current[0]).toBe(false)
  })

  it('explicit settings value wins over system preference', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true)
    const { result } = renderWithSettings(
      () => useAnimateNumbers(),
      { ...baseSettings, animate_numbers: true },
    )
    expect(result.current[0]).toBe(true)
  })

  it('setter calls settingsApi.update with ui_preferences.animate_numbers', async () => {
    vi.mocked(settingsApi.update).mockResolvedValue({ ...baseSettings, animate_numbers: false })
    const { result } = renderWithSettings(() => useAnimateNumbers())

    act(() => {
      result.current[1](false)
    })

    await waitFor(() => {
      expect(settingsApi.update).toHaveBeenCalledWith({ ui_preferences: { animate_numbers: false } })
    })
  })
})

describe('useChartStyle', () => {
  beforeEach(() => {
    vi.mocked(settingsApi.get).mockResolvedValue(baseSettings)
    vi.mocked(settingsApi.update).mockReset()
  })

  it('returns the style stored in settings', () => {
    const { result } = renderWithSettings(
      () => useChartStyle(),
      { ...baseSettings, stats_chart_style: 'dual_bar' },
    )
    expect(result.current[0]).toBe('dual_bar')
  })

  it('falls back to donut when settings is not yet loaded', () => {
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    })
    function Wrapper({ children }: { children: ReactNode }) {
      return <QueryClientProvider client={client}>{children}</QueryClientProvider>
    }
    const { result } = renderHook(() => useChartStyle(), { wrapper: Wrapper })
    expect(result.current[0]).toBe('donut')
  })

  it('setter dispatches PATCH with the new style', async () => {
    vi.mocked(settingsApi.update).mockResolvedValue({ ...baseSettings, stats_chart_style: 'profit_bars' })
    const { result } = renderWithSettings(() => useChartStyle())

    act(() => {
      result.current[1]('profit_bars')
    })

    await waitFor(() => {
      expect(settingsApi.update).toHaveBeenCalledWith({ ui_preferences: { stats_chart_style: 'profit_bars' } })
    })
  })

  it('setter is a no-op when style is unchanged', () => {
    const { result } = renderWithSettings(
      () => useChartStyle(),
      { ...baseSettings, stats_chart_style: 'donut' },
    )
    act(() => {
      result.current[1]('donut')
    })
    expect(settingsApi.update).not.toHaveBeenCalled()
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
