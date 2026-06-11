<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { CircleCheck, CircleX, Info, TriangleAlert, X } from '@lucide/vue'
import { dismiss, useFlashMessages, type FlashMessageType } from '@/runtime/flash-message'

const { messages } = useFlashMessages()
const { t } = useI18n()

const visibleMessages = computed(() => messages.value)

function accentClass(type: FlashMessageType) {
  switch (type) {
    case 'success':
      return 'bg-success/100'
    case 'warning':
      return 'bg-warning/100'
    case 'error':
      return 'bg-error/100'
    default:
      return 'bg-primary'
  }
}

function iconFor(type: FlashMessageType) {
  switch (type) {
    case 'success':
      return CircleCheck
    case 'warning':
      return TriangleAlert
    case 'error':
      return CircleX
    default:
      return Info
  }
}

function iconClass(type: FlashMessageType) {
  switch (type) {
    case 'success':
      return 'text-success'
    case 'warning':
      return 'text-warning'
    case 'error':
      return 'text-error'
    default:
      return 'text-primary'
  }
}

function labelFor(type: FlashMessageType) {
  switch (type) {
    case 'success':
      return t('flash.success')
    case 'warning':
      return t('flash.warning')
    case 'error':
      return t('flash.error')
    default:
      return t('flash.info')
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
        class="gf-alert pointer-events-auto relative flex min-h-[72px] w-full max-w-[380px] items-start gap-3 overflow-hidden px-3.5 py-3 pr-2.5 text-sm backdrop-blur"
        role="status"
      >
        <span
          class="absolute inset-y-0 left-0 w-1"
          :class="accentClass(item.type)"
          aria-hidden="true"
        />
        <component :is="iconFor(item.type)" class="mt-0.5 h-5 w-5 shrink-0" :class="iconClass(item.type)" aria-hidden="true" />
        <div class="min-w-0 flex-1">
          <div class="mb-1 text-[11px] font-bold text-base-content/55">{{ labelFor(item.type) }}</div>
          <p class="leading-5 text-base-content">{{ item.message }}</p>
        </div>
        <button
          type="button"
          class="-mr-1 inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-icon-muted transition hover:bg-base-300 hover:text-base-content"
          :aria-label="t('flash.close')"
          @click="dismiss(item.id)"
        >
          <X class="h-3.5 w-3.5" />
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>
