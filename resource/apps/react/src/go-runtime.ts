import {
  createGooseClient,
  updateDocumentMetadata,
  type AnyPagePayload,
  type ThemePayload,
} from '@gooseforum/client'

const client = createGooseClient<AnyPagePayload>()

export async function loadInitialPage(): Promise<AnyPagePayload> {
  if (document.querySelector('#goose-payload')) {
    return client.readInitialPayload()
  }

  const path = `${window.location.pathname}${window.location.search}`
  return client.pages.fetch(`/__goose_page${path}`)
}

export function prepareDocument(payload: AnyPagePayload) {
  updateDocumentMetadata(payload)
  document.documentElement.lang ||= 'zh-CN'
  applyTheme(payload.layout.theme)
}

function applyTheme(theme: ThemePayload) {
  document.documentElement.dataset.theme = theme.current
  document.documentElement.style.colorScheme = theme.current === 'gf-dark' ? 'dark' : 'light'
  document.querySelector('meta[name="theme-color"]')?.setAttribute('content', theme.themeColor)

  let link = document.querySelector<HTMLLinkElement>('#goose-site-theme-link')
  if (!theme.enabled || !theme.href) {
    link?.remove()
    return
  }
  if (!link) {
    link = document.createElement('link')
    link.id = 'goose-site-theme-link'
    link.rel = 'stylesheet'
    document.head.appendChild(link)
  }
  link.href = theme.href
}
