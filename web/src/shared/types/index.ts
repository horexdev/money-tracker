export type TransactionType = 'expense' | 'income'

export interface Category {
  id: number
  name: string
  icon: string
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
  category_icon: string
  category_color: string
  note: string
  created_at: string
  account_id: number
  account_name?: string
  is_adjustment?: boolean
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
  category_id: number
  category_name: string
  category_icon: string
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
  is_admin: boolean
  notify_budget_alerts: boolean
  notify_recurring_reminders: boolean
  notify_weekly_summary: boolean
  notify_goal_milestones: boolean
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
  category_icon: string
  category_color: string
  limit_cents: number
  spent_cents: number
  period: string
  currency_code: string
  notify_at_percent: number
  notifications_enabled: boolean
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
  account_id: number
  category_id: number
  category_name: string
  category_icon: string
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
  account_id?: number | null
}

export interface GoalsResponse {
  goals: SavingsGoal[]
}

// Accounts
export type AccountType = 'checking' | 'savings' | 'cash' | 'credit' | 'crypto'

export interface Account {
  id: number
  name: string
  icon: string
  color: string
  type: AccountType
  currency_code: string
  is_default: boolean
  include_in_total: boolean
  balance_cents: number
  created_at: string
}

// Admin
export interface AdminUser {
  id: number
  username: string
  first_name: string
  last_name: string
  currency_code: string
  language: string
  created_at: string
}

export interface AdminUsersResponse {
  users: AdminUser[]
  total: number
  page: number
  page_size: number
}

export interface AdminStatsResponse {
  total_users: number
  new_today: number
  new_this_week: number
  new_this_month: number
  retention_day1: number
  retention_day7: number
  retention_day30: number
}

// Transfers
export interface Transfer {
  id: number
  from_account_id: number
  from_account_name: string
  to_account_id: number
  to_account_name: string
  amount_cents: number
  from_currency_code: string
  to_currency_code: string
  exchange_rate: number
  note: string
  created_at: string
}
