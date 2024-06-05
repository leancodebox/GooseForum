<script setup>
import {NEllipsis, NGi, NGrid, NList, NListItem, NMenu, NSpace} from "naive-ui"
import {h} from "vue"

let options = [
  {
    label: () =>
        h(NEllipsis, null, {default: () => '个人自料'}),
    key: '1'
  },
  {
    label: () =>
        h(NEllipsis, null, {default: () => '头像设置'}),
    key: '2'
  },
  {
    label: () =>
        h(NEllipsis, null, {default: () => '修改密码'}),
    key: '2'
  }
]
</script>
<template>
  <n-grid cols="5" style="padding: 0 10px;height: 100%">
    <n-gi span="1">
      <n-space>
        <n-menu :options="options" style="width: 180px" default-value="1"/>
      </n-space>
    </n-gi>
    <n-gi span="4">
      <n-list>
        <n-list-item>
          消息1
        </n-list-item>
        <n-list-item>
          消息2
        </n-list-item>
        <n-list-item>
          消息3
        </n-list-item>
      </n-list>
    </n-gi>
  </n-grid>
</template>
