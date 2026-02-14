import { Outlet } from 'react-router-dom'
import { SidebarProvider } from '@/components/ui/sidebar'
import { AppSidebar } from './AppSidebar'
import { Header } from './Header'

export function Layout() {
  return (
    <SidebarProvider defaultOpen>
      <AppSidebar />
      <main className="flex min-h-screen w-full flex-col">
        <Header />
        <div className="flex-1 p-4 md:p-6 lg:p-8">
          <Outlet />
        </div>
      </main>
    </SidebarProvider>
  )
}
