import type {
  BalanceResponse,
  ListTransactionsResponse,
  CategoriesResponse,
  StatsResponse,
  BudgetsResponse,
  RecurringListResponse,
  GoalsResponse,
  UserSettings,
} from '../types'

export const mockBalance: BalanceResponse = {
  by_currency: [
    {
      currency_code: 'USD',
      income_cents: 450000,
      expense_cents: 187500,
      net_cents: 262500,
    },
    {
      currency_code: 'EUR',
      income_cents: 85000,
      expense_cents: 32000,
      net_cents: 53000,
    },
  ],
  display_conversions: [],
  total_in_base_cents: 262500 + Math.round(53000 * 1.08), // EUR→USD at ~1.08
}

const now = new Date()
function daysAgo(n: number): string {
  return new Date(now.getTime() - n * 86400000).toISOString()
}

export const mockTransactions: ListTransactionsResponse = {
  transactions: [
    { id: 1, type: 'expense', amount_cents: 4500, currency_code: 'USD', category_id: 1, category_name: 'Food', category_emoji: '🍔', note: 'Lunch at cafe', created_at: daysAgo(0) },
    { id: 2, type: 'expense', amount_cents: 12000, currency_code: 'USD', category_id: 2, category_name: 'Transport', category_emoji: '🚕', note: 'Uber ride', created_at: daysAgo(0) },
    { id: 3, type: 'income', amount_cents: 250000, currency_code: 'USD', category_id: 6, category_name: 'Salary', category_emoji: '💰', note: 'Monthly salary', created_at: daysAgo(1) },
    { id: 4, type: 'expense', amount_cents: 8900, currency_code: 'USD', category_id: 3, category_name: 'Entertainment', category_emoji: '🎬', note: 'Cinema tickets', created_at: daysAgo(1) },
    { id: 5, type: 'expense', amount_cents: 35000, currency_code: 'USD', category_id: 4, category_name: 'Shopping', category_emoji: '🛍️', note: 'New shoes', created_at: daysAgo(2) },
    { id: 6, type: 'expense', amount_cents: 6200, currency_code: 'USD', category_id: 1, category_name: 'Food', category_emoji: '🍔', note: 'Groceries', created_at: daysAgo(2) },
    { id: 7, type: 'income', amount_cents: 50000, currency_code: 'USD', category_id: 7, category_name: 'Freelance', category_emoji: '💻', note: 'Design project', created_at: daysAgo(3) },
    { id: 8, type: 'expense', amount_cents: 15000, currency_code: 'USD', category_id: 5, category_name: 'Health', category_emoji: '💊', note: 'Pharmacy', created_at: daysAgo(3) },
    { id: 9, type: 'expense', amount_cents: 75000, currency_code: 'USD', category_id: 8, category_name: 'Rent', category_emoji: '🏠', note: 'Monthly rent', created_at: daysAgo(5) },
    { id: 10, type: 'expense', amount_cents: 3200, currency_code: 'USD', category_id: 9, category_name: 'Coffee', category_emoji: '☕', note: 'Starbucks', created_at: daysAgo(5) },
    { id: 11, type: 'income', amount_cents: 150000, currency_code: 'USD', category_id: 7, category_name: 'Freelance', category_emoji: '💻', note: 'App development', created_at: daysAgo(7) },
    { id: 12, type: 'expense', amount_cents: 22000, currency_code: 'USD', category_id: 10, category_name: 'Bills', category_emoji: '📱', note: 'Phone + internet', created_at: daysAgo(8) },
    { id: 13, type: 'income', amount_cents: 85000, currency_code: 'EUR', category_id: 7, category_name: 'Freelance', category_emoji: '💻', note: 'EU client project', created_at: daysAgo(4) },
    { id: 14, type: 'expense', amount_cents: 32000, currency_code: 'EUR', category_id: 4, category_name: 'Shopping', category_emoji: '🛍️', note: 'Online order', created_at: daysAgo(6) },
  ],
  total_pages: 2,
  current_page: 1,
}

export const mockCategories: CategoriesResponse = {
  categories: [
    { id: 1, name: 'Food', emoji: '🍔', type: 'expense', is_system: true },
    { id: 2, name: 'Transport', emoji: '🚕', type: 'expense', is_system: true },
    { id: 3, name: 'Entertainment', emoji: '🎬', type: 'expense', is_system: true },
    { id: 4, name: 'Shopping', emoji: '🛍️', type: 'expense', is_system: true },
    { id: 5, name: 'Health', emoji: '💊', type: 'expense', is_system: true },
    { id: 6, name: 'Salary', emoji: '💰', type: 'income', is_system: true },
    { id: 7, name: 'Freelance', emoji: '💻', type: 'income', is_system: false },
    { id: 8, name: 'Rent', emoji: '🏠', type: 'expense', is_system: false },
    { id: 9, name: 'Coffee', emoji: '☕', type: 'expense', is_system: false },
    { id: 10, name: 'Bills', emoji: '📱', type: 'both', is_system: false },
  ],
}

