<script setup lang="ts">
import { ArrowUpRight } from '@lucide/vue'
import { Button } from '@/admin/components/ui/button'
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
            这个入口已经接入 resource 管理后台。完整功能会按旧后台能力逐页迁移，迁移期间可以继续通过旧后台处理生产操作。
          </p>
          <div class="mt-5">
            <Button as-child>
              <a :href="payload.props.legacyAdminUrl" target="_blank" rel="noopener noreferrer">
                打开旧后台
                <ArrowUpRight class="size-4" />
              </a>
            </Button>
          </div>
        </div>
      </section>
    </BasicPage>
  </AdminLayout>
</template>
