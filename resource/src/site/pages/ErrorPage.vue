<script setup lang="ts">
import { computed } from 'vue'
import { ArrowLeft, Home } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import { resolveApiMessage } from '@/runtime/api-message'
import EmptyState from '@/site/components/EmptyState.vue'
import type { ErrorPageProps, LayoutPayload } from '@gooseforum/client'

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
          <Button
            type="button"
            variant="surface"
            class="px-3"
            @click="goBack"
          >
            <ArrowLeft class="h-4 w-4" />
            {{ t('common.back') }}
          </Button>
          <Button as-child variant="brand" class="px-3">
            <a href="/">
              <Home class="h-4 w-4" />
              {{ t('common.home') }}
            </a>
          </Button>
        </EmptyState>
      </section>
    </main>
</template>
