import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { AnimatePresence } from 'framer-motion'
import { DownloadSimple, CalendarBlank } from '@phosphor-icons/react'
import { downloadExport } from '../api/export'
import { PageTransition } from '../components/PageTransition'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { Button, RangeDateModal, fmtDisplay } from '../components/ui'

export function ExportPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  useTgBackButton(() => navigate('/more'))

  const today    = new Date().toISOString().split('T')[0]
  const monthAgo = new Date(Date.now() - 30 * 86400000).toISOString().split('T')[0]

  const [from, setFrom]         = useState(monthAgo)
  const [to, setTo]             = useState(today)
  const [showPicker, setShowPicker] = useState(false)
  const [loading, setLoading]   = useState(false)
  const [error, setError]       = useState('')

  async function handleDownload() {
    setLoading(true); setError('')
    try {
      await downloadExport(from, to, 'csv')
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setLoading(false)
    }
  }

  const canDownload = from && to && !loading

  return (
    <PageTransition>
      <div className="p-4 space-y-4">
        <div className="card-elevated overflow-hidden">
          <div className="hero-gradient h-2 relative !rounded-none !overflow-visible" />
          <div className="p-5 space-y-4">

            <div>
              <label className="block text-xs font-semibold text-muted uppercase tracking-widest mb-2">
                {t('export.format')}
              </label>
              <div className="bg-bg rounded-2xl px-3 py-2.5 text-sm font-medium text-text">
                CSV
              </div>
            </div>

            {/* Date range selector */}
            <div>
              <label className="block text-xs font-semibold text-muted uppercase tracking-widest mb-2">
                {t('export.from')} — {t('export.to')}
              </label>
              <button
                onClick={() => setShowPicker(true)}
                className="w-full bg-bg rounded-2xl px-4 py-3 flex items-center justify-between active:bg-accent/5 transition-colors"
              >
                <div className="text-left">
                  <span className="text-sm font-bold text-text">{fmtDisplay(from)}</span>
                  <span className="text-sm text-muted mx-2">—</span>
                  <span className="text-sm font-bold text-text">{fmtDisplay(to)}</span>
                </div>
                <CalendarBlank size={18} weight="bold" className="text-muted shrink-0" />
              </button>
            </div>

            <Button size="lg" className="w-full" onClick={handleDownload} disabled={!canDownload}>
              <DownloadSimple size={18} weight="bold" className="mr-2" />
              {loading ? t('common.loading') : t('export.download')}
            </Button>

            {error && <p className="text-xs text-destructive text-center">{error}</p>}
          </div>
        </div>
      </div>

      <AnimatePresence>
        {showPicker && (
          <RangeDateModal
            initialFrom={from}
            initialTo={to}
            onApply={(f, t) => { setFrom(f); setTo(t) }}
            onClose={() => setShowPicker(false)}
            labelFrom={t('export.from')}
            labelTo={t('export.to')}
            applyLabel={t('stats.apply')}
          />
        )}
      </AnimatePresence>
    </PageTransition>
  )
}
