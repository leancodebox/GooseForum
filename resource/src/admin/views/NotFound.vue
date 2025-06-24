<template>
  <div class="min-h-screen flex items-center justify-center bg-base-200">
    <div class="text-center">
      <div class="mb-8">
        <h1 class="text-9xl font-bold text-primary opacity-20">404</h1>
      </div>
      
      <div class="space-y-4">
        <h2 class="text-3xl font-bold text-base-content">页面不存在</h2>
        <p class="text-base-content/70 max-w-md mx-auto">
          抱歉，您访问的页面不存在或已被移除。请检查URL是否正确，或返回管理后台首页。
        </p>
      </div>
      
      <div class="mt-8 space-x-4">
        <router-link to="/admin" class="btn btn-primary">
          <HomeIcon class="w-4 h-4" />
          返回首页
        </router-link>
        <button @click="$router.go(-1)" class="btn btn-outline">
          <ArrowLeftIcon class="w-4 h-4" />
          返回上页
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { HomeIcon, ArrowLeftIcon } from '@heroicons/vue/24/outline'
</script>

<style scoped>
/* 添加一些动画效果 */
.text-9xl {
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-10px);
  }
}
</style>