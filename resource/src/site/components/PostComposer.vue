<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import type { Component } from 'vue'
import { Check, Image, Loader2, MessageSquare, Send, X } from '@lucide/vue'
import { uploadImage } from '@/runtime/api'
import { formatNumber } from '@/runtime/format'
import { processImageFile, validateImageFile } from '@/runtime/image'
import { markdownFromClipboard } from '@/runtime/rich-paste'
import type { PostPayload } from '@/types/payload'
import PostPositionRail from '@/site/components/PostPositionRail.vue'
import { useI18n } from 'vue-i18n'

type PostAction = {
  key: string
  icon: Component
  active: boolean
  acting: boolean
  fill?: boolean
  title: string
  activeClass: string
  onClick: () => void | Promise<void>
}

const props = defineProps<{
  actions: PostAction[]
  authenticated: boolean
  canPost: boolean
  currentLabel: string
  currentNo: number
  endLabel: string
  errorMessage: string
  hasRail: boolean
  maxNo: number
  mobileRailOpen: boolean
  mode?: 'create' | 'edit'
  open: boolean
  progressCurrent?: number
  progressEnd?: number
  progressStart?: number
  railBusy: boolean
  startLabel: string
  submitting: boolean
  successMessage: string
  target?: PostPayload
}>()

const emit = defineEmits<{
  clearTarget: []
  clearValidation: []
  earliest: []
  imageError: [message: string]
  imageInserted: [count: number]
  latest: []
  openReply: []
  selectRail: [postNo: number]
  submit: []
  'update:mobileRailOpen': [value: boolean]
  'update:open': [value: boolean]
}>()

const content = defineModel<string>({ default: '' })
const { t } = useI18n()
const editorEl = ref<HTMLTextAreaElement | null>(null)
const uploadingImage = ref(false)
const dragOver = ref(false)
const composerBusy = computed(() => props.submitting || uploadingImage.value)
const showFloatingControls = computed(() => props.hasRail || props.authenticated)
const editing = computed(() => props.mode === 'edit')
const composerTitle = computed(() => editing.value ? t('article.editOwnReply') : t('article.joinDiscussion'))
const composerPlaceholder = computed(() => editing.value ? t('article.editReplyPlaceholder') : t('article.replyPlaceholder'))
const submitText = computed(() => {
  if (uploadingImage.value) return t('publish.processingImage')
  if (props.submitting) return editing.value ? t('common.saving') : t('article.publishing')
  return editing.value ? t('common.save') : t('article.publishReply')
})

watch(
  () => props.open,
  async (open) => {
    if (!open) return
    await nextTick()
    window.requestAnimationFrame(() => editorEl.value?.focus())
  },
)

function openReply() {
  emit('openReply')
}

function closeComposer() {
  if (composerBusy.value) return
  emit('update:open', false)
}

function toggleMobileRail() {
  emit('update:mobileRailOpen', !props.mobileRailOpen)
}

function closeMobileRail() {
  emit('update:mobileRailOpen', false)
}

function imageAlt(filename: string) {
  return filename.replace(/\.[^.]+$/, '').replace(/[[\]\n\r]/g, ' ').trim() || 'image'
}

function insertMarkdown(text: string) {
  const el = editorEl.value
  if (!el) {
    content.value = content.value ? `${content.value}\n${text}` : text
    return
  }

  const start = el.selectionStart
  const end = el.selectionEnd
  const before = content.value.slice(0, start)
  const after = content.value.slice(end)
  const prefix = before && !before.endsWith('\n') ? '\n' : ''
  const suffix = after && !after.startsWith('\n') ? '\n' : ''
  content.value = `${before}${prefix}${text}${suffix}${after}`
  nextTick(() => {
    el.focus()
    const cursor = start + prefix.length + text.length
    el.setSelectionRange(cursor, cursor)
  })
}

function imageFilesFromList(files: FileList | File[] | null | undefined) {
  return Array.from(files || []).filter((file) => file.type.startsWith('image/'))
}

function imageFilesFromDataTransfer(dataTransfer: DataTransfer | null) {
  if (!dataTransfer) return []
  return imageFilesFromList(dataTransfer.files)
}

function hasImageDataTransfer(dataTransfer: DataTransfer | null) {
  if (!dataTransfer) return false
  if (Array.from(dataTransfer.items || []).some((item) => item.kind === 'file' && item.type.startsWith('image/'))) return true
  return imageFilesFromList(dataTransfer.files).length > 0
}

