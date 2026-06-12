<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Search, UsersRound } from '@lucide/vue'
import { formatNumber } from '@/runtime/format'
import TopicList from '@/site/components/TopicList.vue'
import type { LayoutPayload, SearchPageProps } from '@/types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: SearchPageProps
}>()
const { t } = useI18n()

const query = ref(page.props.query)
const topics = computed(() => page.props.topics || [])
const hasQuery = computed(() => (page.props.query || '').trim().length > 0)
const hasResults = computed(() => topics.value.length > 0)

watch(
  () => page.props.query,
  () => {
    query.value = page.props.query
  },
)
</script>

<template>
    <main class="min-w-0 pb-12">
      <section class="gf-card overflow-hidden">
        <header class="border-b border-line px-4 py-4">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <h1 class="text-2xl font-bold text-base-content">{{ t('searchPage.title') }}</h1>
                <span class="rounded-md bg-base-300 px-2 py-0.5 text-[11px] font-bold uppercase text-base-content/75">{{ t('searchPage.label') }}</span>
              </div>
              <p class="mt-1 text-sm text-base-content/55">
                <template v-if="hasQuery">
                  “<span class="font-semibold text-base-content">{{ page.props.query }}</span>”
                  <span class="text-base-content/35"> · </span>
                  <span>{{ t('searchPage.resultCount', { count: formatNumber(page.props.total) }) }}</span>
                </template>
                <template v-else>
                  {{ t('searchPage.emptyPrompt') }}
                </template>
              </p>
            </div>

            <form action="/search" method="GET" class="w-full lg:max-w-xl">
              <label class="flex h-11 items-center gap-2 rounded-md border border-line bg-base-200 px-3 text-sm text-base-content/55 transition focus-within:border-primary focus-within:bg-base-100 focus-within:ring-4 focus-within:ring-primary/20">
                <Search class="h-4 w-4 shrink-0" />
                <input v-model="query" name="q" class="min-w-0 flex-1 bg-transparent text-base-content outline-none placeholder:text-base-content/55" :placeholder="t('searchPage.inputPlaceholder')" />
                <button type="submit" class="gf-button gf-button-sm gf-button-neutral shrink-0">{{ t('common.search') }}</button>
              </label>
            </form>
          </div>
        </header>

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

        <div v-else-if="hasQuery" class="px-6 py-16 text-center">
          <UsersRound class="mx-auto h-8 w-8 text-base-content/35" />
          <h2 class="mt-3 text-base font-semibold text-base-content">{{ t('searchPage.noResultsTitle') }}</h2>
          <p class="mt-1 text-sm text-base-content/55">{{ t('searchPage.noResultsDescription') }}</p>
        </div>

        <div v-else class="px-6 py-16 text-center">
          <Search class="mx-auto h-8 w-8 text-base-content/35" />
          <h2 class="mt-3 text-base font-semibold text-base-content">{{ t('searchPage.startTitle') }}</h2>
          <p class="mt-1 text-sm text-base-content/55">{{ t('searchPage.startDescription') }}</p>
        </div>
      </section>
    </main>
</template>
