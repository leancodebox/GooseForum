<script setup lang="ts">
import { computed } from 'vue'
import type { ButtonVariants } from '@/admin/components/ui/button'
import { Button } from '@/admin/components/ui/button'

const props = withDefaults(defineProps<{
  compact?: boolean
  disabled?: boolean
  title?: string
  tone?: 'default' | 'danger' | 'primary' | 'success'
}>(), {
  compact: false,
  disabled: false,
  title: '',
  tone: 'default',
})

const size = computed<ButtonVariants['size']>(() => props.compact ? 'icon-sm' : 'sm')
const toneClass = computed(() => {
  if (props.tone === 'danger') return props.compact ? 'text-destructive hover:text-destructive' : 'border-destructive/30 text-destructive hover:bg-destructive/5 hover:text-destructive'
  if (props.tone === 'primary') return 'text-primary hover:text-primary'
  if (props.tone === 'success') return 'text-emerald-600 hover:text-emerald-700'
  return props.compact ? 'text-muted-foreground hover:text-foreground' : 'text-foreground'
})
</script>

<template>
  <Button
    :variant="compact ? 'ghost' : 'outline'"
    :size="size"
    type="button"
    :class="[compact ? '' : 'h-8 text-xs', toneClass]"
    :disabled="disabled"
    :title="title"
  >
    <slot />
  </Button>
</template>
