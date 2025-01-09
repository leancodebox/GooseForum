<script setup>
import {NDropdown, NFloatButton, NIcon} from 'naive-ui'
import {h} from 'vue'
import {SparklesOutline} from '@vicons/ionicons5'
import {RouterLink} from "vue-router";

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

</script>

<template>
  <!-- 全局浮动按钮 -->
  <n-dropdown trigger="hover" :options="options" :size="'huge'">
    <n-float-button circle :right="56" :bottom="30" style="z-index:1501">
      <n-icon>
        <sparkles-outline/>
      </n-icon>
    </n-float-button>
  </n-dropdown>
</template>

<style>
</style>
