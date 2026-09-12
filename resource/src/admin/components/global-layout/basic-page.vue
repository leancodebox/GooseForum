<script lang="ts" setup>
import type { LayoutHeaderProps } from './types'

import BasicHeader from './basic-header.vue'

defineProps<LayoutHeaderProps>()
</script>

<template>
  <div class="flex min-w-0 flex-1 flex-col">
    <BasicHeader
      :title="title"
      :description="description"
    >
      <template #primary-action>
        <slot name="primary-action" />
      </template>
    </BasicHeader>

    <main class="min-w-0 px-3 py-3 md:px-4 md:py-4">
      <slot />
    </main>
  </div>
</template>
