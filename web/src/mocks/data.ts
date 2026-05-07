import type {
  BalanceResponse,
  ListTransactionsResponse,
  CategoriesResponse,
  StatsResponse,
  BudgetsResponse,
  RecurringListResponse,
  GoalsResponse,
  UserSettings,
  Account,
  Transfer,
  AdminUsersResponse,
  AdminStatsResponse,
} from '../shared/types'

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
    { id: 1, type: 'expense', amount_cents: 4500, currency_code: 'USD', category_id: 1, category_name: 'Food', category_icon: 'fork-knife', category_color: '#6366f1', note: 'Lunch at cafe', created_at: daysAgo(0), account_id: 1, account_name: 'Main Card' },
    { id: 2, type: 'expense', amount_cents: 12000, currency_code: 'USD', category_id: 2, category_name: 'Transport', category_icon: 'taxi', category_color: '#8b5cf6', note: 'Uber ride', created_at: daysAgo(0), account_id: 1, account_name: 'Main Card' },
    { id: 3, type: 'income', amount_cents: 250000, currency_code: 'USD', category_id: 6, category_name: 'Salary', category_icon: 'money', category_color: '#10b981', note: 'Monthly salary', created_at: daysAgo(1), account_id: 1, account_name: 'Main Card' },
    { id: 4, type: 'expense', amount_cents: 8900, currency_code: 'USD', category_id: 3, category_name: 'Entertainment', category_icon: 'film-slate', category_color: '#ec4899', note: 'Cinema tickets', created_at: daysAgo(1), account_id: 1, account_name: 'Main Card' },
    { id: 5, type: 'expense', amount_cents: 35000, currency_code: 'USD', category_id: 4, category_name: 'Shopping', category_icon: 'shopping-bag', category_color: '#ef4444', note: 'New shoes', created_at: daysAgo(2), account_id: 2, account_name: 'Savings' },
    { id: 6, type: 'expense', amount_cents: 6200, currency_code: 'USD', category_id: 1, category_name: 'Food', category_icon: 'fork-knife', category_color: '#6366f1', note: 'Groceries', created_at: daysAgo(2), account_id: 1, account_name: 'Main Card' },
    { id: 7, type: 'income', amount_cents: 50000, currency_code: 'USD', category_id: 7, category_name: 'Freelance', category_icon: 'laptop', category_color: '#3b82f6', note: 'Design project', created_at: daysAgo(3), account_id: 2, account_name: 'Savings' },
    { id: 8, type: 'expense', amount_cents: 15000, currency_code: 'USD', category_id: 5, category_name: 'Health', category_icon: 'first-aid', category_color: '#22c55e', note: 'Pharmacy', created_at: daysAgo(3), account_id: 3, account_name: 'Cash' },
    { id: 9, type: 'expense', amount_cents: 75000, currency_code: 'USD', category_id: 8, category_name: 'Rent', category_icon: 'house', category_color: '#f97316', note: 'Monthly rent', created_at: daysAgo(5), account_id: 1, account_name: 'Main Card' },
    { id: 10, type: 'expense', amount_cents: 3200, currency_code: 'USD', category_id: 9, category_name: 'Coffee', category_icon: 'coffee', category_color: '#eab308', note: 'Starbucks', created_at: daysAgo(5), account_id: 3, account_name: 'Cash' },
    { id: 11, type: 'income', amount_cents: 150000, currency_code: 'USD', category_id: 7, category_name: 'Freelance', category_icon: 'laptop', category_color: '#3b82f6', note: 'App development', created_at: daysAgo(7), account_id: 2, account_name: 'Savings' },
    { id: 12, type: 'expense', amount_cents: 22000, currency_code: 'USD', category_id: 10, category_name: 'Bills', category_icon: 'device-mobile', category_color: '#64748b', note: 'Phone + internet', created_at: daysAgo(8), account_id: 1, account_name: 'Main Card' },
    { id: 13, type: 'income', amount_cents: 85000, currency_code: 'EUR', category_id: 7, category_name: 'Freelance', category_icon: 'laptop', category_color: '#3b82f6', note: 'EU client project', created_at: daysAgo(4), account_id: 1, account_name: 'Main Card' },
    { id: 14, type: 'expense', amount_cents: 32000, currency_code: 'EUR', category_id: 4, category_name: 'Shopping', category_icon: 'shopping-bag', category_color: '#ef4444', note: 'Online order', created_at: daysAgo(6), account_id: 1, account_name: 'Main Card' },
  ],
  total_pages: 2,
  current_page: 1,
}

