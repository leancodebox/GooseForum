<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { KeyRound, LoaderCircle, LockKeyhole, Mail, UserRound } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { forgotPassword, getCaptcha, login, register } from '@/runtime/api'
import { queueFlashMessage } from '@/runtime/flash-message'
import { setLocale, supportedLocales, type Locale } from '@/runtime/i18n'
import type { LayoutPayload, LoginPageProps } from '@gooseforum/client'

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

const showSocial = computed(() => mode.value !== 'forgot' && page.props.oauthProviders.length > 0)
const homeUrl = computed(() => page.props.redirectUrl || '/')
const inputClass = 'h-10 rounded-[var(--gf-radius-field)] border-line bg-base-100 shadow-none focus-visible:border-primary focus-visible:ring-4 focus-visible:ring-primary/20'

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
  loginForm.username = loginForm.username.trim()
  loginForm.captcha = loginForm.captcha.trim()
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
  registerForm.username = registerForm.username.trim()
  registerForm.email = registerForm.email.trim()
  registerForm.captcha = registerForm.captcha.trim()
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
    const message = await register(registerForm.username, registerForm.email, registerForm.password, captchaId.value, registerForm.captcha, String(locale.value))
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
  forgotForm.email = forgotForm.email.trim()
  forgotForm.captcha = forgotForm.captcha.trim()
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
  <main class="relative min-h-screen bg-base-100 text-base-content sm:bg-base-200 sm:px-6 sm:py-8 lg:px-8">
    <Tabs
      :model-value="locale"
      class="absolute right-3 top-3 z-10 sm:right-4 sm:top-4"
      @update:model-value="switchLocale($event as Locale)"
    >
      <TabsList class="inline-flex h-auto items-center gap-1 rounded-full border border-line bg-base-100 p-1 shadow-[0_8px_22px_-16px_rgb(15_23_42/calc(var(--gf-depth)*0.5))]">
        <TabsTrigger
          v-for="item in supportedLocales"
          :key="item"
          :value="item"
          class="h-8 min-w-8 rounded-full px-2 text-xs font-semibold text-base-content/55 shadow-none data-[state=active]:bg-neutral data-[state=active]:text-neutral-content data-[state=active]:shadow-sm hover:text-base-content data-[state=active]:hover:bg-neutral data-[state=active]:hover:text-neutral-content"
        >
          {{ t(`locale.short.${item}`) }}
        </TabsTrigger>
      </TabsList>
    </Tabs>

    <section class="mx-auto flex min-h-screen w-full max-w-[880px] items-stretch justify-center sm:min-h-[calc(100vh-4rem)] sm:items-center">
      <div class="gf-card grid w-full overflow-hidden border-0 shadow-none sm:border sm:shadow-[0_2px_12px_rgb(0_0_0/calc(var(--gf-depth)*0.04))] md:grid-cols-2">
        <div class="flex min-h-screen flex-col justify-center px-4 py-12 sm:min-h-[470px] sm:px-8 sm:py-6">
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

          <Tabs v-if="mode !== 'forgot'" :model-value="mode" class="mb-4" @update:model-value="value => switchMode(value as Mode)">
            <TabsList class="grid w-full grid-cols-2">
              <TabsTrigger value="login">{{ t('shell.login') }}</TabsTrigger>
              <TabsTrigger value="register">{{ t('shell.register') }}</TabsTrigger>
            </TabsList>
          </Tabs>

          <p v-if="error" class="gf-status-message gf-status-message-error mb-4">{{ error }}</p>
          <p v-if="notice" class="gf-status-message gf-status-message-success mb-4">{{ notice }}</p>

          <form v-if="mode === 'login'" class="space-y-3" @submit.prevent="handleLogin">
            <label class="block">
              <span class="sr-only">{{ t('auth.usernameOrEmail') }}</span>
              <span class="relative block">
                <UserRound class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <Input v-model="loginForm.username" :class="[inputClass, 'pl-10']" :placeholder="t('auth.usernameOrEmail')" autocomplete="username" />
              </span>
            </label>
            <label class="block">
              <span class="sr-only">{{ t('auth.password') }}</span>
              <span class="relative block">
                <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <Input v-model="loginForm.password" type="password" :class="[inputClass, 'pl-10']" :placeholder="t('auth.password')" autocomplete="current-password" />
              </span>
            </label>
            <div class="flex gap-3">
              <Input v-model="loginForm.captcha" :class="[inputClass, 'min-w-0 flex-1']" :placeholder="t('auth.captcha')" />
              <Button type="button" variant="surface" class="relative h-10 w-28 overflow-hidden rounded-md p-0 shadow-none gf-panel" @click="refreshCaptcha">
                <LoaderCircle v-if="captchaLoading || !captchaImg" class="mx-auto h-5 w-5 animate-spin text-base-content/55" />
                <img v-else :src="captchaImg" :alt="t('auth.captchaAlt')" class="h-full w-full object-cover" />
              </Button>
            </div>
            <div class="flex justify-end">
              <Button type="button" variant="brand-ghost" size="sm" class="h-auto px-0 py-0" @click="switchMode('forgot')">{{ t('auth.forgotPassword') }}</Button>
            </div>
            <Button type="submit" variant="brand" size="xl" class="w-full" :disabled="loading.login">
              <LoaderCircle v-if="loading.login" class="h-4 w-4 animate-spin" />
              {{ t('shell.login') }}
            </Button>
          </form>

          <form v-else-if="mode === 'register'" class="space-y-3" @submit.prevent="handleRegister">
            <label class="block">
              <span class="sr-only">{{ t('auth.username') }}</span>
              <span class="relative block">
                <UserRound class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <Input v-model="registerForm.username" :class="[inputClass, 'pl-10']" :placeholder="t('auth.username')" autocomplete="username" />
              </span>
            </label>
            <label class="block">
              <span class="sr-only">{{ t('auth.email') }}</span>
              <span class="relative block">
                <Mail class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <Input v-model="registerForm.email" type="email" :class="[inputClass, 'pl-10']" :placeholder="t('auth.email')" autocomplete="email" />
              </span>
            </label>
            <label class="block">
              <span class="sr-only">{{ t('auth.password') }}</span>
              <span class="relative block">
                <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <Input v-model="registerForm.password" type="password" :class="[inputClass, 'pl-10']" :placeholder="t('auth.password')" autocomplete="new-password" />
              </span>
            </label>
            <label class="block">
              <span class="sr-only">{{ t('auth.confirmPassword') }}</span>
              <span class="relative block">
                <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <Input v-model="registerForm.confirmPassword" type="password" :class="[inputClass, 'pl-10']" :placeholder="t('auth.confirmPassword')" autocomplete="new-password" />
              </span>
            </label>
            <div class="flex gap-3">
              <span class="relative min-w-0 flex-1">
                <Input v-model="registerForm.captcha" :class="inputClass" :placeholder="t('auth.captcha')" />
              </span>
              <Button type="button" variant="surface" class="relative h-10 w-28 overflow-hidden rounded-md p-0 shadow-none gf-panel" @click="refreshCaptcha">
                <LoaderCircle v-if="captchaLoading || !captchaImg" class="mx-auto h-5 w-5 animate-spin text-base-content/55" />
                <img v-else :src="captchaImg" :alt="t('auth.captchaAlt')" class="h-full w-full object-cover" />
              </Button>
            </div>
            <label class="flex items-start gap-2 text-sm leading-5 text-base-content/55">
              <Checkbox v-model="registerForm.agree" class="mt-1 border-line shadow-none focus-visible:ring-primary" />
              <span>{{ t('auth.agreeTerms') }}</span>
            </label>
            <Button type="submit" variant="neutral" size="xl" class="w-full" :disabled="loading.register">
              <LoaderCircle v-if="loading.register" class="h-4 w-4 animate-spin" />
              {{ t('auth.createAccount') }}
            </Button>
          </form>

          <form v-else class="space-y-3.5" @submit.prevent="handleForgot">
            <label class="block">
              <span class="relative block">
                <Mail class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <Input v-model="forgotForm.email" type="email" :class="[inputClass, 'pl-10']" :placeholder="t('auth.registeredEmail')" autocomplete="email" />
              </span>
            </label>
            <div class="flex gap-3">
              <Input v-model="forgotForm.captcha" :class="[inputClass, 'min-w-0 flex-1']" :placeholder="t('auth.captcha')" />
              <Button type="button" variant="surface" class="relative h-10 w-28 overflow-hidden rounded-md p-0 shadow-none gf-panel" @click="refreshCaptcha">
                <LoaderCircle v-if="captchaLoading || !captchaImg" class="mx-auto h-5 w-5 animate-spin text-base-content/55" />
                <img v-else :src="captchaImg" :alt="t('auth.captchaAlt')" class="h-full w-full object-cover" />
              </Button>
            </div>
            <Button type="submit" variant="brand" size="xl" class="w-full" :disabled="loading.forgot">
              <LoaderCircle v-if="loading.forgot" class="h-4 w-4 animate-spin" />
              {{ t('auth.sendResetEmail') }}
            </Button>
            <Button type="button" variant="brand-ghost" size="sm" class="h-auto w-full py-0" @click="switchMode('login')">{{ t('auth.backToLogin') }}</Button>
          </form>

          <div v-if="showSocial" class="mt-5 border-t border-line pt-4 md:hidden">
            <h2 class="mb-2 text-xs font-bold uppercase text-base-content/45">{{ t('auth.continueWith') }}</h2>
            <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
              <Button v-for="provider in page.props.oauthProviders" :key="provider.key" as-child variant="surface" class="w-full px-3">
                <a :href="provider.loginUrl">
                  <svg v-if="provider.key === 'github'" class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                    <path d="M12 0C5.37 0 0 5.37 0 12c0 5.3 3.44 9.8 8.21 11.39.6.11.79-.26.79-.58v-2.03c-3.34.73-4.04-1.42-4.04-1.42-.55-1.39-1.34-1.76-1.34-1.76-1.09-.75.08-.73.08-.73 1.21.08 1.85 1.24 1.85 1.24 1.07 1.83 2.81 1.3 3.49 1 .11-.78.42-1.3.76-1.6-2.67-.31-5.47-1.34-5.47-5.93 0-1.31.47-2.38 1.24-3.22-.12-.3-.54-1.52.12-3.18 0 0 1.01-.32 3.3 1.23A11.5 11.5 0 0 1 12 6c1.02 0 2.05.14 3.01.4 2.29-1.55 3.3-1.23 3.3-1.23.65 1.66.24 2.88.12 3.18.77.84 1.24 1.91 1.24 3.22 0 4.61-2.81 5.62-5.48 5.92.43.37.81 1.1.81 2.22v3.29c0 .32.19.69.8.58A12.01 12.01 0 0 0 24 12c0-6.63-5.37-12-12-12Z" />
                  </svg>
                  <KeyRound v-else class="h-4 w-4" />
                  {{ provider.displayName }}
                </a>
              </Button>
            </div>
          </div>
        </div>

        <aside class="hidden border-t border-line bg-base-200/70 px-4 py-6 sm:px-8 md:block md:border-l md:border-t-0">
          <div class="flex h-full flex-col justify-center">
            <div v-if="showSocial">
              <h2 class="text-sm font-bold text-base-content">{{ t('auth.continueWith') }}</h2>
              <div class="mt-8 space-y-3.5">
                <Button v-for="provider in page.props.oauthProviders" :key="provider.key" as-child variant="surface" size="lg" class="w-full px-4">
                  <a :href="provider.loginUrl">
                    <svg v-if="provider.key === 'github'" class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                      <path d="M12 0C5.37 0 0 5.37 0 12c0 5.3 3.44 9.8 8.21 11.39.6.11.79-.26.79-.58v-2.03c-3.34.73-4.04-1.42-4.04-1.42-.55-1.39-1.34-1.76-1.34-1.76-1.09-.75.08-.73.08-.73 1.21.08 1.85 1.24 1.85 1.24 1.07 1.83 2.81 1.3 3.49 1 .11-.78.42-1.3.76-1.6-2.67-.31-5.47-1.34-5.47-5.93 0-1.31.47-2.38 1.24-3.22-.12-.3-.54-1.52.12-3.18 0 0 1.01-.32 3.3 1.23A11.5 11.5 0 0 1 12 6c1.02 0 2.05.14 3.01.4 2.29-1.55 3.3-1.23 3.3-1.23.65 1.66.24 2.88.12 3.18.77.84 1.24 1.91 1.24 3.22 0 4.61-2.81 5.62-5.48 5.92.43.37.81 1.1.81 2.22v3.29c0 .32.19.69.8.58A12.01 12.01 0 0 0 24 12c0-6.63-5.37-12-12-12Z" />
                    </svg>
                    <KeyRound v-else class="h-5 w-5" />
                    {{ provider.displayName }}
                  </a>
                </Button>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </section>
  </main>
</template>
