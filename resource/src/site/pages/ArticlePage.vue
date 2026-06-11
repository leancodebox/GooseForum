<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, Teleport, watch } from 'vue'
import { AlertTriangle, Bell, Bookmark, Check, ChevronsUp, Clock, CornerDownLeft, Eye, Heart, Loader2, MessageSquare, PencilLine, Send, Trash2, X } from '@lucide/vue'
import { bookmarkArticle, deleteReply, getArticleRepliesWindow, likeArticle, postReply, updateReply, watchArticle } from '@/runtime/api'
import { formatDateTime, formatNumber } from '@/runtime/format'
import { fetchPage } from '@/runtime/router'
import { useShellState } from '@/runtime/shell-state'
import { showUserCard } from '@/runtime/user-card-events'
import ReplyPositionRail from '@/site/components/ReplyPositionRail.vue'
import UserAvatar from '@/site/components/UserAvatar.vue'
import type { ArticleDetailProps, LayoutPayload, ReplyPayload } from '@/types/payload'
import { useI18n } from 'vue-i18n'

const page = defineProps<{
  layout: LayoutPayload
  props: ArticleDetailProps
}>()

const { t } = useI18n()
const replyContent = ref('')
const replyContents = reactive<Record<number, string>>({})
const openReplyId = ref<number | null>(null)
const currentReplyId = ref(0)
const likeCount = ref(page.props.article.likeCount)
const isLiked = ref(page.props.article.isLiked)
const isBookmarked = ref(page.props.article.isBookmarked)
const isWatched = ref(page.props.article.isWatched)
const actionMessage = ref('')
const actingLike = ref(false)
const actingBookmark = ref(false)
const actingWatch = ref(false)
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
const replyBeforeReplyNo = ref(firstReplyNo(page.props.replies))
const replyAfterReplyNo = ref(lastReplyNo(page.props.replies))
const replyMaxNo = ref(initialMaxReplyNo())
const replyTailLoaded = ref(!hasMoreInitialReplies())
const replyAutoLoadAfter = ref(true)
const loadingReplyWindow = ref(false)
const loadingReplyDirection = ref<'before' | 'after' | 'anchor' | 'tail' | null>(null)
const replyWindowError = ref('')
const deleteErrorMessage = ref('')
const errorMessage = ref('')
const successMessage = ref('')
const inlineReplyErrors = reactive<Record<number, string>>({})
const editReplyContents = reactive<Record<number, string>>({})
const editReplyErrors = reactive<Record<number, string>>({})
const articleHeaderEl = ref<HTMLElement | null>(null)
const titleEl = ref<HTMLElement | null>(null)
const replyEditorEl = ref<HTMLTextAreaElement | null>(null)
const replySectionEl = ref<HTMLElement | null>(null)
const replyLoadMoreEl = ref<HTMLElement | null>(null)
const articleRailTopOffset = ref(0)
const showHeaderTitle = ref(false)
const isMobileHeaderViewport = ref(false)
const mobileHeaderTitleVisible = ref(false)
const effectiveShowHeaderTitle = computed(() => showHeaderTitle.value && (!isMobileHeaderViewport.value || mobileHeaderTitleVisible.value))
const showFloatingReply = ref(false)
const floatingReplyExpanded = ref(false)
const mobileReplyRailOpen = ref(false)
const activeReplyNo = ref(firstReplyNo(page.props.replies))
const replyMaxRange = computed(() => Math.max(replyMaxNo.value, ...replies.value.map((reply) => reply.replyNo || 0)))
const hasReplyRail = computed(() => page.props.article.replyCount > 0 && replyMaxRange.value > 0)
const replyRailCurrentNo = computed(() => {
  const fallback = firstReplyNo(replies.value) || 1
  return clampReplyNo(activeReplyNo.value || fallback)
})
const replyRailCurrentLabel = computed(() => {
  const activeReply = replies.value.find((reply) => reply.replyNo === replyRailCurrentNo.value)
  return activeReply ? formatRailDate(activeReply.createdAt) : ''
})
const replyRailStartLabel = computed(() => formatRailDate(page.props.article.createdAt))
const replyRailEndLabel = computed(() => formatRailDate(page.props.article.updatedAt))
const replyRailBusy = computed(() => loadingReplyWindow.value && (loadingReplyDirection.value === 'anchor' || loadingReplyDirection.value === 'tail'))
const floatingArticleActions = computed(() => [
  {
    key: 'like',
    icon: Heart,
    active: isLiked.value,
    acting: actingLike.value,
    title: t('article.like'),
    activeClass: 'bg-error/10 text-error hover:bg-error/10',
    onClick: toggleLike,
  },
  {
    key: 'bookmark',
    icon: Bookmark,
    active: isBookmarked.value,
    acting: actingBookmark.value,
    title: isBookmarked.value ? t('article.bookmarked') : t('article.bookmark'),
    activeClass: 'bg-info/10 text-primary hover:bg-info/10',
    onClick: toggleBookmark,
  },
  {
    key: 'watch',
    icon: Bell,
    active: isWatched.value,
    acting: actingWatch.value,
    title: isWatched.value ? t('article.watched') : t('article.watch'),
    activeClass: 'bg-success/10 text-success hover:bg-success/15',
    onClick: toggleWatch,
  },
])
const shellState = useShellState()
let titleObserver: IntersectionObserver | undefined
let articleHeaderResizeObserver: ResizeObserver | undefined
let replyEditorObserver: IntersectionObserver | undefined
let replyLoadObserver: IntersectionObserver | undefined
let replyVisibilityObserver: IntersectionObserver | undefined
let lastHeaderScrollY = 0
let headerScrollFrame = 0
const highlightedReplyId = ref<number | null>(null)
let highlightTimer: number | undefined
let replyVisibilityFrame = 0
let replyVisibilityPaused = false
let replyVisibilityResumeTimer: number | undefined
let replyBottomLoadFrame = 0
let pendingReplyJumpNo: number | null = null
const visibleReplyRatios = new Map<number, number>()

function updateArticleRailTopOffset() {
  if (!articleHeaderEl.value) {
    articleRailTopOffset.value = 0
    return
  }

  const style = window.getComputedStyle(articleHeaderEl.value)
  articleRailTopOffset.value = Math.ceil(articleHeaderEl.value.offsetHeight + (Number.parseFloat(style.marginBottom) || 0))
}

function observeArticleHeader() {
  articleHeaderResizeObserver?.disconnect()
  updateArticleRailTopOffset()

  if (!articleHeaderEl.value || !('ResizeObserver' in window)) return

  articleHeaderResizeObserver = new ResizeObserver(updateArticleRailTopOffset)
  articleHeaderResizeObserver.observe(articleHeaderEl.value)
}

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
  void nextTick(observeArticleHeader)
  void nextTick(observeTitle)
  void nextTick(observeReplyEditor)
  void nextTick(observeReplyLoader)
  void nextTick(scheduleObserveReplyVisibility)
  setupReplyBottomLoadFallback()
  void syncReplyHash()
})

