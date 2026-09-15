import { Suspense, useCallback, useEffect, useMemo, useRef, useState, type CSSProperties } from 'react'
import type { GooseAdminApi, PagePayload } from '@gooseforum/client'
import { Toaster } from '@gooseforum/ui/components/sonner'
import { SidebarInset, SidebarProvider } from '@gooseforum/ui/components/sidebar'
import { TooltipProvider } from '@gooseforum/ui/components/tooltip'
import { applyBrowserLocale, applyBrowserTheme, detectBrowserLocale, detectBrowserTheme } from '../host/browser-runtime'
import { AppSidebar } from './components/app-sidebar'
import { SiteHeader } from './components/site-header'
import { createAdminText } from './i18n'
import { createAccessGroupText } from './access-groups-i18n'
import { createRoleText } from './roles-i18n'
import { createUserText } from './users-i18n'
import { createPostText } from './posts-i18n'
import { createPageConfigText } from './page-config-i18n'
import { createAssetText } from './assets-i18n'
import { createAuditText } from './audit-i18n'
import { createSettingsText } from './settings-i18n'
import { createSystemSettingsText } from './system-settings-i18n'
import { createContentSettingsText } from './content-settings-i18n'
import { createModerationSettingsText } from './moderation-settings-i18n'
import { createIdentityText } from './identity-settings-i18n'
import { createDashboardText } from './dashboard-i18n'
import { adminNavGroups } from './nav'
import { normalizeAdminPath } from './navigation'
import { prepareAdminTranslations, adminPageNamespaces, loadedAdminNamespaces } from './translation-loader'

import { DashboardPage, CategoriesManagementPage, AccessGroupsManagementPage, RolesManagementPage, UsersManagementPage, PostsManagementPage, LinksManagementPage, SponsorsManagementPage, BadgesManagementPage, FileResourcesManagementPage, OptRecordsManagementPage, SiteInfoManagementPage, SiteChromeManagementPage, MailSettingsPage, SecuritySettingsPage, PostingSettingsPage, AnnouncementSettingsPage, HttpNotifySettingsPage, SensitiveWordsSettingsPage, OAuthSettingsPage, OIDCProviderSettingsPage, prepareAdminPage } from './page-registry'

const shellVariables = {
  '--sidebar-width': 'calc(var(--spacing) * 72)',
  '--header-height': 'calc(var(--spacing) * 12)',
} as CSSProperties

