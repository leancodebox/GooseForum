<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, Teleport, watch } from 'vue'
import { AlertTriangle, Bookmark, Check, Clock, CornerDownLeft, Eye, Heart, Loader2, MessageSquare, PencilLine, Send, Trash2, X } from '@lucide/vue'
import { bookmarkArticle, deleteReply, getArticleRepliesWindow, likeArticle, postReply, updateReply } from '@/runtime/api'
import { formatDateTime, formatNumber } from '@/runtime/format'
import { fetchPage } from '@/runtime/router'
import { useShellState } from '@/runtime/shell-state'
import { showUserCard } from '@/runtime/user-card-events'
import type { ArticleDetailProps, LayoutPayload, ReplyPayload } from '@/types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: ArticleDetailProps
}>()

const replyContent = ref('')
const replyContents = reactive<Record<number, string>>({})
const openReplyId = ref<number | null>(null)
const currentReplyId = ref(0)
const likeCount = ref(page.props.article.likeCount)
const isLiked = ref(page.props.article.isLiked)
const isBookmarked = ref(page.props.article.isBookmarked)
const actionMessage = ref('')
const actingLike = ref(false)
const actingBookmark = ref(false)
const submitting = ref(false)
const deletingReplyId = ref(0)
const editingReplyId = ref(0)
const savingEditReplyId = ref(0)
const pendingDeleteReply = ref<ReplyPayload | null>(null)
const replies = ref<ReplyPayload[]>([...page.props.replies])
const replyWindowMode = ref(false)
const replyHasBefore = ref(false)
const replyHasAfter = ref(hasMoreInitialReplies())
const replyBeforeCursor = ref(firstReplyId(page.props.replies))
const replyAfterCursor = ref(lastReplyId(page.props.replies))
const loadingReplyWindow = ref(false)
const loadingReplyDirection = ref<'before' | 'after' | 'anchor' | null>(null)
const replyWindowError = ref('')
const deleteErrorMessage = ref('')
const errorMessage = ref('')
const successMessage = ref('')
const inlineReplyErrors = reactive<Record<number, string>>({})
const editReplyContents = reactive<Record<number, string>>({})
const editReplyErrors = reactive<Record<number, string>>({})
const titleEl = ref<HTMLElement | null>(null)
const replyEditorEl = ref<HTMLTextAreaElement | null>(null)
const replySectionEl = ref<HTMLElement | null>(null)
const replyLoadMoreEl = ref<HTMLElement | null>(null)
const showHeaderTitle = ref(false)
const isMobileHeaderViewport = ref(false)
const mobileHeaderTitleVisible = ref(false)
const effectiveShowHeaderTitle = computed(() => showHeaderTitle.value && (!isMobileHeaderViewport.value || mobileHeaderTitleVisible.value))
const showFloatingReply = ref(false)
const floatingReplyExpanded = ref(false)
const shellState = useShellState()
let titleObserver: IntersectionObserver | undefined
let replyEditorObserver: IntersectionObserver | undefined
let replyLoadObserver: IntersectionObserver | undefined
let lastHeaderScrollY = 0
let headerScrollFrame = 0
const highlightedReplyId = ref<number | null>(null)
let highlightTimer: number | undefined

function observeTitle() {
  titleObserver?.disconnect()
  showHeaderTitle.value = false

  if (!titleEl.value || !('IntersectionObserver' in window)) return

  titleObserver = new IntersectionObserver(
    (entries) => {
      showHeaderTitle.value = !entries[0]?.isIntersecting
    },
    { threshold: 0, rootMargin: '-80px 0px 0px 0px' },
  )
  titleObserver.observe(titleEl.value)
}

onMounted(() => {
  setupHeaderTitleBehavior()
  void nextTick(observeTitle)
  void nextTick(observeReplyEditor)
  void nextTick(observeReplyLoader)
  void syncReplyHash()
})

watch(
  () => page.props.article.id,
  () => {
    likeCount.value = page.props.article.likeCount
    isLiked.value = page.props.article.isLiked
    isBookmarked.value = page.props.article.isBookmarked
    mobileHeaderTitleVisible.value = false
    if (typeof window !== 'undefined') {
      lastHeaderScrollY = window.scrollY
    }
    resetRepliesFromProps()
    void nextTick(observeTitle)
    void nextTick(observeReplyEditor)
    void nextTick(observeReplyLoader)
    void nextTick(syncReplyHash)
  },
  { immediate: true },
)

watch(
  () => [page.props.article.title, page.props.article.categories, effectiveShowHeaderTitle.value] as const,
  ([title, categories, show]) => {
    shellState.headerTitle = title
    shellState.headerTags = categories.map((category) => ({
      id: category.id,
      name: category.name,
      color: category.color,
    }))
    shellState.showHeaderTitle = show
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  titleObserver?.disconnect()
  replyEditorObserver?.disconnect()
  replyLoadObserver?.disconnect()
  window.removeEventListener('scroll', updateMobileHeaderTitle)
  window.removeEventListener('resize', updateHeaderViewport)
  window.cancelAnimationFrame(headerScrollFrame)
  window.clearTimeout(highlightTimer)
  shellState.headerTitle = ''
  shellState.headerTags = []
  shellState.showHeaderTitle = false
})

function setupHeaderTitleBehavior() {
  lastHeaderScrollY = window.scrollY
  updateHeaderViewport()
  window.addEventListener('scroll', updateMobileHeaderTitle, { passive: true })
  window.addEventListener('resize', updateHeaderViewport)
}

function updateHeaderViewport() {
  const wasMobile = isMobileHeaderViewport.value
  const isMobile = window.innerWidth < 768
  isMobileHeaderViewport.value = isMobile
  if (isMobile && !wasMobile) {
    mobileHeaderTitleVisible.value = false
    return
  }
  if (!isMobile) {
    mobileHeaderTitleVisible.value = true
  }
}

