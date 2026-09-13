<script setup lang="ts">
import { FileText, PenSquare, ShieldAlert } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { formatDateTime } from '@/runtime/format'
import EmptyState from '@/site/components/EmptyState.vue'
import PageHeader from '@/site/components/PageHeader.vue'
import type { DraftPayload, DraftsPageProps, LayoutPayload } from '@gooseforum/client'

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
    <PageHeader :title="t('drafts.title')" :description="t('drafts.summary')" compact>
      <template #badge>
        <Badge variant="muted" class="h-5 tabular-nums text-base-content/75">
          {{ t('drafts.total', { count: props.total }) }}
        </Badge>
      </template>
      <template #actions>
        <Button as-child variant="surface" size="sm" class="w-fit text-xs">
          <a href="/publish">
            <PenSquare class="h-4 w-4" />
            {{ t('drafts.newDraft') }}
          </a>
        </Button>
      </template>
    </PageHeader>

    <section class="gf-card overflow-hidden">
      <div class="hidden grid-cols-[minmax(0,1fr)_152px_132px] gap-4 border-b border-line bg-base-200/60 px-4 py-2 text-[11px] font-bold uppercase text-base-content/75 md:grid">
        <div>{{ t('drafts.table.draft') }}</div>
        <div class="text-right">{{ t('drafts.table.updatedAt') }}</div>
        <div class="text-right">{{ t('drafts.table.action') }}</div>
      </div>

      <div v-if="props.drafts.length" class="divide-y divide-line">
        <article
          v-for="draft in props.drafts"
          :key="draft.id"
          class="grid gap-3 px-4 py-3 transition hover:bg-base-200/70 md:grid-cols-[minmax(0,1fr)_152px_132px] md:items-center md:gap-4"
        >
          <div class="min-w-0">
            <div class="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
              <a :href="draft.editUrl" class="min-w-0 max-w-full truncate text-base font-semibold text-base-content hover:text-primary">
                {{ draft.title || t('drafts.untitled') }}
              </a>
              <span
                v-for="category in draft.categories"
                :key="category.id"
                class="inline-flex shrink-0 items-center gap-1 rounded-full border border-line px-2 py-0.5 text-[11px] font-medium text-base-content/75"
              >
                <span class="h-1.5 w-1.5 rounded-full" :style="{ backgroundColor: category.color }" />
                {{ category.name }}
              </span>
              <Badge
                v-if="isBlocked(draft)"
                variant="error"
                class="shrink-0 text-[11px]"
              >
                <ShieldAlert class="h-3.5 w-3.5" />
                {{ t('drafts.blocked') }}
              </Badge>
            </div>
            <p class="mt-1.5 line-clamp-1 text-sm leading-6 text-base-content/55">
              {{ draft.description || t('drafts.emptyDescription') }}
            </p>
            <div class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-base-content/55">
              <span>{{ t('drafts.meta.createdAt') }} {{ formatDateTime(draft.createdAt) }}</span>
              <span>{{ t('drafts.meta.views') }} {{ draft.viewCount }}</span>
              <span>{{ t('drafts.meta.replies') }} {{ draft.replyCount }}</span>
            </div>
          </div>
          <time class="text-xs font-medium text-base-content/55 md:text-right">{{ formatDateTime(draft.updatedAt) }}</time>
          <div class="flex flex-wrap items-center gap-2 md:flex-nowrap md:justify-end">
            <Button as-child variant="brand" size="sm" class="shrink-0 text-xs">
              <a :href="draft.editUrl">
                <FileText class="h-4 w-4" />
                {{ t('drafts.edit') }}
              </a>
            </Button>
          </div>
        </article>
      </div>

      <EmptyState v-else :icon="FileText" :title="t('drafts.emptyTitle')" :description="t('drafts.emptyHint')">
        <Button as-child variant="brand" class="px-4">
          <a href="/publish">{{ t('drafts.newDraft') }}</a>
        </Button>
      </EmptyState>
    </section>
  </main>
</template>
