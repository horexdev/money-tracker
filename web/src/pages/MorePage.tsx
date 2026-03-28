import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Tag, Wallet, Repeat, Target, Download, Settings } from 'lucide-react'
import { PageTransition } from '../components/PageTransition'
import { Card } from '../components/ui'
import type { ReactNode } from 'react'

interface MenuItem {
  to: string
  icon: ReactNode
  labelKey: string
  color: string
}

const MENU_ITEMS: MenuItem[] = [
  { to: '/categories', icon: <Tag size={20} />,      labelKey: 'more.categories', color: 'bg-blue-500/10 text-blue-500'    },
  { to: '/budgets',    icon: <Wallet size={20} />,    labelKey: 'more.budgets',    color: 'bg-amber-500/10 text-amber-500'  },
  { to: '/recurring',  icon: <Repeat size={20} />,    labelKey: 'more.recurring',  color: 'bg-purple-500/10 text-purple-500'},
  { to: '/savings',    icon: <Target size={20} />,    labelKey: 'more.savings',    color: 'bg-green-500/10 text-green-500'  },
  { to: '/export',     icon: <Download size={20} />,  labelKey: 'more.export',     color: 'bg-cyan-500/10 text-cyan-500'    },
  { to: '/settings',   icon: <Settings size={20} />,  labelKey: 'more.settings',   color: 'bg-gray-500/10 text-gray-500'    },
]

export function MorePage() {
  const { t } = useTranslation()

  return (
    <PageTransition>
      <div className="p-4">
        <h1 className="text-xl font-bold mb-4">{t('more.title')}</h1>
        <Card padding="p-0">
          <div className="divide-y divide-border">
            {MENU_ITEMS.map(item => (
              <Link
                key={item.to}
                to={item.to}
                className="flex items-center gap-3 px-4 py-3.5 active:bg-accent/5 transition-colors"
              >
                <div className={`w-9 h-9 rounded-[--radius-sm] flex items-center justify-center ${item.color}`}>
                  {item.icon}
                </div>
                <span className="text-sm font-medium">{t(item.labelKey)}</span>
              </Link>
            ))}
          </div>
        </Card>
      </div>
    </PageTransition>
  )
}
