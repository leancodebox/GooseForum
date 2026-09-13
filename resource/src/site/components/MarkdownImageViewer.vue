<script setup lang="ts">
import { computed, ref } from 'vue'
import { ChevronLeft, ChevronRight, Maximize2, Minimize2, X } from '@lucide/vue'
import { Dialog, DialogContent, DialogDescription, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useI18n } from 'vue-i18n'

interface MarkdownPreviewImage {
  src: string
  alt: string
}

const { t } = useI18n()
const images = ref<MarkdownPreviewImage[]>([])
const currentIndex = ref(0)
const actualSize = ref(false)
const viewerOpen = computed(() => images.value.length > 0)
const currentImage = computed(() => images.value[currentIndex.value])
const hasMultipleImages = computed(() => images.value.length > 1)

function open(nextImages: MarkdownPreviewImage[], index: number) {
  const normalizedImages = nextImages.filter((image) => image.src)
  if (!normalizedImages.length) return

  images.value = normalizedImages
  currentIndex.value = Math.max(0, Math.min(index, normalizedImages.length - 1))
  actualSize.value = false
}

function close() {
  if (!viewerOpen.value) return
  images.value = []
  currentIndex.value = 0
  actualSize.value = false
}

function showPrevious() {
  if (!hasMultipleImages.value) return
  currentIndex.value = currentIndex.value <= 0 ? images.value.length - 1 : currentIndex.value - 1
  actualSize.value = false
}

function showNext() {
  if (!hasMultipleImages.value) return
  currentIndex.value = currentIndex.value >= images.value.length - 1 ? 0 : currentIndex.value + 1
  actualSize.value = false
}

function toggleActualSize() {
  actualSize.value = !actualSize.value
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'ArrowLeft') {
    event.preventDefault()
    showPrevious()
    return
  }
  if (event.key === 'ArrowRight') {
    event.preventDefault()
    showNext()
    return
  }
  if (event.key === '0') {
    event.preventDefault()
    toggleActualSize()
  }
}

function onOpenChange(open: boolean) {
  if (!open) close()
}

defineExpose({
  open,
  close,
})
</script>

<template>
  <Dialog :open="viewerOpen" @update:open="onOpenChange">
    <DialogContent
      class="h-dvh max-h-none w-full max-w-none border-0 bg-transparent p-0 shadow-none"
      :show-close-button="false"
      @keydown="handleKeydown"
    >
      <DialogTitle v-if="currentImage" class="sr-only">{{ currentImage.alt || t('common.preview') }}</DialogTitle>
      <DialogDescription class="sr-only">{{ t('common.preview') }}</DialogDescription>

      <div v-if="currentImage" class="relative h-full w-full">
        <div
          v-if="hasMultipleImages"
          class="gf-markdown-image-viewer-count absolute left-3 top-3 z-10 rounded-full px-3 py-2 text-xs font-semibold tabular-nums sm:left-5 sm:top-5"
        >
          {{ currentIndex + 1 }} / {{ images.length }}
        </div>

        <div class="absolute right-3 top-3 z-10 flex items-center gap-2 sm:right-5 sm:top-5">
          <Button
            variant="muted"
            class="gf-markdown-image-viewer-button inline-flex h-10 w-10 items-center justify-center rounded-full transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
            :aria-label="actualSize ? t('common.preview') : t('image.originalSize')"
            :title="actualSize ? t('common.preview') : t('image.originalSize')"
            @click.stop="toggleActualSize"
          >
            <Minimize2 v-if="actualSize" class="h-4 w-4" />
            <Maximize2 v-else class="h-4 w-4" />
            <span class="sr-only">{{ actualSize ? t('common.preview') : t('image.originalSize') }}</span>
          </Button>

          <Button
            variant="muted"
            class="gf-markdown-image-viewer-button inline-flex h-10 w-10 items-center justify-center rounded-full transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
            :aria-label="t('common.close')"
            :title="t('common.close')"
            @click="close"
          >
            <X class="h-4 w-4" />
            <span class="sr-only">{{ t('common.close') }}</span>
          </Button>
        </div>

        <Button
          v-if="hasMultipleImages"
          variant="muted"
          class="gf-markdown-image-viewer-button absolute left-3 top-1/2 z-10 inline-flex h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary sm:left-5"
          :aria-label="t('common.previousPage')"
          :title="t('common.previousPage')"
          @click.stop="showPrevious"
        >
          <ChevronLeft class="h-5 w-5" />
          <span class="sr-only">{{ t('common.previousPage') }}</span>
        </Button>

        <div class="gf-markdown-image-viewer-stage grid h-full w-full place-items-center overflow-hidden p-0">
          <img
            :key="currentImage.src"
            :src="currentImage.src"
            :alt="currentImage.alt"
            class="gf-markdown-image-viewer-image rounded-md object-contain"
            :class="actualSize ? 'max-h-[calc(100dvh-3rem)] max-w-[calc(100vw-1.5rem)] cursor-zoom-out sm:max-w-[calc(100vw-3rem)]' : 'max-h-[calc(100dvh-6rem)] max-w-[calc(100vw-1.5rem)] cursor-zoom-in sm:max-w-[calc(100vw-7rem)]'"
            decoding="async"
            @click.stop="toggleActualSize"
          >
        </div>

        <Button
          v-if="hasMultipleImages"
          variant="muted"
          class="gf-markdown-image-viewer-button absolute right-3 top-1/2 z-10 inline-flex h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary sm:right-5"
          :aria-label="t('common.nextPage')"
          :title="t('common.nextPage')"
          @click.stop="showNext"
        >
          <ChevronRight class="h-5 w-5" />
          <span class="sr-only">{{ t('common.nextPage') }}</span>
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>

<style scoped>
.gf-markdown-image-viewer-button {
  border: var(--gf-border) solid color-mix(in oklch, var(--gf-color-line) 76%, transparent);
  background: color-mix(in oklch, var(--gf-color-base-100) 86%, transparent);
  color: color-mix(in oklch, var(--gf-color-base-content) 78%, transparent);
  box-shadow:
    0 16px 36px -26px color-mix(in oklch, var(--gf-color-neutral) 70%, transparent),
    0 2px 10px -8px color-mix(in oklch, var(--gf-color-neutral) 55%, transparent);
}

.gf-markdown-image-viewer-button:hover {
  background: color-mix(in oklch, var(--gf-color-base-200) 90%, var(--gf-color-base-100));
  color: var(--gf-color-base-content);
}

.gf-markdown-image-viewer-count {
  border: var(--gf-border) solid color-mix(in oklch, var(--gf-color-line) 70%, transparent);
  background: color-mix(in oklch, var(--gf-color-base-100) 82%, transparent);
  color: color-mix(in oklch, var(--gf-color-base-content) 72%, transparent);
  box-shadow: 0 12px 30px -24px color-mix(in oklch, var(--gf-color-neutral) 65%, transparent);
  backdrop-filter: blur(8px);
}

.gf-markdown-image-viewer-image {
  box-shadow:
    0 28px 80px -36px color-mix(in oklch, var(--gf-color-neutral) 82%, transparent),
    0 10px 32px -24px color-mix(in oklch, var(--gf-color-neutral) 68%, transparent);
}
</style>
