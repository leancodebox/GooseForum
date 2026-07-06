<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref } from 'vue'
import { ChevronLeft, ChevronRight, Maximize2, Minimize2, X } from '@lucide/vue'
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
let lastBodyOverflow = ''
let bodyScrollLocked = false

function open(nextImages: MarkdownPreviewImage[], index: number) {
  const normalizedImages = nextImages.filter((image) => image.src)
  if (!normalizedImages.length) return

  images.value = normalizedImages
  currentIndex.value = Math.max(0, Math.min(index, normalizedImages.length - 1))
  actualSize.value = false
  lockBodyScroll()
  void nextTick(() => {
    window.addEventListener('keydown', handleKeydown)
  })
}

function close() {
  if (!viewerOpen.value) return
  images.value = []
  currentIndex.value = 0
  actualSize.value = false
  window.removeEventListener('keydown', handleKeydown)
  unlockBodyScroll()
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
  if (event.key === 'Escape') {
    event.preventDefault()
    close()
    return
  }
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

function lockBodyScroll() {
  if (typeof document === 'undefined') return
  if (document.body.style.overflow === 'hidden') return
  lastBodyOverflow = document.body.style.overflow
  document.body.style.overflow = 'hidden'
  bodyScrollLocked = true
}

function unlockBodyScroll() {
  if (typeof document === 'undefined') return
  if (!bodyScrollLocked) return
  document.body.style.overflow = lastBodyOverflow
  lastBodyOverflow = ''
  bodyScrollLocked = false
}

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeydown)
  unlockBodyScroll()
})

defineExpose({
  open,
  close,
})
</script>

<template>
  <Teleport to="body">
    <Transition name="gf-modal">
      <div
        v-if="viewerOpen && currentImage"
        class="gf-markdown-image-viewer fixed inset-0 z-[120] flex items-center justify-center px-3 py-4 backdrop-blur-sm sm:px-6"
        role="dialog"
        aria-modal="true"
        :aria-label="currentImage.alt || t('common.preview')"
        @click.self="close"
      >
        <div class="absolute right-3 top-3 z-10 flex items-center gap-2 sm:right-5 sm:top-5">
          <button
            type="button"
            class="gf-markdown-image-viewer-button inline-flex h-10 w-10 items-center justify-center rounded-full transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
            :aria-label="actualSize ? t('common.preview') : t('image.originalSize')"
            :title="actualSize ? t('common.preview') : t('image.originalSize')"
            @click.stop="toggleActualSize"
          >
            <Minimize2 v-if="actualSize" class="h-4 w-4" />
            <Maximize2 v-else class="h-4 w-4" />
            <span class="sr-only">{{ actualSize ? t('common.preview') : t('image.originalSize') }}</span>
          </button>

          <button
            type="button"
            class="gf-markdown-image-viewer-button inline-flex h-10 w-10 items-center justify-center rounded-full transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
            :aria-label="t('common.close')"
            :title="t('common.close')"
            @click="close"
          >
            <X class="h-4 w-4" />
            <span class="sr-only">{{ t('common.close') }}</span>
          </button>
        </div>

        <button
          v-if="hasMultipleImages"
          type="button"
          class="gf-markdown-image-viewer-button absolute left-3 top-1/2 z-10 inline-flex h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary sm:left-5"
          :aria-label="t('common.previousPage')"
          :title="t('common.previousPage')"
          @click.stop="showPrevious"
        >
          <ChevronLeft class="h-5 w-5" />
          <span class="sr-only">{{ t('common.previousPage') }}</span>
        </button>

        <div class="gf-markdown-image-viewer-stage flex h-full w-full items-center justify-center overflow-auto" :class="actualSize ? 'p-4 sm:p-8' : 'p-0'">
          <img
            :key="currentImage.src"
            :src="currentImage.src"
            :alt="currentImage.alt"
            class="gf-markdown-image-viewer-image rounded-md object-contain"
            :class="actualSize ? 'max-h-none max-w-none cursor-zoom-out' : 'max-h-[calc(100dvh-6rem)] max-w-[calc(100vw-1.5rem)] cursor-zoom-in sm:max-w-[calc(100vw-7rem)]'"
            decoding="async"
            @click.stop="toggleActualSize"
          >
        </div>

        <button
          v-if="hasMultipleImages"
          type="button"
          class="gf-markdown-image-viewer-button absolute right-3 top-1/2 z-10 inline-flex h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary sm:right-5"
          :aria-label="t('common.nextPage')"
          :title="t('common.nextPage')"
          @click.stop="showNext"
        >
          <ChevronRight class="h-5 w-5" />
          <span class="sr-only">{{ t('common.nextPage') }}</span>
        </button>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.gf-markdown-image-viewer {
  --gf-image-viewer-backdrop: color-mix(in oklch, var(--gf-color-base-content) 62%, transparent);
  --gf-image-viewer-backdrop-glow: color-mix(in oklch, var(--gf-color-base-100) 16%, transparent);
  background:
    radial-gradient(circle at top, var(--gf-image-viewer-backdrop-glow), transparent 42%),
    var(--gf-image-viewer-backdrop);
}

:global([data-theme="gf-dark"]) .gf-markdown-image-viewer {
  --gf-image-viewer-backdrop: color-mix(in oklch, var(--gf-color-base-200) 82%, transparent);
  --gf-image-viewer-backdrop-glow: color-mix(in oklch, var(--gf-color-base-content) 8%, transparent);
}

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

.gf-markdown-image-viewer-image {
  box-shadow:
    0 28px 80px -36px color-mix(in oklch, var(--gf-color-neutral) 82%, transparent),
    0 10px 32px -24px color-mix(in oklch, var(--gf-color-neutral) 68%, transparent);
}
</style>
