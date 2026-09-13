import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BootstrapError, GooseApp } from '@gooseforum/react/app'
import '@gooseforum/react/globals.css'
import { loadInitialPage, prepareDocument } from './go-runtime'

const rootElement = document.querySelector<HTMLDivElement>('#goose-app')

if (!rootElement) {
  throw new Error('Missing #goose-app mount element')
}

const root = createRoot(rootElement)

try {
  const page = await loadInitialPage()
  prepareDocument(page)
  root.render(
    <StrictMode>
      <GooseApp initialPage={page} />
    </StrictMode>,
  )
} catch (error) {
  root.render(
    <StrictMode>
      <BootstrapError error={error} />
    </StrictMode>,
  )
}