watch(
  () => page.props.article.id,
  () => {
    likeCount.value = page.props.article.likeCount
    isLiked.value = page.props.article.isLiked
    isBookmarked.value = page.props.article.isBookmarked
    isWatched.value = page.props.article.isWatched
    mobileHeaderTitleVisible.value = false
    if (typeof window !== 'undefined') {
      lastHeaderScrollY = window.scrollY
    }
    resetRepliesFromProps()
    mobileReplyRailOpen.value = false
    void nextTick(observeArticleHeader)
    void nextTick(observeTitle)
    void nextTick(observeReplyEditor)
    void nextTick(observeReplyLoader)
    void nextTick(scheduleObserveReplyVisibility)
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

watch(
  () => replies.value.map((reply) => `${reply.id}:${reply.replyNo}`).join(','),
  () => {
    void nextTick(scheduleObserveReplyVisibility)
  },
)

onBeforeUnmount(() => {
  titleObserver?.disconnect()
  articleHeaderResizeObserver?.disconnect()
  replyEditorObserver?.disconnect()
  replyLoadObserver?.disconnect()
  replyVisibilityObserver?.disconnect()
  window.removeEventListener('scroll', updateMobileHeaderTitle)
  window.removeEventListener('scroll', scheduleReplyBottomLoadCheck)
  window.removeEventListener('resize', updateArticleRailTopOffset)
  window.removeEventListener('resize', updateHeaderViewport)
  window.removeEventListener('resize', scheduleReplyBottomLoadCheck)
  window.cancelAnimationFrame(headerScrollFrame)
  window.cancelAnimationFrame(replyVisibilityFrame)
  window.cancelAnimationFrame(replyBottomLoadFrame)
  window.clearTimeout(replyVisibilityResumeTimer)
  window.clearTimeout(highlightTimer)
  shellState.headerTitle = ''
  shellState.headerTags = []
  shellState.showHeaderTitle = false
})

function setupHeaderTitleBehavior() {
  lastHeaderScrollY = window.scrollY
  updateHeaderViewport()
  window.addEventListener('scroll', updateMobileHeaderTitle, { passive: true })
  window.addEventListener('resize', updateArticleRailTopOffset)
  window.addEventListener('resize', updateHeaderViewport)
}

function setupReplyBottomLoadFallback() {
  window.addEventListener('scroll', scheduleReplyBottomLoadCheck, { passive: true })
  window.addEventListener('resize', scheduleReplyBottomLoadCheck)
}

function scheduleReplyBottomLoadCheck() {
  if (replyBottomLoadFrame) return
  replyBottomLoadFrame = window.requestAnimationFrame(() => {
    replyBottomLoadFrame = 0
    void maybeLoadRepliesAtPageBottom()
  })
}

function isNearDocumentBottom() {
  const documentElement = document.documentElement
  const fullHeight = Math.max(documentElement.scrollHeight, document.body?.scrollHeight || 0)
  return fullHeight - (window.scrollY + window.innerHeight) <= 480
}

async function maybeLoadRepliesAtPageBottom() {
  if (!replyHasAfter.value || loadingReplyWindow.value || replyWindowError.value) return
  if (!isNearDocumentBottom()) return

  replyAutoLoadAfter.value = true
  await loadReplyWindow('after')
  await nextTick()
  if (replyHasAfter.value && isNearDocumentBottom()) {
    scheduleReplyBottomLoadCheck()
  }
}

async function loadMoreRepliesManually() {
  replyAutoLoadAfter.value = true
  await loadReplyWindow('after')
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
  if (!replyLoadMoreEl.value || !replyHasAfter.value || !replyAutoLoadAfter.value || !('IntersectionObserver' in window)) return

  replyLoadObserver = new IntersectionObserver(
    (entries) => {
      if (entries[0]?.isIntersecting && replyHasAfter.value && replyAutoLoadAfter.value && !loadingReplyWindow.value && !replyWindowError.value) {
        void loadReplyWindow('after')
      }
    },
    { rootMargin: '360px 0px' },
  )
  replyLoadObserver.observe(replyLoadMoreEl.value)
}

function scheduleObserveReplyVisibility() {
  if (replyVisibilityPaused) return
  if (replyVisibilityFrame) return
  replyVisibilityFrame = window.requestAnimationFrame(() => {
    replyVisibilityFrame = 0
    observeReplyVisibility()
  })
}

function observeReplyVisibility() {
  if (replyVisibilityPaused) return
  replyVisibilityObserver?.disconnect()
  visibleReplyRatios.clear()
  if (!('IntersectionObserver' in window)) {
    activeReplyNo.value = firstReplyNo(replies.value)
    return
  }

  const elements = document.querySelectorAll<HTMLElement>('[data-reply-no]')
  if (!elements.length) {
    activeReplyNo.value = 0
    return
  }

  replyVisibilityObserver = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        const replyNo = Number((entry.target as HTMLElement).dataset.replyNo || 0)
        if (!replyNo) continue
        if (entry.isIntersecting) {
          visibleReplyRatios.set(replyNo, entry.intersectionRatio)
        } else {
          visibleReplyRatios.delete(replyNo)
        }
      }

      const bestReplyNo = findViewportReplyNo()
      if (bestReplyNo > 0) {
        activeReplyNo.value = bestReplyNo
      }
    },
    { threshold: [0.05, 0.25, 0.5, 0.75], rootMargin: '-28% 0px -48% 0px' },
  )

  elements.forEach((element) => replyVisibilityObserver?.observe(element))
}

function pauseReplyVisibility(duration = 600) {
  replyVisibilityPaused = true
  window.clearTimeout(replyVisibilityResumeTimer)
  replyVisibilityResumeTimer = window.setTimeout(() => {
    replyVisibilityPaused = false
    scheduleObserveReplyVisibility()
  }, duration)
}

function findViewportReplyNo() {
  const markerY = Math.min(window.innerHeight * 0.42, 360)
  let coveringReplyNo = 0
  let coveringDistance = Number.POSITIVE_INFINITY
  let nearestReplyNo = 0
  let nearestDistance = Number.POSITIVE_INFINITY

  for (const replyNo of visibleReplyRatios.keys()) {
    const element = document.querySelector<HTMLElement>(`[data-reply-no="${replyNo}"]`)
    if (!element) continue
    const rect = element.getBoundingClientRect()
    if (rect.bottom <= 96 || rect.top >= window.innerHeight) continue

    if (rect.top <= markerY && rect.bottom >= markerY) {
      const distance = Math.abs(rect.top - markerY)
      if (distance < coveringDistance) {
        coveringReplyNo = replyNo
        coveringDistance = distance
      }
      continue
    }

    const distance = Math.abs(rect.top - markerY)
    if (distance < nearestDistance) {
      nearestReplyNo = replyNo
      nearestDistance = distance
    }
  }

  return coveringReplyNo || nearestReplyNo
}

