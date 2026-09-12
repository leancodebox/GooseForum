<script lang="ts" setup>
import type { LayoutHeaderProps } from './types'

defineProps<LayoutHeaderProps>()
</script>

<template>
  <Teleport defer to="#admin-topbar-page-context">
    <div class="min-w-0">
      <h1 class="truncate text-base font-semibold tracking-tight text-foreground">
        {{ title }}
      </h1>
    </div>
  </Teleport>

  <Teleport defer to="#admin-topbar-primary-action">
    <div class="flex shrink-0 items-center gap-2">
      <slot name="primary-action" />
    </div>
  </Teleport>
</template>
