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
  ],
})

export default router
