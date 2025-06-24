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
  PersonOutline,
  LinkOutline,
  TicketOutline,
  LayersOutline
} from '@vicons/ionicons5'
import {
  NLayout,
  NLayoutSider,
  NLayoutHeader,
  NLayoutContent,
  NLayoutFooter,
  NScrollbar ,
  NFlex,
  NMenu,
  NBreadcrumb,
  NBreadcrumbItem,
  NDropdown,
  NButton,
  NIcon,
  NCard
} from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import {useUserStore} from "@/admin/stores/auth.ts";

const router = useRouter()
const route = useRoute()
const collapsed = ref(false)

// 计算主内容区样式
const contentStyle = computed(() => {
  return {
    transition: 'padding-left 0.2s cubic-bezier(.4,0,.2,1)',
    paddingLeft: collapsed.value ? '64px' : '240px',
    minHeight: '100vh',
    background: '#f5f6fa'
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
  },
  {
    label: '友情链接',
    key: 'friendLinks',
    icon: () => h(LinkOutline), // 需要导入LinkOutline图标
    path: '/admin/friend-links'
  },
  {
    label: '赞助商管理',
    key: 'sponsorManager',
    icon: () => h(LinkOutline), // 需要导入LinkOutline图标
    path: '/admin/sponsor-manager'
  },
  {
    label: '外部工单',
    key: 'externalTickets',
    icon: () => h(TicketOutline), // 需要导入TicketOutline图标
    path: '/admin/external-tickets'
  },
  {
    label: 'Footer管理',
    key: 'footerManagement',
    icon: () => h(LayersOutline),
    path: '/admin/footer-management'
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
    useUserStore().handleLogout()
    // router.push('/login')
  } else if (key === 'profile') {
    // 跳转到个人信息页
    // router.push('/profile')
  }
}
</script>

<template>
  <n-layout style="height: 100vh; overflow: hidden;">
    <!-- 顶部导航栏 -->
    <n-layout-header bordered style="z-index: 999; width: 100%; height: 64px;">
      <n-flex justify="space-between" align="center" style="height: 100%; padding: 0 24px; background: #fff;">
        <n-flex align="center" gap="16">
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
        </n-flex>
        <n-flex align="center">
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
        </n-flex>
      </n-flex>
    </n-layout-header>
    <!-- 主体部分：侧边栏+内容区 -->
    <n-layout has-sider style="height: calc(100vh - 64px);">
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
      >
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
      <!-- 内容区 -->
      <n-layout-content
       
      >
      <n-scrollbar style="height: 100%;">
        <n-card :bordered="false" style="height: 100%; min-height: 0; display: flex; flex-direction: column;">
          
            <router-view style="flex: 1 1 auto;"/>
          
        </n-card>
      </n-scrollbar>
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<style scoped>

.logo {
  height: 32px;
}

.logo-small {
  height: 32px;
  width: 32px;
}

.footer-container {
  text-align: center;
  padding: 16px 0;
  color: rgba(0, 0, 0, 0.45);
}
</style>
