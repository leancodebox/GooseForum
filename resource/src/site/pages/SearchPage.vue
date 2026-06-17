<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Search, UsersRound } from '@lucide/vue'
import { formatNumber } from '@/runtime/format'
import EmptyState from '@/site/components/EmptyState.vue'
import PageHeader from '@/site/components/PageHeader.vue'
import TopicList from '@/site/components/TopicList.vue'
import type { LayoutPayload, SearchPageProps } from '@/types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: SearchPageProps
  pageUrl: string
}>()
const { t } = useI18n()

const query = ref(page.props.query)
const topics = computed(() => page.props.topics || [])
const hasQuery = computed(() => (page.props.query || '').trim().length > 0)
const hasResults = computed(() => topics.value.length > 0)
const searchDescription = computed(() => {
  if (!hasQuery.value) return t('searchPage.emptyPrompt')
  return `${page.props.query} · ${t('searchPage.resultCount', { count: formatNumber(page.props.total) })}`
})

watch(
  () => page.props.query,
  () => {
    query.value = page.props.query
  },
)
</script>

<template>
    <main class="min-w-0 pb-8">
      <PageHeader :title="t('searchPage.title')" :description="searchDescription" compact>
        <template #badge>
          <span class="gf-badge gf-badge-muted h-5 text-[11px] uppercase">{{ t('searchPage.label') }}</span>
        </template>
        <template #actions>
          <form action="/search" method="GET" class="w-full sm:w-80 lg:w-96">
            <label class="flex h-10 items-center gap-2 rounded-field border border-line bg-base-100 px-3 text-sm text-base-content/55 transition focus-within:border-primary focus-within:ring-4 focus-within:ring-primary/20">
              <Search class="h-4 w-4 shrink-0" />
              <input v-model="query" name="q" class="min-w-0 flex-1 bg-transparent text-base-content outline-none placeholder:text-base-content/55" :placeholder="t('searchPage.inputPlaceholder')" />
              <button type="submit" class="gf-button gf-button-sm gf-button-neutral shrink-0">{{ t('common.search') }}</button>
            </label>
          </form>
        </template>
      </PageHeader>

      <section class="gf-card overflow-hidden">
        <template v-if="hasResults">
          <TopicList :topics="topics" />

          <footer v-if="page.props.totalPages > 1" class="flex flex-col gap-3 border-t border-line bg-base-200/50 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
            <div class="text-sm text-base-content/55">
              {{ t('common.page', { page: page.props.pagination.page, total: page.props.totalPages }) }}
            </div>
            <div class="flex items-center gap-2">
              <a
                v-if="page.props.pagination.page > 1"
                :href="`/search?q=${encodeURIComponent(page.props.query)}&page=${page.props.pagination.page - 1}`"
                class="gf-button gf-button-sm gf-button-secondary"
              >
                {{ t('common.previousPage') }}
              </a>
              <a
                v-if="page.props.pagination.hasNext"
                :href="page.props.pagination.nextUrl"
                class="gf-button gf-button-sm gf-button-secondary"
                rel="next"
              >
                {{ t('common.nextPage') }}
              </a>
            </div>
          </footer>
        </template>

        <EmptyState v-else-if="hasQuery" :icon="UsersRound" :title="t('searchPage.noResultsTitle')" :description="t('searchPage.noResultsDescription')" />

        <EmptyState v-else :icon="Search" :title="t('searchPage.startTitle')" :description="t('searchPage.startDescription')" />
      </section>
    </main>
</template>
