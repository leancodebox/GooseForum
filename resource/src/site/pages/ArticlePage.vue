<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, Teleport, watch } from 'vue'
import { AlertTriangle, Ban, Bell, Bookmark, Check, ChevronsUp, Clock, CornerDownLeft, Eye, Flag, Heart, Loader2, MessageSquare, PencilLine, RotateCcw, Trash2, X } from '@lucide/vue'
import { bookmarkArticle, deleteReply, getArticleRepliesWindow, likeArticle, postReply, submitReport, updateModerationArticleStatus, updateModerationReplyStatus, updateReply, watchArticle } from '@/runtime/api'
import { formatDateTime, formatNumber } from '@/runtime/format'
import { useFlashMessages } from '@/runtime/flash-message'
import { fetchPage } from '@/runtime/router'
import { useShellState } from '@/runtime/shell-state'
import { showUserCard } from '@/runtime/user-card-events'
import ArticleReplyComposer from '@/site/components/ArticleReplyComposer.vue'
import ReplyPositionRail from '@/site/components/ReplyPositionRail.vue'
import TopicList from '@/site/components/TopicList.vue'
import UserAvatar from '@/site/components/UserAvatar.vue'
import type { ArticleDetailProps, LayoutPayload, ReplyPayload } from '@/types/payload'
import { useI18n } from 'vue-i18n'

const page = defineProps<{
  layout: LayoutPayload
  props: ArticleDetailProps
}>()

const { t } = useI18n()
const { push: pushFlash } = useFlashMessages()
const replyContent = ref('')
const replyTargetId = ref(0)
const likeCount = ref(page.props.article.likeCount)
const isLiked = ref(page.props.article.isLiked)
const isBookmarked = ref(page.props.article.isBookmarked)
const isWatched = ref(page.props.article.isWatched)
const actionMessage = ref('')
const actingLike = ref(false)
const actingBookmark = ref(false)
const actingWatch = ref(false)
const actingModeration = ref(false)
const submitting = ref(false)
const deletingReplyId = ref(0)
const editingReplyId = ref(0)
const savingEditReplyId = ref(0)
const pendingDeleteReply = ref<ReplyPayload | null>(null)
const pendingModerationAction = ref<'ban' | 'unban' | null>(null)
const pendingReport = ref<{ targetType: 'article' | 'reply'; targetId: number; title: string; excerpt: string } | null>(null)
const reportReason = ref('spam')
const reportNote = ref('')
const reportSubmitting = ref(false)
const reportError = ref('')
const moderatingReplyIds = ref<number[]>([])
const replies = ref<ReplyPayload[]>([...page.props.replies])
const articleProcessStatus = ref(page.props.article.processStatus)
const replyTarget = computed(() => replies.value.find((reply) => reply.id === replyTargetId.value))
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
const editReplyContents = reactive<Record<number, string>>({})
const editReplyErrors = reactive<Record<number, string>>({})
const articleHeaderEl = ref<HTMLElement | null>(null)
const titleEl = ref<HTMLElement | null>(null)
const replyLoadMoreEl = ref<HTMLElement | null>(null)
const replyListEndEl = ref<HTMLElement | null>(null)
const articleRailTopOffset = ref(0)
const showHeaderTitle = ref(false)
const isMobileHeaderViewport = ref(false)
const mobileHeaderTitleVisible = ref(false)
const effectiveShowHeaderTitle = computed(() => showHeaderTitle.value && (!isMobileHeaderViewport.value || mobileHeaderTitleVisible.value))
const composerOpen = ref(false)
const mobileReplyRailOpen = ref(false)
const activeReplyNo = ref(firstReplyNo(page.props.replies) || 1)
const replyRailProgressCurrent = ref(0)
const replyRailProgressStart = ref(0)
const replyRailProgressEnd = ref(0)
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
const replyRailEndLabel = computed(() => formatRailDate(replyTailLoaded.value ? replies.value[replies.value.length - 1]?.createdAt || page.props.article.updatedAt : page.props.article.updatedAt))
const replyRailBusy = computed(() => loadingReplyWindow.value && (loadingReplyDirection.value === 'anchor' || loadingReplyDirection.value === 'tail'))
const actionMessageSuccess = computed(() =>
  [
    t('article.bookmarkAdded'),
    t('article.bookmarkRemoved'),
    t('article.watchAdded'),
    t('article.watchRemoved'),
    t('article.moderationBanSuccess'),
    t('article.moderationUnbanSuccess'),
  ].includes(actionMessage.value),
)
const reportReasons = ['spam', 'abuse', 'illegal', 'irrelevant', 'other']
const floatingArticleActions = computed(() => {
  const actions = [
    {
      key: 'like',
      icon: Heart,
      active: isLiked.value,
      acting: actingLike.value,
      fill: true,
      title: t('article.like'),
      activeClass: 'bg-error/10 text-error hover:bg-error/10',
      onClick: toggleLike,
    },
    {
      key: 'bookmark',
      icon: Bookmark,
      active: isBookmarked.value,
      acting: actingBookmark.value,
      fill: true,
      title: isBookmarked.value ? t('article.bookmarked') : t('article.bookmark'),
      activeClass: 'bg-info/10 text-primary hover:bg-info/10',
      onClick: toggleBookmark,
    },
    {
      key: 'watch',
      icon: Bell,
      active: isWatched.value,
      acting: actingWatch.value,
      fill: true,
      title: isWatched.value ? t('article.watched') : t('article.watch'),
      activeClass: 'bg-success/10 text-success hover:bg-success/15',
      onClick: toggleWatch,
    },
  ]

  if (page.props.permissions.canModerateArticle) {
    const isBanned = articleProcessStatus.value === 1
    actions.push({
      key: isBanned ? 'unban' : 'ban',
      icon: isBanned ? RotateCcw : Ban,
      active: false,
      acting: actingModeration.value,
      fill: false,
      title: isBanned ? t('article.moderationUnban') : t('article.moderationBan'),
      activeClass: 'text-base-content/75 hover:bg-base-200 hover:text-base-content',
      onClick: async () => requestArticleModeration(isBanned ? 'unban' : 'ban'),
    })
  }

  return actions
})
const shellState = useShellState()
let titleObserver: IntersectionObserver | undefined
let articleHeaderResizeObserver: ResizeObserver | undefined
let replyLoadObserver: IntersectionObserver | undefined
let lastHeaderScrollY = 0
let headerScrollFrame = 0
const highlightedReplyId = ref<number | null>(null)
let highlightTimer: number | undefined
let replyBottomLoadFrame = 0
let activeReplyScrollFrame = 0
let pendingReplyJumpNo: number | null = null
let replyRailSyncPaused = false
let replyRailResumeFrame = 0
let replyRailResumeLastScrollY = 0
let replyRailResumeStableFrames = 0
let replyElements: HTMLElement[] = []

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
  void nextTick(observeReplyLoader)
  void nextTick(collectReplyElements)
  void nextTick(scheduleActiveReplyFromScroll)
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
    articleProcessStatus.value = page.props.article.processStatus
    pendingModerationAction.value = null
    actingModeration.value = false
    mobileHeaderTitleVisible.value = false
    if (typeof window !== 'undefined') {
      lastHeaderScrollY = window.scrollY
    }
    resetRepliesFromProps()
    mobileReplyRailOpen.value = false
    void nextTick(observeArticleHeader)
    void nextTick(observeTitle)
    void nextTick(observeReplyLoader)
    void nextTick(collectReplyElements)
    void nextTick(scheduleActiveReplyFromScroll)
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
    void nextTick(() => {
      collectReplyElements()
      scheduleActiveReplyFromScroll()
    })
  },
)

