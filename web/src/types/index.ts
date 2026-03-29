export type TransactionType = 'expense' | 'income'

export interface Category {
  id: number
  name: string
  emoji: string
  type: string
  color: string
  is_system: boolean
}

// GET /api/v1/transactions
export interface Transaction {
  id: number
  type: TransactionType
  amount_cents: number
  currency_code: string
  category_id: number
  category_name: string
  category_emoji: string
  category_color: string
  note: string
  created_at: string
}

export interface ListTransactionsResponse {
  transactions: Transaction[]
  total_pages: number
  current_page: number
}

// GET /api/v1/balance
export interface BalanceCurrency {
  currency_code: string
  income_cents: number
  expense_cents: number
  net_cents: number
}

export interface DisplayConversion {
  currency_code: string
  net_cents: number
}

export interface BalanceResponse {
  by_currency: BalanceCurrency[]
  display_conversions: DisplayConversion[]
  total_in_base_cents: number
}

// GET /api/v1/stats
export interface CategoryStat {
  category_name: string
  category_emoji: string
  category_color: string
  type: TransactionType
  total_cents: number
  tx_count: number
  currency_code: string
}

export interface StatsResponse {
  period: string
  items: CategoryStat[]
}

// GET /api/v1/settings
export interface UserSettings {
  base_currency: string
  display_currencies: string[]
  language: string
}

// GET /api/v1/categories
export interface CategoriesResponse {
  categories: Category[]
}

// Budgets
export interface Budget {
  id: number
  category_id: number
  category_name: string
  category_emoji: string
  category_color: string
  limit_cents: number
  spent_cents: number
  period: string
  currency_code: string
  notify_at_percent: number
  usage_percent: number
  is_over_limit: boolean
}

export interface BudgetsResponse {
  budgets: Budget[]
}

// Recurring transactions
export interface RecurringTransaction {
  id: number
  type: TransactionType
  amount_cents: number
  currency_code: string
  category_id: number
  category_name: string
  category_emoji: string
  category_color: string
  note: string
  frequency: string
  next_run_at: string
  is_active: boolean
  created_at: string
}

export interface RecurringListResponse {
  recurring: RecurringTransaction[]
}

// Savings goals
export interface SavingsGoal {
  id: number
  name: string
  target_cents: number
  current_cents: number
  currency_code: string
  deadline: string | null
  progress_percent: number
  is_completed: boolean
  remaining_cents: number
  created_at: string
}

export interface GoalsResponse {
  goals: SavingsGoal[]
}
