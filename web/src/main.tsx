import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import './i18n/config'
import './index.css'
import App from './App'

let isTelegram = false

try {
  const { init } = await import('@tma.js/sdk-react')
  init()
  isTelegram = true
} catch {
  // Not inside Telegram — skip SDK init
}

if (!isTelegram && import.meta.env.DEV) {
  // Simulate admin Telegram user for local development
  ;(window as unknown as Record<string, unknown>).Telegram = {
    WebApp: {
      initDataUnsafe: { user: { id: 6554524765, language_code: 'en' } },
      colorScheme: 'light',
    },
  }
  const { setupMockFetch } = await import('./mocks/setup')
  setupMockFetch()
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 30_000, retry: 1 },
  },
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  </StrictMode>,
)
