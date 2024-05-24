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
  zhCN
} from 'naive-ui'
import {h} from 'vue'
import {SparklesOutline} from '@vicons/ionicons5'
import {RouterLink} from "vue-router";
// import HelloWorld from './components/HelloWorld.vue'
// This starter template is using Vue 3 experimental <script setup> SFCs
// Check out https://github.com/vuejs/rfcs/blob/master/active-rfcs/0040-script-setup.md


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
            path: '/login/'
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
  <n-message-provider>
    <n-loading-bar-provider>
      <n-config-provider :locale="zhCN" :date-locale="dateZhCN">
        <n-layout position="absolute">
          <router-view></router-view>
        </n-layout>

        <div
            style="right: 15px;bottom: 100px;position:absolute;z-index:1501"
        >
          <n-dropdown trigger="hover" :options="options" @select="handleSelect" :size="'huge'">
            <n-button circle size="large" type="success">
              <template #icon>
                <n-icon>
                  <sparkles-outline/>
                </n-icon>
              </template>
            </n-button>
          </n-dropdown>

        </div>
      </n-config-provider>
    </n-loading-bar-provider>
  </n-message-provider>
</template>

<style>
@import "app.css";
</style>
