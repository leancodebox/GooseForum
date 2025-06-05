<template>
  <div class="flex items-center justify-center min-h-[70vh] py-8">
    <div class="card w-full max-w-md bg-base-100 shadow-xl">
      <div class="card-body">
        <!-- Tab 切换 -->
        <div class="tabs tabs-boxed mb-6">
          <a class="tab" :class="{ 'tab-active': activeTab === 'login' }" @click="activeTab = 'login'">登录</a>
          <a class="tab" :class="{ 'tab-active': activeTab === 'register' }" @click="activeTab = 'register'">注册</a>
        </div>

        <!-- 登录表单 -->
        <div v-if="activeTab === 'login'">
          <h2 class="text-2xl font-bold text-base-content mb-6 text-center">欢迎回来</h2>
          <form @submit.prevent="handleLogin" class="space-y-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text">用户名</span>
              </label>
              <input
                  v-model="loginForm.username"
                  type="text"
                  placeholder="请输入用户名"
                  class="input input-bordered w-full"
                  :class="{ 'input-error': loginErrors.username }"
                  required
              />
              <label v-if="loginErrors.username" class="label">
                <span class="label-text-alt text-error">{{ loginErrors.username }}</span>
              </label>
            </div>
            <div class="form-control">
              <label class="label">
                <span class="label-text">密码</span>
              </label>
              <input
                  v-model="loginForm.password"
                  type="password"
                  placeholder="请输入密码"
                  class="input input-bordered w-full"
                  :class="{ 'input-error': loginErrors.password }"
                  required
              />
              <label v-if="loginErrors.password" class="label">
                <span class="label-text-alt text-error">{{ loginErrors.password }}</span>
              </label>
            </div>
            <div class="form-control">
              <label class="label">
                <span class="label-text">验证码</span>
              </label>
              <div class="flex gap-2">
                <input
                    v-model="loginForm.captcha"
                    type="text"
                    placeholder="请输入验证码"
                    class="input input-bordered flex-1"
                    :class="{ 'input-error': loginErrors.captcha }"
                    required
                />
                <img
                    v-if="captchaImg"
                    :src="captchaImg"
                    alt="验证码"
                    class="h-12 cursor-pointer border border-base-300 rounded"
                    @click="refreshCaptcha"
                />
              </div>
              <label v-if="loginErrors.captcha" class="label">
                <span class="label-text-alt text-error">{{ loginErrors.captcha }}</span>
              </label>
            </div>
            <div v-if="loginErrors.general" class="alert alert-error mb-4">
              <span>{{ loginErrors.general }}</span>
            </div>
            <div class="form-control mt-6">
              <button type="submit" class="btn btn-primary w-full" :disabled="loginLoading">
                <span v-if="loginLoading" class="loading loading-spinner loading-sm"></span>
                {{ loginLoading ? '登录中...' : '登录' }}
              </button>
            </div>
            <div class="text-center">
              <a href="#" class="link link-primary text-sm">忘记密码？</a>
            </div>
          </form>
        </div>

        <!-- 注册表单 -->
        <div v-if="activeTab === 'register'">
          <h2 class="text-2xl font-bold text-base-content mb-6 text-center">创建账户</h2>
          <form @submit.prevent="handleRegister" class="space-y-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text">用户名</span>
              </label>
              <input
                  v-model="registerForm.username"
                  type="text"
                  placeholder="请输入用户名"
                  class="input input-bordered w-full"
                  :class="{ 'input-error': registerErrors.username }"
                  required
              />
              <label v-if="registerErrors.username" class="label">
                <span class="label-text-alt text-error">{{ registerErrors.username }}</span>
              </label>
            </div>
            <div class="form-control">
              <label class="label">
                <span class="label-text">邮箱</span>
              </label>
              <input
                  v-model="registerForm.email"
                  type="email"
                  placeholder="请输入邮箱地址"
                  class="input input-bordered w-full"
                  :class="{ 'input-error': registerErrors.email }"
                  required
              />
              <label v-if="registerErrors.email" class="label">
                <span class="label-text-alt text-error">{{ registerErrors.email }}</span>
              </label>
            </div>
            <div class="form-control">
              <label class="label">
                <span class="label-text">密码</span>
              </label>
              <input
                  v-model="registerForm.password"
                  type="password"
                  placeholder="请输入密码（至少6位）"
                  class="input input-bordered w-full"
                  :class="{ 'input-error': registerErrors.password }"
                  required
              />
              <label v-if="registerErrors.password" class="label">
                <span class="label-text-alt text-error">{{ registerErrors.password }}</span>
              </label>
            </div>
            <div class="form-control">
              <label class="label">
                <span class="label-text">确认密码</span>
              </label>
              <input
                  v-model="registerForm.confirmPassword"
                  type="password"
                  placeholder="请再次输入密码"
                  class="input input-bordered w-full"
                  :class="{ 'input-error': registerErrors.confirmPassword }"
                  required
              />
              <label v-if="registerErrors.confirmPassword" class="label">
                <span class="label-text-alt text-error">{{ registerErrors.confirmPassword }}</span>
              </label>
            </div>
            <div class="form-control">
              <label class="label">
                <span class="label-text">验证码</span>
              </label>
              <div class="flex gap-2">
                <input
                    v-model="registerForm.captcha"
                    type="text"
                    placeholder="请输入验证码"
                    class="input input-bordered flex-1"
                    :class="{ 'input-error': registerErrors.captcha }"
                    required
                />
                <img
                    v-if="captchaImg"
                    :src="captchaImg"
                    alt="验证码"
                    class="h-12 cursor-pointer border border-base-300 rounded"
                    @click="refreshCaptcha"
                />
              </div>
              <label v-if="registerErrors.captcha" class="label">
                <span class="label-text-alt text-error">{{ registerErrors.captcha }}</span>
              </label>
            </div>
            <div class="form-control">
              <label class="cursor-pointer label justify-start">
                <input v-model="registerForm.agree" type="checkbox" class="checkbox checkbox-primary" required />
                <span class="label-text ml-2">我同意 <a href="#" class="link link-primary">用户协议</a> 和 <a href="#" class="link link-primary">隐私政策</a></span>
              </label>
            </div>
            <div v-if="registerErrors.general" class="alert alert-error mb-4">
              <span>{{ registerErrors.general }}</span>
            </div>
            <div class="form-control mt-6">
              <button type="submit" class="btn btn-primary w-full" :disabled="registerLoading">
                <span v-if="registerLoading" class="loading loading-spinner loading-sm"></span>
                {{ registerLoading ? '注册中...' : '注册' }}
              </button>
            </div>
          </form>
        </div>

        <!-- 第三方登录 -->
        <div class="divider">或</div>
        <div class="space-y-2">
          <button class="btn btn-outline w-full">
            <svg class="w-5 h-5 mr-2" viewBox="0 0 24 24">
              <path fill="currentColor" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
              <path fill="currentColor" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
              <path fill="currentColor" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
              <path fill="currentColor" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
            </svg>
            使用 Google 登录
          </button>
          <button class="btn btn-outline w-full">
            <svg class="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 24 24">
              <path d="M24 4.557c-.883.392-1.832.656-2.828.775 1.017-.609 1.798-1.574 2.165-2.724-.951.564-2.005.974-3.127 1.195-.897-.957-2.178-1.555-3.594-1.555-3.179 0-5.515 2.966-4.797 6.045-4.091-.205-7.719-2.165-10.148-5.144-1.29 2.213-.669 5.108 1.523 6.574-.806-.026-1.566-.247-2.229-.616-.054 2.281 1.581 4.415 3.949 4.89-.693.188-1.452.232-2.224.084.626 1.956 2.444 3.379 4.6 3.419-2.07 1.623-4.678 2.348-7.29 2.04 2.179 1.397 4.768 2.212 7.548 2.212 9.142 0 14.307-7.721 13.995-14.646.962-.695 1.797-1.562 2.457-2.549z"/>
            </svg>
            使用 Twitter 登录
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'

