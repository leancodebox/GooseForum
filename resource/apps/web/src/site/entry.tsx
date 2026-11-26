import type { AnyPagePayload } from '@gooseforum/client'
import type { AreaNavigation } from '@gooseforum/runtime'
import { detectBrowserLocale, readLocalePreference } from '@gooseforum/runtime/browser-host'
import { prepareGoosePage } from '@gooseforum/runtime/prepared-page'
import { prepareGooseTranslations, goosePageNamespaces } from '@gooseforum/runtime/i18n'
import { preloadPageIcons } from '@gooseforum/theme-default/app/site-app'
import { sitePageSource } from '../host/browser-runtime'
import { SiteApp } from './SiteApp'
export { default as stylesheet } from '@gooseforum/theme-default/styles/site.css?url'

export async function prepare(page: AnyPagePayload, url: URL) {
  const locale = detectBrowserLocale(readLocalePreference(url), page.layout.site.defaultLocale)
  await Promise.all([
    preloadPageIcons(page),
    prepareGooseTranslations(locale, goosePageNamespaces(page.component)),
    prepareGoosePage(page.component),
  ])
  return (navigateArea: AreaNavigation, ready: () => void) =>
    <SiteApp key={url.href} pageSource={sitePageSource} initialPage={page} navigateArea={navigateArea} onInitialContentReady={ready} />
}
