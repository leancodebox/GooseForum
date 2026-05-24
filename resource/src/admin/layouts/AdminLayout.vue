<script setup lang="ts">
import { inject, provide } from 'vue'
import AppSidebar from '@/admin/components/layout/AppSidebar.vue'
import AdminTopbar from '@/admin/components/layout/AdminTopbar.vue'
import {
  SidebarInset,
  SidebarProvider,
} from '@/admin/components/ui/sidebar'
import type { LayoutPayload } from '@/types/payload'

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
  <SidebarProvider v-else>
    <AppSidebar :layout="layout" />
    <SidebarInset class="w-full max-w-full peer-data-[state=collapsed]:w-[calc(100%-var(--sidebar-width-icon)-1rem)] peer-data-[state=expanded]:w-[calc(100%-var(--sidebar-width))]">
      <AdminTopbar :layout="layout" />
      <div class="grow px-4 pb-4">
        <slot />
      </div>
    </SidebarInset>
  </SidebarProvider>
</template>
