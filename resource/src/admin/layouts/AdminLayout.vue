<template>
  <div class="drawer lg:drawer-open">
    <!-- 移动端抽屉切换 -->
    <input id="drawer-toggle" type="checkbox" class="drawer-toggle" />

    <!-- 主内容区 -->
    <div class="drawer-content flex flex-col">
      <!-- 顶部导航栏 -->
      <div class="navbar bg-base-100 shadow-sm border-b border-base-300">
        <div class="flex-none lg:hidden">
          <label for="drawer-toggle" class="btn btn-square btn-ghost">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path>
            </svg>
          </label>
        </div>

        <div class="flex-1">
          <div class="breadcrumbs text-sm" v-if="breadcrumbs.length > 0">
            <ul>
              <li v-for="(crumb, index) in breadcrumbs" :key="index">
                <router-link v-if="crumb.path && index < breadcrumbs.length - 1" :to="crumb.path"
                  class="link link-hover">
                  {{ crumb.name }}
                </router-link>
                <span v-else>{{ crumb.name }}</span>
              </li>
            </ul>
          </div>
          <span v-else>{{ pageTitle }}</span>
        </div>

        <div class="flex-none">
          <!-- 主题切换 -->
          <label class="swap swap-rotate btn btn-ghost btn-circle">
            <input type="checkbox" class="theme-controller" value="dark" @change="toggleTheme" />
            <svg class="swap-off fill-current w-6 h-6" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
              <path
                d="M5.64,17l-.71.71a1,1,0,0,0,0,1.41,1,1,0,0,0,1.41,0l.71-.71A1,1,0,0,0,5.64,17ZM5,12a1,1,0,0,0-1-1H3a1,1,0,0,0,0,2H4A1,1,0,0,0,5,12Zm7-7a1,1,0,0,0,1-1V3a1,1,0,0,0-2,0V4A1,1,0,0,0,12,5ZM5.64,7.05a1,1,0,0,0,.7.29,1,1,0,0,0,.71-.29,1,1,0,0,0,0-1.41l-.71-.71A1,1,0,0,0,4.93,6.34Zm12,.29a1,1,0,0,0,.7-.29l.71-.71a1,1,0,1,0-1.41-1.41L17,5.64a1,1,0,0,0,0,1.41A1,1,0,0,0,17.66,7.34ZM21,11H20a1,1,0,0,0,0,2h1a1,1,0,0,0,0-2Zm-9,8a1,1,0,0,0-1,1v1a1,1,0,0,0,2,0V20A1,1,0,0,0,12,19ZM18.36,17A1,1,0,0,0,17,18.36l.71.71a1,1,0,0,0,1.41,0,1,1,0,0,0,0-1.41ZM12,6.5A5.5,5.5,0,1,0,17.5,12,5.51,5.51,0,0,0,12,6.5Zm0,9A3.5,3.5,0,1,1,15.5,12,3.5,3.5,0,0,1,12,15.5Z" />
            </svg>
            <svg class="swap-on fill-current w-6 h-6" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
              <path
                d="M21.64,13a1,1,0,0,0-1.05-.14,8.05,8.05,0,0,1-3.37.73A8.15,8.15,0,0,1,9.08,5.49a8.59,8.59,0,0,1,.25-2A1,1,0,0,0,8,2.36,10.14,10.14,0,1,0,22,14.05,1,1,0,0,0,21.64,13Zm-9.5,6.69A8.14,8.14,0,0,1,7.08,5.22v.27A10.15,10.15,0,0,0,17.22,15.63a9.79,9.79,0,0,0,2.1-.22A8.11,8.11,0,0,1,12.14,19.73Z" />
            </svg>
          </label>

          <!-- 用户菜单 -->
          <div class="dropdown dropdown-end">
            <div tabindex="0" role="button" class="btn btn-ghost btn-circle avatar">
              <div class="w-8 rounded-full">
                <img :src="userStore.user?.avatar || '/static/pic/default-avatar.png'"
                  :alt="userStore.user?.username" />
              </div>
            </div>
            <ul tabindex="0" class="menu menu-sm dropdown-content mt-3 z-[1] p-2 shadow bg-base-100 rounded-box w-52">
              <li><a class="justify-between">{{ userStore.user?.username }}<span class="badge">管理员</span></a></li>
              <li><a @click="logout">退出登录</a></li>
            </ul>
          </div>
        </div>
      </div>

      <!-- 主内容 -->
      <main class="flex-1 p-6">
        <router-view />
      </main>
    </div>

    <!-- 侧边栏 -->
    <div class="drawer-side">
      <label for="drawer-toggle" aria-label="close sidebar" class="drawer-overlay"></label>
      <aside :class="[
        'min-h-full bg-base-100 border-r border-base-300 transition-all duration-300',
        isCollapsed ? 'w-12' : 'w-64'
      ]">
        <!-- Logo 和折叠按钮 -->
        <div class="p-2 border-b border-base-300 flex items-center h-16"
          :class="isCollapsed ? 'justify-center' : 'justify-between'">
          <div class="transition-all duration-300 overflow-hidden"
            :class="isCollapsed ? 'w-0 opacity-0' : 'w-auto opacity-100'">
            <h2 class="text-xl font-bold text-primary whitespace-nowrap">GooseForum</h2>
            <p class="text-sm text-base-content/70 whitespace-nowrap">管理后台</p>
          </div>
          <button @click="toggleSidebar" class="btn btn-ghost btn-sm btn-circle hidden lg:flex transition-transform duration-300"
            :title="isCollapsed ? '展开侧边栏' : '折叠侧边栏'"
            :class="isCollapsed ? 'rotate-180' : ''">
            <ChevronLeftIcon class="w-4 h-4" />
          </button>
        </div>

        <!-- 菜单 -->
        <ul class="menu space-y-2 w-full" :class="isCollapsed ? 'p-1 pt-4' : 'p-1 pt-4'">
          <li v-for="item in menuItems" :key="item.key">
            <router-link :to="item.path" :class="[
              'flex items-center p-2 rounded-lg',
              {
                'bg-primary text-primary-content': $route.path === item.path,
                'hover:bg-base-200': $route.path !== item.path,
                'justify-center': isCollapsed
              }
            ]" :title="isCollapsed ? item.label : ''">
              <component :is="item.icon" class="w-5 h-5 flex-shrink-0" />
              <span v-if="!isCollapsed" class="truncate">{{ item.label }}</span>
            </router-link>
          </li>
        </ul>
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import {
  HomeIcon,
  UsersIcon,
  DocumentTextIcon,
  TagIcon,
  CogIcon,
  LinkIcon,
  TicketIcon,
  ShieldCheckIcon,
  GiftIcon,
  ChevronLeftIcon,
  ChevronRightIcon
} from '@heroicons/vue/24/outline'

