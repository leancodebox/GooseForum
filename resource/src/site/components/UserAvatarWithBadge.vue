<script setup lang="ts">
import UserAvatar from '@/site/components/UserAvatar.vue'
import { badgeClass, badgeIconURL, badgeTooltip } from '@/site/utils/badge-style'
import type { UserBadgePayload } from '@/types/payload'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<{
  src: string
  alt: string
  badge?: UserBadgePayload | null
  size?: 'medium' | 'large'
  imgClass?: string
}>(), {
  size: 'medium',
  imgClass: '',
  badge: null,
})

</script>

<template>
  <span v-bind="$attrs" class="relative inline-block shrink-0">
    <UserAvatar :src="src" :alt="alt" :size="size" class="h-full w-full object-cover" :class="imgClass" />
    <span
      v-if="badge"
      class="absolute -bottom-1 -right-1 z-10 flex h-[38%] min-h-4 w-[38%] min-w-4 items-center justify-center rounded-full p-0.5 shadow-sm ring-1 ring-inset"
      :class="badgeClass(badge.color, badge.level)"
      :title="badgeTooltip(badge)"
    >
      <img :src="badgeIconURL(badge)" :alt="badge.name" class="h-full w-full object-contain" />
    </span>
  </span>
</template>
