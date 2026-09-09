<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, ChevronRight, Loader2 } from '@lucide/vue'
import type { TopicListMode } from '@/site/composables/useTopicListMode'

const props = defineProps<{
  pagination: {
    page: number
    nextPage: number
    hasNext: boolean
    nextUrl: string
  }
  mode: TopicListMode
  loadingMore: boolean
  hasTopics: boolean
  loadError: string
}>()

const emit = defineEmits<{
  loadMore: []
}>()

const { t } = useI18n()

const previousUrl = computed(() => {
  if (props.pagination.page <= 1 || typeof window === 'undefined') return ''
  const url = new URL(window.location.href)
  const previousPage = props.pagination.page - 1
  if (previousPage <= 1) {
    url.searchParams.delete('page')
  } else {
    url.searchParams.set('page', String(previousPage))
  }
  return `${url.pathname}${url.search}${url.hash}`
})
// Keep native navigation for new-tab/modifier clicks and enhance ordinary clicks.
function loadNextPage(event: MouseEvent) {
  if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return
  event.preventDefault()
  if (!props.loadingMore) emit('loadMore')
}
</script>

<template>
  <nav :aria-label="t('topicList.mode.pagination')" class="border-t border-line bg-base-200/50 p-3 text-center">
    <template v-if="mode === 'pagination'">
      <div class="flex flex-wrap items-center justify-center gap-2">
        <a
          v-if="previousUrl"
          :href="previousUrl"
          rel="prev"
          class="gf-button gf-button-sm gf-button-secondary"
        >
          <ChevronLeft class="h-4 w-4" />
          {{ t('common.previousPage') }}
        </a>
        <span class="inline-flex h-8 items-center px-2 text-xs font-semibold text-base-content/55" style="border-radius: var(--gf-radius-field)">
          {{ t('common.currentPage', { page: pagination.page }) }}
        </span>
        <a
          v-if="pagination.hasNext"
          :href="pagination.nextUrl"
          rel="next"
          class="gf-button gf-button-sm gf-button-secondary"
        >
          {{ t('common.nextPage') }}
          <ChevronRight class="h-4 w-4" />
        </a>
      </div>
      <p v-if="!pagination.hasNext && hasTopics" class="mt-2 text-xs font-medium text-base-content/55">{{ t('topicList.allShown') }}</p>
    </template>

    <template v-else>
      <a
        v-if="pagination.hasNext"
        :href="pagination.nextUrl"
        rel="next"
        class="gf-button gf-button-sm gf-button-ghost gap-2 aria-disabled:cursor-wait"
        :aria-disabled="loadingMore"
        @click="loadNextPage"
      >
        <Loader2 v-if="loadingMore" class="h-4 w-4 animate-spin" />
        {{ loadingMore ? t('common.loadingShort') : t('common.loadMore') }}
      </a>
      <p v-else-if="hasTopics" class="text-xs font-medium text-base-content/55">{{ t('topicList.allShown') }}</p>
      <p v-if="loadError" class="mt-2 text-xs text-error">{{ t('topicList.autoLoadFailed') }}</p>
    </template>
  </nav>
</template>
