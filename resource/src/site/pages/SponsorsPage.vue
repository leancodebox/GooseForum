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
  if (section.tone === 'diamond') return 'bg-info/10 text-primary'
  if (section.tone === 'gold') return 'bg-warning/10 text-warning'
  if (section.tone === 'silver') return 'bg-base-300 text-base-content/75'
  return 'bg-error/10 text-error'
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
      <header class="mb-4 border-b border-line/70 pb-4">
        <div class="flex flex-wrap items-center gap-2">
          <h1 class="text-2xl font-bold text-base-content">{{ props.content.title }}</h1>
          <span class="gf-badge gf-badge-muted">{{ props.totalCount }}</span>
        </div>
        <p class="mt-2 max-w-2xl text-sm leading-6 text-base-content/55">{{ props.content.description }}</p>
      </header>

      <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_260px]">
        <div class="space-y-5">
          <section v-for="section in props.sections" :key="section.key" class="space-y-3">
            <div class="flex items-center justify-between gap-3 border-b border-line pb-2">
              <h2 class="flex min-w-0 items-center gap-2 text-base font-bold text-base-content">
                <span class="rounded px-1.5 py-0.5 text-[11px] font-semibold" :class="sectionBadgeClass(section)">{{ section.label }}</span>
              </h2>
              <span class="gf-badge gf-badge-muted text-[11px]">{{ section.sponsors.length }}</span>
            </div>

            <div class="grid gap-3" :class="sectionGrid(section)">
              <a
                v-for="sponsor in section.sponsors"
                :key="`${section.key}-${sponsor.name}`"
                :href="sponsor.link || '#'"
                :target="sponsor.link ? '_blank' : undefined"
                :rel="sponsor.link ? 'noopener noreferrer' : undefined"
                class="gf-panel group relative transition hover:border-primary/20 hover:bg-info/10"
                :class="sponsorCardClass(section)"
              >
                <div class="flex items-start gap-3">
                  <img
                    :src="sponsor.avatarUrl"
                    :alt="sponsor.name"
                    class="h-11 w-11 shrink-0 rounded-md border border-line object-cover"
                    loading="lazy"
                  />
                  <div class="min-w-0 flex-1">
                    <div class="flex min-w-0 items-center gap-2">
                      <h3 class="truncate text-sm font-semibold text-base-content group-hover:text-primary">{{ sponsor.name }}</h3>
                      <ExternalLink v-if="sponsor.link" class="h-3.5 w-3.5 shrink-0 text-base-content/35 group-hover:text-primary" />
                    </div>
                    <p class="mt-1 line-clamp-2 text-xs leading-5 text-base-content/55">{{ showMessage(sponsor) }}</p>
                  </div>
                </div>
              </a>
            </div>
          </section>

          <div v-if="!hasSponsors" class="gf-panel px-5 py-16 text-center">
            <HeartHandshake class="mx-auto h-8 w-8 text-base-content/35" />
            <h2 class="mt-3 text-base font-semibold text-base-content">{{ t('sponsors.emptyTitle') }}</h2>
            <p class="mt-1 text-sm text-base-content/55">{{ t('sponsors.emptyDescription') }}</p>
          </div>
        </div>

        <aside class="space-y-3">
          <div class="gf-panel p-4">
            <h2 class="text-sm font-semibold text-base-content">{{ props.contact.title }}</h2>
            <p class="mt-2 text-sm leading-6 text-base-content/55">{{ props.contact.description }}</p>
            <a :href="props.contact.buttonLink" class="gf-button gf-button-md gf-button-primary mt-4">
              <Mail class="h-4 w-4" />
              {{ props.contact.buttonText }}
            </a>
          </div>

          <div v-if="props.rules.length" class="gf-panel p-4">
            <h2 class="text-sm font-semibold text-base-content">{{ t('sponsors.rulesTitle') }}</h2>
            <div class="mt-3 space-y-2 text-sm text-base-content/75">
              <div v-for="rule in props.rules" :key="rule.content" class="flex gap-2">
                <ShieldCheck class="mt-0.5 h-4 w-4 shrink-0 text-success" />
                <span>{{ rule.content }}</span>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </div>
</template>
