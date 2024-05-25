<script setup>
import {h, ref} from 'vue'
import {RouterLink} from 'vue-router'
import {Flash} from '@vicons/ionicons5'
import {NDrawer, NDrawerContent, NButton, NDropdown, NIcon, NLayout, NLayoutHeader, NLayoutSider, NMenu} from 'naive-ui'
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
</script>
<template>
    <n-layout position="absolute">
        <n-layout-header style="height: 64px; padding: 18px;" bordered
                         position="absolute"
        >
            <n-layout style="float: right">
                <n-dropdown trigger="hover" :options="options">
                    <n-button>旺财</n-button>
                </n-dropdown>
                <n-button @click="activate('left')">阿福</n-button>
            </n-layout>

        </n-layout-header
        >
        <n-layout position="absolute" style="top: 64px; bottom: 0;" has-sider>

            <n-layout-sider
                    collapse-mode="width"
                    :collapsed-width="64"
                    @collapse="collapsed = true"
                    @expand="collapsed = false"
                    :collapsed="collapsed"
                    content-style="padding-top: 24px;"
                    :native-scrollbar="false" bordered
                    width="180"
                    show-trigger
                    v-show="!isMobile"
            >
                <n-menu :options="menuOptions"
                        :collapsed="collapsed"
                        :collapsed-width="64"
                        :collapsed-icon-size="22"
                />

            </n-layout-sider>
            <n-layout content-style="padding: 24px;" :native-scrollbar="false">
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
</style>