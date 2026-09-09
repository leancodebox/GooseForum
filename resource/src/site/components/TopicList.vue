<script setup lang="ts">
import { useSlots } from 'vue'
import { useI18n } from 'vue-i18n'
import TopicRow from '@/site/components/TopicRow.vue'
import type { TopicPayload } from '@gooseforum/client'

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
const slots = useSlots()
</script>

<template>
  <table class="gf-topic-table" :aria-label="t('topicList.columns.topic')">
    <thead>
      <tr class="gf-topic-list-header">
        <th scope="col" class="text-left font-normal">{{ t('topicList.columns.topic') }}</th>
        <th scope="col" class="text-center font-normal">{{ t('topicList.columns.users') }}</th>
        <th scope="col" class="text-center font-normal">{{ t('topicList.columns.replies') }}</th>
        <th scope="col" class="text-center font-normal">{{ t('topicList.columns.views') }}</th>
        <th scope="col" class="text-right font-normal">
          <slot name="activity-header">
            {{ t('topicList.columns.activity') }}
          </slot>
        </th>
      </tr>
    </thead>
    <tbody class="relative bg-base-100">
      <TopicRow
        v-for="topic in topics"
        :key="topic.id"
        :topic="topic"
        :home="home"
        :show-categories="showCategories"
        :show-hot="showHot"
        :show-pinned="showPinned"
      >
        <template v-if="slots.activity" #activity="{ topic: rowTopic }">
          <slot name="activity" :topic="rowTopic" />
        </template>
        <template v-if="slots['mobile-action']" #mobile-action="{ topic: rowTopic }">
          <slot name="mobile-action" :topic="rowTopic" />
        </template>
      </TopicRow>
    </tbody>
  </table>
  <slot name="empty" />
</template>
