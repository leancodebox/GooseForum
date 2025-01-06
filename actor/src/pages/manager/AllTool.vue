<script setup>
import {NButton, NCard, NGrid, NGridItem, NIcon, NStatistic, useMessage} from 'naive-ui'
import {getUserInfo} from "@/service/request"

const message = useMessage()

function getUserInfoAction() {
  getUserInfo().then(r => {
    message.success(JSON.stringify(r.result))
  })
}
</script>
<template>
  <n-grid :cols="24">
    <n-grid-item :span="12">

      <n-card title="统计数据">
        <n-grid :cols="24">
          <n-grid-item :span="8">
            <n-statistic label="文章/总数(包含草稿)" :value="99">
              <template #prefix>
                <n-icon>

                </n-icon>
              </template>
              <template #suffix>/ 100</template>
            </n-statistic>
          </n-grid-item>
          <n-grid-item :span="8">
            <n-statistic label="用户数量">1,234,123</n-statistic>
          </n-grid-item>
          <n-grid-item :span="8">
            <n-statistic label="回复量">1,234,123</n-statistic>
          </n-grid-item>
        </n-grid>
      </n-card>
    </n-grid-item>

    <n-grid-item :span="12">
      <n-card title="语言包">
        <n-grid :cols="24">
          <n-grid-item :span="8">
            简体中文
          </n-grid-item>
          <n-grid-item :span="8">
            简体中文
          </n-grid-item>
          <n-grid-item :span="8">
            简体中文
          </n-grid-item>
        </n-grid>
      </n-card>
    </n-grid-item>
    <n-grid-item :span="12">
      <n-card title="功能模块">
        <n-button @click="getUserInfoAction">获取用户信息</n-button>
      </n-card>
    </n-grid-item>
  </n-grid>

</template>
<style>
.carousel-img {
  width: 100%;
  height: 240px;
  object-fit: cover;
}
</style>
