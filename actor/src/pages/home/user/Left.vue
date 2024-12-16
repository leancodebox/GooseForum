<script setup>
import {NEllipsis, NMenu, NSpace} from 'naive-ui'
import {h} from 'vue'

let options = [
  {
    label: () =>
        h(NEllipsis, null, {default: () => '我的发布'}),
    key: '1'
  },
  {
    label: () =>
        h(NEllipsis, null, {default: () => '我的回复'}),
    key: '2'
  }
]
</script>
<template>
    <n-menu :options="options" style="min-width: 180px" default-value="1"/>
</template>
