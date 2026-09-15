import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './styles/admin.css'
import { BrowserBootstrapError } from '../host/bootstrap'
import { adminPageSource, prepareDocument, detectBrowserLocale } from '../host/browser-runtime'
import { AdminApp } from './AdminApp'
import { canVisitAdminPath, firstAdminPath } from './access'
import { prepareAdminPage } from './page-registry'
import { adminPageNamespaces, prepareAdminTranslations } from './translation-loader'

const rootElement = document.querySelector<HTMLDivElement>('#goose-admin-app')

if (!rootElement) {
  throw new Error('Missing #goose-admin-app mount element')
}

const root = createRoot(rootElement)

try {
  const page = adminPageSource.readInitial?.() ?? await adminPageSource.load(new URL(window.location.href))
  const pathname = window.location.pathname.replace(/\/+$/, '') || '/admin'
  if (!page.layout.viewer.canAccessAdmin) {
    window.location.replace(`/login?redirect=${encodeURIComponent(`${pathname}${window.location.search}`)}`)
  } else if (!canVisitAdminPath(page.layout.viewer.adminPermissions, pathname)) {
    window.location.replace(firstAdminPath(page.layout.viewer.adminPermissions))
  } else {
    await Promise.all([prepareAdminPage(pathname), prepareAdminTranslations(detectBrowserLocale(), adminPageNamespaces(pathname))])
    prepareDocument(page)
    root.render(
      <StrictMode>
        <AdminApp page={page} api={adminPageSource.admin} />
      </StrictMode>,
    )
  }
} catch (error) {
  root.render(
    <StrictMode>
      <BrowserBootstrapError error={error} />
    </StrictMode>,
  )
}
