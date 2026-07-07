import type { Component } from 'vue'

export const pageLoaders = {
  'home.index': () => import('@/site/pages/HomePage.vue'),
  'article.detail': () => import('@/site/pages/TopicPage.vue'),
  'user.profile': () => import('@/site/pages/UserPage.vue'),
  'category.index': () => import('@/site/pages/CategoryPage.vue'),
  'links.index': () => import('@/site/pages/LinksPage.vue'),
  'sponsors.index': () => import('@/site/pages/SponsorsPage.vue'),
  'notifications.index': () => import('@/site/pages/NotificationsPage.vue'),
  'messages.index': () => import('@/site/pages/MessagesPage.vue'),
  'drafts.index': () => import('@/site/pages/DraftsPage.vue'),
  'moderation.index': () => import('@/site/pages/ModerationPage.vue'),
  'settings.index': () => import('@/site/pages/SettingsPage.vue'),
  'theme.preview': () => import('@/site/pages/ThemePreviewPage.vue'),
  'publish.index': () => import('@/site/pages/PublishPage.vue'),
  'search.index': () => import('@/site/pages/SearchPage.vue'),
  'auth.login': () => import('@/site/pages/LoginPage.vue'),
  'auth.resetPassword': () => import('@/site/pages/ResetPasswordPage.vue'),
  'error.notFound': () => import('@/site/pages/ErrorPage.vue'),
} as const

export type PageComponentName = keyof typeof pageLoaders

const componentCache = new Map<string, Component>()

export async function resolvePageComponent(component: string): Promise<Component | null> {
  const cached = componentCache.get(component)
  if (cached) return cached

  const loader = pageLoaders[component as PageComponentName]
  if (!loader) return null

  const mod = await loader()
  const resolved = mod.default
  componentCache.set(component, resolved)
  return resolved
}
