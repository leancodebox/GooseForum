<template>
  <n-flex vertical>
    <n-grid :cols="24" :x-gap="16" :y-gap="16">
      <n-grid-item :span="6">
        <n-statistic label="总用户数">
          <n-number-animation :from="0" :to="1234" />
        </n-statistic>
      </n-grid-item>
      <n-grid-item :span="6">
        <n-statistic label="总帖子数">
          <n-number-animation :from="0" :to="5678" />
        </n-statistic>
      </n-grid-item>
      <n-grid-item :span="6">
        <n-statistic label="今日新增用户">
          <n-number-animation :from="0" :to="42" />
        </n-statistic>
      </n-grid-item>
      <n-grid-item :span="6">
        <n-statistic label="今日新增帖子">
          <n-number-animation :from="0" :to="87" />
        </n-statistic>
      </n-grid-item>
    </n-grid>

    <n-card title="最近活动" style="margin-top: 16px">
      <n-list>
        <n-list-item v-for="(item, index) in recentActivities" :key="index">
          <n-thing :title="item.title" :description="item.time">
            {{ item.content }}
          </n-thing>
        </n-list-item>
      </n-list>
    </n-card>
  </n-flex>
</template>

<script setup lang="ts">
import { ref } from 'vue'

// 模拟最近活动数据
const recentActivities = ref([
  {
    title: '用户注册',
    content: '新用户 user123 完成注册',
    time: '10分钟前'
  },
  {
    title: '帖子发布',
    content: '用户 admin 发布了新帖子《Vue3 最佳实践》',
    time: '30分钟前'
  },
  {
    title: '评论',
    content: '用户 guest 评论了帖子《如何使用 Naive UI》',
    time: '1小时前'
  },
  {
    title: '系统更新',
    content: '系统完成了安全更新',
    time: '2小时前'
  },
  {
    title: '用户登录',
    content: '管理员登录了系统',
    time: '3小时前'
  }
])
</script>

<style scoped>

</style>
