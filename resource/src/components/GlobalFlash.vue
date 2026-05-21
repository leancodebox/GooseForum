<script setup lang="ts">
import { computed } from 'vue'
import { X } from '@lucide/vue'
import FlashSpriteIcon from '@/components/FlashSpriteIcon.vue'
import { dismiss, useFlashMessages, type FlashMessageType } from '@/runtime/flash-message'

const { messages } = useFlashMessages()

const visibleMessages = computed(() => messages.value)

function accentClass(type: FlashMessageType) {
  switch (type) {
    case 'success':
      return 'from-emerald-500/90 via-emerald-400/40'
    case 'warning':
      return 'from-amber-500/90 via-amber-400/40'
    case 'error':
      return 'from-red-500/90 via-red-400/40'
    default:
      return 'from-blue-500/90 via-blue-400/40'
  }
}

function labelFor(type: FlashMessageType) {
  switch (type) {
    case 'success':
      return '已完成'
    case 'warning':
      return '请注意'
    case 'error':
      return '出错了'
    default:
      return '提示'
  }
}
</script>

<template>
  <div class="pointer-events-none sticky top-16 z-[120] h-0 px-3 sm:px-5 lg:px-8">
    <TransitionGroup
      name="gf-flash"
      tag="div"
      class="mx-auto flex w-full max-w-[1600px] flex-col items-start gap-2 pt-3"
      aria-live="polite"
      aria-atomic="true"
    >
      <div
        v-for="item in visibleMessages"
        :key="item.id"
        class="pointer-events-auto relative flex min-h-[86px] w-full max-w-[380px] items-start gap-3 overflow-hidden rounded-xl border border-gray-200 bg-white/95 px-3.5 py-3.5 pr-2.5 text-sm text-gray-800 shadow-[0_18px_55px_-34px_rgba(15,23,42,0.7)] ring-1 ring-gray-950/5 backdrop-blur"
        role="status"
      >
        <span
          class="absolute inset-y-0 left-0 w-1 bg-gradient-to-b to-transparent"
          :class="accentClass(item.type)"
          aria-hidden="true"
        />
        <div class="mt-0.5 flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-gray-50 ring-1 ring-gray-200/80">
          <FlashSpriteIcon :type="item.type" class="h-9 w-9" />
        </div>
        <div class="min-w-0 flex-1">
          <div class="mb-1 text-[11px] font-bold text-gray-400">{{ labelFor(item.type) }}</div>
          <p class="leading-5 text-gray-800">{{ item.message }}</p>
        </div>
        <button
          type="button"
          class="-mr-1 inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-gray-400 transition hover:bg-gray-100 hover:text-gray-700"
          aria-label="关闭提示"
          @click="dismiss(item.id)"
        >
          <X class="h-3.5 w-3.5" />
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>