function resetRepliesFromProps() {
  replies.value = [...page.props.replies]
  replyWindowMode.value = false
  replyHasBefore.value = false
  replyHasAfter.value = hasMoreInitialReplies()
  replyBeforeCursor.value = firstReplyId(page.props.replies)
  replyAfterCursor.value = lastReplyId(page.props.replies)
  replyBeforeReplyNo.value = firstReplyNo(page.props.replies)
  replyAfterReplyNo.value = lastReplyNo(page.props.replies)
  replyMaxNo.value = initialMaxReplyNo()
  replyTailLoaded.value = !hasMoreInitialReplies()
  replyAutoLoadAfter.value = true
  activeReplyNo.value = firstReplyNo(page.props.replies)
  replyWindowError.value = ''
  editingReplyId.value = 0
}

function firstReplyId(items: ReplyPayload[]) {
  return items.length ? items[0].id : 0
}

function lastReplyId(items: ReplyPayload[]) {
  return items.length ? items[items.length - 1].id : 0
}

function firstReplyNo(items: ReplyPayload[]) {
  return items.length ? items[0].replyNo || 0 : 0
}

function lastReplyNo(items: ReplyPayload[]) {
  return items.length ? items[items.length - 1].replyNo || 0 : 0
}

function initialMaxReplyNo() {
  return Math.max(page.props.article.maxReplyNo || 0, page.props.article.replyCount || 0, lastReplyNo(page.props.replies))
}

function clampReplyNo(replyNo: number) {
  const maxReplyNo = Math.max(1, replyMaxRange.value || 1)
  return Math.min(maxReplyNo, Math.max(1, Math.round(replyNo)))
}

function findClosestLoadedReply(replyNo: number) {
  let closest: ReplyPayload | undefined
  let closestDistance = Number.POSITIVE_INFINITY
  for (const reply of replies.value) {
    if (!reply.replyNo) continue
    const distance = Math.abs(reply.replyNo - replyNo)
    if (distance < closestDistance) {
      closest = reply
      closestDistance = distance
    }
  }
  return closest
}

