<script setup>
import {
  dateZhCN,
  NButton,
  NConfigProvider,
  NDropdown,
  NIcon,
  NLayout,
  NLoadingBarProvider,
  NMessageProvider,
  NNotificationProvider,
  NFloatButton,
  zhCN
} from 'naive-ui'
import {h} from 'vue'
import {SparklesOutline} from '@vicons/ionicons5'
import {RouterLink} from "vue-router";
// import HelloWorld from './components/HelloWorld.vue'
// This starter template is using Vue 3 experimental <script setup> SFCs
// Check out https://github.com/vuejs/rfcs/blob/master/active-rfcs/0040-script-setup.md


// 全局浮动按钮的下拉菜单选项
const options = [
  {
    label: 'super menu',
    key: 'marina bay sands',
    disabled: false
  },
  {
    label: () => h(
        RouterLink,
        {
          to: {
            path: '/manager/',
          }
        },
        {default: () => 'manager'}),
    key: "manager"
  },

  {
    label: () => h(
        RouterLink,
        {
          to: {
            path: '/home/',
          }
        },
        {default: () => 'home'}),
    key: "home"
  },
  {
    label: () => h(
        RouterLink,
        {
          to: {
            path: '/home/regOrLogin'
          }
        },
        {default: () => 'login'},
    ),
    key: "login"
  }
]

function handleSelect(key) {
  console.log(key)
}
</script>

<template>
  <!-- 全局通知提供者，限制最大显示数量为3 -->
  <n-notification-provider :max="3">
    <!-- 全局消息提供者 -->
    <n-message-provider>
      <!-- 全局加载条提供者 -->
      <n-loading-bar-provider>
        <!-- 全局配置提供者，设置中文语言 -->
        <n-config-provider :locale="zhCN" :date-locale="dateZhCN">
          <!-- 主布局容器 -->
          <n-layout position="absolute" content-style="width: 100%;height: 100%;">
            <router-view></router-view>
          </n-layout>
          <!-- 全局浮动按钮 -->
          <n-dropdown trigger="hover" :options="options" @select="handleSelect" :size="'huge'">
            <n-float-button circle :right="56" :bottom="30" style="z-index:1501">
              <n-icon>
                <sparkles-outline/>
              </n-icon>
            </n-float-button>
          </n-dropdown>
        </n-config-provider>
      </n-loading-bar-provider>
    </n-message-provider>
  </n-notification-provider>
</template>

<style>
@import "app.css";
</style>
