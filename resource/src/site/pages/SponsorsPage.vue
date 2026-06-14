<script setup lang="ts">
import { computed } from 'vue'
import { ExternalLink, HeartHandshake, Mail, ShieldCheck } from '@lucide/vue'
import EmptyState from '@/site/components/EmptyState.vue'
import PageHeader from '@/site/components/PageHeader.vue'
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
  if (section.tone === 'diamond') return 'p-3.5 bg-base-100'
  if (section.tone === 'gold') return 'p-3 bg-base-100'
  return 'p-2.5 bg-base-200/45 sm:bg-base-100'
}

function sponsorBodyClass(section: SponsorSectionPayload) {
  return 'flex items-start gap-2.5'
}

function sponsorAvatarClass(section: SponsorSectionPayload) {
  if (section.tone === 'diamond') return 'h-11 w-11'
  if (section.tone === 'gold') return 'h-10 w-10'
  return 'h-8 w-8'
}

function showMessage(sponsor: SponsorPayload) {
  return sponsor.message || t('sponsors.defaultMessage')
}
</script>

<template>
    <div class="pb-12">
      <PageHeader :title="props.content.title" :description="props.content.description">
        <template #badge>
          <span class="gf-badge gf-badge-muted">{{ props.totalCount }}</span>
        </template>
      </PageHeader>

      <div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_260px]">
        <div class="space-y-6">
          <section v-for="section in props.sections" :key="section.key" class="space-y-2.5">
            <div class="flex items-center justify-between gap-3">
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
                class="group relative rounded-[var(--gf-radius-box)] border border-line/70 transition hover:border-primary/25 hover:bg-info/10"
                :class="sponsorCardClass(section)"
              >
                <div :class="sponsorBodyClass(section)">
                  <img
                    :src="sponsor.avatarUrl"
                    :alt="sponsor.name"
                    class="shrink-0 rounded-[var(--gf-radius-field)] border border-line object-cover"
                    :class="sponsorAvatarClass(section)"
                    loading="lazy"
                  />
                  <div class="min-w-0 flex-1">
                    <div class="flex min-w-0 items-center gap-1.5">
                      <h3 class="truncate text-[13px] font-semibold text-base-content group-hover:text-primary sm:text-sm">{{ sponsor.name }}</h3>
                      <ExternalLink v-if="sponsor.link" class="h-3 w-3 shrink-0 text-base-content/35 group-hover:text-primary sm:h-3.5 sm:w-3.5" />
                    </div>
                    <p class="mt-0.5 line-clamp-2 text-[11px] leading-4 text-base-content/55 sm:mt-1 sm:text-xs sm:leading-5">{{ showMessage(sponsor) }}</p>
                  </div>
                </div>
              </a>
            </div>
          </section>

          <EmptyState v-if="!hasSponsors" class="gf-panel" :icon="HeartHandshake" :title="t('sponsors.emptyTitle')" :description="t('sponsors.emptyDescription')" />
        </div>

        <aside class="space-y-3">
          <div class="rounded-[var(--gf-radius-box)] border border-line/70 bg-base-200/45 p-4 sm:bg-base-100">
            <h2 class="text-sm font-semibold text-base-content">{{ props.contact.title }}</h2>
            <p class="mt-2 text-sm leading-6 text-base-content/55">{{ props.contact.description }}</p>
            <a :href="props.contact.buttonLink" class="gf-button gf-button-md gf-button-primary mt-4">
              <Mail class="h-4 w-4" />
              {{ props.contact.buttonText }}
            </a>
          </div>

          <div v-if="props.rules.length" class="rounded-[var(--gf-radius-box)] border border-line/70 bg-base-200/45 p-4 sm:bg-base-100">
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
