<script setup lang="ts">
import type { FlashMessageType } from '@/runtime/flash-message'

defineProps<{
  type: FlashMessageType
}>()

function accentClass(type: FlashMessageType) {
  switch (type) {
    case 'success':
      return 'fill-emerald-400'
    case 'warning':
      return 'fill-amber-400'
    case 'error':
      return 'fill-red-400'
    default:
      return 'fill-blue-400'
  }
}
</script>

<template>
  <svg
    viewBox="0 0 96 96"
    role="img"
    aria-hidden="true"
    class="h-full w-full"
  >
    <path
      d="M48 13.5c17.2 0 29.6 12.7 29.6 29.9v30.2c0 4.4-4.9 7-8.5 4.6l-4.3-2.9-5.2 4.2a6.1 6.1 0 0 1-7.6 0L48 76.2l-4 3.3a6.1 6.1 0 0 1-7.6 0l-5.2-4.2-4.3 2.9c-3.6 2.4-8.5-.2-8.5-4.6V43.4C18.4 26.2 30.8 13.5 48 13.5Z"
      fill="#F9FAFB"
      stroke="#111827"
      stroke-width="4"
      stroke-linejoin="round"
    />
    <path
      d="M25.4 53.5c-6.3-.9-11.6-3.8-15.5-8.5 5.2-3.2 10.8-3.6 16.2-1.2"
      fill="#F8FAFC"
      stroke="#111827"
      stroke-width="3.4"
      stroke-linecap="round"
      stroke-linejoin="round"
    />
    <path
      d="M70.6 53.5c6.3-.9 11.6-3.8 15.5-8.5-5.2-3.2-10.8-3.6-16.2-1.2"
      fill="#F8FAFC"
      stroke="#111827"
      stroke-width="3.4"
      stroke-linecap="round"
      stroke-linejoin="round"
    />
    <circle cx="37.6" cy="39.4" r="3.4" fill="#111827" />
    <circle cx="58.4" cy="39.4" r="3.4" fill="#111827" />
    <circle cx="38.7" cy="38.3" r="0.9" fill="#fff" />
    <circle cx="59.5" cy="38.3" r="0.9" fill="#fff" />
    <path
      d="M32.9 56.2c9.2 5.6 21 5.6 30.2 0l-2.4 9.9c-.7 2.9-3.3 5-6.3 5H41.6c-3 0-5.6-2.1-6.3-5l-2.4-9.9Z"
      :class="accentClass(type)"
      stroke="#111827"
      stroke-width="3.3"
      stroke-linejoin="round"
    />
    <path
      d="M37.2 82.7c-1.7 2.4-4 4.1-6.8 5.1M58.8 82.7c1.7 2.4 4 4.1 6.8 5.1"
      fill="none"
      stroke="#111827"
      stroke-width="3.5"
      stroke-linecap="round"
    />
  </svg>
</template>
