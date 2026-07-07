<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, Teleport, watch } from 'vue'
import { AlertTriangle, Ban, Bell, Bookmark, ChevronsUp, Clock, CornerDownLeft, Eye, Flag, Heart, Loader2, MessageSquare, PencilLine, RotateCcw, Trash2, X } from '@lucide/vue'
import { bookmarkTopic, deletePost, getPostWindow, likeTopic, createPost, submitReport, updateModerationTopicStatus, updateModerationPostStatus, updatePost, watchTopic } from '@/runtime/api'
import { formatDateTime, formatNumber } from '@/runtime/format'
import { useFlashMessages } from '@/runtime/flash-message'
import { fetchPage } from '@/runtime/router'
import { useShellState } from '@/runtime/shell-state'
import { showUserCard } from '@/runtime/user-card-events'
import PostComposer from '@/site/components/PostComposer.vue'
import MarkdownImageViewer from '@/site/components/MarkdownImageViewer.vue'
import PostPositionRail from '@/site/components/PostPositionRail.vue'
import TopicList from '@/site/components/TopicList.vue'
import UserAvatar from '@/site/components/UserAvatar.vue'
import type { TopicDetailProps, LayoutPayload, PostPayload } from '@/types/payload'
import { useI18n } from 'vue-i18n'

const page = defineProps<{
  layout: LayoutPayload
  props: TopicDetailProps
}>()

const { t } = useI18n()
const { push: pushFlash } = useFlashMessages()
const replyContent = ref('')
const replyTargetId = ref(0)
const likeCount = ref(page.props.topic.likeCount)
const isLiked = ref(page.props.topic.isLiked)
const isBookmarked = ref(page.props.topic.isBookmarked)
const isWatched = ref(page.props.topic.isWatched)
const actionMessage = ref('')
const actingLike = ref(false)
const actingBookmark = ref(false)
const actingWatch = ref(false)
const actingModeration = ref(false)
const submitting = ref(false)
const deletingReplyId = ref(0)
const editingReplyId = ref(0)
const savingEditReplyId = ref(0)
const replyDraftBeforeEdit = ref('')
const replyTargetBeforeEdit = ref(0)
const pendingDeleteReply = ref<PostPayload | null>(null)
const pendingModerationAction = ref<'ban' | 'unban' | null>(null)
const pendingReport = ref<{ targetType: 'topic' | 'post'; targetId: number; title: string; excerpt: string } | null>(null)
const reportReason = ref('spam')
const reportNote = ref('')
const reportSubmitting = ref(false)
const reportError = ref('')
const moderatingReplyIds = ref<number[]>([])
const replies = ref<PostPayload[]>([...page.props.posts])
const topicProcessStatus = ref(page.props.topic.processStatus)
const replyTarget = computed(() => replies.value.find((reply) => reply.id === replyTargetId.value))
const replyWindowMode = ref(false)
const replyHasBefore = ref(false)
const replyHasAfter = ref(hasMoreInitialReplies())
const replyBeforeCursor = ref(firstReplyId(page.props.posts))
const replyAfterCursor = ref(lastReplyId(page.props.posts))
const replyBeforePostNo = ref(firstPostNo(page.props.posts))
const replyAfterPostNo = ref(lastPostNo(page.props.posts))
const replyMaxNo = ref(initialMaxPostNo())
const replyTailLoaded = ref(!hasMoreInitialReplies())
const replyAutoLoadAfter = ref(true)
const loadingReplyWindow = ref(false)
const loadingReplyDirection = ref<'before' | 'after' | 'anchor' | 'tail' | null>(null)
const replyWindowError = ref('')
const deleteErrorMessage = ref('')
const errorMessage = ref('')
const successMessage = ref('')
const topicHeaderEl = ref<HTMLElement | null>(null)
const titleEl = ref<HTMLElement | null>(null)
const replyLoadMoreEl = ref<HTMLElement | null>(null)
const replyListEndEl = ref<HTMLElement | null>(null)
const markdownImageViewer = ref<InstanceType<typeof MarkdownImageViewer> | null>(null)
const topicRailTopOffset = ref(0)
const showHeaderTitle = ref(false)
const isMobileHeaderViewport = ref(false)
const mobileHeaderTitleVisible = ref(false)
const effectiveShowHeaderTitle = computed(() => showHeaderTitle.value && (!isMobileHeaderViewport.value || mobileHeaderTitleVisible.value))
const composerOpen = ref(false)
const composerMode = computed(() => editingReplyId.value ? 'edit' : 'create')
const mobileReplyRailOpen = ref(false)
const activePostNo = ref(firstPostNo(page.props.posts) || 1)
const replyRailProgressCurrent = ref(0)
const replyRailProgressStart = ref(0)
const replyRailProgressEnd = ref(0)
const replyMaxRange = computed(() => Math.max(replyMaxNo.value, ...replies.value.map((reply) => reply.postNo || 0)))
const hasReplyRail = computed(() => page.props.topic.replyCount > 0 && replyMaxRange.value > 0)
const replyRailCurrentNo = computed(() => {
  const fallback = firstPostNo(replies.value) || 1
  return clampPostNo(activePostNo.value || fallback)
})
const replyRailCurrentLabel = computed(() => {
  const activeReply = replies.value.find((reply) => reply.postNo === replyRailCurrentNo.value)
  return activeReply ? formatRailDate(activeReply.createdAt) : ''
})
const replyRailStartLabel = computed(() => formatRailDate(page.props.topic.createdAt))
const replyRailEndLabel = computed(() => formatRailDate(replyTailLoaded.value ? replies.value[replies.value.length - 1]?.createdAt || page.props.topic.updatedAt : page.props.topic.updatedAt))
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
const floatingTopicActions = computed(() => {
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

  if (page.props.permissions.canModerateTopic) {
    const isBanned = topicProcessStatus.value === 1
    actions.push({
      key: isBanned ? 'unban' : 'ban',
      icon: isBanned ? RotateCcw : Ban,
      active: false,
      acting: actingModeration.value,
      fill: false,
      title: isBanned ? t('article.moderationUnban') : t('article.moderationBan'),
      activeClass: 'text-base-content/75 hover:bg-base-200 hover:text-base-content',
      onClick: async () => requestTopicModeration(isBanned ? 'unban' : 'ban'),
    })
  }

  return actions
})
const shellState = useShellState()
let titleObserver: IntersectionObserver | undefined
let topicHeaderResizeObserver: ResizeObserver | undefined
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

