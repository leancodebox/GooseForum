<script lang="ts" setup>
import type { LayoutHeaderProps } from './types'

import BasicHeader from './basic-header.vue'

defineProps<LayoutHeaderProps>()
</script>

<template>
  <div class="min-w-0">
    <BasicHeader
      :title="title"
      :description="description"
    >
      <template #primary-action>
        <slot name="primary-action" />
      </template>
    </BasicHeader>

    <main class="min-w-0">
      <slot />
    </main>
  </div>
</template>
