import { Suspense, useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState, type CSSProperties } from 'react'
import { flushSync } from 'react-dom'
import { applyBrowserTheme, createThemeTransition, detectBrowserTheme, detectBrowserThemePreference, resolveBrowserTheme, saveBrowserThemePreference, type BrowserThemePreference } from '@gooseforum/runtime/browser-host'
import type { GooseAdminApi, PagePayload } from '@gooseforum/client'
import { Toaster } from '@gooseforum/ui/components/sonner'
import { SidebarInset, SidebarProvider } from '@gooseforum/ui/components/sidebar'
import { TooltipProvider } from '@gooseforum/ui/components/tooltip'
import { applyBrowserLocale, detectBrowserLocale } from '../host/browser-runtime'
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
import { AuthLogsPage } from './page-registry'
import { AgentSettingsPage } from './page-registry'
import { TrustSettingsPage } from './page-registry'
import { normalizeAdminPath } from './navigation'
import { prepareAdminTranslations, adminPageNamespaces, loadedAdminNamespaces } from './translation-loader'
import { prepareAdminListData, type AdminListData } from './list-data'

import { DashboardPage, CategoriesManagementPage, AccessGroupsManagementPage, RolesManagementPage, UsersManagementPage, PostsManagementPage, LinksManagementPage, SponsorsManagementPage, BadgesManagementPage, FileResourcesManagementPage, OptRecordsManagementPage, SiteInfoManagementPage, SiteChromeManagementPage, MailSettingsPage, SecuritySettingsPage, PostingSettingsPage, AnnouncementSettingsPage, HttpNotifySettingsPage, SensitiveWordsSettingsPage, OAuthSettingsPage, OIDCProviderSettingsPage, prepareAdminPage } from './page-registry'

const shellVariables = {
  '--sidebar-width': 'calc(var(--spacing) * 72)',
  '--header-height': 'calc(var(--spacing) * 12)',
} as CSSProperties

const historyIndexKey = 'gooseAdminIndex'
function readHistoryIndex() {
  const state: unknown = window.history.state
  if (!state || typeof state !== 'object') return 0
  const index = (state as Record<string, unknown>)[historyIndexKey]
  return typeof index === 'number' && Number.isSafeInteger(index) && index >= 0 ? index : 0
}

