<script setup lang="ts">
import type { DropdownMenuItemProps } from 'reka-ui'
import type { HTMLAttributes } from 'vue'
import { reactiveOmit } from '@vueuse/core'
import { DropdownMenuItem, useForwardProps } from 'reka-ui'
import { cn } from '@/admin/utils/cn'

const props = withDefaults(defineProps<DropdownMenuItemProps & {
  class?: HTMLAttributes['class']
  variant?: 'default' | 'destructive'
}>(), {
  variant: 'default',
})
const delegatedProps = reactiveOmit(props, 'variant', 'class')
const forwardedProps = useForwardProps(delegatedProps)
</script>

<template>
  <DropdownMenuItem
    data-slot="dropdown-menu-item"
    :data-variant="variant"
    v-bind="forwardedProps"
    :class="cn('relative flex cursor-default select-none items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-hidden focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50 data-[variant=destructive]:text-destructive data-[variant=destructive]:focus:bg-destructive/10 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0', props.class)"
  >
    <slot />
  </DropdownMenuItem>
</template>
