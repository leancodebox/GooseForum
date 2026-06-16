<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/admin/utils/cn'

const props = defineProps<{
  class?: HTMLAttributes['class']
  bodyClass?: HTMLAttributes['class']
}>()
</script>

<template>
  <section :class="cn('overflow-hidden rounded-lg border bg-card text-card-foreground', props.class)">
    <div v-if="$slots.header" class="border-b bg-muted/20 px-3 py-2">
      <slot name="header" />
    </div>
    <div :class="cn(props.bodyClass)">
      <slot />
    </div>
  </section>
</template>
