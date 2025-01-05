<script setup>
import {NButton, NFlex, NIcon, NLayout, NLayoutFooter, NLayoutHeader, NMenu, NText} from 'naive-ui';
import {useIsMobile, useIsTablet} from "@/utils/composables";
import {h, ref} from "vue";
import {RouterLink} from "vue-router";
import UserInfoCard from "@/components/UserInfoMenu.vue";

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
    // icon: renderIcon(FishOutline),
    children: null,
  }, {
    label: () => h(RouterLink, {to: {path: "/home/bbs",}},
        {
          default: () => "七嘴八舌"
        }
    ),
    key: "bbs",
    // icon: renderIcon(InfiniteOutline),
    children: null,
  }, {
    label: () => h(RouterLink, {to: {path: "/home/about",}},
        {
          default: () => "关于"
        }
    ),
    key: "about",
    // icon: renderIcon(CashIcon),
    children: null,
  }
]
let isTablet = useIsTablet()
let isMobile = useIsMobile()

let topHeight = ref('56px')
</script>
<template>
  <n-layout-header
      position="absolute"
      :style="{height: topHeight,  padding: '8px',align:'center'}"
      bordered
  >
    <n-flex align="center" style="text-align: center;height: 100%;"
            :justify="(!isTablet&&!isMobile)? 'space-between':'end'">
      <n-text tag="div" class="ui-logo" :depth="1" @click="console.log(1)" style="align-items: center;"
              v-if="!isTablet&&!isMobile">
        <img alt="" src="/quote-left.png" style="height: 20px"/>
        <span>GooseForum</span>
        <n-menu v-model:value="activeKey" mode="horizontal" :options="menuOptions"/>
      </n-text>
      <n-flex style="padding-right: 24px;">
        <router-link :to="'/home/bbs/articlesEdit'">
          <n-button quaternary>我也写一篇</n-button>
        </router-link>
        <router-link :to="'/home/notificationCenter'">
          <n-button quaternary>消息中心</n-button>
        </router-link>
        <!--          <n-button quaternary>审核中心</n-button>-->
        <user-info-card></user-info-card>
      </n-flex>
    </n-flex>
  </n-layout-header>
  <n-layout :style="{top: topHeight}"
            :native-scrollbar="false"
            :position="'absolute'"
            content-style="min-height: calc(100vh - var(--header-height)); display: flex; flex-direction: column;"
  >
    <router-view></router-view>
    <n-layout-footer
        style="--x-padding: 56px;margin-top: auto;height: 100px;display: flex;justify-content: center;align-items: center;width: 100%;background-color: #f0f3f5;"
        bordered>

      <p>&copy; My Homepage 2023</p>
    </n-layout-footer>
  </n-layout>
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
