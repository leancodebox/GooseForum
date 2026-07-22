<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { Bold, ClipboardPaste, Code, Code2, CornerDownLeft, Eye, Heading, Image, Italic, Link, List, ListChecks, ListOrdered, MessageSquareQuote, Minus, Send, Strikethrough, Table2, X } from '@lucide/vue'
import { submitTopic, uploadImage } from '@/runtime/api'
import { processImageFile, validateImageFile } from '@/runtime/image'
import { renderMarkdownPreview } from '@/runtime/markdown'
import { createMarkdownTable, fencedCodeBlock, formatMarkdownLines, prefixMarkdownBlock, replaceMarkdownSelectionWithBlock, type MarkdownBlockType } from '@/runtime/markdown-editing'
import { hasUnsupportedVisualMarkdown, markdownFromClipboard } from '@/runtime/rich-paste'
import { useUnsavedDraftGuard } from '@/site/composables/useUnsavedDraftGuard'
import PageHeader from '@/site/components/PageHeader.vue'
import VisualMarkdownEditor from '@/site/components/VisualMarkdownEditor.vue'
import type { LayoutPayload, PublishPageProps } from '@/types/payload'
import { useI18n } from 'vue-i18n'

const page = defineProps<{
  layout: LayoutPayload
  props: PublishPageProps
}>()

const { t } = useI18n()
const title = ref(page.props.topic.title || '')
const content = ref(page.props.topic.content || '')
const categoryIds = ref<number[]>([...(page.props.topic.categoryIds || [])])
const currentTopicId = ref(page.props.topicId)
const editorMode = ref<'markdown' | 'visual'>(hasUnsupportedVisualMarkdown(content.value) ? 'markdown' : 'visual')
const preview = ref(false)
const submitting = ref(false)
const uploading = ref(false)
const dragOver = ref(false)
const uploadTotal = ref(0)
const uploadDone = ref(0)
const message = ref('')
const error = ref('')
const editor = ref<HTMLTextAreaElement | null>(null)
const visualEditor = ref<InstanceType<typeof VisualMarkdownEditor> | null>(null)
const blockPicker = ref<HTMLElement | null>(null)
const blockPickerOpen = ref(false)
const linkPicker = ref<HTMLElement | null>(null)
const linkInput = ref<HTMLInputElement | null>(null)
const linkPickerOpen = ref(false)
const linkUrl = ref('https://')
const tablePicker = ref<HTMLElement | null>(null)
const tablePickerOpen = ref(false)
const tablePickerRows = ref(3)
const tablePickerColumns = ref(3)
const tablePickerMaxRows = 8
const tablePickerMaxColumns = 8
const tablePickerCells = Array.from({ length: tablePickerMaxRows * tablePickerMaxColumns }, (_, index) => ({
  row: Math.floor(index / tablePickerMaxColumns) + 1,
  column: (index % tablePickerMaxColumns) + 1,
}))

const isValid = computed(() => Boolean(title.value.trim() && content.value.trim() && categoryIds.value.length > 0))
const selectedCategories = computed(() => page.props.categories.filter((category) => categoryIds.value.includes(category.id)))
const renderedPreview = computed(() => renderMarkdownPreview(content.value))
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
    categoryIds: [...categoryIds.value].sort((a, b) => a - b),
  })
}

function syncSavedSnapshot() {
  savedSnapshot.value = editorSnapshot()
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
  if (editorMode.value === 'visual') {
    visualEditor.value?.insertMarkdown(text)
    return
  }
  const el = editor.value
  if (!el) {
    content.value = content.value ? `${content.value}\n${text}` : text
    return
  }

  const start = el.selectionStart
  const end = el.selectionEnd
  const result = replaceMarkdownSelectionWithBlock(content.value, start, end, text)
  content.value = result.value

  nextTick(() => {
    el.focus()
    el.setSelectionRange(result.selectionEnd, result.selectionEnd)
  })
}

function selectedMarkdownText(placeholder: string) {
  const el = editor.value
  if (!el) return placeholder
  return content.value.slice(el.selectionStart, el.selectionEnd) || placeholder
}

