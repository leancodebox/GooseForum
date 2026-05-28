<script setup lang="ts">
import { adminText } from '@/admin/runtime/i18n-text'
import { BasicPage } from '@/admin/components/global-layout'
import type { AdminPayload, ManageHomeProps } from '@/admin/types'

defineProps<{
  payload: AdminPayload<ManageHomeProps>
  title: string
  description: string
}>()
</script>

<template>
  <BasicPage :title="title" :description="description" sticky>
    <section class="rounded-xl border bg-card p-6 text-card-foreground shadow-sm">
      <div class="max-w-2xl">
        <p class="text-sm font-medium text-muted-foreground">Resource Admin</p>
        <h2 class="mt-2 text-2xl font-semibold tracking-tight">{{ title }}</h2>
        <p class="mt-3 text-sm leading-6 text-muted-foreground">
          {{ adminText('k004b') }}
        </p>
      </div>
    </section>
  </BasicPage>


</template>
