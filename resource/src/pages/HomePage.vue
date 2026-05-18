<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Loader2, MessageSquare, Plus, Sparkles, UsersRound } from '@lucide/vue'
import AppShell from '../components/AppShell.vue'
import { formatNumber, timeAgo } from '../runtime/format'
import { fetchPage } from '../runtime/router'
import { scheduleHideUserCard, showUserCard } from '../runtime/user-card-events'
import type { HomeProps, LayoutPayload, PagePayload, TopicPayload } from '../types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: HomeProps
}>()

const topics = ref<TopicPayload[]>([])
const pagination = ref<HomeProps['pagination']>(page.props.pagination)
const loadingMore = ref(false)
const loadError = ref('')
const loadMoreSentinel = ref<HTMLElement | null>(null)
let observer: IntersectionObserver | undefined

const hasTopics = computed(() => topics.value.length > 0)

watch(
  () => [page.props.sort, page.props.pagination.page, page.props.topics],
  () => {
    topics.value = [...page.props.topics]
    pagination.value = page.props.pagination
    loadError.value = ''
    void nextTick(observeSentinel)
  },
  { immediate: true },
)

async function loadMore() {
  if (loadingMore.value || !pagination.value.hasNext || !pagination.value.nextUrl) return

  loadingMore.value = true
  loadError.value = ''
  try {
    const payload = (await fetchPage(new URL(pagination.value.nextUrl, window.location.origin))) as PagePayload<HomeProps>
    topics.value = mergeTopics(topics.value, payload.props.topics)
    pagination.value = payload.props.pagination
  } catch (error) {
    loadError.value = error instanceof Error ? error.message : '加载失败'
  } finally {
    loadingMore.value = false
  }
}

function mergeTopics(current: TopicPayload[], incoming: TopicPayload[]) {
  const seen = new Set(current.map((topic) => topic.id))
  return [...current, ...incoming.filter((topic) => !seen.has(topic.id))]
}

function observeSentinel() {
  observer?.disconnect()
  if (!loadMoreSentinel.value || !('IntersectionObserver' in window)) return
  observer = new IntersectionObserver(
    (entries) => {
      if (entries.some((entry) => entry.isIntersecting)) void loadMore()
    },
    { rootMargin: '480px 0px' },
  )
  observer.observe(loadMoreSentinel.value)
}

onMounted(observeSentinel)

onBeforeUnmount(() => {
  observer?.disconnect()
})
</script>

