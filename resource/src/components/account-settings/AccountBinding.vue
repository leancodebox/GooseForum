<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getOAuthBindings, unbindOAuth } from '@/utils/articleService.ts'
import type { OAuthBindings } from "@/utils/articleInterfaces";

// 提示消息状态
const showSuccessMessage = ref(false)
const showErrorMessage = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

// OAuth绑定相关状态
const oauthBindings = ref<OAuthBindings>({})
const loadingBindings = ref(false)

// 检查URL参数中的消息
const checkUrlMessages = () => {
  const urlParams = new URLSearchParams(window.location.search);
  const success = urlParams.get('success');
  const error = urlParams.get('error');
  
  if (success === 'bind_success') {
    successMessage.value = 'OAuth账户绑定成功！'
    showSuccessMessage.value = true
    // 3秒后自动隐藏
    setTimeout(() => {
      showSuccessMessage.value = false
    }, 3000)
    // 重新获取绑定状态
    fetchOAuthBindings();
    // 清理success参数，保留其他参数
    urlParams.delete('success')
    const newUrl = urlParams.toString() ? `${window.location.pathname}?${urlParams.toString()}` : window.location.pathname
    window.history.replaceState({}, '', newUrl)
  } else if (error) {
    errorMessage.value = decodeURIComponent(error)
    showErrorMessage.value = true
    // 5秒后自动隐藏
    setTimeout(() => {
      showErrorMessage.value = false
    }, 5000)
    // 清理error参数，保留其他参数
    urlParams.delete('error')
    const newUrl = urlParams.toString() ? `${window.location.pathname}?${urlParams.toString()}` : window.location.pathname
    window.history.replaceState({}, '', newUrl)
  }
}

// 获取OAuth绑定状态
const fetchOAuthBindings = async () => {
  try {
    loadingBindings.value = true
    const response = await getOAuthBindings()
    if (response && response.result) {
      oauthBindings.value = response.result
    }
  } catch (error) {
    console.error('获取OAuth绑定状态失败:', error)
    showErrorMessage.value = true
    errorMessage.value = '获取绑定状态失败，请重试'
    setTimeout(() => {
      showErrorMessage.value = false
    }, 5000)
  } finally {
    loadingBindings.value = false
  }
}

// 绑定OAuth账户
const bindOAuth = (provider: string) => {
  // 直接跳转到登录URL（已登录状态下会自动进入绑定模式）
  window.location.href = `/api/auth/${provider}`
}

// 解绑OAuth账户
const unbindOAuthAccount = async (provider: string) => {
  if (!confirm(`确定要解绑${provider.toUpperCase()}账户吗？`)) {
    return
  }

  try {
    const response = await unbindOAuth(provider)
    if (response.code === 0) {
       successMessage.value = 'OAuth账户解绑成功！'
       showSuccessMessage.value = true
       setTimeout(() => {
         showSuccessMessage.value = false
       }, 3000)
      await fetchOAuthBindings() // 重新获取绑定状态
    } else {
      showErrorMessage.value = true
      errorMessage.value = response.message || '解绑失败，请重试'
      setTimeout(() => {
        showErrorMessage.value = false
      }, 5000)
    }
  } catch (error: any) {
    console.error('解绑失败:', error)
    showErrorMessage.value = true
    errorMessage.value = error.message || '解绑失败，请重试'
    setTimeout(() => {
      showErrorMessage.value = false
    }, 5000)
  }
}

// 组件挂载时获取OAuth绑定状态
onMounted(() => {
  // 获取OAuth绑定状态
  fetchOAuthBindings()
  // 检查URL消息
  checkUrlMessages()
})
</script>

<template>
  <!-- 提示消息 -->
  <div class="fixed top-20 right-4 z-50 space-y-2">
    <!-- 成功消息 -->
    <div v-if="showSuccessMessage" class="alert alert-success shadow-lg max-w-sm">
      <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
          d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <span>{{ successMessage }}</span>
    </div>

    <!-- 错误消息 -->
    <div v-if="showErrorMessage" class="alert alert-error shadow-lg max-w-sm">
      <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
          d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <span>绑定失败: {{ errorMessage }}</span>
    </div>
  </div>

  <div>
    <h3 class="card-title text-lg mb-6 border-b border-base-300 pb-3">第三方账号绑定</h3>
    <div v-if="loadingBindings" class="flex justify-center py-8">
      <span class="loading loading-spinner loading-lg"></span>
    </div>
    <div v-else class="grid grid-cols-1 gap-6">
      <!-- GitHub绑定 -->
      <div class="card bg-base-200 shadow-sm">
        <div class="card-body p-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 bg-gray-900 rounded-full flex items-center justify-center">
                <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd"
                    d="M10 0C4.477 0 0 4.484 0 10.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0110 4.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.203 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.942.359.31.678.921.678 1.856 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0020 10.017C20 4.484 15.522 0 10 0z"
                    clip-rule="evenodd" />
                </svg>
              </div>
              <div>
                <h4 class="font-medium">GitHub</h4>
                <p class="text-sm text-base-content/60">绑定GitHub账户</p>
              </div>
            </div>
            <div v-if="oauthBindings.github?.bound" class="flex items-center gap-2">
              <span class="badge badge-success">已绑定</span>
              <button @click="unbindOAuthAccount('github')" class="btn btn-sm btn-error">解绑</button>
            </div>
            <button v-else @click="bindOAuth('github')" class="btn btn-sm btn-primary">绑定</button>
          </div>
        </div>
      </div>

      <!-- Google绑定 -->
      <div class="card bg-base-200 shadow-sm" v-if="false">
        <div class="card-body p-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 bg-white rounded-full flex items-center justify-center border">
                <svg class="w-5 h-5" viewBox="0 0 24 24">
                  <path fill="#4285F4"
                    d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" />
                  <path fill="#34A853"
                    d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
                  <path fill="#FBBC05"
                    d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
                  <path fill="#EA4335"
                    d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
                </svg>
              </div>
              <div>
                <h4 class="font-medium">Google</h4>
                <p class="text-sm text-base-content/60">绑定Google账户</p>
              </div>
            </div>
            <div v-if="oauthBindings.google?.bound" class="flex items-center gap-2">
              <span class="badge badge-success">已绑定</span>
              <button @click="unbindOAuthAccount('google')" class="btn btn-sm btn-error">解绑</button>
            </div>
            <button v-else @click="bindOAuth('google')" class="btn btn-sm btn-primary">绑定</button>
          </div>
        </div>
      </div>

      <div class="alert alert-info">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"
          class="stroke-current shrink-0 w-6 h-6">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
        </svg>
        <span>绑定第三方账户后，您可以使用这些账户快速登录。解绑后将无法使用对应账户登录。</span>
      </div>
    </div>
  </div>
</template>

<style scoped></style>