<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { Bold, Code2, Image, Italic, Link, ListChecks, MessageSquareQuote, Send, X } from '@lucide/vue'
import MarkdownIt from 'markdown-it'
import anchor from 'markdown-it-anchor'
import taskLists from 'markdown-it-task-lists'
import { submitArticle, uploadImage } from '@/runtime/api'
import { processImageFile, validateImageFile } from '@/runtime/image'
import { useUnsavedDraftGuard } from '@/site/composables/useUnsavedDraftGuard'
import PageHeader from '@/site/components/PageHeader.vue'
import type { LayoutPayload, PublishPageProps } from '@/types/payload'
import { useI18n } from 'vue-i18n'

const page = defineProps<{
  layout: LayoutPayload
  props: PublishPageProps
}>()

const { t } = useI18n()
const title = ref(page.props.article.title || '')
const content = ref(page.props.article.content || '')
const type = ref(page.props.article.type || page.props.types[0]?.value || 0)
const categoryIds = ref<number[]>([...(page.props.article.categoryIds || [])])
const currentArticleId = ref(page.props.articleId)
const preview = ref(false)
const submitting = ref(false)
const uploading = ref(false)
const dragOver = ref(false)
const uploadTotal = ref(0)
const uploadDone = ref(0)
const message = ref('')
const error = ref('')
const editor = ref<HTMLTextAreaElement | null>(null)
const markdown = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
  breaks: false,
})
  .use(anchor)
  .use(taskLists, { enabled: true })

const isValid = computed(() => Boolean(title.value.trim() && content.value.trim() && categoryIds.value.length > 0))
const selectedCategories = computed(() => page.props.categories.filter((category) => categoryIds.value.includes(category.id)))
const renderedPreview = computed(() => markdown.render(content.value || ''))
const draftSaveable = computed(() => isValid.value && !submitting.value && !uploading.value)
const savedSnapshot = ref(editorSnapshot())
const hasUnsavedChanges = computed(() => editorSnapshot() !== savedSnapshot.value)
const uploadText = computed(() => {
  if (!uploading.value) return ''
  return uploadTotal.value > 1 ? t('publish.processingImages', { done: uploadDone.value, total: uploadTotal.value }) : t('publish.processingImage')
})
const {
  leavePromptOpen,
  forceNextNavigation,
  closeLeavePrompt,
  discardAndLeave,
  saveDraftAndLeave,
} = useUnsavedDraftGuard({
  hasUnsavedChanges,
  canSaveDraft: draftSaveable,
  saveDraftBeforeLeave: () => persistDraft(undefined, false),
})

function editorSnapshot() {
  return JSON.stringify({
    title: title.value.trim(),
    content: content.value.trim(),
    type: type.value,
    categoryIds: [...categoryIds.value].sort((a, b) => a - b),
  })
}

function syncSavedSnapshot() {
  savedSnapshot.value = editorSnapshot()
}

function typeLabel(item: { name: string }) {
  return t(`publish.types.${item.name}`)
}

function toggleCategory(id: number) {
  if (categoryIds.value.includes(id)) {
    categoryIds.value = categoryIds.value.filter((item) => item !== id)
    return
  }
  if (categoryIds.value.length >= 3) return
  categoryIds.value = [...categoryIds.value, id]
}

function insert(before: string, after = '', placeholder = '') {
  const el = editor.value
  if (!el) return
  const start = el.selectionStart
  const end = el.selectionEnd
  const selected = content.value.slice(start, end) || placeholder
  content.value = `${content.value.slice(0, start)}${before}${selected}${after}${content.value.slice(end)}`
  nextTick(() => {
    el.focus()
    const cursor = start + before.length + selected.length
    el.setSelectionRange(start + before.length, cursor)
  })
}

function imageAlt(filename: string) {
  return filename.replace(/\.[^.]+$/, '').replace(/[[\]\n\r]/g, ' ').trim() || 'image'
}

