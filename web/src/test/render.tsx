import type { ReactElement, ReactNode } from 'react'
import { render, type RenderOptions } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { I18nextProvider, initReactI18next } from 'react-i18next'
import i18n from 'i18next'
import en from '../i18n/locales/en.json'

function createTestI18n() {
  const instance = i18n.createInstance()
  instance.use(initReactI18next).init({
    lng: 'en',
    fallbackLng: 'en',
    resources: { en: { translation: en } },
    interpolation: { escapeValue: false },
    react: { useSuspense: false },
  })
  return instance
}

function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0, staleTime: 0 },
      mutations: { retry: false },
    },
  })
}

interface ProviderOptions {
  route?: string
  queryClient?: QueryClient
}

export function renderWithProviders(
  ui: ReactElement,
  options: ProviderOptions & Omit<RenderOptions, 'wrapper'> = {},
) {
  const { route = '/', queryClient: providedClient, ...renderOptions } = options
  const client = providedClient ?? createTestQueryClient()
  const i18nInstance = createTestI18n()

  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <I18nextProvider i18n={i18nInstance}>
        <QueryClientProvider client={client}>
          <MemoryRouter initialEntries={[route]}>{children}</MemoryRouter>
        </QueryClientProvider>
      </I18nextProvider>
    )
  }

  return {
    ...render(ui, { wrapper: Wrapper, ...renderOptions }),
    queryClient: client,
    i18n: i18nInstance,
  }
}
