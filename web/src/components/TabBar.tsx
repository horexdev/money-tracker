import { NavLink } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Home, PlusCircle, Clock, BarChart3, MoreHorizontal } from 'lucide-react'
import { useHaptic } from '../hooks/useHaptic'
import type { ReactNode } from 'react'

interface Tab {
  to: string
  icon: ReactNode
  labelKey: string
  isCenter?: boolean
}

const TABS: Tab[] = [
  { to: '/',        icon: <Home size={22} />,           labelKey: 'tabs.home'    },
  { to: '/history', icon: <Clock size={22} />,          labelKey: 'tabs.history' },
  { to: '/add',     icon: <PlusCircle size={28} />,     labelKey: 'tabs.add',    isCenter: true },
  { to: '/stats',   icon: <BarChart3 size={22} />,      labelKey: 'tabs.stats'   },
  { to: '/more',    icon: <MoreHorizontal size={22} />, labelKey: 'tabs.more'    },
]

export function TabBar() {
  const { t } = useTranslation()
  const { selection } = useHaptic()

  return (
    <nav
      className="fixed bottom-0 left-0 right-0 z-50 flex glass border-t border-border"
      style={{ paddingBottom: 'var(--safe-bottom)' }}
    >
      {TABS.map((tab) => (
        <NavLink
          key={tab.to}
          to={tab.to}
          end={tab.to === '/'}
          onClick={selection}
          className={({ isActive }) =>
            `flex flex-1 flex-col items-center justify-center gap-0.5 py-3 transition-all duration-200 ${
              tab.isCenter ? '' : isActive ? 'text-accent' : 'text-muted'
            }`
          }
        >
          {({ isActive }) =>
            tab.isCenter ? (
              <div className={`
                aurora-bg w-12 h-12 rounded-2xl flex items-center justify-center text-white
                shadow-lg transition-transform duration-200 active:scale-90
                ${isActive ? 'scale-95' : 'scale-100'}
              `}>
                {tab.icon}
              </div>
            ) : (
              <>
                <div className={`transition-transform duration-200 ${isActive ? 'scale-110' : 'scale-100'}`}>
                  {tab.icon}
                </div>
                <span className="text-[10px] leading-none font-medium">{t(tab.labelKey)}</span>
              </>
            )
          }
        </NavLink>
      ))}
    </nav>
  )
}
