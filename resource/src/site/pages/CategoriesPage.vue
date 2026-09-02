<script setup lang="ts">
import { ChevronRight, LayoutGrid, MessageCircle } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import EmptyState from '@/site/components/EmptyState.vue'
import PageHeader from '@/site/components/PageHeader.vue'
import { formatNumber } from '@/runtime/format'
import type { CategoriesPageProps, LayoutPayload } from '@gooseforum/client'

defineProps<{
  layout: LayoutPayload
  props: CategoriesPageProps
}>()

const { t } = useI18n()

function isImageIcon(icon: string) {
  return /^(https?:\/\/|\/)/.test(icon)
}
</script>

<template>
  <div class="pb-12">
    <PageHeader :title="t('categoriesPage.title')" :description="t('categoriesPage.subtitle')" compact>
      <template #badge>
        <span class="gf-badge gf-badge-muted">{{ t('categoriesPage.total', { count: props.total }) }}</span>
      </template>
    </PageHeader>

    <section v-if="props.categories.length" class="grid gap-3 sm:grid-cols-2">
      <a
        v-for="category in props.categories"
        :key="category.id"
        :href="category.url"
        class="group gf-card flex min-w-0 items-stretch p-3.5 transition hover:border-primary/25 hover:bg-base-200"
      >
        <span
          class="mr-3 w-1 shrink-0 rounded-full"
          :style="{ backgroundColor: category.color || 'var(--gf-color-primary)' }"
          aria-hidden="true"
        />

        <div class="min-w-0 flex-1 py-0.5">
          <div class="flex min-w-0 items-center gap-2">
            <span v-if="category.icon" class="inline-flex shrink-0 items-center justify-center text-base leading-none" aria-hidden="true">
              <img v-if="isImageIcon(category.icon)" :src="category.icon" alt="" class="h-5 w-5 rounded object-cover">
              <span v-else>{{ category.icon }}</span>
            </span>
            <h2 class="truncate text-[15px] font-bold text-base-content transition group-hover:text-primary">{{ category.name }}</h2>
          </div>
          <p class="mt-1 line-clamp-2 text-xs leading-5 text-base-content/55">
            {{ category.description || t('categoriesPage.noDescription') }}
          </p>
        </div>

        <div class="ml-3 flex shrink-0 items-center gap-2 self-center">
          <span class="gf-badge gf-badge-muted inline-flex gap-1 text-[10px] font-medium">
            <MessageCircle class="h-3 w-3" aria-hidden="true" />
            {{ t('categoriesPage.topicCount', { count: formatNumber(category.topicCount) }) }}
          </span>
          <ChevronRight class="h-4 w-4 text-icon-muted transition group-hover:translate-x-0.5 group-hover:text-primary" aria-hidden="true" />
        </div>
      </a>
    </section>

    <EmptyState
      v-else
      class="gf-panel"
      :icon="LayoutGrid"
      :title="t('categoriesPage.emptyTitle')"
      :description="t('categoriesPage.emptyDescription')"
    />
  </div>
</template>
