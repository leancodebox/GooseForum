import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BootstrapError } from '@gooseforum/react/app'
import '@gooseforum/react/globals.css'
import { loadInitialAdminPage, prepareDocument } from '../go-runtime'
import { AdminApp } from './AdminApp'

const rootElement = document.querySelector<HTMLDivElement>('#goose-admin-app')

if (!rootElement) {
  throw new Error('Missing #goose-admin-app mount element')
}

const root = createRoot(rootElement)

try {
  const page = await loadInitialAdminPage()
  prepareDocument(page)
  root.render(
    <StrictMode>
      <AdminApp />
    </StrictMode>,
  )
} catch (error) {
  root.render(
    <StrictMode>
      <BootstrapError error={error} />
    </StrictMode>,
  )
}
