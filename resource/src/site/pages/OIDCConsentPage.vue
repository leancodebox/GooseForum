<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ArrowRight, Check, KeyRound, Loader2, X } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { useI18n } from 'vue-i18n'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { setLocale, supportedLocales, type Locale } from '@/runtime/i18n'
import type { LayoutPayload } from '@gooseforum/client'

const page = defineProps<{ layout: LayoutPayload; props: { interaction: string } }>()
const { t, locale } = useI18n()
const siteName = computed(() => page.layout.site.name)
const details = ref<{ client: { id: string; name: string; public: boolean }; scopes: string[] } | null>(null)
const error = ref('')
const deciding = ref(false)

const scopeLabels = computed<Record<string, string>>(() => ({
  openid: t('oidcConsent.openid'),
  profile: t('oidcConsent.profile'),
  email: t('oidcConsent.email'),
  offline_access: t('oidcConsent.offline_access'),
}))

onMounted(async () => {
  try {
    const response = await fetch(`/oauth2/consent/details?interaction=${encodeURIComponent(page.props.interaction)}`, { credentials: 'same-origin', headers: { Accept: 'application/json' } })
    if (!response.ok) throw new Error(t('oidcConsent.expired'))
    details.value = await response.json()
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : t('oidcConsent.loadFailed')
  }
})

async function decide(decision: 'approve' | 'deny') {
  if (deciding.value) return
  deciding.value = true
  error.value = ''
  try {
    const response = await fetch('/oauth2/consent', { method: 'POST', credentials: 'same-origin', headers: { 'Content-Type': 'application/json', Accept: 'application/json' }, body: JSON.stringify({ interaction: page.props.interaction, decision }) })
    const payload = await response.json()
    if (!response.ok || !payload.redirect_url) throw new Error(payload.error_description || t('oidcConsent.decisionFailed'))
    window.location.assign(payload.redirect_url)
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : t('oidcConsent.decisionFailed')
    deciding.value = false
  }
}
</script>

<template>
  <main class="relative min-h-screen bg-base-100 text-base-content sm:bg-base-200 sm:px-6 sm:py-8 lg:px-8">
    <Tabs
      :model-value="locale"
      class="absolute right-3 top-3 z-10 sm:right-4 sm:top-4"
      @update:model-value="setLocale($event as Locale)"
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

    <section class="mx-auto flex min-h-screen w-full max-w-[520px] items-stretch justify-center sm:min-h-[calc(100vh-4rem)] sm:items-center">
      <div class="gf-card w-full overflow-hidden border-0 shadow-none sm:border sm:shadow-[0_2px_12px_rgb(0_0_0/calc(var(--gf-depth)*0.04))]">
        <div class="flex min-h-screen flex-col justify-center px-4 py-12 sm:min-h-[470px] sm:px-8 sm:py-6">
          <a href="/" class="mb-6 inline-flex items-baseline text-[27px] font-semibold leading-none tracking-[-0.04em] text-primary">
            <img v-if="page.layout.site.brandType === 'image' && page.layout.site.brandImage" :src="page.layout.site.brandImage" :alt="siteName" class="h-8 w-auto object-contain" />
            <template v-else-if="page.layout.site.brandType === 'text'">{{ page.layout.site.brandText || siteName }}</template>
            <template v-else>Goose<span class="text-base-content">Forum</span></template>
          </a>

          <div class="mb-4">
            <h1 class="text-[27px] font-bold leading-tight tracking-tight text-base-content">{{ t('oidcConsent.title', { site: siteName }) }}</h1>
            <p class="mt-1.5 text-sm leading-6 text-base-content/55">{{ t('oidcConsent.subtitle') }}</p>
          </div>

          <div v-if="!details && !error" class="grid min-h-56 place-items-center">
            <div class="text-center text-sm text-base-content/55"><Loader2 class="mx-auto mb-3 size-6 animate-spin text-primary" />{{ t('oidcConsent.verifying') }}</div>
          </div>

          <div v-else-if="error">
            <p class="gf-status-message gf-status-message-error">{{ error }}</p>
            <Button as-child variant="surface" class="mt-5 w-full px-3">
              <a href="/">{{ t('oidcConsent.back', { site: siteName }) }}</a>
            </Button>
          </div>

          <template v-else-if="details">
            <div class="flex items-center gap-3 rounded-[var(--gf-radius-box)] border border-line bg-base-200/45 p-4">
              <span class="grid size-10 shrink-0 place-items-center rounded-[var(--gf-radius-field)] border border-line bg-base-100 text-base-content"><KeyRound class="size-5" /></span>
              <div class="min-w-0 flex-1">
                <p class="break-words text-sm font-semibold">{{ details.client.name }}</p>
                <p class="mt-0.5 text-xs leading-5 text-base-content/55">{{ t('oidcConsent.accessAccount', { site: siteName }) }}</p>
              </div>
              <ArrowRight class="size-4 shrink-0 text-base-content/35" />
            </div>

            <div class="mt-6">
              <h2 class="text-sm font-semibold">{{ t('oidcConsent.permissions') }}</h2>
              <ul class="mt-3 divide-y divide-line/70 rounded-[var(--gf-radius-box)] border border-line">
                <li v-for="scope in details.scopes" :key="scope" class="flex items-center gap-3 px-4 py-3 text-sm">
                  <Check class="size-4 shrink-0 text-success" />
                  <span>{{ scopeLabels[scope] || scope }}</span>
                </li>
              </ul>
            </div>

            <div class="mt-4 rounded-[var(--gf-radius-field)] bg-base-200/50 px-3 py-2.5 text-xs leading-5 text-base-content/55">
              {{ t('oidcConsent.clientId') }} <span class="ml-1 break-all font-mono text-base-content/65">{{ details.client.id }}</span>
            </div>
            <p class="mt-5 text-xs leading-5 text-base-content/55">{{ t('oidcConsent.trust') }}</p>

            <div class="mt-6 flex flex-wrap gap-3">
              <Button type="button" variant="surface" size="xl" class="flex-1" :disabled="deciding" @click="decide('deny')"><X class="size-4" />{{ t('oidcConsent.deny') }}</Button>
              <Button type="button" variant="brand" size="xl" class="flex-1" :disabled="deciding" @click="decide('approve')">
                <Loader2 v-if="deciding" class="size-4 animate-spin" /><Check v-else class="size-4" />{{ deciding ? t('common.loading') : t('oidcConsent.approve') }}
              </Button>
            </div>
          </template>
        </div>
      </div>
    </section>
  </main>
</template>
