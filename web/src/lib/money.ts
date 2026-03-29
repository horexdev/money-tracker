/** Format integer cents to human-readable string, e.g. 150050 → "1 500.50" */
export function formatCents(cents: number, currency = 'USD'): string {
  const amount = cents / 100
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    currencyDisplay: 'narrowSymbol',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount)
}

/** Map i18n language code to a BCP 47 locale for Intl APIs */
const LANG_TO_LOCALE: Record<string, string> = {
  en: 'en-US', ru: 'ru-RU', uk: 'uk-UA', be: 'be-BY',
  kk: 'kk-KZ', uz: 'uz-UZ', tr: 'tr-TR', ar: 'ar-SA',
  es: 'es-ES', pt: 'pt-BR', fr: 'fr-FR', de: 'de-DE',
  it: 'it-IT', nl: 'nl-NL', ko: 'ko-KR', ms: 'ms-MY', id: 'id-ID',
}

export function getLocale(lang: string): string {
  return LANG_TO_LOCALE[lang] ?? 'en-US'
}

/** Format a date using the user's language locale */
export function formatDate(
  date: Date | string,
  lang: string,
  options: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric' },
): string {
  return new Date(date).toLocaleDateString(getLocale(lang), options)
}

/** Parse a string like "1500.50" → 150050 cents */
export function parseCents(value: string): number {
  const cleaned = value.replace(/[^0-9.]/g, '')
  const n = parseFloat(cleaned)
  if (isNaN(n)) return 0
  return Math.round(n * 100)
}

/** Get the currency symbol for a currency code, e.g. "USD" → "$", "EUR" → "€" */
export function getCurrencySymbol(currency = 'USD'): string {
  const parts = new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    currencyDisplay: 'narrowSymbol',
  }).formatToParts(0)
  return parts.find(p => p.type === 'currency')?.value ?? currency
}
