<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessageSquare, Search, Sparkles, UsersRound } from '@lucide/vue'
import { formatNumber, timeAgo } from '@/runtime/format'
import { topicDescription } from '@/runtime/topic-description'
import { showUserCard } from '@/runtime/user-card-events'
import UserAvatar from '@/site/components/UserAvatar.vue'
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
      <section class="overflow-hidden rounded-lg border border-line/70 bg-base-100 shadow-[0_2px_12px_rgba(0,0,0,0.04)]">
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
                <button type="submit" class="shrink-0 rounded-md bg-neutral px-3 py-1.5 text-sm font-semibold text-neutral-content hover:bg-neutral/90">{{ t('common.search') }}</button>
              </label>
            </form>
          </div>
        </header>

        <template v-if="hasResults">
          <div class="hidden grid-cols-[minmax(0,1fr)_112px_72px_72px_88px] border-b border-line bg-base-200/60 px-4 py-2 text-[11px] font-bold uppercase text-base-content/75 lg:grid">
            <div>{{ t('topicList.columns.topic') }}</div>
            <div class="text-center">{{ t('topicList.columns.users') }}</div>
            <div class="text-center">{{ t('topicList.columns.replies') }}</div>
            <div class="text-center">{{ t('topicList.columns.views') }}</div>
            <div class="text-right">{{ t('topicList.columns.activity') }}</div>
          </div>

          <div class="relative bg-base-100">
            <article
              v-for="topic in topics"
              :key="topic.id"
              class="group gf-topic-row"
            >
              <div class="min-w-0">
                <div class="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
                  <a :href="topic.url" class="truncate text-[15px] font-semibold leading-snug text-base-content group-hover:text-primary sm:text-base">
                    {{ topic.title }}
                  </a>
                  <a
                    v-for="category in topic.categories"
                    :key="category.id"
                    :href="category.url"
                    class="inline-flex h-6 items-center gap-1.5 rounded-full bg-base-300 px-2 text-[11px] font-medium text-base-content/55 hover:bg-info/10 hover:text-primary"
                  >
                    <span class="h-1.5 w-1.5 rounded-full" :style="{ backgroundColor: category.color }" />
                    {{ category.name }}
                  </a>
                  <span v-if="topic.viewCount > 500" class="inline-flex h-6 items-center gap-1 text-[11px] font-semibold text-warning">
                    <Sparkles class="h-3 w-3" /> hot
                  </span>
                </div>
                <p class="mt-1 truncate text-[13px] leading-relaxed text-base-content/55">{{ topicDescription(topic) }}</p>
                <div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-base-content/55 lg:hidden">
                  <div class="flex -space-x-2">
                    <a
                      v-for="participant in topic.participants"
                      :key="participant.id"
                      :href="`/u/${participant.id}`"
                      :title="participant.username"
                      class="rounded-full ring-2 ring-base-100 transition hover:z-10 hover:scale-110"
                      @click="showUserCard(participant, $event)"
                    >
                      <UserAvatar :src="participant.avatarUrl" :alt="participant.username" class="h-6 w-6 rounded-full object-cover" />
                    </a>
                  </div>
                  <span>{{ timeAgo(topic.lastUpdateTime) }}</span>
                  <span class="inline-flex items-center gap-1">
                    <MessageSquare class="h-3.5 w-3.5" /> {{ formatNumber(topic.replyCount) }}
                  </span>
                </div>
              </div>
              <div class="hidden justify-center lg:flex">
                <div class="flex -space-x-3">
                  <a
                    v-for="participant in topic.participants"
                    :key="participant.id"
                    :href="`/u/${participant.id}`"
                    :title="participant.username"
                    class="rounded-full ring-2 ring-base-100 transition hover:z-10 hover:scale-110"
                    @click="showUserCard(participant, $event)"
                  >
                    <UserAvatar :src="participant.avatarUrl" :alt="participant.username" class="h-8 w-8 rounded-full object-cover" />
                  </a>
                </div>
              </div>
              <div class="hidden text-center text-sm font-semibold tabular-nums text-base-content/75 lg:block">{{ formatNumber(topic.replyCount) }}</div>
              <div class="hidden text-center text-sm tabular-nums text-base-content/55 lg:block">{{ formatNumber(topic.viewCount) }}</div>
              <div class="hidden text-right text-[13px] font-medium tabular-nums text-base-content/55 lg:block">{{ timeAgo(topic.lastUpdateTime) }}</div>
            </article>
          </div>

          <footer v-if="page.props.totalPages > 1" class="flex flex-col gap-3 border-t border-line bg-base-200/50 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
            <div class="text-sm text-base-content/55">
              {{ t('common.page', { page: page.props.pagination.page, total: page.props.totalPages }) }}
            </div>
            <div class="flex items-center gap-2">
              <a
                v-if="page.props.pagination.page > 1"
                :href="`/search?q=${encodeURIComponent(page.props.query)}&page=${page.props.pagination.page - 1}`"
                class="rounded-md border border-line bg-base-100 px-3 py-1.5 text-sm font-semibold text-base-content/75 hover:bg-base-200 hover:text-base-content"
              >
                {{ t('common.previousPage') }}
              </a>
              <a
                v-if="page.props.pagination.hasNext"
                :href="page.props.pagination.nextUrl"
                class="rounded-md border border-line bg-base-100 px-3 py-1.5 text-sm font-semibold text-base-content/75 hover:bg-base-200 hover:text-base-content"
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
