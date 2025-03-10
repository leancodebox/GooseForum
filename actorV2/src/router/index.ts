import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/userStore'

import Empty from '../views/Empty.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Empty,
    },
    {
      path: '/post-edit',
      name: 'postEdit',
      component: ()=>import('../views/ArticlePublish.vue'),
    },
    {
      path: '/about',
      name: 'about',
      component: ()=>import('../views/AboutView.vue'),
    },
    {
      path: '/notifications',
      name: 'notifications',
      component: () => import('../views/NotificationsView.vue')
    }
  ],
})

// 路由守卫
router.beforeEach(async (to, from) => {
  const userStore = useUserStore()
  await userStore.fetchUserInfo()
})

export default router