const route = useRoute()
const router = useRouter()
const userStore = useAuthStore()

// 侧边栏折叠状态
const isCollapsed = ref(false)

// 切换侧边栏折叠状态
const toggleSidebar = () => {
  isCollapsed.value = !isCollapsed.value
  localStorage.setItem('sidebar-collapsed', isCollapsed.value.toString())
}

// 菜单项
const menuItems = ref([
  {
    key: 'dashboard',
    label: '仪表盘-todo',
    path: '/admin',
    icon: HomeIcon
  },
  {
    key: 'users',
    label: '用户管理',
    path: '/admin/users',
    icon: UsersIcon
  },
  {
    key: 'roles',
    label: '角色管理 *',
    path: '/admin/roles',
    icon: ShieldCheckIcon
  },
  {
    key: 'posts',
    label: '帖子管理 *',
    path: '/admin/posts',
    icon: DocumentTextIcon
  },
  {
    key: 'categories',
    label: '分类管理 *',
    path: '/admin/categories',
    icon: TagIcon
  },
  {
    key: 'friend-links',
    label: '友情链接 *',
    path: '/admin/friend-links',
    icon: LinkIcon
  },
  {
    key: 'sponsors',
    label: '赞助管理-todo',
    path: '/admin/sponsors',
    icon: GiftIcon
  },
  {
    key: 'tickets',
    label: '工单管理-todo',
    path: '/admin/tickets',
    icon: TicketIcon
  },
  {
    key: 'settings',
    label: '系统设置-todo',
    path: '/admin/settings',
    icon: CogIcon
  }
])

// 页面标题
const pageTitle = computed(() => {
  const currentItem = menuItems.value.find(item => item.path === route.path)
  return currentItem?.label || '管理后台'
})

// 面包屑导航
const breadcrumbs = computed(() => {
  const crumbs = [{ name: '首页', path: '/admin' }]

  if (route.path !== '/admin') {
    const currentItem = menuItems.value.find(item => item.path === route.path)
    if (currentItem) {
      crumbs.push({ name: currentItem.label, path: route.path })
    }
  }

  return crumbs
})

// 主题切换
const toggleTheme = (event: Event) => {
  const target = event.target as HTMLInputElement
  const theme = target.checked ? 'dark' : 'light'
  document.documentElement.setAttribute('data-theme', theme)
  localStorage.setItem('theme', theme)
}

// 退出登录
const logout = async () => {
  await userStore.logout()
  router.push('/admin/login')
}

// 初始化主题和侧边栏状态
onMounted(() => {
  const savedTheme = localStorage.getItem('theme') || 'light'
  document.documentElement.setAttribute('data-theme', savedTheme)

  const themeController = document.querySelector('.theme-controller') as HTMLInputElement
  if (themeController) {
    themeController.checked = savedTheme === 'dark'
  }

  // 恢复侧边栏折叠状态
  const savedCollapsed = localStorage.getItem('sidebar-collapsed')
  if (savedCollapsed) {
    isCollapsed.value = savedCollapsed === 'true'
  }
})
</script>