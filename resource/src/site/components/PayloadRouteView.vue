<script setup lang="ts">
import { computed } from 'vue'
import type { PreparedPage } from '@/runtime/router'
import AppShell from '@/site/components/AppShell.vue'
import { useShellState } from '@/runtime/shell-state'

const props = defineProps<{
  page: PreparedPage
}>()

const shellState = useShellState()
const standaloneComponents = new Set(['auth.login', 'auth.resetPassword'])
const stableViewComponents = new Set(['home.index', 'category.index', 'search.index'])
const isStandalone = computed(() => standaloneComponents.has(props.page.payload.component))
const hasRail = computed(() => props.page.payload.component === 'article.detail')
const pageViewKey = computed(() => (
  stableViewComponents.has(props.page.payload.component)
    ? props.page.payload.component
    : props.page.payload.url
))
</script>

<template>
  <template v-if="isStandalone">
    <component
      :is="page.component"
      :key="page.payload.url"
      :layout="page.payload.layout"
      :props="page.payload.props"
    />
  </template>
  <template v-else>
    <AppShell
      :layout="page.payload.layout"
      :header-title="shellState.headerTitle"
      :header-tags="shellState.headerTags"
      :show-header-title="shellState.showHeaderTitle"
      :rail="hasRail"
    >
      <Transition name="gf-page">
        <component
          :is="page.component"
          :key="pageViewKey"
          :layout="page.payload.layout"
          :props="page.payload.props"
        />
      </Transition>
    </AppShell>
  </template>
</template>
