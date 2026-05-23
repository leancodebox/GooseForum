<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue'
import type { Component } from 'vue'
import AdminPageLoading from './pages/AdminPageLoading.vue'
import { currentAdminPath, installAdminRouter } from '@/admin/runtime/router'
import type { AdminPayload, ManageHomeProps } from '@/admin/types'

const props = defineProps<{
  payload: AdminPayload
}>()

installAdminRouter()

const asyncPage = (loader: () => Promise<{ default: Component }>) => defineAsyncComponent({
  loader,
  loadingComponent: AdminPageLoading,
  delay: 80,
})

const pages = {
  placeholder: asyncPage(() => import('./pages/AdminPlaceholderPage.vue')),
  settings: asyncPage(() => import('./pages/AdminSettingsPage.vue')),
  home: asyncPage(() => import('./pages/ManageHomePage.vue')),
  stats: asyncPage(() => import('./pages/StatsPage.vue')),
  badges: asyncPage(() => import('./pages/management/BadgesManagementPage.vue')),
  categories: asyncPage(() => import('./pages/management/CategoriesManagementPage.vue')),
  links: asyncPage(() => import('./pages/management/LinksManagementPage.vue')),
  posts: asyncPage(() => import('./pages/management/PostsManagementPage.vue')),
  roles: asyncPage(() => import('./pages/management/RolesManagementPage.vue')),
  sponsors: asyncPage(() => import('./pages/management/SponsorsManagementPage.vue')),
  users: asyncPage(() => import('./pages/management/UsersManagementPage.vue')),
}

const placeholderPages: Record<string, { title: string, description: string }> = {
  '/manage/unknown': { title: '管理页面', description: '这个管理入口还在迁移中。' },
}

const settingsPages = {
  '/manage/settings/site-info': 'site-info',
  '/manage/settings/mail': 'mail',
  '/manage/settings/security': 'security',
  '/manage/settings/posting': 'posting',
  '/manage/settings/announcement': 'announcement',
} as const

const settingsKind = computed(() => settingsPages[currentAdminPath.value as keyof typeof settingsPages])

const placeholder = computed(() => placeholderPages[currentAdminPath.value])

const componentProps = computed(() => {
  if (settingsKind.value) return { kind: settingsKind.value }
  return {
    title: placeholder.value?.title,
    description: placeholder.value?.description,
  }
})

const component = computed(() => {
  switch (currentAdminPath.value) {
    case '/manage/metrics':
      return pages.stats
    case '/manage/users':
      return pages.users
    case '/manage/roles':
      return pages.roles
    case '/manage/categories':
      return pages.categories
    case '/manage/posts':
      return pages.posts
    case '/manage/links':
      return pages.links
    case '/manage/sponsors':
      return pages.sponsors
    case '/manage/badges':
      return pages.badges
    case '/manage':
      return pages.home
    default:
      if (settingsKind.value) return pages.settings
      return placeholderPages[currentAdminPath.value] ? pages.placeholder : pages.home
  }
})
</script>

<template>
  <component
    :is="component"
    :payload="payload as AdminPayload<ManageHomeProps>"
    v-bind="componentProps"
  />
</template>
