<script setup lang="ts">
import { computed } from 'vue'
import { ArrowLeft, Home } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { resolveApiMessage } from '@/runtime/api-message'
import EmptyState from '@/site/components/EmptyState.vue'
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
    <main class="min-w-0 pb-8">
      <section class="gf-card overflow-hidden">
        <EmptyState :title="`${page.props.code} · ${localizedTitle}`" :description="localizedMessage">
          <button
            type="button"
            class="gf-button gf-button-md gf-button-secondary"
            @click="goBack"
          >
            <ArrowLeft class="h-4 w-4" />
            {{ t('common.back') }}
          </button>
          <a href="/" class="gf-button gf-button-md gf-button-primary">
            <Home class="h-4 w-4" />
            {{ t('common.home') }}
          </a>
        </EmptyState>
      </section>
    </main>
</template>
