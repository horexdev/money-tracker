import { NavLink } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Home, PlusCircle, Clock, BarChart3, MoreHorizontal } from 'lucide-react'
import { useHaptic } from '../hooks/useHaptic'
import type { ReactNode } from 'react'

interface Tab {
  to: string
  icon: ReactNode
  labelKey: string
}

const TABS: Tab[] = [
  { to: '/',        icon: <Home size={22} />,           labelKey: 'tabs.home'    },
  { to: '/history', icon: <Clock size={22} />,          labelKey: 'tabs.history' },
  { to: '/add',     icon: <PlusCircle size={26} />,     labelKey: 'tabs.add'     },
  { to: '/stats',   icon: <BarChart3 size={22} />,      labelKey: 'tabs.stats'   },
  { to: '/more',    icon: <MoreHorizontal size={22} />, labelKey: 'tabs.more'    },
]

export function TabBar() {
  const { t } = useTranslation()
  const { selection } = useHaptic()

  return (
    <nav
      className="fixed bottom-0 left-0 right-0 z-50 flex bg-surface border-t border-border-strong"
      style={{ paddingBottom: 'var(--safe-bottom)' }}
    >
      {TABS.map((tab) => (
        <NavLink
          key={tab.to}
          to={tab.to}
          end={tab.to === '/'}
          onClick={selection}
          className={({ isActive }) =>
            `flex flex-1 flex-col items-center justify-center gap-0.5 py-2 transition-colors duration-200 ${
              isActive ? 'text-accent' : 'text-muted'
            }`
          }
        >
          {({ isActive }) => (
            <>
              {tab.icon}
              <span className="text-[10px] leading-none">{t(tab.labelKey)}</span>
              <div
                className={`w-1 h-1 rounded-full transition-all duration-200 ${
                  isActive ? 'bg-accent scale-100' : 'bg-transparent scale-0'
                }`}
              />
            </>
          )}
        </NavLink>
      ))}
    </nav>
  )
}
