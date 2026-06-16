<script setup lang="ts">import { adminText } from '@/admin/runtime/i18n-text'

import { FileText, Link as LinkIcon, MessageSquare, Users } from '@lucide/vue'
import { computed, defineAsyncComponent, onMounted, ref } from 'vue'
import AdminSection from '@/admin/components/AdminSection.vue'
import { BasicPage } from '@/admin/components/global-layout'
import {
  getGithubReleases,
  getServerVersion,
  getSiteStatistics,
  getTrafficOverview,
} from '@/admin/runtime/api'
import type {
  AdminPayload,
  DailyTraffic,
  GithubRelease,
  ManageHomeProps,
  ServerVersion,
  SiteStatistics,
} from '@/admin/types'

defineProps<{
  payload: AdminPayload<ManageHomeProps>
}>()

const DateRangePicker = defineAsyncComponent(() => import('@/admin/pages/stats/DateRangePicker.vue'))
const ProjectVersion = defineAsyncComponent(() => import('@/admin/pages/stats/ProjectVersion.vue'))
const TrafficOverview = defineAsyncComponent(() => import('@/admin/pages/stats/TrafficOverview.vue'))

const stats = ref<SiteStatistics>()
const statsLoading = ref(true)
const traffic = ref<DailyTraffic[]>([])
const trafficLoading = ref(true)
const serverVersion = ref<ServerVersion>()
const serverVersionLoading = ref(true)
const releases = ref<GithubRelease[]>([])
const releasesLoading = ref(true)
const releasesError = ref('')

const endDate = ref(formatDate(new Date()))
const startDate = ref(formatDate(addDays(new Date(), -7)))

const summaryItems = computed(() => [
  {
    label: adminText('k002g'),
    value: stats.value?.userCount,
    delta: `+${stats.value?.userMonthCount ?? 0}`,
    icon: Users,
    visible: true,
  },
  {
    label: adminText('k002h'),
    value: stats.value?.articleCount,
    delta: `+${stats.value?.articleMonthCount ?? 0}`,
    icon: FileText,
    visible: true,
  },
  {
    label: adminText('k002i'),
    value: stats.value?.reply,
    icon: MessageSquare,
    visible: true,
  },
  {
    label: adminText('k002j'),
    value: stats.value?.linksCount,
    icon: LinkIcon,
    visible: true,
  },
])

function addDays(date: Date, days: number) {
  const next = new Date(date)
  next.setDate(next.getDate() + days)
  return next
}

function formatDate(date: Date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function formatNumber(value?: number) {
  return statsLoading.value || value === undefined ? '...' : value.toLocaleString()
}

function modeLabel(mode?: string) {
  switch (mode) {
    case 'release':
      return adminText('k002k')
    case 'snapshot':
      return adminText('k002l')
    case 'development':
      return adminText('k002m')
    case 'custom':
      return adminText('k002n')
    default:
      return adminText('k002o')
  }
}

function shortCommit(commit?: string) {
  return commit ? commit.slice(0, 7) : ''
}

async function loadStats() {
  statsLoading.value = true
  try {
    stats.value = await getSiteStatistics()
  } finally {
    statsLoading.value = false
  }
}

async function loadTraffic() {
  trafficLoading.value = true
  try {
    traffic.value = await getTrafficOverview(startDate.value, endDate.value)
  } finally {
    trafficLoading.value = false
  }
}

async function loadServerVersion() {
  serverVersionLoading.value = true
  try {
    serverVersion.value = await getServerVersion()
  } finally {
    serverVersionLoading.value = false
  }
}

async function loadReleases() {
  releasesLoading.value = true
  releasesError.value = ''
  try {
    releases.value = await getGithubReleases()
  } catch (error) {
    console.error('Failed to fetch releases:', error)
    releasesError.value = error instanceof Error ? error.message : adminText('k002p')
  } finally {
    releasesLoading.value = false
  }
}

onMounted(() => {
  void loadStats()
  void loadTraffic()
  void loadServerVersion()
  void loadReleases()
})
</script>

<template>
  <BasicPage
    :title="adminText('k004c')"
    :description="adminText('k004d')"
    sticky
  >
    <template #actions>
      <div class="flex flex-wrap items-center gap-3">
        <div class="inline-flex max-w-full items-center gap-2 rounded-md border bg-muted/35 px-2.5 py-1 text-xs text-muted-foreground">
          <span class="h-1.5 w-1.5 shrink-0 rounded-full bg-emerald-500" />
          <span class="shrink-0">{{ adminText('k002q') }}</span>
          <span class="truncate font-semibold text-foreground">{{ serverVersionLoading ? adminText('k004e') : serverVersion?.version || 'dev' }}</span>
          <span
            v-if="!serverVersionLoading"
            class="inline-flex h-5 shrink-0 items-center rounded-md bg-secondary px-1.5 text-[10px] font-medium text-secondary-foreground"
          >
            {{ modeLabel(serverVersion?.mode) }}
          </span>
          <span v-if="!serverVersionLoading && shortCommit(serverVersion?.commit)" class="hidden text-muted-foreground sm:inline">
            #{{ shortCommit(serverVersion?.commit) }}
          </span>
        </div>
      </div>
    </template>

      <AdminSection class="mb-4">
        <div class="grid sm:grid-cols-2 xl:grid-cols-4">
        <div
          v-for="(item, index) in summaryItems"
          :key="item.label"
          class="flex min-h-20 items-center justify-between border-b px-4 py-3 text-card-foreground sm:[&:nth-child(2n+1)]:border-r xl:border-b-0 xl:border-r xl:last:border-r-0"
          :class="index >= 2 ? 'sm:border-b-0' : ''"
        >
          <div class="min-w-0">
            <div class="mb-2 flex items-center gap-1.5 text-[13px] font-medium text-muted-foreground">
              <component :is="item.icon" class="h-4 w-4" />
              <span>{{ item.label }}</span>
            </div>
            <div class="flex items-baseline gap-1.5">
              <span class="text-2xl font-bold leading-none tracking-tight">{{ formatNumber(item.value) }}</span>
              <span v-if="item.delta" class="text-xs font-medium text-muted-foreground">{{ item.delta }}</span>
            </div>
          </div>
        </div>
        </div>
      </AdminSection>

      <div class="grid grid-cols-1 gap-4 xl:grid-cols-[minmax(0,1fr)_360px]">
        <div class="min-w-0">
          <TrafficOverview :data="traffic" :loading="trafficLoading">
            <template #headerAction>
              <DateRangePicker
                v-model:start-date="startDate"
                v-model:end-date="endDate"
                @change="loadTraffic"
              />
            </template>
          </TrafficOverview>
        </div>
        <div class="min-w-0">
          <ProjectVersion
            :releases="releases"
            :loading="releasesLoading"
            :error="releasesError"
          />
        </div>
      </div>
    </BasicPage>
</template>