function updateMobileHeaderTitle() {
  if (headerScrollFrame) return
  headerScrollFrame = window.requestAnimationFrame(applyMobileHeaderTitle)
}

function applyMobileHeaderTitle() {
  headerScrollFrame = 0
  const scrollY = window.scrollY
  const delta = scrollY - lastHeaderScrollY
  if (Math.abs(delta) < 4) {
    return
  }

  if (isMobileHeaderViewport.value) {
    mobileHeaderTitleVisible.value = delta > 0
  }
  lastHeaderScrollY = scrollY
}

function observeReplyEditor() {
  replyEditorObserver?.disconnect()
  showFloatingReply.value = false
  if (!replySectionEl.value || !page.props.permissions.canReply || !('IntersectionObserver' in window)) return

  replyEditorObserver = new IntersectionObserver(
    (entries) => {
      showFloatingReply.value = !entries[0]?.isIntersecting
    },
    { threshold: 0.08, rootMargin: '0px 0px -96px 0px' },
  )
  replyEditorObserver.observe(replySectionEl.value)
}

function observeReplyLoader() {
  replyLoadObserver?.disconnect()
  if (!replyLoadMoreEl.value || !replyHasAfter.value || !('IntersectionObserver' in window)) return

  replyLoadObserver = new IntersectionObserver(
    (entries) => {
      if (entries[0]?.isIntersecting && replyHasAfter.value && !loadingReplyWindow.value && !replyWindowError.value) {
        void loadReplyWindow('after')
      }
    },
    { rootMargin: '360px 0px' },
  )
  replyLoadObserver.observe(replyLoadMoreEl.value)
}

function resetRepliesFromProps() {
  replies.value = [...page.props.replies]
  replyWindowMode.value = false
  replyHasBefore.value = false
  replyHasAfter.value = hasMoreInitialReplies()
  replyBeforeCursor.value = firstReplyId(page.props.replies)
  replyAfterCursor.value = lastReplyId(page.props.replies)
  replyWindowError.value = ''
  editingReplyId.value = 0
}

function firstReplyId(items: ReplyPayload[]) {
  return items.length ? items[0].id : 0
}

function lastReplyId(items: ReplyPayload[]) {
  return items.length ? items[items.length - 1].id : 0
}

function hasMoreInitialReplies() {
  return page.props.article.replyCount > page.props.replies.length
}

