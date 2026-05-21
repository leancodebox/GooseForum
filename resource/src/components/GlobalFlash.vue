<script setup lang="ts">
import { computed } from 'vue'
import { X } from '@lucide/vue'
import { dismiss, useFlashMessages, type FlashMessageType } from '@/runtime/flash-message'

const { messages } = useFlashMessages()

const visibleMessages = computed(() => messages.value)

function dotClass(type: FlashMessageType) {
  switch (type) {
    case 'success':
      return 'bg-emerald-500'
    case 'warning':
      return 'bg-amber-500'
    case 'error':
      return 'bg-red-500'
    default:
      return 'bg-blue-500'
  }
}
</script>

<template>
  <TransitionGroup
    name="gf-flash"
    tag="div"
    class="fixed left-1/2 top-[4.75rem] z-[120] flex w-[min(92vw,380px)] -translate-x-1/2 flex-col gap-2 sm:left-auto sm:right-6 sm:translate-x-0"
    aria-live="polite"
    aria-atomic="true"
  >
    <div
      v-for="item in visibleMessages"
      :key="item.id"
      class="flex items-start gap-2.5 rounded-md border border-gray-200 bg-white/95 px-3 py-2.5 text-sm text-gray-800 shadow-[0_10px_28px_-22px_rgba(15,23,42,0.55)] ring-1 ring-black/[0.02] backdrop-blur"
      role="status"
    >
      <span
        class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full"
        :class="dotClass(item.type)"
        aria-hidden="true"
      />
      <p class="min-w-0 flex-1 leading-5">{{ item.message }}</p>
      <button
        type="button"
        class="-mr-1 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-md text-gray-400 hover:bg-gray-100 hover:text-gray-700"
        aria-label="关闭提示"
        @click="dismiss(item.id)"
      >
        <X class="h-3.5 w-3.5" />
      </button>
    </div>
  </TransitionGroup>
</template>
