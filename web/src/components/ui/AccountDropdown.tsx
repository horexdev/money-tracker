import { useState, useRef, useEffect } from 'react'
import { CaretDown, Bank, PiggyBank, Money, CreditCard, Coins, type Icon } from '@phosphor-icons/react'
import { motion, AnimatePresence } from 'framer-motion'
import { formatCents } from '../../lib/money'
import type { Account, AccountType } from '../../types'

const ICONS: Record<AccountType, Icon> = {
  checking: Bank,
  savings: PiggyBank,
  cash: Money,
  credit: CreditCard,
  crypto: Coins,
}

interface Props {
  accounts: Account[]
  selectedId: number | null
  onChange: (id: number | null) => void
  allLabel?: string
  showBalance?: boolean
  /** 'hero' = white glass button (for gradient cards), 'surface' = uses theme tokens */
  variant?: 'hero' | 'surface'
  /** open the menu upward instead of downward */
  dropUp?: boolean
}

export function AccountDropdown({ accounts, selectedId, onChange, allLabel, showBalance = false, variant = 'hero', dropUp = false }: Props) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  const selected = accounts.find(a => a.id === selectedId) ?? null

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    if (open) document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [open])

  const TypeIcon = selected ? ICONS[selected.type] : Bank

  const triggerCls = variant === 'hero'
    ? 'flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-white/10 backdrop-blur-sm border border-white/15 active:scale-95 transition-transform'
    : 'flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-accent-subtle border border-border active:scale-95 transition-transform'

  const labelCls = variant === 'hero' ? 'text-white/90' : 'text-text'
  const allLabelCls = variant === 'hero' ? 'text-white/70' : 'text-muted'
  const caretCls = variant === 'hero' ? 'text-white/50' : 'text-muted'

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen(v => !v)}
        className={triggerCls}
      >
        {selected ? (
          <>
            <div
              className="w-4 h-4 rounded-full flex items-center justify-center shrink-0"
              style={{ background: selected.color }}
            >
              <TypeIcon size={9} weight="fill" className="text-white" />
            </div>
            <span className={`${labelCls} text-[12px] font-semibold max-w-[7rem] truncate`}>
              {selected.name}
            </span>
          </>
        ) : (
          <span className={`${allLabelCls} text-[12px] font-semibold`}>{allLabel ?? 'All'}</span>
        )}
        <CaretDown
          size={10}
          weight="bold"
          className={`${caretCls} transition-transform shrink-0 ${open ? 'rotate-180' : ''}`}
        />
      </button>

      <AnimatePresence>
        {open && (
          <motion.div
            initial={{ opacity: 0, scale: 0.95, y: dropUp ? 4 : -4 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: dropUp ? 4 : -4 }}
            transition={{ duration: 0.12 }}
            className={`absolute right-0 z-50 min-w-[160px] bg-card rounded-2xl shadow-lg border border-border overflow-hidden ${
              dropUp ? 'bottom-full mb-1.5' : 'top-full mt-1.5'
            }`}
            style={{ boxShadow: '0 8px 32px rgba(0,0,0,0.18)' }}
          >
            {allLabel && (
              <button
                type="button"
                onClick={() => { onChange(null); setOpen(false) }}
                className={`w-full flex items-center gap-2.5 px-3.5 py-2.5 text-left transition-colors ${
                  selectedId === null ? 'bg-accent/10 text-accent' : 'text-muted hover:bg-accent/5'
                }`}
              >
                <div className="w-6 h-6 rounded-xl bg-border flex items-center justify-center shrink-0">
                  <Bank size={12} weight="fill" className="text-muted" />
                </div>
                <span className="text-[13px] font-semibold">{allLabel}</span>
              </button>
            )}
            {accounts.map(acc => {
              const Icon = ICONS[acc.type]
              const isActive = selectedId === acc.id
              return (
                <button
                  key={acc.id}
                  type="button"
                  onClick={() => { onChange(acc.id); setOpen(false) }}
                  className={`w-full flex items-center gap-2.5 px-3.5 py-2.5 text-left transition-colors ${
                    isActive ? 'bg-accent/10' : 'hover:bg-accent/5'
                  }`}
                >
                  <div
                    className="w-6 h-6 rounded-xl flex items-center justify-center shrink-0"
                    style={{ background: acc.color }}
                  >
                    <Icon size={12} weight="fill" className="text-white" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className={`text-[13px] font-semibold truncate ${isActive ? 'text-accent' : 'text-text'}`}>
                      {acc.name}
                    </p>
                    {showBalance && (
                      <p className="text-[11px] text-muted tabular-nums">
                        {formatCents(acc.balance_cents, acc.currency_code)}
                      </p>
                    )}
                  </div>
                </button>
              )
            })}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
