import { lazy, Suspense, useCallback, useEffect, useMemo, useState, type CSSProperties } from 'react'
import type { GooseAdminApi, PagePayload } from '@gooseforum/client'
import { Toaster } from '@gooseforum/react/components/ui/sonner'
import { SidebarInset, SidebarProvider } from '@gooseforum/react/components/ui/sidebar'
import { TooltipProvider } from '@gooseforum/react/components/ui/tooltip'
import { applyBrowserLocale, applyBrowserTheme, detectBrowserLocale, detectBrowserTheme } from '../browser-runtime'
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

const loadDashboardPage = () => import('./pages/dashboard-page')
const loadCategoriesPage = () => import('./pages/categories-management-page').then((module) => ({ default: module.CategoriesManagementPage }))
const loadAccessGroupsPage = () => import('./pages/access-groups-management-page').then((module) => ({ default: module.AccessGroupsManagementPage }))
const loadRolesPage = () => import('./pages/roles-management-page').then((module) => ({ default: module.RolesManagementPage }))
const loadUsersPage = () => import('./pages/users-management-page').then((module) => ({ default: module.UsersManagementPage }))
const loadPostsPage = () => import('./pages/posts-management-page').then((module) => ({ default: module.PostsManagementPage }))
const loadLinksPage = () => import('./pages/links-management-page').then((module) => ({ default: module.LinksManagementPage }))
const loadSponsorsPage = () => import('./pages/sponsors-management-page').then((module) => ({ default: module.SponsorsManagementPage }))
const loadBadgesPage = () => import('./pages/badges-management-page').then((module) => ({ default: module.BadgesManagementPage }))
const loadFilesPage = () => import('./pages/file-resources-management-page').then((module) => ({ default: module.FileResourcesManagementPage }))
const loadAuditPage = () => import('./pages/opt-records-management-page').then((module) => ({ default: module.OptRecordsManagementPage }))
const loadSiteInfoPage = () => import('./pages/site-info-management-page').then((module) => ({ default: module.SiteInfoManagementPage }))
const loadSiteChromePage = () => import('./pages/site-chrome-management-page').then((module) => ({ default: module.SiteChromeManagementPage }))
const loadMailPage = () => import('./pages/mail-settings-page').then((module) => ({ default: module.MailSettingsPage }))
const loadSecurityPage = () => import('./pages/security-settings-page').then((module) => ({ default: module.SecuritySettingsPage }))
const loadPostingPage = () => import('./pages/posting-settings-page').then((module) => ({ default: module.PostingSettingsPage }))
const loadAnnouncementPage = () => import('./pages/announcement-settings-page').then((module) => ({ default: module.AnnouncementSettingsPage }))
const loadHttpNotifyPage = () => import('./pages/http-notify-settings-page').then((module) => ({ default: module.HttpNotifySettingsPage }))
const loadSensitiveWordsPage = () => import('./pages/sensitive-words-settings-page').then((module) => ({ default: module.SensitiveWordsSettingsPage }))
const loadOAuthPage = () => import('./pages/oauth-settings-page').then((module) => ({ default: module.OAuthSettingsPage }))
const loadOIDCProviderPage = () => import('./pages/oidc-provider-settings-page').then((module) => ({ default: module.OIDCProviderSettingsPage }))
const DashboardPage = lazy(loadDashboardPage)
const CategoriesManagementPage = lazy(loadCategoriesPage)
const AccessGroupsManagementPage = lazy(loadAccessGroupsPage)
const RolesManagementPage = lazy(loadRolesPage)
const UsersManagementPage = lazy(loadUsersPage)
const PostsManagementPage = lazy(loadPostsPage)
const LinksManagementPage = lazy(loadLinksPage)
const SponsorsManagementPage = lazy(loadSponsorsPage)
const BadgesManagementPage = lazy(loadBadgesPage)
const FileResourcesManagementPage = lazy(loadFilesPage)
const OptRecordsManagementPage = lazy(loadAuditPage)
const SiteInfoManagementPage = lazy(loadSiteInfoPage)
const SiteChromeManagementPage = lazy(loadSiteChromePage)
const MailSettingsPage = lazy(loadMailPage)
const SecuritySettingsPage = lazy(loadSecurityPage)
const PostingSettingsPage = lazy(loadPostingPage)
const AnnouncementSettingsPage = lazy(loadAnnouncementPage)
const HttpNotifySettingsPage = lazy(loadHttpNotifyPage)
const SensitiveWordsSettingsPage = lazy(loadSensitiveWordsPage)
const OAuthSettingsPage = lazy(loadOAuthPage)
const OIDCProviderSettingsPage = lazy(loadOIDCProviderPage)

const shellVariables = {
  '--sidebar-width': 'calc(var(--spacing) * 72)',
  '--header-height': 'calc(var(--spacing) * 12)',
} as CSSProperties

export function AdminApp({ page, api }: { page: PagePayload; api: GooseAdminApi }) {
  const [pathname, setPathname] = useState(() => normalizeAdminPath(window.location.pathname))
  const [locale, setLocale] = useState(() => detectBrowserLocale())
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
    if (path === '/admin/roles') void loadRolesPage()
    if (path === '/admin/users') void loadUsersPage()
    if (path === '/admin/posts') void loadPostsPage()
    if (path === '/admin/links') void loadLinksPage()
    if (path === '/admin/sponsors') void loadSponsorsPage()
    if (path === '/admin/badges') void loadBadgesPage()
    if (path === '/admin/files/resources') void loadFilesPage()
    if (path === '/admin/opt-records') void loadAuditPage()
    if (path === '/admin/settings/site-info') void loadSiteInfoPage()
    if (path === '/admin/settings/site-chrome') void loadSiteChromePage()
    if (path === '/admin/settings/mail') void loadMailPage()
    if (path === '/admin/settings/security') void loadSecurityPage()
    if (path === '/admin/settings/posting') void loadPostingPage()
    if (path === '/admin/settings/announcement') void loadAnnouncementPage()
    if (path === '/admin/settings/http-notify') void loadHttpNotifyPage()
    if (path === '/admin/settings/sensitive-words') void loadSensitiveWordsPage()
    if (path === '/admin/settings/oauth') void loadOAuthPage()
    if (path === '/admin/settings/oidc-provider') void loadOIDCProviderPage()
  }, [])

  const changeLocale = useCallback((nextLocale: typeof locale) => {
    applyBrowserLocale(nextLocale)
    setLocale(nextLocale)
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

  return <TooltipProvider><SidebarProvider style={shellVariables}><AppSidebar layout={page.layout} pathname={pathname} text={text} onNavigate={navigate} onPrefetch={prefetch} variant="inset" /><SidebarInset className="overflow-clip"><SiteHeader title={text(titleKey)} locale={locale} theme={theme} onLocaleChange={changeLocale} onThemeToggle={toggleTheme} /><div className="flex flex-1 flex-col"><Suspense fallback={<div className="flex min-h-48 items-center justify-center text-sm text-muted-foreground">{text('loading')}</div>}>{content}</Suspense></div></SidebarInset></SidebarProvider><Toaster position="bottom-right" /></TooltipProvider>
}