function insertMarkdownBlock(text: string) {
  const el = editor.value
  if (!el) {
    content.value = content.value ? `${content.value}\n${text}` : text
    return
  }

  const start = el.selectionStart
  const end = el.selectionEnd
  const before = content.value.slice(0, start)
  const after = content.value.slice(end)
  const prefix = before && !before.endsWith('\n') ? '\n' : ''
  const suffix = after && !text.endsWith('\n') ? '\n' : ''
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
  if (!files.length || uploading.value) return

  uploading.value = true
  uploadTotal.value = files.length
  uploadDone.value = 0
  message.value = ''
  error.value = ''

  const markdownImages: string[] = []
  const failed: string[] = []

  try {
    for (const file of files) {
      const validation = validateImageFile(file)
      if (validation) {
        failed.push(`${file.name}: ${validation}`)
        uploadDone.value += 1
        continue
      }

      try {
        const optimized = await processImageFile(file)
        const url = await uploadImage(optimized.file)
        markdownImages.push(`![${imageAlt(file.name)}](${url})`)
      } catch (err) {
        failed.push(`${file.name}: ${err instanceof Error ? err.message : t('api.imageUploadFailed')}`)
      } finally {
        uploadDone.value += 1
      }
    }

    if (markdownImages.length) {
      insertMarkdownBlock(markdownImages.join('\n'))
      message.value = markdownImages.length > 1 ? t('publish.imagesInserted', { count: markdownImages.length }) : t('publish.imageInserted')
    }

    if (failed.length) {
      error.value = failed.slice(0, 3).join(t('punctuation.semicolon')) + (failed.length > 3 ? t('publish.moreImageFailures', { count: failed.length - 3 }) : '')
    } else if (!markdownImages.length) {
      error.value = t('publish.noUploadableImages')
    }
  } finally {
    uploading.value = false
    uploadTotal.value = 0
    uploadDone.value = 0
  }
}

async function handleImage(event: Event) {
  const input = event.target as HTMLInputElement
  const files = imageFilesFromList(input.files)
  input.value = ''
  await uploadImageFiles(files)
}

