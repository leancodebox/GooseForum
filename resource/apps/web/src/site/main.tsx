import { startSiteStatus } from '../host/site-status'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '@gooseforum/theme-default/styles/site.css'
import { sitePageSource, detectBrowserLocale } from '../host/browser-runtime'
import { SiteApp } from './SiteApp'
import { preloadPageIcons } from '@gooseforum/theme-default/app/site-app'
import { BrowserBootstrapError } from '../host/bootstrap'
import { prepareGoosePage } from '@gooseforum/runtime/prepared-page'
import { prepareGooseTranslations, goosePageNamespaces } from '@gooseforum/runtime/i18n'

const rootElement = document.querySelector<HTMLDivElement>('#goose-app')

if (!rootElement) {
  throw new Error('Missing #goose-app mount element')
}

const stopSiteStatus = startSiteStatus(sitePageSource.api, false)
if (import.meta.hot) import.meta.hot.dispose(stopSiteStatus)

let splash = document.getElementById('goose-splash')
function dismissSplash() {
  splash?.remove()
  splash = null
}
const root = createRoot(rootElement)
try {
  const initialPage = sitePageSource.readInitial?.()
  await Promise.all([
    initialPage ? preloadPageIcons(initialPage) : undefined,
    prepareGooseTranslations(detectBrowserLocale(), initialPage ? goosePageNamespaces(initialPage.component) : []),
    initialPage ? prepareGoosePage(initialPage.component) : undefined,
  ])

  root.render(
    <StrictMode>
      <SiteApp pageSource={sitePageSource} initialPage={initialPage} onInitialContentReady={dismissSplash} />
    </StrictMode>,
  )
} catch (error) {
  root.render(<BrowserBootstrapError onReady={dismissSplash} error={error} onRetry={() => window.location.reload()} />)
}
