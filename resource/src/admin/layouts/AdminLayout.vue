<template>
  <div class="min-h-screen ">
    <!-- 顶部导航栏 - 100%宽度 -->
    <header
        class="navbar bg-base-100 shadow-sm border-b border-base-300 fixed top-0 left-0 right-0 z-50 rounded-bl-lg rounded-br-lg">
      <div class="flex-none">
        <!-- 移动端菜单按钮 -->
        <label for="drawer-toggle" class="btn btn-square btn-ghost lg:hidden">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path>
          </svg>
        </label>
      </div>

      <!-- Logo 区域 - 独立容器 -->
      <div class="flex-none hidden lg:flex items-center mr-4">
        <div class="transition-all duration-300 ease-in-out overflow-hidden">
          <a href="/" class="font-normal text-lg mr-4 hover:text-primary">GooseForum</a>
        </div>
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
        <span v-else class="text-lg font-normal">{{ pageTitle }}</span>
      </div>

      <div class="flex-none gap-2">
        <!-- 主题切换 -->
        <label class="swap swap-rotate btn btn-ghost btn-circle">
          <input type="checkbox" class="theme-controller" value="dark" @change="toggleTheme"/>
          <svg class="swap-off fill-current w-5 h-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
            <path
                d="M5.64,17l-.71.71a1,1,0,0,0,0,1.41,1,1,0,0,0,1.41,0l.71-.71A1,1,0,0,0,5.64,17ZM5,12a1,1,0,0,0-1-1H3a1,1,0,0,0,0,2H4A1,1,0,0,0,5,12Zm7-7a1,1,0,0,0,1-1V3a1,1,0,0,0-2,0V4A1,1,0,0,0,12,5ZM5.64,7.05a1,1,0,0,0,.7.29,1,1,0,0,0,.71-.29,1,1,0,0,0,0-1.41l-.71-.71A1,1,0,0,0,4.93,6.34Zm12,.29a1,1,0,0,0,.7-.29l.71-.71a1,1,0,1,0-1.41-1.41L17,5.64a1,1,0,0,0,0,1.41A1,1,0,0,0,17.66,7.34ZM21,11H20a1,1,0,0,0,0,2h1a1,1,0,0,0,0-2Zm-9,8a1,1,0,0,0-1,1v1a1,1,0,0,0,2,0V20A1,1,0,0,0,12,19ZM18.36,17A1,1,0,0,0,17,18.36l.71.71a1,1,0,0,0,1.41,0,1,1,0,0,0,0-1.41ZM12,6.5A5.5,5.5,0,1,0,17.5,12,5.51,5.51,0,0,0,12,6.5Zm0,9A3.5,3.5,0,1,1,15.5,12,3.5,3.5,0,0,1,12,15.5Z"/>
          </svg>
          <svg class="swap-on fill-current w-5 h-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
            <path
                d="M21.64,13a1,1,0,0,0-1.05-.14,8.05,8.05,0,0,1-3.37.73A8.15,8.15,0,0,1,9.08,5.49a8.59,8.59,0,0,1,.25-2A1,1,0,0,0,8,2.36,10.14,10.14,0,1,0,22,14.05,1,1,0,0,0,21.64,13Zm-9.5,6.69A8.14,8.14,0,0,1,7.08,5.22v.27A10.15,10.15,0,0,0,17.22,15.63a9.79,9.79,0,0,0,2.1-.22A8.11,8.11,0,0,1,12.14,19.73Z"/>
          </svg>
        </label>

        <!-- 用户菜单 -->
        <div class="dropdown dropdown-end">
          <div tabindex="0" role="button" class="btn btn-ghost btn-circle avatar">
            <div class="w-8 rounded-full">
              <img :src="userStore.user?.avatar || '/static/pic/default-avatar.webp'"
                   :alt="userStore.user?.username"/>
            </div>
          </div>
          <ul tabindex="0" class="menu menu-sm dropdown-content mt-3 z-[1] p-2 shadow bg-base-100 rounded-box w-52">
            <li><a class="justify-between">{{ userStore.user?.username }}<span class="badge">管理员</span></a></li>
            <li><a @click="logout">退出登录</a></li>
          </ul>
        </div>
      </div>
    </header>

    <!-- 主体布局 -->
    <div class="drawer lg:drawer-open pt-16">
      <!-- 移动端抽屉切换 -->
      <input id="drawer-toggle" type="checkbox" class="drawer-toggle"/>

      <!-- 主内容区 -->
      <div class="drawer-content flex flex-col lg:ml-48">
        <main class="flex-1 p-6 min-h-screen">
          <router-view/>
        </main>
      </div>

      <!-- 侧边栏 -->
      <div class="drawer-side z-30">
        <label for="drawer-toggle" aria-label="close sidebar" class="drawer-overlay lg:hidden"></label>
        <aside class="w-48 min-h-full bg-base-100 border-r border-base-300 fixed top-0 left-0 bottom-0 z-40">
          <!-- 菜单 -->
          <nav class="flex-1 scrollbar-ultra-thin pt-20">
            <ul class="menu bg-base-100 w-full h-full">
              <template v-for="item in menuItems" :key="item.key">
                <!-- 普通菜单项 -->
                <li v-if="!item.children">
                  <router-link :to="item.path" :class="{
                    'menu-active': isPathActive(item.path)
                  }">
                    <component :is="item.icon" class="w-5 h-5"/>
                    <span>{{ item.label }}</span>
                  </router-link>
                </li>

                <!-- 带子菜单的菜单项 -->
                <li v-else>
                  <details :open="isParentActive(item)">
                    <summary>
                      <component :is="item.icon" class="w-5 h-5"/>
                      <span>{{ item.label }}</span>
                    </summary>
                    <ul>
                      <li v-for="child in item.children" :key="child.key">
                        <router-link :to="child.path" :class="[
                          {
                            'menu-active': isPathActive(child.path)
                          }
                        ]">
                          <component :is="child.icon" class="w-4 h-4"/>
                          <span>{{ child.label }}</span>
                        </router-link>
                      </li>
                    </ul>
                  </details>
                </li>
              </template>
            </ul>
          </nav>
        </aside>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type {Component} from 'vue'