onBeforeUnmount(() => {
  titleObserver?.disconnect()
  articleHeaderResizeObserver?.disconnect()
  replyLoadObserver?.disconnect()
  window.removeEventListener('scroll', updateMobileHeaderTitle)
  window.removeEventListener('scroll', scheduleActiveReplyFromScroll)
  window.removeEventListener('scroll', scheduleReplyBottomLoadCheck)
  window.removeEventListener('resize', updateArticleRailTopOffset)
  window.removeEventListener('resize', updateHeaderViewport)
  window.removeEventListener('resize', scheduleActiveReplyFromScroll)
  window.removeEventListener('resize', scheduleReplyBottomLoadCheck)
  window.cancelAnimationFrame(headerScrollFrame)
  window.cancelAnimationFrame(replyBottomLoadFrame)
  window.cancelAnimationFrame(activeReplyScrollFrame)
  window.cancelAnimationFrame(replyRailResumeFrame)
  window.clearTimeout(highlightTimer)
  shellState.headerTitle = ''
  shellState.headerTags = []
  shellState.showHeaderTitle = false
})

function setupHeaderTitleBehavior() {
  lastHeaderScrollY = window.scrollY
  updateHeaderViewport()
  window.addEventListener('scroll', updateMobileHeaderTitle, { passive: true })
  window.addEventListener('scroll', scheduleActiveReplyFromScroll, { passive: true })
  window.addEventListener('resize', updateArticleRailTopOffset)
  window.addEventListener('resize', updateHeaderViewport)
  window.addEventListener('resize', scheduleActiveReplyFromScroll)
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

function collectReplyElements() {
  replyElements = Array.from(document.querySelectorAll<HTMLElement>('[data-reply-no]'))
}

function pauseReplyRailSync() {
  replyRailSyncPaused = true
  window.cancelAnimationFrame(replyRailResumeFrame)
  replyRailResumeFrame = 0
  replyRailResumeLastScrollY = window.scrollY
  replyRailResumeStableFrames = 0
}

function resumeReplyRailSyncWhenSettled() {
  pauseReplyRailSync()
  const startedAt = performance.now()
  const settle = () => {
    const currentY = window.scrollY
    if (Math.abs(currentY - replyRailResumeLastScrollY) < 1) {
      replyRailResumeStableFrames += 1
    } else {
      replyRailResumeStableFrames = 0
      replyRailResumeLastScrollY = currentY
    }
    if (replyRailResumeStableFrames >= 4 || performance.now() - startedAt > 1600) {
      replyRailSyncPaused = false
      replyRailResumeFrame = 0
      syncReplyRailProgress()
      return
    }
    replyRailResumeFrame = window.requestAnimationFrame(settle)
  }
  replyRailResumeFrame = window.requestAnimationFrame(settle)
}

function scheduleActiveReplyFromScroll() {
  if (replyRailSyncPaused || activeReplyScrollFrame) return
  activeReplyScrollFrame = window.requestAnimationFrame(() => {
    activeReplyScrollFrame = 0
    syncReplyRailProgress()
  })
}

function syncReplyRailProgress() {
  const progress = measureReplyViewportProgress()
  if (progress.replyNo >= 0) {
    activeReplyNo.value = progress.replyNo
    replyRailProgressCurrent.value = progress.current
    replyRailProgressStart.value = progress.start
    replyRailProgressEnd.value = progress.end
  }
}

function measureReplyViewportProgress() {
  const markerY = Math.min(window.innerHeight * 0.38, 340)
  const viewportTop = 88
  const viewportBottom = window.innerHeight - 96
  const firstVisibleReplyNo = firstReplyNo(replies.value) || 1

  let coveringReplyNo: number | null = null
  let coveringProgress = 0
  let coveringDistance = Number.POSITIVE_INFINITY
  let nearestReplyNo: number | null = null
  let nearestProgress = 0
  let nearestDistance = Number.POSITIVE_INFINITY

  for (const element of replyElements) {
    const replyNo = Number(element.dataset.replyNo || 0)
    if (!replyNo) continue
    const rect = element.getBoundingClientRect()
    if (rect.bottom <= viewportTop || rect.top >= viewportBottom) continue

    const visibleTop = Math.max(viewportTop, rect.top)
    const visibleBottom = Math.min(viewportBottom, rect.bottom)
    if (visibleBottom <= visibleTop) continue

    if (rect.top <= markerY && rect.bottom >= markerY) {
      const distance = Math.abs(rect.top - markerY)
      if (distance < coveringDistance) {
        coveringReplyNo = replyNo
        coveringProgress = progressForReplyNoFraction(replyNo, (markerY - rect.top) / Math.max(1, rect.height))
        coveringDistance = distance
      }
      continue
    }

    const distance = Math.abs(rect.top - markerY)
    if (distance < nearestDistance) {
      nearestReplyNo = replyNo
      nearestProgress = progressForReplyNoFraction(replyNo, rect.top > markerY ? 0 : 1)
      nearestDistance = distance
    }
  }

  const fallbackReplyNo = firstVisibleReplyNo
  const replyNo = coveringReplyNo ?? nearestReplyNo ?? fallbackReplyNo
  const current = coveringReplyNo !== null
    ? coveringProgress
    : nearestReplyNo !== null
      ? nearestProgress
      : 0
  return {
    replyNo,
    current,
    start: Math.max(0, current - visibleSlotSize() / 2),
    end: Math.min(1, current + visibleSlotSize() / 2),
  }
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
  activeReplyNo.value = firstReplyNo(page.props.replies) || 1
  syncProgressForReplyNo(activeReplyNo.value)
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

function progressForReplyNo(replyNo: number) {
  return progressForReplyNoFraction(replyNo, 0.5)
}

function progressForReplyNoFraction(replyNo: number, fraction: number) {
  const maxReplyNo = Math.max(1, replyMaxRange.value || 1)
  if (maxReplyNo <= 1) return Math.min(1, Math.max(0, fraction))
  return Math.min(1, Math.max(0, (Math.max(1, replyNo) - 1 + Math.min(1, Math.max(0, fraction))) / maxReplyNo))
}

function visibleSlotSize() {
  return 1 / Math.max(1, replyMaxRange.value || 1)
}

function syncProgressForReplyNo(replyNo: number) {
  const progress = progressForReplyNo(replyNo)
  replyRailProgressCurrent.value = progress
  replyRailProgressStart.value = Math.max(0, progress - visibleSlotSize() / 2)
  replyRailProgressEnd.value = Math.min(1, progress + visibleSlotSize() / 2)
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
      syncProgressForReplyNo(activeReplyNo.value || 1)
    } else if (direction === 'tail') {
      activeReplyNo.value = lastReplyNo(payload.replies) || lastReplyNo(replies.value) || replyMaxRange.value
      syncProgressForReplyNo(activeReplyNo.value || 1)
    }
    await nextTick()
    collectReplyElements()
    if (replyAutoLoadAfter.value) {
      observeReplyLoader()
    }
    replyRailSyncPaused = false
    scheduleActiveReplyFromScroll()
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
  if (target >= replyMaxRange.value) {
    await jumpToLatestReply()
    return
  }

  if (loadingReplyWindow.value) {
    pendingReplyJumpNo = target
    activeReplyNo.value = target
    syncProgressForReplyNo(target)
    return
  }

  disableReplyAutoLoadAfter()
  activeReplyNo.value = target
  syncProgressForReplyNo(target)
  pauseReplyRailSync()
  const loaded = replies.value.find((reply) => reply.replyNo === target)
  if (loaded) {
    activeReplyNo.value = loaded.replyNo
    syncProgressForReplyNo(loaded.replyNo)
    await nextTick()
    document.getElementById(`reply-${loaded.id}`)?.scrollIntoView({ block: 'center', behavior: 'smooth' })
    resumeReplyRailSyncWhenSettled()
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
      syncProgressForReplyNo(closest.replyNo)
      collectReplyElements()
      document.getElementById(`reply-${closest.id}`)?.scrollIntoView({ block: 'center', behavior: 'smooth' })
      resumeReplyRailSyncWhenSettled()
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
    syncProgressForReplyNo(replyMaxRange.value)
    return
  }
  disableReplyAutoLoadAfter()
  activeReplyNo.value = replyMaxRange.value
  syncProgressForReplyNo(replyMaxRange.value)
  pauseReplyRailSync()
  if (replyTailLoaded.value) {
    const latest = replies.value[replies.value.length - 1]
    if (latest) {
      activeReplyNo.value = latest.replyNo
      syncProgressForReplyNo(latest.replyNo)
      await nextTick()
      scrollReplyListEndIntoView()
      resumeReplyRailSyncWhenSettled()
    }
    return
  }
  const loadedLatest = replies.value.find((reply) => reply.replyNo === replyMaxRange.value)
  if (loadedLatest) {
    activeReplyNo.value = loadedLatest.replyNo
    syncProgressForReplyNo(loadedLatest.replyNo)
    await nextTick()
    scrollReplyListEndIntoView()
    resumeReplyRailSyncWhenSettled()
    return
  }
  await loadReplyWindow('tail')
  await nextTick()
  const latest = replies.value[replies.value.length - 1]
  if (latest) {
    activeReplyNo.value = latest.replyNo
    syncProgressForReplyNo(latest.replyNo)
    scrollReplyListEndIntoView()
    resumeReplyRailSyncWhenSettled()
  }
}

function scrollReplyListEndIntoView() {
  if (replyListEndEl.value) {
    replyListEndEl.value.scrollIntoView({ block: 'end', behavior: 'smooth' })
    return
  }

  const latest = replies.value[replies.value.length - 1]
  if (latest) {
    document.getElementById(`reply-${latest.id}`)?.scrollIntoView({ block: 'end', behavior: 'smooth' })
  }
}

function flushPendingReplyJump() {
  if (!pendingReplyJumpNo || loadingReplyWindow.value) return
  const replyNo = pendingReplyJumpNo
  pendingReplyJumpNo = null
  void jumpToReplyNo(replyNo)
}

function jumpToArticleBody() {
  void jumpToReplyNo(1)
}

function focusReplyEditor() {
  mobileReplyRailOpen.value = false
  composerOpen.value = true
}

function openFloatingReply() {
  replyTargetId.value = 0
  focusReplyEditor()
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

function waitForAnimationFrame() {
  return new Promise<void>((resolve) => {
    window.requestAnimationFrame(() => resolve())
  })
}

async function findReplyElementAfterLayout(replyId: number) {
  for (let attempts = 0; attempts < 4; attempts += 1) {
    await nextTick()
    await waitForAnimationFrame()
    const element = document.getElementById(`reply-${replyId}`)
    if (element) return element
  }
  return null
}

async function revealCreatedReply(replyId: number) {
  if (!replyId) return

  pauseReplyRailSync()
  const payload = await getArticleRepliesWindow({
    articleId: page.props.article.id,
    anchorReplyId: replyId,
    limit: 20,
  })
  applyReplyWindowPayload(payload, 'replace', true)
  const createdReply = payload.replies.find((reply) => reply.id === replyId)
  if (createdReply?.replyNo) {
    activeReplyNo.value = createdReply.replyNo
    syncProgressForReplyNo(createdReply.replyNo)
  }
  highlightReply(replyId)
  const element = await findReplyElementAfterLayout(replyId)
  if (element && !isElementMostlyVisible(element)) {
    scrollReplyIntoComfortView(element)
    resumeReplyRailSyncWhenSettled()
    return
  }
  replyRailSyncPaused = false
  collectReplyElements()
  scheduleActiveReplyFromScroll()
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

function replyTo(reply: ReplyPayload) {
  replyTargetId.value = reply.id
  editingReplyId.value = 0
  errorMessage.value = ''
  successMessage.value = ''
  focusReplyEditor()
}

function cancelReplyTarget() {
  replyTargetId.value = 0
  errorMessage.value = ''
}

function clearReplyValidation() {
  errorMessage.value = ''
  successMessage.value = ''
}

function handleReplyImageInserted(count: number) {
  errorMessage.value = ''
  successMessage.value = count > 1 ? t('publish.imagesInserted', { count }) : t('publish.imageInserted')
}

function handleReplyImageError(message: string) {
  errorMessage.value = message
}

function startEditReply(reply: ReplyPayload) {
  replyTargetId.value = 0
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

async function submitReply() {
  const replyId = replyTarget.value?.id || 0
  const content = replyContent.value.trim()
  if (submitting.value) return

  if (!content) {
    errorMessage.value = t('article.replyRequired')
    successMessage.value = ''
    return
  }

  submitting.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const createdReply = await postReply(page.props.article.id, content, replyId)
    replyContent.value = ''
    replyTargetId.value = 0
    successMessage.value = t('article.replyPosted')
    const createdReplyId = typeof createdReply === 'object' && createdReply !== null ? createdReply.id : createdReply
    if (typeof createdReplyId === 'number') {
      await revealCreatedReply(createdReplyId)
    } else {
      await refreshCurrentPage()
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : t('api.replyFailed')
  } finally {
    submitting.value = false
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

function requestArticleModeration(action: 'ban' | 'unban') {
  actionMessage.value = ''
  pendingModerationAction.value = action
}

function closeArticleModerationDialog() {
  if (actingModeration.value) return
  pendingModerationAction.value = null
}

async function updateArticleModerationFromDetail() {
  if (actingModeration.value || !pendingModerationAction.value) return
  actingModeration.value = true
  actionMessage.value = ''
  const action = pendingModerationAction.value
  try {
    await updateModerationArticleStatus(page.props.article.id, action)
    articleProcessStatus.value = action === 'ban' ? 1 : 0
    pendingModerationAction.value = null
    actionMessage.value = action === 'ban' ? t('article.moderationBanSuccess') : t('article.moderationUnbanSuccess')
    pushFlash(actionMessage.value, 'success')
  } catch (error) {
    actionMessage.value = error instanceof Error ? error.message : t('api.moderationActionFailed')
    pushFlash(actionMessage.value, 'error')
  } finally {
    actingModeration.value = false
  }
}

function openLogin() {
  window.location.href = `/login?next=${encodeURIComponent(window.location.pathname + window.location.search + window.location.hash)}`
}

function requestReport(target: { targetType: 'article' | 'reply'; targetId: number; title: string; excerpt: string }) {
  if (!page.layout.viewer.isAuthenticated) {
    openLogin()
    return
  }
  pendingReport.value = target
  reportReason.value = 'spam'
  reportNote.value = ''
  reportError.value = ''
}

function requestArticleReport() {
  requestReport({
    targetType: 'article',
    targetId: page.props.article.id,
    title: page.props.article.title,
    excerpt: page.props.article.description,
  })
}

function requestReplyReport(reply: ReplyPayload) {
  requestReport({
    targetType: 'reply',
    targetId: reply.id,
    title: t('article.replyReportTitle', { no: reply.replyNo || reply.id }),
    excerpt: reply.content,
  })
}

function closeReportDialog() {
  if (reportSubmitting.value) return
  pendingReport.value = null
  reportError.value = ''
}

async function submitCurrentReport() {
  if (!pendingReport.value || reportSubmitting.value) return
  reportSubmitting.value = true
  reportError.value = ''
  try {
    await submitReport(pendingReport.value.targetType, pendingReport.value.targetId, reportReason.value, reportNote.value)
    pendingReport.value = null
    pushFlash(t('article.reportSubmitted'), 'success')
  } catch (error) {
    reportError.value = error instanceof Error ? error.message : t('api.reportFailed')
  } finally {
    reportSubmitting.value = false
  }
}

function replyModerationBusy(replyId: number) {
  return moderatingReplyIds.value.includes(replyId)
}

async function moderateReply(reply: ReplyPayload, action: 'ban' | 'unban') {
  if (replyModerationBusy(reply.id)) return
  moderatingReplyIds.value = [...moderatingReplyIds.value, reply.id]
  try {
    await updateModerationReplyStatus(reply.id, action)
    reply.processStatus = action === 'ban' ? 1 : 0
    reply.isHidden = action === 'ban'
    pushFlash(action === 'ban' ? t('article.replyModerationBanSuccess') : t('article.replyModerationUnbanSuccess'), 'success')
  } catch (error) {
    pushFlash(error instanceof Error ? error.message : t('api.moderationActionFailed'), 'error')
  } finally {
    moderatingReplyIds.value = moderatingReplyIds.value.filter(id => id !== reply.id)
  }
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
    <article class="min-w-0">
      <header ref="articleHeaderEl" class="relative z-10 border-b border-line/70 px-4 py-4 sm:mb-4 sm:px-0 sm:pb-4 sm:pt-0 xl:w-[calc(100%+292px)]">
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

      <section class="gf-card xl:w-[calc(100%+292px)]">
        <div class="min-w-0 xl:grid xl:grid-cols-[minmax(0,1fr)_256px]">
          <div class="min-w-0">
            <div class="grid grid-cols-[44px_minmax(0,1fr)] gap-3 p-4 sm:grid-cols-[52px_minmax(0,1fr)] sm:gap-4 sm:p-5">
              <a
                :href="`/u/${page.props.article.author.id}`"
                class="sticky top-19 self-start pt-1"
                @click="showUserCard(page.props.article.author, $event)"
              >
                <UserAvatar :src="page.props.article.author.avatarUrl" :alt="page.props.article.author.username" :badge="page.props.article.author.wornBadge" class="h-11 w-11 rounded-full ring-1 ring-line" img-class="rounded-full" />
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
                  <button
                    v-if="!page.props.permissions.isOwnArticle"
                    type="button"
                    class="gf-button gf-button-sm px-2.5 text-base-content/55 hover:bg-warning/10 hover:text-warning"
                    @click="requestArticleReport"
                  >
                    <Flag class="h-4 w-4" />
                    {{ t('article.report') }}
                  </button>
                  <button
                    v-if="page.props.permissions.canModerateArticle && articleProcessStatus === 0"
                    type="button"
                    class="gf-button gf-button-sm px-2.5 text-base-content/55 hover:bg-base-200 hover:text-base-content"
                    :disabled="actingModeration"
                    @click="requestArticleModeration('ban')"
                  >
                    <Ban class="h-4 w-4" />
                    {{ t('article.moderationBan') }}
                  </button>
                  <button
                    v-else-if="page.props.permissions.canModerateArticle && articleProcessStatus === 1"
                    type="button"
                    class="gf-button gf-button-sm px-2.5 text-base-content/55 hover:bg-base-200 hover:text-base-content"
                    :disabled="actingModeration"
                    @click="requestArticleModeration('unban')"
                  >
                    <RotateCcw class="h-4 w-4" />
                    {{ t('article.moderationUnban') }}
                  </button>
                  <span v-if="actionMessage" class="text-xs" :class="actionMessageSuccess ? 'text-base-content/75' : 'text-error'">{{ actionMessage }}</span>
                </div>
              </div>
            </div>

            <span v-if="replies.length" id="replies" class="block scroll-mt-20" aria-hidden="true" />

            <div v-if="replyHasBefore" class="relative border-t border-line px-4 py-3 text-center xl:border-t-transparent">
              <div class="pointer-events-none absolute left-5 right-5 top-0 hidden border-t border-line xl:block" aria-hidden="true" />
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
              class="group relative grid scroll-mt-20 grid-cols-[40px_minmax(0,1fr)] gap-2.5 border-t border-line px-3 py-4 transition hover:bg-base-200/70 sm:grid-cols-[52px_minmax(0,1fr)] sm:gap-4 sm:p-5 xl:border-t-transparent"
              :class="{ 'bg-info/10 ring-1 ring-inset ring-primary/20': highlightedReplyId === reply.id }"
            >
              <div class="pointer-events-none absolute left-5 right-5 top-0 hidden border-t border-line xl:block" aria-hidden="true" />
              <a
                :href="`/u/${reply.author.id}`"
                class="sticky top-19 self-start pt-1"
                @click="showUserCard(reply.author, $event)"
              >
                <UserAvatar :src="reply.author.avatarUrl" :alt="reply.author.username" :badge="reply.author.wornBadge" class="h-9 w-9 rounded-full ring-1 ring-line sm:h-10 sm:w-10" img-class="rounded-full" />
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
                      v-if="reply.isOwnReply && !reply.isHidden"
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
                      v-if="reply.isOwnReply && !reply.isHidden"
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
                      v-if="page.props.permissions.canReply && !reply.isHidden"
                      type="button"
                      class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-icon-muted transition hover:bg-info/10 hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2"
                      :title="t('article.reply')"
                      @click="replyTo(reply)"
                    >
                      <CornerDownLeft class="h-3.5 w-3.5" />
                      <span class="sr-only">{{ t('article.reply') }}</span>
                    </button>
                    <button
                      v-if="!reply.isOwnReply && !reply.isHidden"
                      type="button"
                      class="gf-icon-button h-8 w-8 shrink-0 hover:bg-warning/10 hover:text-warning focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-warning focus-visible:ring-offset-2"
                      :title="t('article.report')"
                      @click="requestReplyReport(reply)"
                    >
                      <Flag class="h-3.5 w-3.5" />
                      <span class="sr-only">{{ t('article.report') }}</span>
                    </button>
                    <button
                      v-if="reply.canModerate && reply.processStatus === 0"
                      type="button"
                      class="gf-icon-button h-8 w-8 shrink-0 hover:bg-error/10 hover:text-error focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-error focus-visible:ring-offset-2 disabled:opacity-50"
                      :disabled="replyModerationBusy(reply.id)"
                      :title="t('article.moderationBan')"
                      @click="moderateReply(reply, 'ban')"
                    >
                      <Ban class="h-3.5 w-3.5" />
                      <span class="sr-only">{{ t('article.moderationBan') }}</span>
                    </button>
                    <button
                      v-else-if="reply.canModerate && reply.processStatus === 1"
                      type="button"
                      class="gf-icon-button h-8 w-8 shrink-0 hover:bg-info/10 hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 disabled:opacity-50"
                      :disabled="replyModerationBusy(reply.id)"
                      :title="t('article.moderationUnban')"
                      @click="moderateReply(reply, 'unban')"
                    >
                      <RotateCcw class="h-3.5 w-3.5" />
                      <span class="sr-only">{{ t('article.moderationUnban') }}</span>
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
                  <div v-if="reply.isHidden && !reply.canModerate" class="rounded border border-line bg-base-200/60 px-3 py-2 text-sm text-base-content/45">
                    {{ t('article.hiddenReplyPlaceholder') }}
                  </div>
                  <div v-else class="gf-prose gf-prose-comment" v-html="reply.renderedContent" />
                  <div v-if="reply.isHidden && reply.canModerate" class="mt-2 inline-flex rounded bg-base-200 px-2 py-1 text-xs font-semibold text-base-content/45">
                    {{ t('article.hiddenReplyBadge') }}
                  </div>
                  <div v-if="reply.updatedAt && reply.updatedAt !== reply.createdAt" class="mt-2 text-xs font-medium text-base-content/55">
                    {{ t('article.editedAt', { time: formatDateTime(reply.updatedAt) }) }}
                  </div>
                </template>
              </div>
            </div>

            <div v-if="replyHasAfter || loadingReplyDirection === 'after' || replyWindowError || (!replyHasAfter && replies.length)" ref="replyLoadMoreEl" class="relative border-t border-line px-4 py-3 text-center xl:border-t-transparent">
              <div class="pointer-events-none absolute left-5 right-5 top-0 hidden border-t border-line xl:block" aria-hidden="true" />
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
            <span ref="replyListEndEl" class="block h-px scroll-mb-28" aria-hidden="true" />
          </div>

          <aside class="hidden min-w-0 xl:block">
            <div
              class="sticky top-19"
            >
              <div class="px-4 py-4">
                <h2 class="text-sm font-semibold text-base-content/55">{{ t('article.overview') }}</h2>
              </div>

              <dl class="space-y-4 border-t border-line px-4 py-5 text-sm">
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

              <ReplyPositionRail
                v-if="page.props.article.replyCount > 0 && replyMaxRange > 0"
                class="border-t border-line"
                :current="replyRailCurrentNo"
                :max="replyMaxRange"
                :start-label="replyRailStartLabel"
                :end-label="replyRailEndLabel"
                :current-label="replyRailCurrentLabel"
                :busy="replyRailBusy"
                :progress-current="replyRailProgressCurrent"
                :progress-end="replyRailProgressEnd"
                :progress-start="replyRailProgressStart"
                @earliest="jumpToArticleBodyFromRail"
                @latest="jumpToLatestReplyFromRail"
                @select="selectReplyFromRail"
              />
            </div>
          </aside>

          <section v-if="page.props.hotTopics.length" class="border-t border-line xl:col-span-2">
            <div class="overflow-hidden bg-base-100 [border-bottom-left-radius:calc(var(--gf-radius-box)-1px)] [border-bottom-right-radius:calc(var(--gf-radius-box)-1px)]">
              <TopicList :topics="page.props.hotTopics" home />
            </div>
          </section>
        </div>
      </section>

      <ArticleReplyComposer
        v-model="replyContent"
        v-model:mobile-rail-open="mobileReplyRailOpen"
        v-model:open="composerOpen"
        :actions="floatingArticleActions"
        :authenticated="page.layout.viewer.isAuthenticated"
        :can-reply="page.props.permissions.canReply"
        :current-label="replyRailCurrentLabel"
        :current-no="replyRailCurrentNo"
        :end-label="replyRailEndLabel"
        :error-message="errorMessage"
        :has-rail="hasReplyRail"
        :max-no="replyMaxRange"
        :progress-current="replyRailProgressCurrent"
        :progress-end="replyRailProgressEnd"
        :progress-start="replyRailProgressStart"
        :rail-busy="replyRailBusy"
        :start-label="replyRailStartLabel"
        :submitting="submitting"
        :success-message="successMessage"
        :target="replyTarget"
        @clear-target="cancelReplyTarget"
        @clear-validation="clearReplyValidation"
        @earliest="jumpToArticleBodyFromRail"
        @image-error="handleReplyImageError"
        @image-inserted="handleReplyImageInserted"
        @latest="jumpToLatestReplyFromRail"
        @open-reply="openFloatingReply"
        @select-rail="selectReplyFromRail"
        @submit="submitReply"
      />

    </article>

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

    <Teleport to="body">
      <Transition name="gf-modal">
        <div
          v-if="pendingReport"
          class="fixed inset-0 z-[110] flex items-center justify-center bg-neutral/45 px-4 py-6 backdrop-blur-sm"
          role="dialog"
          aria-modal="true"
          aria-labelledby="report-title"
          @click.self="closeReportDialog"
        >
          <div class="gf-menu-surface w-full max-w-sm p-4">
            <div class="flex items-start gap-3">
              <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-warning/10 text-warning">
                <Flag class="h-5 w-5" />
              </div>
              <div class="min-w-0 flex-1">
                <h2 id="report-title" class="text-base font-bold text-base-content">{{ t('article.reportTitle') }}</h2>
                <p class="mt-1 line-clamp-2 text-sm leading-6 text-base-content/55">{{ pendingReport.title }}</p>
              </div>
              <button
                type="button"
                class="rounded-md p-1 text-base-content/55 transition hover:bg-base-300 hover:text-base-content/75 disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="reportSubmitting"
                @click="closeReportDialog"
              >
                <X class="h-4 w-4" />
              </button>
            </div>

            <div class="mt-4 space-y-3">
              <label v-for="reason in reportReasons" :key="reason" class="flex cursor-pointer items-center gap-2 text-sm text-base-content/75">
                <input v-model="reportReason" class="radio radio-sm" type="radio" name="report-reason" :value="reason" />
                <span>{{ t(`article.reportReasons.${reason}`) }}</span>
              </label>
              <textarea
                v-model="reportNote"
                class="gf-textarea min-h-24"
                maxlength="300"
                :placeholder="t('article.reportNotePlaceholder')"
              />
            </div>

            <p v-if="reportError" class="mt-3 text-sm text-error">{{ reportError }}</p>

            <div class="mt-4 flex justify-end gap-2">
              <button
                type="button"
                class="gf-button gf-button-md gf-button-muted"
                :disabled="reportSubmitting"
                @click="closeReportDialog"
              >
                {{ t('common.cancel') }}
              </button>
              <button
                type="button"
                class="gf-button gf-button-md gf-button-primary"
                :disabled="reportSubmitting"
                @click="submitCurrentReport"
              >
                <Loader2 v-if="reportSubmitting" class="h-4 w-4 animate-spin" />
                <Flag v-else class="h-4 w-4" />
                {{ reportSubmitting ? t('common.loadingShort') : t('article.submitReport') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <Teleport to="body">
      <Transition name="gf-modal">
        <div
          v-if="pendingModerationAction"
          class="fixed inset-0 z-[110] flex items-center justify-center bg-neutral/45 px-4 py-6 backdrop-blur-sm"
          role="dialog"
          aria-modal="true"
          aria-labelledby="ban-article-title"
          @click.self="closeArticleModerationDialog"
        >
          <div class="gf-menu-surface w-full max-w-sm p-4">
            <div class="flex items-start gap-3">
              <AlertTriangle class="mt-0.5 h-5 w-5 shrink-0 text-error" />
              <div class="min-w-0 flex-1">
                <h2 id="ban-article-title" class="text-base font-bold text-base-content">
                  {{ pendingModerationAction === 'ban' ? t('article.moderationBanTitle') : t('article.moderationUnbanTitle') }}
                </h2>
                <p class="mt-1 text-sm leading-6 text-base-content/55">
                  {{ pendingModerationAction === 'ban' ? t('article.moderationBanDescription') : t('article.moderationUnbanDescription') }}
                </p>
              </div>
              <button
                type="button"
                class="rounded-md p-1 text-base-content/55 transition hover:bg-base-300 hover:text-base-content/75 disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="actingModeration"
                @click="closeArticleModerationDialog"
              >
                <X class="h-4 w-4" />
              </button>
            </div>

            <div class="mt-4 flex justify-end gap-2">
              <button
                type="button"
                class="gf-button gf-button-md gf-button-muted"
                :disabled="actingModeration"
                @click="closeArticleModerationDialog"
              >
                {{ t('common.cancel') }}
              </button>
              <button
                type="button"
                class="gf-button gf-button-md gf-button-danger"
                :disabled="actingModeration"
                @click="updateArticleModerationFromDetail"
              >
                <Loader2 v-if="actingModeration" class="h-4 w-4 animate-spin" />
                <component :is="pendingModerationAction === 'ban' ? Ban : RotateCcw" v-else class="h-4 w-4" />
                {{ actingModeration ? t('common.loadingShort') : (pendingModerationAction === 'ban' ? t('article.confirmModerationBan') : t('article.confirmModerationUnban')) }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
</template>
