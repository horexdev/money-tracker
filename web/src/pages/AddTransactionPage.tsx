import { useState, useCallback, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Check, CaretDown, MagnifyingGlass, X, CalendarBlank } from '@phosphor-icons/react'
import { motion, AnimatePresence } from 'framer-motion'
import { categoriesApi } from '../api/categories'
import { transactionsApi } from '../api/transactions'
import { balanceApi } from '../api/balance'
import { parseCents, getCurrencySymbol } from '../lib/money'
import { CategoryIcon } from '../lib/categoryIcons'
import { useTgMainButton } from '../hooks/useMainButton'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { useHaptic } from '../hooks/useHaptic'
import { Spinner } from '../components/Spinner'
import { PageTransition } from '../components/PageTransition'
import { useCategoryName } from '../hooks/useCategoryName'
import { SingleDateModal, fmtDisplay } from '../components/ui/DatePicker'
import type { TransactionType } from '../types'

const ALL_CURRENCIES = [
  'AED','AFN','ALL','AMD','ANG','AOA','ARS','AUD','AWG','AZN',
  'BAM','BBD','BDT','BGN','BHD','BMD','BND','BOB','BRL','BSD',
  'BWP','BYN','BZD','CAD','CDF','CHF','CLP','CNY','COP','CRC',
  'CUP','CVE','CZK','DJF','DKK','DOP','DZD','EGP','ETB','EUR',
  'FJD','GBP','GEL','GHS','GMD','GTQ','GYD','HKD','HNL','HRK',
  'HTG','HUF','IDR','ILS','INR','IQD','IRR','ISK','JMD','JOD',
  'JPY','KES','KGS','KHR','KRW','KWD','KZT','LAK','LBP','LKR',
  'LYD','MAD','MDL','MKD','MMK','MNT','MOP','MRU','MUR','MVR',
  'MWK','MXN','MYR','MZN','NAD','NGN','NIO','NOK','NPR','NZD',
  'OMR','PAB','PEN','PGK','PHP','PKR','PLN','PYG','QAR','RON',
  'RSD','RUB','RWF','SAR','SBD','SCR','SDG','SEK','SGD','SLL',
  'SOS','SRD','SZL','THB','TJS','TMT','TND','TOP','TRY','TTD',
  'TWD','TZS','UAH','UGX','USD','UYU','UZS','VES','VND','VUV',
  'WST','XAF','XCD','XOF','XPF','YER','ZAR','ZMW',
]

/** Only allow digits and a single dot with up to 2 decimal places */
function sanitizeAmount(value: string): string {
  // Strip everything except digits and dots
  let cleaned = value.replace(/[^0-9.]/g, '')
  // Only allow one dot
  const dotIndex = cleaned.indexOf('.')
  if (dotIndex !== -1) {
    cleaned = cleaned.slice(0, dotIndex + 1) + cleaned.slice(dotIndex + 1).replace(/\./g, '')
  }
  // Limit to 2 decimal places
  if (dotIndex !== -1 && cleaned.length - dotIndex > 3) {
    cleaned = cleaned.slice(0, dotIndex + 3)
  }
  // Prevent leading zeros (except "0." for decimals)
  if (cleaned.length > 1 && cleaned[0] === '0' && cleaned[1] !== '.') {
    cleaned = cleaned.slice(1)
  }
  return cleaned
}

