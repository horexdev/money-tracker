import { useState, useRef, useEffect } from 'react'
import { createPortal } from 'react-dom'
import { CaretDown, Bank, PiggyBank, Money, CreditCard, Coins, type Icon } from '@phosphor-icons/react'
import { motion, AnimatePresence } from 'framer-motion'
import { MoneyText } from './MoneyText'
import type { Account, AccountType } from '../types'

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
  /** If provided, shows an "all accounts" option with this label */
  allLabel?: string
  showBalance?: boolean
  /** 'hero' = white glass button (for gradient cards), 'surface' = uses theme tokens */
  variant?: 'hero' | 'surface'
}

export function AccountDropdown({ accounts, selectedId, onChange, allLabel, showBalance = false, variant = 'hero' }: Props) {
  const [open, setOpen] = useState(false)
  const [menuStyle, setMenuStyle] = useState<React.CSSProperties>({})
  const triggerRef = useRef<HTMLButtonElement>(null)

  const selected = accounts.find(a => a.id === selectedId) ?? null
  const TypeIcon = selected ? ICONS[selected.type] : Bank

  // Position the portal menu relative to the trigger button
  useEffect(() => {
    if (!open || !triggerRef.current) return
    const rect = triggerRef.current.getBoundingClientRect()
    const spaceBelow = window.innerHeight - rect.bottom
    const menuH = Math.min(accounts.length * 56 + (allLabel ? 56 : 0), 280)

    const menuWidth = Math.max(160, rect.width)
    const rightAligned = rect.right - menuWidth
    const left = Math.max(8, rightAligned)
    const openUp = spaceBelow < menuH + 8

    setMenuStyle({
      position: 'fixed',
      ...(openUp
        ? { bottom: window.innerHeight - rect.top + 6 }
        : { top: rect.bottom + 6 }
      ),
      left,
      minWidth: menuWidth,
      zIndex: 9999,
    })
  }, [open, accounts.length, allLabel])

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      const target = e.target as Node
      if (triggerRef.current && !triggerRef.current.contains(target)) {
        // Check if click is inside the portal menu
        const menu = document.getElementById('account-dropdown-menu')
        if (!menu || !menu.contains(target)) setOpen(false)
      }
    }
    if (open) document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [open])

  const triggerCls = variant === 'hero'
    ? 'flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-white/10 backdrop-blur-sm border border-white/15 active:scale-95 transition-transform'
    : 'flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-accent-subtle border border-border active:scale-95 transition-transform'

  const labelCls = variant === 'hero' ? 'text-white/90' : 'text-text'
  const allLabelCls = variant === 'hero' ? 'text-white/70' : 'text-muted'
  const caretCls = variant === 'hero' ? 'text-white/50' : 'text-muted'

  const menu = (
    <AnimatePresence>
      {open && (
        <motion.div
          id="account-dropdown-menu"
          initial={{ opacity: 0, scale: 0.95, y: -4 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.95, y: -4 }}
          transition={{ duration: 0.12 }}
          style={{ ...menuStyle, boxShadow: 'var(--shadow-dropdown)', backgroundColor: 'var(--app-surface)', borderColor: 'var(--app-border, rgba(0,0,0,0.08))' }}
          className="rounded-2xl border overflow-hidden"
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
            const AccIcon = ICONS[acc.type]
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
                  <AccIcon size={12} weight="fill" className="text-white" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className={`text-[13px] font-semibold truncate ${isActive ? 'text-accent' : 'text-text'}`}>
                    {acc.name}
                  </p>
                  {showBalance && (
                    <p className="text-[11px] text-muted tabular-nums">
                      <MoneyText cents={acc.balance_cents} currency={acc.currency_code} />
                    </p>
                  )}
                </div>
              </button>
            )
          })}
        </motion.div>
      )}
    </AnimatePresence>
  )

  return (
    <>
      <button
        ref={triggerRef}
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
      {createPortal(menu, document.body)}
    </>
  )
}
