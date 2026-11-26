import { readLocalePreference } from "@gooseforum/runtime/browser-host";
import { lazy, Suspense, useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'
import { flushSync } from 'react-dom'
import { commitHistory, historyIndex, listenToHistory, rememberScroll, restoreScroll, useLatestRequest, type AreaNavigation } from '@gooseforum/runtime'
import { applyBrowserTheme, createThemeTransition, detectBrowserTheme, detectBrowserThemePreference, resolveBrowserTheme, saveBrowserThemePreference, type BrowserThemePreference } from '@gooseforum/runtime/browser-host'
import type { GooseAdminApi, PagePayload } from '@gooseforum/client'
import { ShellGrid, useShellSidebar } from '@gooseforum/ui/components/shell-layout'
import { logout } from '@gooseforum/runtime/account'
import { applyBrowserLocale, browserPageSource, detectBrowserLocale } from '../host/browser-runtime'
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
import { createTrustSettingsText } from './trust-settings-i18n'
import { createModerationSettingsText } from './moderation-settings-i18n'
import { createIdentityText } from './identity-settings-i18n'
import { createDashboardText } from './dashboard-i18n'
import { adminNavGroups } from './nav'
import { AuthLogsPage } from './page-registry'
import { AgentSettingsPage } from './page-registry'
import { TrustSettingsPage } from './page-registry'
import { normalizeAdminPath, communityRulePaths, identitySettingPaths } from './navigation'
import { prepareAdminTranslations, adminPageNamespaces, loadedAdminNamespaces } from './translation-loader'
import { prepareAdminListData, type AdminListData } from './list-data'
import { errorMessage } from './error-message'
import { AdminPermission, hasAdminPermission } from './access'

import { DashboardPage, CategoriesManagementPage, AccessGroupsManagementPage, RolesManagementPage, UsersManagementPage, RegistrationInvitationsPage, PostsManagementPage, LinksManagementPage, SponsorsManagementPage, BadgesManagementPage, FileResourcesManagementPage, OptRecordsManagementPage, SiteInfoManagementPage, SiteChromeManagementPage, MailSettingsPage, SecuritySettingsPage, PostingSettingsPage, AnnouncementSettingsPage, HttpNotifySettingsPage, SensitiveWordsSettingsPage, OAuthSettingsPage, OIDCProviderSettingsPage, prepareAdminPage } from './page-registry'

const SettingsSectionLayout = lazy(() => import('./components/settings-section-layout').then(module => ({ default: module.SettingsSectionLayout })))
const Toaster = lazy(() => import('@gooseforum/ui/components/sonner').then(module => ({ default: module.Toaster })))

export function AdminApp({ page, api, initialListData, navigateArea }: { navigateArea?: AreaNavigation; page: PagePayload; api: GooseAdminApi; initialListData?: AdminListData }) {
  const sidebar = useShellSidebar({ keyboardShortcut: true })
  const [pathname, setPathname] = useState(() => normalizeAdminPath(window.location.pathname))
  const [routeKey, setRouteKey] = useState(() => `${historyIndex()}:${window.location.pathname}${window.location.search}`)
  const [listData, setListData] = useState(initialListData)
  const rememberedUrls = useRef(new Map<string, string>())
  const activeHistoryIndex = useRef(historyIndex())
  const restoringHistory = useRef(false)
  const committedUrl = useRef(window.location.pathname + window.location.search)
  const pendingScroll = useRef<{ index?: number } | undefined>(undefined)
  const beginNavigationRequest = useLatestRequest()
  const [isNavigating, setIsNavigating] = useState(false)
  const [navigationError, setNavigationError] = useState('')
  const [locale, setLocale] = useState(() => detectBrowserLocale())
  const [defaultLocale, setDefaultLocale] = useState(page.layout.site?.defaultLocale)
  const localeRef = useRef(locale)
  const [localePreference, setLocalePreference] = useState(readLocalePreference);
  const beginLocaleRequest = useLatestRequest()
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
  const trustSettingsText = useMemo(() => createTrustSettingsText(locale), [locale])
  const moderationSettingsText = useMemo(() => createModerationSettingsText(locale), [locale])
  const identityText = useMemo(() => createIdentityText(locale), [locale])
  const dashboardText = useMemo(() => createDashboardText(locale), [locale])
  const isCommunityRules = communityRulePaths.some(path => path === pathname)
  const isIdentitySettings = identitySettingPaths.some(path => path === pathname)
  const titleKey = isCommunityRules ? 'communityRules' : isIdentitySettings ? 'oauthOidc' : adminNavGroups.flatMap((group) => group.items).find((item) => item.url === pathname)?.label || 'console'
  const rememberQueryUrl = useCallback((url: string) => { committedUrl.current = url }, [])

  useEffect(() => {
    document.documentElement.lang = locale
  }, [locale])

  useEffect(() => {
    document.title = `${text(titleKey)} - ${page.layout.site.name || 'GooseForum'}`
  }, [page.layout.site.name, text, titleKey])

  const navigate = useCallback(async (path: string, updateHistory = true, targetHistoryIndex = historyIndex()) => {
    const isCurrent = beginNavigationRequest()
    const requestedUrl = new URL(path, window.location.href)
    if (navigateArea?.(requestedUrl, updateHistory ? "push" : "pop", activeHistoryIndex.current)) return
    const nextPath = normalizeAdminPath(requestedUrl.pathname)
    // A popstate already changed the address; remember the page being left separately.
    const leavingUrl = committedUrl.current
    rememberedUrls.current.set(pathname, leavingUrl)
    rememberScroll(activeHistoryIndex.current)
    const nextUrl = updateHistory && !requestedUrl.search
      ? new URL(rememberedUrls.current.get(nextPath) || requestedUrl.href, window.location.href)
      : requestedUrl
    nextUrl.pathname = nextPath
    const nextKey = nextUrl.pathname + nextUrl.search
    setNavigationError('')
    function restoreHistory() {
      const delta = activeHistoryIndex.current - historyIndex()
      if (delta) {
        restoringHistory.current = true
        window.history.go(delta)
      } else if (window.location.pathname + window.location.search !== committedUrl.current) {
        commitHistory(committedUrl.current, 'replace', activeHistoryIndex.current)
      }
    }
    if (updateHistory && nextKey === leavingUrl) { restoreHistory(); setIsNavigating(false); return }
    setIsNavigating(true)
    try {
      const language = localeRef.current
      const [, , nextData] = await Promise.all([prepareAdminPage(nextPath), prepareAdminTranslations(language, adminPageNamespaces(nextPath)), prepareAdminListData(api, nextUrl)])
      if (language !== localeRef.current) await prepareAdminTranslations(localeRef.current, adminPageNamespaces(nextPath))
      if (!isCurrent()) return
      const nextHistoryIndex = updateHistory ? commitHistory(nextKey, 'push') : targetHistoryIndex
      committedUrl.current = nextKey
      pendingScroll.current = { index: updateHistory ? undefined : nextHistoryIndex }
      activeHistoryIndex.current = nextHistoryIndex
      pathnameRef.current = nextPath
      setListData(nextData)
      setRouteKey(`${nextHistoryIndex}:${nextKey}`)
      setPathname(nextPath)
    } catch (error) {
      if (isCurrent()) {
        restoreHistory()
        setNavigationError(errorMessage(error, 'Navigation failed'))
      }
    } finally {
      if (isCurrent()) setIsNavigating(false)
    }
  }, [api, beginNavigationRequest, pathname, navigateArea])

  useLayoutEffect(() => {
    if (pendingScroll.current === undefined) return
    restoreScroll(pendingScroll.current.index)
    pendingScroll.current = undefined
  }, [routeKey])

  useEffect(() => {
    const handlePopState = () => {
      if (restoringHistory.current) {
        restoringHistory.current = false
        if (historyIndex() === activeHistoryIndex.current) return
      }
      void navigate(window.location.pathname + window.location.search, false, historyIndex())
    }
    return listenToHistory(handlePopState)
  }, [navigate])

  const prefetch = useCallback((path: string) => {
    void Promise.all([prepareAdminPage(path), prepareAdminTranslations(localeRef.current, adminPageNamespaces(path))]).catch(() => undefined)
  }, [])

  const changeLocale = useCallback(async (preference: typeof locale | "browser", persist = true) => {
    const nextLocale = detectBrowserLocale(preference);
    const isCurrent = beginLocaleRequest()
    setNavigationError('')
    try {
      await prepareAdminTranslations(nextLocale, loadedAdminNamespaces())
      await prepareAdminTranslations(nextLocale, adminPageNamespaces(pathnameRef.current))
      if (!isCurrent()) return
      localeRef.current = nextLocale
      if (persist) {
        applyBrowserLocale(nextLocale, preference)
        committedUrl.current = window.location.pathname + window.location.search
      }
      setLocalePreference(preference)
      setLocale(nextLocale)
    } catch (error) {
      if (isCurrent()) setNavigationError(errorMessage(error, 'Unable to change language'))
    }
  }, [beginLocaleRequest])

  useEffect(() => {
    if (defaultLocale) document.documentElement.dataset.defaultLocale = defaultLocale;
    const update = () => {
      const preference = readLocalePreference();
      if (preference !== localePreference || detectBrowserLocale(preference) !== localeRef.current) {
        void changeLocale(preference, false);
      }
    };
    update();
    window.addEventListener("languagechange", update);
    return () => window.removeEventListener("languagechange", update);
  }, [localePreference, changeLocale, routeKey, defaultLocale]);

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
    ? <CategoriesManagementPage api={api} permissions={page.layout.viewer.adminPermissions} text={text} iconText={settingsText} />
    : pathname === '/admin/access-groups'
      ? <AccessGroupsManagementPage api={api} text={accessGroupText} onNavigate={navigate} />
      : pathname === '/admin/roles'
        ? <RolesManagementPage locale={locale} api={api} text={roleText} />
      : pathname === '/admin/users'
        ? <UsersManagementPage key={routeKey} api={api} initialData={listData?.kind === 'users' ? listData.data : undefined} onUrlChange={rememberQueryUrl} text={userText} locale={locale} canResetMFA={page.layout.viewer?.adminPermissions?.includes(0)} currentUserId={page.layout.viewer?.id} />
      : pathname === '/admin/registration-invitations'
        ? <RegistrationInvitationsPage api={api} text={assetText} locale={locale} />
      : pathname === '/admin/posts'
        ? <PostsManagementPage key={routeKey} api={api} initialData={listData?.kind === 'posts' ? listData.data : undefined} onUrlChange={rememberQueryUrl} text={postText} />
      : pathname === '/admin/links'
        ? <LinksManagementPage api={api} text={pageConfigText} />
      : pathname === '/admin/sponsors'
        ? <SponsorsManagementPage api={api} text={pageConfigText} />
      : pathname === '/admin/badges'
        ? <BadgesManagementPage api={api} text={assetText} canManageRecipients={hasAdminPermission(page.layout.viewer?.adminPermissions || [], AdminPermission.UserManager)} />
      : pathname === '/admin/files/resources'
        ? <FileResourcesManagementPage api={api} text={assetText} />
      : pathname === '/admin/opt-records'
        ? <OptRecordsManagementPage api={api} text={auditText} locale={locale} />
      : pathname === '/admin/auth-logs'
        ? <AuthLogsPage key={page.layout.viewer.id} api={api} locale={locale} />
      : pathname === '/admin/settings/site-info'
        ? <SiteInfoManagementPage api={api} text={settingsText} onDefaultLocaleSaved={setDefaultLocale} />
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
        ? <TrustSettingsPage api={api} locale={locale} text={trustSettingsText} />
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

  const navigation = <AppSidebar layout={page.layout}
    pathname={isCommunityRules ? '/admin/settings/security' : isIdentitySettings ? '/admin/settings/oauth' : pathname}
    text={text} onPrefetch={prefetch} onNavigate={path => {
      sidebar.setMobileOpen(false)
      void navigate(path)
    }} />

  return <>
    <div className="admin-shell min-h-[calc(100svh-4rem)] bg-muted text-foreground">
      <SiteHeader layout={page.layout} homeLabel={text('openSite')} sidebar={sidebar} navigation={navigation}
        onLogout={() => logout(browserPageSource.api, page.layout.viewer.id, () => window.location.assign('/'))}
        navigate={navigate} locale={locale} themePreference={themePreference} onLocaleChange={changeLocale} onThemeChange={changeTheme} />
      {isNavigating && <div role="progressbar" aria-label={text('loading')} className="admin-navigation-progress fixed inset-x-0 top-0 z-50 h-0.5 overflow-hidden bg-primary/15"><div className="h-full w-1/3 bg-primary" /></div>}
      <ShellGrid as="div" sidebar={sidebar} navigation={navigation}>
        <div className="flex min-w-0 flex-col rounded-xl bg-background shadow-sm" aria-busy={isNavigating}>
          {navigationError && <p role="alert" className="px-4 py-2 text-sm text-destructive">{navigationError}</p>}
          <Suspense fallback={<div className="flex min-h-48 items-center justify-center text-sm text-muted-foreground">{text('loading')}</div>}>
            {isCommunityRules || isIdentitySettings
              ? <SettingsSectionLayout section={isIdentitySettings ? 'identity' : 'community'} pathname={pathname} permissions={page.layout.viewer.adminPermissions} text={text} navigate={navigate}>{content}</SettingsSectionLayout>
              : content}
          </Suspense>
        </div>
      </ShellGrid>
    </div>
    <Suspense fallback={null}><Toaster position="bottom-right" /></Suspense>
  </>
}