export function AdminApp({ page, api, initialListData }: { page: PagePayload; api: GooseAdminApi; initialListData?: AdminListData }) {
  const [pathname, setPathname] = useState(() => normalizeAdminPath(window.location.pathname))
  const [routeKey, setRouteKey] = useState(() => `${readHistoryIndex()}:${window.location.pathname}${window.location.search}`)
  const [listData, setListData] = useState(initialListData)
  const rememberedUrls = useRef(new Map<string, string>())
  const scrollPositions = useRef(new Map<number, number>())
  const activeHistoryIndex = useRef(readHistoryIndex())
  const restoringHistory = useRef(false)
  const committedUrl = useRef(window.location.pathname + window.location.search)
  const pendingScroll = useRef<number | undefined>(undefined)
  const requestVersion = useRef(0)
  const [isNavigating, setIsNavigating] = useState(false)
  const [navigationError, setNavigationError] = useState('')
  const [locale, setLocale] = useState(() => detectBrowserLocale())
  const localeRef = useRef(locale)
  const localeVersion = useRef(0)
  const pathnameRef = useRef(pathname)
  const [theme, setTheme] = useState(() => detectBrowserTheme())
  const themeRef = useRef(theme)
  const [themePreference, setThemePreference] = useState(detectBrowserThemePreference)
  const themePreferenceRef = useRef(themePreference)
  const [themeTransition] = useState(createThemeTransition)
  useEffect(() => () => themeTransition.cancel(), [themeTransition])
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
  const rememberQueryUrl = useCallback((url: string) => { committedUrl.current = url }, [])

  useEffect(() => {
    document.documentElement.lang = locale
  }, [locale])

  useEffect(() => {
    document.title = `${text(titleKey)} - ${page.layout.site.name || 'GooseForum'}`
  }, [page.layout.site.name, text, titleKey])

  const navigate = useCallback(async (path: string, updateHistory = true, targetHistoryIndex = readHistoryIndex()) => {
    const version = ++requestVersion.current
    const requestedUrl = new URL(path, window.location.href)
    const nextPath = normalizeAdminPath(requestedUrl.pathname)
    // A popstate already changed the address; remember the page being left separately.
    const leavingUrl = committedUrl.current
    rememberedUrls.current.set(pathname, leavingUrl)
    scrollPositions.current.set(activeHistoryIndex.current, window.scrollY)
    const nextUrl = updateHistory && !requestedUrl.search
      ? new URL(rememberedUrls.current.get(nextPath) || requestedUrl.href, window.location.href)
      : requestedUrl
    nextUrl.pathname = nextPath
    const nextKey = nextUrl.pathname + nextUrl.search
    setNavigationError('')
    function restoreHistory() {
      const delta = activeHistoryIndex.current - readHistoryIndex()
      if (delta) {
        restoringHistory.current = true
        window.history.go(delta)
      } else if (window.location.pathname + window.location.search !== committedUrl.current) {
        window.history.replaceState(window.history.state, '', committedUrl.current)
      }
    }
    if (updateHistory && nextKey === leavingUrl) { restoreHistory(); setIsNavigating(false); return }
    setIsNavigating(true)
    try {
      const language = localeRef.current
      const [, , nextData] = await Promise.all([prepareAdminPage(nextPath), prepareAdminTranslations(language, adminPageNamespaces(nextPath)), prepareAdminListData(api, nextUrl)])
      if (language !== localeRef.current) await prepareAdminTranslations(localeRef.current, adminPageNamespaces(nextPath))
      if (version !== requestVersion.current) return
      const nextHistoryIndex = updateHistory ? readHistoryIndex() + 1 : targetHistoryIndex
      if (updateHistory) window.history.pushState({ ...window.history.state, [historyIndexKey]: nextHistoryIndex }, '', nextKey)
      committedUrl.current = nextKey
      pendingScroll.current = updateHistory ? 0 : scrollPositions.current.get(nextHistoryIndex) || 0
      activeHistoryIndex.current = nextHistoryIndex
      pathnameRef.current = nextPath
      setListData(nextData)
      setRouteKey(`${nextHistoryIndex}:${nextKey}`)
      setPathname(nextPath)
    } catch (error) {
      if (version === requestVersion.current) {
        restoreHistory()
        setNavigationError(error instanceof Error ? error.message : String(error))
      }
    } finally {
      if (version === requestVersion.current) setIsNavigating(false)
    }
  }, [api, pathname])

  useLayoutEffect(() => {
    if (pendingScroll.current === undefined) return
    window.scrollTo({ top: pendingScroll.current, behavior: 'instant' })
    pendingScroll.current = undefined
  }, [routeKey])

  useEffect(() => {
    window.history.replaceState({ ...window.history.state, [historyIndexKey]: activeHistoryIndex.current }, '', window.location.href)
  }, [])

  useEffect(() => {
    const previousRestoration = window.history.scrollRestoration
    window.history.scrollRestoration = 'manual'
    const handlePopState = () => {
      if (restoringHistory.current) {
        restoringHistory.current = false
        if (readHistoryIndex() === activeHistoryIndex.current) return
      }
      void navigate(window.location.pathname + window.location.search, false, readHistoryIndex())
    }
    window.addEventListener('popstate', handlePopState)
    return () => {
      window.removeEventListener('popstate', handlePopState)
      window.history.scrollRestoration = previousRestoration
    }
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

  const changeTheme = useCallback((preference: BrowserThemePreference) => {
    saveBrowserThemePreference(preference)
    themePreferenceRef.current = preference
    const next = resolveBrowserTheme(preference)
    const changed = next !== themeRef.current
    themeRef.current = next
    themeTransition.apply(() => {
      applyBrowserTheme(next, page.layout.theme.colors, false)
      flushSync(() => { setThemePreference(preference); setTheme(next) })
    }, changed)
  }, [page.layout.theme.colors, themeTransition])

  useEffect(() => {
    if (themePreference !== 'system' || !window.matchMedia) return
    const media = window.matchMedia('(prefers-color-scheme: dark)')
    const update = () => {
      if (themePreferenceRef.current !== 'system') return
      const next = media.matches ? 'gf-dark' : 'gf-light'
      if (next === themeRef.current) {
        applyBrowserTheme(next, page.layout.theme.colors, false)
        return
      }
      themeRef.current = next
      themeTransition.apply(() => {
        setTheme(next)
        applyBrowserTheme(next, page.layout.theme.colors, false)
      }, false)
    }
    update()
    media.addEventListener('change', update)
    return () => media.removeEventListener('change', update)
  }, [themePreference, page.layout.theme.colors, themeTransition])

  const content = pathname === '/admin/categories'
    ? <CategoriesManagementPage api={api} permissions={page.layout.viewer.adminPermissions} text={text} />
    : pathname === '/admin/access-groups'
      ? <AccessGroupsManagementPage api={api} text={accessGroupText} onNavigate={navigate} />
      : pathname === '/admin/roles'
        ? <RolesManagementPage api={api} text={roleText} />
      : pathname === '/admin/users'
        ? <UsersManagementPage key={routeKey} api={api} initialData={listData?.kind === 'users' ? listData.data : undefined} onUrlChange={rememberQueryUrl} text={userText} locale={locale} canResetMFA={page.layout.viewer?.adminPermissions?.includes(0)} currentUserId={page.layout.viewer?.id} />
      : pathname === '/admin/posts'
        ? <PostsManagementPage key={routeKey} api={api} initialData={listData?.kind === 'posts' ? listData.data : undefined} onUrlChange={rememberQueryUrl} text={postText} />
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
      : pathname === '/admin/auth-logs'
        ? <AuthLogsPage key={page.layout.viewer.id} api={api} locale={locale} />
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
      : pathname === '/admin/settings/agent'
        ? <AgentSettingsPage api={api} text={systemSettingsText} />
      : pathname === '/admin/settings/trust'
        ? <TrustSettingsPage api={api} locale={locale} />
      : pathname === '/admin/settings/announcement'
        ? <AnnouncementSettingsPage api={api} text={contentSettingsText} />
      : pathname === '/admin/settings/http-notify'
        ? <HttpNotifySettingsPage api={api} text={moderationSettingsText} locale={locale} />
      : pathname === '/admin/settings/sensitive-words'
        ? <SensitiveWordsSettingsPage api={api} text={moderationSettingsText} />
      : pathname === '/admin/settings/oauth'
        ? <OAuthSettingsPage api={api} text={identityText} />
      : pathname === '/admin/settings/oidc-provider'
        ? <OIDCProviderSettingsPage api={api} text={identityText} />
      : pathname === '/admin'
                        ? <DashboardPage api={api} text={dashboardText} locale={locale} />
                        : <div className="flex flex-1 items-center justify-center p-6"><div className="max-w-md rounded-xl border bg-card p-6 text-center"><h2 className="text-lg font-semibold">{text(titleKey)}</h2><p className="mt-2 text-sm text-muted-foreground">{text('notAvailable')}</p></div></div>

  return <TooltipProvider><SidebarProvider className="admin-shell mx-auto max-w-[1600px]" style={shellVariables}>{isNavigating ? <div role="progressbar" aria-label={text('loading')} className="admin-navigation-progress fixed inset-x-0 top-0 z-50 h-0.5 overflow-hidden bg-primary/15"><div className="h-full w-1/3 bg-primary" /></div> : null}<AppSidebar layout={page.layout} pathname={pathname} text={text} onNavigate={navigate} onPrefetch={prefetch} variant="inset" /><SidebarInset className="z-20 min-w-0 overflow-clip" aria-busy={isNavigating}><SiteHeader title={text(titleKey)} locale={locale} themePreference={themePreference} onLocaleChange={changeLocale} onThemeChange={changeTheme} /><div className="flex min-w-0 flex-1 flex-col">{navigationError ? <p role="alert" className="px-4 py-2 text-sm text-destructive">{navigationError}</p> : null}<Suspense fallback={<div className="flex min-h-48 items-center justify-center text-sm text-muted-foreground">{text('loading')}</div>}>{content}</Suspense></div></SidebarInset></SidebarProvider><Toaster position="bottom-right" /></TooltipProvider>
}
