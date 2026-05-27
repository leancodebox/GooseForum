<script setup lang="ts">
import type { AdminPayload } from '@/admin/types'

defineProps<{
  payload: AdminPayload
}>()
</script>

<template>
  <section class="space-y-5 py-6">
    <div class="space-y-3">
      <div class="h-10 w-52 animate-pulse rounded-lg bg-muted" />
      <div class="h-5 w-80 max-w-full animate-pulse rounded bg-muted" />
    </div>
    <div class="rounded-lg border bg-card p-4">
      <div class="mb-4 h-10 w-full animate-pulse rounded-md bg-muted" />
      <div class="space-y-3">
        <div class="h-14 animate-pulse rounded-md bg-muted/70" />
        <div class="h-14 animate-pulse rounded-md bg-muted/70" />
        <div class="h-14 animate-pulse rounded-md bg-muted/70" />
      </div>
    </div>
  </section>


</template>
