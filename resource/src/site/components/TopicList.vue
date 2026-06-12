<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import TopicRow from '@/site/components/TopicRow.vue'
import type { TopicPayload } from '@/types/payload'

withDefaults(defineProps<{
  topics: TopicPayload[]
  home?: boolean
  showCategories?: boolean
  showHot?: boolean
  showPinned?: boolean
}>(), {
  home: false,
  showCategories: true,
  showHot: true,
  showPinned: false,
})

const { t } = useI18n()
</script>

<template>
  <div class="gf-topic-list-header">
    <div>{{ t('topicList.columns.topic') }}</div>
    <div class="text-center">{{ t('topicList.columns.users') }}</div>
    <div class="text-center">{{ t('topicList.columns.replies') }}</div>
    <div class="text-center">{{ t('topicList.columns.views') }}</div>
    <div class="text-right">{{ t('topicList.columns.activity') }}</div>
  </div>

  <div class="relative bg-base-100">
    <TopicRow
      v-for="topic in topics"
      :key="topic.id"
      :topic="topic"
      :home="home"
      :show-categories="showCategories"
      :show-hot="showHot"
      :show-pinned="showPinned"
    />
    <slot name="empty" />
  </div>
</template>
