<script setup lang="ts">
import { computed } from 'vue'
import { ExternalLink, HeartHandshake, Mail, ShieldCheck } from '@lucide/vue'
import type { LayoutPayload, SponsorPayload, SponsorSectionPayload, SponsorsPageProps } from '../types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: SponsorsPageProps
}>()

const hasSponsors = computed(() => page.props.sections.length > 0)

function sectionGrid(section: SponsorSectionPayload) {
  if (section.tone === 'diamond') return 'grid-cols-1 md:grid-cols-2'
  if (section.tone === 'gold') return 'grid-cols-1 sm:grid-cols-2 xl:grid-cols-3'
  return 'grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-5'
}

function sectionBadgeClass(section: SponsorSectionPayload) {
  if (section.tone === 'diamond') return 'bg-blue-50 text-blue-700'
  if (section.tone === 'gold') return 'bg-amber-50 text-amber-700'
  if (section.tone === 'silver') return 'bg-gray-100 text-gray-600'
  return 'bg-rose-50 text-rose-700'
}

function sponsorCardClass(section: SponsorSectionPayload) {
  if (section.tone === 'diamond') return 'p-4'
  if (section.tone === 'gold') return 'p-3'
  return 'p-2.5'
}

function showMessage(sponsor: SponsorPayload) {
  return sponsor.message || '感谢支持 GooseForum。'
}
</script>

<template>
    <div class="pb-12">
      <header class="mb-4 border-b border-gray-200/70 pb-4">
        <div class="flex flex-wrap items-center gap-2">
          <h1 class="text-2xl font-bold text-gray-950">赞助</h1>
          <span class="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-semibold text-gray-500">{{ props.totalCount }}</span>
        </div>
        <p class="mt-2 max-w-2xl text-sm leading-6 text-gray-500">感谢这些赞助者帮助 GooseForum 持续变好。</p>
      </header>

      <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_260px]">
        <div class="space-y-5">
          <section v-for="section in props.sections" :key="section.key" class="space-y-3">
            <div class="flex items-center justify-between gap-3 border-b border-gray-100 pb-2">
              <h2 class="flex min-w-0 items-center gap-2 text-base font-bold text-gray-950">
                <span class="rounded px-1.5 py-0.5 text-[11px] font-semibold" :class="sectionBadgeClass(section)">{{ section.label }}</span>
              </h2>
              <span class="rounded-full bg-gray-100 px-2 py-0.5 text-[11px] font-semibold text-gray-500">{{ section.sponsors.length }}</span>
            </div>

            <div class="grid gap-3" :class="sectionGrid(section)">
              <a
                v-for="sponsor in section.sponsors"
                :key="`${section.key}-${sponsor.name}`"
                :href="sponsor.link || '#'"
                :target="sponsor.link ? '_blank' : undefined"
                :rel="sponsor.link ? 'noopener noreferrer' : undefined"
                class="group relative rounded-lg border border-gray-200 bg-white transition hover:border-blue-200 hover:bg-blue-50/20"
                :class="sponsorCardClass(section)"
              >
                <div class="flex items-start gap-3">
                  <img
                    :src="sponsor.avatarUrl"
                    :alt="sponsor.name"
                    class="h-11 w-11 shrink-0 rounded-md border border-gray-100 object-cover"
                    loading="lazy"
                  />
                  <div class="min-w-0 flex-1">
                    <div class="flex min-w-0 items-center gap-2">
                      <h3 class="truncate text-sm font-semibold text-gray-950 group-hover:text-blue-600">{{ sponsor.name }}</h3>
                      <ExternalLink v-if="sponsor.link" class="h-3.5 w-3.5 shrink-0 text-gray-300 group-hover:text-blue-500" />
                    </div>
                    <p class="mt-1 line-clamp-2 text-xs leading-5 text-gray-500">{{ showMessage(sponsor) }}</p>
                  </div>
                </div>
              </a>
            </div>
          </section>

          <div v-if="!hasSponsors" class="rounded-lg border border-gray-200 bg-white px-5 py-16 text-center">
            <HeartHandshake class="mx-auto h-8 w-8 text-gray-300" />
            <h2 class="mt-3 text-base font-semibold text-gray-900">暂无赞助者</h2>
            <p class="mt-1 text-sm text-gray-500">站点还没有配置赞助信息。</p>
          </div>
        </div>

        <aside class="space-y-3">
          <div class="rounded-lg border border-gray-200/70 bg-white p-4">
            <h2 class="text-sm font-semibold text-gray-950">成为赞助者</h2>
            <p class="mt-2 text-sm leading-6 text-gray-500">支持社区建设，赞助者可展示在赞助页，并获得更醒目的社区露出。</p>
            <a href="mailto:contact@gooseforum.online" class="mt-4 inline-flex h-9 items-center gap-1.5 rounded-md bg-blue-600 px-3 text-sm font-semibold text-white hover:bg-blue-700">
              <Mail class="h-4 w-4" />
              联系我们
            </a>
          </div>

          <div class="rounded-lg border border-gray-200/70 bg-white p-4">
            <h2 class="text-sm font-semibold text-gray-950">展示规则</h2>
            <div class="mt-3 space-y-2 text-sm text-gray-600">
              <div class="flex gap-2">
                <ShieldCheck class="mt-0.5 h-4 w-4 shrink-0 text-emerald-600" />
                <span>链接需稳定可访问。</span>
              </div>
              <div class="flex gap-2">
                <ShieldCheck class="mt-0.5 h-4 w-4 shrink-0 text-emerald-600" />
                <span>内容需适合公开社区展示。</span>
              </div>
              <div class="flex gap-2">
                <ShieldCheck class="mt-0.5 h-4 w-4 shrink-0 text-emerald-600" />
                <span>头像或 Logo 建议保持清晰。</span>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </div>
</template>
