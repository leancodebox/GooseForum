<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { Check, LoaderCircle, LockKeyhole } from '@lucide/vue'
import { resetPassword } from '@/runtime/api'
import type { LayoutPayload, ResetPasswordPageProps } from '@/types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: ResetPasswordPageProps
}>()

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
    error.value = '重置链接缺少 token，请重新从邮件打开。'
    return
  }
  if (form.password.length < 6) {
    error.value = '密码长度至少 6 位'
    return
  }
  if (form.password !== form.confirmPassword) {
    error.value = '两次输入的密码不一致'
    return
  }
  loading.value = true
  try {
    success.value = await resetPassword(page.props.token, form.password)
    form.password = ''
    form.confirmPassword = ''
  } catch (err) {
    error.value = err instanceof Error && err.message ? err.message : '密码重置失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="min-h-screen bg-gray-50 px-4 py-8 text-gray-950 sm:px-6 lg:px-8">
    <section class="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-[880px] items-center justify-center">
      <div class="grid w-full overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-[0_18px_50px_rgba(15,23,42,0.07)] md:grid-cols-2">
        <div class="flex min-h-[470px] flex-col justify-center px-5 py-6 sm:px-8">
          <a href="/" class="mb-6 inline-flex items-baseline text-[27px] font-semibold leading-none tracking-[-0.04em] text-blue-600">
            <span v-if="page.layout.site.brandType === 'image' && page.layout.site.brandImage">
              <img :src="page.layout.site.brandImage" :alt="page.layout.site.name" class="h-8 w-auto object-contain" />
            </span>
            <span v-else-if="page.layout.site.brandType === 'text'">
              {{ page.layout.site.brandText || page.layout.site.name }}
            </span>
            <span v-else>Goose<span class="text-gray-950">Forum</span></span>
          </a>

          <div class="mb-4">
            <h1 class="text-[27px] font-bold leading-tight tracking-tight text-gray-950">重置密码</h1>
            <p class="mt-1.5 text-sm leading-6 text-gray-500">设置一个新的登录密码。提交后即可返回登录页使用新密码登录。</p>
          </div>

          <p v-if="error" class="mb-4 rounded-lg border border-red-100 bg-red-50 px-3 py-2 text-sm font-medium text-red-700">{{ error }}</p>
          <p v-if="success" class="mb-4 rounded-lg border border-emerald-100 bg-emerald-50 px-3 py-2 text-sm font-medium text-emerald-700">{{ success }}</p>

          <form class="space-y-3.5" @submit.prevent="submit">
            <label class="block">
              <span class="sr-only">新密码</span>
              <span class="relative block">
                <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                <input v-model="form.password" type="password" class="h-10 w-full rounded-xl border border-gray-200 bg-white pl-10 pr-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100" placeholder="新密码" autocomplete="new-password" />
              </span>
            </label>
            <label class="block">
              <span class="sr-only">确认新密码</span>
              <span class="relative block">
                <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                <input v-model="form.confirmPassword" type="password" class="h-10 w-full rounded-xl border border-gray-200 bg-white pl-10 pr-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100" placeholder="确认新密码" autocomplete="new-password" />
              </span>
            </label>
            <button type="submit" class="inline-flex h-11 w-full items-center justify-center gap-2 rounded-xl bg-blue-600 text-sm font-bold text-white shadow-lg shadow-blue-200 transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-70" :disabled="!canSubmit">
              <LoaderCircle v-if="loading" class="h-4 w-4 animate-spin" />
              保存新密码
            </button>
            <a href="/login" class="inline-flex h-10 w-full items-center justify-center rounded-xl text-sm font-semibold text-blue-600 hover:bg-blue-50">返回登录</a>
          </form>
        </div>

        <aside class="border-t border-gray-100 bg-gray-50/70 px-5 py-6 sm:px-8 md:border-l md:border-t-0">
          <div class="flex h-full flex-col justify-center">
            <h2 class="text-lg font-bold text-gray-950">密码安全建议</h2>
            <p class="mt-3 text-sm leading-6 text-gray-500">设置一个只用于 GooseForum 的新密码。重置链接只在有效期内可用，过期后请重新发送邮件。</p>
            <ul class="mt-6 space-y-3 text-sm text-gray-600">
              <li class="flex items-center gap-2"><Check class="h-4 w-4 text-emerald-600" /> 至少 6 位字符</li>
              <li class="flex items-center gap-2"><Check class="h-4 w-4 text-emerald-600" /> 避免与其他网站共用密码</li>
              <li class="flex items-center gap-2"><Check class="h-4 w-4 text-emerald-600" /> 重置成功后返回登录页使用新密码</li>
            </ul>
          </div>
        </aside>
      </div>
    </section>
  </main>
</template>
