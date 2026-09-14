import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '@gooseforum/react/globals.css'
import { sitePageSource } from '../browser-runtime'
import { SiteApp } from './SiteApp'

const rootElement = document.querySelector<HTMLDivElement>('#goose-app')

if (!rootElement) {
  throw new Error('Missing #goose-app mount element')
}

const root = createRoot(rootElement)

root.render(
  <StrictMode>
    <SiteApp pageSource={sitePageSource} />
  </StrictMode>,
)
