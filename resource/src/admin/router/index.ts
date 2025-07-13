import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import AdminLayout from '../layouts/AdminLayout.vue'

// 路由配置
const routes: RouteRecordRaw[] = [
  {
    path: '/admin/login',
    name: 'AdminLogin',
    component: () => import('../views/Login.vue'),
    meta: {
      title: '管理员登录',
      requiresAuth: false,
      hideInMenu: true
    }
  },
  {
    path: '/admin',
    component: AdminLayout,
    meta: {
      requiresAuth: true
    },
    children: [
      {
        path: '',
        name: 'AdminDashboard',
        component: () => import('../views/Dashboard.vue'),
        meta: {
          title: '仪表盘',
          requiresAuth: true
        }
      },
      {
        path: 'users',
        name: 'AdminUsers',
        component: () => import('../views/UserManagement.vue'),
        meta: {
          title: '用户管理',
          requiresAuth: true
        }
      },
      {
        path: 'roles',
        name: 'AdminRoles',
        component: () => import('../views/RoleManagement.vue'),
        meta: {
          title: '角色管理',
          requiresAuth: true,
          requiresAdmin: true
        }
      },
      {
        path: 'posts',
        name: 'AdminPosts',
        component: () => import('../views/PostManagement.vue'),
        meta: {
          title: '帖子管理',
          requiresAuth: true
        }
      },
      {
        path: 'categories',
        name: 'AdminCategories',
        component: () => import('../views/CategoryManagement.vue'),
        meta: {
          title: '分类管理',
          requiresAuth: true
        }
      },
      {
        path: 'friend-links',
        name: 'AdminFriendLinks',
        component: () => import('../views/FriendLinkManagement.vue'),
        meta: {
          title: '友情链接管理',
          requiresAuth: true
        }
      },
      {
        path: 'sponsors',
        name: 'AdminSponsors',
        component: () => import('../views/SponsorManagement.vue'),
        meta: {
          title: '赞助管理',
          requiresAuth: true
        }
      },
      {
        path: 'tickets',
        name: 'AdminTickets',
        component: () => import('../views/TicketManagement.vue'),
        meta: {
          title: '工单管理',
          requiresAuth: true
        }
      },
      {
        path: 'tickets/view',
        name: 'AdminTicketsView',
        component: () => import('../views/TicketView.vue'),
        meta: {
          title: '工单查看',
          requiresAuth: true
        }
      },
      {
        path: 'web-settings',
        name: 'AdminWebSettings',
        component: () => import('../views/WebSettings.vue'),
        meta: {
          title: '网页设置',
          requiresAuth: true,
          requiresAdmin: true
        }
      },
      {
        path: 'settings',
        name: 'AdminSettings',
        component: () => import('../views/SystemSettings.vue'),
        meta: {
          title: '系统设置',
          requiresAuth: true,
          requiresAdmin: true
        }
      },
      {
        path: 'docs/projects',
        name: 'AdminDocsProjects',
        component: () => import('../views/docs/DocsProjectManagement.vue'),
        meta: {
          title: '文档项目管理',
          requiresAuth: true,
          requiresAdmin: true
        }
      },
      {
        path: 'docs/versions',
        name: 'AdminDocsVersions',
        component: () => import('../views/docs/DocsVersionManagement.vue'),
        meta: {
          title: '文档版本管理',
          requiresAuth: true,
          requiresAdmin: true
        }
      },
      {
        path: 'docs/contents',
        name: 'AdminDocsContents',
        component: () => import('../views/docs/DocsContentManagement.vue'),
        meta: {
          title: '文档内容管理',
          requiresAuth: true,
          requiresAdmin: true
        }
      },
    ]
  },
  {
    path: '/admin/:pathMatch(.*)*',
    name: 'AdminNotFound',
    component: () => import('../views/NotFound.vue'),
    meta: {
      title: '页面不存在',
      hideInMenu: true
    }
  }
]

// 创建路由实例
const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    } else {
      return { top: 0 }
    }
  }
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()

  // 设置页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title} - GooseForum 管理后台`
  }

  // 检查是否需要认证
  if (to.meta.requiresAuth) {
    // 等待认证状态初始化完成
    if (authStore.loading) {
      await new Promise(resolve => {
        const unwatch = watch(() => authStore.loading, (loading) => {
          if (!loading) {
            unwatch()
            resolve(void 0)
          }
        })
      })
    }

    if (!authStore.isAuthenticated) {
      // 未登录，跳转到登录页
      next({
        path: '/admin/login',
        query: { redirect: to.fullPath }
      })
      return
    }

    // 检查是否需要管理员权限
    if (to.meta.requiresAdmin && !authStore.isAdmin) {
      // 权限不足，跳转到仪表盘
      console.warn('用户权限不足，跳转到仪表盘')
      next('/admin')
      return
    }
  }

  // 如果已登录且访问登录页，跳转到仪表盘
  if (to.path === '/admin/login' && authStore.isAuthenticated) {
    next('/admin')
    return
  }

  next()
})

// 路由错误处理
router.onError((error) => {
  console.error('路由错误:', error)
})

export default router