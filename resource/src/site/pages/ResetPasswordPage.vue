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
  <main class="min-h-screen bg-base-100 text-base-content sm:bg-base-200 sm:px-6 sm:py-8 lg:px-8">
    <section class="mx-auto flex min-h-screen w-full max-w-[880px] items-stretch justify-center sm:min-h-[calc(100vh-4rem)] sm:items-center">
      <div class="gf-card grid w-full overflow-hidden border-0 shadow-none sm:border sm:shadow-[0_2px_12px_rgb(0_0_0/calc(var(--gf-depth)*0.04))] md:grid-cols-2">
        <div class="flex min-h-screen flex-col justify-center px-4 py-12 sm:min-h-[470px] sm:px-8 sm:py-6">
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

          <p v-if="error" class="gf-status-message gf-status-message-error mb-4">{{ error }}</p>
          <p v-if="success" class="gf-status-message gf-status-message-success mb-4">{{ success }}</p>

          <form class="space-y-3.5" @submit.prevent="submit">
            <label class="block">
              <span class="sr-only">{{ t('auth.newPassword') }}</span>
              <span class="relative block">
                <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model="form.password" type="password" class="gf-input pl-10" :placeholder="t('auth.newPassword')" autocomplete="new-password" />
              </span>
            </label>
            <label class="block">
              <span class="sr-only">{{ t('auth.confirmPassword') }}</span>
              <span class="relative block">
                <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/55" />
                <input v-model="form.confirmPassword" type="password" class="gf-input pl-10" :placeholder="t('auth.confirmPassword')" autocomplete="new-password" />
              </span>
            </label>
            <button type="submit" class="gf-button gf-button-xl gf-button-primary w-full" :disabled="!canSubmit">
              <LoaderCircle v-if="loading" class="h-4 w-4 animate-spin" />
              {{ t('auth.saveNewPassword') }}
            </button>
            <a href="/login" class="gf-button gf-button-lg gf-button-ghost w-full">{{ t('auth.backToLogin') }}</a>
          </form>
        </div>

        <aside class="border-t border-line bg-base-200/70 px-4 py-6 sm:px-8 md:border-l md:border-t-0">
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
