<script setup lang="ts">
import { FileText, PenSquare, ShieldAlert } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { formatDateTime } from '@/runtime/format'
import type { DraftPayload, DraftsPageProps, LayoutPayload } from '@/types/payload'

defineProps<{
  layout: LayoutPayload
  props: DraftsPageProps
}>()

const { t } = useI18n()

function isBlocked(draft: DraftPayload) {
  return draft.processStatus === 1
}
</script>

<template>
  <main class="min-w-0 pb-8">
    <header class="mb-3 flex flex-col gap-2 border-b border-gray-200/70 pb-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="min-w-0">
        <div class="flex min-w-0 items-center gap-2">
          <h1 class="text-xl font-bold text-gray-950">{{ t('drafts.title') }}</h1>
          <span class="inline-flex h-5 items-center rounded-full bg-gray-100 px-2 text-xs font-semibold tabular-nums text-gray-700">
            {{ t('drafts.total', { count: props.total }) }}
          </span>
        </div>
        <p class="mt-0.5 text-xs text-gray-500">{{ t('drafts.summary') }}</p>
      </div>
      <a
        href="/publish"
        class="inline-flex h-8 w-fit items-center gap-1.5 rounded-md border border-gray-200 bg-white px-2.5 text-xs font-semibold text-gray-600 hover:bg-gray-50 hover:text-gray-900"
      >
        <PenSquare class="h-4 w-4" />
        {{ t('drafts.newDraft') }}
      </a>
    </header>

    <section class="overflow-hidden rounded-lg border border-gray-200/70 bg-white shadow-[0_2px_8px_rgba(0,0,0,0.02)]">
      <div class="hidden grid-cols-[minmax(0,1fr)_152px_132px] gap-4 border-b border-gray-100 bg-gray-50/60 px-4 py-2 text-[11px] font-bold uppercase text-gray-600 md:grid">
        <div>{{ t('drafts.table.draft') }}</div>
        <div class="text-right">{{ t('drafts.table.updatedAt') }}</div>
        <div class="text-right">{{ t('drafts.table.action') }}</div>
      </div>

      <div v-if="props.drafts.length" class="divide-y divide-gray-100">
        <article
          v-for="draft in props.drafts"
          :key="draft.id"
          class="grid gap-3 px-4 py-3 transition hover:bg-gray-50/70 md:grid-cols-[minmax(0,1fr)_152px_132px] md:items-center md:gap-4"
        >
          <div class="min-w-0">
            <div class="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
              <a :href="draft.editUrl" class="min-w-0 max-w-full truncate text-base font-semibold text-gray-950 hover:text-blue-600">
                {{ draft.title || t('drafts.untitled') }}
              </a>
              <span
                v-for="category in draft.categories"
                :key="category.id"
                class="inline-flex shrink-0 items-center gap-1 rounded-full border border-gray-200 px-2 py-0.5 text-[11px] font-medium text-gray-600"
              >
                <span class="h-1.5 w-1.5 rounded-full" :style="{ backgroundColor: category.color }" />
                {{ category.name }}
              </span>
              <span
                v-if="isBlocked(draft)"
                class="inline-flex shrink-0 items-center gap-1 rounded-full bg-red-50 px-2 py-0.5 text-[11px] font-semibold text-red-600"
              >
                <ShieldAlert class="h-3.5 w-3.5" />
                {{ t('drafts.blocked') }}
              </span>
            </div>
            <p class="mt-1.5 line-clamp-1 text-sm leading-6 text-gray-500">
              {{ draft.description || t('drafts.emptyDescription') }}
            </p>
            <div class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-400">
              <span>{{ t('drafts.meta.createdAt') }} {{ formatDateTime(draft.createdAt) }}</span>
              <span>{{ t('drafts.meta.views') }} {{ draft.viewCount }}</span>
              <span>{{ t('drafts.meta.replies') }} {{ draft.replyCount }}</span>
            </div>
          </div>
          <time class="text-xs font-medium text-gray-400 md:text-right">{{ formatDateTime(draft.updatedAt) }}</time>
          <div class="flex flex-wrap items-center gap-2 md:flex-nowrap md:justify-end">
            <a
              :href="draft.editUrl"
              class="inline-flex h-8 shrink-0 items-center gap-1 rounded-md bg-blue-600 px-2.5 text-xs font-semibold text-white hover:bg-blue-700"
            >
              <FileText class="h-4 w-4" />
              {{ t('drafts.edit') }}
            </a>
          </div>
        </article>
      </div>

      <div v-else class="flex min-h-56 flex-col items-center justify-center px-6 text-center">
        <FileText class="h-8 w-8 text-gray-300" />
        <h2 class="mt-2 text-base font-semibold text-gray-950">{{ t('drafts.emptyTitle') }}</h2>
        <p class="mt-1 text-sm text-gray-500">{{ t('drafts.emptyHint') }}</p>
        <a href="/publish" class="mt-4 inline-flex h-9 items-center rounded-md bg-blue-600 px-4 text-sm font-semibold text-white hover:bg-blue-700">
          {{ t('drafts.newDraft') }}
        </a>
      </div>
    </section>
  </main>
</template>