function updateTopicRailTopOffset() {
  if (!topicHeaderEl.value) {
    topicRailTopOffset.value = 0
    return
  }

  const style = window.getComputedStyle(topicHeaderEl.value)
  topicRailTopOffset.value = Math.ceil(topicHeaderEl.value.offsetHeight + (Number.parseFloat(style.marginBottom) || 0))
}

function observeTopicHeader() {
  topicHeaderResizeObserver?.disconnect()
  updateTopicRailTopOffset()

  if (!topicHeaderEl.value || !('ResizeObserver' in window)) return

  topicHeaderResizeObserver = new ResizeObserver(updateTopicRailTopOffset)
  topicHeaderResizeObserver.observe(topicHeaderEl.value)
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
  void nextTick(observeTopicHeader)
  void nextTick(observeTitle)
  void nextTick(observeReplyLoader)
  void nextTick(collectReplyElements)
  void nextTick(scheduleActiveReplyFromScroll)
  setupReplyBottomLoadFallback()
  void syncReplyHash()
})

watch(
  () => page.props.topic.id,
  () => {
    likeCount.value = page.props.topic.likeCount
    isLiked.value = page.props.topic.isLiked
    isBookmarked.value = page.props.topic.isBookmarked
    isWatched.value = page.props.topic.isWatched
    topicProcessStatus.value = page.props.topic.processStatus
    pendingModerationAction.value = null
    actingModeration.value = false
    mobileHeaderTitleVisible.value = false
    if (typeof window !== 'undefined') {
      lastHeaderScrollY = window.scrollY
    }
    resetRepliesFromProps()
    mobileReplyRailOpen.value = false
    void nextTick(observeTopicHeader)
    void nextTick(observeTitle)
    void nextTick(observeReplyLoader)
    void nextTick(collectReplyElements)
    void nextTick(scheduleActiveReplyFromScroll)
    void nextTick(syncReplyHash)
  },
  { immediate: true },
)

watch(
  () => [page.props.topic.title, page.props.topic.categories, effectiveShowHeaderTitle.value] as const,
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
  () => replies.value.map((reply) => `${reply.id}:${reply.postNo}`).join(','),
  () => {
    void nextTick(() => {
      collectReplyElements()
      scheduleActiveReplyFromScroll()
    })
  },
)

