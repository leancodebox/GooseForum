<script setup lang="ts">
import { ref, watch } from 'vue'
import { Ban, CircleAlert, Flag, History, RotateCcw, Scale, XCircle } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { fetchModerationLogs, fetchModerationReports, updateModerationArticleStatus, updateModerationReplyStatus, updateModerationReportStatus } from '@/runtime/api'
import { formatDateTime } from '@/runtime/format'
import { fetchPage } from '@/runtime/router'
import { showUserCard } from '@/runtime/user-card-events'
import EmptyState from '@/site/components/EmptyState.vue'
import PageHeader from '@/site/components/PageHeader.vue'
import TopicList from '@/site/components/TopicList.vue'
import UserAvatar from '@/site/components/UserAvatar.vue'
import type { LayoutPayload, ModerationLogItem, ModerationPageProps, ModerationReportItem, PagePayload, TopicPayload } from '@/types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: ModerationPageProps
}>()

const { t, te } = useI18n()
const currentProps = ref<ModerationPageProps>(page.props)
const topics = ref<TopicPayload[]>([...page.props.topics])
const busyIds = ref<number[]>([])
const actionError = ref('')
const loadingList = ref(false)
const activeConsoleTab = ref<'reports' | 'ban' | 'logs' | 'guidance'>('reports')
const reportItems = ref<ModerationReportItem[]>([])
const reportStatus = ref<'open' | 'closed'>('open')
const reportNextCursor = ref(0)
const reportHasNext = ref(true)
const reportLoading = ref(false)
const reportLoaded = ref(false)
const reportError = ref('')
const reportBusyIds = ref<number[]>([])
const logItems = ref<ModerationLogItem[]>([])
const logNextCursor = ref(0)
const logHasNext = ref(true)
const logLoading = ref(false)
const logLoaded = ref(false)
const logError = ref('')

const managementTabs = [
  { key: 'reports', icon: Flag },
  { key: 'ban', icon: Ban },
  { key: 'logs', icon: History },
  { key: 'guidance', icon: Scale },
]

watch(
  () => page.props,
  (next) => {
    currentProps.value = next
    topics.value = [...next.topics]
    actionError.value = ''
    busyIds.value = []
    loadingList.value = false
  },
  { immediate: true },
)

watch(activeConsoleTab, (tab) => {
  if (tab === 'reports' && !reportLoaded.value) {
    void loadModerationReports(true)
  }
  if (tab === 'logs' && !logLoaded.value) {
    void loadModerationLogs(true)
  }
}, { immediate: true })

function isBusy(id: number) {
  return busyIds.value.includes(id)
}

async function loadModerationURL(url: string, options: { push?: boolean } = {}) {
  if (loadingList.value) return
  loadingList.value = true
  actionError.value = ''
  busyIds.value = []
  try {
    const nextURL = new URL(url, window.location.origin)
    const payload = (await fetchPage(nextURL)) as PagePayload<ModerationPageProps>
    currentProps.value = payload.props
    topics.value = [...payload.props.topics]
    if (options.push !== false) {
      window.history.pushState(window.history.state, '', `${nextURL.pathname}${nextURL.search}${nextURL.hash}`)
    }
  } catch (error) {
    actionError.value = error instanceof Error ? error.message : t('common.loadFailed')
  } finally {
    loadingList.value = false
  }
}

async function moderateTopic(topic: TopicPayload) {
  if (isBusy(topic.id)) return
  busyIds.value = [...busyIds.value, topic.id]
  actionError.value = ''
  try {
    await updateModerationArticleStatus(topic.id, 'unban')
    topics.value = topics.value.filter(item => item.id !== topic.id)
    logLoaded.value = false
    logItems.value = []
    logNextCursor.value = 0
    logHasNext.value = true
  } catch (error) {
    actionError.value = error instanceof Error ? error.message : t('api.moderationActionFailed')
  } finally {
    busyIds.value = busyIds.value.filter(id => id !== topic.id)
  }
}

async function loadModerationLogs(reset = false) {
  if (logLoading.value) return
  logLoading.value = true
  logError.value = ''
  try {
    const payload = await fetchModerationLogs(reset ? 0 : logNextCursor.value, 20)
    logItems.value = reset ? payload.items : mergeModerationLogs(logItems.value, payload.items)
    logNextCursor.value = payload.nextCursor
    logHasNext.value = payload.hasNext
    logLoaded.value = true
  } catch (error) {
    logError.value = error instanceof Error ? error.message : t('api.moderationLogsFailed')
  } finally {
    logLoading.value = false
  }
}

