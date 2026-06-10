<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Languages, LoaderCircle, LockKeyhole, Mail, UserRound } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { forgotPassword, getCaptcha, login, register } from '@/runtime/api'
import { queueFlashMessage } from '@/runtime/flash-message'
import { setLocale, supportedLocales, type Locale } from '@/runtime/i18n'
import type { LayoutPayload, LoginPageProps } from '@/types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: LoginPageProps
}>()

type Mode = 'login' | 'register' | 'forgot'

const { t, locale } = useI18n()
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
  if (mode.value === 'register') return t('auth.registerTitle')
  if (mode.value === 'forgot') return t('auth.forgotTitle')
  return t('auth.loginTitle')
})

const subtitle = computed(() => {
  if (mode.value === 'register') return t('auth.registerSubtitle')
  if (mode.value === 'forgot') return t('auth.forgotSubtitle')
  return t('auth.loginSubtitle')
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
    error.value = errorMessage(err, t('auth.validation.captchaLoadFailed'))
  } finally {
    captchaLoading.value = false
  }
}

async function handleLogin() {
  if (!loginForm.username || !loginForm.password || !loginForm.captcha) {
    error.value = t('auth.validation.loginRequired')
    return
  }
  loading.login = true
  error.value = ''
  try {
    await login(loginForm.username, loginForm.password, captchaId.value, loginForm.captcha)
    window.location.href = homeUrl.value
  } catch (err) {
    error.value = errorMessage(err, t('auth.validation.loginFailed'))
    loginForm.captcha = ''
    refreshCaptcha()
  } finally {
    loading.login = false
  }
}

async function handleRegister() {
  if (!registerForm.username || !registerForm.email || !registerForm.password || !registerForm.captcha) {
    error.value = t('auth.validation.registerRequired')
    return
  }
  if (registerForm.password !== registerForm.confirmPassword) {
    error.value = t('auth.validation.passwordMismatch')
    return
  }
  if (!registerForm.agree) {
    error.value = t('auth.validation.termsRequired')
    return
  }
  loading.register = true
  error.value = ''
  try {
    const message = await register(registerForm.username, registerForm.email, registerForm.password, captchaId.value, registerForm.captcha)
    queueFlashMessage(message || t('auth.validation.registerSuccess'), 'success')
    window.location.href = homeUrl.value
  } catch (err) {
    error.value = errorMessage(err, t('auth.validation.registerFailed'))
    registerForm.captcha = ''
    refreshCaptcha()
  } finally {
    loading.register = false
  }
}

async function handleForgot() {
  if (!forgotForm.email || !forgotForm.captcha) {
    error.value = t('auth.validation.forgotRequired')
    return
  }
  loading.forgot = true
  error.value = ''
  try {
    notice.value = await forgotPassword(forgotForm.email, captchaId.value, forgotForm.captcha)
    forgotForm.captcha = ''
    refreshCaptcha()
  } catch (err) {
    error.value = errorMessage(err, t('auth.validation.resetEmailFailed'))
    forgotForm.captcha = ''
    refreshCaptcha()
  } finally {
    loading.forgot = false
  }
}

function switchLocale(next: Locale) {
  setLocale(next)
}

function errorMessage(err: unknown, fallback: string) {
  return err instanceof Error && err.message ? err.message : fallback
}
</script>

