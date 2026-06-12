<script setup lang="ts">
defineProps<{
  title: string
  description?: string
  compact?: boolean
}>()
</script>

<template>
  <header
    class="gf-page-header"
    :class="compact ? 'sm:mb-3' : 'sm:mb-4'"
  >
    <div class="min-w-0">
      <div class="flex min-w-0 flex-wrap items-center gap-2">
        <h1 class="min-w-0 truncate text-xl font-bold text-base-content sm:text-2xl">{{ title }}</h1>
        <slot name="badge" />
      </div>
      <p v-if="description" class="mt-1 max-w-2xl text-sm leading-6 text-base-content/55">{{ description }}</p>
      <slot name="meta" />
    </div>
    <div v-if="$slots.actions" class="flex shrink-0 flex-wrap items-center gap-2">
      <slot name="actions" />
    </div>
  </header>
</template>
