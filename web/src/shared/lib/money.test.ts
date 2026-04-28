import { describe, expect, it } from 'vitest'

import {
  formatCents,
  formatDate,
  getCurrencySymbol,
  getLocale,
  parseCents,
  sanitizeAmount,
} from './money'

describe('formatCents', () => {
  it('formats USD with the narrow symbol', () => {
    expect(formatCents(150050, 'USD')).toMatch(/\$1,500\.50/)
  })

  it('formats EUR with €', () => {
    expect(formatCents(150050, 'EUR')).toMatch(/€1,500\.50/)
  })

  it('formats zero with two fraction digits', () => {
    expect(formatCents(0, 'USD')).toMatch(/\$0\.00/)
  })

  it('uses minus prefix for negative amounts', () => {
    expect(formatCents(-1500, 'USD')).toMatch(/-\$15\.00/)
  })

  it('always renders exactly two fraction digits', () => {
    expect(formatCents(100, 'USD')).toMatch(/\$1\.00/)
    expect(formatCents(105, 'USD')).toMatch(/\$1\.05/)
  })

  it('defaults to USD when currency is omitted', () => {
    expect(formatCents(100)).toMatch(/\$1\.00/)
  })
})

describe('parseCents', () => {
  it('parses a clean decimal string', () => {
    expect(parseCents('1500.50')).toBe(150050)
  })

  it('parses an integer string', () => {
    expect(parseCents('100')).toBe(10000)
  })

  it('rounds to the nearest cent', () => {
    expect(parseCents('0.005')).toBe(1)
    expect(parseCents('0.004')).toBe(0)
  })

  it('strips non-numeric characters before parsing', () => {
    expect(parseCents('$1500.50')).toBe(150050)
  })

  it('returns 0 for inputs that do not contain a number', () => {
    expect(parseCents('abc')).toBe(0)
    expect(parseCents('')).toBe(0)
  })
})

describe('sanitizeAmount', () => {
  it('keeps a clean decimal as-is', () => {
    expect(sanitizeAmount('1500.50')).toBe('1500.50')
  })

  it('strips letters and other symbols', () => {
    expect(sanitizeAmount('1a2.3b4')).toBe('12.34')
  })

  it('keeps only the first decimal point', () => {
    expect(sanitizeAmount('1.2.3')).toBe('1.23')
  })

  it('limits the fractional part to two digits', () => {
    expect(sanitizeAmount('1.2345')).toBe('1.23')
  })

  it('strips a leading zero unless followed by a dot', () => {
    expect(sanitizeAmount('0123')).toBe('123')
    expect(sanitizeAmount('0.5')).toBe('0.5')
    expect(sanitizeAmount('0')).toBe('0')
  })

  it('returns an empty string for non-numeric input', () => {
    expect(sanitizeAmount('abc')).toBe('')
    expect(sanitizeAmount('')).toBe('')
  })
})

describe('getCurrencySymbol', () => {
  it('returns $ for USD', () => {
    expect(getCurrencySymbol('USD')).toBe('$')
  })

  it('returns € for EUR', () => {
    expect(getCurrencySymbol('EUR')).toBe('€')
  })

  it('returns the currency code for unknown codes', () => {
    expect(getCurrencySymbol('XYZ')).toBe('XYZ')
  })

  it('defaults to USD when called with no argument', () => {
    expect(getCurrencySymbol()).toBe('$')
  })
})

describe('getLocale', () => {
  it('maps known languages to BCP 47 locales', () => {
    expect(getLocale('en')).toBe('en-US')
    expect(getLocale('ru')).toBe('ru-RU')
    expect(getLocale('uk')).toBe('uk-UA')
    expect(getLocale('de')).toBe('de-DE')
  })

  it('falls back to en-US for unknown languages', () => {
    expect(getLocale('xx')).toBe('en-US')
    expect(getLocale('')).toBe('en-US')
  })
})

describe('formatDate', () => {
  it('formats Date objects with default options', () => {
    const date = new Date('2026-04-28T12:00:00Z')
    const result = formatDate(date, 'en')
    expect(result).toMatch(/Apr/)
    expect(result).toMatch(/28/)
  })

  it('accepts ISO strings', () => {
    expect(formatDate('2026-04-28T12:00:00Z', 'en')).toMatch(/Apr/)
  })

  it('honours custom format options', () => {
    const date = new Date('2026-04-28T12:00:00Z')
    const result = formatDate(date, 'en', { year: 'numeric', month: 'long' })
    expect(result).toContain('April')
    expect(result).toContain('2026')
  })
})
