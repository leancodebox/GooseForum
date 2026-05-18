import { defineAsyncComponent } from 'vue'

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

export const asyncPageComponents = {
  HomePage: defineAsyncComponent(pageLoaders['home.index']),
  ArticlePage: defineAsyncComponent(pageLoaders['article.detail']),
  UserPage: defineAsyncComponent(pageLoaders['user.profile']),
  CategoryPage: defineAsyncComponent(pageLoaders['category.index']),
  LinksPage: defineAsyncComponent(pageLoaders['links.index']),
  SponsorsPage: defineAsyncComponent(pageLoaders['sponsors.index']),
  NotificationsPage: defineAsyncComponent(pageLoaders['notifications.index']),
  MessagesPage: defineAsyncComponent(pageLoaders['messages.index']),
  SettingsPage: defineAsyncComponent(pageLoaders['settings.index']),
  PublishPage: defineAsyncComponent(pageLoaders['publish.index']),
  SearchPage: defineAsyncComponent(pageLoaders['search.index']),
  LoginPage: defineAsyncComponent(pageLoaders['auth.login']),
  ResetPasswordPage: defineAsyncComponent(pageLoaders['auth.resetPassword']),
  ErrorPage: defineAsyncComponent(pageLoaders['error.notFound']),
}

export async function preloadPageComponent(component: string) {
  const loader = pageLoaders[component as PageComponentName]
  if (!loader) return
  await loader()
}
