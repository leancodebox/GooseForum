<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Plus, UsersRound } from '@lucide/vue'
import EmptyState from '@/site/components/EmptyState.vue'
import PageHeader from '@/site/components/PageHeader.vue'
import TopicListFooter from '@/site/components/TopicListFooter.vue'
import TopicListModeSwitch from '@/site/components/TopicListModeSwitch.vue'
import TopicList from '@/site/components/TopicList.vue'
import { useTopicList } from '@/site/composables/useTopicList'
import type { CategoryPageProps, LayoutPayload } from '@gooseforum/client'

const page = defineProps<{
  layout: LayoutPayload
  props: CategoryPageProps
  pageUrl: string
}>()
const { t } = useI18n()
const { topics, pagination, hasTopics, listMode, setListMode, loadingMore, loadError, loadMoreSentinel, loadMore } = useTopicList(page)

function sortTabLabel(key: string, fallback?: string) {
  if (key === 'latest') return t('topicList.tabs.latestReplies')
  if (key === 'new') return t('topicList.tabs.latestPublished')
  if (key === 'hot') return t('topicList.tabs.hot')
  if (key === 'popular') return t('topicList.tabs.popular')
  return fallback || key
}

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
            <nav class="gf-home-topic-tabs" :aria-label="t('topicList.columns.topic')">
              <a
                v-for="tab in page.props.tabs"
                :key="tab.key"
                :href="tab.url"
                :aria-current="tab.active ? 'page' : undefined"
                class="gf-tab"
                :class="tab.active ? 'gf-tab-active' : 'gf-tab-idle'"
              >
                {{ sortTabLabel(tab.key, tab.label) }}
              </a>
            </nav>
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
