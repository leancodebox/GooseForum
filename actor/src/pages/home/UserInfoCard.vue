<script setup>

import {NAvatar, NButton, NDropdown, useMessage} from "naive-ui";

const message = useMessage()
let options = [
  {label: "我的发布", key: "我的发布"},
  {label: "个人中心", key: "个人中心"},
  {label: "编辑资料", key: "编辑资料"},
  {label: "账号退出", key: "账号退出"},
]
let handleSelect = function (key) {
  message.info(String(key))
}
</script>

<template>
  <n-dropdown trigger="hover" :options="options" @select="handleSelect">
    <n-button text>
      <n-avatar
          round
          :size="36"
      >
        旺财
      </n-avatar>
      <div style="width: 5px"></div>
      旺财
    </n-button>
  </n-dropdown>
</template>
