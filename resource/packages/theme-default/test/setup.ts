import { vi } from 'vitest'
import { supportedLocales } from '@gooseforum/client/i18n/locale'
import { gooseNamespaces, prepareGooseTranslations } from '@gooseforum/runtime/i18n'

// Component unit tests are synchronous; production prepares only the route's namespaces.
await Promise.all(supportedLocales.map(locale => prepareGooseTranslations(locale, gooseNamespaces)))

HTMLElement.prototype.scrollIntoView = vi.fn()

vi.stubGlobal('ResizeObserver', class ResizeObserverMock {
  observe() {}
  unobserve() {}
  disconnect() {}
})