export function AddTransactionPage() {
  const { t } = useTranslation()
  const tCategory = useCategoryName()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { selection, notification } = useHaptic()

  const [type, setType] = useState<TransactionType>('expense')
  const [amount, setAmount] = useState('')
  const [categoryID, setCategoryID] = useState<number | null>(null)
  const [note, setNote] = useState('')
  const [showCurrencyPicker, setShowCurrencyPicker] = useState(false)
  const [currencySearch, setCurrencySearch] = useState('')
  const [selectedCurrency, setSelectedCurrency] = useState<string | null>(null)
  const [txDate, setTxDate] = useState<string>(new Date().toISOString().split('T')[0])
  const [showDatePicker, setShowDatePicker] = useState(false)

  const { data: catData, isLoading } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  })

  const { data: balanceData } = useQuery({
    queryKey: ['balance'],
    queryFn: balanceApi.get,
  })

  const baseCurrency = balanceData?.by_currency?.[0]?.currency_code ?? 'USD'
  const currencyCode = selectedCurrency ?? baseCurrency
  const currencySymbol = getCurrencySymbol(currencyCode)

  const filteredCurrencies = useMemo(() => {
    const q = currencySearch.toUpperCase().trim()
    if (!q) return ALL_CURRENCIES
    return ALL_CURRENCIES.filter((c) => c.includes(q))
  }, [currencySearch])

  const filtered = (catData?.categories ?? []).filter(
    (c) => c.type === type || c.type === 'both'
  )

  const mutation = useMutation({
    mutationFn: transactionsApi.create,
    onSuccess: () => {
      notification('success')
      qc.invalidateQueries({ queryKey: ['transactions'] })
      qc.invalidateQueries({ queryKey: ['balance'] })
      qc.invalidateQueries({ queryKey: ['stats'] })
      navigate('/')
    },
    onError: () => notification('error'),
  })

  const canSubmit = parseCents(amount) > 0 && categoryID !== null && !mutation.isPending

  const handleSubmit = useCallback(() => {
    if (!canSubmit || categoryID === null) return
    const today = new Date().toISOString().split('T')[0]
    mutation.mutate({
      category_id: categoryID,
      type,
      amount_cents: parseCents(amount),
      note: note.trim() || undefined,
      currency_code: currencyCode,
      created_at: txDate !== today ? txDate : undefined,
    })
  }, [canSubmit, categoryID, type, amount, note, currencyCode, mutation])

  const handleAmountChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setAmount(sanitizeAmount(e.target.value))
  }, [])

  useTgBackButton(() => navigate('/'), true)
  useTgMainButton({
    text: mutation.isPending ? t('common.loading') : t('common.save'),
    onClick: handleSubmit,
    enabled: canSubmit,
    loading: mutation.isPending,
  })

  const isExpense = type === 'expense'

  return (
    <PageTransition>
      <div className="flex flex-col h-[calc(100dvh-var(--tab-bar-h))]">

        {/* Hero — type toggle + amount */}
        <div
          className="mx-4 mt-4 rounded-card p-5 pb-6 relative overflow-hidden shrink-0"
          style={{
            background: isExpense
              ? 'linear-gradient(135deg, #7f1d1d 0%, #ef4444 50%, #f87171 100%)'
              : 'linear-gradient(135deg, #14532d 0%, #22c55e 50%, #4ade80 100%)',
            boxShadow: isExpense
              ? '0 8px 32px rgba(239,68,68,0.3), 0 2px 8px rgba(0,0,0,0.1)'
              : '0 8px 32px rgba(34,197,94,0.3), 0 2px 8px rgba(0,0,0,0.1)',
          }}
        >
          <div className="absolute -top-10 -right-10 w-32 h-32 rounded-full bg-white/[0.06] blur-xl pointer-events-none" />
          <div className="absolute -bottom-8 -left-8 w-24 h-24 rounded-full bg-white/10 blur-2xl pointer-events-none" />

          <div className="relative z-10">
            {/* Type toggle — glass pills */}
            <div className="inline-flex bg-white/10 backdrop-blur-sm rounded-2xl p-1 gap-1 border border-white/[0.08]">
              {(['expense', 'income'] as TransactionType[]).map((v) => (
                <button
                  key={v}
                  onClick={() => { setType(v); setCategoryID(null); selection() }}
                  className={`
                    px-5 py-2 rounded-xl text-xs font-bold transition-all duration-200 select-none
                    ${type === v
                      ? 'bg-white/20 text-white shadow-[0_2px_8px_rgba(0,0,0,0.15)]'
                      : 'text-white/50'
                    }
                  `}
                >
                  {v === 'expense' ? t('transactions.expense') : t('transactions.income')}
                </button>
              ))}
            </div>

            {/* Amount input */}
            <div className="mt-4 flex items-baseline gap-1">
              <button
                onClick={() => setShowCurrencyPicker(true)}
                className="flex items-center gap-1 text-white/50 hover:text-white/80 transition-colors shrink-0"
              >
                <span className="text-3xl font-bold">{currencySymbol}</span>
                <CaretDown size={14} weight="bold" className="mb-0.5" />
              </button>
              <input
                inputMode="decimal"
                placeholder="0.00"
                value={amount}
                onChange={handleAmountChange}
                autoFocus
                className="flex-1 bg-transparent text-white text-4xl font-extrabold outline-none tabular-nums placeholder:text-white/25 min-w-0"
              />
            </div>
            {selectedCurrency && selectedCurrency !== baseCurrency && (
              <p className="text-white/40 text-xs font-medium mt-1">{selectedCurrency}</p>
            )}
          </div>
        </div>

        {/* Note + Date */}
        <div className="mx-4 mt-3 card-elevated overflow-hidden shrink-0">
          <div className="px-4 py-3 flex items-center gap-3 border-b border-border">
            <span className="text-[11px] font-bold text-muted uppercase tracking-widest shrink-0">
              {t('transactions.note')}
            </span>
            <input
              type="text"
              placeholder={t('transactions.note_placeholder')}
              value={note}
              onChange={(e) => setNote(e.target.value)}
              maxLength={120}
              className="flex-1 bg-transparent text-sm text-text outline-none min-w-0"
            />
          </div>
          <button
            onClick={() => setShowDatePicker(true)}
            className="w-full px-4 py-3 flex items-center gap-3 active:bg-accent-subtle/30 transition-colors"
          >
            <CalendarBlank size={16} weight="bold" className="text-muted shrink-0" />
            <span className="text-[11px] font-bold text-muted uppercase tracking-widest shrink-0">
              {t('transactions.date')}
            </span>
            <span className="flex-1 text-sm text-text text-right">{fmtDisplay(txDate)}</span>
          </button>
        </div>

        {/* Categories — scrollable grid */}
        <div className="flex-1 min-h-0 mt-3 flex flex-col">
          <p className="px-5 mb-2 text-[11px] font-bold text-muted uppercase tracking-widest shrink-0">
            {t('transactions.category')}
          </p>
          <div className="flex-1 min-h-0 overflow-y-auto no-scrollbar px-4" style={{ paddingBottom: 'calc(72px + env(safe-area-inset-bottom, 0px) + 16px)' }}>
            {isLoading ? (
              <div className="flex justify-center py-8"><Spinner /></div>
            ) : (
              <div className="grid grid-cols-4 gap-2">
                {filtered.map((cat) => {
                  const isSelected = categoryID === cat.id
                  return (
                    <button
                      key={cat.id}
                      onClick={() => { setCategoryID(cat.id); selection() }}
                      className={`
                        flex flex-col items-center justify-center gap-1.5 py-3 rounded-2xl
                        text-[11px] font-semibold transition-all duration-150 active:scale-[0.93] relative
                        ${isSelected
                          ? isExpense
                            ? 'bg-expense/10 text-expense shadow-[0_2px_12px_rgba(239,68,68,0.2)]'
                            : 'bg-income/10 text-income shadow-[0_2px_12px_rgba(34,197,94,0.2)]'
                          : 'bg-surface text-text shadow-sm'
                        }
                      `}
                    >
                      {isSelected && (
                        <div className={`absolute top-1.5 right-1.5 w-4 h-4 rounded-full flex items-center justify-center ${
                          isExpense ? 'bg-expense' : 'bg-income'
                        }`}>
                          <Check size={10} weight="bold" className="text-white" />
                        </div>
                      )}
                      <div
                        className="w-10 h-10 rounded-2xl flex items-center justify-center"
                        style={{ background: isSelected
                          ? (isExpense ? 'var(--color-expense)' : 'var(--color-income)')
                          : (cat.color || 'var(--color-accent)')
                        }}
                      >
                        <CategoryIcon
                          emoji={cat.emoji}
                          size={20}
                          weight="fill"
                          className="text-white"
                        />
                      </div>
                      <span className="text-center leading-tight px-0.5 truncate w-full">
                        {tCategory(cat.name)}
                      </span>
                    </button>
                  )
                })}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Save button — fixed above tab bar */}
      <div
        className="fixed left-0 right-0 px-4 z-10"
        style={{ bottom: 'calc(var(--tab-bar-h) + 8px)' }}
      >
        <button
          onClick={handleSubmit}
          disabled={!canSubmit}
          className={`
            w-full py-4 rounded-2xl text-[15px] font-bold transition-all active:scale-[0.98]
            ${canSubmit
              ? 'bg-accent text-accent-text shadow-[0_4px_16px_rgba(99,102,241,0.35)]'
              : 'bg-border text-muted'
            }
          `}
        >
          {mutation.isPending ? t('common.loading') : t('common.save')}
        </button>
      </div>
      {/* Currency picker bottom sheet */}
      <AnimatePresence>
        {showCurrencyPicker && (
          <>
            <motion.div
              className="fixed inset-0 bg-black/40 z-40"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={() => { setShowCurrencyPicker(false); setCurrencySearch('') }}
            />
            <motion.div
              className="fixed bottom-0 left-0 right-0 z-[60] bg-surface rounded-t-card flex flex-col"
              style={{ maxHeight: '75dvh' }}
              initial={{ y: '100%' }}
              animate={{ y: 0 }}
              exit={{ y: '100%' }}
              transition={{ type: 'spring', stiffness: 350, damping: 35 }}
              drag="y"
              dragConstraints={{ top: 0 }}
              dragElastic={0.1}
              onDragEnd={(_, info) => {
                if (info.velocity.y > 300 || info.offset.y > 120) {
                  setShowCurrencyPicker(false)
                  setCurrencySearch('')
                }
              }}
            >
              <div className="flex justify-center pt-3 pb-1 shrink-0">
                <div className="w-10 h-1 rounded-full bg-border" />
              </div>
              <div className="flex items-center justify-between px-5 py-3 shrink-0">
                <span className="text-base font-bold text-text">{t('add.select_currency')}</span>
                <button
                  onClick={() => { setShowCurrencyPicker(false); setCurrencySearch('') }}
                  className="w-8 h-8 rounded-full bg-accent-subtle flex items-center justify-center text-muted"
                >
                  <X size={14} weight="bold" />
                </button>
              </div>
              <div className="px-4 pb-2 shrink-0">
                <div className="relative">
                  <MagnifyingGlass size={14} weight="bold" className="absolute left-3.5 top-1/2 -translate-y-1/2 text-muted" />
                  <input
                    type="text"
                    value={currencySearch}
                    onChange={(e) => setCurrencySearch(e.target.value)}
                    placeholder={`${t('common.search')}...`}
                    className="w-full bg-bg rounded-2xl pl-9 pr-9 py-2.5 text-xs font-medium outline-none text-text placeholder:text-muted/50 shadow-sm focus:shadow-[0_0_0_2px_rgba(99,102,241,0.2)] transition-shadow"
                  />
                  {currencySearch && (
                    <button
                      onClick={() => setCurrencySearch('')}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted"
                    >
                      <X size={12} weight="bold" />
                    </button>
                  )}
                </div>
              </div>
              <div className="flex-1 min-h-0 overflow-y-auto no-scrollbar pb-[var(--tab-bar-h)]">
                <div className="divide-y divide-border">
                  {filteredCurrencies.map((c) => {
                    const isActive = currencyCode === c
                    return (
                      <button
                        key={c}
                        onClick={() => {
                          setSelectedCurrency(c)
                          setShowCurrencyPicker(false)
                          setCurrencySearch('')
                          selection()
                        }}
                        className={`w-full flex items-center justify-between px-5 py-3 text-left transition-colors active:bg-border ${
                          isActive ? 'bg-accent-subtle' : ''
                        }`}
                      >
                        <span className="text-[13px] font-semibold text-text">{c}</span>
                        {isActive && <Check size={14} weight="bold" className="text-accent" />}
                      </button>
                    )
                  })}
                </div>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>

      <AnimatePresence>
        {showDatePicker && (
          <SingleDateModal
            value={txDate}
            onApply={(iso) => setTxDate(iso)}
            onClose={() => setShowDatePicker(false)}
            applyLabel={t('common.done')}
          />
        )}
      </AnimatePresence>

    </PageTransition>
  )
}
