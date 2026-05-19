<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { Bold, Code2, Image, Italic, Link, ListChecks, MessageSquareQuote, Send, X } from '@lucide/vue'
import MarkdownIt from 'markdown-it'
import anchor from 'markdown-it-anchor'
import taskLists from 'markdown-it-task-lists'
import { submitArticle, uploadImage } from '@/runtime/api'
import { processImageFile, validateImageFile } from '@/runtime/image'
import type { LayoutPayload, PublishPageProps } from '@/types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: PublishPageProps
}>()

const title = ref(page.props.article.title || '')
const content = ref(page.props.article.content || '')
const type = ref(page.props.article.type || page.props.types[0]?.value || 0)
const categoryIds = ref<number[]>([...(page.props.article.categoryIds || [])])
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

const isValid = computed(() => title.value.trim() && content.value.trim() && categoryIds.value.length > 0)
const selectedCategories = computed(() => page.props.categories.filter((category) => categoryIds.value.includes(category.id)))
const renderedPreview = computed(() => markdown.render(content.value || ''))
const uploadText = computed(() => {
  if (!uploading.value) return ''
  return uploadTotal.value > 1 ? `正在处理图片 ${uploadDone.value}/${uploadTotal.value}` : '正在处理图片...'
})

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
        failed.push(`${file.name}: ${err instanceof Error ? err.message : '图片上传失败'}`)
      } finally {
        uploadDone.value += 1
      }
    }

    if (markdownImages.length) {
      insertMarkdownBlock(markdownImages.join('\n'))
      message.value = markdownImages.length > 1 ? `已插入 ${markdownImages.length} 张图片。` : '图片已插入。'
    }

    if (failed.length) {
      error.value = failed.slice(0, 3).join('；') + (failed.length > 3 ? `；另有 ${failed.length - 3} 张失败` : '')
    } else if (!markdownImages.length) {
      error.value = '没有可上传的图片文件'
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
    insert('**', '**', '加粗文本')
  } else if (key === 'i') {
    event.preventDefault()
    insert('*', '*', '斜体文本')
  } else if (key === 'k') {
    event.preventDefault()
    insert('[', '](https://)', '链接文本')
  }
}

