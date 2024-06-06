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
    <n-card title="User">
      <p>nickname</p>
      <p>积分 声望</p>
      <p>关注 被关注</p>
      <p>2021-12-18 加入</p>
      <n-menu :options="options" style="width: 180px" default-value="1"/>
    </n-card>
  </n-space>
</template>
