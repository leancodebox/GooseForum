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
  progressCurrent?: number
  progressEnd?: number
  progressStart?: number
}>()

const emit = defineEmits<{
  earliest: []
  latest: []
  select: [postNo: number]
}>()

const { t } = useI18n()
const railEl = ref<HTMLElement | null>(null)
const dragging = ref(false)
const previewNo = ref<number | null>(null)
const railHeight = 192
const thumbHeight = 28
const thumbNo = computed(() => clampPostNo(previewNo.value ?? props.current ?? 1))
const displayNo = computed(() => thumbNo.value)
const normalizedCurrent = computed(() => dragging.value ? progressForPostNo(thumbNo.value) : clampProgress(props.progressCurrent ?? progressForPostNo(props.current)))
const thumbStyle = computed(() => ({
  top: `${thumbTopPx.value}px`,
  height: `${thumbHeight}px`,
}))
const thumbCenterPx = computed(() => thumbTopPx.value + thumbHeight / 2)
const indicatorStyle = computed(() => ({
  top: `${Math.min(railHeight - 36, Math.max(0, thumbCenterPx.value - 18))}px`,
}))
const thumbTopPx = computed(() => {
  return Math.min(railHeight - thumbHeight, Math.max(0, normalizedCurrent.value * railHeight - thumbHeight / 2))
})

onBeforeUnmount(() => {
  removePointerListeners()
})

function clampPostNo(postNo: number) {
  return Math.min(Math.max(1, props.max || 1), Math.max(1, Math.round(postNo)))
}

function clampProgress(value: number) {
  if (!Number.isFinite(value)) return 0
  return Math.min(1, Math.max(0, value))
}

function progressForPostNo(postNo: number) {
  if (props.max <= 1) return 0
  return clampProgress((postNo - 1) / (props.max - 1))
}

function postNoFromPointer(event: PointerEvent) {
  const element = railEl.value
  if (!element) return displayNo.value
  const rect = element.getBoundingClientRect()
  const ratio = Math.min(1, Math.max(0, (event.clientY - rect.top) / Math.max(1, rect.height)))
  return clampPostNo(1 + ratio * Math.max(0, props.max - 1))
}

function startDrag(event: PointerEvent) {
  if (!props.max || props.busy) return
  event.preventDefault()
  dragging.value = true
  previewNo.value = postNoFromPointer(event)
  window.addEventListener('pointermove', handlePointerMove)
  window.addEventListener('pointerup', handlePointerUp)
}

function handlePointerMove(event: PointerEvent) {
  if (!dragging.value) return
  previewNo.value = postNoFromPointer(event)
}

function handlePointerUp(event: PointerEvent) {
  if (!dragging.value) return
  previewNo.value = postNoFromPointer(event)
  const target = previewNo.value
  dragging.value = false
  previewNo.value = null
  removePointerListeners()
  if (target && target !== props.current) {
    emit('select', target)
  }
}

function selectPostNo(postNo: number) {
  const target = clampPostNo(postNo)
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
      class="mb-3 block max-w-full truncate text-left text-base font-semibold leading-tight text-base-content/55 transition hover:text-base-content/75 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-base-content/35 focus-visible:ring-offset-2"
      :title="t('topic.earliestContent')"
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
        class="relative h-48 rounded-full focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-base-content/35 focus-visible:ring-offset-2"
        role="slider"
        tabindex="0"
        :aria-label="t('topic.replyPosition')"
        :aria-valuemin="1"
        :aria-valuemax="max"
        :aria-valuenow="displayNo"
        @keydown.up.prevent="selectPostNo(thumbNo - 1)"
        @keydown.down.prevent="selectPostNo(thumbNo + 1)"
        @keydown.home.prevent="emit('earliest')"
        @keydown.end.prevent="emit('latest')"
      >
        <div class="absolute left-1/2 top-0 h-full w-px -translate-x-1/2 bg-line" />
        <div
          class="absolute left-1/2 w-1.5 -translate-x-1/2 rounded-full bg-success/100 transition-[box-shadow]"
          :class="{ 'transition-none shadow-[0_0_0_4px_rgba(16,185,129,0.16)]': dragging, 'opacity-70': busy }"
          :style="thumbStyle"
        />
      </div>

      <div class="relative min-w-0">
        <div class="absolute left-0 min-w-0" :style="indicatorStyle">
          <div class="whitespace-nowrap text-base font-black leading-none tabular-nums text-base-content">
            {{ `${displayNo} / ${formatNumber(max)}` }}
          </div>
          <div v-if="currentLabel && !dragging && !busy" class="mt-2 truncate text-sm font-semibold leading-tight text-base-content/55">
            {{ currentLabel }}
          </div>
        </div>
      </div>
    </div>

    <button
      type="button"
      class="mt-3 block max-w-full truncate text-left text-base font-semibold leading-tight text-base-content/55 transition hover:text-base-content/75 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-base-content/35 focus-visible:ring-offset-2"
      :title="t('topic.latestReply')"
      @click="emit('latest')"
    >
      {{ endLabel }}
    </button>
  </div>
</template>
