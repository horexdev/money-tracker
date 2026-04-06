import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { CaretLeft, CaretRight } from '@phosphor-icons/react'

/* ─── Helpers ─── */
export function fmtISO(year: number, month: number, day: number) {
  return `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`
}

export function fmtDisplay(iso: string) {
  const d = new Date(iso + 'T00:00:00')
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}

/* ─── Range Calendar ─── */
export function RangeCalendar({
  from,
  to,
  activeField,
  onSelect,
  singleDate = false,
}: {
  from: string
  to: string
  activeField: 'from' | 'to'
  onSelect: (iso: string) => void
  singleDate?: boolean
}) {
  const initDate = new Date((activeField === 'from' ? from : to) + 'T00:00:00')
  const [viewYear, setViewYear] = useState(initDate.getFullYear())
  const [viewMonth, setViewMonth] = useState(initDate.getMonth())
  const [swipeDir, setSwipeDir] = useState(0)
  const [touchStart, setTouchStart] = useState<number | null>(null)
  const [showPicker, setShowPicker] = useState(false)

  const todayISO = new Date().toISOString().split('T')[0]
  const daysInMonth = new Date(viewYear, viewMonth + 1, 0).getDate()
  const firstDow = (new Date(viewYear, viewMonth, 1).getDay() + 6) % 7
  const days: (number | null)[] = [...Array(firstDow).fill(null)]
  for (let d = 1; d <= daysInMonth; d++) days.push(d)
  while (days.length % 7 !== 0) days.push(null)

  function go(dir: -1 | 1) {
    setSwipeDir(dir)
    const m = viewMonth + dir
    if (m < 0) { setViewMonth(11); setViewYear(viewYear - 1) }
    else if (m > 11) { setViewMonth(0); setViewYear(viewYear + 1) }
    else setViewMonth(m)
  }

  function handleTouchStart(e: React.TouchEvent) { setTouchStart(e.touches[0].clientX) }
  function handleTouchEnd(e: React.TouchEvent) {
    if (touchStart === null) return
    const dx = e.changedTouches[0].clientX - touchStart
    if (Math.abs(dx) > 50) go(dx > 0 ? -1 : 1)
    setTouchStart(null)
  }

  const monthName = new Date(viewYear, viewMonth).toLocaleDateString(undefined, { month: 'long' })
  const weekDays = ['Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa', 'Su']
  const MONTHS = Array.from({ length: 12 }, (_, i) =>
    new Date(2024, i).toLocaleDateString(undefined, { month: 'short' })
  )

  return (
    <div>
      {/* Month nav */}
      <div className="flex items-center justify-between mb-2 px-1">
        <button
          onClick={() => go(-1)}
          className="w-10 h-10 flex items-center justify-center rounded-full active:bg-accent/10 text-muted active:text-accent transition-colors"
        >
          <CaretLeft size={18} weight="bold" />
        </button>
        <button
          onClick={() => setShowPicker(!showPicker)}
          className="px-3 py-1.5 rounded-full active:bg-accent/10 transition-colors flex items-center gap-1"
        >
          <span className="text-[13px] font-bold text-text capitalize">{monthName}</span>
          <span className="text-[13px] font-bold text-accent">{viewYear}</span>
        </button>
        <button
          onClick={() => go(1)}
          className="w-10 h-10 flex items-center justify-center rounded-full active:bg-accent/10 text-muted active:text-accent transition-colors"
        >
          <CaretRight size={18} weight="bold" />
        </button>
      </div>

      {/* Month+Year picker grid */}
      <AnimatePresence>
        {showPicker && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2, ease: 'easeOut' }}
            className="overflow-hidden"
          >
            <div className="pb-3">
              <div className="flex items-center justify-center gap-1 mb-2">
                <button
                  onClick={() => setViewYear(viewYear - 1)}
                  className="w-9 h-9 flex items-center justify-center rounded-full active:bg-accent/10 text-muted active:text-accent transition-colors"
                >
                  <CaretLeft size={14} weight="bold" />
                </button>
                <span className="text-sm font-bold text-text w-14 text-center tabular-nums">{viewYear}</span>
                <button
                  onClick={() => setViewYear(viewYear + 1)}
                  className="w-9 h-9 flex items-center justify-center rounded-full active:bg-accent/10 text-muted active:text-accent transition-colors"
                >
                  <CaretRight size={14} weight="bold" />
                </button>
              </div>
              <div className="grid grid-cols-4 gap-1.5 px-1">
                {MONTHS.map((m, i) => (
                  <button
                    key={i}
                    onClick={() => { setViewMonth(i); setShowPicker(false) }}
                    className={`
                      h-9 rounded-full text-[12px] font-semibold transition-all capitalize
                      ${i === viewMonth
                        ? 'bg-accent text-accent-text shadow-sm'
                        : 'text-text active:bg-accent/10'
                      }
                    `}
                  >
                    {m}
                  </button>
                ))}
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Weekday headers + day grid */}
      {!showPicker && (
        <>
          <div className="grid grid-cols-7 mb-1">
            {weekDays.map((d, i) => (
              <div key={i} className="text-center text-[10px] font-semibold text-muted/50 py-1">{d}</div>
            ))}
          </div>

          <AnimatePresence mode="wait" initial={false}>
            <motion.div
              key={`${viewYear}-${viewMonth}`}
              className="grid grid-cols-7"
              initial={{ opacity: 0, x: swipeDir * 60 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: swipeDir * -60 }}
              transition={{ duration: 0.15, ease: 'easeOut' }}
              onTouchStart={handleTouchStart}
              onTouchEnd={handleTouchEnd}
            >
              {days.map((d, i) => {
                if (d === null) return <div key={i} className="h-11" />
                const iso = fmtISO(viewYear, viewMonth, d)
                const isFrom = iso === from
                const isTo = iso === to
                const isSelected = isFrom || isTo
                const isToday = iso === todayISO && !isSelected
                const inRange = !singleDate && from && to && iso > from && iso < to

                return (
                  <button
                    key={i}
                    onClick={() => onSelect(iso)}
                    className="h-11 flex items-center justify-center relative"
                  >
                    {inRange && (
                      <div className="absolute inset-y-1 inset-x-0 bg-accent/8" />
                    )}
                    {!singleDate && isFrom && to && from < to && (
                      <div className="absolute inset-y-1 left-1/2 right-0 bg-accent/8" />
                    )}
                    {!singleDate && isTo && from && from < to && (
                      <div className="absolute inset-y-1 left-0 right-1/2 bg-accent/8" />
                    )}
                    <span
                      className={`
                        relative z-10 w-9 h-9 flex items-center justify-center rounded-full text-[13px] font-semibold transition-all
                        ${isSelected
                          ? 'bg-accent text-accent-text shadow-sm'
                          : isToday
                            ? 'text-accent font-bold ring-1 ring-accent/30'
                            : 'text-text active:bg-accent/10'
                        }
                      `}
                    >
                      {d}
                    </span>
                  </button>
                )
              })}
            </motion.div>
          </AnimatePresence>
        </>
      )}
    </div>
  )
}

