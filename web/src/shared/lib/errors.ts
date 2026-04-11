import type { TFunction } from 'i18next'

/** Maps known API error substrings to i18n translation keys. */
const ERROR_KEY_MAP: Array<[string, string]> = [
  ['category has transactions', 'errors.category_in_use'],
  ['system categories cannot be modified', 'errors.category_system_readonly'],
  ['budget already exists', 'errors.budget_already_exists'],
  ['account has transactions', 'errors.account_has_transactions'],
  ['insufficient funds', 'errors.insufficient_goal_funds'],
  ['transfer source and destination', 'errors.transfer_same_account'],
  ['exchange rate temporarily unavailable', 'errors.exchange_rate_unavailable'],
  ['maximum 3 display currencies', 'errors.too_many_currencies'],
  ['invalid currency', 'errors.invalid_currency'],
  ['invalid amount', 'errors.invalid_amount'],
  ['cannot delete the only account', 'errors.cannot_delete_last_account'],
  ['set a new default account', 'errors.must_set_new_default'],
  ['currency cannot be changed', 'errors.currency_immutable'],
]

/** Returns a localized, user-friendly message for an API error.
 *  Falls back to a generic message if the error is not recognized. */
export function friendlyError(error: unknown, t: TFunction): string {
  if (!error) return ''
  const raw = error instanceof Error ? error.message : String(error)
  const lower = raw.toLowerCase()

  for (const [substring, key] of ERROR_KEY_MAP) {
    if (lower.includes(substring)) {
      return t(key)
    }
  }
  return t('errors.something_went_wrong')
}
