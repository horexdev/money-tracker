import {
  mockBalance,
  mockTransactions,
  mockCategories,
  mockStats,
  mockBudgets,
  mockRecurring,
  mockGoals,
  mockSettings,
  mockAccounts,
  mockTransfers,
  mockAdminStats,
  mockAdminUsers,
} from './data'

type RouteHandler = (url: URL) => unknown

const routes: Array<{ pattern: RegExp; handler: RouteHandler }> = [
  {
    pattern: /\/api\/v1\/balance$/,
    handler: (url) => {
      const accountId = url.searchParams.get('account_id')
      if (accountId) {
        const acc = mockAccounts.accounts.find(a => a.id === Number(accountId))
        if (acc) {
          return {
            total_in_base_cents: acc.balance_cents,
            by_currency: [{
              currency_code: acc.currency_code,
              net_cents: acc.balance_cents,
              income_cents: Math.round(acc.balance_cents * 0.6),
              expense_cents: Math.round(acc.balance_cents * 0.4),
            }],
          }
        }
      }
      return mockBalance
    },
  },
  {
    pattern: /\/api\/v1\/transactions$/,
    handler: (url) => {
      const page = Number(url.searchParams.get('page') || '1')
      const pageSize = Number(url.searchParams.get('page_size') || '20')
      const accountId = url.searchParams.get('account_id')
      const all = accountId
        ? mockTransactions.transactions.filter(tx => tx.account_id === Number(accountId))
        : mockTransactions.transactions
      const start = (page - 1) * pageSize
      const txs = all.slice(start, start + pageSize)
      return {
        transactions: txs,
        total_pages: Math.ceil(all.length / pageSize),
        current_page: page,
      }
    },
  },
  {
    pattern: /\/api\/v1\/transactions\/\d+$/,
    handler: () => undefined,
  },
  {
    pattern: /\/api\/v1\/categories$/,
    handler: () => mockCategories,
  },
  {
    pattern: /\/api\/v1\/categories\/\d+$/,
    handler: () => mockCategories.categories[0],
  },
  {
    pattern: /\/api\/v1\/stats/,
    handler: () => ({
      ...mockStats,
      items: mockStats.items.map((item) => ({
        ...item,
        currency_code: mockSettings.base_currency,
      })),
    }),
  },
  {
    pattern: /\/api\/v1\/budgets$/,
    handler: () => mockBudgets,
  },
  {
    pattern: /\/api\/v1\/budgets\/\d+$/,
    handler: () => mockBudgets.budgets[0],
  },
  {
    pattern: /\/api\/v1\/recurring$/,
    handler: () => mockRecurring,
  },
  {
    pattern: /\/api\/v1\/recurring\/\d+/,
    handler: () => mockRecurring.recurring[0],
  },
  {
    pattern: /\/api\/v1\/goals$/,
    handler: () => mockGoals,
  },
  {
    pattern: /\/api\/v1\/goals\/\d+/,
    handler: () => mockGoals.goals[0],
  },
  {
    pattern: /\/api\/v1\/accounts$/,
    handler: () => mockAccounts,
  },
  {
    pattern: /\/api\/v1\/accounts\/\d+$/,
    handler: (url) => {
      const id = Number(url.pathname.split('/').pop())
      return mockAccounts.accounts.find(a => a.id === id) ?? mockAccounts.accounts[0]
    },
  },
  {
    pattern: /\/api\/v1\/transfers$/,
    handler: () => mockTransfers,
  },
  {
    pattern: /\/api\/v1\/transfers\/\d+$/,
    handler: (url) => {
      const id = Number(url.pathname.split('/').pop())
      return mockTransfers.transfers.find(t => t.id === id) ?? mockTransfers.transfers[0]
    },
  },
  {
    pattern: /\/api\/v1\/settings$/,
    handler: () => mockSettings,
  },
  {
    pattern: /\/api\/v1\/admin\/stats$/,
    handler: () => mockAdminStats,
  },
  {
    pattern: /\/api\/v1\/admin\/users$/,
    handler: (url) => {
      const page = Number(url.searchParams.get('page') || '1')
      const pageSize = Number(url.searchParams.get('page_size') || '20')
      return { ...mockAdminUsers, page, page_size: pageSize }
    },
  },
  {
    pattern: /\/api\/v1\/export$/,
    handler: () => new Blob(['id,type,amount\n1,expense,45.00'], { type: 'text/csv' }),
  },
]

