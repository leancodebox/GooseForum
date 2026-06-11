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
    <div class="gf-locale-switch absolute right-4 top-4 z-10">
      <button
        v-for="item in supportedLocales"
        :key="item"
        type="button"
        class="gf-locale-switch-item"
        :class="{ 'gf-locale-switch-item-active': locale === item }"
        @click="switchLocale(item)"
      >
        {{ t(`locale.short.${item}`) }}
      </button>
    </div>

    <section class="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-[880px] items-center justify-center">
      <div class="gf-card grid w-full overflow-hidden md:grid-cols-2">
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

          <div v-if="mode !== 'forgot'" class="gf-segmented mb-4 grid-cols-2">
            <button type="button" class="gf-segmented-item" :class="mode === 'login' ? 'gf-segmented-item-active' : 'gf-segmented-item-idle'" @click="switchMode('login')">{{ t('shell.login') }}</button>
            <button type="button" class="gf-segmented-item" :class="mode === 'register' ? 'gf-segmented-item-active' : 'gf-segmented-item-idle'" @click="switchMode('register')">{{ t('shell.register') }}</button>
          </div>

          <p v-if="error" class="gf-status-message gf-status-message-error mb-4">{{ error }}</p>
          <p v-if="notice" class="gf-status-message gf-status-message-success mb-4">{{ notice }}</p>

          <form v-if="mode === 'login'" class="space-y-3.5" @submit.prevent="handleLogin">
            <label class="block">
              <span class="sr-only">{{ t('auth.usernameOrEmail') }}</span>
              <span class="relative block">
                <UserRound class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model.trim="loginForm.username" class="gf-input pl-10" :placeholder="t('auth.usernameOrEmail')" autocomplete="username" />
              </span>
            </label>
            <label class="block">
              <span class="sr-only">{{ t('auth.password') }}</span>
              <span class="relative block">
                <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model="loginForm.password" type="password" class="gf-input pl-10" :placeholder="t('auth.password')" autocomplete="current-password" />
              </span>
            </label>
            <div class="flex gap-3">
              <input v-model.trim="loginForm.captcha" class="gf-input min-w-0 flex-1" :placeholder="t('auth.captcha')" />
              <button type="button" class="relative h-10 w-28 overflow-hidden gf-panel" @click="refreshCaptcha">
                <LoaderCircle v-if="captchaLoading || !captchaImg" class="mx-auto h-5 w-5 animate-spin text-base-content/55" />
                <img v-else :src="captchaImg" :alt="t('auth.captchaAlt')" class="h-full w-full object-cover" />
              </button>
            </div>
            <div class="flex justify-end">
              <button type="button" class="text-sm font-medium text-primary hover:text-primary" @click="switchMode('forgot')">{{ t('auth.forgotPassword') }}</button>
            </div>
            <button type="submit" class="gf-button gf-button-xl gf-button-primary w-full" :disabled="loading.login">
              <LoaderCircle v-if="loading.login" class="h-4 w-4 animate-spin" />
              {{ t('shell.login') }}
            </button>
          </form>

          <form v-else-if="mode === 'register'" class="space-y-3.5" @submit.prevent="handleRegister">
            <label class="block">
              <span class="sr-only">{{ t('auth.username') }}</span>
              <span class="relative block">
                <UserRound class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model.trim="registerForm.username" class="gf-input pl-10" :placeholder="t('auth.username')" autocomplete="username" />
              </span>
            </label>
            <label class="block">
              <span class="sr-only">{{ t('auth.email') }}</span>
              <span class="relative block">
                <Mail class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model.trim="registerForm.email" type="email" class="gf-input pl-10" :placeholder="t('auth.email')" autocomplete="email" />
              </span>
            </label>
            <label class="block">
              <span class="sr-only">{{ t('auth.password') }}</span>
              <span class="relative block">
                <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model="registerForm.password" type="password" class="gf-input pl-10" :placeholder="t('auth.password')" autocomplete="new-password" />
              </span>
            </label>
            <label class="block">
              <span class="sr-only">{{ t('auth.confirmPassword') }}</span>
              <span class="relative block">
                <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model="registerForm.confirmPassword" type="password" class="gf-input pl-10" :placeholder="t('auth.confirmPassword')" autocomplete="new-password" />
              </span>
            </label>
            <div class="flex gap-3">
              <span class="relative min-w-0 flex-1">
                <Languages class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model.trim="registerForm.captcha" class="gf-input pl-10" :placeholder="t('auth.captcha')" />
              </span>
              <button type="button" class="relative h-10 w-28 overflow-hidden gf-panel" @click="refreshCaptcha">
                <LoaderCircle v-if="captchaLoading || !captchaImg" class="mx-auto h-5 w-5 animate-spin text-base-content/55" />
                <img v-else :src="captchaImg" :alt="t('auth.captchaAlt')" class="h-full w-full object-cover" />
              </button>
            </div>
            <label class="flex items-start gap-2 text-sm leading-5 text-base-content/55">
              <input v-model="registerForm.agree" type="checkbox" class="mt-1 h-4 w-4 rounded border-line text-primary focus:ring-primary" />
              <span>{{ t('auth.agreeTerms') }}</span>
            </label>
            <button type="submit" class="gf-button gf-button-xl gf-button-neutral w-full" :disabled="loading.register">
              <LoaderCircle v-if="loading.register" class="h-4 w-4 animate-spin" />
              {{ t('auth.createAccount') }}
            </button>
          </form>

          <form v-else class="space-y-3.5" @submit.prevent="handleForgot">
            <label class="block">
              <span class="relative block">
                <Mail class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model.trim="forgotForm.email" type="email" class="gf-input pl-10" :placeholder="t('auth.registeredEmail')" autocomplete="email" />
              </span>
            </label>
            <div class="flex gap-3">
              <input v-model.trim="forgotForm.captcha" class="gf-input min-w-0 flex-1" :placeholder="t('auth.captcha')" />
              <button type="button" class="relative h-10 w-28 overflow-hidden gf-panel" @click="refreshCaptcha">
                <LoaderCircle v-if="captchaLoading || !captchaImg" class="mx-auto h-5 w-5 animate-spin text-base-content/55" />
                <img v-else :src="captchaImg" :alt="t('auth.captchaAlt')" class="h-full w-full object-cover" />
              </button>
            </div>
            <button type="submit" class="gf-button gf-button-xl gf-button-primary w-full" :disabled="loading.forgot">
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
                <a :href="page.props.githubUrl" class="gf-button gf-button-lg gf-button-secondary w-full">
                  <svg class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                    <path d="M12 0C5.37 0 0 5.37 0 12c0 5.3 3.44 9.8 8.21 11.39.6.11.79-.26.79-.58v-2.03c-3.34.73-4.04-1.42-4.04-1.42-.55-1.39-1.34-1.76-1.34-1.76-1.09-.75.08-.73.08-.73 1.21.08 1.85 1.24 1.85 1.24 1.07 1.83 2.81 1.3 3.49 1 .11-.78.42-1.3.76-1.6-2.67-.31-5.47-1.34-5.47-5.93 0-1.31.47-2.38 1.24-3.22-.12-.3-.54-1.52.12-3.18 0 0 1.01-.32 3.3 1.23A11.5 11.5 0 0 1 12 6c1.02 0 2.05.14 3.01.4 2.29-1.55 3.3-1.23 3.3-1.23.65 1.66.24 2.88.12 3.18.77.84 1.24 1.91 1.24 3.22 0 4.61-2.81 5.62-5.48 5.92.43.37.81 1.1.81 2.22v3.29c0 .32.19.69.8.58A12.01 12.01 0 0 0 24 12c0-6.63-5.37-12-12-12Z" />
                  </svg>
                  GitHub
                </a>
                <button type="button" class="gf-button gf-button-lg gf-button-secondary w-full cursor-not-allowed opacity-70">
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
