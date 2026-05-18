<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Languages, LoaderCircle, LockKeyhole, Mail, UserRound } from '@lucide/vue'
import { forgotPassword, getCaptcha, login, register } from '../runtime/api'
import { useI18n } from '../runtime/i18n'
import type { LayoutPayload, LoginPageProps } from '../types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: LoginPageProps
}>()

type Mode = 'login' | 'register' | 'forgot'

const { locale } = useI18n()
const mode = ref<Mode>(page.props.initialMode || 'login')
const captchaImg = ref('')
const captchaId = ref('')
const captchaLoading = ref(false)
const notice = ref('')
const error = ref('')

const loading = reactive({
  login: false,
  register: false,
  forgot: false,
})

const loginForm = reactive({
  username: '',
  password: '',
  captcha: '',
})

const registerForm = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
  captcha: '',
  agree: false,
})

const forgotForm = reactive({
  email: '',
  captcha: '',
})

const title = computed(() => {
  if (mode.value === 'register') return '创建新账号'
  if (mode.value === 'forgot') return '重置密码'
  return '登录账号'
})

const subtitle = computed(() => {
  if (mode.value === 'register') return '加入 GooseForum，开启你的讨论空间。'
  if (mode.value === 'forgot') return '输入邮箱，我们会发送一封重置密码邮件。'
  return '欢迎回来，继续你的讨论和创作。'
})

const showSocial = computed(() => mode.value !== 'forgot')
const homeUrl = computed(() => page.props.redirectUrl || '/')

onMounted(() => {
  refreshCaptcha()
})

function switchMode(next: Mode) {
  mode.value = next
  error.value = ''
  notice.value = ''
}

async function refreshCaptcha() {
  captchaLoading.value = true
  try {
    const captcha = await getCaptcha()
    captchaId.value = captcha.captchaId
    captchaImg.value = captcha.captchaImg
  } catch (err) {
    error.value = errorMessage(err, '验证码加载失败')
  } finally {
    captchaLoading.value = false
  }
}

async function handleLogin() {
  if (!loginForm.username || !loginForm.password || !loginForm.captcha) {
    error.value = '请填写账号、密码和验证码'
    return
  }
  loading.login = true
  error.value = ''
  try {
    await login(loginForm.username, loginForm.password, captchaId.value, loginForm.captcha)
    window.location.href = homeUrl.value
  } catch (err) {
    error.value = errorMessage(err, '登录失败')
    loginForm.captcha = ''
    refreshCaptcha()
  } finally {
    loading.login = false
  }
}

async function handleRegister() {
  if (!registerForm.username || !registerForm.email || !registerForm.password || !registerForm.captcha) {
    error.value = '请完整填写注册信息'
    return
  }
  if (registerForm.password !== registerForm.confirmPassword) {
    error.value = '两次输入的密码不一致'
    return
  }
  if (!registerForm.agree) {
    error.value = '请先同意服务条款和隐私政策'
    return
  }
  loading.register = true
  error.value = ''
  try {
    const message = await register(registerForm.username, registerForm.email, registerForm.password, captchaId.value, registerForm.captcha)
    if (message.includes('邮箱') || message.includes('验证')) {
      notice.value = message
      switchMode('login')
      return
    }
    window.location.href = homeUrl.value
  } catch (err) {
    error.value = errorMessage(err, '注册失败')
    registerForm.captcha = ''
    refreshCaptcha()
  } finally {
    loading.register = false
  }
}

async function handleForgot() {
  if (!forgotForm.email || !forgotForm.captcha) {
    error.value = '请填写邮箱和验证码'
    return
  }
  loading.forgot = true
  error.value = ''
  try {
    notice.value = await forgotPassword(forgotForm.email, captchaId.value, forgotForm.captcha)
    forgotForm.captcha = ''
    refreshCaptcha()
  } catch (err) {
    error.value = errorMessage(err, '重置邮件发送失败')
    forgotForm.captcha = ''
    refreshCaptcha()
  } finally {
    loading.forgot = false
  }
}

function setLocale(next: 'zh' | 'en' | 'ja') {
  document.cookie = `lang=${next}; path=/; max-age=31536000; samesite=lax`
  window.location.reload()
}

function errorMessage(err: unknown, fallback: string) {
  return err instanceof Error && err.message ? err.message : fallback
}
</script>

