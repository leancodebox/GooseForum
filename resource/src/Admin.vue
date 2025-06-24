<template>
  <div class="min-h-screen bg-base-200">
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from './admin/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

onMounted(() => {
  // 检查认证状态
  if (!authStore.isAuthenticated && router.currentRoute.value.path !== '/admin/login') {
    router.push('/admin/login')
  }
})
</script>

<style>
/* 全局样式 */
body {
  margin: 0;
  padding: 0;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

/* 自定义滚动条 */
::-webkit-scrollbar {
  width: 6px;
}

::-webkit-scrollbar-track {
  background: hsl(var(--b2));
}

::-webkit-scrollbar-thumb {
  background: hsl(var(--bc) / 0.3);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: hsl(var(--bc) / 0.5);
}
</style>