<script setup>
import {h, ref} from 'vue'
import {RouterLink} from 'vue-router'
import {Flash} from '@vicons/ionicons5'
import {NDrawer, NDrawerContent, NButton, NDropdown, NIcon, NLayout, NLayoutHeader, NLayoutSider, NMenu} from 'naive-ui'
import routes from "@/route/routes";
import {useIsMobile} from "@/utils/composables";

function renderIcon(icon) {
    return () => h(NIcon, null, {default: () => h(icon)})
}

function buildMenuOption(parentPath, route, deep = 0) {
    if (route.belongMenu === undefined || route.belongMenu !== true) {
        return []
    }
    let menuList = [];
    if (deep === 0) {
        for (let key in route.children) {
            menuList.push(...buildMenuOption(route.path, route.children[key], deep + 1))
        }
    } else {
        let path
        if (parentPath.charAt(parentPath.length - 1) !== '/' && route.path.charAt(0) !== '/') {
            path = parentPath + '/' + route.path
        } else {
            path = parentPath + route.path
        }
        let children = [];
        if (route.children !== undefined && route.children.length > 0) {
            for (let key in route.children) {
                children.push(...buildMenuOption(path, route.children[key], deep + 1))
            }
        }

        let label = function (path) {
            return () => h(RouterLink, {to: {path: path,}},
                {
                    default: () => {
                        return route.showName === undefined ? route.name : route.showName;
                    }
                }
            )
        }(path)

        if (children.length === 0) {
            children = undefined
        } else {
            label = route.showName === undefined ? route.name : route.showName;
        }
        let icon = renderIcon(Flash)
        if (route.icon !== undefined) {
            icon = renderIcon(route.icon)
        }
        menuList.push({
            label: label,
            key: path,
            icon: icon,
            children: children
        })
    }
    return menuList
}

let menuOptions = []
for (const element of routes) {
    menuOptions.push(...buildMenuOption('', element))
}
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