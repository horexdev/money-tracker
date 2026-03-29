import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Tag, Wallet, ArrowsClockwise, Target, DownloadSimple, GearSix } from '@phosphor-icons/react'
import { PageTransition } from '../components/PageTransition'
import type { ReactNode } from 'react'

interface MenuItem {
  to: string
  icon: ReactNode
  labelKey: string
  gradient: string
  descKey?: string
  comingSoon?: boolean
}

const FEATURED: MenuItem = {
  to: '/budgets',
  icon: <Wallet size={26} weight="fill" />,
  labelKey: 'more.budgets',
  gradient: '',
  descKey: 'budgets.set_budget',
}

const GRID_ITEMS: MenuItem[] = [
  { to: '/savings',   icon: <Target size={22} weight="fill" />,           labelKey: 'more.savings',   gradient: 'from-emerald-400 to-teal-500' },
  { to: '/recurring', icon: <ArrowsClockwise size={22} weight="bold" />,  labelKey: 'more.recurring', gradient: 'from-violet-400 to-purple-600' },
  { to: '/categories',icon: <Tag size={22} weight="fill" />,              labelKey: 'more.categories',gradient: 'from-blue-400 to-indigo-500' },
  { to: '/export',    icon: <DownloadSimple size={22} weight="bold" />,   labelKey: 'more.export',    gradient: 'from-cyan-400 to-sky-500', comingSoon: true },
  { to: '/settings',  icon: <GearSix size={22} weight="fill" />,          labelKey: 'more.settings',  gradient: 'from-slate-400 to-slate-600' },
]

export function MorePage() {
  const { t } = useTranslation()

  return (
    <PageTransition>
      <div className="px-4 pt-4 space-y-4">
        {/* Featured full-width tile */}
        <Link
          to={FEATURED.to}
          className="block hero-gradient p-6 relative active:scale-[0.98] transition-transform"
          style={{ boxShadow: 'var(--shadow-hero)' }}
        >
          <div className="absolute -top-8 -right-8 w-32 h-32 rounded-full bg-white/[0.08] blur-xl pointer-events-none" />
          <div className="absolute -bottom-6 -left-6 w-24 h-24 rounded-full bg-indigo-400/20 blur-2xl pointer-events-none" />
          <div className="relative z-10">
            <div className="w-12 h-12 rounded-2xl bg-white/15 backdrop-blur-sm flex items-center justify-center text-white mb-4 border border-white/10">
              {FEATURED.icon}
            </div>
            <p className="text-white font-bold text-lg">{t(FEATURED.labelKey)}</p>
            {FEATURED.descKey && (
              <p className="text-white/50 text-xs mt-1 font-medium">{t(FEATURED.descKey)}</p>
            )}
          </div>
        </Link>

        {/* 2-column grid — elevated cards with colored icon containers */}
        <div className="grid grid-cols-2 gap-3">
          {GRID_ITEMS.map(item => {
            const inner = (
              <>
                <div className="absolute -top-4 -right-4 w-16 h-16 rounded-full bg-accent/[0.04] pointer-events-none" />
                {item.comingSoon && (
                  <span className="absolute top-2.5 right-2.5 text-[9px] font-bold text-muted bg-border rounded-full px-1.5 py-0.5 leading-none">
                    {t('common.coming_soon')}
                  </span>
                )}
                <div className={`w-11 h-11 rounded-2xl bg-gradient-to-br ${item.gradient} flex items-center justify-center text-white mb-4 shadow-sm ${item.comingSoon ? 'opacity-40' : ''}`}>
                  {item.icon}
                </div>
                <p className={`font-bold text-sm leading-tight ${item.comingSoon ? 'text-muted' : 'text-text'}`}>{t(item.labelKey)}</p>
              </>
            )
            if (item.comingSoon) {
              return (
                <div key={item.to} className="card-elevated p-5 relative select-none">
                  {inner}
                </div>
              )
            }
            return (
              <Link
                key={item.to}
                to={item.to}
                className="card-elevated p-5 active:scale-[0.96] transition-transform relative"
              >
                {inner}
              </Link>
            )
          })}
        </div>
      </div>
    </PageTransition>
  )
}
