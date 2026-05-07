import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route, useLocation } from 'react-router-dom'
import { AnimatePresence } from 'framer-motion'
import { useQuery } from '@tanstack/react-query'
import { Layout } from './Layout'
import { Spinner } from '../shared/ui/Spinner'
import { ErrorBoundary } from '../shared/ui/ErrorBoundary'
import { settingsApi } from '../shared/api/settings'
import { useFirstLaunchSetup } from '../shared/hooks/useFirstLaunchSetup'

const DashboardPage      = lazy(() => import('../features/dashboard').then(m => ({ default: m.DashboardPage })))
const AddTransactionPage = lazy(() => import('../features/add-transaction').then(m => ({ default: m.AddTransactionPage })))
const HistoryPage        = lazy(() => import('../features/history').then(m => ({ default: m.HistoryPage })))
const StatsPage          = lazy(() => import('../features/stats').then(m => ({ default: m.StatsPage })))
const MorePage           = lazy(() => import('../features/more').then(m => ({ default: m.MorePage })))
const SettingsPage       = lazy(() => import('../features/settings').then(m => ({ default: m.SettingsPage })))
const CategoriesPage     = lazy(() => import('../features/categories').then(m => ({ default: m.CategoriesPage })))
const BudgetsPage        = lazy(() => import('../features/budgets').then(m => ({ default: m.BudgetsPage })))
const RecurringPage      = lazy(() => import('../features/recurring').then(m => ({ default: m.RecurringPage })))
const TemplatesPage      = lazy(() => import('../features/templates').then(m => ({ default: m.TemplatesPage })))
const SavingsPage        = lazy(() => import('../features/savings').then(m => ({ default: m.SavingsPage })))
const ExportPage         = lazy(() => import('../features/export').then(m => ({ default: m.ExportPage })))
const AccountsPage       = lazy(() => import('../features/accounts').then(m => ({ default: m.AccountsPage })))
const AdminPage          = lazy(() => import('../features/admin').then(m => ({ default: m.AdminPage })))

function PageLoader() {
  return <div className="flex justify-center items-center h-48"><Spinner /></div>
}

function AnimatedRoutes() {
  const location = useLocation()

  return (
    <AnimatePresence mode="wait">
      <Suspense fallback={<PageLoader />} key={location.pathname}>
        <Routes location={location}>
          <Route element={<Layout />}>
            <Route index element={<DashboardPage />} />
            <Route path="add" element={<AddTransactionPage />} />
            <Route path="history" element={<HistoryPage />} />
            <Route path="stats" element={<StatsPage />} />
            <Route path="more" element={<MorePage />} />
            <Route path="settings" element={<SettingsPage />} />
            <Route path="categories" element={<CategoriesPage />} />
            <Route path="budgets" element={<BudgetsPage />} />
            <Route path="recurring" element={<RecurringPage />} />
            <Route path="templates" element={<TemplatesPage />} />
            <Route path="savings" element={<SavingsPage />} />
            <Route path="export" element={<ExportPage />} />
            <Route path="accounts" element={<AccountsPage />} />
            <Route path="admin" element={<AdminPage />} />
          </Route>
        </Routes>
      </Suspense>
    </AnimatePresence>
  )
}

function AppInit() {
  const { data: settings } = useQuery({
    queryKey: ['settings'],
    queryFn: settingsApi.get,
    staleTime: 60_000,
  })
  useFirstLaunchSetup(settings)
  return <AnimatedRoutes />
}

export default function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <AppInit />
      </BrowserRouter>
    </ErrorBoundary>
  )
}
