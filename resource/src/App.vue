<script setup lang="ts">
import { computed } from 'vue'
import type { Component } from 'vue'
import AppShell from '@/components/AppShell.vue'
import GlobalFlash from '@/components/GlobalFlash.vue'
import { useShellState } from '@/runtime/shell-state'
import type { PagePayload } from '@/types/payload'

const props = defineProps<{
  payload: PagePayload
  component: Component
}>()

const shellState = useShellState()
const standaloneComponents = new Set(['auth.login', 'auth.resetPassword'])
const isStandalone = computed(() => standaloneComponents.has(props.payload.component))
const hasRail = computed(() => props.payload.component === 'article.detail')
</script>

<template>
  <GlobalFlash />
  <template v-if="isStandalone">
    <component
      :is="component"
      :key="payload.url"
      :layout="payload.layout"
      :props="payload.props"
    />
  </template>
  <template v-else>
    <AppShell
      :layout="payload.layout"
      :header-title="shellState.headerTitle"
      :show-header-title="shellState.showHeaderTitle"
      :rail="hasRail"
    >
      <component
        :is="component"
        :key="payload.url"
        :layout="payload.layout"
        :props="payload.props"
      />
    </AppShell>
  </template>
</template>
