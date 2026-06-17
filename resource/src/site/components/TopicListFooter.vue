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
</script>

<template>
  <footer class="border-t border-line bg-base-200/50 p-3 text-center">
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
      <button
        v-if="pagination.hasNext"
        type="button"
        class="gf-button gf-button-sm gf-button-ghost gap-2 disabled:cursor-wait"
        :disabled="loadingMore"
        @click="emit('loadMore')"
      >
        <Loader2 v-if="loadingMore" class="h-4 w-4 animate-spin" />
        {{ loadingMore ? t('common.loadingShort') : t('common.loadMore') }}
      </button>
      <p v-else-if="hasTopics" class="text-xs font-medium text-base-content/55">{{ t('topicList.allShown') }}</p>
      <p v-if="loadError" class="mt-2 text-xs text-error">{{ t('topicList.autoLoadFailed') }}</p>
      <a v-if="pagination.hasNext" :href="pagination.nextUrl" rel="next" class="sr-only">{{ t('common.nextPage') }}</a>
    </template>
  </footer>
</template>