<template>
  <main class="relative min-h-screen bg-base-200 px-4 py-8 text-base-content sm:px-6 lg:px-8">
    <div class="absolute right-4 top-4 z-10 flex rounded-full border border-line bg-base-100 p-1 shadow-sm">
      <button
        v-for="item in supportedLocales"
        :key="item"
        type="button"
        class="inline-flex h-8 min-w-8 items-center justify-center rounded-full px-2 text-xs font-semibold text-base-content/55 hover:bg-base-300 hover:text-base-content"
        :class="{ 'bg-neutral text-neutral-content hover:bg-neutral hover:text-neutral-content': locale === item }"
        @click="switchLocale(item)"
      >
        {{ t(`locale.short.${item}`) }}
      </button>
    </div>

    <section class="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-[880px] items-center justify-center">
      <div class="grid w-full overflow-hidden rounded-2xl border border-line bg-base-100 shadow-[0_18px_50px_rgba(15,23,42,0.07)] md:grid-cols-2">
        <div class="flex min-h-[470px] flex-col justify-center px-5 py-6 sm:px-8">
          <a href="/" class="mb-6 inline-flex items-baseline text-[27px] font-semibold leading-none tracking-[-0.04em] text-primary">
            <span v-if="page.layout.site.brandType === 'image' && page.layout.site.brandImage" class="inline-flex">
              <img :src="page.layout.site.brandImage" :alt="page.layout.site.name" class="h-8 w-auto object-contain" />
            </span>
            <span v-else-if="page.layout.site.brandType === 'text'">
              {{ page.layout.site.brandText || page.layout.site.name }}
            </span>
            <span v-else>
              Goose<span class="text-base-content">Forum</span>
            </span>
          </a>

          <div class="mb-4">
            <h1 class="text-[27px] font-bold leading-tight tracking-tight text-base-content">{{ title }}</h1>
            <p class="mt-1.5 text-sm leading-6 text-base-content/55">{{ subtitle }}</p>
          </div>

          <div v-if="mode !== 'forgot'" class="mb-4 grid grid-cols-2 rounded-xl bg-base-300 p-1">
            <button type="button" class="h-8 rounded-lg text-sm font-semibold transition" :class="mode === 'login' ? 'bg-base-100 text-primary shadow-sm' : 'text-base-content/55 hover:text-base-content'" @click="switchMode('login')">{{ t('shell.login') }}</button>
            <button type="button" class="h-8 rounded-lg text-sm font-semibold transition" :class="mode === 'register' ? 'bg-base-100 text-primary shadow-sm' : 'text-base-content/55 hover:text-base-content'" @click="switchMode('register')">{{ t('shell.register') }}</button>
          </div>

          <p v-if="error" class="mb-4 rounded-lg border border-error/20 bg-error/10 px-3 py-2 text-sm font-medium text-error">{{ error }}</p>
          <p v-if="notice" class="mb-4 rounded-lg border border-success/20 bg-success/10 px-3 py-2 text-sm font-medium text-success">{{ notice }}</p>

          <form v-if="mode === 'login'" class="space-y-3.5" @submit.prevent="handleLogin">
            <label class="block">
              <span class="sr-only">{{ t('auth.usernameOrEmail') }}</span>
              <span class="relative block">
                <UserRound class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model.trim="loginForm.username" class="h-10 w-full rounded-xl border border-line bg-base-100 pl-10 pr-3 text-sm outline-none transition focus:border-primary focus:ring-4 focus:ring-primary/20" :placeholder="t('auth.usernameOrEmail')" autocomplete="username" />
              </span>
            </label>
            <label class="block">
              <span class="sr-only">{{ t('auth.password') }}</span>
              <span class="relative block">
                <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model="loginForm.password" type="password" class="h-10 w-full rounded-xl border border-line bg-base-100 pl-10 pr-3 text-sm outline-none transition focus:border-primary focus:ring-4 focus:ring-primary/20" :placeholder="t('auth.password')" autocomplete="current-password" />
              </span>
            </label>
            <div class="flex gap-3">
              <input v-model.trim="loginForm.captcha" class="h-10 min-w-0 flex-1 rounded-xl border border-line bg-base-100 px-3 text-sm outline-none transition focus:border-primary focus:ring-4 focus:ring-primary/20" :placeholder="t('auth.captcha')" />
              <button type="button" class="relative h-10 w-28 overflow-hidden rounded-xl border border-line bg-base-200" @click="refreshCaptcha">
                <LoaderCircle v-if="captchaLoading || !captchaImg" class="mx-auto h-5 w-5 animate-spin text-base-content/55" />
                <img v-else :src="captchaImg" :alt="t('auth.captchaAlt')" class="h-full w-full object-cover" />
              </button>
            </div>
            <div class="flex justify-end">
              <button type="button" class="text-sm font-medium text-primary hover:text-primary" @click="switchMode('forgot')">{{ t('auth.forgotPassword') }}</button>
            </div>
            <button type="submit" class="inline-flex h-11 w-full items-center justify-center gap-2 rounded-xl bg-primary text-sm font-bold text-primary-content shadow-lg shadow-primary/20 transition hover:bg-primary disabled:cursor-not-allowed disabled:opacity-70" :disabled="loading.login">
              <LoaderCircle v-if="loading.login" class="h-4 w-4 animate-spin" />
              {{ t('shell.login') }}
            </button>
          </form>

          <form v-else-if="mode === 'register'" class="space-y-3.5" @submit.prevent="handleRegister">
            <label class="block">
              <span class="sr-only">{{ t('auth.username') }}</span>
              <span class="relative block">
                <UserRound class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model.trim="registerForm.username" class="h-10 w-full rounded-xl border border-line bg-base-100 pl-10 pr-3 text-sm outline-none transition focus:border-primary focus:ring-4 focus:ring-primary/20" :placeholder="t('auth.username')" autocomplete="username" />
              </span>
            </label>
            <label class="block">
              <span class="sr-only">{{ t('auth.email') }}</span>
              <span class="relative block">
                <Mail class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model.trim="registerForm.email" type="email" class="h-10 w-full rounded-xl border border-line bg-base-100 pl-10 pr-3 text-sm outline-none transition focus:border-primary focus:ring-4 focus:ring-primary/20" :placeholder="t('auth.email')" autocomplete="email" />
              </span>
            </label>
            <label class="block">
              <span class="sr-only">{{ t('auth.password') }}</span>
              <span class="relative block">
                <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model="registerForm.password" type="password" class="h-10 w-full rounded-xl border border-line bg-base-100 pl-10 pr-3 text-sm outline-none transition focus:border-primary focus:ring-4 focus:ring-primary/20" :placeholder="t('auth.password')" autocomplete="new-password" />
              </span>
            </label>
            <label class="block">
              <span class="sr-only">{{ t('auth.confirmPassword') }}</span>
              <span class="relative block">
                <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model="registerForm.confirmPassword" type="password" class="h-10 w-full rounded-xl border border-line bg-base-100 pl-10 pr-3 text-sm outline-none transition focus:border-primary focus:ring-4 focus:ring-primary/20" :placeholder="t('auth.confirmPassword')" autocomplete="new-password" />
              </span>
            </label>
            <div class="flex gap-3">
              <span class="relative min-w-0 flex-1">
                <Languages class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model.trim="registerForm.captcha" class="h-10 w-full rounded-xl border border-line bg-base-100 pl-10 pr-3 text-sm outline-none transition focus:border-primary focus:ring-4 focus:ring-primary/20" :placeholder="t('auth.captcha')" />
              </span>
              <button type="button" class="relative h-10 w-28 overflow-hidden rounded-xl border border-line bg-base-200" @click="refreshCaptcha">
                <LoaderCircle v-if="captchaLoading || !captchaImg" class="mx-auto h-5 w-5 animate-spin text-base-content/55" />
                <img v-else :src="captchaImg" :alt="t('auth.captchaAlt')" class="h-full w-full object-cover" />
              </button>
            </div>
            <label class="flex items-start gap-2 text-sm leading-5 text-base-content/55">
              <input v-model="registerForm.agree" type="checkbox" class="mt-1 h-4 w-4 rounded border-line text-primary focus:ring-primary" />
              <span>{{ t('auth.agreeTerms') }}</span>
            </label>
            <button type="submit" class="inline-flex h-11 w-full items-center justify-center gap-2 rounded-xl bg-neutral text-sm font-bold text-neutral-content shadow-lg shadow-base-300 transition hover:bg-neutral/90 disabled:cursor-not-allowed disabled:opacity-70" :disabled="loading.register">
              <LoaderCircle v-if="loading.register" class="h-4 w-4 animate-spin" />
              {{ t('auth.createAccount') }}
            </button>
          </form>

          <form v-else class="space-y-3.5" @submit.prevent="handleForgot">
            <label class="block">
              <span class="relative block">
                <Mail class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model.trim="forgotForm.email" type="email" class="h-10 w-full rounded-xl border border-line bg-base-100 pl-10 pr-3 text-sm outline-none transition focus:border-primary focus:ring-4 focus:ring-primary/20" :placeholder="t('auth.registeredEmail')" autocomplete="email" />
              </span>
            </label>
            <div class="flex gap-3">
              <input v-model.trim="forgotForm.captcha" class="h-10 min-w-0 flex-1 rounded-xl border border-line bg-base-100 px-3 text-sm outline-none transition focus:border-primary focus:ring-4 focus:ring-primary/20" :placeholder="t('auth.captcha')" />
              <button type="button" class="relative h-10 w-28 overflow-hidden rounded-xl border border-line bg-base-200" @click="refreshCaptcha">
                <LoaderCircle v-if="captchaLoading || !captchaImg" class="mx-auto h-5 w-5 animate-spin text-base-content/55" />
                <img v-else :src="captchaImg" :alt="t('auth.captchaAlt')" class="h-full w-full object-cover" />
              </button>
            </div>
            <button type="submit" class="inline-flex h-11 w-full items-center justify-center gap-2 rounded-xl bg-primary text-sm font-bold text-primary-content shadow-lg shadow-primary/20 transition hover:bg-primary disabled:cursor-not-allowed disabled:opacity-70" :disabled="loading.forgot">
              <LoaderCircle v-if="loading.forgot" class="h-4 w-4 animate-spin" />
              {{ t('auth.sendResetEmail') }}
            </button>
            <button type="button" class="w-full text-sm font-medium text-primary hover:text-primary" @click="switchMode('login')">{{ t('auth.backToLogin') }}</button>
          </form>
        </div>

        <aside class="border-t border-line bg-base-200/70 px-5 py-6 sm:px-8 md:border-l md:border-t-0">
          <div class="flex h-full flex-col justify-center">
            <div v-if="showSocial">
              <h2 class="text-sm font-bold text-base-content">{{ t('auth.continueWith') }}</h2>
              <div class="mt-8 space-y-3.5">
                <a :href="page.props.githubUrl" class="flex h-10 w-full items-center justify-center gap-2 rounded-xl border border-line bg-base-100 text-sm font-semibold text-base-content shadow-sm transition hover:bg-base-200">
                  <svg class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                    <path d="M12 0C5.37 0 0 5.37 0 12c0 5.3 3.44 9.8 8.21 11.39.6.11.79-.26.79-.58v-2.03c-3.34.73-4.04-1.42-4.04-1.42-.55-1.39-1.34-1.76-1.34-1.76-1.09-.75.08-.73.08-.73 1.21.08 1.85 1.24 1.85 1.24 1.07 1.83 2.81 1.3 3.49 1 .11-.78.42-1.3.76-1.6-2.67-.31-5.47-1.34-5.47-5.93 0-1.31.47-2.38 1.24-3.22-.12-.3-.54-1.52.12-3.18 0 0 1.01-.32 3.3 1.23A11.5 11.5 0 0 1 12 6c1.02 0 2.05.14 3.01.4 2.29-1.55 3.3-1.23 3.3-1.23.65 1.66.24 2.88.12 3.18.77.84 1.24 1.91 1.24 3.22 0 4.61-2.81 5.62-5.48 5.92.43.37.81 1.1.81 2.22v3.29c0 .32.19.69.8.58A12.01 12.01 0 0 0 24 12c0-6.63-5.37-12-12-12Z" />
                  </svg>
                  GitHub
                </a>
                <button type="button" class="flex h-10 w-full cursor-not-allowed items-center justify-center gap-2 rounded-xl border border-line bg-base-100 text-sm font-semibold text-base-content/55 opacity-70">
                  {{ t('auth.googleUnavailable') }}
                </button>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </section>
  </main>
</template>
