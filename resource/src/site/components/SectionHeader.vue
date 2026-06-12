<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  icon?: Component
  title: string
  description?: string
}>()
</script>

<template>
  <header class="flex items-center justify-between gap-3 border-b border-line px-4 py-3">
    <div class="flex min-w-0 items-center gap-2">
      <component :is="icon" v-if="icon" class="h-4 w-4 shrink-0 text-base-content/55" />
      <div class="min-w-0">
        <h2 class="truncate text-sm font-semibold text-base-content">{{ title }}</h2>
        <p v-if="description" class="mt-0.5 text-xs text-base-content/55">{{ description }}</p>
      </div>
    </div>
    <div v-if="$slots.actions" class="shrink-0">
      <slot name="actions" />
    </div>
  </header>
</template>