function insertPrefixedMarkdownBlock(prefix: string, placeholder: string) {
  insertMarkdownBlock(prefixMarkdownBlock(selectedMarkdownText(placeholder), prefix))
}

function insertFencedCodeBlock() {
  insertMarkdownBlock(fencedCodeBlock(selectedMarkdownText('code')))
}

async function selectEditorMode(mode: 'markdown' | 'visual') {
  if (editorMode.value === mode && !preview.value) return
  if (mode === 'visual' && hasUnsupportedVisualMarkdown(content.value)) {
    error.value = t('publish.visualUnsupported')
    return
  }
  error.value = ''
  blockPickerOpen.value = false
  linkPickerOpen.value = false
  tablePickerOpen.value = false
  editorMode.value = mode
  preview.value = false
  await nextTick()
  if (mode === 'visual') {
    visualEditor.value?.focus()
  } else {
    editor.value?.focus()
  }
}

async function togglePreview() {
  blockPickerOpen.value = false
  linkPickerOpen.value = false
  tablePickerOpen.value = false
  preview.value = !preview.value
  if (!preview.value && editorMode.value === 'visual') {
    await nextTick()
    visualEditor.value?.restoreEditableBoundaries()
    visualEditor.value?.focus()
  }
}

type ToolbarAction = 'bold' | 'italic' | 'strike' | 'inlineCode' | 'quote' | 'code' | 'bulletList' | 'orderedList' | 'horizontalRule' | 'hardBreak'

function applyToolbarAction(action: ToolbarAction) {
  if (editorMode.value === 'markdown') {
    if (action === 'bold') insert('**', '**', t('publish.placeholder.bold'))
    else if (action === 'italic') insert('*', '*', t('publish.placeholder.italic'))
    else if (action === 'strike') insert('~~', '~~', t('publish.placeholder.strike'))
    else if (action === 'inlineCode') insert('`', '`', 'code')
    else if (action === 'quote') insertPrefixedMarkdownBlock('> ', t('publish.placeholder.quote'))
    else if (action === 'code') insertFencedCodeBlock()
    else if (action === 'bulletList') insertPrefixedMarkdownBlock('- ', t('publish.placeholder.listItem'))
    else if (action === 'orderedList') insertPrefixedMarkdownBlock('1. ', t('publish.placeholder.listItem'))
    else if (action === 'horizontalRule') insertMarkdownBlock('---')
    else insert('  \n')
    return
  }

  visualEditor.value?.applyAction(action)
}

function openBlockPicker() {
  linkPickerOpen.value = false
  tablePickerOpen.value = false
  blockPickerOpen.value = !blockPickerOpen.value
}

function applyBlockType(block: MarkdownBlockType) {
  blockPickerOpen.value = false
  if (editorMode.value === 'visual') {
    visualEditor.value?.setBlock(block)
    return
  }
  if (block === 'code_block') {
    insertFencedCodeBlock()
    return
  }
  setMarkdownLineType(block)
}

function setMarkdownLineType(block: Exclude<MarkdownBlockType, 'code_block'>) {
  const el = editor.value
  if (!el) return
  const result = formatMarkdownLines(content.value, el.selectionStart, el.selectionEnd, block)
  content.value = result.value
  nextTick(() => {
    el.focus()
    el.setSelectionRange(result.selectionStart, result.selectionEnd)
  })
}

async function openLinkPicker() {
  if (editorMode.value === 'markdown') {
    insert('[', '](https://)', t('publish.placeholder.link'))
    return
  }
  blockPickerOpen.value = false
  tablePickerOpen.value = false
  linkPickerOpen.value = !linkPickerOpen.value
  if (!linkPickerOpen.value) return
  linkUrl.value = visualEditor.value?.activeLinkHref() || 'https://'
  await nextTick()
  linkInput.value?.focus()
  linkInput.value?.select()
}

function applyLink() {
  const href = linkUrl.value.trim()
  if (!href) return
  linkPickerOpen.value = false
  visualEditor.value?.setLink(href, t('publish.placeholder.link'))
}

