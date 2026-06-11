<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ExternalLink, Link, Send, ShieldCheck } from '@lucide/vue'
import type { LayoutPayload, LinksPageProps } from '@/types/payload'

defineProps<{
  layout: LayoutPayload
  props: LinksPageProps
}>()
const { t } = useI18n()
</script>

<template>
    <div class="pb-12">
      <header class="mb-3 border-b border-line/70 pb-4">
        <div class="flex flex-wrap items-center gap-2">
          <h1 class="text-2xl font-bold text-base-content">{{ t('linksPage.title') }}</h1>
          <span class="gf-badge gf-badge-muted">{{ props.totalCount }}</span>
        </div>
        <p class="mt-2 max-w-2xl text-sm leading-6 text-base-content/55">{{ t('linksPage.subtitle') }}</p>
      </header>

      <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_260px]">
        <div class="space-y-5">
          <section v-for="group in props.groups" :key="group.name" class="space-y-3">
            <div class="flex items-center justify-between gap-3 border-b border-line pb-2">
              <h2 class="flex min-w-0 items-center gap-2 text-base font-bold text-base-content">
                <span
                  class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-base-200 text-sm"
                  :style="{ color: group.color || 'var(--gf-color-base-content)' }"
                >
                  {{ group.emoji || '↗' }}
                </span>
                <span class="truncate">{{ group.name }}</span>
              </h2>
              <span class="gf-badge gf-badge-muted text-[11px]">{{ group.links.length }}</span>
            </div>

            <div class="grid grid-cols-2 gap-2 md:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-5">
              <a
                v-for="link in group.links"
                :key="`${group.name}-${link.url}`"
                :href="link.url"
                target="_blank"
                rel="noopener noreferrer"
                class="gf-panel group px-2.5 py-2 transition hover:border-primary/20 hover:bg-info/10"
              >
                <div class="flex items-center gap-2">
                  <div class="flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-md border border-line bg-base-200">
                    <img
                      v-if="link.logoUrl"
                      :src="link.logoUrl"
                      :alt="link.name"
                      class="h-full w-full object-cover"
                      loading="lazy"
                    />
                    <Link v-else class="h-4 w-4 text-base-content/55" />
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="flex min-w-0 items-center gap-2">
                      <h3 class="truncate text-[13px] font-semibold text-base-content group-hover:text-primary">{{ link.name }}</h3>
                      <ExternalLink class="h-3 w-3 shrink-0 text-base-content/35 group-hover:text-primary" />
                    </div>
                    <p class="mt-0.5 truncate text-[11px] leading-4 text-base-content/55">{{ link.desc || link.url }}</p>
                  </div>
                </div>
              </a>
            </div>
          </section>

          <div v-if="!props.groups.length" class="gf-panel px-5 py-16 text-center">
            <Link class="mx-auto h-8 w-8 text-base-content/35" />
            <h2 class="mt-3 text-base font-semibold text-base-content">{{ t('linksPage.emptyTitle') }}</h2>
            <p class="mt-1 text-sm text-base-content/55">{{ t('linksPage.emptyDescription') }}</p>
          </div>
        </div>

        <aside class="space-y-3">
          <div class="gf-panel p-4">
            <h2 class="text-sm font-semibold text-base-content">{{ t('linksPage.applyTitle') }}</h2>
            <p class="mt-2 text-sm leading-6 text-base-content/55">{{ t('linksPage.applyDescription') }}</p>
            <a href="/publish" class="gf-button gf-button-md gf-button-primary mt-4">
              <Send class="h-4 w-4" />
              {{ t('linksPage.applyAction') }}
            </a>
          </div>

          <div class="gf-panel p-4">
            <h2 class="text-sm font-semibold text-base-content">{{ t('linksPage.principlesTitle') }}</h2>
            <div class="mt-3 space-y-2 text-sm text-base-content/75">
              <div class="flex gap-2">
                <ShieldCheck class="mt-0.5 h-4 w-4 shrink-0 text-success" />
                <span>{{ t('linksPage.principles.healthy') }}</span>
              </div>
              <div class="flex gap-2">
                <ShieldCheck class="mt-0.5 h-4 w-4 shrink-0 text-success" />
                <span>{{ t('linksPage.principles.relevant') }}</span>
              </div>
              <div class="flex gap-2">
                <ShieldCheck class="mt-0.5 h-4 w-4 shrink-0 text-success" />
                <span>{{ t('linksPage.principles.stable') }}</span>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </div>
</template>