export const mockCategories: CategoriesResponse = {
  categories: [
    { id: 1, name: 'Food', icon: 'fork-knife', type: 'expense', color: '#6366f1', is_system: true },
    { id: 2, name: 'Transport', icon: 'taxi', type: 'expense', color: '#8b5cf6', is_system: true },
    { id: 3, name: 'Entertainment', icon: 'film-slate', type: 'expense', color: '#ec4899', is_system: true },
    { id: 4, name: 'Shopping', icon: 'shopping-bag', type: 'expense', color: '#ef4444', is_system: true },
    { id: 5, name: 'Health', icon: 'first-aid', type: 'expense', color: '#22c55e', is_system: true },
    { id: 6, name: 'Salary', icon: 'money', type: 'income', color: '#10b981', is_system: true },
    { id: 7, name: 'Freelance', icon: 'laptop', type: 'income', color: '#3b82f6', is_system: false },
    { id: 8, name: 'Rent', icon: 'house', type: 'expense', color: '#f97316', is_system: false },
    { id: 9, name: 'Coffee', icon: 'coffee', type: 'expense', color: '#eab308', is_system: false },
    { id: 10, name: 'Bills', icon: 'device-mobile', type: 'both', color: '#64748b', is_system: false },
  ],
}

export const mockStats: StatsResponse = {
  period: 'month',
  items: [
    { category_id: 1, category_name: 'Food', category_icon: 'fork-knife', category_color: '#6366f1', type: 'expense', total_cents: 45000, tx_count: 8, currency_code: 'USD' },
    { category_id: 2, category_name: 'Transport', category_icon: 'taxi', category_color: '#8b5cf6', type: 'expense', total_cents: 32000, tx_count: 5, currency_code: 'USD' },
    { category_id: 3, category_name: 'Entertainment', category_icon: 'film-slate', category_color: '#ec4899', type: 'expense', total_cents: 18000, tx_count: 3, currency_code: 'USD' },
    { category_id: 4, category_name: 'Shopping', category_icon: 'shopping-bag', category_color: '#ef4444', type: 'expense', total_cents: 35000, tx_count: 2, currency_code: 'USD' },
    { category_id: 8, category_name: 'Rent', category_icon: 'house', category_color: '#f97316', type: 'expense', total_cents: 75000, tx_count: 1, currency_code: 'USD' },
    { category_id: 5, category_name: 'Health', category_icon: 'first-aid', category_color: '#22c55e', type: 'expense', total_cents: 15000, tx_count: 2, currency_code: 'USD' },
    { category_id: 6, category_name: 'Salary', category_icon: 'money', category_color: '#10b981', type: 'income', total_cents: 250000, tx_count: 1, currency_code: 'USD' },
    { category_id: 7, category_name: 'Freelance', category_icon: 'laptop', category_color: '#3b82f6', type: 'income', total_cents: 200000, tx_count: 3, currency_code: 'USD' },
  ],
}

export const mockBudgets: BudgetsResponse = {
  budgets: [
    { id: 1, category_id: 1, category_name: 'Food', category_icon: 'fork-knife', category_color: '#6366f1', limit_cents: 60000, spent_cents: 45000, period: 'monthly', currency_code: 'USD', notify_at_percent: 80, notifications_enabled: true, usage_percent: 75, is_over_limit: false },
    { id: 2, category_id: 3, category_name: 'Entertainment', category_icon: 'film-slate', category_color: '#ec4899', limit_cents: 15000, spent_cents: 18000, period: 'monthly', currency_code: 'USD', notify_at_percent: 80, notifications_enabled: true, usage_percent: 120, is_over_limit: true },
    { id: 3, category_id: 2, category_name: 'Transport', category_icon: 'taxi', category_color: '#8b5cf6', limit_cents: 40000, spent_cents: 32000, period: 'monthly', currency_code: 'USD', notify_at_percent: 80, notifications_enabled: true, usage_percent: 80, is_over_limit: false },
  ],
}