function imageFilesFromClipboard(data: DataTransfer | null) {
  if (!data) return []
  return Array.from(data.items || [])
    .filter((item) => item.kind === 'file' && item.type.startsWith('image/'))
    .map((item) => item.getAsFile())
    .filter((file): file is File => Boolean(file))
}

async function uploadImageFiles(files: File[]) {
  if (!files.length || uploadingImage.value) return

  uploadingImage.value = true
  emit('clearValidation')
  const markdownImages: string[] = []
  const failed: string[] = []

  try {
    for (const file of files) {
      const validation = validateImageFile(file)
      if (validation) {
        failed.push(`${file.name}: ${validation}`)
        continue
      }

      try {
        const optimized = await processImageFile(file)
        const url = await uploadImage(optimized.file)
        markdownImages.push(`![${imageAlt(file.name)}](${url})`)
      } catch (error) {
        failed.push(`${file.name}: ${error instanceof Error ? error.message : t('api.imageUploadFailed')}`)
      }
    }

    if (markdownImages.length) {
      insertMarkdown(markdownImages.join('\n'))
      emit('imageInserted', markdownImages.length)
    }

    if (failed.length) {
      emit('imageError', failed.slice(0, 3).join(t('punctuation.semicolon')) + (failed.length > 3 ? t('publish.moreImageFailures', { count: failed.length - 3 }) : ''))
    } else if (!markdownImages.length) {
      emit('imageError', t('publish.noUploadableImages'))
    }
  } finally {
    uploadingImage.value = false
  }
}

async function handleImageInput(event: Event) {
  const input = event.target as HTMLInputElement
  const files = imageFilesFromList(input.files)
  input.value = ''
  await uploadImageFiles(files)
}

async function handlePaste(event: ClipboardEvent) {
  const files = imageFilesFromClipboard(event.clipboardData)
  if (files.length) {
    event.preventDefault()
    await uploadImageFiles(files)
    return
  }

  const markdown = markdownFromClipboard(event.clipboardData)
  if (!markdown) return
  event.preventDefault()
  insertMarkdown(markdown)
}

async function handleDrop(event: DragEvent) {
  dragOver.value = false
  const files = imageFilesFromDataTransfer(event.dataTransfer)
  if (!files.length) return
  event.preventDefault()
  await uploadImageFiles(files)
}

function handleDragOver(event: DragEvent) {
  if (!hasImageDataTransfer(event.dataTransfer)) return
  event.preventDefault()
  dragOver.value = true
}

function submit() {
  if (composerBusy.value) return
  emit('submit')
}
</script>

