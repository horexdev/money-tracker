/** Format integer cents to human-readable string, e.g. 150050 → "1 500.50" */
export function formatCents(cents: number, currency = 'USD'): string {
  const amount = cents / 100
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount)
}

/** Parse a string like "1500.50" → 150050 cents */
export function parseCents(value: string): number {
  const cleaned = value.replace(/[^0-9.]/g, '')
  const n = parseFloat(cleaned)
  if (isNaN(n)) return 0
  return Math.round(n * 100)
}
