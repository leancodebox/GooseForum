<script setup>
import {h, ref} from 'vue'
import {RouterLink} from 'vue-router'
import {Flash} from '@vicons/ionicons5'
import {
  NButton,
  NDrawer,
  NDrawerContent,
  NDropdown,
  NFlex,
  NIcon,
  NLayout,
  NLayoutHeader,
  NLayoutSider,
  NMenu,
  NSpace,
  NText,
  NButtonGroup
} from 'naive-ui'
import {managerRouter} from "@/route/routes";
import {useIsMobile} from "@/utils/composables";

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
          default: () => "BBS"
        }
    ),
    key: "bbs",
    children: null,
  }
]
</script>
<template>
  <n-layout position="absolute" content-style="width: 100%;height: 100%;">
    <n-layout-header
        position="absolute"
        style="height: 56px; padding: 8px;"
        bordered
    >
      <n-flex justify="space-between" align="center">
        <n-text tag="div" class="ui-logo" :depth="1" >
          <img alt="" src="/quote-left.png"/>
          <span>GooseForum</span>
          <n-menu v-model:value="activeKey" mode="horizontal" :options="menuOptions2"/>
        </n-text>
        <n-flex align="center"  style="padding-right: 24px;">>
          <n-button-group>
            <n-dropdown trigger="hover" :options="options">
              <n-button>旺财</n-button>
            </n-dropdown>
            <n-button @click="activate('left')">阿福</n-button>
          </n-button-group>
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
        <n-menu :options="menuOptions"
                :collapsed="collapsed"
                :collapsed-width="64"
                :collapsed-icon-size="22"
                :icon-size="22"
                :indent="22"
        />

      </n-layout-sider>
      <n-layout content-style="padding: 16px;width: 100%;height: 100%;"   :native-scrollbar="false">
        <router-view></router-view>
      </n-layout>
    </n-layout>
  </n-layout>
  <n-drawer v-model:show="active" :width="320" :placement="placement">
    <n-drawer-content title="斯通纳" :body-content-style="{padding:'12px 0 0 0 '}">
      <n-menu :options="menuOptions"
              :collapsed="collapsed"
              :collapsed-width="64"
              :collapsed-icon-size="22"
              :indent="16"
              @select="handleSelect"
      />

    </n-drawer-content>
  </n-drawer>
</template>

<style>

.ui-logo {
  cursor: pointer;
  display: flex;
  align-items: center;
  font-size: 18px;
}

.ui-logo > img {
  margin-right: 12px;
  height: 32px;
  width: 32px;
}
</style>
