import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import './i18n/config'
import './index.css'
import App from './app/App'

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

  // Auto-detect real backend: probe localhost:8080 via Vite proxy.
  // Any HTTP response (even 401) means the backend is up → skip mocks.
  // Network error means backend is down → fall back to mocks.
  let backendAvailable = false
  try {
    const probe = await fetch('/api/v1/settings', { signal: AbortSignal.timeout(1000) })
    backendAvailable = probe.ok || probe.status === 401 || probe.status === 403
  } catch {
    backendAvailable = false
  }

  if (!backendAvailable) {
    const { setupMockFetch } = await import('./mocks/setup')
    setupMockFetch()
  }
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
