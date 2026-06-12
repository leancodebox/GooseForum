<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  icon?: Component
  title: string
  description?: string
  loading?: boolean
}>()
</script>

<template>
  <div class="flex min-h-56 flex-col items-center justify-center px-6 text-center">
    <component
      :is="icon"
      v-if="icon"
      class="h-8 w-8 text-base-content/35"
      :class="{ 'animate-spin': loading }"
    />
    <h2 class="mt-2 text-base font-semibold text-base-content">{{ title }}</h2>
    <p v-if="description" class="mt-1 text-sm text-base-content/55">{{ description }}</p>
    <div v-if="$slots.default" class="mt-4">
      <slot />
    </div>
  </div>
</template>
