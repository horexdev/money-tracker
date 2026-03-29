import {
  mockBalance,
  mockTransactions,
  mockCategories,
  mockStats,
  mockBudgets,
  mockRecurring,
  mockGoals,
  mockSettings,
} from './data'

// Approximate exchange rates relative to USD for mock purposes
const MOCK_RATES: Record<string, number> = {
  USD: 1.0,
  EUR: 0.93,
  GBP: 0.79,
  UAH: 41.5,
  RUB: 92.0,
  TRY: 32.0,
  KZT: 450.0,
  UZS: 12700.0,
  BRL: 5.0,
  JPY: 155.0,
}

function rateToUSD(code: string): number {
  return MOCK_RATES[code] ?? 1.0
}

// Rebuild mockBalance.by_currency with new base currency
function rebuildBalance(newBaseCurrency: string) {
  const usdToNew = rateToUSD(newBaseCurrency)

  // Convert each currency group to the new base
  mockBalance.by_currency.forEach((b) => {
    if (b.currency_code === newBaseCurrency) return
    // keep amounts as-is; only base currency entry gets relabelled
  })

  // Relabel the primary (previously base) entry to new currency
  const oldBase = mockBalance.by_currency.find(
    (b) => b.currency_code === mockSettings.base_currency
  )
  if (oldBase) {
    // Convert old base amounts to new base currency
    const factor = usdToNew / rateToUSD(mockSettings.base_currency)
    oldBase.income_cents = Math.round(oldBase.income_cents * factor)
    oldBase.expense_cents = Math.round(oldBase.expense_cents * factor)
    oldBase.net_cents = oldBase.income_cents - oldBase.expense_cents
    oldBase.currency_code = newBaseCurrency
  }

  // Recalculate total_in_base_cents
  mockBalance.total_in_base_cents = mockBalance.by_currency.reduce((sum, b) => {
    if (b.currency_code === newBaseCurrency) return sum + b.net_cents
    const rate = usdToNew / rateToUSD(b.currency_code)
    return sum + Math.round(b.net_cents * rate)
  }, 0)

  // Convert amounts on budgets and stats
  const factor2 = usdToNew / rateToUSD(mockSettings.base_currency)
  mockBudgets.budgets.forEach((b) => {
    if (b.currency_code === mockSettings.base_currency) {
      b.limit_cents = Math.round(b.limit_cents * factor2)
      b.spent_cents = Math.round(b.spent_cents * factor2)
      b.currency_code = newBaseCurrency
    }
  })
  mockStats.items.forEach((item) => {
    if (item.currency_code === mockSettings.base_currency) {
      item.total_cents = Math.round(item.total_cents * factor2)
      item.currency_code = newBaseCurrency
    }
  })
  mockGoals.goals.forEach((g) => {
    if (g.currency_code === mockSettings.base_currency) {
      g.target_cents = Math.round(g.target_cents * factor2)
      g.current_cents = Math.round(g.current_cents * factor2)
      g.remaining_cents = Math.round(g.remaining_cents * factor2)
      g.currency_code = newBaseCurrency
    }
  })
  mockRecurring.recurring.forEach((r) => {
    if (r.currency_code === mockSettings.base_currency) {
      r.amount_cents = Math.round(r.amount_cents * factor2)
      r.currency_code = newBaseCurrency
    }
  })
}

type RouteHandler = (url: URL) => unknown

const routes: Array<{ pattern: RegExp; handler: RouteHandler }> = [
  {
    pattern: /\/api\/v1\/balance$/,
    handler: () => mockBalance,
  },
  {
    pattern: /\/api\/v1\/transactions$/,
    handler: (url) => {
      const page = Number(url.searchParams.get('page') || '1')
      const pageSize = Number(url.searchParams.get('page_size') || '20')
      const start = (page - 1) * pageSize
      const txs = mockTransactions.transactions.slice(start, start + pageSize)
      return {
        transactions: txs,
        total_pages: Math.ceil(mockTransactions.transactions.length / pageSize),
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
    pattern: /\/api\/v1\/settings$/,
    handler: () => mockSettings,
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

    // Handle PATCH /api/v1/settings — mutate mockSettings in-place, update balance if currency changed
    if (/\/api\/v1\/settings$/.test(url.pathname) && init?.method === 'PATCH') {
      await new Promise((r) => setTimeout(r, 150))
      const body = JSON.parse(init.body as string)
      if (body.base_currency && body.base_currency !== mockSettings.base_currency) {
        rebuildBalance(body.base_currency)
      }
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
