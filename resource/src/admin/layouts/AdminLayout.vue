<script setup lang="ts">
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
</script>

<template>
  <SidebarProvider>
    <AppSidebar :layout="layout" />
    <SidebarInset class="w-full max-w-full peer-data-[state=collapsed]:w-[calc(100%-var(--sidebar-width-icon)-1rem)] peer-data-[state=expanded]:w-[calc(100%-var(--sidebar-width))]">
      <AdminTopbar :layout="layout" />
      <div class="grow px-4 pb-4">
        <slot />
      </div>
    </SidebarInset>
  </SidebarProvider>
</template>
