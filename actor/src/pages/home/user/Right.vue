<script setup>
import {NCard, NEllipsis, NMenu, NSpace} from 'naive-ui'
import {h} from 'vue'

let options = [
  {
    label: () =>
        h(NEllipsis, null, {default: () => '快捷入口1'}),
    key: '1'
  },
  {
    label: () =>
        h(NEllipsis, null, {default: () => '快捷入口2'}),
    key: '2'
  }
]
</script>
<template>

  <n-space vertical>
    <n-card title="User"></n-card>
    <n-menu :options="options" style="width: 180px" default-value="1"/>
  </n-space>
</template>
