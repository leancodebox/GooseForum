<script setup>
import {NGi, NGrid, NIcon, NLayout, NLayoutFooter, NLayoutHeader, NMenu,NButton} from 'naive-ui';
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
</script>
<template>
  <n-layout position="absolute">

    <n-layout-header class="l-header" position="absolute"
                     style="height: 64px; padding: 8px;"
                     bordered
    >
      <n-grid cols="24" item-responsive>
        <n-gi span="8">
          <n-menu v-model:value="activeKey" mode="horizontal" :options="menuOptions"/>
        </n-gi>
        <n-gi span="0 800:8">
          <!--                    <n-input type="text">asdasd</n-input>-->
        </n-gi>
        <n-gi span="0 800:8">
          <n-gi style="float:right;padding-right: 30px">
            <user-info-card></user-info-card>
          </n-gi>

          <n-gi style="float:right;padding-right: 30px">
            <router-link :to="'/home/bbs/articlesEdit'"><n-button>我也写一篇</n-button></router-link>
          </n-gi>

        </n-gi>
      </n-grid>
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
  padding: 12px 12px;
}
</style>