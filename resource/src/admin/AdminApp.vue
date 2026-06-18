<script setup lang="ts">
import { RouterView } from 'vue-router'
import AdminLayout from '@/admin/layouts/AdminLayout.vue'
import type { AdminPayload, ManageHomeProps } from '@/admin/types'

const props = defineProps<{
  payload: AdminPayload
}>()
</script>

<template>
  <AdminLayout :layout="payload.layout">
    <RouterView v-slot="{ Component, route }">
      <component
        :is="Component"
        :payload="props.payload as AdminPayload<ManageHomeProps>"
        v-bind="route.meta.pageProps || {}"
      />
    </RouterView>
  </AdminLayout>
</template>