function findReplyHashId() {
  const match = window.location.hash.match(/^#reply-(\d+)$/)
  return match ? Number(match[1]) : 0
}

async function syncReplyHash() {
  const replyId = findReplyHashId()
  if (!replyId) return

  if (!replies.value.some((reply) => reply.id === replyId)) {
    await loadReplyWindow('anchor', replyId)
  }

  highlightReply(replyId)
  await nextTick()
  document.getElementById(`reply-${replyId}`)?.scrollIntoView({ block: 'center' })
}

function highlightReply(replyId: number) {
  highlightedReplyId.value = replyId
  window.clearTimeout(highlightTimer)
  highlightTimer = window.setTimeout(() => {
    highlightedReplyId.value = null
  }, 2400)
}

function mergeReplies(nextReplies: ReplyPayload[], mode: 'replace' | 'prepend' | 'append') {
  if (mode === 'replace') {
    replies.value = nextReplies
    return
  }

  const seen = new Set(replies.value.map((reply) => reply.id))
  const filtered = nextReplies.filter((reply) => !seen.has(reply.id))
  replies.value = mode === 'prepend' ? [...filtered, ...replies.value] : [...replies.value, ...filtered]
}

async function loadReplyWindow(direction: 'before' | 'after' | 'anchor', anchorReplyId = 0) {
  if (loadingReplyWindow.value) return

  const wasWindowMode = replyWindowMode.value
  loadingReplyWindow.value = true
  loadingReplyDirection.value = direction
  replyWindowError.value = ''
  try {
    const payload = await getArticleRepliesWindow({
      articleId: page.props.article.id,
      anchorReplyId: direction === 'anchor' ? anchorReplyId : undefined,
      before: direction === 'before' ? replyBeforeCursor.value : undefined,
      after: direction === 'after' ? replyAfterCursor.value : undefined,
      limit: 20,
    })

    replyWindowMode.value = direction === 'anchor' || direction === 'before' || wasWindowMode
    mergeReplies(payload.replies, direction === 'before' ? 'prepend' : direction === 'after' ? 'append' : 'replace')
    replyHasBefore.value = replyWindowMode.value ? payload.hasBefore : false
    replyHasAfter.value = payload.hasAfter
    replyBeforeCursor.value = payload.beforeCursor ?? firstReplyId(replies.value)
    replyAfterCursor.value = payload.afterCursor ?? lastReplyId(replies.value)
    await nextTick()
    observeReplyLoader()
  } catch (error) {
    replyWindowError.value = error instanceof Error ? error.message : '回复加载失败'
  } finally {
    loadingReplyWindow.value = false
    loadingReplyDirection.value = null
  }
}

function focusReplyEditor() {
  replyEditorEl.value?.focus()
  replyEditorEl.value?.scrollIntoView({ block: 'center' })
}

function openFloatingReply() {
  floatingReplyExpanded.value = true
  openReplyId.value = null
}

function closeFloatingReply() {
  floatingReplyExpanded.value = false
}

function isElementMostlyVisible(element: HTMLElement) {
  const rect = element.getBoundingClientRect()
  return rect.top >= 96 && rect.bottom <= window.innerHeight - 120
}

function scrollReplyIntoComfortView(element: HTMLElement) {
  const targetTop = element.getBoundingClientRect().top + window.scrollY - 160
  window.scrollTo({
    top: Math.max(0, targetTop),
    behavior: 'smooth',
  })
}

async function revealCreatedReply(reply: ReplyPayload) {
  if (!reply.id) return

  mergeReplies([reply], 'append')
  replyHasAfter.value = false
  replyAfterCursor.value = Math.max(replyAfterCursor.value, reply.id)
  await nextTick()
  highlightReply(reply.id)
  const element = document.getElementById(`reply-${reply.id}`)
  if (element && !isElementMostlyVisible(element)) {
    scrollReplyIntoComfortView(element)
  }
}

function buildCreatedReply(replyId: number, content: string, renderedContent: string, replyToId: number): ReplyPayload {
  const parentReply = replyToId > 0 ? replies.value.find((reply) => reply.id === replyToId) : undefined
  const viewer = page.layout.viewer
  return {
    id: replyId,
    articleId: page.props.article.id,
    content,
    renderedContent,
    author: {
      id: viewer.id,
      username: viewer.username,
      avatarUrl: viewer.avatarUrl,
    },
    createdAt: new Date().toISOString().slice(0, 19).replace('T', ' '),
    replyToId: replyToId || undefined,
    replyToUserId: parentReply?.author.id,
    replyToUsername: parentReply?.author.username,
    isOwnReply: true,
    updatedAt: new Date().toISOString().slice(0, 19).replace('T', ' '),
  }
}

function escapePlainText(content: string) {
  return content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
    .replace(/\r\n/g, '\n')
    .replace(/\r/g, '\n')
    .replace(/\n/g, '<br>\n')
}

async function fetchAndRevealCreatedReply(replyId: number) {
  if (!replyId) return

  const payload = await getArticleRepliesWindow({
    articleId: page.props.article.id,
    anchorReplyId: replyId,
    limit: 3,
  })
  mergeReplies(payload.replies, 'append')
  replyHasAfter.value = payload.hasAfter
  replyAfterCursor.value = payload.afterCursor ?? lastReplyId(replies.value)
  await nextTick()
  highlightReply(replyId)
  const element = document.getElementById(`reply-${replyId}`)
  if (element && !isElementMostlyVisible(element)) {
    scrollReplyIntoComfortView(element)
  }
}

async function toggleLike() {
  if (actingLike.value) return

  const nextLiked = !isLiked.value
  const previousLiked = isLiked.value
  const previousCount = likeCount.value
  actingLike.value = true
  actionMessage.value = ''
  isLiked.value = nextLiked
  likeCount.value = Math.max(0, likeCount.value + (nextLiked ? 1 : -1))
  try {
    await likeArticle(page.props.article.id, nextLiked ? 1 : 2)
  } catch (error) {
    isLiked.value = previousLiked
    likeCount.value = previousCount
    actionMessage.value = error instanceof Error ? error.message : '点赞失败'
  } finally {
    actingLike.value = false
  }
}

async function toggleBookmark() {
  if (actingBookmark.value) return

  const nextBookmarked = !isBookmarked.value
  const previousBookmarked = isBookmarked.value
  actingBookmark.value = true
  actionMessage.value = ''
  isBookmarked.value = nextBookmarked
  try {
    await bookmarkArticle(page.props.article.id, nextBookmarked ? 1 : 2)
    actionMessage.value = nextBookmarked ? '已收藏。' : '已取消收藏。'
  } catch (error) {
    isBookmarked.value = previousBookmarked
    actionMessage.value = error instanceof Error ? error.message : '收藏失败'
  } finally {
    actingBookmark.value = false
  }
}

function toggleReplyForm(replyId: number) {
  openReplyId.value = openReplyId.value === replyId ? null : replyId
  editingReplyId.value = 0
  if (replyContents[replyId] === undefined) {
    replyContents[replyId] = ''
  }
  inlineReplyErrors[replyId] = ''
}

function clearReplyValidation(replyId = 0) {
  if (replyId > 0) {
    inlineReplyErrors[replyId] = ''
    return
  }

  errorMessage.value = ''
  successMessage.value = ''
}

function startEditReply(reply: ReplyPayload) {
  openReplyId.value = null
  editingReplyId.value = reply.id
  editReplyContents[reply.id] = reply.content
  editReplyErrors[reply.id] = ''
}

function cancelEditReply(replyId: number) {
  if (savingEditReplyId.value) return
  editingReplyId.value = 0
  editReplyErrors[replyId] = ''
}

function clearEditReplyValidation(replyId: number) {
  editReplyErrors[replyId] = ''
}

async function saveReplyEdit(reply: ReplyPayload) {
  if (savingEditReplyId.value) return

  const content = (editReplyContents[reply.id] || '').trim()
  if (!content) {
    editReplyErrors[reply.id] = '回复内容不能为空'
    return
  }
  if (content === reply.content.trim()) {
    editingReplyId.value = 0
    editReplyErrors[reply.id] = ''
    return
  }

  savingEditReplyId.value = reply.id
  editReplyErrors[reply.id] = ''
  try {
    const updated = await updateReply(reply.id, content)
    const index = replies.value.findIndex((item) => item.id === reply.id)
    if (index >= 0) {
      replies.value[index] = {
        ...replies.value[index],
        content: updated.content,
        renderedContent: updated.renderedContent,
        updatedAt: updated.updatedAt,
      }
    }
    editingReplyId.value = 0
    successMessage.value = '回复已更新。'
  } catch (error) {
    editReplyErrors[reply.id] = error instanceof Error ? error.message : '更新回复失败'
  } finally {
    savingEditReplyId.value = 0
  }
}

async function submitReply(replyId = 0) {
  const content = replyId > 0 ? (replyContents[replyId] || '').trim() : replyContent.value.trim()
  if (submitting.value) return

  if (!content) {
    if (replyId > 0) {
      inlineReplyErrors[replyId] = '回复内容不能为空'
    } else {
      errorMessage.value = '回复内容不能为空'
      successMessage.value = ''
    }
    return
  }

  submitting.value = true
  currentReplyId.value = replyId
  errorMessage.value = ''
  successMessage.value = ''
  if (replyId > 0) {
    inlineReplyErrors[replyId] = ''
  }
  try {
    const createdReply = await postReply(page.props.article.id, content, replyId)
    if (replyId > 0) {
      replyContents[replyId] = ''
      openReplyId.value = null
    } else {
      replyContent.value = ''
      closeFloatingReply()
      successMessage.value = '回复已发布。'
    }
    const createdReplyId = typeof createdReply === 'object' && createdReply !== null ? createdReply.id : createdReply
    const renderedContent = typeof createdReply === 'object' && createdReply !== null ? createdReply.renderedContent : escapePlainText(content)
    if (typeof createdReplyId === 'number') {
      if (page.layout.viewer.isAuthenticated) {
        await revealCreatedReply(buildCreatedReply(createdReplyId, content, renderedContent, replyId))
      } else {
        await fetchAndRevealCreatedReply(createdReplyId)
      }
    } else {
      await refreshCurrentPage()
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : '回复失败'
    if (replyId > 0) {
      inlineReplyErrors[replyId] = message
    } else {
      errorMessage.value = message
    }
  } finally {
    submitting.value = false
    currentReplyId.value = 0
  }
}

async function refreshCurrentPage() {
  const payload = await fetchPage(new URL(window.location.href))
  history.replaceState({ goose: true, payload }, '', window.location.href)
  window.dispatchEvent(new CustomEvent('goose:page', { detail: payload }))
}

function requestDeleteReply(reply: ReplyPayload) {
  if (savingEditReplyId.value === reply.id) return
  pendingDeleteReply.value = reply
  deleteErrorMessage.value = ''
}

function closeDeleteDialog() {
  if (deletingReplyId.value) return
  pendingDeleteReply.value = null
  deleteErrorMessage.value = ''
}

async function removeReply(replyId: number) {
  if (deletingReplyId.value || savingEditReplyId.value === replyId) return

  deletingReplyId.value = replyId
  errorMessage.value = ''
  successMessage.value = ''
  deleteErrorMessage.value = ''
  try {
    await deleteReply(replyId)
    successMessage.value = '回复已删除。'
    pendingDeleteReply.value = null
    await refreshCurrentPage()
  } catch (error) {
    deleteErrorMessage.value = error instanceof Error ? error.message : '删除回复失败'
  } finally {
    deletingReplyId.value = 0
  }
}
</script>

<template>
  <div>
    <article class="min-w-0 pb-12">
      <header class="mb-4 border-b border-gray-200/70 pb-4">
        <h1 ref="titleEl" class="text-2xl font-bold leading-tight text-gray-950 sm:text-3xl">{{ page.props.article.title }}</h1>
        <div class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2 text-[13px] text-gray-500">
          <a
            :href="`/u/${page.props.article.author.id}`"
            class="inline-flex items-center gap-2 font-medium text-gray-700 hover:text-blue-600"
            @click="showUserCard(page.props.article.author, $event)"
          >
            <img :src="page.props.article.author.avatarUrl" :alt="page.props.article.author.username" class="h-5 w-5 rounded-full object-cover" />
            {{ page.props.article.author.username }}
          </a>
          <span class="inline-flex items-center gap-1.5">
            <Clock class="h-3.5 w-3.5" />
            {{ formatDateTime(page.props.article.createdAt) }}
          </span>
          <a
            v-for="category in page.props.article.categories"
            :key="category.id"
            :href="category.url"
            class="inline-flex items-center gap-1.5 rounded-sm text-gray-600 hover:text-blue-600"
          >
            <span class="h-2 w-2 rounded-[3px]" :style="{ backgroundColor: category.color }" />
            {{ category.name }}
          </a>
        </div>
      </header>

      <section class="rounded-lg border border-gray-200/70 bg-white shadow-[0_2px_8px_rgba(0,0,0,0.02)]">
        <div class="grid grid-cols-[44px_minmax(0,1fr)] gap-3 p-4 sm:grid-cols-[52px_minmax(0,1fr)] sm:gap-4 sm:p-5">
          <a
            :href="`/u/${page.props.article.author.id}`"
            class="sticky top-19 self-start pt-1"
            @click="showUserCard(page.props.article.author, $event)"
          >
            <img :src="page.props.article.author.avatarUrl" :alt="page.props.article.author.username" class="h-11 w-11 rounded-full object-cover ring-1 ring-gray-100" />
          </a>
          <div class="min-w-0">
            <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
              <div>
                <a :href="`/u/${page.props.article.author.id}`" class="font-semibold text-gray-950 hover:text-blue-600">{{ page.props.article.author.username }}</a>
                <div class="text-xs font-medium text-gray-600">正文</div>
              </div>
              <div class="flex flex-wrap items-center justify-end gap-3 text-xs font-medium text-gray-600">
                <div class="flex items-center gap-3">
                  <span class="inline-flex items-center gap-1"><MessageSquare class="h-3.5 w-3.5" />{{ formatNumber(page.props.article.replyCount) }}</span>
                  <span class="inline-flex items-center gap-1"><Eye class="h-3.5 w-3.5" />{{ formatNumber(page.props.article.viewCount) }}</span>
                  <span class="inline-flex items-center gap-1"><Heart class="h-3.5 w-3.5" />{{ formatNumber(likeCount) }}</span>
                </div>
                <a
                  v-if="page.props.permissions.isOwnArticle"
                  :href="`/publish?id=${page.props.article.id}`"
                  class="inline-flex h-7 items-center gap-1.5 rounded-md border border-gray-200 bg-white px-2 text-xs font-semibold text-gray-600 transition hover:border-blue-200 hover:bg-blue-50 hover:text-blue-700"
                >
                  <PencilLine class="h-3.5 w-3.5" />
                  编辑
                </a>
              </div>
            </div>
            <div class="gf-prose gf-prose-article" v-html="page.props.article.html" />
            <div class="mt-6 flex flex-wrap items-center gap-3 border-t border-gray-100 pt-4">
              <button
                type="button"
                class="inline-flex h-8 items-center gap-1.5 rounded-md px-2.5 text-sm font-semibold transition disabled:cursor-not-allowed disabled:opacity-60"
                :class="isLiked ? 'bg-red-50 text-red-600 hover:bg-red-100' : 'text-gray-500 hover:bg-gray-50 hover:text-gray-800'"
                :disabled="actingLike"
                @click="toggleLike"
              >
                <Heart class="h-4 w-4" :fill="isLiked ? 'currentColor' : 'none'" />
                {{ likeCount ? formatNumber(likeCount) : '点赞' }}
              </button>
              <button
                type="button"
                class="inline-flex h-8 items-center gap-1.5 rounded-md px-2.5 text-sm font-semibold transition disabled:cursor-not-allowed disabled:opacity-60"
                :class="isBookmarked ? 'bg-blue-50 text-blue-700 hover:bg-blue-100' : 'text-gray-500 hover:bg-gray-50 hover:text-gray-800'"
                :disabled="actingBookmark"
                @click="toggleBookmark"
              >
                <Bookmark class="h-4 w-4" :fill="isBookmarked ? 'currentColor' : 'none'" />
                {{ isBookmarked ? '已收藏' : '收藏' }}
              </button>
              <span v-if="actionMessage" class="text-xs" :class="actionMessage.includes('失败') ? 'text-red-600' : 'text-gray-600'">{{ actionMessage }}</span>
            </div>
          </div>
        </div>

        <span v-if="replies.length" id="replies" class="block scroll-mt-20" aria-hidden="true" />

        <div v-if="replyHasBefore" class="border-t border-gray-100 px-4 py-3 text-center">
          <button
            v-if="replyHasBefore"
            type="button"
            class="inline-flex h-8 items-center gap-1.5 rounded-md px-3 text-xs font-semibold text-gray-600 transition hover:bg-gray-50 hover:text-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="loadingReplyWindow"
            @click="loadReplyWindow('before')"
          >
            <Loader2 v-if="loadingReplyDirection === 'before'" class="h-3.5 w-3.5 animate-spin" />
            加载更早回复
          </button>
        </div>

        <div
          v-for="reply in replies"
          :id="`reply-${reply.id}`"
          :key="reply.id"
          class="group grid scroll-mt-20 grid-cols-[40px_minmax(0,1fr)] gap-2.5 border-t border-gray-100 px-3 py-4 transition hover:bg-gray-50/70 sm:grid-cols-[52px_minmax(0,1fr)] sm:gap-4 sm:p-5"
          :class="{ 'bg-blue-50/50 ring-1 ring-inset ring-blue-100': highlightedReplyId === reply.id }"
        >
          <a
            :href="`/u/${reply.author.id}`"
            class="sticky top-19 self-start pt-1"
            @click="showUserCard(reply.author, $event)"
          >
            <img :src="reply.author.avatarUrl" :alt="reply.author.username" class="h-9 w-9 rounded-full object-cover ring-1 ring-gray-100 sm:h-10 sm:w-10" />
          </a>
          <div class="min-w-0">
            <div class="mb-1.5 flex min-w-0 items-start justify-between gap-2">
              <div class="min-w-0">
                <a :href="`/u/${reply.author.id}`" class="min-w-0 truncate font-semibold text-gray-950 hover:text-blue-600">{{ reply.author.username }}</a>
                <time class="mt-0.5 block truncate text-xs text-gray-400 sm:hidden">{{ formatDateTime(reply.createdAt) }}</time>
              </div>
              <div class="flex shrink-0 items-center gap-0.5 sm:gap-3">
                <button
                  v-if="page.props.permissions.canReply"
                  type="button"
                  class="inline-flex h-8 shrink-0 items-center gap-1.5 whitespace-nowrap rounded-md px-2 text-xs font-semibold text-gray-600 opacity-100 transition hover:bg-blue-50 hover:text-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 sm:h-9 sm:px-3 sm:opacity-0 sm:group-hover:opacity-100 sm:focus-visible:opacity-100"
                  title="回复"
                  @click="toggleReplyForm(reply.id)"
                >
                  <CornerDownLeft class="h-3.5 w-3.5" />
                  <span class="sr-only sm:not-sr-only">回复</span>
                </button>
                <button
                  v-if="reply.isOwnReply"
                  type="button"
                  class="inline-flex h-8 shrink-0 items-center gap-1.5 whitespace-nowrap rounded-md px-2 text-xs font-semibold text-gray-500 opacity-100 transition hover:bg-blue-50 hover:text-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 sm:h-9 sm:px-3 sm:opacity-0 sm:group-hover:opacity-100 sm:focus-visible:opacity-100"
                  :disabled="savingEditReplyId === reply.id || deletingReplyId === reply.id"
                  title="编辑"
                  @click="startEditReply(reply)"
                >
                  <PencilLine class="h-3.5 w-3.5" />
                  <span class="sr-only sm:not-sr-only">编辑</span>
                </button>
                <button
                  v-if="reply.isOwnReply"
                  type="button"
                  class="inline-flex h-8 shrink-0 items-center gap-1.5 whitespace-nowrap rounded-md px-2 text-xs font-semibold text-gray-500 opacity-100 transition hover:bg-red-50 hover:text-red-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-red-500 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 sm:h-9 sm:px-3 sm:opacity-0 sm:group-hover:opacity-100 sm:focus-visible:opacity-100"
                  :disabled="deletingReplyId === reply.id"
                  :title="deletingReplyId === reply.id ? '删除中...' : '删除'"
                  @click="requestDeleteReply(reply)"
                >
                  <Trash2 class="h-3.5 w-3.5" />
                  <span class="sr-only sm:not-sr-only">{{ deletingReplyId === reply.id ? '删除中...' : '删除' }}</span>
                </button>
                <time class="hidden shrink-0 text-xs text-gray-400 sm:block">{{ formatDateTime(reply.createdAt) }}</time>
              </div>
            </div>
            <p v-if="reply.replyToUsername" class="mb-1.5 inline-flex max-w-full items-center rounded bg-gray-50 px-2 py-1 text-sm text-gray-500">
              回复 <a :href="`/u/${reply.replyToUserId}`" class="font-medium text-gray-700 hover:text-blue-600">@{{ reply.replyToUsername }}</a>
            </p>
            <div v-if="editingReplyId === reply.id" class="mt-3 rounded-lg border border-blue-100 bg-blue-50/40 p-3">
              <div class="mb-2 flex items-center justify-between">
                <div class="text-xs font-semibold text-blue-700">编辑自己的回复</div>
                <button type="button" class="rounded-md p-1 text-gray-400 hover:bg-white hover:text-gray-700" @click="cancelEditReply(reply.id)">
                  <X class="h-4 w-4" />
                </button>
              </div>
              <textarea
                v-model="editReplyContents[reply.id]"
                class="min-h-24 w-full resize-y rounded-md border border-blue-100 bg-white p-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100"
                placeholder="修改你的回复..."
                @input="clearEditReplyValidation(reply.id)"
              />
              <p v-if="editReplyErrors[reply.id]" class="mt-2 text-sm text-red-600">{{ editReplyErrors[reply.id] }}</p>
              <div class="mt-2 flex justify-end gap-2">
                <button type="button" class="h-8 rounded-md px-3 text-xs font-semibold text-gray-500 hover:bg-white" @click="cancelEditReply(reply.id)">取消</button>
                <button
                  type="button"
                  class="inline-flex h-8 items-center gap-1.5 rounded-md bg-blue-600 px-3 text-xs font-semibold text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
                  :disabled="savingEditReplyId === reply.id"
                  @click="saveReplyEdit(reply)"
                >
                  <Loader2 v-if="savingEditReplyId === reply.id" class="h-3.5 w-3.5 animate-spin" />
                  <Check v-else class="h-3.5 w-3.5" />
                  {{ savingEditReplyId === reply.id ? '保存中...' : '保存' }}
                </button>
              </div>
            </div>
            <template v-else>
              <div class="gf-prose gf-prose-comment" v-html="reply.renderedContent" />
              <div v-if="reply.updatedAt && reply.updatedAt !== reply.createdAt" class="mt-2 text-xs font-medium text-gray-400">
                已编辑于 {{ formatDateTime(reply.updatedAt) }}
              </div>
            </template>

            <div v-if="openReplyId === reply.id" class="mt-4 border-l-2 border-blue-100 pl-3">
              <div class="mb-2 flex items-center justify-between">
                <div class="text-xs font-medium text-gray-500">回复 @{{ reply.author.username }}</div>
                <button type="button" class="rounded-md p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-700" @click="openReplyId = null">
                  <X class="h-4 w-4" />
                </button>
              </div>
              <textarea
                v-model="replyContents[reply.id]"
                class="min-h-20 w-full resize-y rounded-md border border-gray-200 bg-white p-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100"
                placeholder="写下你的回复..."
                @input="clearReplyValidation(reply.id)"
              />
              <p v-if="inlineReplyErrors[reply.id]" class="mt-2 text-sm text-red-600">{{ inlineReplyErrors[reply.id] }}</p>
              <div class="mt-2 flex justify-end gap-2">
                <button type="button" class="h-8 rounded-md px-3 text-xs font-semibold text-gray-500 hover:bg-gray-100" @click="openReplyId = null">取消</button>
                <button
                  type="button"
                  class="inline-flex h-8 items-center gap-1.5 rounded-md bg-blue-600 px-3 text-xs font-semibold text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
                  :disabled="submitting"
                  @click="submitReply(reply.id)"
                >
                  <Send class="h-3.5 w-3.5" />
                  {{ submitting && currentReplyId === reply.id ? '发布中...' : '回复' }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <div v-if="replyHasAfter || loadingReplyDirection === 'after' || replyWindowError || replies.length" ref="replyLoadMoreEl" class="border-t border-gray-100 px-4 py-3 text-center">
          <button
            v-if="replyHasAfter && replyWindowError"
            type="button"
            class="inline-flex h-8 items-center gap-1.5 rounded-md px-3 text-xs font-semibold text-gray-600 transition hover:bg-gray-50 hover:text-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="loadingReplyWindow"
            @click="loadReplyWindow('after')"
          >
            <Loader2 v-if="loadingReplyDirection === 'after'" class="h-3.5 w-3.5 animate-spin" />
            重试加载回复
          </button>
          <p v-else-if="replyWindowError" class="text-xs text-red-600">{{ replyWindowError }}</p>
          <p v-else-if="replyHasAfter && loadingReplyDirection === 'after'" class="inline-flex items-center justify-center gap-1.5 text-xs font-medium text-gray-500">
            <Loader2 class="h-3.5 w-3.5 animate-spin" />
            正在加载更多回复...
          </p>
          <p v-else-if="!replyHasAfter && replies.length" class="text-xs font-medium text-gray-400">已显示全部回复</p>
        </div>
      </section>

      <section ref="replySectionEl" class="mt-4 rounded-lg border border-gray-200/70 bg-white p-4 shadow-[0_2px_8px_rgba(0,0,0,0.02)] sm:p-5">
        <template v-if="page.props.permissions.canReply">
          <div class="mb-3 flex items-center justify-between">
            <label class="text-sm font-semibold text-gray-950" for="reply-content">参与讨论</label>
            <span class="text-xs text-gray-400">Markdown 支持稍后补齐</span>
          </div>
          <textarea
            id="reply-content"
            ref="replyEditorEl"
            v-model="replyContent"
            class="min-h-28 w-full resize-y rounded-md border border-gray-200 p-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100"
            placeholder="写下你的回复..."
            @input="clearReplyValidation()"
          />
          <p v-if="errorMessage" class="mt-2 text-sm text-red-600">{{ errorMessage }}</p>
          <p v-if="successMessage" class="mt-2 text-sm text-green-600">{{ successMessage }}</p>
          <div class="mt-3 flex justify-end">
            <button
              class="inline-flex h-9 items-center gap-1.5 rounded-md bg-blue-600 px-3 text-sm font-semibold text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
              :disabled="submitting"
              @click="submitReply()"
            >
              <Send class="h-4 w-4" />
              {{ submitting ? '发布中...' : '发布回复' }}
            </button>
          </div>
        </template>
        <template v-else>
          <div class="text-center">
            <h2 class="text-base font-semibold text-gray-950">加入讨论</h2>
            <p class="mt-1 text-sm text-gray-500">登录后可以回复这个主题。</p>
            <a href="/login" class="mt-4 inline-flex h-9 items-center rounded-md bg-blue-600 px-3 text-sm font-semibold text-white hover:bg-blue-700">登录</a>
          </div>
        </template>
      </section>
    </article>

    <Teleport v-if="page.props.permissions.canReply && showFloatingReply" to="body">
      <div class="pointer-events-none fixed inset-x-0 bottom-4 z-[80] px-3 sm:px-6">
        <Transition name="floating-reply" mode="out-in">
          <button
            v-if="!floatingReplyExpanded"
            type="button"
            class="pointer-events-auto mx-auto flex h-10 items-center gap-2 rounded-full border border-gray-200/80 bg-white/95 px-4 text-sm font-semibold text-gray-700 shadow-[0_14px_34px_-22px_rgba(15,23,42,0.55),0_4px_14px_-10px_rgba(15,23,42,0.35)] backdrop-blur transition hover:border-blue-200 hover:bg-blue-50 hover:text-blue-700"
            @click="openFloatingReply"
          >
            <MessageSquare class="h-4 w-4" />
            参与讨论
          </button>
          <div
            v-else
            class="pointer-events-auto mx-auto w-full max-w-2xl rounded-lg border border-gray-200/80 bg-white/95 p-3 shadow-[0_18px_48px_-24px_rgba(15,23,42,0.5),0_4px_16px_-12px_rgba(15,23,42,0.35)] backdrop-blur"
          >
            <div class="mb-2 flex items-center justify-between">
              <div class="text-sm font-semibold text-gray-950">参与讨论</div>
              <button type="button" class="rounded-md p-1 text-gray-400 transition hover:bg-gray-100 hover:text-gray-700" @click="closeFloatingReply">
                <X class="h-4 w-4" />
              </button>
            </div>
            <textarea
              v-model="replyContent"
              rows="3"
              class="min-h-24 w-full resize-y rounded-md border border-gray-200 bg-white p-3 text-sm leading-6 outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100"
              placeholder="写下你的回复..."
              @focus="openReplyId = null"
              @input="clearReplyValidation()"
            />
            <p v-if="errorMessage" class="mt-2 text-sm text-red-600">{{ errorMessage }}</p>
            <p v-if="successMessage" class="mt-2 text-sm text-green-600">{{ successMessage }}</p>
            <div class="mt-3 flex justify-end gap-2">
              <button type="button" class="h-9 rounded-md px-3 text-sm font-semibold text-gray-500 transition hover:bg-gray-100 hover:text-gray-800" @click="focusReplyEditor">
                完整编辑
              </button>
              <button
                type="button"
                class="inline-flex h-9 items-center gap-1.5 rounded-md bg-blue-600 px-3 text-sm font-semibold text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="submitting"
                @click="submitReply()"
              >
                <Loader2 v-if="submitting && currentReplyId === 0" class="h-4 w-4 animate-spin" />
                <Send v-else class="h-4 w-4" />
                {{ submitting && currentReplyId === 0 ? '发布中...' : '发布回复' }}
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </Teleport>

    <Teleport defer to="#goose-shell-rail">
      <div class="sticky top-19 space-y-3">
        <div class="overflow-hidden rounded-lg border border-gray-200/70 bg-white shadow-[0_2px_8px_rgba(0,0,0,0.02)]">
          <div class="border-b border-gray-100 px-4 py-4">
            <h2 class="text-sm font-semibold text-gray-500">主题概览</h2>
          </div>

          <dl class="space-y-4 px-4 py-5 text-sm">
            <div class="flex items-center justify-between gap-4">
              <dt class="font-semibold text-gray-500">创建于</dt>
              <dd class="text-right font-semibold tabular-nums text-gray-950">{{ formatDateTime(page.props.article.createdAt) }}</dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="font-semibold text-gray-500">最后回复</dt>
              <dd class="text-right font-semibold tabular-nums text-gray-950">{{ formatDateTime(page.props.article.updatedAt) }}</dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="font-semibold text-gray-500">回复数</dt>
              <dd class="text-right font-semibold tabular-nums text-gray-950">{{ formatNumber(page.props.article.replyCount) }}</dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="font-semibold text-gray-500">阅读数</dt>
              <dd class="text-right font-semibold tabular-nums text-gray-950">{{ formatNumber(page.props.article.viewCount) }}</dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="font-semibold text-gray-500">参与者</dt>
              <dd class="text-right font-semibold tabular-nums text-gray-950">{{ page.props.article.participants.length }}</dd>
            </div>
          </dl>

          <div v-if="page.props.article.participants.length" class="border-t border-gray-100 px-4 py-4">
            <h3 class="mb-3 text-sm font-semibold text-gray-500">活跃参与者</h3>
            <div class="flex flex-wrap gap-1.5">
              <a
                v-for="participant in page.props.article.participants"
                :key="participant.id"
                :href="`/u/${participant.id}`"
                class="rounded-full"
                @click="showUserCard(participant, $event)"
              >
                <img :src="participant.avatarUrl" :alt="participant.username" class="h-8 w-8 rounded-full object-cover ring-1 ring-gray-100 transition hover:ring-blue-300" />
              </a>
            </div>
          </div>
        </div>

        <div v-if="page.props.hotTopics.length" class="rounded-lg border border-gray-200/70 bg-white p-3 shadow-[0_2px_8px_rgba(0,0,0,0.02)]">
          <div class="mb-2 flex items-center justify-between">
            <h2 class="text-sm font-semibold text-gray-950">热门内容</h2>
            <a href="/?sort=hot" class="text-xs font-medium text-blue-600 hover:text-blue-700">更多</a>
          </div>
          <div class="space-y-0.5">
            <a
              v-for="topic in page.props.hotTopics"
              :key="topic.id"
              :href="topic.url"
              class="block rounded-md px-2 py-2 hover:bg-gray-50"
            >
              <div class="line-clamp-2 text-sm font-semibold leading-snug text-gray-900">{{ topic.title }}</div>
              <div class="mt-1 flex items-center gap-2 text-xs font-medium text-gray-600">
                <span class="truncate">{{ topic.author.username }}</span>
                <span class="shrink-0 tabular-nums">{{ formatNumber(topic.replyCount) }} 回复</span>
              </div>
            </a>
          </div>
        </div>

      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="pendingDeleteReply"
        class="fixed inset-0 z-[110] flex items-center justify-center bg-gray-950/45 px-4 py-6 backdrop-blur-sm"
        role="dialog"
        aria-modal="true"
        aria-labelledby="delete-reply-title"
        @click.self="closeDeleteDialog"
      >
        <div class="w-full max-w-sm rounded-lg border border-gray-200 bg-white p-4 shadow-[0_24px_80px_-32px_rgba(15,23,42,0.55),0_10px_32px_-18px_rgba(15,23,42,0.28)]">
          <div class="flex items-start gap-3">
            <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-red-50 text-red-600">
              <AlertTriangle class="h-5 w-5" />
            </div>
            <div class="min-w-0 flex-1">
              <h2 id="delete-reply-title" class="text-base font-bold text-gray-950">删除这条回复？</h2>
              <p class="mt-1 text-sm leading-6 text-gray-500">删除后这条回复会从主题中移除，相关回复数也会同步更新。</p>
            </div>
            <button
              type="button"
              class="rounded-md p-1 text-gray-400 transition hover:bg-gray-100 hover:text-gray-700 disabled:cursor-not-allowed disabled:opacity-50"
              :disabled="Boolean(deletingReplyId)"
              @click="closeDeleteDialog"
            >
              <X class="h-4 w-4" />
            </button>
          </div>

          <div class="mt-4 rounded-md border border-gray-100 bg-gray-50 px-3 py-2">
            <div class="text-xs font-semibold text-gray-500">@{{ pendingDeleteReply.author.username }}</div>
            <p class="mt-1 line-clamp-3 whitespace-pre-wrap text-sm leading-6 text-gray-700">{{ pendingDeleteReply.content }}</p>
          </div>

          <p v-if="deleteErrorMessage" class="mt-3 text-sm text-red-600">{{ deleteErrorMessage }}</p>

          <div class="mt-4 flex justify-end gap-2">
            <button
              type="button"
              class="h-9 rounded-md px-3 text-sm font-semibold text-gray-600 transition hover:bg-gray-100 disabled:cursor-not-allowed disabled:opacity-50"
              :disabled="Boolean(deletingReplyId)"
              @click="closeDeleteDialog"
            >
              取消
            </button>
            <button
              type="button"
              class="inline-flex h-9 items-center gap-1.5 rounded-md bg-red-600 px-3 text-sm font-semibold text-white transition hover:bg-red-700 disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="Boolean(deletingReplyId)"
              @click="removeReply(pendingDeleteReply.id)"
            >
              <Loader2 v-if="deletingReplyId === pendingDeleteReply.id" class="h-4 w-4 animate-spin" />
              <Trash2 v-else class="h-4 w-4" />
              {{ deletingReplyId === pendingDeleteReply.id ? '删除中...' : '确认删除' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.floating-reply-enter-active,
.floating-reply-leave-active {
  transition:
    opacity 180ms ease,
    transform 180ms ease;
}

.floating-reply-enter-from,
.floating-reply-leave-to {
  opacity: 0;
  transform: translateY(10px) scale(0.98);
}

@media (prefers-reduced-motion: reduce) {
  .floating-reply-enter-active,
  .floating-reply-leave-active {
    transition: none;
  }
}
</style>
