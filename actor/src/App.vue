<script setup>
import {
  darkTheme,
  dateZhCN,
  NConfigProvider,
  NDialogProvider,
  NLayout,
  NLoadingBarProvider,
  NMessageProvider,
  NNotificationProvider,
  zhCN
} from 'naive-ui'
import {useThemeStore} from '@/modules/theme.js';
import SuperMenu from "@/components/SuperMenu.vue";

const themeStore = useThemeStore();
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
          <n-config-provider :locale="zhCN" :date-locale="dateZhCN" :theme="themeStore.isDarkTheme ? darkTheme : null">
            <!-- 主布局容器 -->
            <n-layout position="absolute" content-style="width: 100%;height: 100%;">
              <router-view></router-view>
            </n-layout>
            <!-- 全局浮动按钮 -->
            <super-menu v-if="false"></super-menu>
          </n-config-provider>
        </n-loading-bar-provider>
      </n-message-provider>
    </n-dialog-provider>
  </n-notification-provider>
</template>

<style>
@import "app.css";
</style>
