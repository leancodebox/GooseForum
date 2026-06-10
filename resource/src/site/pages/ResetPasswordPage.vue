<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Check, LoaderCircle, LockKeyhole } from '@lucide/vue'
import { resetPassword } from '@/runtime/api'
import type { LayoutPayload, ResetPasswordPageProps } from '@/types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: ResetPasswordPageProps
}>()
const { t } = useI18n()

const form = reactive({
  password: '',
  confirmPassword: '',
})
const loading = ref(false)
const error = ref('')
const success = ref('')

const canSubmit = computed(() => Boolean(page.props.token && form.password && form.confirmPassword && !loading.value))

async function submit() {
  error.value = ''
  success.value = ''
  if (!page.props.token) {
    error.value = t('auth.resetMissingToken')
    return
  }
  if (form.password.length < 6) {
    error.value = t('auth.passwordMinLength')
    return
  }
  if (form.password !== form.confirmPassword) {
    error.value = t('auth.validation.passwordMismatch')
    return
  }
  loading.value = true
  try {
    success.value = await resetPassword(page.props.token, form.password)
    form.password = ''
    form.confirmPassword = ''
  } catch (err) {
    error.value = err instanceof Error && err.message ? err.message : t('api.passwordResetFailed')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="min-h-screen bg-base-200 px-4 py-8 text-base-content sm:px-6 lg:px-8">
    <section class="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-[880px] items-center justify-center">
      <div class="grid w-full overflow-hidden rounded-2xl border border-line bg-base-100 shadow-[0_18px_50px_rgba(15,23,42,0.07)] md:grid-cols-2">
        <div class="flex min-h-[470px] flex-col justify-center px-5 py-6 sm:px-8">
          <a href="/" class="mb-6 inline-flex items-baseline text-[27px] font-semibold leading-none tracking-[-0.04em] text-primary">
            <span v-if="page.layout.site.brandType === 'image' && page.layout.site.brandImage">
              <img :src="page.layout.site.brandImage" :alt="page.layout.site.name" class="h-8 w-auto object-contain" />
            </span>
            <span v-else-if="page.layout.site.brandType === 'text'">
              {{ page.layout.site.brandText || page.layout.site.name }}
            </span>
            <span v-else>Goose<span class="text-base-content">Forum</span></span>
          </a>

          <div class="mb-4">
            <h1 class="text-[27px] font-bold leading-tight tracking-tight text-base-content">{{ t('auth.resetPasswordTitle') }}</h1>
            <p class="mt-1.5 text-sm leading-6 text-base-content/55">{{ t('auth.resetPasswordSubtitle') }}</p>
          </div>

          <p v-if="error" class="mb-4 rounded-lg border border-error/20 bg-error/10 px-3 py-2 text-sm font-medium text-error">{{ error }}</p>
          <p v-if="success" class="mb-4 rounded-lg border border-success/20 bg-success/10 px-3 py-2 text-sm font-medium text-success">{{ success }}</p>

          <form class="space-y-3.5" @submit.prevent="submit">
            <label class="block">
              <span class="sr-only">{{ t('auth.newPassword') }}</span>
              <span class="relative block">
                <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model="form.password" type="password" class="h-10 w-full rounded-xl border border-line bg-base-100 pl-10 pr-3 text-sm outline-none transition focus:border-primary focus:ring-4 focus:ring-primary/20" :placeholder="t('auth.newPassword')" autocomplete="new-password" />
              </span>
            </label>
            <label class="block">
              <span class="sr-only">{{ t('auth.confirmPassword') }}</span>
              <span class="relative block">
                <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model="form.confirmPassword" type="password" class="h-10 w-full rounded-xl border border-line bg-base-100 pl-10 pr-3 text-sm outline-none transition focus:border-primary focus:ring-4 focus:ring-primary/20" :placeholder="t('auth.confirmPassword')" autocomplete="new-password" />
              </span>
            </label>
            <button type="submit" class="inline-flex h-11 w-full items-center justify-center gap-2 rounded-xl bg-primary text-sm font-bold text-primary-content shadow-lg shadow-primary/20 transition hover:bg-primary disabled:cursor-not-allowed disabled:opacity-70" :disabled="!canSubmit">
              <LoaderCircle v-if="loading" class="h-4 w-4 animate-spin" />
              {{ t('auth.saveNewPassword') }}
            </button>
            <a href="/login" class="inline-flex h-10 w-full items-center justify-center rounded-xl text-sm font-semibold text-primary hover:bg-info/10">{{ t('auth.backToLogin') }}</a>
          </form>
        </div>

        <aside class="border-t border-line bg-base-200/70 px-5 py-6 sm:px-8 md:border-l md:border-t-0">
          <div class="flex h-full flex-col justify-center">
            <h2 class="text-lg font-bold text-base-content">{{ t('auth.passwordAdviceTitle') }}</h2>
            <p class="mt-3 text-sm leading-6 text-base-content/55">{{ t('auth.passwordAdviceDescription') }}</p>
            <ul class="mt-6 space-y-3 text-sm text-base-content/75">
              <li class="flex items-center gap-2"><Check class="h-4 w-4 text-success" /> {{ t('auth.passwordAdvice.length') }}</li>
              <li class="flex items-center gap-2"><Check class="h-4 w-4 text-success" /> {{ t('auth.passwordAdvice.unique') }}</li>
              <li class="flex items-center gap-2"><Check class="h-4 w-4 text-success" /> {{ t('auth.passwordAdvice.loginAfterReset') }}</li>
            </ul>
          </div>
        </aside>
      </div>
    </section>
  </main>
</template>
