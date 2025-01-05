<script setup>
import {NEllipsis, NMenu, NSpace} from 'naive-ui'
import {h} from 'vue'

let options = [
  {
    label: () =>
        h(NEllipsis, null, {default: () => '我的博文'}),
    key: '2'
  },
  {
    label: () =>
        h(NEllipsis, null, {default: () => '我的回复'}),
    key: '3'
  }
]
</script>
<template>
    <n-menu :options="options" style="min-width: 180px" default-value="1"/>
</template>