export const mockRecurring: RecurringListResponse = {
  recurring: [
    { id: 1, type: 'expense', amount_cents: 75000, currency_code: 'USD', account_id: 1, category_id: 8, category_name: 'Rent', category_icon: 'house', category_color: '#f97316', note: 'Monthly rent', frequency: 'monthly', next_run_at: daysAgo(-5), is_active: true, created_at: daysAgo(60) },
    { id: 2, type: 'expense', amount_cents: 1500, currency_code: 'USD', account_id: 1, category_id: 10, category_name: 'Bills', category_icon: 'device-mobile', category_color: '#64748b', note: 'Spotify subscription', frequency: 'monthly', next_run_at: daysAgo(-12), is_active: true, created_at: daysAgo(90) },
    { id: 3, type: 'income', amount_cents: 250000, currency_code: 'USD', account_id: 1, category_id: 6, category_name: 'Salary', category_icon: 'money', category_color: '#10b981', note: 'Monthly salary', frequency: 'monthly', next_run_at: daysAgo(-2), is_active: true, created_at: daysAgo(120) },
    { id: 4, type: 'expense', amount_cents: 500, currency_code: 'USD', account_id: 1, category_id: 9, category_name: 'Coffee', category_icon: 'coffee', category_color: '#eab308', note: 'Daily coffee', frequency: 'daily', next_run_at: daysAgo(-1), is_active: false, created_at: daysAgo(30) },
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
  is_admin: true,
  notify_budget_alerts: false,
  notify_recurring_reminders: false,
  notify_weekly_summary: false,
  notify_goal_milestones: false,
}

export const mockAccounts: { accounts: Account[] } = {
  accounts: [
    {
      id: 1,
      name: 'Main Card',
      icon: 'credit-card',
      color: '#6366f1',
      type: 'checking',
      currency_code: 'USD',
      is_default: true,
      include_in_total: true,
      balance_cents: 262500,
      created_at: daysAgo(90),
    },
    {
      id: 2,
      name: 'Savings',
      icon: 'piggy-bank',
      color: '#10b981',
      type: 'savings',
      currency_code: 'USD',
      is_default: false,
      include_in_total: true,
      balance_cents: 500000,
      created_at: daysAgo(60),
    },
    {
      id: 3,
      name: 'Cash',
      icon: 'money',
      color: '#f59e0b',
      type: 'cash',
      currency_code: 'USD',
      is_default: false,
      include_in_total: true,
      balance_cents: 15000,
      created_at: daysAgo(30),
    },
  ],
}

export const mockTransfers: { transfers: Transfer[]; total: number; limit: number; offset: number } = {
  transfers: [
    {
      id: 1,
      from_account_id: 1,
      from_account_name: 'Main Card',
      to_account_id: 2,
      to_account_name: 'Savings',
      amount_cents: 50000,
      from_currency_code: 'USD',
      to_currency_code: 'USD',
      exchange_rate: 1,
      note: 'Monthly savings',
      created_at: daysAgo(5),
    },
    {
      id: 2,
      from_account_id: 1,
      from_account_name: 'Main Card',
      to_account_id: 3,
      to_account_name: 'Cash',
      amount_cents: 10000,
      from_currency_code: 'USD',
      to_currency_code: 'USD',
      exchange_rate: 1,
      note: 'ATM withdrawal',
      created_at: daysAgo(12),
    },
  ],
  total: 2,
  limit: 50,
  offset: 0,
}

export const mockAdminStats: AdminStatsResponse = {
  total_users: 142,
  new_today: 3,
  new_this_week: 17,
  new_this_month: 54,
  retention_day1: 68.5,
  retention_day7: 41.2,
  retention_day30: 22.8,
}

export const mockAdminUsers: AdminUsersResponse = {
  users: [
    { id: 6554524765, username: 'horexdev', first_name: 'Alex', last_name: 'Dev', currency_code: 'USD', language: 'en', created_at: daysAgo(90) },
    { id: 123456789, username: 'ivan_petrov', first_name: 'Ivan', last_name: 'Petrov', currency_code: 'RUB', language: 'ru', created_at: daysAgo(45) },
    { id: 987654321, username: 'maria_k', first_name: 'Maria', last_name: 'Kovaleva', currency_code: 'UAH', language: 'uk', created_at: daysAgo(30) },
    { id: 111222333, username: '', first_name: 'John', last_name: 'Smith', currency_code: 'USD', language: 'en', created_at: daysAgo(21) },
    { id: 444555666, username: 'ali_hassan', first_name: 'Ali', last_name: 'Hassan', currency_code: 'USD', language: 'ar', created_at: daysAgo(14) },
    { id: 777888999, username: 'jung_min', first_name: 'Jung', last_name: 'Min', currency_code: 'USD', language: 'ko', created_at: daysAgo(7) },
    { id: 100200300, username: 'anna_m', first_name: 'Anna', last_name: 'Müller', currency_code: 'EUR', language: 'de', created_at: daysAgo(5) },
    { id: 400500600, username: 'pierre_d', first_name: 'Pierre', last_name: 'Dupont', currency_code: 'EUR', language: 'fr', created_at: daysAgo(3) },
    { id: 700800900, username: '', first_name: 'Sofia', last_name: '', currency_code: 'USD', language: 'es', created_at: daysAgo(2) },
    { id: 101010101, username: 'test_user', first_name: 'Test', last_name: 'User', currency_code: 'USD', language: 'en', created_at: daysAgo(1) },
  ],
  total: 142,
  page: 1,
  page_size: 20,
}
