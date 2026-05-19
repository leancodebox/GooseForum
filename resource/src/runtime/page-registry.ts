import type { Component } from 'vue'

export const pageLoaders = {
  'home.index': () => import('../pages/HomePage.vue'),
  'article.detail': () => import('../pages/ArticlePage.vue'),
  'user.profile': () => import('../pages/UserPage.vue'),
  'category.index': () => import('../pages/CategoryPage.vue'),
  'links.index': () => import('../pages/LinksPage.vue'),
  'sponsors.index': () => import('../pages/SponsorsPage.vue'),
  'notifications.index': () => import('../pages/NotificationsPage.vue'),
  'messages.index': () => import('../pages/MessagesPage.vue'),
  'settings.index': () => import('../pages/SettingsPage.vue'),
  'publish.index': () => import('../pages/PublishPage.vue'),
  'search.index': () => import('../pages/SearchPage.vue'),
  'auth.login': () => import('../pages/LoginPage.vue'),
  'auth.resetPassword': () => import('../pages/ResetPasswordPage.vue'),
  'error.notFound': () => import('../pages/ErrorPage.vue'),
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