export function AdminApp({ page, api }: { page: PagePayload; api: GooseAdminApi }) {
  const [pathname, setPathname] = useState(() => normalizeAdminPath(window.location.pathname))
  const requestVersion = useRef(0)
  const [isNavigating, setIsNavigating] = useState(false)
  const [navigationError, setNavigationError] = useState('')
  const [locale, setLocale] = useState(() => detectBrowserLocale())
  const localeRef = useRef(locale)
  const localeVersion = useRef(0)
  const pathnameRef = useRef(pathname)
  const [theme, setTheme] = useState(() => detectBrowserTheme())
  const text = useMemo(() => createAdminText(locale), [locale])
  const accessGroupText = useMemo(() => createAccessGroupText(locale), [locale])
  const roleText = useMemo(() => createRoleText(locale), [locale])
  const userText = useMemo(() => createUserText(locale), [locale])
  const postText = useMemo(() => createPostText(locale), [locale])
  const pageConfigText = useMemo(() => createPageConfigText(locale), [locale])
  const assetText = useMemo(() => createAssetText(locale), [locale])
  const auditText = useMemo(() => createAuditText(locale), [locale])
  const settingsText = useMemo(() => createSettingsText(locale), [locale])
  const systemSettingsText = useMemo(() => createSystemSettingsText(locale), [locale])
  const contentSettingsText = useMemo(() => createContentSettingsText(locale), [locale])
  const moderationSettingsText = useMemo(() => createModerationSettingsText(locale), [locale])
  const identityText = useMemo(() => createIdentityText(locale), [locale])
  const dashboardText = useMemo(() => createDashboardText(locale), [locale])
  const titleKey = adminNavGroups.flatMap((group) => group.items).find((item) => item.url === pathname)?.label || 'console'

  useEffect(() => {
    document.documentElement.lang = locale
  }, [locale])

  useEffect(() => {
    document.title = `${text(titleKey)} - ${page.layout.site.name || 'GooseForum'}`
  }, [page.layout.site.name, text, titleKey])

  const navigate = useCallback(async (path: string, updateHistory = true) => {
    const version = ++requestVersion.current
    const nextPath = normalizeAdminPath(path)
    setNavigationError('')
    if (nextPath === pathname) { setIsNavigating(false); return }
    setIsNavigating(true)
    try {
      const language = localeRef.current
      await Promise.all([prepareAdminPage(nextPath), prepareAdminTranslations(language, adminPageNamespaces(nextPath))])
      if (language !== localeRef.current) await prepareAdminTranslations(localeRef.current, adminPageNamespaces(nextPath))
      if (version !== requestVersion.current) return
      if (updateHistory) window.history.pushState(null, '', nextPath)
      pathnameRef.current = nextPath
      setPathname(nextPath)
    } catch (error) {
      if (version === requestVersion.current) setNavigationError(error instanceof Error ? error.message : String(error))
    } finally {
      if (version === requestVersion.current) setIsNavigating(false)
    }
  }, [pathname])

  useEffect(() => {
    const handlePopState = () => { void navigate(window.location.pathname, false) }
    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  }, [navigate])

  useEffect(() => () => { requestVersion.current++; localeVersion.current++ }, [])

  const prefetch = useCallback((path: string) => {
    void Promise.all([prepareAdminPage(path), prepareAdminTranslations(localeRef.current, adminPageNamespaces(path))]).catch(() => undefined)
  }, [])

  const changeLocale = useCallback(async (nextLocale: typeof locale) => {
    const version = ++localeVersion.current
    setNavigationError('')
    try {
      await prepareAdminTranslations(nextLocale, loadedAdminNamespaces())
      await prepareAdminTranslations(nextLocale, adminPageNamespaces(pathnameRef.current))
      if (version !== localeVersion.current) return
      localeRef.current = nextLocale
      applyBrowserLocale(nextLocale)
      setLocale(nextLocale)
    } catch (error) {
      if (version === localeVersion.current) setNavigationError(error instanceof Error ? error.message : String(error))
    }
  }, [])

  const toggleTheme = useCallback(() => {
    setTheme((current) => {
      const next = current === 'gf-dark' ? 'gf-light' : 'gf-dark'
      applyBrowserTheme(next, page.layout.theme.colors)
      return next
    })
  }, [page.layout.theme.colors])

  const content = pathname === '/admin/categories'
    ? <CategoriesManagementPage api={api} permissions={page.layout.viewer.adminPermissions} text={text} />
    : pathname === '/admin/access-groups'
      ? <AccessGroupsManagementPage api={api} text={accessGroupText} onNavigate={navigate} />
      : pathname === '/admin/roles'
        ? <RolesManagementPage api={api} text={roleText} />
      : pathname === '/admin/users'
        ? <UsersManagementPage api={api} text={userText} />
      : pathname === '/admin/posts'
        ? <PostsManagementPage api={api} text={postText} />
      : pathname === '/admin/links'
        ? <LinksManagementPage api={api} text={pageConfigText} />
      : pathname === '/admin/sponsors'
        ? <SponsorsManagementPage api={api} text={pageConfigText} />
      : pathname === '/admin/badges'
        ? <BadgesManagementPage api={api} text={assetText} />
      : pathname === '/admin/files/resources'
        ? <FileResourcesManagementPage api={api} text={assetText} />
      : pathname === '/admin/opt-records'
        ? <OptRecordsManagementPage api={api} text={auditText} locale={locale} />
      : pathname === '/admin/settings/site-info'
        ? <SiteInfoManagementPage api={api} text={settingsText} />
      : pathname === '/admin/settings/site-chrome'
        ? <SiteChromeManagementPage api={api} text={settingsText} layout={page.layout} />
      : pathname === '/admin/settings/mail'
        ? <MailSettingsPage api={api} text={systemSettingsText} />
      : pathname === '/admin/settings/security'
        ? <SecuritySettingsPage api={api} text={systemSettingsText} />
      : pathname === '/admin/settings/posting'
        ? <PostingSettingsPage api={api} text={contentSettingsText} />
      : pathname === '/admin/settings/announcement'
        ? <AnnouncementSettingsPage api={api} text={contentSettingsText} />
      : pathname === '/admin/settings/http-notify'
        ? <HttpNotifySettingsPage api={api} text={moderationSettingsText} />
      : pathname === '/admin/settings/sensitive-words'
        ? <SensitiveWordsSettingsPage api={api} text={moderationSettingsText} />
      : pathname === '/admin/settings/oauth'
        ? <OAuthSettingsPage api={api} text={identityText} />
      : pathname === '/admin/settings/oidc-provider'
        ? <OIDCProviderSettingsPage api={api} text={identityText} />
      : pathname === '/admin'
                        ? <DashboardPage api={api} text={dashboardText} locale={locale} />
                        : <div className="flex flex-1 items-center justify-center p-6"><div className="max-w-md rounded-xl border bg-card p-6 text-center"><h2 className="text-lg font-semibold">{text(titleKey)}</h2><p className="mt-2 text-sm text-muted-foreground">{text('notAvailable')}</p></div></div>

  return <TooltipProvider><SidebarProvider style={shellVariables}><AppSidebar layout={page.layout} pathname={pathname} text={text} onNavigate={navigate} onPrefetch={prefetch} variant="inset" /><SidebarInset className="overflow-clip" aria-busy={isNavigating}><SiteHeader title={text(titleKey)} locale={locale} theme={theme} onLocaleChange={changeLocale} onThemeToggle={toggleTheme} /><div className="flex flex-1 flex-col">{navigationError ? <p role="alert" className="px-4 py-2 text-sm text-destructive">{navigationError}</p> : null}<Suspense fallback={<div className="flex min-h-48 items-center justify-center text-sm text-muted-foreground">{text('loading')}</div>}>{content}</Suspense></div></SidebarInset></SidebarProvider><Toaster position="bottom-right" /></TooltipProvider>
}