<template>
  <AppShell :layout="layout">
    <div class="pb-12">
      <aside v-if="page.props.announcement.enabled" class="mb-3 rounded-lg border border-gray-200/80 bg-white shadow-[0_2px_12px_rgba(0,0,0,0.035)]" aria-label="公告">
        <div class="border-l-[3px] border-blue-500 px-4 py-3">
          <div class="mb-2 flex items-center gap-2">
            <span class="h-1.5 w-1.5 rounded-full bg-blue-500" />
            <span class="text-[11px] font-bold uppercase text-gray-600">Announcement</span>
          </div>
          <div class="gf-prose gf-prose-announcement" v-html="page.props.announcement.html" />
        </div>
      </aside>

      <section class="overflow-hidden rounded-lg border border-gray-200/70 bg-white shadow-[0_2px_12px_rgba(0,0,0,0.04)]">
        <div class="flex flex-col gap-3 border-b border-gray-100 bg-white px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
          <div class="flex min-w-0 items-center gap-2 overflow-x-auto">
            <a
              v-for="tab in page.props.tabs"
              :key="tab.key"
              :href="tab.url"
              class="inline-flex h-8 shrink-0 items-center rounded-md px-3 text-sm font-semibold"
              :class="tab.active ? 'bg-gray-900 text-white' : 'text-gray-500 hover:bg-gray-100 hover:text-gray-950'"
            >
              {{ tab.label }}
            </a>
          </div>
          <a href="/publish" class="inline-flex h-8 items-center justify-center gap-1.5 rounded-md bg-blue-600 px-3 text-sm font-semibold text-white hover:bg-blue-700">
            <Plus class="h-4 w-4" />
            新建主题
          </a>
        </div>

        <div class="hidden gap-3 grid-cols-[minmax(0,1fr)_112px_72px_72px_88px] border-b border-gray-100 bg-gray-50/60 px-4 py-2 text-[11px] font-bold uppercase text-gray-600 lg:grid">
          <div>Topic</div>
          <div class="text-center">Users</div>
          <div class="text-center">Replies</div>
          <div class="text-center">Views</div>
          <div class="text-right">Activity</div>
        </div>

        <div class="relative bg-white">
          <article
            v-for="topic in topics"
            :key="topic.id"
            class="group relative isolate grid gap-3 bg-white px-4 py-3 transition before:absolute before:inset-0 before:-z-10 before:bg-white before:content-[''] after:absolute after:inset-x-4 after:bottom-0 after:h-px after:bg-gray-100/70 after:content-[''] last:after:hidden hover:before:bg-gray-50 lg:grid-cols-[minmax(0,1fr)_112px_72px_72px_88px] lg:items-center"
          >
            <div class="min-w-0">
              <div class="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
                <a :href="topic.url" class="truncate text-[15px] font-semibold leading-snug text-gray-950 group-hover:text-blue-600 sm:text-base">
                  {{ topic.title }}
                </a>
                <a
                  v-for="category in topic.categories"
                  :key="category.id"
                  :href="category.url"
                  class="inline-flex h-6 items-center gap-1.5 rounded-full bg-gray-100 px-2 text-[11px] font-medium text-gray-500 hover:bg-blue-50 hover:text-blue-600"
                >
                  <span class="h-1.5 w-1.5 rounded-full" :style="{ backgroundColor: category.color }" />
                  {{ category.name }}
                </a>
                <span v-if="topic.viewCount > 500" class="inline-flex h-6 items-center gap-1 text-[11px] font-semibold text-orange-500">
                  <Sparkles class="h-3 w-3" /> hot
                </span>
              </div>
              <p v-if="topic.description" class="mt-1 truncate text-[13px] leading-relaxed text-gray-500">{{ topic.description }}</p>
              <div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-400 lg:hidden">
                <div class="flex -space-x-2">
                  <a
                    v-for="participant in topic.participants"
                    :key="participant.id"
                    :href="`/u/${participant.id}`"
                    :title="participant.username"
                    class="rounded-full ring-2 ring-white transition hover:z-10 hover:scale-110"
                    @mouseenter="showUserCard(participant, $event)"
                    @mouseleave="scheduleHideUserCard"
                    @focus="showUserCard(participant, $event)"
                    @blur="scheduleHideUserCard"
                  >
                    <img :src="participant.avatarUrl" :alt="participant.username" class="h-6 w-6 rounded-full object-cover" />
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
                    class="rounded-full ring-2 ring-white transition hover:z-10 hover:scale-110"
                    @mouseenter="showUserCard(participant, $event)"
                    @mouseleave="scheduleHideUserCard"
                    @focus="showUserCard(participant, $event)"
                    @blur="scheduleHideUserCard"
                  >
                    <img :src="participant.avatarUrl" :alt="participant.username" class="h-8 w-8 rounded-full object-cover" />
                  </a>
              </div>
            </div>
            <div class="hidden text-center text-sm font-semibold tabular-nums text-gray-700 lg:block">{{ formatNumber(topic.replyCount) }}</div>
            <div class="hidden text-center text-sm tabular-nums text-gray-500 lg:block">{{ formatNumber(topic.viewCount) }}</div>
            <div class="hidden text-right text-[13px] font-medium tabular-nums text-gray-400 lg:block">{{ timeAgo(topic.lastUpdateTime) }}</div>
          </article>

          <div v-if="!hasTopics" class="px-5 py-16 text-center">
            <UsersRound class="mx-auto h-8 w-8 text-gray-300" />
            <h2 class="mt-3 text-base font-semibold text-gray-900">暂无主题</h2>
            <p class="mt-1 text-sm text-gray-500">这里还没有公开主题。</p>
          </div>
        </div>

        <div ref="loadMoreSentinel" class="border-t border-gray-100 bg-gray-50/50 p-3 text-center">
          <button
            v-if="pagination.hasNext"
            type="button"
            class="inline-flex h-8 items-center gap-2 rounded-md px-3 text-sm font-semibold text-blue-600 hover:bg-blue-50 hover:text-blue-700 disabled:cursor-wait disabled:opacity-70"
            :disabled="loadingMore"
            @click="loadMore"
          >
            <Loader2 v-if="loadingMore" class="h-4 w-4 animate-spin" />
            {{ loadingMore ? '加载中...' : '加载更多' }}
          </button>
          <p v-else-if="hasTopics" class="text-xs font-medium text-gray-400">已显示全部主题</p>
          <p v-if="loadError" class="mt-2 text-xs text-red-600">自动加载失败，可以稍后重试。</p>
          <a v-if="pagination.hasNext" :href="pagination.nextUrl" rel="next" class="sr-only">下一页</a>
        </div>
      </section>
    </div>
  </AppShell>
</template>
