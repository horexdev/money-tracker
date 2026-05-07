import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('./client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}))

import { api } from './client'
import { accountsApi } from './accounts'
import { adminApi } from './admin'
import { balanceApi } from './balance'
import {
  fetchBudgets,
  createBudget,
  updateBudget,
  deleteBudget,
  fetchBudgetTransactions,
} from './budgets'
import { categoriesApi } from './categories'
import {
  fetchGoals,
  createGoal,
  updateGoal,
  depositGoal,
  withdrawGoal,
  deleteGoal,
  fetchGoalHistory,
} from './goals'
import {
  fetchRecurring,
  createRecurring,
  updateRecurring,
  toggleRecurring,
  deleteRecurring,
} from './recurring'
import { settingsApi } from './settings'
import { statsApi } from './stats'
import { transfersApi, exchangeApi } from './transfers'

const apiMock = api as unknown as Record<'get' | 'post' | 'put' | 'patch' | 'delete', ReturnType<typeof vi.fn>>

beforeEach(() => {
  Object.values(apiMock).forEach((fn) => {
    fn.mockReset().mockResolvedValue({})
  })
})

describe('accountsApi', () => {
  it('list unwraps the accounts envelope', async () => {
    apiMock.get.mockResolvedValueOnce({ accounts: [{ id: 1, name: 'Main' }] })
    const got = await accountsApi.list()
    expect(apiMock.get).toHaveBeenCalledWith('/v1/accounts')
    expect(got).toEqual([{ id: 1, name: 'Main' }])
  })

  it('getById hits /v1/accounts/:id', async () => {
    await accountsApi.getById(7)
    expect(apiMock.get).toHaveBeenCalledWith('/v1/accounts/7')
  })

  it('create POSTs to /v1/accounts with body', async () => {
    const data = { name: 'X', icon: 'wallet', color: '#000', type: 'cash', currency_code: 'USD', include_in_total: true }
    await accountsApi.create(data)
    expect(apiMock.post).toHaveBeenCalledWith('/v1/accounts', data)
  })

  it('update PUTs to /v1/accounts/:id', async () => {
    await accountsApi.update(3, { name: 'Y' })
    expect(apiMock.put).toHaveBeenCalledWith('/v1/accounts/3', { name: 'Y' })
  })

  it('setDefault POSTs to set-default subroute', async () => {
    await accountsApi.setDefault(5)
    expect(apiMock.post).toHaveBeenCalledWith('/v1/accounts/5/set-default', {})
  })

  it('delete DELETEs /v1/accounts/:id', async () => {
    await accountsApi.delete(8)
    expect(apiMock.delete).toHaveBeenCalledWith('/v1/accounts/8')
  })

  it('adjust POSTs to /v1/accounts/:id/adjust with delta', async () => {
    await accountsApi.adjust(1, { delta_cents: 500, note: 'fix' })
    expect(apiMock.post).toHaveBeenCalledWith('/v1/accounts/1/adjust', { delta_cents: 500, note: 'fix' })
  })
})

describe('adminApi', () => {
  it('getUsers paginates with default values', async () => {
    await adminApi.getUsers()
    expect(apiMock.get).toHaveBeenCalledWith('/v1/admin/users?page=1&page_size=20')
  })

  it('getUsers honours custom page and pageSize', async () => {
    await adminApi.getUsers(2, 50)
    expect(apiMock.get).toHaveBeenCalledWith('/v1/admin/users?page=2&page_size=50')
  })

  it('getStats hits /v1/admin/stats', async () => {
    await adminApi.getStats()
    expect(apiMock.get).toHaveBeenCalledWith('/v1/admin/stats')
  })

  it('resetUser DELETEs /v1/admin/users/:id/data', async () => {
    await adminApi.resetUser(99)
    expect(apiMock.delete).toHaveBeenCalledWith('/v1/admin/users/99/data')
  })

  it('resetAllUsers DELETEs the bulk endpoint', async () => {
    await adminApi.resetAllUsers()
    expect(apiMock.delete).toHaveBeenCalledWith('/v1/admin/users/data')
  })
})

describe('balanceApi', () => {
  it('omits account_id when not provided', async () => {
    await balanceApi.get()
    expect(apiMock.get).toHaveBeenCalledWith('/v1/balance')
  })

  it('appends account_id query param', async () => {
    await balanceApi.get(7)
    expect(apiMock.get).toHaveBeenCalledWith('/v1/balance?account_id=7')
  })

  it('omits account_id when null is passed', async () => {
    await balanceApi.get(null)
    expect(apiMock.get).toHaveBeenCalledWith('/v1/balance')
  })
})

describe('budgets api', () => {
  it('fetchBudgets', async () => {
    await fetchBudgets()
    expect(apiMock.get).toHaveBeenCalledWith('/v1/budgets')
  })

  it('createBudget', async () => {
    await createBudget({ category_id: 1, limit_cents: 1000, period: 'monthly', currency_code: 'USD' })
    expect(apiMock.post).toHaveBeenCalledWith('/v1/budgets', expect.objectContaining({ category_id: 1 }))
  })

  it('updateBudget PUTs /v1/budgets/:id', async () => {
    await updateBudget(2, { limit_cents: 2000 })
    expect(apiMock.put).toHaveBeenCalledWith('/v1/budgets/2', { limit_cents: 2000 })
  })

  it('deleteBudget DELETEs /v1/budgets/:id', async () => {
    await deleteBudget(3)
    expect(apiMock.delete).toHaveBeenCalledWith('/v1/budgets/3')
  })

  it('fetchBudgetTransactions', async () => {
    await fetchBudgetTransactions(4)
    expect(apiMock.get).toHaveBeenCalledWith('/v1/budgets/4/transactions')
  })
})

describe('categoriesApi', () => {
  it('list with no filters', async () => {
    await categoriesApi.list()
    expect(apiMock.get).toHaveBeenCalledWith('/v1/categories')
  })

  it('list with type filter', async () => {
    await categoriesApi.list('expense')
    expect(apiMock.get).toHaveBeenCalledWith('/v1/categories?type=expense')
  })

  it('list with type and order', async () => {
    await categoriesApi.list('income', 'frequency')
    expect(apiMock.get).toHaveBeenCalledWith('/v1/categories?type=income&order=frequency')
  })

  it('create / update / delete', async () => {
    await categoriesApi.create({ name: 'C', icon: 'i', type: 'expense', color: '#fff' })
    expect(apiMock.post).toHaveBeenCalledWith('/v1/categories', expect.objectContaining({ name: 'C' }))

    await categoriesApi.update(5, { name: 'D', icon: 'i', type: 'expense', color: '#fff' })
    expect(apiMock.put).toHaveBeenCalledWith('/v1/categories/5', expect.objectContaining({ name: 'D' }))

    await categoriesApi.delete(7)
    expect(apiMock.delete).toHaveBeenCalledWith('/v1/categories/7')
  })
})

describe('goals api', () => {
  it('fetchGoals', async () => {
    await fetchGoals()
    expect(apiMock.get).toHaveBeenCalledWith('/v1/goals')
  })

  it('createGoal', async () => {
    await createGoal({ name: 'Vacation', target_cents: 100, currency_code: 'USD' })
    expect(apiMock.post).toHaveBeenCalledWith('/v1/goals', expect.objectContaining({ name: 'Vacation' }))
  })

  it('updateGoal PUTs', async () => {
    await updateGoal(1, { name: 'New' })
    expect(apiMock.put).toHaveBeenCalledWith('/v1/goals/1', { name: 'New' })
  })

  it('depositGoal POSTs to deposit subroute', async () => {
    await depositGoal(2, 500)
    expect(apiMock.post).toHaveBeenCalledWith('/v1/goals/2/deposit', { amount_cents: 500 })
  })

  it('withdrawGoal POSTs to withdraw subroute', async () => {
    await withdrawGoal(3, 200)
    expect(apiMock.post).toHaveBeenCalledWith('/v1/goals/3/withdraw', { amount_cents: 200 })
  })

  it('deleteGoal', async () => {
    await deleteGoal(4)
    expect(apiMock.delete).toHaveBeenCalledWith('/v1/goals/4')
  })

  it('fetchGoalHistory', async () => {
    await fetchGoalHistory(5)
    expect(apiMock.get).toHaveBeenCalledWith('/v1/goals/5/history')
  })
})

describe('recurring api', () => {
  it('fetchRecurring', async () => {
    await fetchRecurring()
    expect(apiMock.get).toHaveBeenCalledWith('/v1/recurring')
  })

  it('createRecurring', async () => {
    await createRecurring({
      account_id: 1, type: 'expense', amount_cents: 100, currency_code: 'USD',
      category_id: 2, note: 'rent', frequency: 'monthly',
    })
    expect(apiMock.post).toHaveBeenCalledWith('/v1/recurring', expect.objectContaining({ frequency: 'monthly' }))
  })

  it('updateRecurring PUTs', async () => {
    await updateRecurring(3, { amount_cents: 200 })
    expect(apiMock.put).toHaveBeenCalledWith('/v1/recurring/3', { amount_cents: 200 })
  })

  it('toggleRecurring PATCHes /toggle subroute', async () => {
    await toggleRecurring(4)
    expect(apiMock.patch).toHaveBeenCalledWith('/v1/recurring/4/toggle', {})
  })

  it('deleteRecurring', async () => {
    await deleteRecurring(5)
    expect(apiMock.delete).toHaveBeenCalledWith('/v1/recurring/5')
  })
})

describe('settingsApi', () => {
  it('get hits /v1/settings', async () => {
    await settingsApi.get()
    expect(apiMock.get).toHaveBeenCalledWith('/v1/settings')
  })

  it('update PATCHes /v1/settings', async () => {
    const patch = { language: 'ru', display_currencies: ['USD'] }
    await settingsApi.update(patch)
    expect(apiMock.patch).toHaveBeenCalledWith('/v1/settings', patch)
  })

  it('resetData DELETEs /v1/user/data', async () => {
    await settingsApi.resetData()
    expect(apiMock.delete).toHaveBeenCalledWith('/v1/user/data')
  })
})

describe('statsApi', () => {
  it('get with default period=month', async () => {
    await statsApi.get()
    expect(apiMock.get).toHaveBeenCalledWith('/v1/stats?period=month')
  })

  it('get with custom period', async () => {
    await statsApi.get('week')
    expect(apiMock.get).toHaveBeenCalledWith('/v1/stats?period=week')
  })

  it('get with accountId appends query param', async () => {
    await statsApi.get('month', 7)
    const url = apiMock.get.mock.calls[0][0] as string
    expect(url).toContain('period=month')
    expect(url).toContain('account_id=7')
  })

  it('getRange uses from/to', async () => {
    await statsApi.getRange('2026-01-01', '2026-01-31')
    const url = apiMock.get.mock.calls[0][0] as string
    expect(url).toContain('from=2026-01-01')
    expect(url).toContain('to=2026-01-31')
  })
})

describe('transfers api', () => {
  it('list with no params', async () => {
    await transfersApi.list()
    expect(apiMock.get).toHaveBeenCalledWith('/v1/transfers')
  })

  it('list with limit and offset', async () => {
    await transfersApi.list({ limit: 10, offset: 20 })
    expect(apiMock.get).toHaveBeenCalledWith('/v1/transfers?limit=10&offset=20')
  })

  it('create POSTs body', async () => {
    await transfersApi.create({ from_account_id: 1, to_account_id: 2, amount_cents: 100 })
    expect(apiMock.post).toHaveBeenCalledWith('/v1/transfers', expect.objectContaining({ from_account_id: 1 }))
  })

  it('delete DELETEs /v1/transfers/:id', async () => {
    await transfersApi.delete(7)
    expect(apiMock.delete).toHaveBeenCalledWith('/v1/transfers/7')
  })
})

describe('exchangeApi', () => {
  it('getRate hits /v1/exchange/rate with from and to', async () => {
    await exchangeApi.getRate('USD', 'EUR')
    expect(apiMock.get).toHaveBeenCalledWith('/v1/exchange/rate?from=USD&to=EUR')
  })
})
