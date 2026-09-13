<script setup lang="ts">
import type { AcceptableValue } from 'reka-ui'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

type SelectOption = {
  value: string
  label: string
}

const props = defineProps<{
  modelValue: string
  options: SelectOption[]
  placeholder?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

function updateValue(value: AcceptableValue) {
  if (typeof value === 'string') emit('update:modelValue', value)
}
</script>

<template>
  <Select :model-value="modelValue" @update:model-value="updateValue">
    <SelectTrigger
      class="h-10 w-full rounded-[var(--gf-radius-field)] border-line bg-base-100 text-left text-base-content shadow-none focus-visible:border-primary focus-visible:ring-4 focus-visible:ring-primary/20"
    >
      <SelectValue :placeholder="placeholder" />
    </SelectTrigger>
    <SelectContent
      class="rounded-[var(--gf-radius-box)] border-line bg-base-100 text-base-content shadow-[0_18px_40px_-24px_rgb(15_23_42_/_calc(var(--gf-depth)*0.45))]"
    >
      <SelectItem
        v-for="option in options"
        :key="option.value"
        :value="option.value"
        class="h-9 rounded-md px-2.5 font-medium focus:bg-base-200 focus:text-base-content data-[state=checked]:bg-primary/10 data-[state=checked]:text-primary"
      >
        {{ option.label }}
      </SelectItem>
    </SelectContent>
  </Select>
</template>