import {computed, onMounted, ref} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useAuthStore} from '../stores/auth'
import {
  BookOpenIcon,
  CogIcon,
  DocumentDuplicateIcon,
  DocumentTextIcon,
  FolderIcon,
  GiftIcon,
  GlobeAltIcon,
  HomeIcon,
  LinkIcon,
  ShieldCheckIcon,
  TagIcon,
  TicketIcon,
  UsersIcon
} from '@heroicons/vue/24/outline'

// 菜单项类型定义
interface MenuItem {
  key: string
  label: string
  path?: string
  icon: Component
  children?: MenuItem[]
}

interface MenuItemResult {
  item: MenuItem
  parent: MenuItem | null
}

const route = useRoute()
const router = useRouter()
const userStore = useAuthStore()


// 路径匹配函数 - 精确匹配
const isPathActive = (itemPath: string) => {
  if (!itemPath) return false
  const currentPath = route.path.replace(/\/$/, '') || '/'
  const targetPath = itemPath.replace(/\/$/, '') || '/'
  return currentPath === targetPath
}

// 判断父级菜单是否激活（子菜单中有激活项）
const isParentActive = (parentItem: MenuItem) => {
  if (!parentItem.children) return false
  return parentItem.children.some((child: MenuItem) => child.path && isPathActive(child.path))
}


// 菜单项
const menuItems = ref<MenuItem[]>([
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
    label: '角色管理',
    path: '/admin/roles',
    icon: ShieldCheckIcon
  },
  {
    key: 'posts',
    label: '帖子管理',
    path: '/admin/posts',
    icon: DocumentTextIcon
  },
  {
    key: 'categories',
    label: '分类管理',
    path: '/admin/categories',
    icon: TagIcon
  },
  {
    key: 'friend-links',
    label: '友情链接',
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
    key: 'docs',
    label: '文档管理',
    icon: FolderIcon,
    children: [
      {
        key: 'docs-projects',
        label: '项目管理',
        path: '/admin/docs/projects',
        icon: BookOpenIcon
      },
      {
        key: 'docs-versions',
        label: '版本管理',
        path: '/admin/docs/versions',
        icon: DocumentDuplicateIcon
      }
    ]
  },
  {
    key: 'web-settings',
    label: '网页设置',
    path: '/admin/web-settings',
    icon: GlobeAltIcon
  },
  {
    key: 'settings',
    label: '系统设置-todo',
    path: '/admin/settings',
    icon: CogIcon
  },
])

// 查找当前菜单项（支持嵌套菜单）
const findCurrentMenuItem = (path: string): MenuItemResult | null => {
  for (const item of menuItems.value) {
    if (item.path === path) {
      return {item, parent: null}
    }
    if (item.children) {
      for (const child of item.children) {
        if (child.path === path) {
          return {item: child, parent: item}
        }
      }
    }
  }
  return null
}

// 页面标题
const pageTitle = computed(() => {
  const result = findCurrentMenuItem(route.path)
  return result?.item?.label || '管理后台'
})

// 面包屑导航
const breadcrumbs = computed(() => {
  const crumbs = [{name: '首页', path: '/admin'}]

  if (route.path !== '/admin') {
    const result = findCurrentMenuItem(route.path)
    if (result) {
      // 如果有父级菜单，先添加父级
      if (result.parent) {
        crumbs.push({name: result.parent.label, path: null})
      }
      // 添加当前页面
      crumbs.push({name: result.item.label, path: route.path})
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


// 组件挂载时初始化
onMounted(() => {
  // 从localStorage恢复主题
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme) {
    document.documentElement.setAttribute('data-theme', savedTheme)
    const themeController = document.querySelector('.theme-controller') as HTMLInputElement
    if (themeController) {
      themeController.checked = savedTheme === 'dark'
    }
  }
})

</script>