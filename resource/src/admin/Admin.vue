<template>
  <div class="min-h-screen bg-base-200">
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth.ts'

const router = useRouter()
const authStore = useAuthStore()

onMounted(async () => {
  // 等待认证状态初始化完成
  await authStore.initAuth()
  
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

</style>