/* ─── Single Date Picker Modal ─── */
export function SingleDateModal({
  value,
  onApply,
  onClose,
  applyLabel = 'Apply',
}: {
  value: string
  onApply: (iso: string) => void
  onClose: () => void
  applyLabel?: string
}) {
  const [selected, setSelected] = useState(value || new Date().toISOString().split('T')[0])

  return (
    <>
      <motion.div
        className="fixed inset-0 bg-black/40 z-[70]"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        onClick={onClose}
      />
      <motion.div
        className="fixed bottom-0 left-0 right-0 z-[70] bg-surface rounded-t-(--radius-card) px-5 pt-4 pb-8 max-h-[70dvh] overflow-y-auto"
        style={{ boxShadow: 'var(--shadow-modal)' }}
        initial={{ y: '100%' }}
        animate={{ y: 0 }}
        exit={{ y: '100%' }}
        transition={{ type: 'spring', damping: 28, stiffness: 300 }}
        drag="y"
        dragConstraints={{ top: 0 }}
        dragElastic={0.1}
        onDragEnd={(_, info) => {
          if (info.velocity.y > 500 || info.offset.y > 100) onClose()
        }}
      >
        <div className="w-10 h-1 rounded-full bg-border mx-auto mb-4" />

        <RangeCalendar
          from={selected}
          to={selected}
          activeField="from"
          singleDate
          onSelect={(iso) => setSelected(iso)}
        />

        <button
          onClick={() => { onApply(selected); onClose() }}
          className="w-full hero-gradient text-white font-bold text-sm py-2.5 rounded-2xl active:scale-[0.98] transition-transform mt-3"
          style={{ boxShadow: 'var(--shadow-hero)' }}
        >
          {applyLabel}
        </button>
      </motion.div>
    </>
  )
}

