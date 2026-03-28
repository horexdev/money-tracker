import { getInitDataRaw } from '../hooks/useTelegramApp'

export async function downloadExport(from: string, to: string, format = 'csv'): Promise<void> {
  const initData = getInitDataRaw()
  const res = await fetch(`/api/v1/export?format=${format}&from=${from}&to=${to}`, {
    headers: {
      'X-Telegram-Init-Data': initData,
    },
  })

  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText)
    throw new Error(text || `HTTP ${res.status}`)
  }

  const blob = await res.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `transactions_${from}_${to}.${format}`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}
