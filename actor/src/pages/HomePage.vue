<script setup>
import {NButton, NFlex, NIcon, NLayout, NLayoutFooter, NLayoutHeader, NMenu, NSpace} from 'naive-ui';
import {useIsMobile, useIsTablet} from "@/utils/composables";
import {FastFoodOutline as CashIcon, FishOutline, InfiniteOutline} from '@vicons/ionicons5'
import {h, ref} from "vue";
import {RouterLink} from "vue-router";
import UserInfoCard from "@/pages/home/UserInfoCard.vue";

function renderIcon(icon) {
  return () => h(NIcon, null, {default: () => h(icon)});
}

const activeKey = ref('home')
const menuOptions = [
  {
    label: () => h(RouterLink, {to: {path: "/home/index",}},
        {
          default: () => "首页"
        }
    ),
    key: "index",
    icon: renderIcon(FishOutline),
    children: null,
  }, {
    label: () => h(RouterLink, {to: {path: "/home/bbs",}},
        {
          default: () => "论坛"
        }
    ),
    key: "bbs",
    icon: renderIcon(InfiniteOutline),
    children: null,
  }, {
    label: () => h(RouterLink, {to: {path: "/home/about",}},
        {
          default: () => "关于"
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
let isTablet = useIsTablet()
let isMobile = useIsMobile()
</script>
<template>
  <n-layout position="absolute">
    <n-layout-header position="absolute"
                     style="height: 64px;align-items: center;"
                     bordered
                     :style="style"
    >
      <n-flex justify="space-between" style="padding: 0 var(--side-padding);height: 100%">
        <n-space justify="space-around" style="align-items: center;" v-show="!isTablet&&!isMobile">
          <span style="font-size: 18px;">GooseForum</span>
          <n-menu v-model:value="activeKey" mode="horizontal" :options="menuOptions"/>
        </n-space>
        <n-flex justify="end" style="align-items: center;">
          <router-link :to="'/home/bbs/articlesEdit'">
            <n-button quaternary>我也写一篇</n-button>
          </router-link>
          <n-button quaternary>消息</n-button>
          <n-button quaternary>审核中心</n-button>
          <user-info-card></user-info-card>
        </n-flex>
      </n-flex>
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

</style>
