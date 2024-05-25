<script setup>
import {NEllipsis, NMenu, NSpace} from 'naive-ui'
import {h} from 'vue'

let options = [
  {
    label: () =>
        h(NEllipsis, null, {default: () => '电灯熄灭 物换星移 泥牛入海'}),
    key: '1'
  },
  {
    label: () =>
        h(NEllipsis, null, {default: () => '黑暗好像 一颗巨石 按在胸口'}),
    key: '2'
  }
]
</script>
<template>
  <n-space vertical>
    <n-menu :options="options" style="width: 180px" default-value="1"/>
    <n-menu :options="options" style="width: 180px" default-value="1"/>
    <n-menu :options="options" style="width: 180px" default-value="1"/>
  </n-space>
  <n-space vertical>
    <n-menu :options="options" style="width: 180px" default-value="1"/>
    <n-menu :options="options" style="width: 180px" default-value="1"/>
    <n-menu :options="options" style="width: 180px" default-value="1"/>
  </n-space>
</template>