// 当前激活的标签页
const activeTab = ref('login')

// 验证码相关
const captchaImg = ref('')
const captchaId = ref('')

// 登录表单数据
const loginForm = reactive({
  username: '',
  password: '',
  captcha: ''
})

// 注册表单数据
const registerForm = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
  captcha: '',
  agree: false
})

// 错误信息
const loginErrors = reactive({})
const registerErrors = reactive({})

// 加载状态
const loginLoading = ref(false)
const registerLoading = ref(false)

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
  getCaptcha()
}

// 表单验证函数
const validateLogin = () => {
  const errors = {}

  if (!loginForm.username.trim()) {
    errors.username = '请输入用户名'
  }

  if (!loginForm.password.trim()) {
    errors.password = '请输入密码'
  }

  if (!loginForm.captcha.trim()) {
    errors.captcha = '请输入验证码'
  }

  Object.assign(loginErrors, errors)
  return Object.keys(errors).length === 0
}

const validateRegister = () => {
  const errors = {}

  if (!registerForm.username.trim()) {
    errors.username = '请输入用户名'
  }

  if (!registerForm.email.trim()) {
    errors.email = '请输入邮箱地址'
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(registerForm.email)) {
    errors.email = '请输入有效的邮箱地址'
  }

  if (!registerForm.password.trim()) {
    errors.password = '请输入密码'
  }

  if (!registerForm.confirmPassword.trim()) {
    errors.confirmPassword = '请确认密码'
  } else if (registerForm.password !== registerForm.confirmPassword) {
    errors.confirmPassword = '两次输入的密码不一致'
  }

  if (!registerForm.captcha.trim()) {
    errors.captcha = '请输入验证码'
  }

  Object.assign(registerErrors, errors)
  return Object.keys(errors).length === 0
}

