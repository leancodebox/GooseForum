<script setup lang="ts">
import { CalendarDays, ChevronRight, FileText, MessageCircle, Sparkles, UsersRound } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import EmptyState from '@/site/components/EmptyState.vue'
import PageHeader from '@/site/components/PageHeader.vue'
import UserAvatar from '@/site/components/UserAvatar.vue'
import { formatNumber } from '@/runtime/format'
import type { LayoutPayload, MembersPageProps } from '@gooseforum/client'

defineProps<{
  layout: LayoutPayload
  props: MembersPageProps
}>()

const { t } = useI18n()
</script>

<template>
  <div class="pb-12">
    <PageHeader :title="t('membersPage.title')" :description="t('membersPage.subtitle')" compact />

    <section v-if="props.members.length" class="grid gap-2.5 md:grid-cols-2 xl:grid-cols-3">
      <a
        v-for="member in props.members"
        :key="member.id"
        :href="member.url"
        class="group gf-card flex min-w-0 flex-col overflow-hidden transition hover:-translate-y-px hover:border-primary/30 hover:shadow-[0_5px_16px_rgb(0_0_0/calc(var(--gf-depth)*0.06))]"
      >
        <div class="flex min-w-0 items-center gap-2.5 px-3.5 pt-3.5">
          <span class="h-11 w-11 shrink-0 overflow-hidden rounded-full bg-base-200 ring-1 ring-line/70 transition group-hover:ring-primary/30">
            <UserAvatar :src="member.avatarUrl" :alt="member.nickname" class="h-full w-full" img-class="h-full w-full object-cover" />
          </span>
          <div class="min-w-0 flex-1">
            <div class="flex min-w-0 items-baseline gap-1.5">
              <h2 class="truncate text-sm font-bold text-base-content transition group-hover:text-primary">{{ member.nickname }}</h2>
              <span class="truncate text-[11px] text-base-content/45">@{{ member.username }}</span>
            </div>
            <span class="mt-0.5 inline-flex items-center gap-1 text-[10px] text-base-content/45">
              <CalendarDays class="h-3 w-3" aria-hidden="true" />{{ t('membersPage.joinedAt', { date: member.joinedAt }) }}
            </span>
          </div>
          <ChevronRight class="h-4 w-4 shrink-0 text-icon-muted transition group-hover:translate-x-0.5 group-hover:text-primary" aria-hidden="true" />
        </div>

        <p class="mx-3.5 mb-3 mt-2 line-clamp-2 min-h-10 text-xs leading-5 text-base-content/55">
          {{ member.bio || t('membersPage.noBio') }}
        </p>

        <div class="mt-auto grid grid-cols-3 divide-x divide-line/60 border-t border-line/60 bg-base-200/35 py-2">
          <span class="flex min-w-0 items-center justify-center gap-1 text-[10px] text-base-content/50" :title="t('membersPage.prestige')">
            <Sparkles class="h-3 w-3 shrink-0 text-warning" aria-hidden="true" />
            <strong class="truncate font-semibold text-base-content/75">{{ formatNumber(member.prestige) }}</strong>
          </span>
          <span class="flex min-w-0 items-center justify-center gap-1 text-[10px] text-base-content/50" :title="t('membersPage.topics')">
            <FileText class="h-3 w-3 shrink-0" aria-hidden="true" />
            <strong class="truncate font-semibold text-base-content/75">{{ formatNumber(member.topicCount) }}</strong>
          </span>
          <span class="flex min-w-0 items-center justify-center gap-1 text-[10px] text-base-content/50" :title="t('membersPage.replies')">
            <MessageCircle class="h-3 w-3 shrink-0" aria-hidden="true" />
            <strong class="truncate font-semibold text-base-content/75">{{ formatNumber(member.replyCount) }}</strong>
          </span>
        </div>
      </a>
    </section>

    <EmptyState
      v-else
      class="gf-panel"
      :icon="UsersRound"
      :title="t('membersPage.emptyTitle')"
      :description="t('membersPage.emptyDescription')"
    />

    <nav
      v-if="props.previousUrl || props.pagination.hasNext"
      class="mt-4 flex items-center justify-between gap-3 border-t border-line/70 pt-4"
      :aria-label="t('membersPage.pagination')"
    >
      <a
        v-if="props.previousUrl"
        :href="props.previousUrl"
        rel="prev"
        class="gf-button gf-button-sm gf-button-secondary"
      >{{ t('common.previousPage') }}</a>
      <span v-else />
      <a
        v-if="props.pagination.hasNext"
        :href="props.pagination.nextUrl"
        rel="next"
        class="gf-button gf-button-sm gf-button-secondary"
      >{{ t('common.nextPage') }}</a>
      <span v-else />
    </nav>
  </div>
</template>
