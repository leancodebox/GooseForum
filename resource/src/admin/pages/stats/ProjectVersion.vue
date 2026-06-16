<script setup lang="ts">
import { adminText } from '@/admin/runtime/i18n-text'
import { AlertCircle, ChevronRight, Code2, ExternalLink, Loader2, Tag } from '@lucide/vue'
import AdminSection from '@/admin/components/AdminSection.vue'
import { Button } from '@/admin/components/ui/button'
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
  if (diff < hour) return adminText('k00c9', { count: Math.max(1, Math.floor(diff / minute)) })
  if (diff < day) return adminText('k00ca', { count: Math.floor(diff / hour) })
  return adminText('k00cb', { count: Math.floor(diff / day) })
}
</script>

<template>
  <AdminSection class="flex h-full flex-col" body-class="flex min-h-0 flex-1 flex-col p-0">
    <template #header>
      <div class="flex items-start justify-between gap-4">
      <div class="space-y-1">
        <h2 class="flex items-center gap-2 text-base font-semibold">
          <Code2 class="h-4 w-4" />
          {{ adminText('k0047') }}
        </h2>
        <p class="text-sm text-muted-foreground">{{ adminText('k002a') }}</p>
      </div>
      <Button variant="ghost" size="sm" as-child>
        <a
          href="https://github.com/leancodebox/GooseForum/releases"
          target="_blank"
          rel="noopener noreferrer"
          class="flex items-center gap-1"
        >
          {{ adminText('k0048') }} <ExternalLink class="h-3 w-3" />
        </a>
      </Button>
      </div>
    </template>
      <div class="min-h-0 flex-1 overflow-y-auto" style="max-height: 383px">
        <div v-if="loading" class="flex h-full w-full flex-col items-center justify-center space-y-2 text-muted-foreground">
          <Loader2 class="h-6 w-6 animate-spin" />
          <p class="text-sm">{{ adminText('k002b') }}</p>
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
                    {{ release.body || adminText('k004a') }}
                  </span>
                  <span
                    v-if="release.prerelease"
                    class="inline-flex h-4 shrink-0 items-center rounded bg-amber-100 px-1 text-[10px] font-medium text-amber-700"
                  >
                    {{ adminText('k0049') }}
                  </span>
                  <span
                    v-if="release.draft"
                    class="inline-flex h-4 shrink-0 items-center rounded border px-1 text-[10px] font-medium"
                  >
                    {{ adminText('k003r') }}
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
          <p class="text-sm">{{ adminText('k002c') }}</p>
        </div>
      </div>
  </AdminSection>
</template>