<template>
  <main class="relative min-h-screen bg-gray-50 px-4 py-8 text-gray-950 sm:px-6 lg:px-8">
    <div class="absolute right-4 top-4 z-10 flex rounded-full border border-gray-200 bg-white p-1 shadow-sm">
      <button
        v-for="item in [
          { key: 'zh', label: '中' },
          { key: 'en', label: 'EN' },
          { key: 'ja', label: '日' },
        ]"
        :key="item.key"
        type="button"
        class="inline-flex h-8 min-w-8 items-center justify-center rounded-full px-2 text-xs font-semibold text-gray-500 hover:bg-gray-100 hover:text-gray-900"
        :class="{ 'bg-gray-900 text-white hover:bg-gray-900 hover:text-white': locale === item.key }"
        @click="setLocale(item.key as 'zh' | 'en' | 'ja')"
      >
        {{ item.label }}
      </button>
    </div>

    <section class="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-[880px] items-center justify-center">
      <div class="grid w-full overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-[0_18px_50px_rgba(15,23,42,0.07)] md:grid-cols-2">
        <div class="flex min-h-[470px] flex-col justify-center px-5 py-6 sm:px-8">
          <a href="/" class="mb-6 inline-flex items-baseline text-[27px] font-semibold leading-none tracking-[-0.04em] text-blue-600">
            <span v-if="page.layout.site.brandType === 'image' && page.layout.site.brandImage" class="inline-flex">
              <img :src="page.layout.site.brandImage" :alt="page.layout.site.name" class="h-8 w-auto object-contain" />
            </span>
            <span v-else-if="page.layout.site.brandType === 'text'">
              {{ page.layout.site.brandText || page.layout.site.name }}
            </span>
            <span v-else>
              Goose<span class="text-gray-950">Forum</span>
            </span>
          </a>

          <div class="mb-4">
            <h1 class="text-[27px] font-bold leading-tight tracking-tight text-gray-950">{{ title }}</h1>
            <p class="mt-1.5 text-sm leading-6 text-gray-500">{{ subtitle }}</p>
          </div>

          <div v-if="mode !== 'forgot'" class="mb-4 grid grid-cols-2 rounded-xl bg-gray-100 p-1">
            <button type="button" class="h-8 rounded-lg text-sm font-semibold transition" :class="mode === 'login' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:text-gray-900'" @click="switchMode('login')">登录</button>
            <button type="button" class="h-8 rounded-lg text-sm font-semibold transition" :class="mode === 'register' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:text-gray-900'" @click="switchMode('register')">注册</button>
          </div>

          <p v-if="error" class="mb-4 rounded-lg border border-red-100 bg-red-50 px-3 py-2 text-sm font-medium text-red-700">{{ error }}</p>
          <p v-if="notice" class="mb-4 rounded-lg border border-emerald-100 bg-emerald-50 px-3 py-2 text-sm font-medium text-emerald-700">{{ notice }}</p>

          <form v-if="mode === 'login'" class="space-y-3.5" @submit.prevent="handleLogin">
            <label class="block">
              <span class="sr-only">用户名或邮箱</span>
              <span class="relative block">
                <UserRound class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                <input v-model.trim="loginForm.username" class="h-10 w-full rounded-xl border border-gray-200 bg-white pl-10 pr-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100" placeholder="用户名或邮箱" autocomplete="username" />
              </span>
            </label>
            <label class="block">
              <span class="sr-only">密码</span>
              <span class="relative block">
                <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                <input v-model="loginForm.password" type="password" class="h-10 w-full rounded-xl border border-gray-200 bg-white pl-10 pr-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100" placeholder="密码" autocomplete="current-password" />
              </span>
            </label>
            <div class="flex gap-3">
              <input v-model.trim="loginForm.captcha" class="h-10 min-w-0 flex-1 rounded-xl border border-gray-200 bg-white px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100" placeholder="验证码" />
              <button type="button" class="relative h-10 w-28 overflow-hidden rounded-xl border border-gray-200 bg-gray-50" @click="refreshCaptcha">
                <LoaderCircle v-if="captchaLoading || !captchaImg" class="mx-auto h-5 w-5 animate-spin text-gray-400" />
                <img v-else :src="captchaImg" alt="验证码" class="h-full w-full object-cover" />
              </button>
            </div>
            <div class="flex justify-end">
              <button type="button" class="text-sm font-medium text-blue-600 hover:text-blue-700" @click="switchMode('forgot')">忘记密码？</button>
            </div>
            <button type="submit" class="inline-flex h-11 w-full items-center justify-center gap-2 rounded-xl bg-blue-600 text-sm font-bold text-white shadow-lg shadow-blue-200 transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-70" :disabled="loading.login">
              <LoaderCircle v-if="loading.login" class="h-4 w-4 animate-spin" />
              登录
            </button>
          </form>

          <form v-else-if="mode === 'register'" class="space-y-3.5" @submit.prevent="handleRegister">
            <input v-model.trim="registerForm.username" class="h-10 w-full rounded-xl border border-gray-200 bg-white px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100" placeholder="用户名" autocomplete="username" />
            <input v-model.trim="registerForm.email" type="email" class="h-10 w-full rounded-xl border border-gray-200 bg-white px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100" placeholder="邮箱" autocomplete="email" />
            <input v-model="registerForm.password" type="password" class="h-10 w-full rounded-xl border border-gray-200 bg-white px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100" placeholder="密码" autocomplete="new-password" />
            <input v-model="registerForm.confirmPassword" type="password" class="h-10 w-full rounded-xl border border-gray-200 bg-white px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100" placeholder="确认密码" autocomplete="new-password" />
            <div class="flex gap-3">
              <input v-model.trim="registerForm.captcha" class="h-10 min-w-0 flex-1 rounded-xl border border-gray-200 bg-white px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100" placeholder="验证码" />
              <button type="button" class="relative h-10 w-28 overflow-hidden rounded-xl border border-gray-200 bg-gray-50" @click="refreshCaptcha">
                <LoaderCircle v-if="captchaLoading || !captchaImg" class="mx-auto h-5 w-5 animate-spin text-gray-400" />
                <img v-else :src="captchaImg" alt="验证码" class="h-full w-full object-cover" />
              </button>
            </div>
            <label class="flex items-start gap-2 text-sm leading-5 text-gray-500">
              <input v-model="registerForm.agree" type="checkbox" class="mt-1 h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500" />
              <span>我已阅读并同意服务条款和隐私政策</span>
            </label>
            <button type="submit" class="inline-flex h-11 w-full items-center justify-center gap-2 rounded-xl bg-gray-950 text-sm font-bold text-white shadow-lg shadow-gray-200 transition hover:bg-gray-800 disabled:cursor-not-allowed disabled:opacity-70" :disabled="loading.register">
              <LoaderCircle v-if="loading.register" class="h-4 w-4 animate-spin" />
              创建账号
            </button>
          </form>

          <form v-else class="space-y-3.5" @submit.prevent="handleForgot">
            <label class="block">
              <span class="relative block">
                <Mail class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                <input v-model.trim="forgotForm.email" type="email" class="h-10 w-full rounded-xl border border-gray-200 bg-white pl-10 pr-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100" placeholder="注册邮箱" autocomplete="email" />
              </span>
            </label>
            <div class="flex gap-3">
              <input v-model.trim="forgotForm.captcha" class="h-10 min-w-0 flex-1 rounded-xl border border-gray-200 bg-white px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100" placeholder="验证码" />
              <button type="button" class="relative h-10 w-28 overflow-hidden rounded-xl border border-gray-200 bg-gray-50" @click="refreshCaptcha">
                <LoaderCircle v-if="captchaLoading || !captchaImg" class="mx-auto h-5 w-5 animate-spin text-gray-400" />
                <img v-else :src="captchaImg" alt="验证码" class="h-full w-full object-cover" />
              </button>
            </div>
            <button type="submit" class="inline-flex h-11 w-full items-center justify-center gap-2 rounded-xl bg-blue-600 text-sm font-bold text-white shadow-lg shadow-blue-200 transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-70" :disabled="loading.forgot">
              <LoaderCircle v-if="loading.forgot" class="h-4 w-4 animate-spin" />
              发送重置邮件
            </button>
            <button type="button" class="w-full text-sm font-medium text-blue-600 hover:text-blue-700" @click="switchMode('login')">返回登录</button>
          </form>
        </div>

        <aside class="border-t border-gray-100 bg-gray-50/70 px-5 py-6 sm:px-8 md:border-l md:border-t-0">
          <div class="flex h-full flex-col justify-center">
            <div v-if="showSocial">
              <h2 class="text-sm font-bold text-gray-950">或继续使用</h2>
              <div class="mt-8 space-y-3.5">
                <a :href="page.props.githubUrl" class="flex h-10 w-full items-center justify-center gap-2 rounded-xl border border-gray-200 bg-white text-sm font-semibold text-gray-800 shadow-sm transition hover:bg-gray-50">
                  <svg class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                    <path d="M12 0C5.37 0 0 5.37 0 12c0 5.3 3.44 9.8 8.21 11.39.6.11.79-.26.79-.58v-2.03c-3.34.73-4.04-1.42-4.04-1.42-.55-1.39-1.34-1.76-1.34-1.76-1.09-.75.08-.73.08-.73 1.21.08 1.85 1.24 1.85 1.24 1.07 1.83 2.81 1.3 3.49 1 .11-.78.42-1.3.76-1.6-2.67-.31-5.47-1.34-5.47-5.93 0-1.31.47-2.38 1.24-3.22-.12-.3-.54-1.52.12-3.18 0 0 1.01-.32 3.3 1.23A11.5 11.5 0 0 1 12 6c1.02 0 2.05.14 3.01.4 2.29-1.55 3.3-1.23 3.3-1.23.65 1.66.24 2.88.12 3.18.77.84 1.24 1.91 1.24 3.22 0 4.61-2.81 5.62-5.48 5.92.43.37.81 1.1.81 2.22v3.29c0 .32.19.69.8.58A12.01 12.01 0 0 0 24 12c0-6.63-5.37-12-12-12Z" />
                  </svg>
                  GitHub
                </a>
                <button type="button" class="flex h-10 w-full cursor-not-allowed items-center justify-center gap-2 rounded-xl border border-gray-200 bg-white text-sm font-semibold text-gray-400 opacity-70">
                  Google 暂未开放
                </button>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </section>
  </main>
</template>