function matchRoute(url: URL): RouteHandler | undefined {
  for (const route of routes) {
    if (route.pattern.test(url.pathname)) {
      return route.handler
    }
  }
  return undefined
}

export function setupMockFetch() {
  const originalFetch = window.fetch

  window.fetch = async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
    const urlStr = typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url
    const url = new URL(urlStr, window.location.origin)

    // Handle PUT /api/v1/budgets/:id — mutate mockBudgets in-place
    if (/\/api\/v1\/budgets\/\d+$/.test(url.pathname) && init?.method === 'PUT') {
      await new Promise((r) => setTimeout(r, 150))
      const id = Number(url.pathname.split('/').pop())
      const body = JSON.parse(init.body as string)
      const budget = mockBudgets.budgets.find(b => b.id === id)
      if (budget) Object.assign(budget, body)
      return new Response(JSON.stringify(budget ?? mockBudgets.budgets[0]), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    }

    // Handle PUT /api/v1/recurring/:id — mutate mockRecurring in-place
    if (/\/api\/v1\/recurring\/\d+$/.test(url.pathname) && init?.method === 'PUT') {
      await new Promise((r) => setTimeout(r, 150))
      const id = Number(url.pathname.split('/').pop())
      const body = JSON.parse(init.body as string)
      const item = mockRecurring.recurring.find(r => r.id === id)
      if (item) Object.assign(item, body)
      return new Response(JSON.stringify(item ?? mockRecurring.recurring[0]), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    }

    // Handle PUT /api/v1/goals/:id — mutate mockGoals in-place
    if (/\/api\/v1\/goals\/\d+$/.test(url.pathname) && init?.method === 'PUT') {
      await new Promise((r) => setTimeout(r, 150))
      const id = Number(url.pathname.split('/').pop())
      const body = JSON.parse(init.body as string)
      const goal = mockGoals.goals.find(g => g.id === id)
      if (goal) Object.assign(goal, body)
      return new Response(JSON.stringify(goal ?? mockGoals.goals[0]), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    }

    // Handle DELETE /api/v1/user/data — reset all mock user data
    if (/\/api\/v1\/user\/data$/.test(url.pathname) && init?.method === 'DELETE') {
      await new Promise((r) => setTimeout(r, 200))
      mockTransactions.transactions = []
      mockBudgets.budgets = []
      mockRecurring.recurring = []
      mockGoals.goals = []
      mockBalance.by_currency.forEach((b) => {
        b.income_cents = 0
        b.expense_cents = 0
        b.net_cents = 0
      })
      mockBalance.total_in_base_cents = 0
      mockStats.items = []
      return new Response(null, { status: 204 })
    }

    // Handle POST /api/v1/accounts — create account
    if (/\/api\/v1\/accounts$/.test(url.pathname) && init?.method === 'POST') {
      await new Promise((r) => setTimeout(r, 150))
      const body = JSON.parse(init.body as string)
      const newAccount = {
        id: Date.now(),
        name: body.name ?? 'Account',
        icon: body.icon ?? 'wallet',
        color: body.color ?? '#6366f1',
        type: body.type ?? 'checking',
        currency_code: body.currency_code ?? 'USD',
        is_default: mockAccounts.accounts.length === 0,
        include_in_total: body.include_in_total ?? true,
        balance_cents: 0,
        created_at: new Date().toISOString(),
      }
      mockAccounts.accounts.push(newAccount)
      return new Response(JSON.stringify(newAccount), {
        status: 201,
        headers: { 'Content-Type': 'application/json' },
      })
    }

    // Handle PUT /api/v1/accounts/:id — update account
    if (/\/api\/v1\/accounts\/\d+$/.test(url.pathname) && init?.method === 'PUT') {
      await new Promise((r) => setTimeout(r, 150))
      const id = Number(url.pathname.split('/').pop())
      const body = JSON.parse(init.body as string)
      const account = mockAccounts.accounts.find(a => a.id === id)
      if (account) Object.assign(account, body)
      return new Response(JSON.stringify(account ?? mockAccounts.accounts[0]), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    }

    // Handle POST /api/v1/accounts/:id/set-default — set default account
    if (/\/api\/v1\/accounts\/\d+\/set-default$/.test(url.pathname) && init?.method === 'POST') {
      await new Promise((r) => setTimeout(r, 150))
      const id = Number(url.pathname.split('/').at(-2))
      mockAccounts.accounts.forEach(a => { a.is_default = a.id === id })
      const account = mockAccounts.accounts.find(a => a.id === id)
      return new Response(JSON.stringify(account ?? mockAccounts.accounts[0]), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    }

    // Handle DELETE /api/v1/accounts/:id — delete account
    if (/\/api\/v1\/accounts\/\d+$/.test(url.pathname) && init?.method === 'DELETE') {
      await new Promise((r) => setTimeout(r, 150))
      const id = Number(url.pathname.split('/').pop())
      mockAccounts.accounts = mockAccounts.accounts.filter(a => a.id !== id)
      return new Response(null, { status: 204 })
    }

    // Handle POST /api/v1/transfers — create transfer
    if (/\/api\/v1\/transfers$/.test(url.pathname) && init?.method === 'POST') {
      await new Promise((r) => setTimeout(r, 150))
      const body = JSON.parse(init.body as string)
      const fromAcc = mockAccounts.accounts.find(a => a.id === body.from_account_id)
      const toAcc = mockAccounts.accounts.find(a => a.id === body.to_account_id)
      const newTransfer = {
        id: Date.now(),
        from_account_id: body.from_account_id,
        from_account_name: fromAcc?.name ?? '',
        to_account_id: body.to_account_id,
        to_account_name: toAcc?.name ?? '',
        amount_cents: body.amount_cents,
        from_currency_code: fromAcc?.currency_code ?? 'USD',
        to_currency_code: toAcc?.currency_code ?? 'USD',
        exchange_rate: body.exchange_rate ?? 1,
        note: body.note ?? '',
        created_at: new Date().toISOString(),
      }
      mockTransfers.transfers.unshift(newTransfer)
      mockTransfers.total += 1
      return new Response(JSON.stringify(newTransfer), {
        status: 201,
        headers: { 'Content-Type': 'application/json' },
      })
    }

    // Handle DELETE /api/v1/transfers/:id — delete transfer
    if (/\/api\/v1\/transfers\/\d+$/.test(url.pathname) && init?.method === 'DELETE') {
      await new Promise((r) => setTimeout(r, 150))
      const id = Number(url.pathname.split('/').pop())
      mockTransfers.transfers = mockTransfers.transfers.filter(t => t.id !== id)
      mockTransfers.total = Math.max(0, mockTransfers.total - 1)
      return new Response(null, { status: 204 })
    }

    // Handle PATCH /api/v1/settings
    if (/\/api\/v1\/settings$/.test(url.pathname) && init?.method === 'PATCH') {
      await new Promise((r) => setTimeout(r, 150))
      const body = JSON.parse(init.body as string)
      Object.assign(mockSettings, body)
      return new Response(JSON.stringify(mockSettings), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    }

    const handler = matchRoute(url)
    if (handler) {
      // Simulate network delay
      await new Promise((r) => setTimeout(r, 200 + Math.random() * 300))

      const result = handler(url)

      if (result instanceof Blob) {
        return new Response(result, { status: 200 })
      }

      if (result === undefined) {
        return new Response(null, { status: 204 })
      }

      return new Response(JSON.stringify(result), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    }

    return originalFetch(input, init)
  }
}
