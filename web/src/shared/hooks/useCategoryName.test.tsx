import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { renderHook } from '@testing-library/react'
import { I18nextProvider, initReactI18next } from 'react-i18next'
import i18n from 'i18next'

import { useCategoryName } from './useCategoryName'

function buildI18n(lang: 'en' | 'ru' = 'en') {
  const instance = i18n.createInstance()
  instance.use(initReactI18next).init({
    lng: lang,
    fallbackLng: 'en',
    resources: {
      en: {
        translation: {
          categories: { names: { Food: 'Food', Transport: 'Transport' } },
        },
      },
      ru: {
        translation: {
          categories: { names: { Food: 'Еда', Transport: 'Транспорт' } },
        },
      },
    },
    interpolation: { escapeValue: false },
    react: { useSuspense: false },
  })
  return instance
}

function makeWrapper(instance: ReturnType<typeof buildI18n>) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return <I18nextProvider i18n={instance}>{children}</I18nextProvider>
  }
}

describe('useCategoryName', () => {
  it('translates known category names', () => {
    const wrapper = makeWrapper(buildI18n('en'))
    const { result } = renderHook(() => useCategoryName(), { wrapper })
    expect(result.current('Food')).toBe('Food')
    expect(result.current('Transport')).toBe('Transport')
  })

  it('falls back to the input value when translation is missing', () => {
    const wrapper = makeWrapper(buildI18n('en'))
    const { result } = renderHook(() => useCategoryName(), { wrapper })
    expect(result.current('Unknown')).toBe('Unknown')
    expect(result.current('CustomCategory')).toBe('CustomCategory')
  })

  it('uses the active i18n language for translations', () => {
    const wrapper = makeWrapper(buildI18n('ru'))
    const { result } = renderHook(() => useCategoryName(), { wrapper })
    expect(result.current('Food')).toBe('Еда')
    expect(result.current('Transport')).toBe('Транспорт')
  })

  it('keeps the fallback identity even after switching languages', () => {
    const wrapper = makeWrapper(buildI18n('ru'))
    const { result } = renderHook(() => useCategoryName(), { wrapper })
    expect(result.current('Mystery')).toBe('Mystery')
  })
})
