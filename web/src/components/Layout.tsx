import { Outlet } from 'react-router-dom'
import { TabBar } from './TabBar'
import { useTelegramApp } from '../hooks/useTelegramApp'

export function Layout() {
  useTelegramApp()

  return (
    <div className="flex flex-col min-h-svh bg-bg" style={{ paddingTop: 'var(--safe-top)' }}>
      <main className="flex-1 overflow-y-auto">
        <Outlet />
        <div className="tab-spacer" />
      </main>
      <TabBar />
    </div>
  )
}
