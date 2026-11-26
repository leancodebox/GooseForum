import type { PagePayload } from '@gooseforum/client'
import type { AreaNavigation } from '@gooseforum/runtime'
import { detectBrowserLocale, readLocalePreference } from '@gooseforum/runtime/browser-host'
import { browserPageSource } from '../host/browser-runtime'
import { AdminApp } from './AdminApp'
import { canVisitAdminPath, firstAdminPath } from './access'
import { normalizeAdminPath } from './navigation'
import { prepareAdminPage } from './page-registry'
import { adminPageNamespaces, prepareAdminTranslations } from './translation-loader'
import { prepareAdminListData } from './list-data'
export { default as stylesheet } from './styles/admin.css?url'

export function adminDestination(page: PagePayload, url: URL) {
  const path = normalizeAdminPath(url.pathname)
  const viewer = page.layout.viewer
  if (!viewer.canAccessAdmin) return new URL('/login?redirect=' + encodeURIComponent(path + url.search), url)
  if (!canVisitAdminPath(viewer.adminPermissions, path)) return new URL(firstAdminPath(viewer.adminPermissions), url)
  return url
}

export async function prepare(page: PagePayload, url: URL) {
  const path = normalizeAdminPath(url.pathname)
  const locale = detectBrowserLocale(readLocalePreference(url), page.layout.site.defaultLocale)
  const [, , initialListData] = await Promise.all([
    prepareAdminPage(path),
    prepareAdminTranslations(locale, adminPageNamespaces(path)),
    prepareAdminListData(browserPageSource.admin, url),
  ])
  return (navigateArea: AreaNavigation) =>
    <AdminApp page={page} api={browserPageSource.admin} initialListData={initialListData} navigateArea={navigateArea} />
}
