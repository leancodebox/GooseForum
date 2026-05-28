import { adminText } from '@/admin/runtime/i18n-text'
import { createRouter, createWebHistory } from 'vue-router'

const settingsPages = {
  '/admin/settings/site-info': 'site-info',
  '/admin/settings/mail': 'mail',
  '/admin/settings/security': 'security',
  '/admin/settings/posting': 'posting',
  '/admin/settings/announcement': 'announcement',
} as const

export const adminRouter = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/admin',
      component: () => import('@/admin/pages/StatsPage.vue'),
    },
    {
      path: '/admin/users',
      component: () => import('@/admin/pages/management/UsersManagementPage.vue'),
    },
    {
      path: '/admin/roles',
      component: () => import('@/admin/pages/management/RolesManagementPage.vue'),
    },
    {
      path: '/admin/categories',
      component: () => import('@/admin/pages/management/CategoriesManagementPage.vue'),
    },
    {
      path: '/admin/posts',
      component: () => import('@/admin/pages/management/PostsManagementPage.vue'),
    },
    {
      path: '/admin/links',
      component: () => import('@/admin/pages/management/LinksManagementPage.vue'),
    },
    {
      path: '/admin/sponsors',
      component: () => import('@/admin/pages/management/SponsorsManagementPage.vue'),
    },
    {
      path: '/admin/badges',
      component: () => import('@/admin/pages/management/BadgesManagementPage.vue'),
    },
    {
      path: '/admin/opt-records',
      component: () => import('@/admin/pages/management/OptRecordsManagementPage.vue'),
    },
    ...Object.entries(settingsPages).map(([path, kind]) => ({
      path,
      component: () => import('@/admin/pages/AdminSettingsPage.vue'),
      meta: { pageProps: { kind } },
    })),
    {
      path: '/admin/unknown',
      component: () => import('@/admin/pages/AdminPlaceholderPage.vue'),
      meta: {
        pageProps: {
          title: adminText('k002d'),
          description: adminText('k002e'),
        },
      },
    },
    {
      path: '/admin/:pathMatch(.*)*',
      redirect: '/admin',
    },
  ],
})