function openTablePicker() {
  blockPickerOpen.value = false
  linkPickerOpen.value = false
  tablePickerRows.value = 3
  tablePickerColumns.value = 3
  tablePickerOpen.value = !tablePickerOpen.value
}

function selectTableSize(row: number, column: number) {
  tablePickerRows.value = row
  tablePickerColumns.value = column
}

function insertTable(row: number, column: number) {
  tablePickerOpen.value = false
  if (editorMode.value === 'visual') visualEditor.value?.insertTable(row, column)
  else insertMarkdownBlock(createMarkdownTable(row, column))
}

function closeToolbarPopovers(event: PointerEvent) {
  const target = event.target as Node
  if (blockPicker.value?.contains(target) || linkPicker.value?.contains(target) || tablePicker.value?.contains(target)) return
  blockPickerOpen.value = false
  linkPickerOpen.value = false
  tablePickerOpen.value = false
}

onMounted(() => document.addEventListener('pointerdown', closeToolbarPopovers))
onBeforeUnmount(() => document.removeEventListener('pointerdown', closeToolbarPopovers))

async function pastePlainText() {
  error.value = ''
  try {
    const text = await navigator.clipboard.readText()
    if (!text) return
    if (editorMode.value === 'visual') {
      visualEditor.value?.insertText(text)
    } else {
      insertPlainTextInMarkdown(text)
    }
  } catch {
    error.value = t('publish.clipboardReadFailed')
  }
}

