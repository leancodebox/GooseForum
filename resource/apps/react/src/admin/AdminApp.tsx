import type { CSSProperties } from 'react'
import { Toaster } from '@gooseforum/react/components/ui/sonner'
import {
  SidebarInset,
  SidebarProvider,
} from '@gooseforum/react/components/ui/sidebar'
import { TooltipProvider } from '@gooseforum/react/components/ui/tooltip'
import { AppSidebar } from './components/app-sidebar'
import { ChartAreaInteractive } from './components/chart-area-interactive'
import { DataTable } from './components/data-table'
import { SectionCards } from './components/section-cards'
import { SiteHeader } from './components/site-header'
import dashboardData from './data/dashboard.json'

const shellVariables = {
  '--sidebar-width': 'calc(var(--spacing) * 72)',
  '--header-height': 'calc(var(--spacing) * 12)',
} as CSSProperties

export function AdminApp() {
  return (
    <TooltipProvider>
      <SidebarProvider style={shellVariables}>
        <AppSidebar variant="inset" />
        <SidebarInset>
          <SiteHeader />
          <div className="flex flex-1 flex-col">
            <div className="@container/main flex flex-1 flex-col gap-2">
              <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
                <SectionCards />
                <div className="px-4 lg:px-6">
                  <ChartAreaInteractive />
                </div>
                <DataTable data={dashboardData} />
              </div>
            </div>
          </div>
        </SidebarInset>
      </SidebarProvider>
      <Toaster position="bottom-right" />
    </TooltipProvider>
  )
}
