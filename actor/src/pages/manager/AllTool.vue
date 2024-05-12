<script setup>
import {NButton, NCard, NGrid, NGridItem, NIcon, NSpace, NStatistic, useMessage} from 'naive-ui'
import {getUserInfo} from "@/service/request"
import CpTool from "@/pages/manager/tool/CpTool.vue";

const message = useMessage()

function getUserInfoAction() {
  getUserInfo().then(r => {
    message.success(JSON.stringify(r.result))
  })
}
</script>
<template>
  <n-space vertical>
    <n-card>
      <cp-tool></cp-tool>
    </n-card>
    <n-card>
      <n-button @click="getUserInfoAction"> 获取用户信息</n-button>
    </n-card>
    <n-card title="统计数据">
      <n-grid>
        <n-grid-item :span="12">
          <n-statistic label="统计数据" :value="99">
            <template #prefix>
              <n-icon>

              </n-icon>
            </template>
            <template #suffix>/ 100</template>
          </n-statistic>
        </n-grid-item>
        <n-grid-item :span="12">
          <n-statistic label="活跃用户">1,234,123</n-statistic>
        </n-grid-item>
      </n-grid>
    </n-card>
  </n-space>
</template>
<style>
.carousel-img {
  width: 100%;
  height: 240px;
  object-fit: cover;
}
</style>
