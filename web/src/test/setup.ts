import '@testing-library/jest-dom/vitest'
import { afterEach, beforeEach, vi } from 'vitest'
import { cleanup } from '@testing-library/react'
import * as React from 'react'

afterEach(() => {
  cleanup()
})

beforeEach(() => {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    configurable: true,
    value: vi.fn().mockImplementation((query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  })

  ;(window as unknown as Record<string, unknown>).Telegram = {
    WebApp: {
      colorScheme: 'light',
      contentSafeAreaInset: { top: 0 },
      safeAreaInset: { top: 0 },
      onEvent: vi.fn(),
      offEvent: vi.fn(),
      ready: vi.fn(),
      expand: vi.fn(),
      close: vi.fn(),
      MainButton: { setText: vi.fn(), show: vi.fn(), hide: vi.fn(), onClick: vi.fn(), offClick: vi.fn() },
      BackButton: { show: vi.fn(), hide: vi.fn(), onClick: vi.fn(), offClick: vi.fn() },
      HapticFeedback: {
        impactOccurred: vi.fn(),
        notificationOccurred: vi.fn(),
        selectionChanged: vi.fn(),
      },
      initDataUnsafe: { user: { id: 1, language_code: 'en' } },
      initData: '',
    },
  }

  window.localStorage.clear()
  window.sessionStorage.clear()
})

vi.mock('@tma.js/sdk-react', () => ({
  miniApp: { mount: vi.fn(), unmount: vi.fn() },
  themeParams: { state: { value: {} } },
  viewport: {
    isStable: () => true,
    isExpanded: () => true,
    expand: vi.fn(),
  },
  backButton: {
    isSupported: () => false,
    show: vi.fn(),
    hide: vi.fn(),
    onClick: vi.fn(() => () => {}),
  },
  useSignal: () => undefined,
  retrieveRawInitData: () => 'mock_init_data',
}))

const motionProps = new Set([
  'initial', 'animate', 'exit', 'transition', 'variants',
  'whileHover', 'whileTap', 'whileDrag', 'whileFocus', 'whileInView',
  'drag', 'dragConstraints', 'dragElastic', 'dragMomentum', 'dragSnapToOrigin',
  'dragListener', 'dragTransition', 'onDragStart', 'onDragEnd', 'onDrag',
  'onAnimationStart', 'onAnimationComplete', 'onUpdate',
  'layout', 'layoutId', 'layoutScroll', 'layoutDependency',
  'viewport', 'custom',
])

function stripMotionProps(props: Record<string, unknown>): Record<string, unknown> {
  const out: Record<string, unknown> = {}
  for (const [k, v] of Object.entries(props)) {
    if (!motionProps.has(k)) out[k] = v
  }
  return out
}

vi.mock('framer-motion', () => {
  const motion = new Proxy(
    {},
    {
      get(_target, key) {
        if (typeof key !== 'string') return undefined
        const tag = key
        const Component = React.forwardRef<HTMLElement, Record<string, unknown>>((props, ref) => {
          const { children, ...rest } = props
          return React.createElement(tag, { ...stripMotionProps(rest), ref }, children as React.ReactNode)
        })
        Component.displayName = `motion.${tag}`
        return Component
      },
    },
  )

  return {
    motion,
    AnimatePresence: ({ children }: { children: React.ReactNode }) => children,
    LayoutGroup: ({ children }: { children: React.ReactNode }) => children,
    useAnimation: () => ({ start: () => Promise.resolve(), stop: () => {}, set: () => {} }),
    useMotionValue: (v: number) => ({ get: () => v, set: () => {}, on: () => () => {} }),
    useTransform: () => ({ get: () => 0, set: () => {}, on: () => () => {} }),
    useSpring: (v: number) => ({ get: () => v, set: () => {}, on: () => () => {} }),
    useInView: () => true,
    useScroll: () => ({ scrollY: { get: () => 0, on: () => () => {} } }),
    animate: () => ({ stop: () => {} }),
  }
})
