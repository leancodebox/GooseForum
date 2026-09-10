<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { formatDateTime, formatNumber } from '@/runtime/format'
import type { PostPayload } from '@gooseforum/client'

defineProps<{ post: PostPayload; first: boolean; permalink: string }>()
const { t } = useI18n()
</script>

<template>
  <header class="mb-1.5 flex min-w-0 items-start justify-between gap-2">
    <div class="min-w-0">
      <div class="flex min-w-0 items-center gap-2">
        <a :id="`post-author-${post.id}`" rel="author" :href="`/u/${post.author.id}`" class="min-w-0 truncate font-semibold text-base-content hover:text-primary">{{ post.author.username }}</a>
        <span v-if="first" class="rounded bg-base-200 px-1.5 py-0.5 text-xs font-semibold text-base-content/55">{{ t('topic.originalPost') }}</span>
        <a v-if="post.postNo" :href="permalink" class="hidden shrink-0 text-xs font-semibold tabular-nums text-base-content/55 hover:text-primary sm:inline">#{{ formatNumber(post.postNo) }}</a>
      </div>
      <div class="mt-0.5 flex items-center gap-2 text-xs text-base-content/55 sm:hidden">
        <a v-if="post.postNo" :href="permalink" class="font-semibold tabular-nums text-base-content/55 hover:text-primary">#{{ formatNumber(post.postNo) }}</a>
        <time :datetime="post.createdAt" class="truncate">{{ formatDateTime(post.createdAt) }}</time>
      </div>
    </div>
    <div class="flex shrink-0 items-center gap-0.5 sm:gap-1.5">
      <slot />
      <time :datetime="post.createdAt" class="hidden w-36 shrink-0 text-right text-xs text-base-content/55 sm:-ml-1 sm:block">{{ formatDateTime(post.createdAt) }}</time>
    </div>
  </header>
</template>
