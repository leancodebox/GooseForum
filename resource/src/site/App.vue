<script setup lang="ts">
import { RouterView } from 'vue-router'
import type { PreparedPage } from '@/runtime/router'

const props = defineProps<{
  page: PreparedPage
}>()
</script>

<template>
  <RouterView v-slot="{ Component }">
    <component :is="Component" :page="props.page" />
  </RouterView>
</template>