function formatRailDate(value: string) {
  const normalized = value.replace(' ', 'T')
  const date = new Date(normalized)
  if (Number.isNaN(date.getTime())) return value.slice(0, 10)
  const now = new Date()
  const options: Intl.DateTimeFormatOptions = date.getFullYear() === now.getFullYear()
    ? { month: 'short', day: 'numeric' }
    : { year: 'numeric', month: 'short', day: 'numeric' }
  return new Intl.DateTimeFormat(undefined, options).format(date)
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

function applyReplyWindowPayload(
  payload: Awaited<ReturnType<typeof getArticleRepliesWindow>>,
  mergeMode: 'replace' | 'prepend' | 'append',
  forceWindowMode: boolean,
) {
  replyWindowMode.value = forceWindowMode || replyWindowMode.value
  mergeReplies(payload.replies, mergeMode)
  replyHasBefore.value = replyWindowMode.value ? payload.hasBefore : false
  replyHasAfter.value = payload.hasAfter
  replyBeforeCursor.value = payload.beforeCursor ?? firstReplyId(replies.value)
  replyAfterCursor.value = payload.afterCursor ?? lastReplyId(replies.value)
  replyBeforeReplyNo.value = payload.beforeReplyNo ?? firstReplyNo(replies.value)
  replyAfterReplyNo.value = payload.afterReplyNo ?? lastReplyNo(replies.value)
  replyMaxNo.value = Math.max(replyMaxNo.value, payload.maxReplyNo || 0)
  if (mergeMode === 'replace') {
    replyTailLoaded.value = payloadEndsAtTail(payload)
  } else if (mergeMode === 'append' && payloadEndsAtTail(payload)) {
    replyTailLoaded.value = true
  }
}

function payloadEndsAtTail(payload: Awaited<ReturnType<typeof getArticleRepliesWindow>>) {
  const afterReplyNo = payload.afterReplyNo || lastReplyNo(payload.replies)
  const maxReplyNo = Math.max(replyMaxNo.value, payload.maxReplyNo || 0)
  return payload.replies.length > 0 && !payload.hasAfter && afterReplyNo >= maxReplyNo
}

function disableReplyAutoLoadAfter() {
  replyAutoLoadAfter.value = false
  replyLoadObserver?.disconnect()
}

async function loadReplyWindow(direction: 'before' | 'after' | 'anchor' | 'tail', anchorValue = 0) {
  if (loadingReplyWindow.value) return
  if (direction === 'after' && (!replyHasAfter.value || !replyAutoLoadAfter.value)) return
  if (direction === 'tail' && replyTailLoaded.value) return

  if (direction !== 'after') {
    disableReplyAutoLoadAfter()
  }

  const wasWindowMode = replyWindowMode.value
  loadingReplyWindow.value = true
  loadingReplyDirection.value = direction
  replyWindowError.value = ''
  try {
    const payload = await getArticleRepliesWindow({
      articleId: page.props.article.id,
      anchorReplyId: direction === 'anchor' ? anchorValue : undefined,
      beforeReplyNo: direction === 'before' ? replyBeforeReplyNo.value : undefined,
      afterReplyNo: direction === 'after' ? replyAfterReplyNo.value : undefined,
      before: direction === 'before' && !replyBeforeReplyNo.value ? replyBeforeCursor.value : undefined,
      after: direction === 'after' && !replyAfterReplyNo.value ? replyAfterCursor.value : undefined,
      tail: direction === 'tail',
      limit: 20,
    })

    applyReplyWindowPayload(
      payload,
      direction === 'before' ? 'prepend' : direction === 'after' ? 'append' : 'replace',
      direction === 'anchor' || direction === 'tail' || direction === 'before' || wasWindowMode,
    )
    if (direction === 'after' && !payload.hasAfter) {
      replyTailLoaded.value = true
      disableReplyAutoLoadAfter()
    }
    if (direction === 'tail') {
      replyTailLoaded.value = true
      replyHasAfter.value = false
      disableReplyAutoLoadAfter()
    }
    if (direction === 'before') {
      activeReplyNo.value = firstReplyNo(payload.replies) || firstReplyNo(replies.value)
      pauseReplyVisibility(250)
    } else if (direction === 'tail') {
      activeReplyNo.value = lastReplyNo(payload.replies) || lastReplyNo(replies.value) || replyMaxRange.value
      pauseReplyVisibility()
    }
    await nextTick()
    if (replyAutoLoadAfter.value) {
      observeReplyLoader()
      scheduleObserveReplyVisibility()
    }
  } catch (error) {
    replyWindowError.value = error instanceof Error ? error.message : t('api.repliesLoadFailed')
  } finally {
    loadingReplyWindow.value = false
    loadingReplyDirection.value = null
    flushPendingReplyJump()
  }
}

async function jumpToReplyNo(replyNo: number) {
  const target = clampReplyNo(replyNo)
  if (loadingReplyWindow.value) {
    pendingReplyJumpNo = target
    activeReplyNo.value = target
    return
  }

  disableReplyAutoLoadAfter()
  activeReplyNo.value = target
  pauseReplyVisibility()
  const loaded = replies.value.find((reply) => reply.replyNo === target)
  if (loaded) {
    activeReplyNo.value = loaded.replyNo
    await nextTick()
    document.getElementById(`reply-${loaded.id}`)?.scrollIntoView({ block: 'center' })
    return
  }

  loadingReplyWindow.value = true
  loadingReplyDirection.value = 'anchor'
  replyWindowError.value = ''
  try {
    const payload = await getArticleRepliesWindow({
      articleId: page.props.article.id,
      anchorReplyNo: target,
      limit: 20,
    })
    applyReplyWindowPayload(payload, 'replace', true)
    await nextTick()
    const closest = findClosestLoadedReply(target)
    if (closest) {
      activeReplyNo.value = closest.replyNo
      document.getElementById(`reply-${closest.id}`)?.scrollIntoView({ block: 'center' })
      pauseReplyVisibility()
    }
  } catch (error) {
    replyWindowError.value = error instanceof Error ? error.message : t('api.repliesLoadFailed')
  } finally {
    loadingReplyWindow.value = false
    loadingReplyDirection.value = null
    flushPendingReplyJump()
  }
}

async function jumpToLatestReply() {
  if (!replyMaxRange.value) return
  if (loadingReplyWindow.value) {
    pendingReplyJumpNo = replyMaxRange.value
    activeReplyNo.value = replyMaxRange.value
    return
  }
  disableReplyAutoLoadAfter()
  activeReplyNo.value = replyMaxRange.value
  pauseReplyVisibility()
  if (replyTailLoaded.value) {
    const latest = replies.value[replies.value.length - 1]
    if (latest) {
      activeReplyNo.value = latest.replyNo
      await nextTick()
      document.getElementById(`reply-${latest.id}`)?.scrollIntoView({ block: 'center' })
      pauseReplyVisibility()
    }
    return
  }
  const loadedLatest = replies.value.find((reply) => reply.replyNo === replyMaxRange.value)
  if (loadedLatest) {
    await jumpToReplyNo(loadedLatest.replyNo)
    return
  }
  await loadReplyWindow('tail')
  await nextTick()
  const latest = replies.value[replies.value.length - 1]
  if (latest) {
    activeReplyNo.value = latest.replyNo
    document.getElementById(`reply-${latest.id}`)?.scrollIntoView({ block: 'center' })
  }
}

function flushPendingReplyJump() {
  if (!pendingReplyJumpNo || loadingReplyWindow.value) return
  const replyNo = pendingReplyJumpNo
  pendingReplyJumpNo = null
  void jumpToReplyNo(replyNo)
}

function jumpToArticleBody() {
  titleEl.value?.scrollIntoView({ block: 'start', behavior: 'smooth' })
}

function focusReplyEditor() {
  replyEditorEl.value?.focus()
  replyEditorEl.value?.scrollIntoView({ block: 'center' })
}

function openFloatingReply() {
  mobileReplyRailOpen.value = false
  floatingReplyExpanded.value = true
  openReplyId.value = null
}

function closeFloatingReply() {
  floatingReplyExpanded.value = false
}

function toggleMobileReplyRail() {
  floatingReplyExpanded.value = false
  mobileReplyRailOpen.value = !mobileReplyRailOpen.value
}

function closeMobileReplyRail() {
  mobileReplyRailOpen.value = false
}

async function selectReplyFromRail(replyNo: number) {
  closeMobileReplyRail()
  await jumpToReplyNo(replyNo)
}

async function jumpToLatestReplyFromRail() {
  closeMobileReplyRail()
  await jumpToLatestReply()
}

function jumpToArticleBodyFromRail() {
  closeMobileReplyRail()
  jumpToArticleBody()
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
  replyAfterReplyNo.value = Math.max(replyAfterReplyNo.value, reply.replyNo || 0)
  replyMaxNo.value = Math.max(replyMaxNo.value, reply.replyNo || 0)
  await nextTick()
  scheduleObserveReplyVisibility()
  highlightReply(reply.id)
  const element = document.getElementById(`reply-${reply.id}`)
  if (element && !isElementMostlyVisible(element)) {
    scrollReplyIntoComfortView(element)
  }
}

function buildCreatedReply(replyId: number, replyNo: number, content: string, renderedContent: string, replyToId: number): ReplyPayload {
  const parentReply = replyToId > 0 ? replies.value.find((reply) => reply.id === replyToId) : undefined
  const viewer = page.layout.viewer
  return {
    id: replyId,
    articleId: page.props.article.id,
    replyNo,
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
  applyReplyWindowPayload(payload, 'append', replyWindowMode.value)
  await nextTick()
  scheduleObserveReplyVisibility()
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
    actionMessage.value = error instanceof Error ? error.message : t('api.likeFailed')
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
    actionMessage.value = nextBookmarked ? t('article.bookmarkAdded') : t('article.bookmarkRemoved')
  } catch (error) {
    isBookmarked.value = previousBookmarked
    actionMessage.value = error instanceof Error ? error.message : t('api.bookmarkFailed')
  } finally {
    actingBookmark.value = false
  }
}

async function toggleWatch() {
  if (actingWatch.value) return

  const nextWatched = !isWatched.value
  const previousWatched = isWatched.value
  actingWatch.value = true
  actionMessage.value = ''
  isWatched.value = nextWatched
  try {
    await watchArticle(page.props.article.id, nextWatched ? 1 : 2)
    actionMessage.value = nextWatched ? t('article.watchAdded') : t('article.watchRemoved')
  } catch (error) {
    isWatched.value = previousWatched
    actionMessage.value = error instanceof Error ? error.message : t('api.watchFailed')
  } finally {
    actingWatch.value = false
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
    editReplyErrors[reply.id] = t('article.replyRequired')
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
    successMessage.value = t('article.replyUpdated')
  } catch (error) {
    editReplyErrors[reply.id] = error instanceof Error ? error.message : t('api.replyUpdateFailed')
  } finally {
    savingEditReplyId.value = 0
  }
}

async function submitReply(replyId = 0) {
  const content = replyId > 0 ? (replyContents[replyId] || '').trim() : replyContent.value.trim()
  if (submitting.value) return

  if (!content) {
    if (replyId > 0) {
      inlineReplyErrors[replyId] = t('article.replyRequired')
    } else {
      errorMessage.value = t('article.replyRequired')
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
      successMessage.value = t('article.replyPosted')
    }
    const createdReplyId = typeof createdReply === 'object' && createdReply !== null ? createdReply.id : createdReply
    const createdReplyNo = typeof createdReply === 'object' && createdReply !== null ? createdReply.replyNo || 0 : 0
    const renderedContent = typeof createdReply === 'object' && createdReply !== null ? createdReply.renderedContent : escapePlainText(content)
    if (typeof createdReplyId === 'number') {
      if (page.layout.viewer.isAuthenticated) {
        await revealCreatedReply(buildCreatedReply(createdReplyId, createdReplyNo || replyMaxRange.value + 1, content, renderedContent, replyId))
      } else {
        await fetchAndRevealCreatedReply(createdReplyId)
      }
    } else {
      await refreshCurrentPage()
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : t('api.replyFailed')
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
    successMessage.value = t('article.replyDeleted')
    pendingDeleteReply.value = null
    await refreshCurrentPage()
  } catch (error) {
    deleteErrorMessage.value = error instanceof Error ? error.message : t('api.replyDeleteFailed')
  } finally {
    deletingReplyId.value = 0
  }
}
</script>

<template>
  <div class="pb-20 xl:pb-0">
    <article class="min-w-0">
      <header ref="articleHeaderEl" class="relative z-10 mb-4 border-b border-line/70 pb-4 xl:w-[calc(100%+292px)]">
        <h1 ref="titleEl" class="break-words text-2xl font-bold leading-tight text-base-content [overflow-wrap:anywhere] sm:text-3xl">{{ page.props.article.title }}</h1>
        <div class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2 text-[13px] text-base-content/55">
          <a
            :href="`/u/${page.props.article.author.id}`"
            class="inline-flex items-center gap-2 font-medium text-base-content/75 hover:text-primary"
            @click="showUserCard(page.props.article.author, $event)"
          >
            <UserAvatar :src="page.props.article.author.avatarUrl" :alt="page.props.article.author.username" class="h-5 w-5 rounded-full object-cover" />
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
            class="inline-flex items-center gap-1.5 rounded-sm text-base-content/75 hover:text-primary"
          >
            <span class="h-2 w-2 rounded-[3px]" :style="{ backgroundColor: category.color }" />
            {{ category.name }}
          </a>
        </div>
      </header>

      <section class="gf-card">
        <div class="grid grid-cols-[44px_minmax(0,1fr)] gap-3 p-4 sm:grid-cols-[52px_minmax(0,1fr)] sm:gap-4 sm:p-5">
          <a
            :href="`/u/${page.props.article.author.id}`"
            class="sticky top-19 self-start pt-1"
            @click="showUserCard(page.props.article.author, $event)"
          >
            <UserAvatar :src="page.props.article.author.avatarUrl" :alt="page.props.article.author.username" class="h-11 w-11 rounded-full object-cover ring-1 ring-line" />
          </a>
          <div class="min-w-0">
            <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
              <div>
                <a :href="`/u/${page.props.article.author.id}`" class="font-semibold text-base-content hover:text-primary">{{ page.props.article.author.username }}</a>
                <div class="text-xs font-medium text-base-content/75">{{ t('article.body') }}</div>
              </div>
              <div class="flex flex-wrap items-center justify-end gap-3 text-xs font-medium text-base-content/75">
                <div class="flex items-center gap-3">
                  <span class="inline-flex items-center gap-1"><MessageSquare class="h-3.5 w-3.5" />{{ formatNumber(page.props.article.replyCount) }}</span>
                  <span class="inline-flex items-center gap-1"><Eye class="h-3.5 w-3.5" />{{ formatNumber(page.props.article.viewCount) }}</span>
                  <span class="inline-flex items-center gap-1"><Heart class="h-3.5 w-3.5" />{{ formatNumber(likeCount) }}</span>
                </div>
                <a
                  v-if="page.props.permissions.isOwnArticle"
                  :href="`/publish?id=${page.props.article.id}`"
                  class="gf-button gf-button-secondary h-7 px-2 text-xs hover:border-primary/20 hover:bg-info/10 hover:text-primary"
                >
                  <PencilLine class="h-3.5 w-3.5" />
                  {{ t('common.edit') }}
                </a>
              </div>
            </div>
            <div class="gf-prose gf-prose-article" v-html="page.props.article.html" />
            <div class="mt-6 flex flex-wrap items-center gap-3 border-t border-line pt-4">
              <button
                type="button"
                class="gf-button gf-button-sm px-2.5"
                :class="isLiked ? 'bg-error/10 text-error hover:bg-error/10' : 'text-base-content/55 hover:bg-base-200 hover:text-base-content'"
                :disabled="actingLike"
                @click="toggleLike"
              >
                <Heart class="h-4 w-4" :fill="isLiked ? 'currentColor' : 'none'" />
                {{ likeCount ? formatNumber(likeCount) : t('article.like') }}
              </button>
              <button
                type="button"
                class="gf-button gf-button-sm px-2.5"
                :class="isBookmarked ? 'bg-info/10 text-primary hover:bg-info/10' : 'text-base-content/55 hover:bg-base-200 hover:text-base-content'"
                :disabled="actingBookmark"
                @click="toggleBookmark"
              >
                <Bookmark class="h-4 w-4" :fill="isBookmarked ? 'currentColor' : 'none'" />
                {{ isBookmarked ? t('article.bookmarked') : t('article.bookmark') }}
              </button>
              <button
                type="button"
                class="gf-button gf-button-sm px-2.5"
                :class="isWatched ? 'bg-success/10 text-success hover:bg-success/15' : 'text-base-content/55 hover:bg-base-200 hover:text-base-content'"
                :disabled="actingWatch"
                @click="toggleWatch"
              >
                <Bell class="h-4 w-4" :fill="isWatched ? 'currentColor' : 'none'" />
                {{ isWatched ? t('article.watched') : t('article.watch') }}
              </button>
              <span v-if="actionMessage" class="text-xs" :class="actionMessage === t('article.bookmarkAdded') || actionMessage === t('article.bookmarkRemoved') || actionMessage === t('article.watchAdded') || actionMessage === t('article.watchRemoved') ? 'text-base-content/75' : 'text-error'">{{ actionMessage }}</span>
            </div>
          </div>
        </div>

        <span v-if="replies.length" id="replies" class="block scroll-mt-20" aria-hidden="true" />

        <div v-if="replyHasBefore" class="border-t border-line px-4 py-3 text-center">
          <button
            v-if="replyHasBefore"
            type="button"
            class="inline-flex h-8 items-center gap-1.5 rounded-md px-2 text-xs font-semibold text-primary transition hover:bg-info/10 hover:text-primary/90 disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="loadingReplyWindow"
            @click="loadReplyWindow('before')"
          >
            <Loader2 v-if="loadingReplyDirection === 'before'" class="h-3.5 w-3.5 animate-spin" />
            <ChevronsUp v-else class="h-3.5 w-3.5" />
            {{ t('article.loadEarlierReplies') }}
          </button>
        </div>

        <div
          v-for="reply in replies"
          :id="`reply-${reply.id}`"
          :key="reply.id"
          :data-reply-no="reply.replyNo"
          class="group grid scroll-mt-20 grid-cols-[40px_minmax(0,1fr)] gap-2.5 border-t border-line px-3 py-4 transition hover:bg-base-200/70 sm:grid-cols-[52px_minmax(0,1fr)] sm:gap-4 sm:p-5"
          :class="{ 'bg-info/10 ring-1 ring-inset ring-primary/20': highlightedReplyId === reply.id }"
        >
          <a
            :href="`/u/${reply.author.id}`"
            class="sticky top-19 self-start pt-1"
            @click="showUserCard(reply.author, $event)"
          >
            <UserAvatar :src="reply.author.avatarUrl" :alt="reply.author.username" class="h-9 w-9 rounded-full object-cover ring-1 ring-line sm:h-10 sm:w-10" />
          </a>
          <div class="min-w-0">
            <div class="mb-1.5 flex min-w-0 items-start justify-between gap-2">
              <div class="min-w-0">
                <div class="flex min-w-0 items-center gap-2">
                  <a :href="`/u/${reply.author.id}`" class="min-w-0 truncate font-semibold text-base-content hover:text-primary">{{ reply.author.username }}</a>
                  <span v-if="reply.replyNo" class="hidden shrink-0 text-xs font-semibold tabular-nums text-base-content/55 sm:inline">#{{ formatNumber(reply.replyNo) }}</span>
                </div>
                <div class="mt-0.5 flex items-center gap-2 text-xs text-base-content/55 sm:hidden">
                  <span v-if="reply.replyNo" class="font-semibold tabular-nums text-base-content/55">#{{ formatNumber(reply.replyNo) }}</span>
                  <time class="truncate">{{ formatDateTime(reply.createdAt) }}</time>
                </div>
              </div>
              <div class="flex shrink-0 items-center gap-0.5 sm:gap-1.5">
                <button
                  v-if="reply.isOwnReply"
                  type="button"
                  class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-icon-muted transition hover:bg-info/10 hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  :disabled="savingEditReplyId === reply.id || deletingReplyId === reply.id"
                  :title="t('common.edit')"
                  @click="startEditReply(reply)"
                >
                  <PencilLine class="h-3.5 w-3.5" />
                  <span class="sr-only">{{ t('common.edit') }}</span>
                </button>
                <button
                  v-if="reply.isOwnReply"
                  type="button"
                  class="gf-icon-button h-8 w-8 shrink-0 hover:bg-error/10 hover:text-error focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-error focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  :disabled="deletingReplyId === reply.id"
                  :title="deletingReplyId === reply.id ? t('article.deleting') : t('article.delete')"
                  @click="requestDeleteReply(reply)"
                >
                  <Trash2 class="h-3.5 w-3.5" />
                  <span class="sr-only">{{ deletingReplyId === reply.id ? t('article.deleting') : t('article.delete') }}</span>
                </button>
                <button
                  v-if="page.props.permissions.canReply"
                  type="button"
                  class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-icon-muted transition hover:bg-info/10 hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2"
                  :title="t('article.reply')"
                  @click="toggleReplyForm(reply.id)"
                >
                  <CornerDownLeft class="h-3.5 w-3.5" />
                  <span class="sr-only">{{ t('article.reply') }}</span>
                </button>
                <time class="hidden w-36 shrink-0 text-right text-xs text-base-content/55 sm:-ml-1 sm:block">{{ formatDateTime(reply.createdAt) }}</time>
              </div>
            </div>
            <p v-if="reply.replyToUsername" class="mb-1.5 inline-flex max-w-full min-w-0 items-center gap-1 rounded bg-base-200 px-2 py-1 text-sm text-base-content/55">
              <span class="shrink-0">{{ t('article.reply') }}</span>
              <a :href="`/u/${reply.replyToUserId}`" class="min-w-0 truncate font-medium text-base-content/75 hover:text-primary">@{{ reply.replyToUsername }}</a>
            </p>
            <Transition name="gf-local-expand">
              <div v-if="editingReplyId === reply.id" class="mt-3 rounded-lg border border-primary/20 bg-info/10 p-3">
                <div class="mb-2 flex items-center justify-between">
                  <div class="text-xs font-semibold text-primary">{{ t('article.editOwnReply') }}</div>
                  <button type="button" class="rounded-md p-1 text-base-content/55 hover:bg-base-100 hover:text-base-content/75" @click="cancelEditReply(reply.id)">
                    <X class="h-4 w-4" />
                  </button>
                </div>
                <textarea
                  v-model="editReplyContents[reply.id]"
                  class="gf-textarea min-h-24 border-primary/20"
                  :placeholder="t('article.editReplyPlaceholder')"
                  @input="clearEditReplyValidation(reply.id)"
                />
                <p v-if="editReplyErrors[reply.id]" class="mt-2 text-sm text-error">{{ editReplyErrors[reply.id] }}</p>
                <div class="mt-2 flex justify-end gap-2">
                  <button type="button" class="gf-button gf-button-sm gf-button-muted text-xs hover:bg-base-100" @click="cancelEditReply(reply.id)">{{ t('common.cancel') }}</button>
                  <button
                    type="button"
                    class="gf-button gf-button-sm gf-button-primary text-xs"
                    :disabled="savingEditReplyId === reply.id"
                    @click="saveReplyEdit(reply)"
                  >
                    <Loader2 v-if="savingEditReplyId === reply.id" class="h-3.5 w-3.5 animate-spin" />
                    <Check v-else class="h-3.5 w-3.5" />
                    {{ savingEditReplyId === reply.id ? t('common.saving') : t('common.save') }}
                  </button>
                </div>
              </div>
            </Transition>
            <template v-if="editingReplyId !== reply.id">
              <div class="gf-prose gf-prose-comment" v-html="reply.renderedContent" />
              <div v-if="reply.updatedAt && reply.updatedAt !== reply.createdAt" class="mt-2 text-xs font-medium text-base-content/55">
                {{ t('article.editedAt', { time: formatDateTime(reply.updatedAt) }) }}
              </div>
            </template>

            <Transition name="gf-local-expand">
              <div v-if="openReplyId === reply.id" class="mt-4 border-l-2 border-primary/20 pl-3">
                <div class="mb-2 flex items-center justify-between">
                  <div class="text-xs font-medium text-base-content/55">{{ t('article.replyTo', { user: `@${reply.author.username}` }) }}</div>
                  <button type="button" class="rounded-md p-1 text-base-content/55 hover:bg-base-300 hover:text-base-content/75" @click="openReplyId = null">
                    <X class="h-4 w-4" />
                  </button>
                </div>
                <textarea
                  v-model="replyContents[reply.id]"
                  class="gf-textarea min-h-20"
                  :placeholder="t('article.replyPlaceholder')"
                  @input="clearReplyValidation(reply.id)"
                />
                <p v-if="inlineReplyErrors[reply.id]" class="mt-2 text-sm text-error">{{ inlineReplyErrors[reply.id] }}</p>
                <div class="mt-2 flex justify-end gap-2">
                  <button type="button" class="gf-button gf-button-sm gf-button-muted text-xs" @click="openReplyId = null">{{ t('common.cancel') }}</button>
                  <button
                    type="button"
                    class="gf-button gf-button-sm gf-button-primary text-xs"
                    :disabled="submitting"
                    @click="submitReply(reply.id)"
                  >
                    <Send class="h-3.5 w-3.5" />
                    {{ submitting && currentReplyId === reply.id ? t('article.publishing') : t('article.reply') }}
                  </button>
                </div>
              </div>
            </Transition>
          </div>
        </div>

        <div v-if="replyHasAfter || loadingReplyDirection === 'after' || replyWindowError || (!replyHasAfter && replies.length)" ref="replyLoadMoreEl" class="border-t border-line px-4 py-3 text-center">
          <button
            v-if="replyHasAfter && replyWindowError"
            type="button"
            class="gf-button gf-button-sm gf-button-secondary text-xs"
            :disabled="loadingReplyWindow"
            @click="loadReplyWindow('after')"
          >
            <Loader2 v-if="loadingReplyDirection === 'after'" class="h-3.5 w-3.5 animate-spin" />
            {{ t('article.retryLoadReplies') }}
          </button>
          <p v-else-if="replyWindowError" class="text-xs text-error">{{ replyWindowError }}</p>
          <p v-else-if="replyHasAfter && loadingReplyDirection === 'after'" class="inline-flex items-center justify-center gap-1.5 text-xs font-medium text-base-content/55">
            <Loader2 class="h-3.5 w-3.5 animate-spin" />
            {{ t('article.loadingMoreReplies') }}
          </p>
          <button
            v-else-if="replyHasAfter"
            type="button"
            class="gf-button gf-button-sm gf-button-secondary text-xs"
            :disabled="loadingReplyWindow"
            @click="loadMoreRepliesManually"
          >
            {{ t('article.loadMoreReplies') }}
          </button>
          <p v-else-if="!replyHasAfter && replies.length" class="text-xs font-medium text-base-content/55">{{ t('article.allRepliesShown') }}</p>
        </div>
      </section>

      <section ref="replySectionEl" class="gf-card mt-4 p-4 sm:p-5">
        <template v-if="page.props.permissions.canReply">
          <div class="mb-3 flex items-center justify-between">
            <label class="text-sm font-semibold text-base-content" for="reply-content">{{ t('article.joinDiscussion') }}</label>
            <span class="text-xs text-base-content/55">{{ t('article.markdownSoon') }}</span>
          </div>
          <textarea
            id="reply-content"
            ref="replyEditorEl"
            v-model="replyContent"
            class="gf-textarea min-h-28"
            :placeholder="t('article.replyPlaceholder')"
            @input="clearReplyValidation()"
          />
          <p v-if="errorMessage" class="mt-2 text-sm text-error">{{ errorMessage }}</p>
          <p v-if="successMessage" class="mt-2 text-sm text-success">{{ successMessage }}</p>
          <div class="mt-3 flex justify-end">
            <button
              class="gf-button gf-button-md gf-button-primary"
              :disabled="submitting"
              @click="submitReply()"
            >
              <Send class="h-4 w-4" />
              {{ submitting ? t('article.publishing') : t('article.publishReply') }}
            </button>
          </div>
        </template>
        <template v-else>
          <div class="text-center">
            <h2 class="text-base font-semibold text-base-content">{{ t('article.joinDiscussion') }}</h2>
            <p class="mt-1 text-sm text-base-content/55">{{ t('article.loginToReply') }}</p>
            <a href="/login" class="gf-button gf-button-md gf-button-primary mt-4">{{ t('auth.loginTitle') }}</a>
          </div>
        </template>
      </section>

    </article>

    <Teleport defer to="#goose-shell-wide-content">
      <section v-if="page.props.hotTopics.length" class="gf-card w-full">
        <div class="flex items-center justify-between gap-3 border-b border-line px-4 py-4 sm:px-5">
          <h2 class="text-base font-bold text-base-content">{{ t('article.hotContent') }}</h2>
          <a href="/?sort=hot" class="text-sm font-semibold text-primary hover:text-primary">{{ t('article.more') }}</a>
        </div>
        <div class="divide-y divide-line">
          <a
            v-for="topic in page.props.hotTopics"
            :key="topic.id"
            :href="topic.url"
            class="block px-4 py-4 transition hover:bg-base-200 sm:px-5"
          >
            <div class="line-clamp-2 text-base font-bold leading-snug text-base-content">{{ topic.title }}</div>
            <p v-if="topic.description" class="mt-1 line-clamp-2 text-sm leading-6 text-base-content/55">{{ topic.description }}</p>
            <div class="mt-3 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs font-semibold text-base-content/55">
              <span>{{ topic.author.username }}</span>
              <span class="tabular-nums">{{ t('article.replyCountValue', { count: formatNumber(topic.replyCount) }) }}</span>
              <span class="tabular-nums">{{ formatNumber(topic.viewCount) }} {{ t('article.viewCount') }}</span>
            </div>
          </a>
        </div>
      </section>
    </Teleport>

    <Teleport v-if="hasReplyRail || (page.props.permissions.canReply && showFloatingReply)" to="body">
      <div class="pointer-events-none fixed inset-x-0 bottom-4 z-[90] px-3 sm:px-6">
        <div class="relative mx-auto flex w-fit max-w-full justify-center">
          <Transition name="floating-reply">
            <div
              v-if="mobileReplyRailOpen"
              class="gf-floating-surface pointer-events-auto absolute bottom-full left-0 mb-2 w-[min(18rem,calc(100vw-1.5rem))] p-2 xl:hidden"
            >
              <div class="mb-1 flex items-center justify-between gap-3 px-1">
                <div class="text-xs font-semibold text-base-content/55">{{ t('article.replyPosition') }}</div>
                <button
                  type="button"
                  class="inline-flex h-7 w-7 items-center justify-center rounded-md text-icon-muted transition hover:bg-base-300 hover:text-base-content"
                  :aria-label="t('common.close')"
                  @click="closeMobileReplyRail"
                >
                  <X class="h-4 w-4" />
                </button>
              </div>
              <ReplyPositionRail
                :current="replyRailCurrentNo"
                :max="replyMaxRange"
                :start-label="replyRailStartLabel"
                :end-label="replyRailEndLabel"
                :current-label="replyRailCurrentLabel"
                :busy="replyRailBusy"
                @earliest="jumpToArticleBodyFromRail"
                @latest="jumpToLatestReplyFromRail"
                @select="selectReplyFromRail"
              />
            </div>
          </Transition>

          <Transition name="floating-reply" mode="out-in">
            <div
              v-if="!floatingReplyExpanded || !page.props.permissions.canReply || !showFloatingReply"
              class="gf-floating-surface pointer-events-auto flex w-fit max-w-full items-center gap-1 rounded-full p-1"
            >
              <button
                v-if="hasReplyRail"
                type="button"
                class="inline-flex h-9 items-center rounded-full px-2.5 text-sm font-black tabular-nums text-primary transition hover:bg-info/10 hover:text-primary xl:hidden"
                :aria-expanded="mobileReplyRailOpen"
                :aria-label="t('article.replyPosition')"
                @click="toggleMobileReplyRail"
              >
                {{ replyRailCurrentNo }} / {{ formatNumber(replyMaxRange) }}
              </button>
              <template v-if="page.props.permissions.canReply && showFloatingReply">
                <button
                  v-for="action in floatingArticleActions"
                  :key="action.key"
                  type="button"
                  class="inline-flex h-9 w-9 items-center justify-center rounded-full text-sm font-semibold transition disabled:cursor-not-allowed disabled:opacity-60"
                  :class="action.active ? action.activeClass : 'text-base-content/75 hover:bg-base-200 hover:text-base-content'"
                  :disabled="action.acting"
                  :title="action.title"
                  @click="action.onClick"
                >
                  <Loader2 v-if="action.acting" class="h-4 w-4 animate-spin" />
                  <component :is="action.icon" v-else class="h-4 w-4" :fill="action.active ? 'currentColor' : 'none'" />
                </button>
                <button
                  type="button"
                  class="inline-flex h-9 items-center gap-1.5 rounded-full px-3 text-sm font-semibold text-base-content/75 transition hover:bg-info/10 hover:text-primary"
                  :title="t('article.joinDiscussion')"
                  @click="openFloatingReply"
                >
                  <MessageSquare class="h-4 w-4" />
                  <span>{{ t('article.joinDiscussion') }}</span>
                </button>
              </template>
            </div>
            <div
              v-else-if="page.props.permissions.canReply && showFloatingReply"
              class="gf-floating-surface pointer-events-auto w-[min(42rem,calc(100vw-1.5rem))] p-3"
            >
              <div class="mb-2 flex items-center justify-between">
                <div class="text-sm font-semibold text-base-content">{{ t('article.joinDiscussion') }}</div>
                <button type="button" class="rounded-md p-1 text-base-content/55 transition hover:bg-base-300 hover:text-base-content/75" @click="closeFloatingReply">
                  <X class="h-4 w-4" />
                </button>
              </div>
              <textarea
                v-model="replyContent"
                rows="3"
                class="gf-textarea min-h-24 leading-6"
                :placeholder="t('article.replyPlaceholder')"
                @focus="openReplyId = null"
                @input="clearReplyValidation()"
              />
              <p v-if="errorMessage" class="mt-2 text-sm text-error">{{ errorMessage }}</p>
              <p v-if="successMessage" class="mt-2 text-sm text-success">{{ successMessage }}</p>
              <div class="mt-3 flex justify-end gap-2">
                <button type="button" class="gf-button gf-button-md gf-button-muted" @click="focusReplyEditor">
                  {{ t('article.fullEditor') }}
                </button>
                <button
                  type="button"
                  class="gf-button gf-button-md gf-button-primary"
                  :disabled="submitting"
                  @click="submitReply()"
                >
                  <Loader2 v-if="submitting && currentReplyId === 0" class="h-4 w-4 animate-spin" />
                  <Send v-else class="h-4 w-4" />
                  {{ submitting && currentReplyId === 0 ? t('article.publishing') : t('article.publishReply') }}
                </button>
              </div>
            </div>
          </Transition>
        </div>
      </div>
    </Teleport>

    <Teleport defer to="#goose-shell-rail">
      <div
        class="sticky top-19 space-y-3"
        :style="{ marginTop: `${articleRailTopOffset}px` }"
      >
        <div class="gf-card overflow-hidden">
          <div class="border-b border-line px-4 py-4">
            <h2 class="text-sm font-semibold text-base-content/55">{{ t('article.overview') }}</h2>
          </div>

          <dl class="space-y-4 px-4 py-5 text-sm">
            <div class="flex items-center justify-between gap-4">
              <dt class="font-semibold text-base-content/55">{{ t('article.replyCount') }}</dt>
              <dd class="text-right font-semibold tabular-nums text-base-content">{{ formatNumber(page.props.article.replyCount) }}</dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="font-semibold text-base-content/55">{{ t('article.viewCount') }}</dt>
              <dd class="text-right font-semibold tabular-nums text-base-content">{{ formatNumber(page.props.article.viewCount) }}</dd>
            </div>
            <div class="flex items-center justify-between gap-4">
              <dt class="font-semibold text-base-content/55">{{ t('article.participants') }}</dt>
              <dd class="text-right font-semibold tabular-nums text-base-content">{{ page.props.article.participants.length }}</dd>
            </div>
          </dl>

          <div v-if="page.props.article.participants.length" class="border-t border-line px-4 py-4">
            <h3 class="mb-3 text-sm font-semibold text-base-content/55">{{ t('article.activeParticipants') }}</h3>
            <div class="flex flex-wrap gap-1.5">
              <a
                v-for="participant in page.props.article.participants"
                :key="participant.id"
                :href="`/u/${participant.id}`"
                class="rounded-full"
                @click="showUserCard(participant, $event)"
              >
                <UserAvatar :src="participant.avatarUrl" :alt="participant.username" class="h-8 w-8 rounded-full object-cover ring-1 ring-line transition hover:ring-primary/40" />
              </a>
            </div>
          </div>
        </div>

        <ReplyPositionRail
          v-if="page.props.article.replyCount > 0 && replyMaxRange > 0"
          :current="replyRailCurrentNo"
          :max="replyMaxRange"
          :start-label="replyRailStartLabel"
          :end-label="replyRailEndLabel"
          :current-label="replyRailCurrentLabel"
          :busy="replyRailBusy"
          @earliest="jumpToArticleBodyFromRail"
          @latest="jumpToLatestReplyFromRail"
          @select="selectReplyFromRail"
        />

      </div>
    </Teleport>

    <Teleport to="body">
      <Transition name="gf-modal">
        <div
          v-if="pendingDeleteReply"
          class="fixed inset-0 z-[110] flex items-center justify-center bg-neutral/45 px-4 py-6 backdrop-blur-sm"
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-reply-title"
          @click.self="closeDeleteDialog"
        >
          <div class="gf-menu-surface w-full max-w-sm p-4">
            <div class="flex items-start gap-3">
              <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-error/10 text-error">
                <AlertTriangle class="h-5 w-5" />
              </div>
              <div class="min-w-0 flex-1">
                <h2 id="delete-reply-title" class="text-base font-bold text-base-content">{{ t('article.deleteReplyTitle') }}</h2>
                <p class="mt-1 text-sm leading-6 text-base-content/55">{{ t('article.deleteReplyDescription') }}</p>
              </div>
              <button
                type="button"
                class="rounded-md p-1 text-base-content/55 transition hover:bg-base-300 hover:text-base-content/75 disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="Boolean(deletingReplyId)"
                @click="closeDeleteDialog"
              >
                <X class="h-4 w-4" />
              </button>
            </div>

            <div class="mt-4 rounded-md border border-line bg-base-200 px-3 py-2">
              <div class="text-xs font-semibold text-base-content/55">@{{ pendingDeleteReply.author.username }}</div>
              <p class="mt-1 line-clamp-3 whitespace-pre-wrap text-sm leading-6 text-base-content/75">{{ pendingDeleteReply.content }}</p>
            </div>

            <p v-if="deleteErrorMessage" class="mt-3 text-sm text-error">{{ deleteErrorMessage }}</p>

            <div class="mt-4 flex justify-end gap-2">
              <button
                type="button"
                class="gf-button gf-button-md gf-button-muted"
                :disabled="Boolean(deletingReplyId)"
                @click="closeDeleteDialog"
              >
                {{ t('common.cancel') }}
              </button>
              <button
                type="button"
                class="gf-button gf-button-md gf-button-danger"
                :disabled="Boolean(deletingReplyId)"
                @click="removeReply(pendingDeleteReply.id)"
              >
                <Loader2 v-if="deletingReplyId === pendingDeleteReply.id" class="h-4 w-4 animate-spin" />
                <Trash2 v-else class="h-4 w-4" />
                {{ deletingReplyId === pendingDeleteReply.id ? t('article.deleting') : t('article.confirmDelete') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
