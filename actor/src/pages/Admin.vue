<script setup>
import {h, ref} from 'vue'
import {RouterLink} from 'vue-router'
import {Flash, GridOutline, MenuOutline} from '@vicons/ionicons5'
import {
  NButton,
  NDrawer,
  NDrawerContent,
  NFlex,
  NIcon,
  NLayout,
  NLayoutHeader,
  NLayoutSider,
  NMenu,
  NSpace
} from 'naive-ui'
import {managerRouter} from "@/route/routes";
import {useIsMobile, useIsTablet} from "@/utils/composables";
import UpdateTheme from "@/components/UpdateTheme.vue";
import UserInfoMenu from "@/components/UserInfoMenu.vue";
import HeaderLogo from "@/components/HeaderLogo.vue"; // Import the theme store


function renderIcon(icon) {
  return () => h(NIcon, null, {default: () => h(icon)})
}

function buildCompletePath(parentPath, currentPath) {
  parentPath = parentPath === undefined ? "" : parentPath
  return parentPath.endsWith('/') ? parentPath + currentPath : `${parentPath}/${currentPath}`;
}

function buildMenuItem(route, parentPath) {
  const {path, showName, name, icon} = route;
  let currentPath = buildCompletePath(parentPath, path)
  let label = () => h(RouterLink, {to: {path: currentPath}}, () => showName === undefined ? name : showName);
  let iconComponent = icon !== undefined ? renderIcon(icon) : renderIcon(Flash);
  return {
    label: label,
    key: currentPath,
    icon: iconComponent,
  }
}

function buildMenuOptionV2(routeList, parentPath) {
  if (!routeList) {
    return [];
  }
  return routeList.filter(item => item.belongMenu).map(childrenRoute => {
    let menuItem = buildMenuItem(childrenRoute, parentPath)
    let childrenList = buildMenuOptionV2(childrenRoute.children, buildCompletePath(parentPath, childrenRoute.path))
    menuItem.children = (childrenList && childrenList.length > 0) ? childrenList : undefined
    return menuItem;
  })
}

let menuOptions = buildMenuOptionV2(managerRouter.children, managerRouter.path)
const options = menuOptions
const collapsed = ref(false)
const isMobile = useIsMobile()
const active = ref(false);
const placement = ref("left");
const activate = (place) => {
  active.value = true;
  placement.value = place;
};

function handleSelect() {
  active.value = false
}


const activeKey = ref('home')
const menuOptions2 = [
  {
    label: () => h(RouterLink, {to: {path: "/home/index",}},
        {
          default: () => "首页"
        }
    ),
    key: "index",
    children: null,
  }, {
    label: () => h(RouterLink, {to: {path: "/home/bbs",}},
        {
          default: () => "Forum"
        }
    ),
    key: "bbs",
    children: null,
  }
]

let isTablet = useIsTablet()
const showDrawer = ref(false)
const showAdminDrawer = ref(false)


</script>
<template>
  <n-layout position="absolute" content-style="width: 100%;height: 100%;">
    <n-layout-header
        position="absolute"
        style="height: 56px;"
        bordered
    >
      <n-flex align="center" style="height: 100%; padding: 0 16px" justify="space-between">
        <!-- Logo and Navigation Area -->
        <n-flex align="center" style="height: 100%">
          <header-logo/>
          <!-- Desktop Navigation Menu -->
          <div v-if="!isTablet && !isMobile" class="menu-container">
            <n-menu v-model:value="activeKey" mode="horizontal" :options="menuOptions2"/>
          </div>
        </n-flex>
        <n-flex v-if="!isTablet && !isMobile" align="center" class="action-buttons">
          <update-theme/>
          <user-info-menu/>
        </n-flex>

        <n-flex v-if="isTablet || isMobile" align="center" class="action-buttons">
          <update-theme/>
          <!-- Navigation Menu Button -->
          <n-button quaternary @click="showDrawer = true">
            <n-icon size="20">
              <menu-outline/>
            </n-icon>
          </n-button>
          <!-- Admin Menu Button -->
          <n-button quaternary @click="showAdminDrawer = true">
            <n-icon size="20">
              <grid-outline/>
            </n-icon>
          </n-button>
        </n-flex>
      </n-flex>
    </n-layout-header>
    <n-layout position="absolute" style="top: 56px; bottom: 0;" has-sider>
      <n-layout-sider
          collapse-mode="width"
          :collapsed-width="64"
          @collapse="collapsed = true"
          @expand="collapsed = false"
          :collapsed="collapsed"
          content-style="padding-top: 6px;"
          :native-scrollbar="false" bordered
          width="140"
          show-trigger
          v-show="!isMobile"
      >
        <n-menu
            :options="menuOptions"
            :collapsed="collapsed"
            :collapsed-width="64"
            :collapsed-icon-size="22"
            :icon-size="22"
            :indent="24"
        />
      </n-layout-sider>

      <n-layout content-style="padding: 16px;width: 100%;height: 100%;" :native-scrollbar="false">
        <router-view></router-view>
      </n-layout>
    </n-layout>
  </n-layout>

  <!-- Mobile Navigation Drawer -->
  <n-drawer v-model:show="showDrawer" :width="280">
    <n-drawer-content title="导航菜单" closable>
      <n-space vertical size="large">
        <n-menu
            :options="menuOptions2"

        />
      </n-space>
    </n-drawer-content>
  </n-drawer>

  <!-- Mobile Admin Menu Drawer -->
  <n-drawer v-model:show="showAdminDrawer" :width="280">
    <n-drawer-content title="管理菜单" closable>
      <n-menu
          :options="menuOptions"

      />
    </n-drawer-content>
  </n-drawer>
</template>

<style scoped>

.menu-container {
  height: 100%;
}

.action-buttons {
  height: 100%;
  gap: 8px;
}

.menu-container :deep(.n-menu.n-menu--horizontal) {
  height: 100%;
  border: none;
}

.menu-container :deep(.n-menu-item) {
  height: 100%;
  display: flex;
  align-items: center;
}

.menu-container :deep(.n-menu-item a) {
  display: flex;
  align-items: center;
  height: 100%;
}

</style>
