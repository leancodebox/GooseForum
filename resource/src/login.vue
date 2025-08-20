<template>
  <div class="flex items-center justify-center min-h-[70vh] py-8">
    <div class="card w-full max-w-md bg-base-100 shadow-xl">
      <div class="card-body">
        <!-- Tab 切换 -->
        <div class="tabs tabs-boxed mb-6">
          <a class="tab" :class="{ 'tab-active': activeTab === 'login' }" @click="activeTab = 'login'">登录</a>
          <a class="tab" :class="{ 'tab-active': activeTab === 'register' }" @click="activeTab = 'register'">注册</a>
          <a class="tab" :class="{ 'tab-active': activeTab === 'forgot' }" @click="activeTab = 'forgot'">忘记密码</a>
        </div>

        <!-- 登录表单 -->
        <div v-if="activeTab === 'login'">
          <h2 class="text-2xl font-normal text-base-content mb-6 text-center">欢迎回来</h2>
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
                    class="h-10 cursor-pointer border border-base-300 rounded bg-white"
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
            <div class="text-center mt-4">
              <a href="#" class="link link-primary text-sm" @click.prevent="activeTab = 'forgot'">忘记密码？</a>
            </div>
          </form>
        </div>

        <!-- 注册表单 -->
        <div v-if="activeTab === 'register'">
          <h2 class="text-2xl font-normal text-base-content mb-6 text-center">创建账户</h2>
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
                <span class="label-text ml-2">我同意 <a href="/terms-of-service" target="_blank" class="link link-primary">用户协议</a> 和 <a href="/privacy-policy" target="_blank" class="link link-primary">隐私政策</a></span>
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

        <!-- 忘记密码表单 -->
        <div v-if="activeTab === 'forgot'">
          <h2 class="text-2xl font-normal text-base-content mb-6 text-center">重置密码</h2>
          <form @submit.prevent="handleForgotPassword" class="space-y-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text">邮箱地址</span>
              </label>
              <input
                  v-model="forgotForm.email"
                  type="email"
                  placeholder="请输入注册时使用的邮箱地址"
                  class="input input-bordered w-full"
                  :class="{ 'input-error': forgotErrors.email }"
                  required
              />
              <label v-if="forgotErrors.email" class="label">
                <span class="label-text-alt text-error">{{ forgotErrors.email }}</span>
              </label>
            </div>
            <div class="form-control">
              <label class="label">
                <span class="label-text">验证码</span>
              </label>
              <div class="flex gap-2">
                <input
                    v-model="forgotForm.captcha"
                    type="text"
                    placeholder="请输入验证码"
                    class="input input-bordered flex-1"
                    :class="{ 'input-error': forgotErrors.captcha }"
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
              <label v-if="forgotErrors.captcha" class="label">
                <span class="label-text-alt text-error">{{ forgotErrors.captcha }}</span>
              </label>
            </div>
            <div v-if="forgotErrors.general" class="alert alert-error mb-4">
              <span>{{ forgotErrors.general }}</span>
            </div>
            <div v-if="forgotSuccess" class="alert alert-success mb-4">
              <span>{{ forgotSuccess }}</span>
            </div>
            <div class="form-control mt-6">
              <button type="submit" class="btn btn-primary w-full" :disabled="forgotLoading">
                <span v-if="forgotLoading" class="loading loading-spinner loading-sm"></span>
                {{ forgotLoading ? '发送中...' : '发送重置邮件' }}
              </button>
            </div>
            <div class="text-center mt-4">
              <a href="#" class="link link-primary text-sm" @click.prevent="activeTab = 'login'">返回登录</a>
            </div>
          </form>
        </div>

        <!-- 第三方登录 -->
        <div v-if="activeTab !== 'forgot'" class="divider">或</div>
        <div v-if="activeTab !== 'forgot'" class="space-y-2">
          <button @click="handleGitHubLogin" class="btn btn-outline w-full">
            <svg class="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
            </svg>
            使用 GitHub 登录
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getCaptcha as getCaptchaApi, login, register, forgotPassword } from '@/utils/gooseForumService.ts'

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

// 忘记密码表单数据
const forgotForm = reactive({
  email: '',
  captcha: ''
})

// 错误信息
const loginErrors = reactive({})
const registerErrors = reactive({})
const forgotErrors = reactive({})

// 加载状态
const loginLoading = ref(false)
const registerLoading = ref(false)
const forgotLoading = ref(false)

// 忘记密码成功信息
const forgotSuccess = ref('')

// 获取验证码
const getCaptcha = async () => {
  try {
    const data = await getCaptchaApi()
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
    const data = await login(loginForm.username, loginForm.password, captchaId.value, loginForm.captcha)

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
    loginErrors.general = '登录失败，请稍后重试'+error
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
    const data = await register(registerForm.username, registerForm.email, registerForm.password, captchaId.value, registerForm.captcha)

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
    registerErrors.general = '注册失败，请稍后重试'+error
    refreshCaptcha()
  } finally {
    registerLoading.value = false
  }
}

// 处理忘记密码
const handleForgotPassword = async () => {
  // 清空之前的错误和成功信息
  Object.keys(forgotErrors).forEach(key => delete forgotErrors[key])
  forgotSuccess.value = ''

  // 验证表单
  const errors = {}
  if (!forgotForm.email.trim()) {
    errors.email = '请输入邮箱地址'
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(forgotForm.email)) {
    errors.email = '请输入有效的邮箱地址'
  }

  if (!forgotForm.captcha.trim()) {
    errors.captcha = '请输入验证码'
  }

  if (Object.keys(errors).length > 0) {
    Object.assign(forgotErrors, errors)
    return
  }

  forgotLoading.value = true

  try {
    const data = await forgotPassword(forgotForm.email, captchaId.value, forgotForm.captcha)

    if (data.code === 0) {
      forgotSuccess.value = data.result
      // 清空表单
      forgotForm.captcha = ''
    } else {
      forgotErrors.general = data.msg || '发送失败，请稍后重试'
    }
    refreshCaptcha()
  } catch (error) {
    console.error('忘记密码请求失败:', error)
    forgotErrors.general = '发送失败，请稍后重试'+error
    refreshCaptcha()
  } finally {
    forgotLoading.value = false
  }
}

// 处理GitHub登录
const handleGitHubLogin = () => {
  // 保存当前页面的重定向参数
  const urlParams = new URLSearchParams(window.location.search)
  const redirectUrl = urlParams.get('redirect')

  // 构建GitHub OAuth登录URL
  let githubLoginUrl = '/api/auth/github'
  if (redirectUrl) {
    githubLoginUrl += '?redirect=' + encodeURIComponent(redirectUrl)
  }

  // 跳转到GitHub OAuth登录
  window.location.href = githubLoginUrl
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
