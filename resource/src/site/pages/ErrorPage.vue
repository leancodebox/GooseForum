<script setup lang="ts">
import { computed } from 'vue'
import { ArrowLeft, Home } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { resolveApiMessage } from '@/runtime/api-message'
import type { ErrorPageProps, LayoutPayload } from '@/types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: ErrorPageProps
}>()

const { t } = useI18n()
const localizedMessage = computed(() => resolveApiMessage({
  messageCode: page.props.messageCode,
  params: page.props.params,
}, t('common.loadFailed')))
const localizedTitle = computed(() => page.props.code === '404' ? t('error.notFound.title') : page.props.title)

function goBack() {
  if (window.history.length > 1) {
    window.history.back()
    return
  }
  window.location.href = '/'
}
</script>

<template>
    <main class="min-w-0 pb-12">
      <section class="rounded-lg border border-gray-200/70 bg-white px-6 py-12 text-center shadow-[0_2px_8px_rgba(0,0,0,0.02)]">
        <div class="text-sm font-bold uppercase tracking-wide text-gray-400">{{ page.props.code }}</div>
        <h1 class="mt-2 text-2xl font-bold text-gray-950">{{ localizedTitle }}</h1>
        <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-gray-500">{{ localizedMessage }}</p>
        <div class="mt-6 flex flex-wrap justify-center gap-2">
          <button
            type="button"
            class="inline-flex h-9 items-center gap-1.5 rounded-md border border-gray-200 bg-white px-3 text-sm font-semibold text-gray-700 hover:bg-gray-50"
            @click="goBack"
          >
            <ArrowLeft class="h-4 w-4" />
            {{ t('common.back') }}
          </button>
          <a href="/" class="inline-flex h-9 items-center gap-1.5 rounded-md bg-blue-600 px-3 text-sm font-semibold text-white hover:bg-blue-700">
            <Home class="h-4 w-4" />
            {{ t('common.home') }}
          </a>
        </div>
      </section>
    </main>
</template>