onBeforeUnmount(() => {
  titleObserver?.disconnect()
  topicHeaderResizeObserver?.disconnect()
  replyLoadObserver?.disconnect()
  window.removeEventListener('scroll', updateMobileHeaderTitle)
  window.removeEventListener('scroll', scheduleActiveReplyFromScroll)
  window.removeEventListener('scroll', scheduleReplyBottomLoadCheck)
  window.removeEventListener('resize', updateTopicRailTopOffset)
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
  window.addEventListener('resize', updateTopicRailTopOffset)
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
  replyElements = Array.from(document.querySelectorAll<HTMLElement>('[data-post-no]'))
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
  if (progress.postNo >= 0) {
    activePostNo.value = progress.postNo
    replyRailProgressCurrent.value = progress.current
    replyRailProgressStart.value = progress.start
    replyRailProgressEnd.value = progress.end
  }
}

function measureReplyViewportProgress() {
  const markerY = Math.min(window.innerHeight * 0.38, 340)
  const viewportTop = 88
  const viewportBottom = window.innerHeight - 96
  const firstVisiblePostNo = firstPostNo(replies.value) || 1

  let coveringPostNo: number | null = null
  let coveringProgress = 0
  let coveringDistance = Number.POSITIVE_INFINITY
  let nearestPostNo: number | null = null
  let nearestProgress = 0
  let nearestDistance = Number.POSITIVE_INFINITY

  for (const element of replyElements) {
    const postNo = Number(element.dataset.postNo || 0)
    if (!postNo) continue
    const rect = element.getBoundingClientRect()
    if (rect.bottom <= viewportTop || rect.top >= viewportBottom) continue

    const visibleTop = Math.max(viewportTop, rect.top)
    const visibleBottom = Math.min(viewportBottom, rect.bottom)
    if (visibleBottom <= visibleTop) continue

    if (rect.top <= markerY && rect.bottom >= markerY) {
      const distance = Math.abs(rect.top - markerY)
      if (distance < coveringDistance) {
        coveringPostNo = postNo
        coveringProgress = progressForPostNoFraction(postNo, (markerY - rect.top) / Math.max(1, rect.height))
        coveringDistance = distance
      }
      continue
    }

    const distance = Math.abs(rect.top - markerY)
    if (distance < nearestDistance) {
      nearestPostNo = postNo
      nearestProgress = progressForPostNoFraction(postNo, rect.top > markerY ? 0 : 1)
      nearestDistance = distance
    }
  }

  const fallbackPostNo = firstVisiblePostNo
  const postNo = coveringPostNo ?? nearestPostNo ?? fallbackPostNo
  const current = coveringPostNo !== null
    ? coveringProgress
    : nearestPostNo !== null
      ? nearestProgress
      : 0
  return {
    postNo,
    current,
    start: Math.max(0, current - visibleSlotSize() / 2),
    end: Math.min(1, current + visibleSlotSize() / 2),
  }
}

function resetRepliesFromProps() {
  replies.value = [...page.props.posts]
  replyWindowMode.value = false
  replyHasBefore.value = false
  replyHasAfter.value = hasMoreInitialReplies()
  replyBeforeCursor.value = firstReplyId(page.props.posts)
  replyAfterCursor.value = lastReplyId(page.props.posts)
  replyBeforePostNo.value = firstPostNo(page.props.posts)
  replyAfterPostNo.value = lastPostNo(page.props.posts)
  replyMaxNo.value = initialMaxPostNo()
  replyTailLoaded.value = !hasMoreInitialReplies()
  replyAutoLoadAfter.value = true
  activePostNo.value = firstPostNo(page.props.posts) || 1
  syncProgressForPostNo(activePostNo.value)
  replyWindowError.value = ''
  editingReplyId.value = 0
}

function firstReplyId(items: PostPayload[]) {
  return items.length ? items[0].id : 0
}

function lastReplyId(items: PostPayload[]) {
  return items.length ? items[items.length - 1].id : 0
}

function firstPostNo(items: PostPayload[]) {
  return items.length ? items[0].postNo || 0 : 0
}

function lastPostNo(items: PostPayload[]) {
  return items.length ? items[items.length - 1].postNo || 0 : 0
}

function initialMaxPostNo() {
  return Math.max(page.props.topic.maxPostNo || 0, page.props.topic.replyCount || 0, lastPostNo(page.props.posts))
}

function clampPostNo(postNo: number) {
  const maxPostNo = Math.max(1, replyMaxRange.value || 1)
  return Math.min(maxPostNo, Math.max(1, Math.round(postNo)))
}

function progressForPostNo(postNo: number) {
  return progressForPostNoFraction(postNo, 0.5)
}

function progressForPostNoFraction(postNo: number, fraction: number) {
  const maxPostNo = Math.max(1, replyMaxRange.value || 1)
  if (maxPostNo <= 1) return Math.min(1, Math.max(0, fraction))
  return Math.min(1, Math.max(0, (Math.max(1, postNo) - 1 + Math.min(1, Math.max(0, fraction))) / maxPostNo))
}

function visibleSlotSize() {
  return 1 / Math.max(1, replyMaxRange.value || 1)
}

function syncProgressForPostNo(postNo: number) {
  const progress = progressForPostNo(postNo)
  replyRailProgressCurrent.value = progress
  replyRailProgressStart.value = Math.max(0, progress - visibleSlotSize() / 2)
  replyRailProgressEnd.value = Math.min(1, progress + visibleSlotSize() / 2)
}

function findClosestLoadedReply(postNo: number) {
  let closest: PostPayload | undefined
  let closestDistance = Number.POSITIVE_INFINITY
  for (const reply of replies.value) {
    if (!reply.postNo) continue
    const distance = Math.abs(reply.postNo - postNo)
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
  return page.props.topic.replyCount > page.props.posts.length
}

function findReplyHashId() {
  const match = window.location.hash.match(/^#post-(\d+)$/)
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
  const element = document.getElementById(`post-${replyId}`)
  if (element) {
    scrollReplyIntoComfortView(element, 'auto')
  }
}

function highlightReply(replyId: number) {
  highlightedReplyId.value = replyId
  window.clearTimeout(highlightTimer)
  highlightTimer = window.setTimeout(() => {
    highlightedReplyId.value = null
  }, 2400)
}

function mergeReplies(nextReplies: PostPayload[], mode: 'replace' | 'prepend' | 'append') {
  if (mode === 'replace') {
    replies.value = nextReplies
    return
  }

  const seen = new Set(replies.value.map((reply) => reply.id))
  const filtered = nextReplies.filter((reply) => !seen.has(reply.id))
  replies.value = mode === 'prepend' ? [...filtered, ...replies.value] : [...replies.value, ...filtered]
}

function applyPostWindowPayload(
  payload: Awaited<ReturnType<typeof getPostWindow>>,
  mergeMode: 'replace' | 'prepend' | 'append',
  forceWindowMode: boolean,
) {
  replyWindowMode.value = forceWindowMode || replyWindowMode.value
  mergeReplies(payload.posts, mergeMode)
  replyHasBefore.value = replyWindowMode.value ? payload.hasBefore : false
  replyHasAfter.value = payload.hasAfter
  replyBeforeCursor.value = payload.beforeCursor ?? firstReplyId(replies.value)
  replyAfterCursor.value = payload.afterCursor ?? lastReplyId(replies.value)
  replyBeforePostNo.value = payload.beforePostNo ?? firstPostNo(replies.value)
  replyAfterPostNo.value = payload.afterPostNo ?? lastPostNo(replies.value)
  replyMaxNo.value = Math.max(replyMaxNo.value, payload.maxPostNo || 0)
  if (mergeMode === 'replace') {
    replyTailLoaded.value = payloadEndsAtTail(payload)
  } else if (mergeMode === 'append' && payloadEndsAtTail(payload)) {
    replyTailLoaded.value = true
  }
}

function payloadEndsAtTail(payload: Awaited<ReturnType<typeof getPostWindow>>) {
  const afterPostNo = payload.afterPostNo || lastPostNo(payload.posts)
  const maxPostNo = Math.max(replyMaxNo.value, payload.maxPostNo || 0)
  return payload.posts.length > 0 && !payload.hasAfter && afterPostNo >= maxPostNo
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
    const payload = await getPostWindow({
      topicId: page.props.topic.id,
      anchorPostId: direction === 'anchor' ? anchorValue : undefined,
      beforePostNo: direction === 'before' ? replyBeforePostNo.value : undefined,
      afterPostNo: direction === 'after' ? replyAfterPostNo.value : undefined,
      before: direction === 'before' && !replyBeforePostNo.value ? replyBeforeCursor.value : undefined,
      after: direction === 'after' && !replyAfterPostNo.value ? replyAfterCursor.value : undefined,
      tail: direction === 'tail',
      limit: 20,
    })

    applyPostWindowPayload(
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
      activePostNo.value = firstPostNo(payload.posts) || firstPostNo(replies.value)
      syncProgressForPostNo(activePostNo.value || 1)
    } else if (direction === 'tail') {
      activePostNo.value = lastPostNo(payload.posts) || lastPostNo(replies.value) || replyMaxRange.value
      syncProgressForPostNo(activePostNo.value || 1)
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

async function jumpToPostNo(postNo: number) {
  const target = clampPostNo(postNo)
  if (target >= replyMaxRange.value) {
    await jumpToLatestReply()
    return
  }

  if (loadingReplyWindow.value) {
    pendingReplyJumpNo = target
    activePostNo.value = target
    syncProgressForPostNo(target)
    return
  }

  disableReplyAutoLoadAfter()
  activePostNo.value = target
  syncProgressForPostNo(target)
  pauseReplyRailSync()
  const loaded = replies.value.find((reply) => reply.postNo === target)
  if (loaded) {
    activePostNo.value = loaded.postNo
    syncProgressForPostNo(loaded.postNo)
    await nextTick()
    const element = document.getElementById(`post-${loaded.id}`)
    if (element) {
      scrollReplyIntoComfortView(element)
    }
    resumeReplyRailSyncWhenSettled()
    return
  }

  loadingReplyWindow.value = true
  loadingReplyDirection.value = 'anchor'
  replyWindowError.value = ''
  try {
    const payload = await getPostWindow({
      topicId: page.props.topic.id,
      anchorPostNo: target,
      limit: 20,
    })
    applyPostWindowPayload(payload, 'replace', true)
    await nextTick()
    const closest = findClosestLoadedReply(target)
    if (closest) {
      activePostNo.value = closest.postNo
      syncProgressForPostNo(closest.postNo)
      collectReplyElements()
      const element = document.getElementById(`post-${closest.id}`)
      if (element) {
        scrollReplyIntoComfortView(element)
      }
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
    activePostNo.value = replyMaxRange.value
    syncProgressForPostNo(replyMaxRange.value)
    return
  }
  disableReplyAutoLoadAfter()
  activePostNo.value = replyMaxRange.value
  syncProgressForPostNo(replyMaxRange.value)
  pauseReplyRailSync()
  if (replyTailLoaded.value) {
    const latest = replies.value[replies.value.length - 1]
    if (latest) {
      activePostNo.value = latest.postNo
      syncProgressForPostNo(latest.postNo)
      await nextTick()
      scrollReplyListEndIntoView()
      resumeReplyRailSyncWhenSettled()
    }
    return
  }
  const loadedLatest = replies.value.find((reply) => reply.postNo === replyMaxRange.value)
  if (loadedLatest) {
    activePostNo.value = loadedLatest.postNo
    syncProgressForPostNo(loadedLatest.postNo)
    await nextTick()
    scrollReplyListEndIntoView()
    resumeReplyRailSyncWhenSettled()
    return
  }
  await loadReplyWindow('tail')
  await nextTick()
  const latest = replies.value[replies.value.length - 1]
  if (latest) {
    activePostNo.value = latest.postNo
    syncProgressForPostNo(latest.postNo)
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
    document.getElementById(`post-${latest.id}`)?.scrollIntoView({ block: 'end', behavior: 'smooth' })
  }
}

function flushPendingReplyJump() {
  if (!pendingReplyJumpNo || loadingReplyWindow.value) return
  const postNo = pendingReplyJumpNo
  pendingReplyJumpNo = null
  void jumpToPostNo(postNo)
}

function jumpToTopicBody() {
  void jumpToPostNo(1)
}

function focusReplyEditor() {
  mobileReplyRailOpen.value = false
  composerOpen.value = true
}

function updateComposerOpen(open: boolean) {
  composerOpen.value = open
  if (!open && editingReplyId.value) {
    cancelEditReply()
  }
}

function openFloatingReply() {
  if (editingReplyId.value) {
    cancelEditReply()
  }
  replyTargetId.value = 0
  focusReplyEditor()
}

function closeMobileReplyRail() {
  mobileReplyRailOpen.value = false
}

async function selectReplyFromRail(postNo: number) {
  closeMobileReplyRail()
  await jumpToPostNo(postNo)
}

async function jumpToLatestReplyFromRail() {
  closeMobileReplyRail()
  await jumpToLatestReply()
}

function jumpToTopicBodyFromRail() {
  closeMobileReplyRail()
  jumpToTopicBody()
}

function isElementMostlyVisible(element: HTMLElement) {
  const rect = element.getBoundingClientRect()
  return rect.top >= 96 && rect.bottom <= window.innerHeight - 120
}

function scrollReplyIntoComfortView(element: HTMLElement, behavior: ScrollBehavior = 'smooth') {
  const targetTop = element.getBoundingClientRect().top + window.scrollY - 160
  window.scrollTo({
    top: Math.max(0, targetTop),
    behavior,
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
    const element = document.getElementById(`post-${replyId}`)
    if (element) return element
  }
  return null
}

async function revealCreatedReply(replyId: number) {
  if (!replyId) return

  pauseReplyRailSync()
  const payload = await getPostWindow({
    topicId: page.props.topic.id,
    anchorPostId: replyId,
    limit: 20,
  })
  applyPostWindowPayload(payload, 'replace', true)
  const createdReply = payload.posts.find((reply) => reply.id === replyId)
  if (createdReply?.postNo) {
    activePostNo.value = createdReply.postNo
    syncProgressForPostNo(createdReply.postNo)
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
    await likeTopic(page.props.topic.id, nextLiked ? 1 : 2)
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
    await bookmarkTopic(page.props.topic.id, nextBookmarked ? 1 : 2)
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
    await watchTopic(page.props.topic.id, nextWatched ? 1 : 2)
    actionMessage.value = nextWatched ? t('article.watchAdded') : t('article.watchRemoved')
  } catch (error) {
    isWatched.value = previousWatched
    actionMessage.value = error instanceof Error ? error.message : t('api.watchFailed')
  } finally {
    actingWatch.value = false
  }
}

function replyTo(reply: PostPayload) {
  if (editingReplyId.value) {
    cancelEditReply()
  }
  replyTargetId.value = reply.id
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

function startEditReply(reply: PostPayload) {
  if (savingEditReplyId.value || deletingReplyId.value === reply.id) return
  if (!editingReplyId.value) {
    replyDraftBeforeEdit.value = replyContent.value
    replyTargetBeforeEdit.value = replyTargetId.value
  }
  replyTargetId.value = 0
  editingReplyId.value = reply.id
  replyContent.value = reply.content
  errorMessage.value = ''
  successMessage.value = ''
  focusReplyEditor()
}

function cancelEditReply() {
  if (savingEditReplyId.value) return
  editingReplyId.value = 0
  errorMessage.value = ''
  replyContent.value = replyDraftBeforeEdit.value
  replyTargetId.value = replyTargetBeforeEdit.value
  replyDraftBeforeEdit.value = ''
  replyTargetBeforeEdit.value = 0
}

async function saveReplyEdit() {
  if (savingEditReplyId.value) return

  const reply = replies.value.find((item) => item.id === editingReplyId.value)
  if (!reply) {
    cancelEditReply()
    return
  }

  const content = replyContent.value.trim()
  if (!content) {
    errorMessage.value = t('article.replyRequired')
    return
  }
  if (content === reply.content.trim()) {
    cancelEditReply()
    composerOpen.value = false
    return
  }

  savingEditReplyId.value = reply.id
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const updated = await updatePost(reply.id, content)
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
    replyContent.value = replyDraftBeforeEdit.value
    replyTargetId.value = replyTargetBeforeEdit.value
    replyDraftBeforeEdit.value = ''
    replyTargetBeforeEdit.value = 0
    composerOpen.value = false
    pushFlash(t('article.replyUpdated'), 'success')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : t('api.replyUpdateFailed')
  } finally {
    savingEditReplyId.value = 0
  }
}

async function submitReply() {
  if (editingReplyId.value) {
    await saveReplyEdit()
    return
  }

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
    const createdReply = await createPost(page.props.topic.id, content, replyId)
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

function requestDeleteReply(reply: PostPayload) {
  if (savingEditReplyId.value === reply.id) return
  pendingDeleteReply.value = reply
  deleteErrorMessage.value = ''
}

function closeDeleteDialog() {
  if (deletingReplyId.value) return
  pendingDeleteReply.value = null
  deleteErrorMessage.value = ''
}

function requestTopicModeration(action: 'ban' | 'unban') {
  actionMessage.value = ''
  pendingModerationAction.value = action
}

function closeTopicModerationDialog() {
  if (actingModeration.value) return
  pendingModerationAction.value = null
}

async function updateTopicModerationFromDetail() {
  if (actingModeration.value || !pendingModerationAction.value) return
  actingModeration.value = true
  actionMessage.value = ''
  const action = pendingModerationAction.value
  try {
    await updateModerationTopicStatus(page.props.topic.id, action)
    topicProcessStatus.value = action === 'ban' ? 1 : 0
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

function requestReport(target: { targetType: 'topic' | 'post'; targetId: number; title: string; excerpt: string }) {
  if (!page.layout.viewer.isAuthenticated) {
    openLogin()
    return
  }
  pendingReport.value = target
  reportReason.value = 'spam'
  reportNote.value = ''
  reportError.value = ''
}

function handleMarkdownImageClick(event: MouseEvent) {
  const target = event.target
  if (!(target instanceof HTMLElement)) return

  const image = target.closest('.gf-prose-article img, .gf-prose-comment img')
  if (!(image instanceof HTMLImageElement)) return

  const imageSrc = image.currentSrc || image.src
  if (!imageSrc) return

  const anchor = image.closest('a')
  if (anchor && !sameUrl(anchor.href, imageSrc)) return

  event.preventDefault()
  event.stopPropagation()

  const markdownImages = Array.from(document.querySelectorAll<HTMLImageElement>('.gf-prose-article img, .gf-prose-comment img'))
    .map((item) => ({
      src: item.currentSrc || item.src,
      alt: item.alt || '',
    }))
    .filter((item) => item.src)
  const index = markdownImages.findIndex((item) => sameUrl(item.src, imageSrc))

  markdownImageViewer.value?.open(markdownImages, index >= 0 ? index : 0)
}

function sameUrl(left: string, right: string) {
  try {
    return new URL(left, window.location.href).href === new URL(right, window.location.href).href
  } catch {
    return left === right
  }
}

function requestTopicReport() {
  requestReport({
    targetType: 'topic',
    targetId: page.props.topic.id,
    title: page.props.topic.title,
    excerpt: page.props.topic.description,
  })
}

function requestReplyReport(reply: PostPayload) {
  requestReport({
    targetType: 'post',
    targetId: reply.id,
    title: t('article.replyReportTitle', { no: reply.postNo || reply.id }),
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

async function moderateReply(reply: PostPayload, action: 'ban' | 'unban') {
  if (replyModerationBusy(reply.id)) return
  moderatingReplyIds.value = [...moderatingReplyIds.value, reply.id]
  try {
    await updateModerationPostStatus(reply.id, action)
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
    await deletePost(replyId)
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
    <article class="min-w-0" @click="handleMarkdownImageClick">
      <header ref="topicHeaderEl" class="relative z-10 border-b border-line/70 px-4 py-4 sm:mb-4 sm:px-0 sm:pb-4 sm:pt-0 xl:w-[calc(100%+292px)]">
        <h1 ref="titleEl" class="break-words text-2xl font-bold leading-tight text-base-content [overflow-wrap:anywhere] sm:text-3xl">{{ page.props.topic.title }}</h1>
        <div class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2 text-[13px] text-base-content/55">
          <a
            :href="`/u/${page.props.topic.author.id}`"
            class="inline-flex items-center gap-2 font-medium text-base-content/75 hover:text-primary"
            @click="showUserCard(page.props.topic.author, $event)"
          >
            <UserAvatar :src="page.props.topic.author.avatarUrl" :alt="page.props.topic.author.username" class="h-5 w-5 rounded-full object-cover" />
            {{ page.props.topic.author.username }}
          </a>
          <span class="inline-flex items-center gap-1.5">
            <Clock class="h-3.5 w-3.5" />
            {{ formatDateTime(page.props.topic.createdAt) }}
          </span>
          <a
            v-for="category in page.props.topic.categories"
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
                :href="`/u/${page.props.topic.author.id}`"
                class="sticky top-19 self-start pt-1"
                @click="showUserCard(page.props.topic.author, $event)"
              >
                <UserAvatar :src="page.props.topic.author.avatarUrl" :alt="page.props.topic.author.username" :badge="page.props.topic.author.wornBadge" class="h-11 w-11 rounded-full ring-1 ring-line" img-class="rounded-full" />
              </a>
              <div class="min-w-0">
                <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <a :href="`/u/${page.props.topic.author.id}`" class="font-semibold text-base-content hover:text-primary">{{ page.props.topic.author.username }}</a>
                    <div class="text-xs font-medium text-base-content/75">{{ t('article.body') }}</div>
                  </div>
                  <div class="flex flex-wrap items-center justify-end gap-3 text-xs font-medium text-base-content/75">
                    <div class="flex items-center gap-3">
                      <span class="inline-flex items-center gap-1"><MessageSquare class="h-3.5 w-3.5" />{{ formatNumber(page.props.topic.replyCount) }}</span>
                      <span class="inline-flex items-center gap-1"><Eye class="h-3.5 w-3.5" />{{ formatNumber(page.props.topic.viewCount) }}</span>
                      <span class="inline-flex items-center gap-1"><Heart class="h-3.5 w-3.5" />{{ formatNumber(likeCount) }}</span>
                    </div>
                    <a
                      v-if="page.props.permissions.isOwnTopic"
                      :href="`/publish?id=${page.props.topic.id}`"
                      class="gf-button gf-button-secondary h-7 px-2 text-xs hover:border-primary/20 hover:bg-info/10 hover:text-primary"
                    >
                      <PencilLine class="h-3.5 w-3.5" />
                      {{ t('common.edit') }}
                    </a>
                  </div>
                </div>
                <div class="gf-prose gf-prose-article" v-html="page.props.topic.html" />
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
                    v-if="!page.props.permissions.isOwnTopic"
                    type="button"
                    class="gf-button gf-button-sm px-2.5 text-base-content/55 hover:bg-warning/10 hover:text-warning"
                    @click="requestTopicReport"
                  >
                    <Flag class="h-4 w-4" />
                    {{ t('article.report') }}
                  </button>
                  <button
                    v-if="page.props.permissions.canModerateTopic && topicProcessStatus === 0"
                    type="button"
                    class="gf-button gf-button-sm px-2.5 text-base-content/55 hover:bg-base-200 hover:text-base-content"
                    :disabled="actingModeration"
                    @click="requestTopicModeration('ban')"
                  >
                    <Ban class="h-4 w-4" />
                    {{ t('article.moderationBan') }}
                  </button>
                  <button
                    v-else-if="page.props.permissions.canModerateTopic && topicProcessStatus === 1"
                    type="button"
                    class="gf-button gf-button-sm px-2.5 text-base-content/55 hover:bg-base-200 hover:text-base-content"
                    :disabled="actingModeration"
                    @click="requestTopicModeration('unban')"
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
              :id="`post-${reply.id}`"
              :key="reply.id"
              :data-post-no="reply.postNo"
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
                      <span v-if="reply.postNo" class="hidden shrink-0 text-xs font-semibold tabular-nums text-base-content/55 sm:inline">#{{ formatNumber(reply.postNo) }}</span>
                    </div>
                    <div class="mt-0.5 flex items-center gap-2 text-xs text-base-content/55 sm:hidden">
                      <span v-if="reply.postNo" class="font-semibold tabular-nums text-base-content/55">#{{ formatNumber(reply.postNo) }}</span>
                      <time class="truncate">{{ formatDateTime(reply.createdAt) }}</time>
                    </div>
                  </div>
                  <div class="flex shrink-0 items-center gap-0.5 sm:gap-1.5">
                    <button
                      v-if="reply.isOwnPost && !reply.isHidden"
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
                      v-if="reply.isOwnPost && !reply.isHidden"
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
                      v-if="page.props.permissions.canPost && !reply.isHidden"
                      type="button"
                      class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-icon-muted transition hover:bg-info/10 hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2"
                      :title="t('article.reply')"
                      @click="replyTo(reply)"
                    >
                      <CornerDownLeft class="h-3.5 w-3.5" />
                      <span class="sr-only">{{ t('article.reply') }}</span>
                    </button>
                    <button
                      v-if="!reply.isOwnPost && !reply.isHidden"
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
                  <dd class="text-right font-semibold tabular-nums text-base-content">{{ formatNumber(page.props.topic.replyCount) }}</dd>
                </div>
                <div class="flex items-center justify-between gap-4">
                  <dt class="font-semibold text-base-content/55">{{ t('article.viewCount') }}</dt>
                  <dd class="text-right font-semibold tabular-nums text-base-content">{{ formatNumber(page.props.topic.viewCount) }}</dd>
                </div>
                <div class="flex items-center justify-between gap-4">
                  <dt class="font-semibold text-base-content/55">{{ t('article.participants') }}</dt>
                  <dd class="text-right font-semibold tabular-nums text-base-content">{{ page.props.topic.participants.length }}</dd>
                </div>
              </dl>

              <div v-if="page.props.topic.participants.length" class="border-t border-line px-4 py-4">
                <h3 class="mb-3 text-sm font-semibold text-base-content/55">{{ t('article.activeParticipants') }}</h3>
                <div class="flex flex-wrap gap-1.5">
                  <a
                    v-for="participant in page.props.topic.participants"
                    :key="participant.id"
                    :href="`/u/${participant.id}`"
                    class="rounded-full"
                    @click="showUserCard(participant, $event)"
                  >
                    <UserAvatar :src="participant.avatarUrl" :alt="participant.username" class="h-8 w-8 rounded-full object-cover ring-1 ring-line transition hover:ring-primary/40" />
                  </a>
                </div>
              </div>

              <PostPositionRail
                v-if="page.props.topic.replyCount > 0 && replyMaxRange > 0"
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
                @earliest="jumpToTopicBodyFromRail"
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

      <PostComposer
        v-model="replyContent"
        v-model:mobile-rail-open="mobileReplyRailOpen"
        :open="composerOpen"
        :actions="floatingTopicActions"
        :authenticated="page.layout.viewer.isAuthenticated"
        :can-post="page.props.permissions.canPost"
        :current-label="replyRailCurrentLabel"
        :current-no="replyRailCurrentNo"
        :end-label="replyRailEndLabel"
        :error-message="errorMessage"
        :has-rail="hasReplyRail"
        :max-no="replyMaxRange"
        :mode="composerMode"
        :progress-current="replyRailProgressCurrent"
        :progress-end="replyRailProgressEnd"
        :progress-start="replyRailProgressStart"
        :rail-busy="replyRailBusy"
        :start-label="replyRailStartLabel"
        :submitting="editingReplyId ? savingEditReplyId > 0 : submitting"
        :success-message="successMessage"
        :target="replyTarget"
        @clear-target="cancelReplyTarget"
        @clear-validation="clearReplyValidation"
        @earliest="jumpToTopicBodyFromRail"
        @image-error="handleReplyImageError"
        @image-inserted="handleReplyImageInserted"
        @latest="jumpToLatestReplyFromRail"
        @open-reply="openFloatingReply"
        @select-rail="selectReplyFromRail"
        @submit="submitReply"
        @update:open="updateComposerOpen"
      />

    </article>

    <MarkdownImageViewer ref="markdownImageViewer" />

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
          aria-labelledby="ban-topic-title"
          @click.self="closeTopicModerationDialog"
        >
          <div class="gf-menu-surface w-full max-w-sm p-4">
            <div class="flex items-start gap-3">
              <AlertTriangle class="mt-0.5 h-5 w-5 shrink-0 text-error" />
              <div class="min-w-0 flex-1">
                <h2 id="ban-topic-title" class="text-base font-bold text-base-content">
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
                @click="closeTopicModerationDialog"
              >
                <X class="h-4 w-4" />
              </button>
            </div>

            <div class="mt-4 flex justify-end gap-2">
              <button
                type="button"
                class="gf-button gf-button-md gf-button-muted"
                :disabled="actingModeration"
                @click="closeTopicModerationDialog"
              >
                {{ t('common.cancel') }}
              </button>
              <button
                type="button"
                class="gf-button gf-button-md gf-button-danger"
                :disabled="actingModeration"
                @click="updateTopicModerationFromDetail"
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
