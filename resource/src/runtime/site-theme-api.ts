import { createGooseClient, GooseClientError, type SiteThemeConfig } from '@gooseforum/client'
import { resolveApiMessage } from './api-message'
import { i18n } from './i18n'

const client = createGooseClient()

export async function saveSiteTheme(settings: SiteThemeConfig): Promise<SiteThemeConfig> {
  return localize(() => client.api.themes.save(settings))
}

export async function publishSiteTheme(): Promise<SiteThemeConfig> {
  return localize(() => client.api.themes.publish())
}

async function localize<T>(request: () => Promise<T>) {
  try {
    return await request()
  } catch (error) {
    if (error instanceof GooseClientError && error.messageCode) {
      throw new Error(resolveApiMessage(error, i18n.global.t('api.themeSaveFailed')))
    }
    throw error
  }
}
