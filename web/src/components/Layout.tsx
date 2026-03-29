import { Outlet } from 'react-router-dom'
import { TabBar } from './TabBar'
import { useTelegramApp } from '../hooks/useTelegramApp'

export function Layout() {
  useTelegramApp()

  return (
    <div className="flex flex-col bg-bg" style={{ height: '100svh', paddingTop: 'var(--safe-top)' }}>
      <main className="flex-1 overflow-y-auto min-h-0" style={{ paddingBottom: 'var(--tab-bar-h)', overscrollBehavior: 'none' }}>
        <Outlet />
      </main>
      <TabBar />
    </div>
  )
}