async function loadModerationReports(reset = false) {
  if (reportLoading.value) return
  reportLoading.value = true
  reportError.value = ''
  try {
    const payload = await fetchModerationReports(reset ? 0 : reportNextCursor.value, 20, reportStatus.value)
    reportItems.value = reset ? payload.items : mergeModerationReports(reportItems.value, payload.items)
    reportNextCursor.value = payload.nextCursor
    reportHasNext.value = payload.hasNext
    reportLoaded.value = true
  } catch (error) {
    reportError.value = error instanceof Error ? error.message : t('api.moderationReportsFailed')
  } finally {
    reportLoading.value = false
  }
}

function switchReportStatus(status: 'open' | 'closed') {
  if (reportStatus.value === status) return
  reportStatus.value = status
  reportItems.value = []
  reportNextCursor.value = 0
  reportHasNext.value = true
  reportLoaded.value = false
  void loadModerationReports(true)
}

function mergeModerationReports(current: ModerationReportItem[], incoming: ModerationReportItem[]) {
  const seen = new Set(current.map(item => item.id))
  return [...current, ...incoming.filter(item => !seen.has(item.id))]
}

function reportBusy(id: number) {
  return reportBusyIds.value.includes(id)
}

async function handleReport(item: ModerationReportItem, action: 'resolve' | 'reject') {
  if (reportBusy(item.id)) return
  reportBusyIds.value = [...reportBusyIds.value, item.id]
  reportError.value = ''
  try {
    await updateModerationReportStatus(item.id, action)
    reportItems.value = reportItems.value.filter(report => report.id !== item.id)
    logLoaded.value = false
    logItems.value = []
    logNextCursor.value = 0
    logHasNext.value = true
  } catch (error) {
    reportError.value = error instanceof Error ? error.message : t('api.moderationActionFailed')
  } finally {
    reportBusyIds.value = reportBusyIds.value.filter(id => id !== item.id)
  }
}

async function hideReportTarget(item: ModerationReportItem) {
  if (reportBusy(item.id)) return
  reportBusyIds.value = [...reportBusyIds.value, item.id]
  reportError.value = ''
  try {
    if (item.targetType === 'topic') {
      await updateModerationArticleStatus(item.targetId, 'ban')
    } else {
      await updateModerationReplyStatus(item.targetId, 'ban')
    }
    await updateModerationReportStatus(item.id, 'ban')
    reportItems.value = reportItems.value.filter(report => report.id !== item.id)
    if (item.targetType === 'topic') {
      topics.value = topics.value.filter(topic => topic.id !== item.targetId)
    }
    logLoaded.value = false
    logItems.value = []
    logNextCursor.value = 0
    logHasNext.value = true
  } catch (error) {
    reportError.value = error instanceof Error ? error.message : t('api.moderationActionFailed')
  } finally {
    reportBusyIds.value = reportBusyIds.value.filter(id => id !== item.id)
  }
}

function mergeModerationLogs(current: ModerationLogItem[], incoming: ModerationLogItem[]) {
  const seen = new Set(current.map(item => item.id))
  return [...current, ...incoming.filter(item => !seen.has(item.id))]
}

function logActionLabel(item: ModerationLogItem) {
  const key = `moderation.logs.actions.${item.action}`
  return te(key) ? t(key) : t('moderation.logs.actions.operation')
}

function reportResolutionLabel(item: ModerationReportItem) {
  if (item.resolution) return t(`moderation.reports.resolutions.${item.resolution}`)
  if (item.status === 'rejected') return t('moderation.reports.resolutions.ignored')
  return t('moderation.reports.resolutions.resolved')
}
</script>

