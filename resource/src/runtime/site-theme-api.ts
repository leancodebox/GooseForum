import type { SiteThemeConfig } from '@/types/payload'
import { resolveApiMessage } from './api-message'
import { i18n } from './i18n'

interface ApiResponse<T> {
  code?: number
  messageCode?: string
  params?: Record<string, unknown>
  result?: T
  data?: T
}

function t(key: string) {
  return i18n.global.t(key)
}

async function readApiResponse<T>(response: Response, fallback: string): Promise<T> {
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }
  const data = (await response.json()) as ApiResponse<T>
  if (data.code !== undefined && data.code !== 0) {
    throw new Error(resolveApiMessage(data, fallback))
  }
  return (data.result ?? data.data) as T
}

export async function saveSiteTheme(settings: SiteThemeConfig): Promise<SiteThemeConfig> {
  const response = await fetch('/api/admin/save-site-theme', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ settings }),
  })
  return readApiResponse<SiteThemeConfig>(response, t('api.themeSaveFailed'))
}

export async function publishSiteTheme(): Promise<SiteThemeConfig> {
  const response = await fetch('/api/admin/publish-site-theme', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
  })
  return readApiResponse<SiteThemeConfig>(response, t('api.themeSaveFailed'))
}
