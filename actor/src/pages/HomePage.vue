<script setup>
import {NAutoComplete, NButton, NIcon, NLayout, NLayoutFooter, NLayoutHeader, NMenu} from 'naive-ui';
import {useIsMobile, useIsSmallDesktop, useIsTablet} from '@/utils/composables';

import {FastFoodOutline as CashIcon, FishOutline, InfiniteOutline} from '@vicons/ionicons5'
import {h, ref} from "vue";
import {RouterLink} from "vue-router";
import UserInfoCard from "@/pages/home/UserInfoCard.vue";

function renderIcon(icon) {
  return () => h(NIcon, null, {default: () => h(icon)});
}

const isMobileRef = useIsMobile()
const isTabletRef = useIsTablet()
const isSmallDesktop = useIsSmallDesktop()
const isMobile = isMobileRef
const activeKey = ref('home')
const menuOptions = [
  {
    label: () => h(RouterLink, {to: {path: "/home/index",}},
        {
          default: () => "index"
        }
    ),
    key: "index",
    icon: renderIcon(FishOutline),
    children: null,
  }, {
    label: () => h(RouterLink, {to: {path: "/home/bbs",}},
        {
          default: () => "bbs"
        }
    ),
    key: "bbs",
    icon: renderIcon(InfiniteOutline),
    children: null,
  }, {
    label: () => h(RouterLink, {to: {path: "/home/about",}},
        {
          default: () => "about"
        }
    ),
    key: "about",
    icon: renderIcon(CashIcon),
    children: null,
  }
]
let style = {
  '--side-padding': '32px',
  'grid-template-columns':
      'calc(272px - var(--side-padding)) 1fr auto'
}
</script>
<template>
  <n-layout position="absolute">

    <n-layout-header class="l-header" position="absolute"
                     style="height: 64px;align-items: center;"
                     bordered
                     :style="style"
    >
      <n-text tag="div" :depth="1" >
        <span >学</span>
      </n-text>
      <div style="display: flex; align-items: center; overflow: hidden;">
      <div class="nav-menu">
        <n-menu v-model:value="activeKey" mode="horizontal" :options="menuOptions"/>
      </div>
      <n-auto-complete style="width: 280px; margin-left: 24px; margin-right: 12px; flex-shrink: 0;"         clear-after-select
                       blur-after-select>
      </n-auto-complete>
      </div>
      <div class="nav-end">
        <router-link :to="'/home/bbs/articlesEdit'">
          <n-button quaternary>我也写一篇</n-button>
        </router-link>
        <n-button quaternary>消息</n-button>
        <n-button quaternary>审核中心</n-button>
        <div style="width:15px"></div>
        <user-info-card></user-info-card>
      </div>
    </n-layout-header>

    <n-layout style="top: 64px;"
              :native-scrollbar="false"
              :position=" 'absolute'"
              content-style="min-height: calc(100vh - var(--header-height)); display: flex; flex-direction: column;"
    >
      <router-view></router-view>

      <n-layout-footer
          style="--x-padding: 56px;margin-top: auto;height: 100px;display: flex;justify-content: center;align-items: center;width: 100%;background-color: #f0f3f5;"
          bordered>

        <p>&copy; My Homepage 2023</p>
      </n-layout-footer>
    </n-layout>
  </n-layout>
</template>

<style>
.l-header {
  display: grid;
  grid-template-rows: calc(var(--header-height) - 1px);
  align-items: center;
  padding: 0 var(--side-padding);
}

.nav-menu {
  padding-left: 36px;
  overflow: hidden;
  flex-grow: 0;
  flex-shrink: 1;
}

.nav-end {
  display: flex;
  align-items: center;
}
</style>
