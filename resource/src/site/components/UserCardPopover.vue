<script setup lang="ts">
import { defineAsyncComponent, ref } from 'vue'
import { Popover, PopoverTrigger } from '@/components/ui/popover'
import { chooseUserCardSide, type UserCardSide, type UserCardTarget } from '@/runtime/user-card'

defineProps<{ user: UserCardTarget }>()

const UserCard = defineAsyncComponent(() => import('./UserCard.vue'))
const open = ref(false)
const hasOpened = ref(false)
const side = ref<UserCardSide>('bottom')
let ignoreNextOpen = false

function handleTriggerClick(event: MouseEvent) {
  if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) {
    ignoreNextOpen = true
    return
  }
  event.preventDefault()

  const trigger = event.currentTarget
  if (trigger instanceof HTMLElement) {
    side.value = chooseUserCardSide(trigger.getBoundingClientRect(), window.innerHeight)
  }
}

function handleOpenChange(next: boolean) {
  if (next && ignoreNextOpen) {
    ignoreNextOpen = false
    return
  }
  ignoreNextOpen = false
  open.value = next
  if (next) hasOpened.value = true
}
</script>

<template>
  <Popover :open="open" @update:open="handleOpenChange">
    <PopoverTrigger as-child @click.capture="handleTriggerClick">
      <slot />
    </PopoverTrigger>
    <UserCard v-if="hasOpened" :user="user" :side="side" />
  </Popover>
</template>
