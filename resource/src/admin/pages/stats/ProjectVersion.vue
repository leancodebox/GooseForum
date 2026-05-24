<script setup lang="ts">
import { AlertCircle, ChevronRight, Code2, ExternalLink, Loader2, Tag } from '@lucide/vue'
import { Button } from '@/admin/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/admin/components/ui/card'
import type { GithubRelease } from '@/admin/types'

defineProps<{
  releases: GithubRelease[]
  loading: boolean
  error?: string
}>()

function releaseTime(value: string) {
  const time = new Date(value).getTime()
  if (!Number.isFinite(time)) return ''
  const diff = Date.now() - time
  const minute = 60 * 1000
  const hour = 60 * minute
  const day = 24 * hour
  if (diff < hour) return `${Math.max(1, Math.floor(diff / minute))} 分钟前`
  if (diff < day) return `大约 ${Math.floor(diff / hour)} 小时前`
  return `${Math.floor(diff / day)} 天前`
}
</script>

<template>
  <Card class="flex h-full flex-col overflow-hidden">
    <CardHeader class="flex flex-row items-start justify-between gap-4 border-b px-5 py-3.5">
      <div class="space-y-1">
        <CardTitle class="flex items-center gap-2 text-base font-semibold">
          <Code2 class="h-4 w-4" />
          项目版本
        </CardTitle>
        <CardDescription>系统更新与发布记录</CardDescription>
      </div>
      <Button variant="ghost" size="sm" as-child>
        <a
          href="https://github.com/leancodebox/GooseForum/releases"
          target="_blank"
          rel="noopener noreferrer"
          class="flex items-center gap-1"
        >
          查看全部 <ExternalLink class="h-3 w-3" />
        </a>
      </Button>
    </CardHeader>
    <CardContent class="flex min-h-0 flex-1 flex-col p-0">
      <div class="min-h-0 flex-1 overflow-y-auto" style="max-height: 383px">
        <div v-if="loading" class="flex h-full w-full flex-col items-center justify-center space-y-2 text-muted-foreground">
          <Loader2 class="h-6 w-6 animate-spin" />
          <p class="text-sm">加载版本信息中...</p>
        </div>
        <div v-else-if="error" class="flex h-full w-full flex-col items-center justify-center space-y-2 text-destructive">
          <AlertCircle class="h-6 w-6" />
          <p class="text-sm">{{ error }}</p>
        </div>
        <div v-else-if="releases.length > 0" class="divide-y border-t-0">
          <div v-for="release in releases" :key="release.id" class="group px-5 py-3 transition-colors hover:bg-muted/35">
            <div class="flex items-center justify-between gap-3">
              <div class="min-w-0 flex-1">
                <div class="flex min-w-0 items-center gap-2">
                  <Tag class="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                  <span class="shrink-0 text-sm font-semibold">{{ release.tag_name }}</span>
                  <span class="min-w-0 truncate text-xs text-muted-foreground">
                    {{ release.body || '暂无发布说明' }}
                  </span>
                  <span
                    v-if="release.prerelease"
                    class="inline-flex h-4 shrink-0 items-center rounded bg-amber-100 px-1 text-[10px] font-medium text-amber-700"
                  >
                    预发布
                  </span>
                  <span
                    v-if="release.draft"
                    class="inline-flex h-4 shrink-0 items-center rounded border px-1 text-[10px] font-medium"
                  >
                    草稿
                  </span>
                </div>
                <p class="mt-1 text-xs text-muted-foreground">
                  {{ releaseTime(release.published_at) }}
                </p>
              </div>
              <Button variant="ghost" size="icon" class="h-8 w-8 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100" as-child>
                <a :href="release.html_url" target="_blank" rel="noopener noreferrer">
                  <ChevronRight class="h-4 w-4" />
                </a>
              </Button>
            </div>
          </div>
        </div>
        <div v-else class="flex h-full w-full flex-col items-center justify-center space-y-2 text-muted-foreground">
          <Code2 class="h-10 w-10 opacity-20" />
          <p class="text-sm">暂无发布信息</p>
        </div>
      </div>
    </CardContent>
  </Card>
</template>
