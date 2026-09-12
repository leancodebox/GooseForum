<script setup lang="ts">
import { inject, provide } from 'vue'
import AppSidebar from '@/admin/components/layout/AppSidebar.vue'
import AdminTopbar from '@/admin/components/layout/AdminTopbar.vue'
import {
  SidebarInset,
  SidebarProvider,
} from '@/admin/components/ui/sidebar'
import type { LayoutPayload } from '@gooseforum/client'

defineProps<{
  layout: LayoutPayload
}>()

const adminLayoutContextKey = Symbol.for('gooseforum.admin.layout')
const hasParentLayout = inject(adminLayoutContextKey, false)

if (!hasParentLayout) {
  provide(adminLayoutContextKey, true)
}
</script>

<template>
  <slot v-if="hasParentLayout" />
  <SidebarProvider
    v-else
    class="admin-shell"
    :style="{
      '--sidebar-width': '18rem',
      '--header-height': '3rem',
    }"
  >
    <AppSidebar :layout="layout" variant="inset" />
    <SidebarInset class="min-w-0 max-w-full overflow-hidden border-sidebar-border/70 md:rounded-2xl md:border">
      <AdminTopbar :layout="layout" />
      <div class="@container/main flex min-w-0 grow flex-col overflow-x-hidden">
        <slot />
      </div>
    </SidebarInset>
  </SidebarProvider>
</template>
