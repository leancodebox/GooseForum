<script setup lang="ts">
import { MessageSquare, Pin, Sparkles } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { formatNumber, timeAgo } from '@/runtime/format'
import { topicDescription } from '@/runtime/topic-description'
import AvatarStack from '@/site/components/AvatarStack.vue'
import type { TopicPayload } from '@gooseforum/client'

withDefaults(defineProps<{
  topic: TopicPayload
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
  <tr
    class="group gf-topic-row"
    :class="[
      home ? 'gf-topic-row-home' : '',
      topic.pinWeight > 0 ? 'gf-topic-row-pinned' : '',
    ]"
  >
    <th scope="row" class="min-w-0 text-left font-normal">
      <div class="flex min-h-6 min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
        <h2 class="m-0 inline-flex min-w-0 max-w-full items-center gap-2">
          <Pin
            v-if="showPinned && topic.pinWeight > 0"
            class="h-3.5 w-3.5 shrink-0 rotate-45 text-error"
            :aria-label="t('topicList.pinned')"
          />
          <a :href="topic.url" rel="bookmark" class="min-w-0 truncate text-[15px] font-medium leading-6 text-base-content group-hover:text-primary sm:text-base">
            {{ topic.title }}
          </a>
          <span
            v-if="topic.unseen"
            class="h-2 w-2 shrink-0 rounded-full bg-primary"
            aria-hidden="true"
          />
        </h2>
        <a
          v-for="category in showCategories ? topic.categories : []"
          :key="category.id"
          :href="category.url"
          rel="tag"
          class="gf-topic-chip"
          :style="{ '--gf-topic-chip-color': category.color || 'var(--gf-color-primary)' }"
        >
          <span class="gf-topic-chip-dot" />
          {{ category.name }}
        </a>
        <span v-if="showHot && topic.viewCount > 500" class="inline-flex h-5 items-center gap-1 text-[11px] font-semibold text-warning">
          <Sparkles class="h-3 w-3" /> hot
        </span>
      </div>
      <p class="mt-1 min-h-5 truncate text-[13px] leading-5 text-base-content/55">{{ topicDescription(topic) }}</p>
      <div class="mt-1.5 flex min-h-6 flex-wrap items-center gap-x-3 gap-y-1 text-xs text-base-content/55 lg:hidden">
        <AvatarStack :users="topic.participants" size="sm" />
        <a :href="topic.url"><time :datetime="topic.lastUpdateTime">{{ timeAgo(topic.lastUpdateTime) }}</time></a>
        <a :href="topic.url" :aria-label="`${topic.title} · ${t('topicList.columns.replies')}: ${topic.replyCount}`" class="inline-flex items-center gap-1">
          <MessageSquare class="h-3.5 w-3.5" /> {{ formatNumber(topic.replyCount) }}
        </a>
        <slot name="mobile-action" :topic="topic" />
      </div>
    </th>
    <td class="hidden justify-center lg:flex">
      <AvatarStack :users="topic.participants" />
    </td>
    <td class="hidden text-center text-sm font-semibold tabular-nums text-base-content/75 lg:block">
      <a :href="topic.url" :aria-label="`${topic.title} · ${t('topicList.columns.replies')}: ${topic.replyCount}`">
        {{ formatNumber(topic.replyCount) }}
      </a>
    </td>
    <td class="hidden text-center text-sm tabular-nums text-base-content/55 lg:block">{{ formatNumber(topic.viewCount) }}</td>
    <td class="hidden text-right text-[13px] font-medium tabular-nums text-base-content/55 lg:block">
      <slot name="activity" :topic="topic">
        <a :href="topic.url"><time :datetime="topic.lastUpdateTime">{{ timeAgo(topic.lastUpdateTime) }}</time></a>
      </slot>
    </td>
  </tr>
</template>
