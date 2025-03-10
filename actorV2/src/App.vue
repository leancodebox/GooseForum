<script setup lang="ts">
import {RouterView} from 'vue-router'
import TheHeader from '@/components/TheHeader.vue'
import TheFooter from '@/components/TheFooter.vue'
import {
  dateZhCN,
  NConfigProvider,
  NDialogProvider,
  NLoadingBarProvider,
  NMessageProvider,
  NNotificationProvider,
  zhCN
} from 'naive-ui'
</script>

<template>
  <div class="app-container">
    <TheHeader/>
    <main>
      <n-notification-provider :max="3">
        <n-dialog-provider>
          <!-- 全局消息提供者 -->
          <n-message-provider>
            <!-- 全局加载条提供者 -->
            <n-loading-bar-provider>
              <!-- 全局配置提供者，设置中文语言 -->
              <n-config-provider :locale="zhCN" :date-locale="dateZhCN">
                <RouterView/>
              </n-config-provider>
            </n-loading-bar-provider>
          </n-message-provider>
        </n-dialog-provider>
      </n-notification-provider>
    </main>
    <TheFooter/>
  </div>
</template>

<style>
.app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

main {
  flex: 1;
  max-width: 1440px;
  margin: 1rem auto;
  padding: 0 2rem;
  width: 100%;
  box-sizing: border-box;
}

/* 响应式调整 */
@media (max-width: 1200px) {
  main {
    max-width: 1200px;
    padding: 0 1.5rem;
  }
}

@media (max-width: 900px) {
  main {
    padding: 0 0.25rem;
  }
}
</style>
