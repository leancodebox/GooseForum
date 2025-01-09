<script setup>
import {
  NButton,
  NFlex,
  NIcon,
  NLayout,
  NLayoutFooter,
  NLayoutHeader,
  NMenu,
  NText,
  NDrawer,
  NDrawerContent,
  NSpace,
  NDivider,
  NBadge
} from 'naive-ui';
import { MenuOutline } from '@vicons/ionicons5'
import {useIsMobile, useIsTablet} from "@/utils/composables";
import {h, ref, onMounted, onUnmounted, watch,defineEmits} from "vue";
import {RouterLink, useRouter} from "vue-router";
import UserInfoCard from "@/components/UserInfoMenu.vue";
import { getUnreadCount } from '@/service/request'
import { useUserStore } from '@/modules/user'
import { useNotificationStore } from '@/modules/notification'
import { useThemeStore } from '@/modules/theme'; // Import the theme store

const router = useRouter()
const showDrawer = ref(false)

function renderIcon(icon) {
  return () => h(NIcon, null, {default: () => h(icon)});
}

const activeKey = ref('home')
const menuOptions = [
  {
    label: "首页",
    key: "index",
    path: "/home/index"
  },
  {
    label: "BBS",
    key: "bbs",
    path: "/home/bbs"
  },
  {
    label: "关于",
    key: "about",
    path: "/home/about"
  }
]

const notificationStore = useNotificationStore()

const userStore = useUserStore()

const actionOptions = [
  {
    label: "写文章",
    key: "write",
    path: "/home/bbs/articlesEdit"
  },
  {
    label: "消息中心",
    key: "notification",
    path: "/home/notificationCenter",
  }
]

let isTablet = useIsTablet()
let isMobile = useIsMobile()
let topHeight = ref('56px')

function handleMenuClick(path) {
  router.push(path)
  showDrawer.value = false
}

// 监听登录状态变化
watch(() => userStore.isLogin, (newVal) => {
  if (newVal) {
    notificationStore.startPolling()
  } else {
    notificationStore.stopPolling()
    notificationStore.resetUnreadCount()
  }
})

// 在组件挂载时启动轮询
onMounted(() => {
  if (userStore.isLogin) {
    notificationStore.startPolling()
  }
})

// 在组件卸载时停止轮询
onUnmounted(() => {
  notificationStore.stopPolling()
})

// 在组件卸载时清理
onUnmounted(() => {
  notificationStore.cleanup()
})

const themeStore = useThemeStore(); // Use the theme store

function toggleTheme() {
  themeStore.toggleTheme(); // Call the toggleTheme action from the store
}
</script>

<template>
  <n-layout-header
      position="absolute"
      :style="{height: topHeight}"
      bordered
  >
    <n-flex align="center" style="height: 100%; padding: 0 16px" justify="space-between">
      <!-- Logo和导航区域 -->
      <n-flex align="center" style="height: 100%">
        <n-text tag="div" class="ui-logo" :depth="1">
          <img alt="" src="/quote-left.png"/>
          <span>GooseForum</span>
        </n-text>

        <!-- Theme Toggle Button -->
        <n-button @click="toggleTheme">
          {{ themeStore.isDarkTheme ? 'Light Mode' : 'Dark Mode' }}
        </n-button>

        <!-- 桌面端导航菜单 -->
        <div v-if="!isTablet && !isMobile" class="menu-container">
          <n-menu
              v-model:value="activeKey"
              mode="horizontal"
              :options="menuOptions.map(item => ({
                label: () => h(RouterLink, {to: {path: item.path}}, {default: () => item.label}),
                key: item.key
              }))"
          />
        </div>
      </n-flex>

      <!-- 桌面端操作按钮 -->
      <n-flex v-if="!isTablet && !isMobile" align="center" class="action-buttons">
        <router-link v-for="item in actionOptions" :key="item.key" :to="item.path">
          <n-badge v-if="item.key === 'notification'"
                   :value="notificationStore.unreadCount"
                   :max="99"
                   processing
                   :offset="[-5,3]"
                   type="error"
          >
          <n-button >
            {{ item.label }}
          </n-button>
          </n-badge>
          <n-button v-else >
            {{ item.label }}
          </n-button>

        </router-link>
        <user-info-card />
      </n-flex>

      <!-- 移动端菜单按钮和用户信息 -->
      <n-flex v-else align="center" class="action-buttons">
        <user-info-card />
        <n-button quaternary @click="showDrawer = true">
          <n-icon size="20">
            <menu-outline />
          </n-icon>
        </n-button>
      </n-flex>
    </n-flex>
  </n-layout-header>

  <!-- 移动端抽屉菜单 -->
  <n-drawer v-model:show="showDrawer" :width="280">
    <n-drawer-content title="菜单" closable>
      <n-space vertical size="large">
        <!-- 导航菜单 -->
        <n-space vertical>
          <div class="drawer-section-title">导航</div>
          <n-button
              v-for="item in menuOptions"
              :key="item.key"
              quaternary
              block
              @click="handleMenuClick(item.path)"
          >
            {{ item.label }}
          </n-button>
        </n-space>

        <n-divider />

        <!-- 操作按钮 -->
        <n-space vertical>
          <div class="drawer-section-title">操作</div>
          <n-button
              v-for="item in actionOptions"
              :key="item.key"
              quaternary
              block
              @click="handleMenuClick(item.path)"
          >
            {{ item.label }}
            <n-badge v-if="item.key === 'notification' && item.badge > 0"
                    :value="item.badge"
                    :max="99"
                    processing
                    type="error"
            />
          </n-button>
        </n-space>
      </n-space>
    </n-drawer-content>
  </n-drawer>

  <!-- 其他内容保持不变 -->
  <n-layout
    class="n-layout-container"
    :style="{top: topHeight}"
    :native-scrollbar="true"
    :position="'absolute'"
    content-style="display: flex; flex-direction: column; min-height: calc(100vh - var(--header-height));"
  >
    <div style="flex: 1;">
      <router-view></router-view>
    </div>
    <n-layout-footer
        class="main-footer"
        bordered>
      <p>&copy; My Homepage 2023</p>
    </n-layout-footer>
  </n-layout>
</template>

<style scoped>
.ui-logo {
  cursor: pointer;
  display: flex;
  align-items: center;
  font-size: 18px;
  margin-right: 24px;
  height: 100%;
  white-space: nowrap;
}

.ui-logo > img {
  margin-right: 12px;
  height: 32px;
  width: 32px;
}

.menu-container {
  height: 100%;
}

.action-buttons {
  height: 100%;
  gap: 12px;
}

/* 调整菜单样式 */
:deep(.n-menu.n-menu--horizontal) {
  height: 100%;
  border: none;
  display: flex;
  align-items: center;
}

:deep(.n-menu-item) {
  height: 100%;
  display: flex;
  align-items: center;
}

:deep(.n-button.n-button--quaternary) {
  height: 34px;
  display: flex;
  align-items: center;
}

.drawer-section-title {
  font-size: 14px;
  color: var(--n-text-color-3);
  padding: 8px 0;
}

/* 移动端样式调整 */
@media (max-width: 800px) {
  .ui-logo {
    font-size: 16px;
    margin-right: 0;
  }

  .ui-logo > img {
    height: 24px;
    width: 24px;
  }
}

/* 添加 badge 样式调整 */
:deep(.n-badge) {
  margin-left: 4px;
}

.main-content {
  flex: 1;
  overflow: auto;
  padding: 0;
}

.n-layout-container {
  overflow: auto;
}

.main-footer {
  flex-shrink: 0;
  height: 100px;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #f0f3f5;
}
</style>
