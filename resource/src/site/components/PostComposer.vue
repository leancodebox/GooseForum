<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { Bold, Check, Code, Code2, Image, Italic, Link, List, ListOrdered, Loader2, Maximize2, MessageSquareQuote, Minimize2, Minus, Send, Strikethrough, X } from '@lucide/vue'
import { uploadImage } from '@/runtime/api'
import { processImageFile, validateImageFile } from '@/runtime/image'
import { markdownFromClipboard } from '@/runtime/rich-paste'
import { useMarkdownTextarea, type MarkdownToolbarAction } from '@/site/composables/useMarkdownTextarea'
import UserAvatar from '@/site/components/UserAvatar.vue'
import type { PostPayload, ViewerPayload } from '@gooseforum/client'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  authenticated: boolean
  errorMessage: string
  mode?: 'create' | 'edit'
  open: boolean
  submitting: boolean
  successMessage: string
  target?: PostPayload
  topicTitle: string
  viewer: ViewerPayload
}>()

const emit = defineEmits<{
  clearTarget: []
  clearValidation: []
  imageError: [message: string]
  imageInserted: [count: number]
  submit: []
  'update:open': [value: boolean]
}>()

const content = defineModel<string>({ default: '' })
const { t } = useI18n()
const editorEl = ref<HTMLTextAreaElement | null>(null)
const uploadingImage = ref(false)
const dragOver = ref(false)
const dockState = ref<'compact' | 'expanded' | 'minimized'>('compact')
const composerBusy = computed(() => props.submitting || uploadingImage.value)
const editing = computed(() => props.mode === 'edit')
const expanded = computed(() => dockState.value === 'expanded')
const minimized = computed(() => dockState.value === 'minimized')
const composerTitle = computed(() => editing.value ? t('topic.editOwnReply') : t('topic.joinDiscussion'))
const composerPlaceholder = computed(() => editing.value ? t('topic.editReplyPlaceholder') : t('topic.replyPlaceholder'))
const submitText = computed(() => {
  if (uploadingImage.value) return t('publish.processingImage')
  if (props.submitting) return editing.value ? t('common.saving') : t('topic.publishing')
  return editing.value ? t('common.save') : t('topic.publishReply')
})
const markdownEditor = useMarkdownTextarea({ content, editor: editorEl })
watch(
  () => props.open,
  async (open) => {
    if (!open) return
    if (minimized.value) dockState.value = 'compact'
    await nextTick()
    window.requestAnimationFrame(() => editorEl.value?.focus())
  },
  { immediate: true },
)

watch(
  () => [props.target?.id, props.mode] as const,
  async () => {
    if (!props.open) return
    if (minimized.value) dockState.value = 'compact'
    await focusEditor()
  },
)

async function focusEditor() {
  await nextTick()
  window.requestAnimationFrame(() => editorEl.value?.focus())
}

function closeComposer() {
  if (composerBusy.value) return
  emit('update:open', false)
}

function minimizeComposer() {
  if (composerBusy.value) return
  dockState.value = 'minimized'
}

function restoreComposer() {
  dockState.value = 'compact'
  void focusEditor()
}

function toggleExpanded() {
  dockState.value = expanded.value ? 'compact' : 'expanded'
  void focusEditor()
}

function markdownPlaceholders() {
  return {
    bold: t('publish.placeholder.bold'),
    italic: t('publish.placeholder.italic'),
    strike: t('publish.placeholder.strike'),
    link: t('publish.placeholder.link'),
    quote: t('publish.placeholder.quote'),
    listItem: t('publish.placeholder.listItem'),
  }
}

function applyToolbarAction(action: MarkdownToolbarAction) {
  markdownEditor.applyAction(action, markdownPlaceholders())
  emit('clearValidation')
}

function imageAlt(filename: string) {
  return filename.replace(/\.[^.]+$/, '').replace(/[[\]\n\r]/g, ' ').trim() || 'image'
}

