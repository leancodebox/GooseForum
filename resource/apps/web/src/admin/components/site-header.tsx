import type { LayoutPayload } from '@gooseforum/client'
import { GooseLink, GooseNavigationProvider } from '@gooseforum/runtime'
import { searchUrl } from '@gooseforum/runtime/search-url'
import { useUnreadStatus } from '@gooseforum/runtime/unread-status'
import { authLocales, localeLabels, type AuthLocale } from '@gooseforum/runtime/i18n/auth'
import type { BrowserThemePreference } from '@gooseforum/runtime/browser-host'
import { SiteHeader as SharedSiteHeader } from '@gooseforum/ui/components/site-header'
import { HeaderSlot } from '@gooseforum/ui/components/header-host'
import { ShellMenuTrigger, type useShellSidebar } from '@gooseforum/ui/components/shell-layout'
import type { ReactNode } from 'react'
import { PageNavigationIcons } from '@gooseforum/ui/components/navigation-icon'
import { getAdminDictionary } from '../translation-loader'

type HeaderDictionary = Record<string, string | Record<string, string>>

export function SiteHeader({ layout, locale, navigate, themePreference, onLocaleChange, onThemeChange, homeLabel, onLogout, sidebar, navigation }: {
  sidebar: ReturnType<typeof useShellSidebar>
  navigation: ReactNode
  layout: LayoutPayload
  onLogout(): Promise<void>
  homeLabel: string
  navigate(path: string): Promise<void>
  locale: AuthLocale
  themePreference: BrowserThemePreference
  onLocaleChange(locale: AuthLocale): void
  onThemeChange(preference: BrowserThemePreference): void
}) {
  const dictionary = getAdminDictionary<HeaderDictionary>('site-header', locale)
  const unread = useUnreadStatus(layout.unread)
  function translate(key: string, fallback = key) {
    const [group, item] = key.split('.')
    const value = dictionary[group]
    return typeof value === 'string' ? value : value?.[item] ?? fallback
  }
  return <HeaderSlot>
    <GooseNavigationProvider value={href => navigate(href)}>
      <PageNavigationIcons value={layout.icons}>
        <SharedSiteHeader site={layout.site} viewer={layout.viewer} header={layout.header} unread={unread}
          homeLabel={homeLabel} Link={GooseLink} sidebarTrigger={<ShellMenuTrigger sidebar={sidebar} navigation={navigation} t={translate} />}
          t={translate} locale={locale} locales={authLocales} localeLabels={localeLabels} onLocaleChange={onLocaleChange}
          themePreference={themePreference} onThemeChange={onThemeChange} onSearch={query => void navigate(searchUrl(query))}
          onLogout={onLogout} />
      </PageNavigationIcons>
    </GooseNavigationProvider>
  </HeaderSlot>
}
