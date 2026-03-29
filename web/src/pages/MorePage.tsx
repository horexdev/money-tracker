import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Tag, Wallet, Repeat, Target, Download, Settings } from 'lucide-react'
import { PageTransition } from '../components/PageTransition'
import type { ReactNode } from 'react'

interface MenuItem {
  to: string
  icon: ReactNode
  labelKey: string
  gradient: string
  span?: 'full' | 'half'
}

const MENU_ITEMS: MenuItem[] = [
  {
    to: '/budgets',
    icon: <Wallet size={24} />,
    labelKey: 'more.budgets',
    gradient: 'from-amber-400 to-orange-500',
    span: 'full',
  },
  {
    to: '/savings',
    icon: <Target size={22} />,
    labelKey: 'more.savings',
    gradient: 'from-emerald-400 to-teal-500',
  },
  {
    to: '/recurring',
    icon: <Repeat size={22} />,
    labelKey: 'more.recurring',
    gradient: 'from-violet-400 to-purple-600',
  },
  {
    to: '/categories',
    icon: <Tag size={22} />,
    labelKey: 'more.categories',
    gradient: 'from-blue-400 to-cyan-500',
  },
  {
    to: '/export',
    icon: <Download size={22} />,
    labelKey: 'more.export',
    gradient: 'from-cyan-400 to-sky-500',
  },
  {
    to: '/settings',
    icon: <Settings size={22} />,
    labelKey: 'more.settings',
    gradient: 'from-slate-400 to-slate-600',
  },
]

export function MorePage() {
  const { t } = useTranslation()

  const [budgets, ...rest] = MENU_ITEMS

  return (
    <PageTransition>
      <div className="px-4 pt-4 space-y-3">
        <h1 className="text-xl font-bold">{t('more.title')}</h1>

        {/* Full-width featured tile */}
        <Link
          to={budgets.to}
          className={`block bg-gradient-to-br ${budgets.gradient} rounded-[--radius-card] p-5 relative overflow-hidden active:scale-[0.98] transition-transform`}
        >
          <div className="absolute -top-6 -right-6 w-24 h-24 rounded-full bg-white/15 blur-xl pointer-events-none" />
          <div className="w-10 h-10 rounded-2xl bg-white/20 flex items-center justify-center text-white mb-3">
            {budgets.icon}
          </div>
          <p className="text-white font-semibold text-base">{t(budgets.labelKey)}</p>
          <p className="text-white/60 text-xs mt-0.5">{t('budgets.set_budget')}</p>
        </Link>

        {/* 2-column grid for the rest */}
        <div className="grid grid-cols-2 gap-3">
          {rest.map(item => (
            <Link
              key={item.to}
              to={item.to}
              className={`bg-gradient-to-br ${item.gradient} rounded-[--radius-card] p-4 relative overflow-hidden active:scale-[0.97] transition-transform`}
            >
              <div className="absolute -top-4 -right-4 w-16 h-16 rounded-full bg-white/10 blur-lg pointer-events-none" />
              <div className="w-9 h-9 rounded-xl bg-white/20 flex items-center justify-center text-white mb-3">
                {item.icon}
              </div>
              <p className="text-white font-semibold text-sm leading-tight">{t(item.labelKey)}</p>
            </Link>
          ))}
        </div>
      </div>
    </PageTransition>
  )
}
