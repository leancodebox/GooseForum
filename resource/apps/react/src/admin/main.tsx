import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BootstrapError } from '@gooseforum/react/app'
import '@gooseforum/react/globals.css'
import { adminPageSource, prepareDocument } from '../browser-runtime'
import { AdminApp } from './AdminApp'
import { canVisitAdminPath, firstAdminPath } from './access'

const rootElement = document.querySelector<HTMLDivElement>('#goose-admin-app')

if (!rootElement) {
  throw new Error('Missing #goose-admin-app mount element')
}

const root = createRoot(rootElement)

try {
  const page = await adminPageSource.load(new URL(window.location.href))
  const pathname = window.location.pathname.replace(/\/+$/, '') || '/admin'
  if (!page.layout.viewer.canAccessAdmin) {
    window.location.replace(`/login?redirect=${encodeURIComponent(`${pathname}${window.location.search}`)}`)
  } else if (!canVisitAdminPath(page.layout.viewer.adminPermissions, pathname)) {
    window.location.replace(firstAdminPath(page.layout.viewer.adminPermissions))
  } else {
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
      <BootstrapError error={error} />
    </StrictMode>,
  )
}
