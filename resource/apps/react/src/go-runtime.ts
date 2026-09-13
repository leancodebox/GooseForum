import {
  createGooseClient,
  updateDocumentMetadata,
  type AnyPagePayload,
  type PagePayload,
  type ThemePayload,
} from '@gooseforum/client'

const siteClient = createGooseClient<AnyPagePayload>()
const adminClient = createGooseClient<PagePayload>({
  pages: {
    components: ['admin.shell'],
  },
})

export async function loadInitialPage(): Promise<AnyPagePayload> {
  if (document.querySelector('#goose-payload')) {
    return siteClient.readInitialPayload()
  }

  const path = `${window.location.pathname}${window.location.search}`
  return siteClient.pages.fetch(`/__goose_page${path}`)
}

export async function loadInitialAdminPage(): Promise<PagePayload> {
  if (document.querySelector('#goose-payload')) {
    return adminClient.readInitialPayload()
  }

  const path = `${window.location.pathname}${window.location.search}`
  return adminClient.pages.fetch(`/__goose_page${path}`)
}

export function prepareDocument(payload: PagePayload) {
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
