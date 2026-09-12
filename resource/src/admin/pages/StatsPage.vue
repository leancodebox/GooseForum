<script setup lang="ts">import { adminText } from '@/admin/runtime/i18n-text'

import { FileText, Link as LinkIcon, MessageSquare, Users } from '@lucide/vue'
import { computed, defineAsyncComponent, onMounted, ref } from 'vue'
import AdminSection from '@/admin/components/AdminSection.vue'
import { Badge } from '@/admin/components/ui/badge'
import {
  Card,
  CardAction,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/admin/components/ui/card'
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
    value: stats.value?.topicMaxId,
    delta: `+${stats.value?.topicMonthCount ?? 0}`,
    icon: FileText,
    visible: true,
  },
  {
    label: adminText('k002i'),
    value: stats.value?.postMaxId,
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
  >
    <template #primary-action>
      <div class="inline-flex h-8 max-w-[48vw] items-center gap-1.5 rounded-md border bg-muted/35 px-2.5 text-sm text-muted-foreground">
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
    </template>

      <div class="mb-4 grid grid-cols-1 gap-3 sm:grid-cols-2 md:gap-4 @5xl/main:grid-cols-4">
        <Card
          v-for="item in summaryItems"
          :key="item.label"
          class="@container/card min-h-32 gap-3 bg-gradient-to-t from-primary/5 to-card py-4 shadow-xs"
        >
          <CardHeader class="gap-2.5 px-4">
            <CardDescription class="flex items-center gap-2 font-medium">
              <component :is="item.icon" class="size-4" />
              {{ item.label }}
            </CardDescription>
            <CardTitle class="text-3xl font-semibold tabular-nums tracking-tight">
              {{ formatNumber(item.value) }}
            </CardTitle>
            <CardAction>
              <Badge v-if="item.delta" variant="outline" class="bg-background/70 tabular-nums">
                {{ item.delta }}
              </Badge>
            </CardAction>
          </CardHeader>
        </Card>
      </div>

      <div class="grid grid-cols-1 gap-3 md:gap-4 xl:grid-cols-[minmax(0,1fr)_360px]">
        <div class="min-w-0 min-h-110">
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