// 处理登录
const handleLogin = async () => {
  // 清空之前的错误
  Object.keys(loginErrors).forEach(key => delete loginErrors[key])

  if (!validateLogin()) {
    return
  }

  loginLoading.value = true

  try {
    const response = await fetch('/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: loginForm.username,
        password: loginForm.password,
        captchaId: captchaId.value,
        captchaCode: loginForm.captcha
      })
    })

    const data = await response.json()
    
    if (data.code === 0) {
      // 获取重定向参数，如果存在则跳转到原始页面，否则跳转到首页
      const urlParams = new URLSearchParams(window.location.search)
      const redirectUrl = urlParams.get('redirect')
      window.location.href = redirectUrl || '/'
    } else {
      loginErrors.general = data.message || '登录失败，请检查用户名和密码'
      refreshCaptcha()
    }
  } catch (error) {
    console.error('登录请求失败:', error)
    loginErrors.general = '登录失败，请稍后重试'
    refreshCaptcha()
  } finally {
    loginLoading.value = false
  }
}

// 处理注册
const handleRegister = async () => {
  // 清空之前的错误
  Object.keys(registerErrors).forEach(key => delete registerErrors[key])

  if (!validateRegister()) {
    return
  }

  registerLoading.value = true

  try {
    const response = await fetch('/register', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: registerForm.username,
        email: registerForm.email,
        password: registerForm.password,
        captchaId: captchaId.value,
        captchaCode: registerForm.captcha
      })
    })

    const data = await response.json()

    if (data.code === 0) {
      // 获取重定向参数，如果存在则跳转到原始页面，否则跳转到首页
      const urlParams = new URLSearchParams(window.location.search)
      const redirectUrl = urlParams.get('redirect')
      window.location.href = redirectUrl || '/'
    } else {
      registerErrors.general = data.msg || '注册失败，请稍后重试'
      refreshCaptcha()
    }
  } catch (error) {
    console.error('注册请求失败:', error)
    registerErrors.general = '注册失败，请稍后重试'
    refreshCaptcha()
  } finally {
    registerLoading.value = false
  }
}

// 页面加载时获取验证码
onMounted(() => {
  getCaptcha()
  
  // 检查 URL 参数并切换到相应的标签
  const urlParams = new URLSearchParams(window.location.search)
  const model = urlParams.get('model')
  if (model === 'register') {
    activeTab.value = 'register'
  }
})
</script>