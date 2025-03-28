import { createRouter, createWebHistory } from 'vue-router'
import AdminLayout from '../layouts/AdminLayout.vue'
import Dashboard from '../views/Dashboard.vue'
import UserManagement from '../views/UserManagement.vue'
import PostManagement from '../views/PostManagement.vue'
import CategoryManagement from '../views/CategoryManagement.vue'
import SystemSettings from '../views/SystemSettings.vue'
import Login from '../views/Login.vue'
import RolesView from '../views/RolesView.vue'
import {useUserStore} from "../stores/auth.ts";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: Login,
      meta: {
        title: '登录',
        requiresAuth: false
      }
    },
    {
      path: '/admin',
      component: AdminLayout,
      children: [
        {
          path: '',
          name: 'dashboard',
          component: Dashboard,
          meta: {
            title: '仪表盘',
            requiresAuth: true
          }
        },
        {
          path: 'users',
          name: 'users',
          component: UserManagement,
          meta: {
            title: '用户管理',
            requiresAuth: true
          }
        },
        {
          path: 'posts',
          name: 'posts',
          component: PostManagement,
          meta: {
            title: '帖子管理',
            requiresAuth: true
          }
        },
        {
          path: 'categories',
          name: 'categories',
          component: CategoryManagement,
          meta: {
            title: '分类管理',
            requiresAuth: true
          }
        },
        {
          path: 'settings',
          name: 'settings',
          component: SystemSettings,
          meta: {
            title: '系统设置',
            requiresAuth: true
          }
        },
        {
          path: '/admin/roles',
          name: 'AdminRoles',
          component: RolesView,
          meta: {
            title: '角色管理',
            requiresAuth: true,
            isAdmin: true
          }
        }
      ]
    }
  ]
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  // 设置页面标题
  document.title = `${to.meta.title} - GooseForum管理系统`
  const userStore = useUserStore()
  await userStore.fetchUserInfo()
  // 暂时不做验证，直接放行
  next()
})

export default router
