<script setup lang="ts">
import { BasicPage } from '@/admin/components/global-layout'
import AdminLayout from '@/admin/layouts/AdminLayout.vue'
import type { AdminPayload, ManageHomeProps } from '@/admin/types'

defineProps<{
  payload: AdminPayload<ManageHomeProps>
  title: string
  description: string
}>()
</script>

<template>
  <AdminLayout :layout="payload.layout">
    <BasicPage :title="title" :description="description" sticky>
      <section class="rounded-xl border bg-card p-6 text-card-foreground shadow-sm">
        <div class="max-w-2xl">
          <p class="text-sm font-medium text-muted-foreground">Resource Admin</p>
          <h2 class="mt-2 text-2xl font-semibold tracking-tight">{{ title }}</h2>
          <p class="mt-3 text-sm leading-6 text-muted-foreground">
            这个入口已经接入 resource 管理后台，但暂时没有对应的页面配置。
          </p>
        </div>
      </section>
    </BasicPage>
  </AdminLayout>
</template>
