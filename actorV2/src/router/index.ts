import { createRouter, createWebHistory } from 'vue-router'

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
  ],
})

export default router