/* ─── Range Date Picker Modal ─── */
export function RangeDateModal({
  initialFrom,
  initialTo,
  onApply,
  onClose,
  labelFrom = 'From',
  labelTo = 'To',
  applyLabel = 'Apply',
}: {
  initialFrom: string
  initialTo: string
  onApply: (from: string, to: string) => void
  onClose: () => void
  labelFrom?: string
  labelTo?: string
  applyLabel?: string
}) {
  const [from, setFrom] = useState(initialFrom)
  const [to, setTo] = useState(initialTo)
  const [activeField, setActiveField] = useState<'from' | 'to'>('from')

  return (
    <>
      <motion.div
        className="fixed inset-0 bg-black/40 z-[70]"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        onClick={onClose}
      />
      <motion.div
        className="fixed bottom-0 left-0 right-0 z-[70] bg-surface rounded-t-(--radius-card) px-5 pt-4 pb-5 max-h-[70dvh] overflow-y-auto"
        style={{ boxShadow: 'var(--shadow-modal)' }}
        initial={{ y: '100%' }}
        animate={{ y: 0 }}
        exit={{ y: '100%' }}
        transition={{ type: 'spring', damping: 28, stiffness: 300 }}
        drag="y"
        dragConstraints={{ top: 0 }}
        dragElastic={0.1}
        onDragEnd={(_, info) => {
          if (info.velocity.y > 500 || info.offset.y > 100) onClose()
        }}
      >
        <div className="w-10 h-1 rounded-full bg-border mx-auto mb-3" />

        <div className="grid grid-cols-2 gap-2 mb-3">
          {(['from', 'to'] as const).map((field) => (
            <button
              key={field}
              onClick={() => setActiveField(field)}
              className={`px-3 py-2 rounded-2xl text-left transition-all ${
                activeField === field
                  ? 'bg-accent/10 ring-2 ring-accent'
                  : 'bg-bg'
              }`}
            >
              <span className="block text-[9px] font-bold text-muted uppercase tracking-widest">
                {field === 'from' ? labelFrom : labelTo}
              </span>
              <span className="block text-[13px] font-bold text-text mt-0.5">
                {fmtDisplay(field === 'from' ? from : to)}
              </span>
            </button>
          ))}
        </div>

        <RangeCalendar
          from={from}
          to={to}
          activeField={activeField}
          onSelect={(iso) => {
            if (activeField === 'from') {
              setFrom(iso)
              if (iso > to) setTo(iso)
              setActiveField('to')
            } else {
              setTo(iso)
              if (iso < from) setFrom(iso)
            }
          }}
        />

        <button
          onClick={() => { onApply(from, to); onClose() }}
          className="w-full hero-gradient text-white font-bold text-sm py-2.5 rounded-2xl active:scale-[0.98] transition-transform mt-3"
          style={{ boxShadow: 'var(--shadow-hero)' }}
        >
          {applyLabel}
        </button>
      </motion.div>
    </>
  )
}