<template>
  <main class="min-w-0 pb-8">
    <PageHeader :title="t('moderation.title')" :description="t('moderation.description')" compact class="border-b-0 !mb-2 sm:!mb-2 !pb-2 sm:!pb-2" />

    <div class="mb-4 flex flex-wrap gap-2 border-b border-line">
      <button
        v-for="tab in managementTabs"
        :key="tab.key"
        type="button"
        class="-mb-px inline-flex h-10 items-center gap-2 border-b-2 px-1 text-sm font-semibold transition"
        :class="activeConsoleTab === tab.key ? 'border-primary text-primary' : 'border-transparent text-base-content/55 hover:text-base-content'"
        @click="activeConsoleTab = tab.key as 'reports' | 'ban' | 'logs' | 'guidance'"
      >
        <component :is="tab.icon" class="h-4 w-4" />
        {{ t(`moderation.managementTabs.${tab.key}`) }}
      </button>
    </div>

    <section v-if="activeConsoleTab === 'reports'" class="space-y-3">
      <p v-if="reportError" class="rounded border border-error/25 bg-error/10 px-3 py-2 text-sm text-error">
        {{ reportError }}
      </p>

      <div class="gf-card overflow-hidden">
        <div class="flex items-center gap-1 border-b border-line bg-base-200/60 p-2">
          <button
            type="button"
            class="gf-tab"
            :class="reportStatus === 'open' ? 'bg-base-100 text-base-content shadow-sm ring-1 ring-line' : 'text-base-content/55 hover:bg-base-100/70 hover:text-base-content'"
            @click="switchReportStatus('open')"
          >
            {{ t('moderation.reports.statusTabs.open') }}
          </button>
          <button
            type="button"
            class="gf-tab"
            :class="reportStatus === 'closed' ? 'bg-base-100 text-base-content shadow-sm ring-1 ring-line' : 'text-base-content/55 hover:bg-base-100/70 hover:text-base-content'"
            @click="switchReportStatus('closed')"
          >
            {{ t('moderation.reports.statusTabs.closed') }}
          </button>
        </div>

        <div class="hidden grid-cols-[28px_minmax(260px,1fr)_140px_170px_180px_176px] gap-4 border-b border-line bg-base-200/60 px-3 py-1.5 text-[11px] font-bold uppercase text-base-content/75 lg:grid">
          <div />
          <div>{{ t('moderation.reports.table.report') }}</div>
          <div>{{ t('moderation.reports.table.reason') }}</div>
          <div>{{ t('moderation.reports.table.people') }}</div>
          <div class="text-right">{{ t('moderation.reports.table.time') }}</div>
          <div class="text-right">{{ t('moderation.reports.table.action') }}</div>
        </div>

        <div v-if="reportItems.length" class="divide-y divide-line">
          <article
            v-for="item in reportItems"
            :key="item.id"
            class="grid grid-cols-[28px_minmax(0,1fr)] gap-3 px-3 py-1.5 transition hover:bg-base-200/70 lg:grid-cols-[28px_minmax(260px,1fr)_140px_170px_180px_176px] lg:items-center lg:gap-4"
          >
            <div class="flex h-7 w-7 shrink-0 items-center justify-center rounded bg-base-200 text-warning">
              <Flag class="h-4 w-4" />
            </div>
            <div class="min-w-0">
              <div class="flex min-w-0 items-center gap-1.5 text-[15px] leading-5 text-base-content/80">
                <span class="shrink-0 text-base-content/45">{{ t(`moderation.reports.targetTypes.${item.targetType}`) }}</span>
                <a :href="item.targetUrl" class="min-w-0 max-w-full truncate font-medium text-primary/90 hover:text-primary">{{ item.title }}</a>
                <span v-for="category in item.categories" :key="category.id" class="hidden shrink-0 items-center gap-1 text-xs text-base-content/45 lg:inline-flex">
                  <span class="h-2 w-2 rounded-[3px]" :style="{ backgroundColor: category.color }" />
                  {{ category.name }}
                </span>
              </div>
              <p class="line-clamp-1 text-[13px] leading-5 text-base-content/55">{{ item.excerpt || item.note || t('moderation.reports.noExcerpt') }}</p>
              <div class="mt-1 flex flex-wrap items-center gap-x-2.5 gap-y-1 text-xs text-base-content/50 lg:hidden">
                <span>{{ t('moderation.reports.reasonLabel') }} {{ t(`moderation.reports.reasons.${item.reason}`) }}</span>
                <span v-if="reportStatus === 'closed'">{{ t('moderation.reports.statusLabel') }} {{ reportResolutionLabel(item) }}</span>
                <span v-for="category in item.categories" :key="category.id" class="inline-flex items-center gap-1">
                  <span class="h-2 w-2 rounded-[3px]" :style="{ backgroundColor: category.color }" />
                  {{ category.name }}
                </span>
                <a
                  :href="`/u/${item.reporter.id}`"
                  class="inline-flex min-w-0 items-center gap-1.5 hover:text-primary lg:hidden"
                  @click="showUserCard(item.reporter, $event)"
                >
                  <UserAvatar :src="item.reporter.avatarUrl" alt="" class="h-4 w-4 rounded-full object-cover ring-1 ring-line" />
                  <span class="shrink-0">{{ t('moderation.reports.reporterLabel') }}</span>
                  <span class="max-w-28 truncate font-medium text-base-content/65">{{ item.reporter.username }}</span>
                </a>
                <a
                  v-if="reportStatus === 'closed' && item.handler.id"
                  :href="`/u/${item.handler.id}`"
                  class="inline-flex min-w-0 items-center gap-1.5 hover:text-primary lg:hidden"
                  @click="showUserCard(item.handler, $event)"
                >
                  <UserAvatar :src="item.handler.avatarUrl" alt="" class="h-4 w-4 rounded-full object-cover ring-1 ring-line" />
                  <span class="shrink-0">{{ t('moderation.reports.handlerLabel') }}</span>
                  <span class="max-w-28 truncate font-medium text-base-content/65">{{ item.handler.username }}</span>
                </a>
              </div>
              <time class="mt-1 block text-xs text-base-content/55 lg:hidden">{{ formatDateTime(reportStatus === 'closed' && item.handledAt ? item.handledAt : item.createdAt) }}</time>
            </div>
            <div class="hidden min-w-0 space-y-1 text-[13px] text-base-content/55 lg:block">
              <div class="flex min-w-0 items-center gap-1.5">
                <span class="shrink-0 text-base-content/40">{{ t('moderation.reports.reasonLabel') }}</span>
                <span class="min-w-0 truncate font-medium text-base-content/70">{{ t(`moderation.reports.reasons.${item.reason}`) }}</span>
              </div>
              <div v-if="reportStatus === 'closed'" class="flex min-w-0 items-center gap-1.5">
                <span class="shrink-0 text-base-content/40">{{ t('moderation.reports.statusLabel') }}</span>
                <span class="min-w-0 truncate font-medium text-base-content/70">{{ reportResolutionLabel(item) }}</span>
              </div>
            </div>
            <div class="hidden min-w-0 space-y-1 text-[13px] text-base-content/55 lg:block">
              <a
                :href="`/u/${item.reporter.id}`"
                class="flex min-w-0 items-center gap-1.5 hover:text-primary"
                @click="showUserCard(item.reporter, $event)"
              >
                <UserAvatar :src="item.reporter.avatarUrl" alt="" class="h-5 w-5 rounded-full object-cover ring-1 ring-line" />
                <span class="shrink-0">{{ t('moderation.reports.reporterLabel') }}</span>
                <span class="min-w-0 truncate font-medium text-base-content/65">{{ item.reporter.username }}</span>
              </a>
              <a
                v-if="reportStatus === 'closed' && item.handler.id"
                :href="`/u/${item.handler.id}`"
                class="flex min-w-0 items-center gap-1.5 hover:text-primary"
                @click="showUserCard(item.handler, $event)"
              >
                <UserAvatar :src="item.handler.avatarUrl" alt="" class="h-5 w-5 rounded-full object-cover ring-1 ring-line" />
                <span class="shrink-0">{{ t('moderation.reports.handlerLabel') }}</span>
                <span class="min-w-0 truncate font-medium text-base-content/65">{{ item.handler.username }}</span>
              </a>
            </div>
            <div class="col-start-2 mt-1 flex flex-wrap items-center justify-start gap-2 lg:col-start-auto lg:mt-0 lg:block lg:text-right">
              <div class="space-y-0.5 text-xs tabular-nums text-base-content/55">
                <div>{{ t('moderation.reports.submittedAtLabel') }} {{ formatDateTime(item.createdAt) }}</div>
                <div v-if="reportStatus === 'closed' && item.handledAt">{{ t('moderation.reports.handledAtLabel') }} {{ formatDateTime(item.handledAt) }}</div>
              </div>
            </div>
            <div v-if="reportStatus === 'open'" class="col-start-2 mt-1 flex items-center justify-start gap-2 lg:col-start-auto lg:mt-0 lg:justify-end">
              <button
                type="button"
                class="gf-button gf-button-sm gf-button-danger shrink-0 whitespace-nowrap text-xs"
                :disabled="reportBusy(item.id)"
                @click="hideReportTarget(item)"
              >
                <Ban class="h-4 w-4" />
                {{ t('moderation.reports.ban') }}
              </button>
              <button
                type="button"
                class="gf-button gf-button-sm gf-button-muted shrink-0 whitespace-nowrap text-xs"
                :disabled="reportBusy(item.id)"
                @click="handleReport(item, 'reject')"
              >
                <XCircle class="h-4 w-4" />
                {{ t('moderation.reports.reject') }}
              </button>
            </div>
            <div v-else class="hidden lg:block" />
          </article>
        </div>

        <EmptyState
          v-else-if="reportLoading"
          :icon="Flag"
          :title="t('moderation.reports.loading')"
          loading
        />
        <EmptyState
          v-else
          :icon="Flag"
          :title="t('moderation.reports.emptyTitle')"
          :description="t('moderation.reports.emptyDescription')"
        />

        <footer v-if="reportLoaded && (reportItems.length || reportHasNext)" class="border-t border-line px-4 py-3 text-center text-xs font-semibold text-base-content/55">
          <button
            v-if="reportHasNext"
            type="button"
            class="gf-button gf-button-sm gf-button-ghost"
            :disabled="reportLoading"
            @click="loadModerationReports(false)"
          >
            {{ reportLoading ? t('moderation.reports.loading') : t('moderation.reports.loadMore') }}
          </button>
          <span v-else-if="reportItems.length" class="text-xs text-base-content/45">{{ t('moderation.reports.noMore') }}</span>
        </footer>
      </div>
    </section>

    <section v-else-if="activeConsoleTab === 'ban'" class="space-y-3">
      <div class="flex flex-wrap gap-2">
        <a
          v-for="tab in currentProps.categoryTabs"
          :key="tab.key"
          :href="tab.url"
          class="gf-button gf-button-sm text-xs"
          :class="tab.active ? 'gf-button-secondary' : 'gf-button-ghost'"
          @click.prevent="loadModerationURL(tab.url)"
        >
          {{ tab.label }}
        </a>
      </div>

      <p v-if="actionError" class="rounded border border-error/25 bg-error/10 px-3 py-2 text-sm text-error">
        {{ actionError }}
      </p>

      <div class="overflow-hidden rounded border border-line bg-base-100">
        <TopicList :topics="topics" :show-hot="false">
          <template #activity-header>
            {{ t('moderation.table.action') }}
          </template>
          <template #activity="{ topic }">
            <button
              type="button"
              class="gf-button gf-button-sm gf-button-primary shrink-0 text-xs"
              :disabled="isBusy(topic.id)"
              @click="moderateTopic(topic)"
            >
              <RotateCcw class="h-4 w-4" />
              {{ isBusy(topic.id) ? t('common.loadingShort') : t('moderation.unbanAction') }}
            </button>
          </template>
          <template #mobile-action="{ topic }">
            <span class="ml-auto">
              <button
                type="button"
                class="gf-button gf-button-sm gf-button-primary shrink-0 text-xs"
                :disabled="isBusy(topic.id)"
                @click="moderateTopic(topic)"
              >
                <RotateCcw class="h-4 w-4" />
                {{ isBusy(topic.id) ? t('common.loadingShort') : t('moderation.unbanAction') }}
              </button>
            </span>
          </template>
          <template #empty>
            <EmptyState v-if="!topics.length" :icon="Ban" :title="t('moderation.blockedEmptyTitle')" :description="t('moderation.emptyDescription')" />
          </template>
        </TopicList>

        <footer v-if="currentProps.pagination.hasNext" class="border-t border-line bg-base-200/50 px-4 py-3 text-center">
          <a
            :href="currentProps.pagination.nextUrl"
            class="gf-button gf-button-sm gf-button-secondary"
            rel="next"
            @click.prevent="loadModerationURL(currentProps.pagination.nextUrl)"
          >
            {{ t('common.nextPage') }}
          </a>
        </footer>
      </div>
    </section>

    <section v-else-if="activeConsoleTab === 'logs'" class="space-y-3">
      <p v-if="logError" class="rounded border border-error/25 bg-error/10 px-3 py-2 text-sm text-error">
        {{ logError }}
      </p>

      <div class="gf-card overflow-hidden">
        <div class="hidden grid-cols-[34px_minmax(0,1fr)_116px] gap-3 border-b border-line bg-base-200/60 px-3 py-2 text-[11px] font-bold uppercase text-base-content/75 md:grid">
          <div />
          <div>{{ t('moderation.logs.table.operation') }}</div>
          <div class="text-right">{{ t('moderation.logs.table.time') }}</div>
        </div>

        <div v-if="logItems.length" class="divide-y divide-line">
          <article
            v-for="item in logItems"
            :key="item.id"
            class="grid grid-cols-[34px_minmax(0,1fr)] gap-3 px-3 py-2.5 transition hover:bg-base-200/70 md:grid-cols-[34px_minmax(0,1fr)_116px] md:items-start"
          >
            <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-base-200 text-base-content/55">
              <History class="h-4 w-4" />
            </div>
            <div class="min-w-0">
              <div class="flex min-w-0 items-center gap-1.5 text-sm leading-5">
                <span class="max-w-[42%] shrink-0 truncate font-semibold text-base-content">{{ item.actor.username }}</span>
                <span class="shrink-0 text-base-content/55">{{ logActionLabel(item) }}</span>
                <a
                  v-if="item.subject.url"
                  :href="item.subject.url"
                  class="min-w-0 max-w-full truncate font-semibold text-primary hover:text-primary"
                >
                  {{ item.subject.title }}
                </a>
                <span v-else class="min-w-0 max-w-full truncate font-semibold text-base-content/75">{{ item.subject.title }}</span>
              </div>
              <p v-if="item.subject.excerpt" class="mt-0.5 line-clamp-1 text-xs text-base-content/55">
                {{ item.subject.excerpt }}
              </p>
              <time class="mt-1 block text-xs text-base-content/55 md:hidden">{{ formatDateTime(item.createdAt) }}</time>
            </div>
            <time class="hidden text-right text-xs font-medium tabular-nums text-base-content/55 md:block">{{ formatDateTime(item.createdAt) }}</time>
          </article>
        </div>

        <EmptyState
          v-else-if="logLoading"
          :icon="History"
          :title="t('moderation.logs.loading')"
          loading
        />
        <EmptyState
          v-else
          :icon="History"
          :title="t('moderation.logs.emptyTitle')"
          :description="t('moderation.logs.emptyDescription')"
        />

        <footer v-if="logLoaded && (logItems.length || logHasNext)" class="border-t border-line px-4 py-3 text-center text-xs font-semibold text-base-content/55">
          <button
            v-if="logHasNext"
            type="button"
            class="gf-button gf-button-sm gf-button-ghost"
            :disabled="logLoading"
            @click="loadModerationLogs(false)"
          >
            {{ logLoading ? t('moderation.logs.loading') : t('moderation.logs.loadMore') }}
          </button>
          <span v-else-if="logItems.length" class="text-xs text-base-content/45">{{ t('moderation.logs.noMore') }}</span>
        </footer>
      </div>
    </section>

    <section v-else class="space-y-3 px-4 sm:px-0">
      <div>
        <div class="flex items-start gap-3 py-2.5">
          <CircleAlert class="mt-1 h-4 w-4 shrink-0 text-warning" />
          <p class="text-sm leading-6 text-base-content/70">{{ t('moderation.notice') }}</p>
        </div>
        <div class="space-y-3 pt-1">
          <div class="py-2">
            <h3 class="text-sm font-semibold text-base-content">{{ t('moderation.guidanceItems.rule.title') }}</h3>
            <p class="mt-1 text-sm leading-6 text-base-content/60">{{ t('moderation.guidanceItems.rule.description') }}</p>
          </div>
          <div class="py-2">
            <h3 class="text-sm font-semibold text-base-content">{{ t('moderation.guidanceItems.context.title') }}</h3>
            <p class="mt-1 text-sm leading-6 text-base-content/60">{{ t('moderation.guidanceItems.context.description') }}</p>
          </div>
          <div class="py-2">
            <h3 class="text-sm font-semibold text-base-content">{{ t('moderation.guidanceItems.restraint.title') }}</h3>
            <p class="mt-1 text-sm leading-6 text-base-content/60">{{ t('moderation.guidanceItems.restraint.description') }}</p>
          </div>
        </div>
      </div>
    </section>
  </main>
</template>