export const mockStats: StatsResponse = {
  period: 'month',
  items: [
    { category_name: 'Food', category_emoji: '🍔', type: 'expense', total_cents: 45000, tx_count: 8, currency_code: 'USD' },
    { category_name: 'Transport', category_emoji: '🚕', type: 'expense', total_cents: 32000, tx_count: 5, currency_code: 'USD' },
    { category_name: 'Entertainment', category_emoji: '🎬', type: 'expense', total_cents: 18000, tx_count: 3, currency_code: 'USD' },
    { category_name: 'Shopping', category_emoji: '🛍️', type: 'expense', total_cents: 35000, tx_count: 2, currency_code: 'USD' },
    { category_name: 'Rent', category_emoji: '🏠', type: 'expense', total_cents: 75000, tx_count: 1, currency_code: 'USD' },
    { category_name: 'Health', category_emoji: '💊', type: 'expense', total_cents: 15000, tx_count: 2, currency_code: 'USD' },
    { category_name: 'Salary', category_emoji: '💰', type: 'income', total_cents: 250000, tx_count: 1, currency_code: 'USD' },
    { category_name: 'Freelance', category_emoji: '💻', type: 'income', total_cents: 200000, tx_count: 3, currency_code: 'USD' },
  ],
}

export const mockBudgets: BudgetsResponse = {
  budgets: [
    { id: 1, category_id: 1, category_name: 'Food', category_emoji: '🍔', limit_cents: 60000, spent_cents: 45000, period: 'monthly', currency_code: 'USD', notify_at_percent: 80, usage_percent: 75, is_over_limit: false },
    { id: 2, category_id: 3, category_name: 'Entertainment', category_emoji: '🎬', limit_cents: 15000, spent_cents: 18000, period: 'monthly', currency_code: 'USD', notify_at_percent: 80, usage_percent: 120, is_over_limit: true },
    { id: 3, category_id: 2, category_name: 'Transport', category_emoji: '🚕', limit_cents: 40000, spent_cents: 32000, period: 'monthly', currency_code: 'USD', notify_at_percent: 80, usage_percent: 80, is_over_limit: false },
  ],
}

export const mockRecurring: RecurringListResponse = {
  recurring: [
    { id: 1, type: 'expense', amount_cents: 75000, currency_code: 'USD', category_id: 8, category_name: 'Rent', category_emoji: '🏠', note: 'Monthly rent', frequency: 'monthly', next_run_at: daysAgo(-5), is_active: true, created_at: daysAgo(60) },
    { id: 2, type: 'expense', amount_cents: 1500, currency_code: 'USD', category_id: 10, category_name: 'Bills', category_emoji: '📱', note: 'Spotify subscription', frequency: 'monthly', next_run_at: daysAgo(-12), is_active: true, created_at: daysAgo(90) },
    { id: 3, type: 'income', amount_cents: 250000, currency_code: 'USD', category_id: 6, category_name: 'Salary', category_emoji: '💰', note: 'Monthly salary', frequency: 'monthly', next_run_at: daysAgo(-2), is_active: true, created_at: daysAgo(120) },
    { id: 4, type: 'expense', amount_cents: 500, currency_code: 'USD', category_id: 9, category_name: 'Coffee', category_emoji: '☕', note: 'Daily coffee', frequency: 'daily', next_run_at: daysAgo(-1), is_active: false, created_at: daysAgo(30) },
  ],
}

export const mockGoals: GoalsResponse = {
  goals: [
    { id: 1, name: 'New MacBook', target_cents: 200000, current_cents: 135000, currency_code: 'USD', deadline: daysAgo(-60), progress_percent: 67.5, is_completed: false, remaining_cents: 65000, created_at: daysAgo(90) },
    { id: 2, name: 'Vacation Fund', target_cents: 500000, current_cents: 500000, currency_code: 'USD', deadline: null, progress_percent: 100, is_completed: true, remaining_cents: 0, created_at: daysAgo(180) },
    { id: 3, name: 'Emergency Fund', target_cents: 1000000, current_cents: 320000, currency_code: 'USD', deadline: daysAgo(-365), progress_percent: 32, is_completed: false, remaining_cents: 680000, created_at: daysAgo(30) },
  ],
}

export const mockSettings: UserSettings = {
  base_currency: 'USD',
  display_currencies: ['EUR', 'GBP'],
  language: 'en',
}