async function handlePaste(event: ClipboardEvent) {
  const files = imageFilesFromClipboard(event.clipboardData)
  if (!files.length) return
  event.preventDefault()
  await uploadImageFiles(files)
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

function handleEditorKeydown(event: KeyboardEvent) {
  if (!(event.metaKey || event.ctrlKey)) return
  const key = event.key.toLowerCase()
  if (key === 'b') {
    event.preventDefault()
    insert('**', '**', t('publish.placeholder.bold'))
  } else if (key === 'i') {
    event.preventDefault()
    insert('*', '*', t('publish.placeholder.italic'))
  } else if (key === 'k') {
    event.preventDefault()
    insert('[', '](https://)', t('publish.placeholder.link'))
  }
}

async function save() {
  if (!isValid.value || submitting.value) return
  submitting.value = true
  error.value = ''
  message.value = ''
  try {
    const id = await submitArticle({
      id: currentArticleId.value,
      title: title.value.trim(),
      content: content.value.trim(),
      type: type.value,
      categoryId: categoryIds.value,
      articleStatus: 1,
    })
    currentArticleId.value = id
    syncSavedSnapshot()
    forceNextNavigation()
    message.value = page.props.isEditing ? t('publish.topicUpdated') : t('publish.topicPublished')
    window.location.href = `/p/post/${id}`
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('publish.saveFailed')
  } finally {
    submitting.value = false
  }
}

async function saveDraft() {
  if (!isValid.value || submitting.value) return
  await persistDraft('/drafts')
}

async function persistDraft(nextUrl?: string, redirect = true): Promise<boolean> {
  submitting.value = true
  error.value = ''
  message.value = ''
  try {
    const id = await submitArticle({
      id: currentArticleId.value,
      title: title.value.trim(),
      content: content.value.trim(),
      type: type.value,
      categoryId: categoryIds.value,
      articleStatus: 0,
    })
    currentArticleId.value = id
    syncSavedSnapshot()
    forceNextNavigation()
    if (redirect) window.location.href = nextUrl || '/drafts'
    return true
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('publish.draftSaveFailed')
    return false
  } finally {
    submitting.value = false
  }
}
</script>

<template>
    <main class="min-w-0 pb-12">
      <PageHeader :title="props.isEditing ? t('publish.editTitle') : t('publish.createTitle')" :description="t('publish.subtitle')" />

      <div class="grid gap-3 xl:grid-cols-[minmax(0,1fr)_280px]">
        <section class="gf-card p-4 sm:p-5">
          <div class="space-y-5">
            <label class="block">
              <span class="text-sm font-semibold text-base-content/75">{{ t('publish.fields.title') }}</span>
              <input
                v-model="title"
                class="mt-1 h-11 w-full rounded-md border border-line px-3 text-lg font-semibold outline-none transition focus:border-primary focus:ring-4 focus:ring-primary/20"
                :placeholder="t('publish.titlePlaceholder')"
              />
            </label>

            <div>
              <div class="mb-2 text-sm font-semibold text-base-content/75">{{ t('publish.fields.type') }}</div>
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="item in props.types"
                  :key="item.value"
                  type="button"
                  class="rounded-md border px-3 py-1.5 text-sm font-medium transition"
                  :class="type === item.value ? 'border-primary bg-info/10 text-primary' : 'border-line text-base-content/75 hover:border-line hover:bg-base-200'"
                  @click="type = item.value"
                >
                  {{ typeLabel(item) }}
                </button>
              </div>
            </div>

            <div>
              <div class="mb-2 flex items-center justify-between">
                <span class="text-sm font-semibold text-base-content/75">{{ t('publish.fields.category') }}</span>
                <span class="text-xs text-base-content/55">{{ t('publish.maxCategories') }}</span>
              </div>
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="category in props.categories"
                  :key="category.id"
                  type="button"
                  class="inline-flex items-center gap-1.5 rounded-md border px-2.5 py-1.5 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-40"
                  :class="categoryIds.includes(category.id) ? 'border-primary bg-info/10 text-primary' : 'border-line text-base-content/75 hover:border-line hover:bg-base-200'"
                  :disabled="!categoryIds.includes(category.id) && categoryIds.length >= 3"
                  @click="toggleCategory(category.id)"
                >
                  <span class="h-2 w-2 rounded-[3px]" :style="{ backgroundColor: category.color }" />
                  {{ category.name }}
                </button>
              </div>
            </div>

            <div>
              <div class="mb-2 flex items-center justify-between">
                <span class="text-sm font-semibold text-base-content/75">{{ t('publish.fields.body') }}</span>
                <div class="inline-flex rounded-md border border-line p-0.5 text-xs font-semibold">
                  <button type="button" class="rounded px-2 py-1" :class="!preview ? 'bg-neutral text-neutral-content' : 'text-base-content/55'" @click="preview = false">{{ t('common.edit') }}</button>
                  <button type="button" class="rounded px-2 py-1" :class="preview ? 'bg-neutral text-neutral-content' : 'text-base-content/55'" @click="preview = true">{{ t('publish.preview') }}</button>
                </div>
              </div>

              <div
                v-if="!preview"
                class="overflow-hidden rounded-lg border transition"
                :class="dragOver ? 'border-primary bg-info/10 shadow-[0_0_0_4px_rgba(59,130,246,0.12)]' : 'border-line bg-base-100'"
              >
                <div class="flex flex-wrap items-center gap-1 border-b border-line bg-base-200 px-2 py-2">
                  <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-100 hover:text-base-content" :title="t('publish.toolbar.bold')" @click="insert('**', '**', t('publish.placeholder.bold'))"><Bold class="h-4 w-4" /></button>
                  <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-100 hover:text-base-content" :title="t('publish.toolbar.italic')" @click="insert('*', '*', t('publish.placeholder.italic'))"><Italic class="h-4 w-4" /></button>
                  <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-100 hover:text-base-content" :title="t('publish.toolbar.link')" @click="insert('[', '](https://)', t('publish.placeholder.link'))"><Link class="h-4 w-4" /></button>
                  <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-100 hover:text-base-content" :title="t('publish.toolbar.quote')" @click="insert('\\n> ', '', t('publish.placeholder.quote'))"><MessageSquareQuote class="h-4 w-4" /></button>
                  <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-100 hover:text-base-content" :title="t('publish.toolbar.code')" @click="insert('\\n```\\n', '\\n```\\n', 'code')"><Code2 class="h-4 w-4" /></button>
                  <span v-if="uploadText" class="gf-badge gf-badge-info ml-auto rounded">{{ uploadText }}</span>
                  <label
                    class="rounded p-1.5 text-base-content/55 transition hover:bg-base-100 hover:text-base-content"
                    :class="uploadText ? '' : 'ml-auto'"
                    :title="t('publish.uploadImageTitle')"
                  >
                    <Image class="h-4 w-4" />
                    <input type="file" accept="image/*" multiple class="hidden" :disabled="uploading" @change="handleImage" />
                  </label>
                </div>
                <div class="relative">
                  <textarea
                    ref="editor"
                    v-model="content"
                    class="min-h-96 w-full resize-y border-0 bg-transparent px-4 py-3 font-mono text-sm leading-relaxed outline-none"
                    :placeholder="t('publish.bodyPlaceholder')"
                    @keydown="handleEditorKeydown"
                    @paste="handlePaste"
                    @drop="handleDrop"
                    @dragover="handleDragOver"
                    @dragleave="dragOver = false"
                  />
                  <div
                    v-if="dragOver"
                    class="pointer-events-none absolute inset-3 grid place-items-center rounded-lg border-2 border-dashed border-primary/60 bg-info/10 text-sm font-semibold text-primary"
                  >
                    {{ t('publish.dropToUpload') }}
                  </div>
                </div>
              </div>

              <div v-else class="gf-prose gf-prose-article min-h-96 rounded-lg border border-line bg-base-200/50 p-5">
                <div v-if="content.trim()" v-html="renderedPreview" />
                <p v-else class="text-sm text-base-content/55">{{ t('publish.emptyPreview') }}</p>
              </div>
            </div>

            <p v-if="error" class="gf-status-message gf-status-message-error">{{ error }}</p>
            <p v-if="message" class="gf-status-message gf-status-message-success">{{ message }}</p>

            <div class="flex items-center justify-end gap-2 border-t border-line pt-4">
              <a href="/" class="gf-button gf-button-lg gf-button-muted">{{ t('common.cancel') }}</a>
              <button
                type="button"
                class="gf-button gf-button-lg gf-button-secondary"
                :disabled="!isValid || submitting || uploading"
                @click="saveDraft"
              >
                {{ submitting ? t('common.saving') : t('publish.saveDraft') }}
              </button>
              <button
                type="button"
                class="gf-button gf-button-lg gf-button-primary"
                :disabled="!isValid || submitting || uploading"
                @click="save"
              >
                <Send class="h-4 w-4" />
                {{ submitting ? t('common.saving') : props.isEditing ? t('publish.updateTopic') : t('publish.publishTopic') }}
              </button>
            </div>
          </div>
        </section>

        <aside class="space-y-3">
          <section class="gf-card p-4">
            <div class="flex items-center gap-2">
              <ListChecks class="h-4 w-4 text-base-content/55" />
              <h2 class="text-sm font-semibold text-base-content">{{ t('publish.checklist.title') }}</h2>
            </div>
            <ul class="mt-3 space-y-2 text-sm text-base-content/75">
              <li class="flex items-center justify-between gap-3"><span>{{ t('publish.fields.title') }}</span><span :class="title.trim() ? 'text-success' : 'text-base-content/55'">{{ title.trim() ? t('publish.checklist.done') : t('publish.checklist.pending') }}</span></li>
              <li class="flex items-center justify-between gap-3"><span>{{ t('publish.fields.category') }}</span><span :class="categoryIds.length ? 'text-success' : 'text-base-content/55'">{{ categoryIds.length }}/3</span></li>
              <li class="flex items-center justify-between gap-3"><span>{{ t('publish.fields.body') }}</span><span :class="content.trim() ? 'text-success' : 'text-base-content/55'">{{ t('publish.checklist.characters', { count: content.trim().length }) }}</span></li>
            </ul>
          </section>

          <section v-if="selectedCategories.length" class="gf-card p-4">
            <h2 class="text-sm font-semibold text-base-content">{{ t('publish.selectedCategories') }}</h2>
            <div class="mt-3 flex flex-wrap gap-2">
              <button
                v-for="category in selectedCategories"
                :key="category.id"
                type="button"
                class="inline-flex items-center gap-1.5 rounded-md border border-line px-2 py-1 text-sm text-base-content/75 hover:bg-base-200"
                @click="toggleCategory(category.id)"
              >
                <span class="h-2 w-2 rounded-[3px]" :style="{ backgroundColor: category.color }" />
                {{ category.name }}
                <X class="h-3 w-3" />
              </button>
            </div>
          </section>
        </aside>
      </div>

      <div v-if="leavePromptOpen" class="fixed inset-0 z-[100] flex items-center justify-center bg-neutral/50 px-4 backdrop-blur-sm" role="dialog" aria-modal="true">
        <div class="gf-menu-surface w-full max-w-md overflow-hidden">
          <div class="border-b border-line px-5 py-4">
            <h2 class="text-base font-semibold text-base-content">{{ t('publish.leaveTitle') }}</h2>
            <p class="mt-1 text-sm leading-6 text-base-content/55">
              {{ t('publish.leaveDescription') }}
            </p>
          </div>

          <div v-if="!isValid" class="border-b border-warning/20 bg-warning/10 px-5 py-3 text-sm font-medium text-warning">
            {{ t('publish.draftRequirement') }}
          </div>

          <div class="flex flex-wrap items-center justify-end gap-2 bg-base-200 px-5 py-4">
            <button type="button" class="gf-button gf-button-lg gf-button-muted" @click="closeLeavePrompt">
              {{ t('publish.continueEditing') }}
            </button>
            <button type="button" class="gf-button gf-button-lg gf-button-secondary" @click="discardAndLeave">
              {{ t('publish.leaveWithoutSaving') }}
            </button>
            <button
              type="button"
              class="gf-button gf-button-lg gf-button-primary min-w-28"
              :disabled="!draftSaveable"
              @click="saveDraftAndLeave"
            >
              {{ submitting ? t('common.saving') : t('publish.saveDraft') }}
            </button>
          </div>
        </div>
      </div>
    </main>
</template>
