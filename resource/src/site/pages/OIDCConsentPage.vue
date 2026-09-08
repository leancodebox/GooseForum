<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ArrowRight, Check, KeyRound, Loader2, ShieldCheck, X } from '@lucide/vue'
import type { LayoutPayload } from '@gooseforum/client'

const page = defineProps<{ layout: LayoutPayload; props: { interaction: string } }>()
const details = ref<{ client: { id: string; name: string; public: boolean }; scopes: string[] } | null>(null)
const error = ref('')
const deciding = ref(false)

const scopeLabels: Record<string, string> = {
  openid: '确认你的身份',
  profile: '读取名称、用户名和头像',
  email: '读取邮箱和验证状态',
  offline_access: '在你离开后继续保持登录',
}

onMounted(async () => {
  try {
    const response = await fetch(`/oauth2/consent/details?interaction=${encodeURIComponent(page.props.interaction)}`, { credentials: 'same-origin', headers: { Accept: 'application/json' } })
    if (!response.ok) throw new Error('授权请求已失效，请返回应用重新发起登录。')
    details.value = await response.json()
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : '无法加载授权请求。'
  }
})

async function decide(decision: 'approve' | 'deny') {
  if (deciding.value) return
  deciding.value = true
  error.value = ''
  try {
    const response = await fetch('/oauth2/consent', { method: 'POST', credentials: 'same-origin', headers: { 'Content-Type': 'application/json', Accept: 'application/json' }, body: JSON.stringify({ interaction: page.props.interaction, decision }) })
    const payload = await response.json()
    if (!response.ok || !payload.redirect_url) throw new Error(payload.error_description || '授权请求处理失败。')
    window.location.assign(payload.redirect_url)
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : '授权请求处理失败。'
    deciding.value = false
  }
}
</script>

<template>
  <main class="relative grid min-h-screen place-items-center overflow-hidden bg-base-100 px-4 py-8 text-base-content sm:bg-base-200 sm:px-6">
    <div class="pointer-events-none absolute inset-x-0 top-0 h-64 bg-gradient-to-b from-primary/5 to-transparent" />

    <section class="gf-card relative w-full max-w-[520px] overflow-hidden border-0 shadow-none sm:border sm:shadow-[0_12px_40px_rgb(0_0_0/calc(var(--gf-depth)*0.09))]">
      <header class="border-b border-line/70 px-5 py-5 sm:px-7">
        <a href="/" class="inline-flex items-baseline text-xl font-semibold leading-none tracking-[-0.04em] text-primary">
          <img v-if="page.layout.site.brandType === 'image' && page.layout.site.brandImage" :src="page.layout.site.brandImage" :alt="page.layout.site.name" class="h-7 w-auto object-contain" />
          <template v-else-if="page.layout.site.brandType === 'text'">{{ page.layout.site.brandText || page.layout.site.name }}</template>
          <template v-else>Goose<span class="text-base-content">Forum</span></template>
        </a>
      </header>

      <div class="px-5 py-6 sm:px-7 sm:py-7">
        <div class="flex items-start gap-4">
          <span class="grid size-12 shrink-0 place-items-center rounded-xl bg-primary/10 text-primary ring-1 ring-primary/10"><ShieldCheck class="size-6" /></span>
          <div class="min-w-0 pt-0.5">
            <p class="text-xs font-semibold uppercase tracking-[0.12em] text-base-content/45">授权请求</p>
            <h1 class="mt-1 text-xl font-bold tracking-tight sm:text-2xl">使用 GooseForum 登录</h1>
          </div>
        </div>

        <div v-if="!details && !error" class="grid min-h-56 place-items-center">
          <div class="text-center text-sm text-base-content/50"><Loader2 class="mx-auto mb-3 size-6 animate-spin text-primary" />正在验证授权请求…</div>
        </div>

        <div v-else-if="error" class="mt-6">
          <p class="gf-status-message gf-status-message-error">{{ error }}</p>
          <a href="/" class="gf-button gf-button-md gf-button-secondary mt-5 w-full">返回 GooseForum</a>
        </div>

        <template v-else-if="details">
          <div class="mt-6 flex items-center gap-3 rounded-lg border border-line bg-base-200/45 p-4">
            <span class="grid size-10 shrink-0 place-items-center rounded-lg bg-base-100 text-base-content shadow-sm ring-1 ring-line"><KeyRound class="size-5" /></span>
            <div class="min-w-0 flex-1">
              <p class="truncate font-semibold">{{ details.client.name }}</p>
              <p class="mt-0.5 text-xs text-base-content/50">希望访问你的 GooseForum 账号</p>
            </div>
            <ArrowRight class="size-4 shrink-0 text-base-content/35" />
          </div>

          <div class="mt-6">
            <h2 class="text-sm font-semibold">允许后，此应用可以：</h2>
            <ul class="mt-3 divide-y divide-line/70 rounded-lg border border-line">
              <li v-for="scope in details.scopes" :key="scope" class="flex items-center gap-3 px-4 py-3 text-sm">
                <span class="grid size-6 shrink-0 place-items-center rounded-full bg-success/10 text-success"><Check class="size-3.5" stroke-width="2.5" /></span>
                <span>{{ scopeLabels[scope] || scope }}</span>
              </li>
            </ul>
          </div>

          <div class="mt-4 rounded-md bg-base-200/50 px-3 py-2.5 text-xs leading-5 text-base-content/50">
            客户端 ID <span class="ml-1 break-all font-mono text-base-content/65">{{ details.client.id }}</span>
          </div>
          <p class="mt-5 text-xs leading-5 text-base-content/50">请确认你信任此应用。你可以随时在账号设置中撤销访问权限。</p>

          <div class="mt-6 grid grid-cols-2 gap-3">
            <button type="button" class="gf-button gf-button-xl gf-button-secondary" :disabled="deciding" @click="decide('deny')"><X class="size-4" />拒绝</button>
            <button type="button" class="gf-button gf-button-xl gf-button-primary" :disabled="deciding" @click="decide('approve')">
              <Loader2 v-if="deciding" class="size-4 animate-spin" /><Check v-else class="size-4" />{{ deciding ? '处理中…' : '允许并继续' }}
            </button>
          </div>
        </template>
      </div>
    </section>
  </main>
</template>
