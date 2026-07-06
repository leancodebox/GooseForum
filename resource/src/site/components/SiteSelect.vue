<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { Check, ChevronDown } from '@lucide/vue'

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

const open = ref(false)
const root = ref<HTMLElement | null>(null)

const selectedOption = computed(() => props.options.find(option => option.value === props.modelValue))
const triggerLabel = computed(() => selectedOption.value?.label || props.placeholder || '')

function selectOption(value: string) {
  emit('update:modelValue', value)
  open.value = false
}

function handleDocumentPointerDown(event: PointerEvent) {
  const target = event.target
  if (target instanceof Node && root.value?.contains(target)) return
  open.value = false
}

function handleTriggerKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    open.value = false
    return
  }
  if (event.key === 'ArrowDown' || event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    open.value = true
  }
}

onMounted(() => {
  document.addEventListener('pointerdown', handleDocumentPointerDown)
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', handleDocumentPointerDown)
})
</script>

<template>
  <div ref="root" class="relative">
    <button
      type="button"
      class="gf-input flex w-full items-center justify-between gap-2 text-left"
      :aria-expanded="open"
      @click="open = !open"
      @keydown="handleTriggerKeydown"
    >
      <span class="min-w-0 truncate" :class="selectedOption ? 'text-base-content' : 'text-base-content/45'">
        {{ triggerLabel }}
      </span>
      <ChevronDown class="h-4 w-4 shrink-0 text-base-content/45 transition-transform" :class="{ 'rotate-180': open }" />
    </button>

    <Transition name="gf-menu">
      <div v-if="open" class="gf-menu-surface absolute left-0 right-0 top-[calc(100%+0.375rem)] z-30 overflow-hidden p-1">
        <button
          v-for="option in options"
          :key="option.value"
          type="button"
          class="flex h-9 w-full items-center gap-2 rounded-md px-2.5 text-left text-sm font-medium text-base-content hover:bg-base-200"
          :class="option.value === modelValue ? 'bg-primary/10 text-primary' : ''"
          @click="selectOption(option.value)"
        >
          <span class="min-w-0 flex-1 truncate">{{ option.label }}</span>
          <Check v-if="option.value === modelValue" class="h-4 w-4 shrink-0" />
        </button>
      </div>
    </Transition>
  </div>
</template>
