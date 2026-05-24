<script lang="ts" setup>
import { ref } from 'vue'
import AdminLayout from '@/admin/layouts/AdminLayout.vue'
import { BasicPage } from '@/admin/components/global-layout'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/admin/components/ui/tabs'
import OverviewContent from '@/admin/pages/dashboard/components/overview-content.vue'
import type { AdminPayload, ManageHomeProps } from '@/admin/types'

defineProps<{
  payload: AdminPayload<ManageHomeProps>
}>()

const tabs = ref([
  { name: 'Overview', value: 'overview' },
  { name: 'Analytics', value: 'analytics', disabled: true },
  { name: 'Reports', value: 'reports', disabled: true },
  { name: 'Notifications', value: 'notifications', disabled: true },
])

const activeTab = ref(tabs.value[0].value)
</script>

<template>
  <AdminLayout :layout="payload.layout">
    <BasicPage
      title="workspace"
      description="workspace description"
      sticky
    >
      <Tabs :default-value="activeTab" class="w-full">
        <TabsList>
          <TabsTrigger
            v-for="tab in tabs"
            :key="tab.value"
            :value="tab.value"
            :disabled="tab.disabled"
          >
            {{ tab.name }}
          </TabsTrigger>
        </TabsList>
        <TabsContent value="overview" class="space-y-4">
          <OverviewContent />
        </TabsContent>
      </Tabs>
    </BasicPage>
  </AdminLayout>
</template>
