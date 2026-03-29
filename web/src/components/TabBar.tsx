import { useEffect, useState } from 'react'
import { NavLink } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { House, Plus, ClockCounterClockwise, ChartBar, DotsThree } from '@phosphor-icons/react'
import { useHaptic } from '../hooks/useHaptic'
import type { ReactNode } from 'react'

interface Tab {
  to: string
  icon: ReactNode
  labelKey: string
  isCenter?: boolean
}

const TABS: Tab[] = [
  { to: '/',        icon: <House size={24} weight="fill" />,                  labelKey: 'tabs.home'    },
  { to: '/history', icon: <ClockCounterClockwise size={24} weight="bold" />, labelKey: 'tabs.history' },
  { to: '/add',     icon: <Plus size={24} weight="bold" />,                  labelKey: 'tabs.add',    isCenter: true },
  { to: '/stats',   icon: <ChartBar size={24} weight="fill" />,             labelKey: 'tabs.stats'   },
  { to: '/more',    icon: <DotsThree size={24} weight="bold" />,            labelKey: 'tabs.more'    },
]

/** Hide the tab bar while any text input or textarea is focused so it
 *  doesn't float over the virtual keyboard on iOS. */
function useKeyboardVisible() {
  const [visible, setVisible] = useState(false)

  useEffect(() => {
    const INPUT_SELECTORS = 'input, textarea, [contenteditable]'

    function onFocusIn(e: FocusEvent) {
      if ((e.target as Element)?.matches?.(INPUT_SELECTORS)) {
        setVisible(true)
      }
    }
    function onFocusOut() {
      // Small delay so the bar doesn't flash back before keyboard fully dismisses
      setTimeout(() => {
        if (!document.activeElement?.matches?.(INPUT_SELECTORS)) {
          setVisible(false)
        }
      }, 100)
    }

    document.addEventListener('focusin', onFocusIn)
    document.addEventListener('focusout', onFocusOut)
    return () => {
      document.removeEventListener('focusin', onFocusIn)
      document.removeEventListener('focusout', onFocusOut)
    }
  }, [])

  return visible
}

export function TabBar() {
  const { t } = useTranslation()
  const { selection } = useHaptic()
  const keyboardOpen = useKeyboardVisible()

  if (keyboardOpen) return null

  return (
    <nav
      className="fixed bottom-0 left-0 right-0 z-50 glass"
      style={{ paddingBottom: 'var(--safe-bottom)' }}
    >
      <div className="flex items-end">
        {TABS.map((tab) => (
          <NavLink
            key={tab.to}
            to={tab.to}
            end={tab.to === '/'}
            onClick={selection}
            className={({ isActive }) =>
              `flex flex-1 flex-col items-center justify-center gap-1 transition-all duration-200 ${
                tab.isCenter ? 'pb-2 pt-1' : 'py-2.5'
              } ${isActive ? 'text-accent' : 'text-muted'}`
            }
          >
            {({ isActive }) =>
              tab.isCenter ? (
                <div className={`
                  w-[52px] h-[52px] -mt-4 rounded-2xl flex items-center justify-center
                  hero-gradient text-white
                  shadow-[0_4px_20px_rgba(67,56,202,0.45)]
                  transition-transform duration-150
                  ${isActive ? 'scale-95' : 'active:scale-90'}
                `}>
                  {tab.icon}
                </div>
              ) : (
                <>
                  <div className={`transition-transform duration-150 ${isActive ? 'scale-110' : ''}`}>
                    {tab.icon}
                  </div>
                  <span className="text-[10px] leading-none font-semibold">{t(tab.labelKey)}</span>
                </>
              )
            }
          </NavLink>
        ))}
      </div>
    </nav>
  )
}
