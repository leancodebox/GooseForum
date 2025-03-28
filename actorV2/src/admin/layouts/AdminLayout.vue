<script setup lang="ts">
import { h, ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  MenuOutline,
  HomeOutline,
  PeopleOutline,
  DocumentTextOutline,
  GridOutline,
  SettingsOutline,
  LogOutOutline,
  PersonOutline
} from '@vicons/ionicons5'
import {
  NLayout,
  NLayoutSider,
  NLayoutHeader,
  NLayoutContent,
  NLayoutFooter,
  NMenu,
  NBreadcrumb,
  NBreadcrumbItem,
  NDropdown,
  NButton,
  NIcon,
  NCard
} from 'naive-ui'
import type { MenuOption } from 'naive-ui'

const router = useRouter()
const route = useRoute()
const collapsed = ref(false)

// 计算主内容区样式
const contentStyle = computed(() => {
  return {
    paddingLeft: collapsed.value ? '64px' : '240px',
  }
})

// 菜单渲染函数
function renderMenuLabel(option: MenuOption) {
  return option.label as string
}

function renderMenuIcon(option: MenuOption) {
  return h(NIcon, null, { default: () => h(option.icon as any) })
}

// 菜单选项
const menuOptions: MenuOption[] = [
  {
    label: '仪表盘',
    key: 'dashboard',
    icon: () => h(HomeOutline),
    path: '/admin/'
  },
  {
    label: '用户管理',
    key: 'users',
    icon: () => h(PeopleOutline),
    path: '/admin/users'
  },
  {
    label: '角色管理',
    key: 'roles',
    icon: () => h(PeopleOutline),
    path: '/admin/roles'
  },
  {
    label: '帖子管理',
    key: 'posts',
    icon: () => h(DocumentTextOutline),
    path: '/admin/posts'
  },
  {
    label: '分类管理',
    key: 'categories',
    icon: () => h(GridOutline),
    path: '/admin/categories'
  },
  {
    label: '系统设置',
    key: 'settings',
    icon: () => h(SettingsOutline),
    path: '/admin/settings'
  }
]

// 用户下拉菜单选项
const userOptions = [
  {
    label: '个人信息',
    key: 'profile',
    icon: () => h(NIcon, null, { default: () => h(PersonOutline) })
  },
  {
    label: '退出登录',
    key: 'logout',
    icon: () => h(NIcon, null, { default: () => h(LogOutOutline) })
  }
]

// 当前激活的菜单项
const activeKey = computed(() => {
  const path = route.path
  const menuItem = menuOptions.find(item => path === item.path)
  return menuItem ? menuItem.key : 'dashboard'
})

// 当前页面标题
const currentPageTitle = computed(() => {
  return route.meta.title || '仪表盘'
})

// 菜单点击处理
const handleMenuUpdate = (key: string) => {
  const menuItem = menuOptions.find(item => item.key === key)
  if (menuItem && menuItem.path) {
    router.push(menuItem.path)
  }
}

// 用户菜单选择处理
const handleUserSelect = (key: string) => {
  if (key === 'logout') {
    // 退出登录逻辑
    // localStorage.removeItem('admin_token')
    router.push('/login')
  } else if (key === 'profile') {
    // 跳转到个人信息页
    // router.push('/profile')
  }
}
</script>

<template>
  <n-layout has-sider position="absolute">
    <!-- 侧边栏 -->
    <n-layout-sider
      bordered
      collapse-mode="width"
      :collapsed-width="64"
      :width="240"
      :collapsed="collapsed"
      show-trigger
      @collapse="collapsed = true"
      @expand="collapsed = false"
      :native-scrollbar="false"
      position="absolute"
    >
      <div class="logo-container">
        <img v-if="collapsed"  src="/favicon.ico" alt="Logo" class="logo-small" />
        <img v-else src="/favicon.ico" alt="Logo" class="logo" />
      </div>
      <n-menu
        :collapsed="collapsed"
        :collapsed-width="64"
        :collapsed-icon-size="22"
        :options="menuOptions"
        :render-label="renderMenuLabel"
        :render-icon="renderMenuIcon"
        :value="activeKey"
        @update:value="handleMenuUpdate"
      />
    </n-layout-sider>

    <!-- 主内容区 -->
    <n-layout :style="contentStyle">
      <!-- 顶部导航栏 -->
      <n-layout-header bordered position="absolute" style="z-index: 999; width: 100%;">
        <div class="header-container">
          <div class="header-left">
            <n-button quaternary circle @click="collapsed = !collapsed">
              <template #icon>
                <n-icon size="18">
                  <MenuOutline />
                </n-icon>
              </template>
            </n-button>
            <n-breadcrumb>
              <n-breadcrumb-item>GooseForum</n-breadcrumb-item>
              <n-breadcrumb-item>管理系统</n-breadcrumb-item>
              <n-breadcrumb-item>{{ currentPageTitle }}</n-breadcrumb-item>
            </n-breadcrumb>
          </div>
          <div class="header-right">
            <n-dropdown :options="userOptions" @select="handleUserSelect">
              <n-button text>
                管理员
                <template #icon>
                  <n-icon>
                    <PersonOutline />
                  </n-icon>
                </template>
              </n-button>
            </n-dropdown>
          </div>
        </div>
      </n-layout-header>

      <!-- 内容区 -->
      <n-layout-content content-style="padding: 24px;" :style="{ marginTop: '64px' }">
        <n-card>
          <router-view />
        </n-card>
      </n-layout-content>

    </n-layout>
  </n-layout>
</template>



<style scoped>
.logo-container {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid rgba(0, 0, 0, 0.08);
}

.logo {
  height: 32px;
}

.logo-small {
  height: 32px;
  width: 32px;
}

.header-container {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  background-color: #fff;
  box-sizing: border-box;
  width: 100%;
  right: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-right {
  display: flex;
  align-items: center;
  margin-right: 24px;
}

.footer-container {
  text-align: center;
  padding: 16px 0;
  color: rgba(0, 0, 0, 0.45);
}
</style>
