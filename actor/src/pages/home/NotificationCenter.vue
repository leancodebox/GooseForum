<script setup>
import {NEllipsis, NGi, NGrid, NList, NListItem, NMenu, NSpace} from "naive-ui"
import {h} from "vue"

let options = [
  {
    label: () =>
        h(NEllipsis, null, {default: () => '只看未读'}),
    key: '1'
  },
  {
    label: () =>
        h(NEllipsis, null, {default: () => '全部消息'}),
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
          <p><span>故人重来</span> 回答了问题 <span>webman中如何让php文件加载一次后就常驻内存了？</span></p>
          <p>6小时前</p>
        </n-list-item>
        <n-list-item>
          <p><span>故人重来</span> 回答了问题 <span>webman中如何让php文件加载一次后就常驻内存了？</span></p>
          <p>6小时前</p>
        </n-list-item>
        <n-list-item>
          <p><span>故人重来</span> 回答了问题 <span>webman中如何让php文件加载一次后就常驻内存了？</span></p>
          <p>6小时前</p>
        </n-list-item>
      </n-list>
    </n-gi>
  </n-grid>
</template>
