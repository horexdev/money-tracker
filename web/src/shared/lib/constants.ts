import {
  Bank, PiggyBank, Money, CreditCard, Coins,
  type Icon,
} from '@phosphor-icons/react'
import type { AccountType } from '../types'

/** Shared color palette for account and category color pickers */
export const COLOR_SWATCHES: string[] = [
  '#6366f1', '#8b5cf6', '#ec4899', '#ef4444',
  '#f97316', '#eab308', '#22c55e', '#10b981',
  '#14b8a6', '#06b6d4', '#3b82f6', '#64748b',
]

/**
 * Chart color palette for stats donut chart.
 * Mirrors the brand palette defined in index.css @theme.
 */
export const CHART_COLORS: string[] = [
  '#6366f1', '#22c55e', '#06b6d4', '#f59e0b',
  '#ef4444', '#8b5cf6', '#ec4899', '#f97316',
]

/** Ordered list of account types */
export const ACCOUNT_TYPES: AccountType[] = [
  'checking', 'savings', 'cash', 'credit', 'crypto',
]

/** Icon component per account type */
export const ACCOUNT_TYPE_ICONS: Record<AccountType, Icon> = {
  checking: Bank,
  savings:  PiggyBank,
  cash:     Money,
  credit:   CreditCard,
  crypto:   Coins,
}

/** Commonly-used currencies shown as chips at the top of the picker */
export const POPULAR_CURRENCIES: string[] = [
  'USD', 'EUR', 'GBP', 'UAH', 'RUB', 'TRY', 'KZT', 'UZS', 'BRL', 'JPY', 'TJS', 'CNY',
]

/** Full ISO 4217 currency list */
export const ALL_CURRENCIES: string[] = [
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

/** Supported application languages */
export const LANGUAGES: Array<{ code: string; label: string; native: string }> = [
  { code: 'en', label: 'English',    native: 'English'          },
  { code: 'ru', label: 'Russian',    native: 'Русский'          },
  { code: 'uk', label: 'Ukrainian',  native: 'Українська'       },
  { code: 'be', label: 'Belarusian', native: 'Беларуская'       },
  { code: 'kk', label: 'Kazakh',     native: 'Қазақша'          },
  { code: 'uz', label: 'Uzbek',      native: "O'zbek"           },
  { code: 'tr', label: 'Turkish',    native: 'Türkçe'           },
  { code: 'ar', label: 'Arabic',     native: 'العربية'          },
  { code: 'es', label: 'Spanish',    native: 'Español'          },
  { code: 'pt', label: 'Portuguese', native: 'Português'        },
  { code: 'fr', label: 'French',     native: 'Français'         },
  { code: 'de', label: 'German',     native: 'Deutsch'          },
  { code: 'it', label: 'Italian',    native: 'Italiano'         },
  { code: 'nl', label: 'Dutch',      native: 'Nederlands'       },
  { code: 'ko', label: 'Korean',     native: '한국어'            },
  { code: 'ms', label: 'Malay',      native: 'Bahasa Melayu'    },
  { code: 'id', label: 'Indonesian', native: 'Bahasa Indonesia' },
]
