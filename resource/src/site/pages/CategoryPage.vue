<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Plus, UsersRound } from '@lucide/vue'
import { fetchPage } from '@/runtime/router'
import EmptyState from '@/site/components/EmptyState.vue'
import PageHeader from '@/site/components/PageHeader.vue'
import TopicListFooter from '@/site/components/TopicListFooter.vue'
import TopicListModeSwitch from '@/site/components/TopicListModeSwitch.vue'
import TopicList from '@/site/components/TopicList.vue'
import { useTopicListMode } from '@/site/composables/useTopicListMode'
import type { CategoryPageProps, LayoutPayload, PagePayload, TopicPayload } from '@/types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: CategoryPageProps
}>()
const { t } = useI18n()
const { mode: listMode, setMode: setListMode } = useTopicListMode()

const topics = ref<TopicPayload[]>([])
const pagination = ref<CategoryPageProps['pagination']>(page.props.pagination)
const loadingMore = ref(false)
const loadError = ref('')
const loadMoreSentinel = ref<HTMLElement | null>(null)
let observer: IntersectionObserver | undefined

const hasTopics = computed(() => topics.value.length > 0)
const isWaterfallMode = computed(() => listMode.value === 'waterfall')

watch(
  () => [page.props.category.id, page.props.sort, page.props.pagination.page, page.props.topics],
  () => {
    topics.value = [...page.props.topics]
    pagination.value = page.props.pagination
    loadError.value = ''
    void nextTick(observeSentinel)
  },
  { immediate: true },
)

watch(listMode, (mode) => {
  if (mode === 'pagination') {
    observer?.disconnect()
    topics.value = [...page.props.topics]
    pagination.value = page.props.pagination
    return
  }
  void nextTick(observeSentinel)
})

async function loadMore() {
  if (!isWaterfallMode.value || loadingMore.value || !pagination.value.hasNext || !pagination.value.nextUrl) return

  loadingMore.value = true
  loadError.value = ''
  try {
    const payload = (await fetchPage(new URL(pagination.value.nextUrl, window.location.origin))) as PagePayload<CategoryPageProps>
    topics.value = mergeTopics(topics.value, payload.props.topics)
    pagination.value = payload.props.pagination
  } catch (error) {
    loadError.value = error instanceof Error ? error.message : t('common.loadFailed')
  } finally {
    loadingMore.value = false
  }
}

function mergeTopics(current: TopicPayload[], incoming: TopicPayload[]) {
  const seen = new Set(current.map((topic) => topic.id))
  return [...current, ...incoming.filter((topic) => !seen.has(topic.id))]
}

function sortTabLabel(key: string, fallback?: string) {
  if (key === 'latest') return t('topicList.tabs.latestReplies')
  if (key === 'new') return t('topicList.tabs.latestPublished')
  if (key === 'hot') return t('topicList.tabs.hot')
  if (key === 'popular') return t('topicList.tabs.popular')
  return fallback || key
}

function observeSentinel() {
  observer?.disconnect()
  if (!isWaterfallMode.value || !loadMoreSentinel.value || !('IntersectionObserver' in window)) return
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
    <div class="pb-12">
      <PageHeader :title="page.props.category.name" :description="page.props.category.description" compact>
        <template #badge>
          <span class="h-1.5 w-1.5 shrink-0 rounded-full" :style="{ backgroundColor: page.props.category.color || 'var(--gf-color-primary)' }" />
          <span class="text-xs font-semibold uppercase text-base-content/75">{{ t('category.label') }}</span>
        </template>
      </PageHeader>

      <section class="gf-card overflow-hidden">
        <div class="gf-home-topic-toolbar">
          <div class="gf-home-topic-tools">
            <div class="gf-home-topic-tabs">
              <a
                v-for="tab in page.props.tabs"
                :key="tab.key"
                :href="tab.url"
                class="gf-tab"
                :class="tab.active ? 'gf-tab-active' : 'gf-tab-idle'"
              >
                {{ sortTabLabel(tab.key, tab.label) }}
              </a>
            </div>
            <TopicListModeSwitch :model-value="listMode" @update:model-value="setListMode" />
          </div>
          <a href="/publish" class="gf-button gf-button-md gf-button-primary shrink-0 whitespace-nowrap px-3 sm:h-8">
            <Plus class="h-4 w-4" />
            {{ t('topicList.newTopic') }}
          </a>
        </div>

        <TopicList :topics="topics" :show-categories="false" :show-hot="false">
          <template #empty>
            <EmptyState v-if="!hasTopics" :icon="UsersRound" :title="t('topicList.emptyTitle')" :description="t('topicList.emptyCategoryDescription')" />
          </template>
        </TopicList>

        <div ref="loadMoreSentinel">
          <TopicListFooter
            :pagination="pagination"
            :mode="listMode"
            :loading-more="loadingMore"
            :has-topics="hasTopics"
            :load-error="loadError"
            @load-more="loadMore"
          />
        </div>
      </section>
    </div>
</template>
