<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/admin/utils/cn'

const props = defineProps<{
  class?: HTMLAttributes['class']
  bodyClass?: HTMLAttributes['class']
}>()
</script>

<template>
  <section :class="cn('overflow-hidden rounded-xl border bg-card text-card-foreground shadow-xs', props.class)">
    <div v-if="$slots.header" class="border-b px-4 py-3">
      <slot name="header" />
    </div>
    <div :class="cn(props.bodyClass)">
      <slot />
    </div>
  </section>
</template>
