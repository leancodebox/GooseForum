<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { formatNumber } from '@/runtime/format'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  current: number
  max: number
  startLabel: string
  endLabel: string
  currentLabel?: string
  busy?: boolean
}>()

const emit = defineEmits<{
  earliest: []
  latest: []
  select: [replyNo: number]
}>()

const { t } = useI18n()
const railEl = ref<HTMLElement | null>(null)
const dragging = ref(false)
const previewNo = ref<number | null>(null)
const thumbHeight = 32
const thumbNo = computed(() => clampReplyNo(previewNo.value ?? props.current ?? 1))
const displayNo = computed(() => thumbNo.value)
const progress = computed(() => {
  if (props.max <= 1) return 0
  return ((thumbNo.value - 1) / (props.max - 1)) * 100
})
const thumbStyle = computed(() => ({
  top: `clamp(0px, calc(${progress.value}% - ${thumbHeight / 2}px), calc(100% - ${thumbHeight}px))`,
  height: `${thumbHeight}px`,
}))
const indicatorStyle = computed(() => ({
  top: `clamp(0px, calc(${progress.value}% - 18px), calc(100% - 36px))`,
}))

onBeforeUnmount(() => {
  removePointerListeners()
})

function clampReplyNo(replyNo: number) {
  return Math.min(Math.max(1, props.max || 1), Math.max(1, Math.round(replyNo)))
}

function replyNoFromPointer(event: PointerEvent) {
  const element = railEl.value
  if (!element) return displayNo.value
  const rect = element.getBoundingClientRect()
  const ratio = Math.min(1, Math.max(0, (event.clientY - rect.top) / Math.max(1, rect.height)))
  return clampReplyNo(1 + ratio * Math.max(0, props.max - 1))
}

function startDrag(event: PointerEvent) {
  if (!props.max || props.busy) return
  event.preventDefault()
  dragging.value = true
  previewNo.value = replyNoFromPointer(event)
  window.addEventListener('pointermove', handlePointerMove)
  window.addEventListener('pointerup', handlePointerUp)
}

function handlePointerMove(event: PointerEvent) {
  if (!dragging.value) return
  previewNo.value = replyNoFromPointer(event)
}

function handlePointerUp(event: PointerEvent) {
  if (!dragging.value) return
  previewNo.value = replyNoFromPointer(event)
  const target = previewNo.value
  dragging.value = false
  previewNo.value = null
  removePointerListeners()
  if (target && target !== props.current) {
    emit('select', target)
  }
}

function selectReplyNo(replyNo: number) {
  const target = clampReplyNo(replyNo)
  if (target !== props.current) {
    emit('select', target)
  }
}

function removePointerListeners() {
  window.removeEventListener('pointermove', handlePointerMove)
  window.removeEventListener('pointerup', handlePointerUp)
}
</script>

<template>
  <div class="px-3 py-2">
    <button
      type="button"
      class="mb-3 block max-w-full truncate text-left text-base font-semibold leading-tight text-gray-400 transition hover:text-gray-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-gray-400 focus-visible:ring-offset-2"
      :title="t('article.earliestContent')"
      @click="emit('earliest')"
    >
      {{ startLabel }}
    </button>

    <div
      class="grid cursor-ns-resize touch-none select-none grid-cols-[24px_minmax(0,1fr)] gap-3"
      @pointerdown="startDrag"
    >
      <div
        ref="railEl"
        class="relative h-48 rounded-full focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-gray-400 focus-visible:ring-offset-2"
        role="slider"
        tabindex="0"
        :aria-label="t('article.replyPosition')"
        :aria-valuemin="1"
        :aria-valuemax="max"
        :aria-valuenow="displayNo"
        @keydown.up.prevent="selectReplyNo(displayNo - 1)"
        @keydown.down.prevent="selectReplyNo(displayNo + 1)"
        @keydown.home.prevent="selectReplyNo(1)"
        @keydown.end.prevent="emit('latest')"
      >
        <div class="absolute left-1/2 top-0 h-full w-px -translate-x-1/2 bg-gray-200" />
        <div
          class="absolute left-1/2 w-1.5 -translate-x-1/2 rounded-full bg-emerald-500 transition-[top,box-shadow]"
          :class="{ 'transition-none shadow-[0_0_0_4px_rgba(16,185,129,0.16)]': dragging, 'opacity-70': busy }"
          :style="thumbStyle"
        />
      </div>

      <div class="relative min-w-0">
        <div class="absolute left-0 min-w-0" :style="indicatorStyle">
          <div class="whitespace-nowrap text-base font-black leading-none tabular-nums text-gray-950">
            {{ displayNo }} / {{ formatNumber(max) }}
          </div>
          <div v-if="currentLabel && !dragging && !busy" class="mt-2 truncate text-sm font-semibold leading-tight text-gray-400">
            {{ currentLabel }}
          </div>
        </div>
      </div>
    </div>

    <button
      type="button"
      class="mt-3 block max-w-full truncate text-left text-base font-semibold leading-tight text-gray-400 transition hover:text-gray-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-gray-400 focus-visible:ring-offset-2"
      :title="t('article.latestReply')"
      @click="emit('latest')"
    >
      {{ endLabel }}
    </button>
  </div>
</template>