function insertPlainTextInMarkdown(text: string) {
  const el = editor.value
  if (!el) {
    content.value += text
    return
  }
  const start = el.selectionStart
  const end = el.selectionEnd
  content.value = `${content.value.slice(0, start)}${text}${content.value.slice(end)}`
  nextTick(() => {
    el.focus()
    el.setSelectionRange(start + text.length, start + text.length)
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
  if (files.length) {
    event.preventDefault()
    await uploadImageFiles(files)
    return
  }

  const markdown = markdownFromClipboard(event.clipboardData)
  if (!markdown) return
  event.preventDefault()
  insertMarkdownBlock(markdown)
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
    const id = await submitTopic({
      topicId: currentTopicId.value,
      title: title.value.trim(),
      content: content.value.trim(),
      categoryId: categoryIds.value,
      topicStatus: 1,
    })
    currentTopicId.value = id
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
    const id = await submitTopic({
      topicId: currentTopicId.value,
      title: title.value.trim(),
      content: content.value.trim(),
      categoryId: categoryIds.value,
      topicStatus: 0,
    })
    currentTopicId.value = id
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
    <main class="min-w-0 pb-8">
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
              <div class="mb-1 flex items-center">
                <span class="text-sm font-semibold text-base-content/75">{{ t('publish.fields.body') }}</span>
              </div>

              <div class="mb-2 flex flex-wrap items-center gap-1 py-2">
                <div v-if="!preview" class="flex flex-wrap items-center gap-1">
                  <div ref="blockPicker" class="relative mr-1">
                    <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-200 hover:text-base-content" :title="t('publish.toolbar.blockType')" :aria-expanded="blockPickerOpen" @mousedown.prevent @click="openBlockPicker"><Heading class="h-4 w-4" /></button>
                    <div v-if="blockPickerOpen" class="gf-menu-surface absolute left-0 top-full z-30 mt-1.5 w-48 p-2 shadow-lg" role="menu" @keydown.esc.stop="blockPickerOpen = false">
                      <button type="button" class="flex w-full items-center gap-2 rounded px-2 py-1.5 text-left text-sm text-base-content/75 hover:bg-base-200 hover:text-base-content" @click="applyBlockType('paragraph')">
                        <span class="w-7 font-medium">P</span>{{ t('publish.toolbar.paragraph') }}
                      </button>
                      <button v-for="level in 6" :key="level" type="button" class="flex w-full items-center gap-2 rounded px-2 py-1.5 text-left text-sm text-base-content/75 hover:bg-base-200 hover:text-base-content" @click="applyBlockType(`heading_${level}` as MarkdownBlockType)">
                        <span class="w-7 font-semibold">H{{ level }}</span>{{ t('publish.toolbar.heading', { level }) }}
                      </button>
                      <button type="button" class="flex w-full items-center gap-2 rounded px-2 py-1.5 text-left text-sm text-base-content/75 hover:bg-base-200 hover:text-base-content" @click="applyBlockType('code_block')">
                        <Code2 class="h-4 w-7" />{{ t('publish.toolbar.codeBlock') }}
                      </button>
                    </div>
                  </div>
                  <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-200 hover:text-base-content" :title="t('publish.toolbar.bold')" @mousedown.prevent @click="applyToolbarAction('bold')"><Bold class="h-4 w-4" /></button>
                  <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-200 hover:text-base-content" :title="t('publish.toolbar.italic')" @mousedown.prevent @click="applyToolbarAction('italic')"><Italic class="h-4 w-4" /></button>
                  <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-200 hover:text-base-content" :title="t('publish.toolbar.strike')" @mousedown.prevent @click="applyToolbarAction('strike')"><Strikethrough class="h-4 w-4" /></button>
                  <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-200 hover:text-base-content" :title="t('publish.toolbar.inlineCode')" @mousedown.prevent @click="applyToolbarAction('inlineCode')"><Code class="h-4 w-4" /></button>
                  <div ref="linkPicker" class="relative">
                    <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-200 hover:text-base-content" :title="t('publish.toolbar.link')" :aria-expanded="linkPickerOpen" @mousedown.prevent @click="openLinkPicker"><Link class="h-4 w-4" /></button>
                    <form v-if="linkPickerOpen" class="gf-menu-surface absolute left-0 top-full z-30 mt-1.5 flex w-72 items-center gap-1.5 p-2 shadow-lg" @submit.prevent="applyLink">
                      <input ref="linkInput" v-model="linkUrl" type="text" inputmode="url" class="h-8 min-w-0 flex-1 rounded border border-line bg-base-100 px-2 text-sm outline-none focus:border-primary" :placeholder="t('publish.toolbar.linkUrl')" />
                      <button type="submit" class="gf-button gf-button-primary h-8 px-2.5" :disabled="!linkUrl.trim()">{{ t('publish.toolbar.applyLink') }}</button>
                    </form>
                  </div>
                  <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-200 hover:text-base-content" :title="t('publish.toolbar.quote')" @mousedown.prevent @click="applyToolbarAction('quote')"><MessageSquareQuote class="h-4 w-4" /></button>
                  <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-200 hover:text-base-content" :title="t('publish.toolbar.code')" @mousedown.prevent @click="applyToolbarAction('code')"><Code2 class="h-4 w-4" /></button>
                  <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-200 hover:text-base-content" :title="t('publish.toolbar.bulletList')" @mousedown.prevent @click="applyToolbarAction('bulletList')"><List class="h-4 w-4" /></button>
                  <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-200 hover:text-base-content" :title="t('publish.toolbar.orderedList')" @mousedown.prevent @click="applyToolbarAction('orderedList')"><ListOrdered class="h-4 w-4" /></button>
                  <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-200 hover:text-base-content" :title="t('publish.toolbar.horizontalRule')" @mousedown.prevent @click="applyToolbarAction('horizontalRule')"><Minus class="h-4 w-4" /></button>
                  <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-200 hover:text-base-content" :title="t('publish.toolbar.hardBreak')" @mousedown.prevent @click="applyToolbarAction('hardBreak')"><CornerDownLeft class="h-4 w-4" /></button>
                  <div ref="tablePicker" class="relative">
                    <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-200 hover:text-base-content" :title="t('publish.toolbar.table')" :aria-expanded="tablePickerOpen" @mousedown.prevent @click="openTablePicker"><Table2 class="h-4 w-4" /></button>
                    <div
                      v-if="tablePickerOpen"
                      class="gf-menu-surface absolute left-0 top-full z-30 mt-1.5 p-2.5 shadow-lg"
                      role="menu"
                      @keydown.esc.stop="tablePickerOpen = false"
                    >
                      <div class="mb-2 text-center text-xs font-medium text-base-content/75">
                        {{ t('publish.toolbar.tableSize', { rows: tablePickerRows, columns: tablePickerColumns }) }}
                      </div>
                      <div class="grid gap-1" :style="{ gridTemplateColumns: `repeat(${tablePickerMaxColumns}, 1.25rem)` }">
                        <button
                          v-for="cell in tablePickerCells"
                          :key="`${cell.row}-${cell.column}`"
                          type="button"
                          class="h-5 w-5 rounded-[2px] border transition-colors"
                          :class="cell.row <= tablePickerRows && cell.column <= tablePickerColumns ? 'border-primary bg-primary/25' : 'border-line bg-base-100 hover:border-primary/60'"
                          :aria-label="t('publish.toolbar.tableSize', { rows: cell.row, columns: cell.column })"
                          @mouseenter="selectTableSize(cell.row, cell.column)"
                          @focus="selectTableSize(cell.row, cell.column)"
                          @click="insertTable(cell.row, cell.column)"
                        />
                      </div>
                    </div>
                  </div>
                  <span class="mx-1 h-5 w-px bg-line" />
                  <button type="button" class="rounded p-1.5 text-base-content/55 hover:bg-base-200 hover:text-base-content" :title="t('publish.pastePlainText')" @mousedown.prevent @click="pastePlainText"><ClipboardPaste class="h-4 w-4" /></button>
                  <span v-if="uploadText" class="gf-badge gf-badge-info rounded">{{ uploadText }}</span>
                  <label class="rounded p-1.5 text-base-content/55 transition hover:bg-base-200 hover:text-base-content" :title="t('publish.uploadImageTitle')">
                    <Image class="h-4 w-4" />
                    <input type="file" accept="image/*" multiple class="hidden" :disabled="uploading" @change="handleImage" />
                  </label>
                </div>

                <div class="ml-auto flex items-center gap-1.5 text-xs font-semibold">
                  <div class="inline-flex rounded-md border border-line p-0.5">
                    <button type="button" class="rounded px-2 py-1" :class="editorMode === 'visual' ? 'bg-neutral text-neutral-content' : 'text-base-content/55 hover:text-base-content'" @click="selectEditorMode('visual')">{{ t('publish.visualMode') }}</button>
                    <button type="button" class="rounded px-2 py-1" :class="editorMode === 'markdown' ? 'bg-neutral text-neutral-content' : 'text-base-content/55 hover:text-base-content'" @click="selectEditorMode('markdown')">{{ t('publish.markdownMode') }}</button>
                  </div>
                  <button type="button" class="inline-flex items-center gap-1 rounded-md border border-line px-2 py-1.5" :class="preview ? 'bg-neutral text-neutral-content' : 'text-base-content/55 hover:bg-base-200 hover:text-base-content'" @click="togglePreview">
                    <Eye class="h-3.5 w-3.5" />
                    {{ t('publish.preview') }}
                  </button>
                </div>
              </div>

              <div
                v-show="!preview"
                :class="[
                  'min-h-80 bg-transparent',
                  dragOver ? 'bg-info/10 ring-1 ring-inset ring-primary shadow-[0_0_0_4px_rgba(59,130,246,0.12)]' : '',
                ]"
              >
                <div class="relative">
                  <textarea
                    v-if="editorMode === 'markdown'"
                    ref="editor"
                    v-model="content"
                    class="block min-h-80 w-full resize-none border-0 bg-transparent px-1 py-4 text-[15px] leading-relaxed outline-none placeholder:text-base-content/45"
                    :placeholder="t('publish.bodyPlaceholder')"
                    @keydown="handleEditorKeydown"
                    @paste="handlePaste"
                    @drop="handleDrop"
                    @dragover="handleDragOver"
                    @dragleave="dragOver = false"
                  />
                  <VisualMarkdownEditor
                    v-else
                    ref="visualEditor"
                    v-model="content"
                    :placeholder="t('publish.visualPlaceholder')"
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

              <div v-if="preview && content.trim()" class="gf-prose gf-prose-post min-h-80 max-w-none px-1 py-4" v-html="renderedPreview" />
              <div v-else-if="preview" class="gf-prose gf-prose-post min-h-80 max-w-none px-1 py-4">
                <p class="text-sm text-base-content/55">{{ t('publish.emptyPreview') }}</p>
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
