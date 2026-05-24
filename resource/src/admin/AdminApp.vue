<script setup lang="ts">
import { computed, markRaw, shallowRef, watch } from 'vue'
import type { Component } from 'vue'
import AdminLayout from '@/admin/layouts/AdminLayout.vue'
import AdminPageLoading from './pages/AdminPageLoading.vue'
import { currentAdminPath, installAdminRouter } from '@/admin/runtime/router'
import type { AdminPayload, ManageHomeProps } from '@/admin/types'

const props = defineProps<{
  payload: AdminPayload
}>()

installAdminRouter()

type PageKey =
  | 'placeholder'
  | 'settings'
  | 'stats'
  | 'badges'
  | 'categories'
  | 'links'
  | 'posts'
  | 'roles'
  | 'sponsors'
  | 'users'

type PageLoader = () => Promise<{ default: Component }>

const pageLoaders: Record<PageKey, PageLoader> = {
  placeholder: () => import('./pages/AdminPlaceholderPage.vue'),
  settings: () => import('./pages/AdminSettingsPage.vue'),
  stats: () => import('./pages/StatsPage.vue'),
  badges: () => import('./pages/management/BadgesManagementPage.vue'),
  categories: () => import('./pages/management/CategoriesManagementPage.vue'),
  links: () => import('./pages/management/LinksManagementPage.vue'),
  posts: () => import('./pages/management/PostsManagementPage.vue'),
  roles: () => import('./pages/management/RolesManagementPage.vue'),
  sponsors: () => import('./pages/management/SponsorsManagementPage.vue'),
  users: () => import('./pages/management/UsersManagementPage.vue'),
}

const placeholderPages: Record<string, { title: string, description: string }> = {
  '/admin/unknown': { title: '管理页面', description: '这个管理入口还在迁移中。' },
}

const settingsPages = {
  '/admin/settings/site-info': 'site-info',
  '/admin/settings/mail': 'mail',
  '/admin/settings/security': 'security',
  '/admin/settings/posting': 'posting',
  '/admin/settings/announcement': 'announcement',
} as const

const pageCache = new Map<PageKey, Component>()
const activeComponent = shallowRef<Component>()
const activeComponentProps = shallowRef<Record<string, unknown>>({})

const isInitialLoading = computed(() => !activeComponent.value)

function resolvePage(path: string): { key: PageKey, props: Record<string, unknown> } {
  const settingsKind = settingsPages[path as keyof typeof settingsPages]
  const placeholder = placeholderPages[path]
  switch (path) {
    case '/admin':
      return { key: 'stats', props: {} }
    case '/admin/users':
      return { key: 'users', props: {} }
    case '/admin/roles':
      return { key: 'roles', props: {} }
    case '/admin/categories':
      return { key: 'categories', props: {} }
    case '/admin/posts':
      return { key: 'posts', props: {} }
    case '/admin/links':
      return { key: 'links', props: {} }
    case '/admin/sponsors':
      return { key: 'sponsors', props: {} }
    case '/admin/badges':
      return { key: 'badges', props: {} }
    default:
      if (settingsKind) return { key: 'settings', props: { kind: settingsKind } }
      if (placeholder) {
        return {
          key: 'placeholder',
          props: {
            title: placeholder.title,
            description: placeholder.description,
          },
        }
      }
      return { key: 'stats', props: {} }
  }
}

async function loadPage(key: PageKey): Promise<Component> {
  const cached = pageCache.get(key)
  if (cached) return cached
  const page = markRaw((await pageLoaders[key]()).default)
  pageCache.set(key, page)
  return page
}

let requestId = 0

function runWhenIdle(callback: () => void) {
  if (typeof window === 'undefined') return
  if ('requestIdleCallback' in window) {
    window.requestIdleCallback(callback, { timeout: 3000 })
    return
  }
  globalThis.setTimeout(callback, 1200)
}

watch(
  currentAdminPath,
  async (path) => {
    const currentRequestId = ++requestId
    const nextPage = resolvePage(path)
    const nextComponent = await loadPage(nextPage.key)
    if (currentRequestId !== requestId) return
    activeComponent.value = nextComponent
    activeComponentProps.value = nextPage.props
  },
  { immediate: true },
)

runWhenIdle(() => {
  void loadPage('stats')
})
</script>

<template>
  <AdminLayout :layout="payload.layout">
    <component
      :is="isInitialLoading ? AdminPageLoading : activeComponent"
      :payload="payload as AdminPayload<ManageHomeProps>"
      v-bind="activeComponentProps"
    />
  </AdminLayout>
</template>
