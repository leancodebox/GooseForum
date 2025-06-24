<template>
  <div class="min-h-screen bg-gradient-to-br from-primary/10 via-base-200 to-secondary/10 flex items-center justify-center p-4">
    <div class="card w-full max-w-md bg-base-100 shadow-2xl">
      <div class="card-body">
        <!-- Logo 和标题 -->
        <div class="text-center mb-8">
          <h1 class="text-3xl font-bold text-primary mb-2">GooseForum</h1>
          <p class="text-base-content/70">管理后台登录</p>
        </div>
        
        <!-- 登录表单 -->
        <form @submit.prevent="handleLogin" class="space-y-6">
          <!-- 用户名输入 -->
          <div class="form-control">
            <label class="label">
              <span class="label-text font-medium">用户名</span>
            </label>
            <div class="relative">
              <input 
                v-model="form.username"
                type="text" 
                placeholder="请输入用户名" 
                class="input input-bordered w-full pl-12"
                :class="{ 'input-error': errors.username }"
                required
                @blur="handleRememberChange"
              />
              <UserIcon class="absolute left-4 top-1/2 transform -translate-y-1/2 w-5 h-5 text-base-content/50 input-icon" />
            </div>
            <label v-if="errors.username" class="label">
              <span class="label-text-alt text-error">{{ errors.username }}</span>
            </label>
          </div>
          
          <!-- 密码输入 -->
          <div class="form-control">
            <label class="label">
              <span class="label-text font-medium">密码</span>
            </label>
            <div class="relative">
              <input 
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'" 
                placeholder="请输入密码" 
                class="input input-bordered w-full pl-12 pr-12"
                :class="{ 'input-error': errors.password }"
                required
              />
              <LockClosedIcon class="absolute left-4 top-1/2 transform -translate-y-1/2 w-5 h-5 text-base-content/50 input-icon" />
              <button 
                type="button"
                @click="showPassword = !showPassword"
                class="absolute right-4 top-1/2 transform -translate-y-1/2 text-base-content/50 hover:text-base-content transition-colors"
              >
                <EyeIcon v-if="!showPassword" class="w-5 h-5" />
                <EyeSlashIcon v-else class="w-5 h-5" />
              </button>
            </div>
            <label v-if="errors.password" class="label">
              <span class="label-text-alt text-error">{{ errors.password }}</span>
            </label>
          </div>
          
          <!-- 验证码输入 -->
          <div class="form-control">
            <label class="label">
              <span class="label-text font-medium">验证码</span>
            </label>
            <div class="flex gap-2">
              <div class="relative flex-1">
                <input 
                  v-model="form.captcha"
                  type="text" 
                  placeholder="请输入验证码" 
                  class="input input-bordered w-full pl-12"
                  :class="{ 'input-error': errors.captcha }"
                  maxlength="6"
                  required
                />
                <ExclamationTriangleIcon class="absolute left-4 top-1/2 transform -translate-y-1/2 w-5 h-5 text-base-content/50 input-icon" />
              </div>
              <div class="w-24 h-12">
                <img
                  v-if="captchaImg"
                  :src="captchaImg"
                  alt="验证码"
                  class="w-full h-full object-cover rounded cursor-pointer border border-base-300 hover:border-primary"
                  @click="refreshCaptcha"
                  title="点击刷新验证码"
                />
                <div v-else class="w-full h-full bg-base-200 rounded flex items-center justify-center text-xs text-base-content/50">
                  加载中...
                </div>
              </div>
            </div>
            <label v-if="errors.captcha" class="label">
              <span class="label-text-alt text-error">{{ errors.captcha }}</span>
            </label>
          </div>
          
          <!-- 记住我 -->
          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-3">
              <input 
                v-model="form.remember" 
                type="checkbox" 
                class="checkbox checkbox-primary checkbox-sm" 
              />
              <span class="label-text">记住我</span>
            </label>
          </div>
          
          <!-- 错误信息 -->
          <div v-if="authStore.error" class="alert alert-error">
            <ExclamationTriangleIcon class="w-5 h-5" />
            <span>{{ authStore.error }}</span>
          </div>
          
          <!-- 登录按钮 -->
          <div class="form-control mt-8">
            <button 
              type="submit" 
              class="btn btn-primary w-full"
              :class="{ 'loading': authStore.loading }"
              :disabled="authStore.loading"
            >
              <span v-if="!authStore.loading">登录</span>
              <span v-else>登录中...</span>
            </button>
          </div>
        </form>
        
        <!-- 底部信息 -->
        <div class="text-center mt-6 pt-6 border-t border-base-300">
          <p class="text-sm text-base-content/60">
            © 2024 GooseForum. 保留所有权利。
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import {
  UserIcon,
  LockClosedIcon,
  EyeIcon,
  EyeSlashIcon,
  ExclamationTriangleIcon
} from '@heroicons/vue/24/outline'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