async function save() {
  if (!isValid.value || submitting.value) return
  submitting.value = true
  error.value = ''
  message.value = ''
  try {
    const id = await submitArticle({
      id: page.props.articleId,
      title: title.value.trim(),
      content: content.value.trim(),
      type: type.value,
      categoryId: categoryIds.value,
    })
    message.value = page.props.isEditing ? '主题已更新。' : '主题已发布。'
    window.location.href = `/p/post/${id}`
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存失败'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
    <main class="min-w-0 pb-12">
      <header class="mb-4 border-b border-gray-200/70 pb-4">
        <h1 class="text-2xl font-bold text-gray-950">{{ props.isEditing ? '编辑主题' : '发布主题' }}</h1>
        <p class="mt-1 text-sm text-gray-500">写清楚标题，选择合适分类，让讨论更容易被找到。</p>
      </header>

      <div class="grid gap-3 xl:grid-cols-[minmax(0,1fr)_280px]">
        <section class="rounded-lg border border-gray-200/70 bg-white p-4 shadow-[0_2px_8px_rgba(0,0,0,0.02)] sm:p-5">
          <div class="space-y-5">
            <label class="block">
              <span class="text-sm font-semibold text-gray-700">标题</span>
              <input
                v-model="title"
                class="mt-1 h-11 w-full rounded-md border border-gray-200 px-3 text-lg font-semibold outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100"
                placeholder="输入主题标题"
              />
            </label>

            <div>
              <div class="mb-2 text-sm font-semibold text-gray-700">类型</div>
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="item in props.types"
                  :key="item.value"
                  type="button"
                  class="rounded-md border px-3 py-1.5 text-sm font-medium transition"
                  :class="type === item.value ? 'border-blue-500 bg-blue-50 text-blue-700' : 'border-gray-200 text-gray-600 hover:border-gray-300 hover:bg-gray-50'"
                  @click="type = item.value"
                >
                  {{ item.name }}
                </button>
              </div>
            </div>

            <div>
              <div class="mb-2 flex items-center justify-between">
                <span class="text-sm font-semibold text-gray-700">分类</span>
                <span class="text-xs text-gray-400">最多 3 个</span>
              </div>
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="category in props.categories"
                  :key="category.id"
                  type="button"
                  class="inline-flex items-center gap-1.5 rounded-md border px-2.5 py-1.5 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-40"
                  :class="categoryIds.includes(category.id) ? 'border-blue-500 bg-blue-50 text-blue-700' : 'border-gray-200 text-gray-600 hover:border-gray-300 hover:bg-gray-50'"
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
                <span class="text-sm font-semibold text-gray-700">正文</span>
                <div class="inline-flex rounded-md border border-gray-200 p-0.5 text-xs font-semibold">
                  <button type="button" class="rounded px-2 py-1" :class="!preview ? 'bg-gray-900 text-white' : 'text-gray-500'" @click="preview = false">编辑</button>
                  <button type="button" class="rounded px-2 py-1" :class="preview ? 'bg-gray-900 text-white' : 'text-gray-500'" @click="preview = true">预览</button>
                </div>
              </div>

              <div
                v-if="!preview"
                class="overflow-hidden rounded-lg border transition"
                :class="dragOver ? 'border-blue-500 bg-blue-50/50 shadow-[0_0_0_4px_rgba(59,130,246,0.12)]' : 'border-gray-200 bg-white'"
              >
                <div class="flex flex-wrap items-center gap-1 border-b border-gray-100 bg-gray-50 px-2 py-2">
                  <button type="button" class="rounded p-1.5 text-gray-500 hover:bg-white hover:text-gray-900" title="加粗" @click="insert('**', '**', '加粗文本')"><Bold class="h-4 w-4" /></button>
                  <button type="button" class="rounded p-1.5 text-gray-500 hover:bg-white hover:text-gray-900" title="斜体" @click="insert('*', '*', '斜体文本')"><Italic class="h-4 w-4" /></button>
                  <button type="button" class="rounded p-1.5 text-gray-500 hover:bg-white hover:text-gray-900" title="链接" @click="insert('[', '](https://)', '链接文本')"><Link class="h-4 w-4" /></button>
                  <button type="button" class="rounded p-1.5 text-gray-500 hover:bg-white hover:text-gray-900" title="引用" @click="insert('\\n> ', '', '引用内容')"><MessageSquareQuote class="h-4 w-4" /></button>
                  <button type="button" class="rounded p-1.5 text-gray-500 hover:bg-white hover:text-gray-900" title="代码" @click="insert('\\n```\\n', '\\n```\\n', 'code')"><Code2 class="h-4 w-4" /></button>
                  <span v-if="uploadText" class="ml-auto rounded bg-blue-50 px-2 py-1 text-xs font-semibold text-blue-700">{{ uploadText }}</span>
                  <label
                    class="rounded p-1.5 text-gray-500 transition hover:bg-white hover:text-gray-900"
                    :class="uploadText ? '' : 'ml-auto'"
                    title="上传图片，支持多选、粘贴和拖拽"
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
                    placeholder="输入正文，支持 Markdown；可粘贴或拖拽图片"
                    @keydown="handleEditorKeydown"
                    @paste="handlePaste"
                    @drop="handleDrop"
                    @dragover="handleDragOver"
                    @dragleave="dragOver = false"
                  />
                  <div
                    v-if="dragOver"
                    class="pointer-events-none absolute inset-3 grid place-items-center rounded-lg border-2 border-dashed border-blue-400 bg-blue-50/80 text-sm font-semibold text-blue-700"
                  >
                    松开后上传并插入图片
                  </div>
                </div>
              </div>

              <div v-else class="gf-prose gf-prose-article min-h-96 rounded-lg border border-gray-200 bg-gray-50/50 p-5">
                <div v-if="content.trim()" v-html="renderedPreview" />
                <p v-else class="text-sm text-gray-400">还没有可预览的内容。</p>
              </div>
            </div>

            <p v-if="error" class="rounded-md bg-red-50 px-3 py-2 text-sm text-red-600">{{ error }}</p>
            <p v-if="message" class="rounded-md bg-green-50 px-3 py-2 text-sm text-green-700">{{ message }}</p>

            <div class="flex items-center justify-end gap-2 border-t border-gray-100 pt-4">
              <a href="/" class="inline-flex h-10 items-center rounded-md px-3 text-sm font-semibold text-gray-500 hover:bg-gray-100 hover:text-gray-900">取消</a>
              <button
                type="button"
                class="inline-flex h-10 items-center gap-1.5 rounded-md bg-blue-600 px-4 text-sm font-semibold text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="!isValid || submitting || uploading"
                @click="save"
              >
                <Send class="h-4 w-4" />
                {{ submitting ? '保存中...' : props.isEditing ? '更新主题' : '发布主题' }}
              </button>
            </div>
          </div>
        </section>

        <aside class="space-y-3">
          <section class="rounded-lg border border-gray-200/70 bg-white p-4 shadow-[0_2px_8px_rgba(0,0,0,0.02)]">
            <div class="flex items-center gap-2">
              <ListChecks class="h-4 w-4 text-gray-400" />
              <h2 class="text-sm font-semibold text-gray-950">发布检查</h2>
            </div>
            <ul class="mt-3 space-y-2 text-sm text-gray-600">
              <li class="flex items-center justify-between gap-3"><span>标题</span><span :class="title.trim() ? 'text-green-600' : 'text-gray-400'">{{ title.trim() ? '已填写' : '待填写' }}</span></li>
              <li class="flex items-center justify-between gap-3"><span>分类</span><span :class="categoryIds.length ? 'text-green-600' : 'text-gray-400'">{{ categoryIds.length }}/3</span></li>
              <li class="flex items-center justify-between gap-3"><span>正文</span><span :class="content.trim() ? 'text-green-600' : 'text-gray-400'">{{ content.trim().length }} 字</span></li>
            </ul>
          </section>

          <section v-if="selectedCategories.length" class="rounded-lg border border-gray-200/70 bg-white p-4 shadow-[0_2px_8px_rgba(0,0,0,0.02)]">
            <h2 class="text-sm font-semibold text-gray-950">已选分类</h2>
            <div class="mt-3 flex flex-wrap gap-2">
              <button
                v-for="category in selectedCategories"
                :key="category.id"
                type="button"
                class="inline-flex items-center gap-1.5 rounded-md border border-gray-200 px-2 py-1 text-sm text-gray-600 hover:bg-gray-50"
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
    </main>
</template>
