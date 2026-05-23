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
  <Card class="flex h-full flex-col">
    <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-1">
      <div class="space-y-1">
        <CardTitle class="flex items-center gap-2 text-lg font-semibold">
          <Code2 class="h-5 w-5" />
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
    <CardContent class="flex min-h-0 flex-1 flex-col p-4 pt-0">
      <div class="min-h-0 flex-1 overflow-y-auto rounded-lg border bg-muted/30 p-4" style="max-height: 350px">
        <div v-if="loading" class="flex h-full w-full flex-col items-center justify-center space-y-2 text-muted-foreground">
          <Loader2 class="h-6 w-6 animate-spin" />
          <p class="text-sm">加载版本信息中...</p>
        </div>
        <div v-else-if="error" class="flex h-full w-full flex-col items-center justify-center space-y-2 text-destructive">
          <AlertCircle class="h-6 w-6" />
          <p class="text-sm">{{ error }}</p>
        </div>
        <div v-else-if="releases.length > 0" class="space-y-4">
          <div v-for="(release, index) in releases" :key="release.id" class="group">
            <div class="flex items-start justify-between gap-4">
              <div class="min-w-0 flex-1 space-y-1">
                <div class="flex flex-wrap items-center gap-2">
                  <Tag class="h-4 w-4 shrink-0 text-primary" />
                  <span class="truncate font-medium">{{ release.tag_name }}</span>
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
                <p class="text-xs text-muted-foreground">
                  {{ releaseTime(release.published_at) }}
                </p>
                <p class="mt-2 line-clamp-2 break-words text-sm text-foreground/80">
                  {{ release.body || '暂无发布说明' }}
                </p>
              </div>
              <Button variant="ghost" size="icon" class="h-8 w-8 shrink-0 opacity-0 transition-opacity group-hover:opacity-100" as-child>
                <a :href="release.html_url" target="_blank" rel="noopener noreferrer">
                  <ChevronRight class="h-4 w-4" />
                </a>
              </Button>
            </div>
            <div v-if="index < releases.length - 1" class="mt-4 border-t" />
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
