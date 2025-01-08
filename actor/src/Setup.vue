<script setup>
import {
  dateZhCN,
  NConfigProvider,
  NDialogProvider,
  NLoadingBarProvider,
  NMessageProvider,
  NNotificationProvider,
  zhCN
} from 'naive-ui'

import SetupPage from './pages/setup/SetupPage.vue'
</script>

<template>
  <!-- 全局通知提供者，限制最大显示数量为3 -->
  <n-notification-provider :max="3">
    <n-dialog-provider>
      <!-- 全局消息提供者 -->
      <n-message-provider>
        <!-- 全局加载条提供者 -->
        <n-loading-bar-provider>
          <!-- 全局配置提供者，设置中文语言 -->
          <n-config-provider :locale="zhCN" :date-locale="dateZhCN">
            <!-- 主布局容器 -->
            <setup-page></setup-page>
          </n-config-provider>
        </n-loading-bar-provider>
      </n-message-provider>
    </n-dialog-provider>
  </n-notification-provider>
</template>

<style>
@import "app.css";
</style>