function insertMarkdown(text: string) {
  markdownEditor.insertBlock(text)
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

function handleEditorKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    event.preventDefault()
    minimizeComposer()
    return
  }
  if (!(event.metaKey || event.ctrlKey)) return
  const key = event.key.toLowerCase()
  if (key === 'enter') {
    event.preventDefault()
    submit()
  } else if (key === 'b') {
    event.preventDefault()
    applyToolbarAction('bold')
  } else if (key === 'i') {
    event.preventDefault()
    applyToolbarAction('italic')
  } else if (key === 'k') {
    event.preventDefault()
    applyToolbarAction('link')
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="composer-dock" appear>
      <div v-if="open && authenticated" class="pointer-events-none fixed inset-x-0 bottom-0 z-[90] px-0 sm:px-8">
      <div class="relative mx-auto flex w-full max-w-[58rem] justify-center">
        <Transition name="floating-reply" mode="out-in">
          <div v-if="minimized" key="minimized" class="gf-composer-minimized pointer-events-auto flex w-[min(32rem,calc(100vw-1.5rem))] items-center gap-3 px-3 py-2">
            <button type="button" class="flex min-w-0 flex-1 items-center gap-3 text-left" :aria-label="t('topic.restoreComposer')" @click="restoreComposer">
              <UserAvatar :src="viewer.avatarUrl" :alt="viewer.username" class="h-8 w-8 shrink-0 rounded-full object-cover ring-1 ring-line" />
              <span class="min-w-0">
                <span class="block truncate text-sm font-semibold text-base-content">{{ composerTitle }}</span>
                <span class="block truncate text-xs text-base-content/55">{{ topicTitle }}</span>
              </span>
            </button>
            <button type="button" class="gf-icon-button h-8 w-8 shrink-0" :title="t('topic.restoreComposer')" @click="restoreComposer">
              <Maximize2 class="h-4 w-4" />
            </button>
            <button type="button" class="gf-icon-button h-8 w-8 shrink-0" :title="t('common.close')" @click="closeComposer">
              <X class="h-4 w-4" />
            </button>
          </div>

          <section
            v-else
            key="composer"
            class="gf-composer-dock pointer-events-auto relative flex w-full flex-col"
            :class="expanded ? 'h-[calc(100dvh-0.75rem)] sm:h-[min(70dvh,44rem)]' : 'h-[min(21rem,52dvh)] sm:h-[22rem]'"
            :aria-label="composerTitle"
          >
            <header class="flex min-h-13 shrink-0 items-center gap-3 border-b border-line px-3 py-2 sm:px-4">
              <UserAvatar :src="viewer.avatarUrl" :alt="viewer.username" class="h-8 w-8 shrink-0 rounded-full object-cover ring-1 ring-line sm:h-10 sm:w-10" />
              <div class="min-w-0 flex-1">
                <div class="truncate text-sm font-semibold text-base-content">{{ composerTitle }}</div>
                <div class="truncate text-xs text-base-content/55">{{ topicTitle }}</div>
              </div>
              <div class="flex shrink-0 items-center gap-0.5">
                <button type="button" class="gf-icon-button h-8 w-8" :title="t('topic.minimizeComposer')" :disabled="composerBusy" @click="minimizeComposer">
                  <Minus class="h-4 w-4" />
                </button>
                <button type="button" class="gf-icon-button h-8 w-8" :title="expanded ? t('topic.restoreComposer') : t('topic.expandComposer')" @click="toggleExpanded">
                  <Minimize2 v-if="expanded" class="h-4 w-4" />
                  <Maximize2 v-else class="h-4 w-4" />
                </button>
                <button type="button" class="gf-icon-button h-8 w-8" :title="t('common.close')" :disabled="composerBusy" @click="closeComposer">
                  <X class="h-4 w-4" />
                </button>
              </div>
            </header>

            <div v-if="target && !editing" class="flex min-w-0 shrink-0 items-center gap-2 border-b border-primary/15 bg-info/10 px-3 py-2 text-xs font-medium text-base-content/75">
              <span class="min-w-0 flex-1 truncate">{{ t('topic.replyToPost', { user: `@${target.author.username}`, no: target.postNo || target.id }) }}</span>
              <button type="button" class="gf-icon-button h-6 w-6 shrink-0 hover:bg-base-100" :aria-label="t('common.cancel')" @click="emit('clearTarget')">
                <X class="h-3.5 w-3.5" />
              </button>
            </div>

            <div class="relative min-h-0 flex-1">
              <textarea
                id="reply-content"
                ref="editorEl"
                v-model="content"
                class="block h-full min-h-0 w-full resize-none border-0 bg-base-100 px-4 py-3 text-[15px] leading-6 text-base-content outline-none placeholder:text-base-content/45"
                :placeholder="composerPlaceholder"
                @input="emit('clearValidation')"
                @keydown="handleEditorKeydown"
                @paste="handlePaste"
                @drop="handleDrop"
                @dragover="handleDragOver"
                @dragleave="dragOver = false"
              />
              <div v-if="dragOver" class="pointer-events-none absolute inset-3 grid place-items-center rounded-md border-2 border-dashed border-primary/60 bg-info/10 text-sm font-semibold text-primary">
                {{ t('publish.dropToUpload') }}
              </div>
            </div>

            <div v-if="errorMessage || successMessage" class="shrink-0 border-t border-line px-3 py-1.5 text-xs" :class="errorMessage ? 'text-error' : 'text-success'">
              {{ errorMessage || successMessage }}
            </div>

            <footer class="flex shrink-0 items-center gap-1 border-t border-line bg-base-100 px-2 py-2 pb-[max(0.5rem,env(safe-area-inset-bottom))] sm:px-3 sm:pb-2">
              <button type="button" class="gf-button gf-button-md gf-button-primary mr-1 shrink-0 px-3" :disabled="composerBusy" @click="submit">
                <Loader2 v-if="composerBusy" class="h-4 w-4 animate-spin" />
                <Check v-else-if="editing" class="h-4 w-4" />
                <Send v-else class="h-4 w-4" />
                <span class="hidden sm:inline">{{ submitText }}</span>
              </button>

              <div class="flex min-w-0 flex-1 items-center gap-0.5 overflow-x-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
                <label class="gf-composer-tool shrink-0 cursor-pointer" :class="{ 'cursor-wait opacity-60': uploadingImage }" :title="t('publish.uploadImageTitle')">
                  <Loader2 v-if="uploadingImage" class="h-4 w-4 animate-spin" />
                  <Image v-else class="h-4 w-4" />
                  <input type="file" accept="image/*" multiple class="hidden" :disabled="uploadingImage" @change="handleImageInput" />
                </label>
                <span class="mx-1 h-5 w-px shrink-0 bg-line" />
                <button type="button" class="gf-composer-tool" :title="t('publish.toolbar.bold')" @mousedown.prevent @click="applyToolbarAction('bold')"><Bold class="h-4 w-4" /></button>
                <button type="button" class="gf-composer-tool" :title="t('publish.toolbar.italic')" @mousedown.prevent @click="applyToolbarAction('italic')"><Italic class="h-4 w-4" /></button>
                <button type="button" class="gf-composer-tool" :title="t('publish.toolbar.strike')" @mousedown.prevent @click="applyToolbarAction('strike')"><Strikethrough class="h-4 w-4" /></button>
                <button type="button" class="gf-composer-tool" :title="t('publish.toolbar.inlineCode')" @mousedown.prevent @click="applyToolbarAction('inlineCode')"><Code class="h-4 w-4" /></button>
                <button type="button" class="gf-composer-tool" :title="t('publish.toolbar.link')" @mousedown.prevent @click="applyToolbarAction('link')"><Link class="h-4 w-4" /></button>
                <button type="button" class="gf-composer-tool" :title="t('publish.toolbar.quote')" @mousedown.prevent @click="applyToolbarAction('quote')"><MessageSquareQuote class="h-4 w-4" /></button>
                <button type="button" class="gf-composer-tool" :title="t('publish.toolbar.code')" @mousedown.prevent @click="applyToolbarAction('code')"><Code2 class="h-4 w-4" /></button>
                <button type="button" class="gf-composer-tool" :title="t('publish.toolbar.bulletList')" @mousedown.prevent @click="applyToolbarAction('bulletList')"><List class="h-4 w-4" /></button>
                <button type="button" class="gf-composer-tool" :title="t('publish.toolbar.orderedList')" @mousedown.prevent @click="applyToolbarAction('orderedList')"><ListOrdered class="h-4 w-4" /></button>
              </div>
            </footer>
          </section>
        </Transition>
      </div>
      </div>
    </Transition>
  </Teleport>
</template>
