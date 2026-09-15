import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '@gooseforum/theme-default/styles/site.css'
import { sitePageSource, detectBrowserLocale } from '../host/browser-runtime'
import { SiteApp } from './SiteApp'
import { BrowserBootstrapError } from '../host/bootstrap'
import { prepareGoosePage } from '@gooseforum/runtime/prepared-page'
import { prepareGooseTranslations, goosePageNamespaces } from '@gooseforum/runtime/i18n'

const rootElement = document.querySelector<HTMLDivElement>('#goose-app')

if (!rootElement) {
  throw new Error('Missing #goose-app mount element')
}

const root = createRoot(rootElement)
try {
  const initialPage = sitePageSource.readInitial?.()
  await Promise.all([
    prepareGooseTranslations(detectBrowserLocale(), initialPage ? goosePageNamespaces(initialPage.component) : []),
    initialPage ? prepareGoosePage(initialPage.component) : undefined,
  ])

  root.render(
    <StrictMode>
      <SiteApp pageSource={sitePageSource} initialPage={initialPage} />
    </StrictMode>,
  )
} catch (error) {
  root.render(<BrowserBootstrapError error={error} onRetry={() => window.location.reload()} />)
}
