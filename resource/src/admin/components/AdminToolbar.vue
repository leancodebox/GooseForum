<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/admin/utils/cn'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()
</script>

<template>
  <div :class="cn('flex flex-col gap-2 border-b bg-muted/20 px-3 py-2 lg:flex-row lg:items-center lg:justify-between', props.class)">
    <slot />
  </div>
</template>
