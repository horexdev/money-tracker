import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route, useLocation } from 'react-router-dom'
import { AnimatePresence } from 'framer-motion'
import { useQuery } from '@tanstack/react-query'
import { Layout } from './components/Layout'
import { Spinner } from './components/Spinner'
import { settingsApi } from './api/settings'
import { useFirstLaunchSetup } from './hooks/useFirstLaunchSetup'

const DashboardPage      = lazy(() => import('./pages/DashboardPage').then(m => ({ default: m.DashboardPage })))
const AddTransactionPage = lazy(() => import('./pages/AddTransactionPage').then(m => ({ default: m.AddTransactionPage })))
const HistoryPage        = lazy(() => import('./pages/HistoryPage').then(m => ({ default: m.HistoryPage })))
const StatsPage          = lazy(() => import('./pages/StatsPage').then(m => ({ default: m.StatsPage })))
const MorePage           = lazy(() => import('./pages/MorePage').then(m => ({ default: m.MorePage })))
const SettingsPage       = lazy(() => import('./pages/SettingsPage').then(m => ({ default: m.SettingsPage })))
const CategoriesPage     = lazy(() => import('./pages/CategoriesPage').then(m => ({ default: m.CategoriesPage })))
const BudgetsPage        = lazy(() => import('./pages/BudgetsPage').then(m => ({ default: m.BudgetsPage })))
const RecurringPage      = lazy(() => import('./pages/RecurringPage').then(m => ({ default: m.RecurringPage })))
const SavingsPage        = lazy(() => import('./pages/SavingsPage').then(m => ({ default: m.SavingsPage })))
const ExportPage         = lazy(() => import('./pages/ExportPage').then(m => ({ default: m.ExportPage })))
const AccountsPage       = lazy(() => import('./pages/AccountsPage').then(m => ({ default: m.AccountsPage })))
const TransfersPage      = lazy(() => import('./pages/TransfersPage').then(m => ({ default: m.TransfersPage })))
const AdminPage          = lazy(() => import('./pages/AdminPage').then(m => ({ default: m.AdminPage })))

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
            <Route path="savings" element={<SavingsPage />} />
            <Route path="export" element={<ExportPage />} />
            <Route path="accounts" element={<AccountsPage />} />
            <Route path="transfers" element={<TransfersPage />} />
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
    <BrowserRouter>
      <AppInit />
    </BrowserRouter>
  )
}
