import { lazy, Suspense, useCallback, useEffect, useMemo, useState, type CSSProperties } from 'react'
import type { GooseAdminApi, PagePayload } from '@gooseforum/client'
import { Toaster } from '@gooseforum/react/components/ui/sonner'
import { SidebarInset, SidebarProvider } from '@gooseforum/react/components/ui/sidebar'
import { TooltipProvider } from '@gooseforum/react/components/ui/tooltip'
import { detectBrowserLocale } from '../browser-runtime'
import { AppSidebar } from './components/app-sidebar'
import { SiteHeader } from './components/site-header'
import { createAdminText } from './i18n'
import { createAccessGroupText } from './access-groups-i18n'
import { adminNavGroups } from './nav'
import { normalizeAdminPath } from './navigation'

const loadDashboardPage = () => import('./pages/dashboard-page')
const loadCategoriesPage = () => import('./pages/categories-management-page').then((module) => ({ default: module.CategoriesManagementPage }))
const loadAccessGroupsPage = () => import('./pages/access-groups-management-page').then((module) => ({ default: module.AccessGroupsManagementPage }))
const DashboardPage = lazy(loadDashboardPage)
const CategoriesManagementPage = lazy(loadCategoriesPage)
const AccessGroupsManagementPage = lazy(loadAccessGroupsPage)

const shellVariables = {
  '--sidebar-width': 'calc(var(--spacing) * 72)',
  '--header-height': 'calc(var(--spacing) * 12)',
} as CSSProperties

export function AdminApp({ page, api }: { page: PagePayload; api: GooseAdminApi }) {
  const [pathname, setPathname] = useState(() => normalizeAdminPath(window.location.pathname))
  const locale = useMemo(() => detectBrowserLocale(), [])
  const text = useMemo(() => createAdminText(locale), [locale])
  const accessGroupText = useMemo(() => createAccessGroupText(locale), [locale])
  const titleKey = adminNavGroups.flatMap((group) => group.items).find((item) => item.url === pathname)?.label || 'console'

  useEffect(() => {
    document.documentElement.lang = locale
  }, [locale])

  useEffect(() => {
    const handlePopState = () => setPathname(normalizeAdminPath(window.location.pathname))
    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  }, [])

  useEffect(() => {
    document.title = `${text(titleKey)} - ${page.layout.site.name || 'GooseForum'}`
  }, [page.layout.site.name, text, titleKey])

  const navigate = useCallback((path: string) => {
    const nextPath = normalizeAdminPath(path)
    if (nextPath === pathname) return
    window.history.pushState(null, '', nextPath)
    setPathname(nextPath)
  }, [pathname])

  const prefetch = useCallback((path: string) => {
    if (path === '/admin') void loadDashboardPage()
    if (path === '/admin/categories') void loadCategoriesPage()
    if (path === '/admin/access-groups') void loadAccessGroupsPage()
  }, [])

  const content = pathname === '/admin/categories'
    ? <CategoriesManagementPage api={api} permissions={page.layout.viewer.adminPermissions} text={text} />
    : pathname === '/admin/access-groups'
      ? <AccessGroupsManagementPage api={api} text={accessGroupText} onNavigate={navigate} />
      : pathname === '/admin'
        ? <DashboardPage />
        : <div className="flex flex-1 items-center justify-center p-6"><div className="max-w-md rounded-xl border bg-card p-6 text-center"><h2 className="text-lg font-semibold">{text(titleKey)}</h2><p className="mt-2 text-sm text-muted-foreground">{text('notAvailable')}</p></div></div>

  return <TooltipProvider><SidebarProvider style={shellVariables}><AppSidebar layout={page.layout} pathname={pathname} text={text} onNavigate={navigate} onPrefetch={prefetch} variant="inset" /><SidebarInset><SiteHeader title={text(titleKey)} /><div className="flex flex-1 flex-col"><Suspense fallback={<div className="flex min-h-48 items-center justify-center text-sm text-muted-foreground">{text('loading')}</div>}>{content}</Suspense></div></SidebarInset></SidebarProvider><Toaster position="bottom-right" /></TooltipProvider>
}