<template>
  <Teleport v-if="hasRail || showFloatingControls || open" to="body">
    <div
      v-if="mobileRailOpen"
      class="pointer-events-auto fixed inset-0 z-[89] xl:hidden"
      @click="closeMobileRail"
    />
    <div class="pointer-events-none fixed inset-x-0 bottom-4 z-[90] px-3 sm:px-6">
      <div class="relative mx-auto flex w-full max-w-full justify-center">
        <Transition name="floating-reply" mode="out-in">
          <div
            v-if="mobileRailOpen"
            class="gf-floating-surface pointer-events-auto relative w-[min(18rem,calc(100vw-2rem))] p-2 xl:hidden"
            @click.stop
          >
            <div class="mb-1 flex items-center justify-between gap-3 px-1">
              <div class="text-xs font-semibold text-base-content/55">{{ t('article.replyPosition') }}</div>
              <button
                type="button"
                class="inline-flex h-7 w-7 items-center justify-center rounded-md text-icon-muted transition hover:bg-base-300 hover:text-base-content"
                :aria-label="t('common.close')"
                @click="closeMobileRail"
              >
                <X class="h-4 w-4" />
              </button>
            </div>
            <PostPositionRail
              :current="currentNo"
              :max="maxNo"
              :start-label="startLabel"
              :end-label="endLabel"
              :current-label="currentLabel"
              :busy="railBusy"
              :progress-current="progressCurrent"
              :progress-end="progressEnd"
              :progress-start="progressStart"
              @earliest="emit('earliest')"
              @latest="emit('latest')"
              @select="emit('selectRail', $event)"
            />
          </div>
          <div v-else-if="!open" class="pointer-events-auto flex max-w-full flex-col items-center gap-2">
            <div class="gf-floating-surface flex w-fit max-w-full items-center gap-1 rounded-full p-1">
              <button
                v-if="hasRail"
                type="button"
                class="inline-flex h-9 items-center rounded-full px-2.5 text-sm font-black tabular-nums text-primary transition hover:bg-info/10 hover:text-primary xl:hidden"
                :aria-expanded="mobileRailOpen"
                :aria-label="t('article.replyPosition')"
                @click="toggleMobileRail"
              >
                {{ `${currentNo} / ${formatNumber(maxNo)}` }}
              </button>
              <button
                v-if="authenticated"
                v-for="action in actions"
                :key="action.key"
                type="button"
                class="inline-flex h-9 w-9 items-center justify-center rounded-full text-sm font-semibold transition disabled:cursor-not-allowed disabled:opacity-60"
                :class="action.active ? action.activeClass : 'text-base-content/75 hover:bg-base-200 hover:text-base-content'"
                :disabled="action.acting"
                :title="action.title"
                @click="action.onClick"
              >
                <Loader2 v-if="action.acting" class="h-4 w-4 animate-spin" />
                <component :is="action.icon" v-else class="h-4 w-4" :fill="action.active && action.fill !== false ? 'currentColor' : 'none'" />
              </button>
              <button
                v-if="authenticated && canPost"
                type="button"
                class="inline-flex h-9 items-center gap-1.5 rounded-full px-3 text-sm font-semibold text-base-content/75 transition hover:bg-info/10 hover:text-primary"
                :title="t('article.joinDiscussion')"
                @click="openReply"
              >
                <MessageSquare class="h-4 w-4" />
                <span>{{ t('article.joinDiscussion') }}</span>
              </button>
            </div>
          </div>
          <div v-else-if="authenticated" class="gf-floating-surface pointer-events-auto relative w-[min(42rem,calc(100vw-1.5rem))] p-3">
            <div class="mb-2 flex items-center justify-between gap-3">
              <div class="min-w-0">
                <div class="text-sm font-semibold text-base-content">{{ composerTitle }}</div>
              </div>
              <button type="button" class="rounded-md p-1 text-base-content/55 transition hover:bg-base-300 hover:text-base-content/75 disabled:cursor-not-allowed disabled:opacity-60" :disabled="composerBusy" @click="closeComposer">
                <X class="h-4 w-4" />
              </button>
            </div>
            <div v-if="target && !editing" class="mb-2 flex min-w-0 items-center justify-between gap-3 rounded-md border border-primary/20 bg-info/10 px-3 py-2">
              <div class="min-w-0 text-sm font-medium text-base-content/75">
                {{ t('article.replyTo', { user: `@${target.author.username}` }) }}
              </div>
              <button type="button" class="gf-icon-button h-7 w-7 shrink-0 hover:bg-base-100" :aria-label="t('common.cancel')" @click="emit('clearTarget')">
                <X class="h-3.5 w-3.5" />
              </button>
            </div>
            <textarea
              id="reply-content"
              ref="editorEl"
              v-model="content"
              rows="3"
              class="gf-textarea min-h-24 leading-6"
              :placeholder="composerPlaceholder"
              @input="emit('clearValidation')"
              @paste="handlePaste"
              @drop="handleDrop"
              @dragover="handleDragOver"
              @dragleave="dragOver = false"
            />
            <div
              v-if="dragOver"
              class="pointer-events-none absolute inset-x-3 top-11 bottom-15 grid place-items-center rounded-md border-2 border-dashed border-primary/60 bg-info/10 text-sm font-semibold text-primary"
            >
              {{ t('publish.dropToUpload') }}
            </div>
            <p v-if="errorMessage" class="mt-2 text-sm text-error">{{ errorMessage }}</p>
            <p v-if="successMessage" class="mt-2 text-sm text-success">{{ successMessage }}</p>
            <div class="mt-3 flex items-center justify-between gap-2">
              <label class="gf-icon-button h-9 w-9 cursor-pointer" :class="{ 'cursor-wait opacity-60': uploadingImage }" :title="t('publish.uploadImageTitle')">
                <Loader2 v-if="uploadingImage" class="h-4 w-4 animate-spin" />
                <Image v-else class="h-4 w-4" />
                <input type="file" accept="image/*" multiple class="hidden" :disabled="uploadingImage" @change="handleImageInput" />
              </label>
              <div class="flex justify-end gap-2">
                <button v-if="target && !editing" type="button" class="gf-button gf-button-md gf-button-muted" @click="emit('clearTarget')">
                  {{ t('common.cancel') }}
                </button>
                <button type="button" class="gf-button gf-button-md gf-button-primary" :disabled="composerBusy" @click="submit">
                  <Loader2 v-if="composerBusy" class="h-4 w-4 animate-spin" />
                  <Check v-else-if="editing" class="h-4 w-4" />
                  <Send v-else class="h-4 w-4" />
                  {{ submitText }}
                </button>
              </div>
            </div>
          </div>
        </Transition>
      </div>
    </div>
  </Teleport>
</template>