// 表单数据
const form = reactive({
  username: '',
  password: '',
  captcha: '',
  remember: false
})

// 表单验证错误
const errors = reactive({
  username: '',
  password: '',
  captcha: ''
})

// 验证码相关
const captchaImg = ref('')
const captchaId = ref('')

// 显示密码
const showPassword = ref(false)

// 获取验证码
const getCaptcha = async () => {
  try {
    const response = await fetch('/api/get-captcha')
    const data = await response.json()
    if (data.code === 0) {
      captchaImg.value = data.result.captchaImg
      captchaId.value = data.result.captchaId
    }
  } catch (error) {
    console.error('获取验证码失败:', error)
  }
}

// 刷新验证码
const refreshCaptcha = () => {
  form.captcha = ''
  getCaptcha()
}

// 表单验证
const validateForm = () => {
  errors.username = ''
  errors.password = ''
  errors.captcha = ''
  
  let isValid = true
  
  if (!form.username.trim()) {
    errors.username = '请输入用户名'
    isValid = false
  } else if (form.username.length < 3) {
    errors.username = '用户名至少3个字符'
    isValid = false
  }
  
  if (!form.password) {
    errors.password = '请输入密码'
    isValid = false
  } else if (form.password.length < 6) {
    errors.password = '密码至少6个字符'
    isValid = false
  }
  
  if (!form.captcha.trim()) {
    errors.captcha = '请输入验证码'
    isValid = false
  }
  
  return isValid
}

// 处理登录
const handleLogin = async () => {
  if (!validateForm()) {
    return
  }
  
  const result = await authStore.login({
    username: form.username,
    password: form.password,
    captchaId: captchaId.value,
    captchaCode: form.captcha
  })
  
  if (result.success) {
    // 登录成功，跳转到目标页面或仪表盘
    const redirect = route.query.redirect as string || '/admin'
    router.push(redirect)
  } else {
    // 登录失败，刷新验证码
    refreshCaptcha()
  }
}

// 组件挂载时的处理
onMounted(() => {
  // 如果已经登录，直接跳转
  if (authStore.isAuthenticated) {
    router.push('/admin')
  }
  
  // 从 localStorage 恢复用户名
  const savedUsername = localStorage.getItem('admin_username')
  if (savedUsername) {
    form.username = savedUsername
    form.remember = true
  }
  
  // 获取验证码
  getCaptcha()
})

// 监听记住我选项
const handleRememberChange = () => {
  if (form.remember && form.username) {
    localStorage.setItem('admin_username', form.username)
  } else {
    localStorage.removeItem('admin_username')
  }
}
</script>

<style scoped>
/* 自定义样式 */
.card {
  backdrop-filter: blur(10px);
  border: 1px solid hsl(var(--bc) / 0.1);
}

.input:focus {
  box-shadow: 0 0 0 2px hsl(var(--p) / 0.2);
}

/* 修复输入框聚焦时图标颜色问题 */
.input-icon {
  color: hsl(var(--bc) / 0.5) !important;
  transition: color 0.2s ease;
}

/* 当输入框容器获得焦点时，图标颜色变深 */
.relative:focus-within .input-icon {
  color: hsl(var(--bc) / 0.75) !important;
}

/* 强制覆盖Tailwind的text-base-content/50类 */
.input-icon.text-base-content\/50 {
  color: hsl(var(--bc) / 0.5) !important;
}

.relative:focus-within .input-icon.text-base-content\/50 {
  color: hsl(var(--bc) / 0.75) !important;
}

/* 额外的选择器确保覆盖所有可能的Tailwind类 */
.absolute.input-icon {
  color: hsl(var(--bc) / 0.5) !important;
}

.relative:focus-within .absolute.input-icon {
  color: hsl(var(--bc) / 0.75) !important;
}

/* 针对具体的图标元素 */
.relative .absolute[class*="text-base-content"] {
  color: hsl(var(--bc) / 0.5) !important;
}

.relative:focus-within .absolute[class*="text-base-content"] {
  color: hsl(var(--bc) / 0.75) !important;
}

/* 登录按钮动画 */
.btn:not(:disabled):hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px hsl(var(--p) / 0.3);
}

/* 背景动画 */
@keyframes float {
  0%, 100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-10px);
  }
}

.card {
  animation: float 6s ease-in-out infinite;
}
</style>