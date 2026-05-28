<script setup lang="ts">
import { computed } from 'vue'
import { ExternalLink, HeartHandshake, Mail, ShieldCheck } from '@lucide/vue'
import type { LayoutPayload, SponsorPayload, SponsorSectionPayload, SponsorsPageProps } from '@/types/payload'
import { useI18n } from 'vue-i18n'

const page = defineProps<{
  layout: LayoutPayload
  props: SponsorsPageProps
}>()

const { t } = useI18n()
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
  return sponsor.message || t('sponsors.defaultMessage')
}
</script>

<template>
    <div class="pb-12">
      <header class="mb-4 border-b border-gray-200/70 pb-4">
        <div class="flex flex-wrap items-center gap-2">
          <h1 class="text-2xl font-bold text-gray-950">{{ props.content.title }}</h1>
          <span class="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-semibold text-gray-500">{{ props.totalCount }}</span>
        </div>
        <p class="mt-2 max-w-2xl text-sm leading-6 text-gray-500">{{ props.content.description }}</p>
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
            <h2 class="mt-3 text-base font-semibold text-gray-900">{{ t('sponsors.emptyTitle') }}</h2>
            <p class="mt-1 text-sm text-gray-500">{{ t('sponsors.emptyDescription') }}</p>
          </div>
        </div>

        <aside class="space-y-3">
          <div class="rounded-lg border border-gray-200/70 bg-white p-4">
            <h2 class="text-sm font-semibold text-gray-950">{{ props.contact.title }}</h2>
            <p class="mt-2 text-sm leading-6 text-gray-500">{{ props.contact.description }}</p>
            <a :href="props.contact.buttonLink" class="mt-4 inline-flex h-9 items-center gap-1.5 rounded-md bg-blue-600 px-3 text-sm font-semibold text-white hover:bg-blue-700">
              <Mail class="h-4 w-4" />
              {{ props.contact.buttonText }}
            </a>
          </div>

          <div v-if="props.rules.length" class="rounded-lg border border-gray-200/70 bg-white p-4">
            <h2 class="text-sm font-semibold text-gray-950">{{ t('sponsors.rulesTitle') }}</h2>
            <div class="mt-3 space-y-2 text-sm text-gray-600">
              <div v-for="rule in props.rules" :key="rule.content" class="flex gap-2">
                <ShieldCheck class="mt-0.5 h-4 w-4 shrink-0 text-emerald-600" />
                <span>{{ rule.content }}</span>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </div>
</template>
