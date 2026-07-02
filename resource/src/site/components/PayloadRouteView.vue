<script setup lang="ts">
import { computed, KeepAlive } from 'vue'
import type { PreparedPage } from '@/runtime/router'
import AppShell from '@/site/components/AppShell.vue'
import { useShellState } from '@/runtime/shell-state'

const props = defineProps<{
  page: PreparedPage
}>()

const shellState = useShellState()
const standaloneComponents = new Set(['auth.login', 'auth.resetPassword'])
const keepAliveComponents = new Set(['home.index', 'category.index', 'search.index'])
const isStandalone = computed(() => standaloneComponents.has(props.page.payload.component))
const hasRail = computed(() => props.page.payload.component === 'article.detail')
const shouldKeepAlive = computed(() => keepAliveComponents.has(props.page.payload.component))
const pageViewKey = computed(() => {
  if (props.page.payload.component === 'user.profile') {
    const user = (props.page.payload.props as { user?: { userId?: number | string } })?.user
    return `user.profile:${user?.userId || props.page.payload.url}`
  }
  return props.page.payload.url
})
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
      <KeepAlive :max="10">
        <component
          v-if="shouldKeepAlive"
          :is="page.component"
          :key="pageViewKey"
          :layout="page.payload.layout"
          :props="page.payload.props"
          :page-url="page.payload.url"
        />
      </KeepAlive>
      <Transition v-if="!shouldKeepAlive" name="gf-page">
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
