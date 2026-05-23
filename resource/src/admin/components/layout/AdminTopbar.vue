<script setup lang="ts">
import { ArrowLeft, Search } from '@lucide/vue'
import { Button } from '@/admin/components/ui/button'
import { Separator } from '@/admin/components/ui/separator'
import { SidebarTrigger } from '@/admin/components/ui/sidebar'
import type { LayoutPayload } from '@/types/payload'

defineProps<{
  layout: LayoutPayload
}>()
</script>

<template>
  <header class="sticky top-0 z-50 flex h-16 shrink-0 items-center gap-3 bg-background p-4 transition-[width,height] ease-linear sm:gap-4">
    <SidebarTrigger class="-ml-1" />
    <Separator orientation="vertical" class="h-6" />

    <button
      type="button"
      class="hidden h-10 w-72 items-center gap-3 rounded-md border bg-background px-3 text-sm text-muted-foreground shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground md:flex"
    >
      <Search class="size-4" />
      <span class="flex-1 text-left">Search Menu</span>
      <kbd class="pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground">
        ⌘ + K
      </kbd>
    </button>

    <div class="flex-1" />

    <div class="ml-auto flex items-center space-x-4">
      <Button as-child class="hidden md:inline-flex">
        <a href="/">
          <ArrowLeft class="size-4" />
          返回前台
        </a>
      </Button>
      <img
        v-if="layout.viewer.isAuthenticated"
        :src="layout.viewer.avatarUrl"
        :alt="layout.viewer.username"
        class="size-10 rounded-full object-cover"
      />
    </div>
  </header>
</template>
