<template>
  <div class="drawer lg:drawer-open font-sans text-base-content">
    <input id="admin-drawer" type="checkbox" class="drawer-toggle" ref="drawerToggle" />
    
    <!-- Drawer Content -->
    <div class="drawer-content flex flex-col min-h-screen bg-base-200">
      <!-- Navbar -->
      <div class="navbar bg-base-100 sticky top-0 z-30 shadow-sm px-4 sm:px-6 min-h-[3rem] h-14">
        <div class="flex-none lg:hidden mr-2">
          <label for="admin-drawer" aria-label="open sidebar" class="btn btn-square btn-ghost">
            <component :is="Bars3Icon" class="w-6 h-6" />
          </label>
        </div>
        
        <div class="flex-1 flex justify-between items-center gap-4">
           <div class="breadcrumbs text-sm text-base-content/60">
            <ul>
              <li v-for="(crumb, index) in breadcrumbs" :key="index">
                <router-link v-if="crumb.path && index < breadcrumbs.length - 1" :to="crumb.path" class="hover:text-primary transition-colors">
                  {{ crumb.name }}
                </router-link>
                <span v-else>{{ crumb.name }}</span>
              </li>
            </ul>
          </div>
          
          <!-- Breadcrumbs for Mobile/Small Screens -->
          <div class="text-sm breadcrumbs sm:hidden">
            <ul>
              <li>{{ pageTitle }}</li>
            </ul>
          </div>
        </div>
        
        <div class="flex-none gap-2">
          <!-- Theme Toggle -->
          <label class="swap swap-rotate btn btn-ghost btn-circle">
            <input type="checkbox" class="theme-controller" value="dark" @change="toggleTheme"/>
            <!-- sun icon -->
            <SunIcon class="swap-off w-5 h-5" />
            <!-- moon icon -->
            <MoonIcon class="swap-on w-5 h-5" />
          </label>
          
        </div>
      </div>

      <!-- Main Content Area -->
      <main class="flex-1 p-4 lg:p-6 max-w-[1600px] w-full mx-auto">
        <!-- Page Header -->
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-2" v-if="false">
          <h1 class="text-2xl font-normal">{{ pageTitle }}</h1>
          <div class="breadcrumbs text-sm text-base-content/60">
            <ul>
              <li v-for="(crumb, index) in breadcrumbs" :key="index">
                <router-link v-if="crumb.path && index < breadcrumbs.length - 1" :to="crumb.path" class="hover:text-primary transition-colors">
                  {{ crumb.name }}
                </router-link>
                <span v-else>{{ crumb.name }}</span>
              </li>
            </ul>
          </div>
          <!-- Right side actions can go here -->
        </div>

        <!-- Content -->
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <div :key="$route.path" class="w-full">
              <component :is="Component" />
            </div>
          </transition>
        </router-view>
      </main>
    </div>

    <!-- Sidebar -->
    <div class="drawer-side z-40">
      <label for="admin-drawer" aria-label="close sidebar" class="drawer-overlay"></label>
      <aside class="bg-base-100 min-h-screen w-64 flex flex-col border-r border-base-200">
        <!-- Logo -->
        <div class="h-14 flex items-center px-6 border-b border-base-200">
          <a href="/" class="flex items-center gap-3 text-xl font-normal text-primary tracking-wide">
            <div class="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center text-primary">
              <ChatBubbleLeftRightIcon class="w-5 h-5" />
            </div>
            GooseForum
          </a>
        </div>
        
        <!-- Menu -->
        <div class="flex-1 overflow-y-auto py-4 px-3">
          <SidebarMenu :menu-items="menuItems" />
        </div>
        
        <!-- User Profile -->
        <div class="p-4 border-t border-base-200 bg-base-100">
          <div class="dropdown dropdown-top w-full">
            <div tabindex="0" role="button" class="flex items-center gap-3 p-2 rounded-lg hover:bg-base-200 transition-colors w-full">
              <div class="avatar">
                <div class="w-10 h-10 rounded-full ring ring-primary ring-offset-base-100 ring-offset-2">
                  <img :src="userStore.user?.avatar || '/static/pic/default-avatar.webp'" :alt="userStore.user?.username" />
                </div>
              </div>
              <div class="flex-1 text-left overflow-hidden">
                <div class="font-normal truncate">{{ userStore.user?.username || 'User' }}</div>
                <div class="text-xs text-base-content/60 truncate">Administrator</div>
              </div>
              <ChevronDownIcon class="h-4 w-4 opacity-50" />
            </div>
            <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow-lg bg-base-100 rounded-box w-full mb-2 border border-base-200">
              <li><a href="/profile">个人中心</a></li>
              <li><a href="/settings">账户设置</a></li>
              <li class="mt-1 border-t border-base-200"></li>
              <li><a @click="logout" class="text-error">退出登录</a></li>
            </ul>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@admin/stores/auth.ts'
import {
  Bars3Icon,
  MagnifyingGlassIcon,
  SunIcon,
  MoonIcon,
  ChevronDownIcon,
  ChatBubbleLeftRightIcon,
} from '@heroicons/vue/24/outline'
import SidebarMenu from './components/SidebarMenu.vue'
import { menuItems, type MenuItemResult } from '@/admin/utils/menuConfig'

const route = useRoute()
const router = useRouter()
const userStore = useAuthStore()

const drawerToggle = ref<HTMLInputElement | null>(null)

// 查找当前菜单项（支持嵌套菜单）
const findCurrentMenuItem = (path: string): MenuItemResult | null => {
  for (const item of menuItems) {
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

// 监听路由变化，移动端自动关闭抽屉
watch(
  () => route.path,
  () => {
    if (drawerToggle.value && window.innerWidth < 1024) {
      drawerToggle.value.checked = false
    }
  }
)

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

<style scoped>
</style>
