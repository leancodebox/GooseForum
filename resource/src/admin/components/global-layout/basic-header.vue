<script lang="ts" setup>
import { cn } from '@/admin/utils/cn'

import type { LayoutHeaderProps } from './types'

defineProps<LayoutHeaderProps>()
</script>

<template>
  <header
    :class="cn(
      'flex flex-col gap-3 border-b py-4 md:flex-row md:items-start md:justify-between',
      sticky ? 'sticky top-16 z-40 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/85' : '',
    )"
  >
    <main class="min-w-0">
      <h1 class="text-2xl font-semibold tracking-tight">
        {{ title }}
      </h1>
      <p v-if="description" class="mt-1 max-w-3xl text-sm leading-6 text-muted-foreground">
        {{ description }}
      </p>
    </main>

    <aside class="flex flex-wrap items-center gap-2 md:justify-end">
      <slot name="actions" />
    </aside>
  </header>
</template>
