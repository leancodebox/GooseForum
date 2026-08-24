<script lang="ts" setup>
import type { LayoutHeaderProps } from './types'

defineProps<LayoutHeaderProps>()
</script>

<template>
  <Teleport defer to="#admin-topbar-page-context">
    <div class="min-w-0">
      <h1 class="truncate text-base font-semibold sm:text-lg">
        {{ title }}
      </h1>
      <p v-if="description" class="mt-0.5 hidden truncate text-xs text-muted-foreground lg:block">
        {{ description }}
      </p>
    </div>
  </Teleport>

  <Teleport defer to="#admin-topbar-primary-action">
    <div class="flex shrink-0 items-center gap-2">
      <slot name="primary-action" />
    </div>
  </Teleport>
</template>
