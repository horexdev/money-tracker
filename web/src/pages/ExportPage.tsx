import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Download } from 'lucide-react'
import { downloadExport } from '../api/export'
import { PageTransition } from '../components/PageTransition'
import { useTgBackButton } from '../hooks/useTelegramApp'
import { Card, Button } from '../components/ui'

export function ExportPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  useTgBackButton(() => navigate('/more'))

  const today = new Date().toISOString().split('T')[0]
  const monthAgo = new Date(Date.now() - 30 * 86400000).toISOString().split('T')[0]

  const [from, setFrom] = useState(monthAgo)
  const [to, setTo] = useState(today)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function handleDownload() {
    setLoading(true)
    setError('')
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
        <h1 className="text-xl font-bold">{t('export.title')}</h1>

        <Card>
          <div className="space-y-4">
            <div>
              <label className="block text-xs text-muted mb-1">{t('export.format')}</label>
              <div className="bg-surface rounded-[--radius-sm] px-3 py-2 text-sm text-text">
                CSV
              </div>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs text-muted mb-1">{t('export.from')}</label>
                <input
                  type="date"
                  value={from}
                  onChange={e => setFrom(e.target.value)}
                  className="w-full bg-surface rounded-[--radius-sm] px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-accent"
                />
              </div>
              <div>
                <label className="block text-xs text-muted mb-1">{t('export.to')}</label>
                <input
                  type="date"
                  value={to}
                  onChange={e => setTo(e.target.value)}
                  className="w-full bg-surface rounded-[--radius-sm] px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-accent"
                />
              </div>
            </div>

            <Button className="w-full" onClick={handleDownload} disabled={!canDownload}>
              <Download size={16} className="mr-2" />
              {loading ? t('common.loading') : t('export.download')}
            </Button>

            {error && <p className="text-xs text-destructive text-center">{error}</p>}
          </div>
        </Card>
      </div>
    </PageTransition>
  )
}
