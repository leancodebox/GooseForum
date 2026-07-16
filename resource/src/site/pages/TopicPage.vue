<script setup lang="ts">
import { computed, defineAsyncComponent, nextTick, onBeforeUnmount, onMounted, ref, Teleport, watch } from 'vue'
import { AlertTriangle, Ban, Bell, Bookmark, ChevronsUp, Clock, CornerDownLeft, Eye, Flag, Heart, Loader2, MessageSquare, PencilLine, RotateCcw, Trash2, X } from '@lucide/vue'
import { bookmarkTopic, deletePost, getPostWindow, likeTopic, createPost, submitReport, updateModerationTopicStatus, updateModerationPostStatus, updatePost, watchTopic } from '@/runtime/api'
import { formatDateTime, formatNumber } from '@/runtime/format'
import { useFlashMessages } from '@/runtime/flash-message'
import { fetchPage } from '@/runtime/router'
import { useShellState } from '@/runtime/shell-state'
import { showUserCard } from '@/runtime/user-card-events'
import { measurePostViewportProgressFromRects } from '@/runtime/post-viewport-progress'
import MarkdownImageViewer from '@/site/components/MarkdownImageViewer.vue'
import PostPositionRail from '@/site/components/PostPositionRail.vue'
import TopicFloatingControls from '@/site/components/TopicFloatingControls.vue'
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
const PostComposer = defineAsyncComponent(() => import('@/site/components/PostComposer.vue'))
const initialPostStream = page.props.postStream
const initialPosts = initialPostStream.posts
const postContent = ref('')
const targetPostId = ref(0)
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
const deletingPostId = ref(0)
const editingPostId = ref(0)
const savingEditPostId = ref(0)
const postDraftBeforeEdit = ref('')
const targetPostBeforeEdit = ref(0)
const pendingDeletePost = ref<PostPayload | null>(null)
const pendingModerationAction = ref<'ban' | 'unban' | null>(null)
const pendingReport = ref<{ targetType: 'topic' | 'post'; targetId: number; title: string; excerpt: string } | null>(null)
const reportReason = ref('spam')
const reportNote = ref('')
const reportSubmitting = ref(false)
const reportError = ref('')
const moderatingPostIds = ref<number[]>([])
const posts = ref<PostPayload[]>([...initialPosts])
const topicProcessStatus = ref(page.props.topic.processStatus)
const targetPost = computed(() => posts.value.find((post) => post.id === targetPostId.value))
const postHasBefore = ref(initialPostStream.hasBefore)
const postHasAfter = ref(initialPostStream.hasAfter)
const postBeforePostNo = ref(initialPostStream.beforePostNo || firstPostNo(initialPosts))
const postAfterPostNo = ref(initialPostStream.afterPostNo || lastPostNo(initialPosts))
const postMaxNo = ref(initialPostStream.maxPostNo || initialMaxPostNo())
const postAutoLoadAfter = ref(true)
const loadingPostWindow = ref(false)
const loadingPostDirection = ref<'before' | 'after' | 'anchor' | null>(null)
const postWindowError = ref('')
const deleteErrorMessage = ref('')
const errorMessage = ref('')
const successMessage = ref('')
const topicHeaderEl = ref<HTMLElement | null>(null)
const titleEl = ref<HTMLElement | null>(null)
const postLoadMoreEl = ref<HTMLElement | null>(null)
const markdownImageViewer = ref<InstanceType<typeof MarkdownImageViewer> | null>(null)
const topicRailTopOffset = ref(0)
const showHeaderTitle = ref(false)
const isMobileHeaderViewport = ref(false)
const mobileHeaderTitleVisible = ref(false)
const effectiveShowHeaderTitle = computed(() => showHeaderTitle.value && (!isMobileHeaderViewport.value || mobileHeaderTitleVisible.value))
const composerOpen = ref(false)
const composerMode = computed(() => editingPostId.value ? 'edit' : 'create')
const composerMounted = computed(() => composerOpen.value)
const mobilePostRailOpen = ref(false)
const activePostNo = ref(firstPostNo(initialPosts) || 1)
const postRailProgressCurrent = ref(0)
const postRailProgressStart = ref(0)
const postRailProgressEnd = ref(0)
const postMaxRange = computed(() => Math.max(postMaxNo.value, ...posts.value.map((post) => post.postNo || 0)))
const hasPostRail = computed(() => postMaxRange.value > 0)
const postRailCurrentNo = computed(() => {
  const fallback = firstPostNo(posts.value) || 1
  return clampPostNo(activePostNo.value || fallback)
})
const postRailCurrentLabel = computed(() => {
  const activePost = posts.value.find((post) => post.postNo === postRailCurrentNo.value)
  return activePost ? formatRailDate(activePost.createdAt) : formatRailDate(page.props.topic.createdAt)
})
const postRailStartLabel = computed(() => formatRailDate(page.props.topic.createdAt))
const postRailEndLabel = computed(() => formatRailDate(postHasAfter.value ? page.props.topic.updatedAt : posts.value[posts.value.length - 1]?.createdAt || page.props.topic.updatedAt))
const postRailBusy = computed(() => navigationPhase.value !== 'idle' || (loadingPostWindow.value && loadingPostDirection.value === 'anchor'))
const actionMessageSuccess = computed(() =>
  [
    t('topic.bookmarkAdded'),
    t('topic.bookmarkRemoved'),
    t('topic.watchAdded'),
    t('topic.watchRemoved'),
    t('topic.moderationBanSuccess'),
    t('topic.moderationUnbanSuccess'),
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
      title: t('topic.like'),
      activeClass: 'bg-error/10 text-error hover:bg-error/10',
      onClick: toggleLike,
    },
    {
      key: 'bookmark',
      icon: Bookmark,
      active: isBookmarked.value,
      acting: actingBookmark.value,
      fill: true,
      title: isBookmarked.value ? t('topic.bookmarked') : t('topic.bookmark'),
      activeClass: 'bg-info/10 text-primary hover:bg-info/10',
      onClick: toggleBookmark,
    },
    {
      key: 'watch',
      icon: Bell,
      active: isWatched.value,
      acting: actingWatch.value,
      fill: true,
      title: isWatched.value ? t('topic.watched') : t('topic.watch'),
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
      title: isBanned ? t('topic.moderationUnban') : t('topic.moderationBan'),
      activeClass: 'text-base-content/75 hover:bg-base-200 hover:text-base-content',
      onClick: async () => requestTopicModeration(isBanned ? 'unban' : 'ban'),
    })
  }

  return actions
})
const shellState = useShellState()
let titleObserver: IntersectionObserver | undefined
let topicHeaderResizeObserver: ResizeObserver | undefined
let postLoadObserver: IntersectionObserver | undefined
let lastHeaderScrollY = 0
let headerScrollFrame = 0
const highlightedPostId = ref<number | null>(null)
let highlightTimer: number | undefined
let postBottomLoadFrame = 0
let activePostScrollFrame = 0
const navigationPhase = ref<'idle' | 'loading' | 'scrolling'>('idle')
const navigationTargetPostNo = ref(0)
const navigationTargetPostId = ref(0)
let postRailResumeFrame = 0
let postRailResumeLastScrollY = 0
let postRailResumeStableFrames = 0
let postElements: HTMLElement[] = []
const postNavigationTargetTop = 160

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
  void nextTick(observePostLoader)
  void nextTick(collectPostElements)
  void nextTick(scheduleActivePostFromScroll)
  setupPostBottomLoadFallback()
  void syncPostHash()
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
    resetPostsFromProps()
    mobilePostRailOpen.value = false
    void nextTick(observeTopicHeader)
    void nextTick(observeTitle)
    void nextTick(observePostLoader)
    void nextTick(collectPostElements)
    void nextTick(scheduleActivePostFromScroll)
    void nextTick(syncPostHash)
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
  () => posts.value.map((post) => `${post.id}:${post.postNo}`).join(','),
  () => {
    void nextTick(() => {
      collectPostElements()
      scheduleActivePostFromScroll()
    })
  },
)

onBeforeUnmount(() => {
  titleObserver?.disconnect()
  topicHeaderResizeObserver?.disconnect()
  postLoadObserver?.disconnect()
  window.removeEventListener('scroll', updateMobileHeaderTitle)
  window.removeEventListener('scroll', scheduleActivePostFromScroll)
  window.removeEventListener('scroll', schedulePostBottomLoadCheck)
  window.removeEventListener('resize', updateTopicRailTopOffset)
  window.removeEventListener('resize', updateHeaderViewport)
  window.removeEventListener('resize', scheduleActivePostFromScroll)
  window.removeEventListener('resize', schedulePostBottomLoadCheck)
  window.cancelAnimationFrame(headerScrollFrame)
  window.cancelAnimationFrame(postBottomLoadFrame)
  window.cancelAnimationFrame(activePostScrollFrame)
  window.cancelAnimationFrame(postRailResumeFrame)
  navigationTargetPostId.value = 0
  window.clearTimeout(highlightTimer)
  shellState.headerTitle = ''
  shellState.headerTags = []
  shellState.showHeaderTitle = false
})

function setupHeaderTitleBehavior() {
  lastHeaderScrollY = window.scrollY
  updateHeaderViewport()
  window.addEventListener('scroll', updateMobileHeaderTitle, { passive: true })
  window.addEventListener('scroll', scheduleActivePostFromScroll, { passive: true })
  window.addEventListener('resize', updateTopicRailTopOffset)
  window.addEventListener('resize', updateHeaderViewport)
  window.addEventListener('resize', scheduleActivePostFromScroll)
}

function setupPostBottomLoadFallback() {
  window.addEventListener('scroll', schedulePostBottomLoadCheck, { passive: true })
  window.addEventListener('resize', schedulePostBottomLoadCheck)
}

function schedulePostBottomLoadCheck() {
  if (postBottomLoadFrame) return
  postBottomLoadFrame = window.requestAnimationFrame(() => {
    postBottomLoadFrame = 0
    void maybeLoadRepliesNearViewportEdge()
  })
}

function isNearDocumentBottom() {
  const documentElement = document.documentElement
  const fullHeight = Math.max(documentElement.scrollHeight, document.body?.scrollHeight || 0)
  return fullHeight - (window.scrollY + window.innerHeight) <= 480
}

async function maybeLoadRepliesNearViewportEdge() {
  if (loadingPostWindow.value || postWindowError.value) return

  if (!postHasAfter.value || !isNearDocumentBottom()) return

  postAutoLoadAfter.value = true
  await loadPostWindow('after')
  await nextTick()
  if (postHasAfter.value && isNearDocumentBottom()) {
    schedulePostBottomLoadCheck()
  }
}

async function loadMoreRepliesManually() {
  postAutoLoadAfter.value = true
  await loadPostWindow('after')
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

function observePostLoader() {
  observePostAfterLoader()
}

function observePostAfterLoader() {
  postLoadObserver?.disconnect()
  if (!postLoadMoreEl.value || !postHasAfter.value || !postAutoLoadAfter.value || !('IntersectionObserver' in window)) return

  postLoadObserver = new IntersectionObserver(
    (entries) => {
      if (entries[0]?.isIntersecting && postHasAfter.value && postAutoLoadAfter.value && !loadingPostWindow.value && !postWindowError.value) {
        void loadPostWindow('after')
      }
    },
    { rootMargin: '360px 0px' },
  )
  postLoadObserver.observe(postLoadMoreEl.value)
}

function collectPostElements() {
  postElements = Array.from(document.querySelectorAll<HTMLElement>('[data-post-no]'))
}

function pausePostRailSync() {
  navigationPhase.value = 'scrolling'
  window.cancelAnimationFrame(postRailResumeFrame)
  postRailResumeFrame = 0
  postRailResumeLastScrollY = window.scrollY
  postRailResumeStableFrames = 0
}

function keepNavigationTargetPinned() {
  if (!navigationTargetPostId.value) return false

  const element = document.getElementById(`post-${navigationTargetPostId.value}`)
  if (!element) return false

  const delta = element.getBoundingClientRect().top - postNavigationTargetTop
  if (Math.abs(delta) < 1) return false

  window.scrollBy({ top: delta, behavior: 'auto' })
  return true
}

function resumePostRailSyncWhenSettled() {
  navigationPhase.value = 'scrolling'
  window.cancelAnimationFrame(postRailResumeFrame)
  postRailResumeFrame = 0
  postRailResumeLastScrollY = window.scrollY
  postRailResumeStableFrames = 0
  const startedAt = performance.now()
  const settle = () => {
    const pinned = keepNavigationTargetPinned()
    const currentY = window.scrollY
    if (!pinned && Math.abs(currentY - postRailResumeLastScrollY) < 1) {
      postRailResumeStableFrames += 1
    } else {
      postRailResumeStableFrames = 0
      postRailResumeLastScrollY = currentY
    }
    if (postRailResumeStableFrames >= 8 || performance.now() - startedAt > 2600) {
      navigationPhase.value = 'idle'
      navigationTargetPostNo.value = 0
      navigationTargetPostId.value = 0
      postRailResumeFrame = 0
      syncPostRailProgress()
      return
    }
    postRailResumeFrame = window.requestAnimationFrame(settle)
  }
  postRailResumeFrame = window.requestAnimationFrame(settle)
}

function scheduleActivePostFromScroll() {
  if (navigationPhase.value !== 'idle' || activePostScrollFrame) return
  activePostScrollFrame = window.requestAnimationFrame(() => {
    activePostScrollFrame = 0
    syncPostRailProgress()
  })
}

function syncPostRailProgress() {
  const progress = measurePostViewportProgress()
  if (progress.postNo >= 0) {
    activePostNo.value = progress.postNo
    postRailProgressCurrent.value = progress.current
    postRailProgressStart.value = progress.start
    postRailProgressEnd.value = progress.end
  }
}

function measurePostViewportProgress() {
  const markerY = Math.min(window.innerHeight * 0.38, 340)
  const viewportTop = 88
  const viewportBottom = window.innerHeight - 96
  return measurePostViewportProgressFromRects({
    posts: postElements.map((element) => {
      const rect = element.getBoundingClientRect()
      return {
        postNo: Number(element.dataset.postNo || 0),
        top: rect.top,
        bottom: rect.bottom,
        height: rect.height,
      }
    }),
    markerY,
    viewportTop,
    viewportBottom,
    maxPostNo: postMaxRange.value,
    visibleSlotSize: visibleSlotSize(),
  })
}

function resetPostsFromProps() {
  posts.value = [...initialPosts]
  postHasBefore.value = initialPostStream.hasBefore
  postHasAfter.value = initialPostStream.hasAfter
  postBeforePostNo.value = initialPostStream.beforePostNo || firstPostNo(initialPosts)
  postAfterPostNo.value = initialPostStream.afterPostNo || lastPostNo(initialPosts)
  postMaxNo.value = initialPostStream.maxPostNo || initialMaxPostNo()
  postAutoLoadAfter.value = true
  navigationPhase.value = 'idle'
  navigationTargetPostNo.value = 0
  navigationTargetPostId.value = 0
  activePostNo.value = firstPostNo(initialPosts) || 1
  syncProgressForPostNo(activePostNo.value)
  postWindowError.value = ''
  editingPostId.value = 0
}

function firstPostNo(items: PostPayload[]) {
  return items.length ? items[0].postNo || 0 : 0
}

function lastPostNo(items: PostPayload[]) {
  return items.length ? items[items.length - 1].postNo || 0 : 0
}

function initialMaxPostNo() {
  return Math.max(page.props.topic.maxPostNo || 0, page.props.postStream.maxPostNo || 0, lastPostNo(initialPosts))
}

function clampPostNo(postNo: number) {
  const maxPostNo = Math.max(1, postMaxRange.value || 1)
  return Math.min(maxPostNo, Math.max(1, Math.round(postNo)))
}

function progressForPostNo(postNo: number) {
  return progressForPostNoFraction(postNo, 0.5)
}

function progressForPostNoFraction(postNo: number, fraction: number) {
  const maxPostNo = Math.max(1, postMaxRange.value || 1)
  if (maxPostNo <= 1) return Math.min(1, Math.max(0, fraction))
  return Math.min(1, Math.max(0, (Math.max(1, postNo) - 1 + Math.min(1, Math.max(0, fraction))) / maxPostNo))
}

function visibleSlotSize() {
  return 1 / Math.max(1, postMaxRange.value || 1)
}

function syncProgressForPostNo(postNo: number) {
  const progress = progressForPostNo(postNo)
  postRailProgressCurrent.value = progress
  postRailProgressStart.value = Math.max(0, progress - visibleSlotSize() / 2)
  postRailProgressEnd.value = Math.min(1, progress + visibleSlotSize() / 2)
}

function findClosestLoadedPost(postNo: number) {
  let closest: PostPayload | undefined
  let closestDistance = Number.POSITIVE_INFINITY
  for (const post of posts.value) {
    if (!post.postNo) continue
    const distance = Math.abs(post.postNo - postNo)
    if (distance < closestDistance) {
      closest = post
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

function findPostHashId() {
  const match = window.location.hash.match(/^#post-(\d+)$/)
  return match ? Number(match[1]) : 0
}

async function syncPostHash() {
  const postId = findPostHashId()
  if (!postId) return

  if (!posts.value.some((post) => post.id === postId)) {
    navigationPhase.value = 'loading'
    loadingPostWindow.value = true
    loadingPostDirection.value = 'anchor'
    postWindowError.value = ''
    try {
      const payload = await getPostWindow({
        topicId: page.props.topic.id,
        anchorPostId: postId,
        limit: 20,
      })
      applyPostWindowPayload(payload, 'replace')
      await nextTick()
      collectPostElements()
    } catch (error) {
      postWindowError.value = error instanceof Error ? error.message : t('api.repliesLoadFailed')
    } finally {
      loadingPostWindow.value = false
      loadingPostDirection.value = null
      navigationPhase.value = 'idle'
    }
  }

  highlightPost(postId)
  await nextTick()
  const element = document.getElementById(`post-${postId}`)
  if (element) {
    navigationTargetPostId.value = postId
    scrollPostIntoComfortView(element, 'auto')
    resumePostRailSyncWhenSettled()
  }
}

function highlightPost(postId: number) {
  highlightedPostId.value = postId
  window.clearTimeout(highlightTimer)
  highlightTimer = window.setTimeout(() => {
    highlightedPostId.value = null
  }, 2400)
}

function mergePosts(nextReplies: PostPayload[], mode: 'replace' | 'prepend' | 'append') {
  if (mode === 'replace') {
    posts.value = nextReplies
    return
  }

  const seen = new Set(posts.value.map((post) => post.id))
  const filtered = nextReplies.filter((post) => !seen.has(post.id))
  posts.value = mode === 'prepend' ? [...filtered, ...posts.value] : [...posts.value, ...filtered]
}

function applyPostWindowPayload(payload: Awaited<ReturnType<typeof getPostWindow>>, mergeMode: 'replace' | 'prepend' | 'append') {
  mergePosts(payload.posts, mergeMode)
  const nextMaxPostNo = Math.max(postMaxNo.value, payload.maxPostNo || 0)
  postMaxNo.value = nextMaxPostNo
  syncLoadedPostWindowBounds(payload.hasBefore, payload.hasAfter, nextMaxPostNo)
}

function syncLoadedPostWindowBounds(hasBefore = postHasBefore.value, hasAfter = postHasAfter.value, maxPostNo = postMaxNo.value) {
  const loadedFirstPostNo = firstPostNo(posts.value)
  const loadedLastPostNo = lastPostNo(posts.value)
  postHasBefore.value = hasBefore && loadedFirstPostNo > 1
  postHasAfter.value = hasAfter && loadedLastPostNo < maxPostNo
  postBeforePostNo.value = loadedFirstPostNo
  postAfterPostNo.value = loadedLastPostNo
}

function disablePostAutoLoadAfter() {
  postAutoLoadAfter.value = false
  postLoadObserver?.disconnect()
}

function firstVisiblePostElement() {
  for (const element of postElements) {
    const rect = element.getBoundingClientRect()
    if (rect.bottom > 96) return element
  }
  return postElements[0] || null
}

async function keepScrollPositionWhilePrepending<T>(operation: () => Promise<T>) {
  const anchor = firstVisiblePostElement()
  const beforeTop = anchor?.getBoundingClientRect().top ?? 0
  const result = await operation()
  await nextTick()
  collectPostElements()
  if (anchor) {
    const afterTop = anchor.getBoundingClientRect().top
    window.scrollBy({ top: afterTop - beforeTop, behavior: 'auto' })
  }
  return result
}

async function loadPostWindow(direction: 'before' | 'after') {
  if (loadingPostWindow.value) return
  if (direction === 'after' && (!postHasAfter.value || !postAutoLoadAfter.value)) return
  if (direction === 'before' && !postHasBefore.value) return

  loadingPostWindow.value = true
  loadingPostDirection.value = direction
  postWindowError.value = ''
  try {
    if (direction === 'before') {
      await keepScrollPositionWhilePrepending(async () => {
        const payload = await getPostWindow({
          topicId: page.props.topic.id,
          beforePostNo: postBeforePostNo.value,
          limit: 20,
        })
        applyPostWindowPayload(payload, 'prepend')
        return payload
      })
    } else {
      const payload = await getPostWindow({
        topicId: page.props.topic.id,
        afterPostNo: postAfterPostNo.value,
        limit: 20,
      })
      applyPostWindowPayload(payload, 'append')
      await nextTick()
      collectPostElements()
      if (!payload.hasAfter) {
        disablePostAutoLoadAfter()
      }
    }
    if (postAutoLoadAfter.value) {
      observePostLoader()
    }
    scheduleActivePostFromScroll()
  } catch (error) {
    postWindowError.value = error instanceof Error ? error.message : t('api.repliesLoadFailed')
  } finally {
    loadingPostWindow.value = false
    loadingPostDirection.value = null
  }
}

async function jumpToPostNo(postNo: number) {
  const target = clampPostNo(postNo)
  if (loadingPostWindow.value) {
    activePostNo.value = target
    syncProgressForPostNo(target)
    return
  }

  disablePostAutoLoadAfter()
  navigationPhase.value = 'loading'
  navigationTargetPostNo.value = target
  navigationTargetPostId.value = 0
  activePostNo.value = target
  syncProgressForPostNo(target)
  const loaded = posts.value.find((post) => post.postNo === target)
  if (loaded) {
    activePostNo.value = loaded.postNo
    syncProgressForPostNo(loaded.postNo)
    const element = await findPostElementAfterLayout(loaded.id)
    if (element) {
      navigationTargetPostId.value = loaded.id
      scrollPostIntoComfortView(element)
    }
    resumePostRailSyncWhenSettled()
    return
  }

  loadingPostWindow.value = true
  loadingPostDirection.value = 'anchor'
  postWindowError.value = ''
  try {
    const payload = await getPostWindow({
      topicId: page.props.topic.id,
      anchorPostNo: target,
      limit: 20,
    })
    applyPostWindowPayload(payload, 'replace')
    await nextTick()
    collectPostElements()
    const closest = findClosestLoadedPost(target)
    if (closest) {
      activePostNo.value = closest.postNo
      syncProgressForPostNo(closest.postNo)
      const element = await findPostElementAfterLayout(closest.id)
      if (element) {
        navigationTargetPostId.value = closest.id
        scrollPostIntoComfortView(element, 'auto')
      }
      resumePostRailSyncWhenSettled()
    } else {
      navigationPhase.value = 'idle'
      navigationTargetPostNo.value = 0
      navigationTargetPostId.value = 0
    }
  } catch (error) {
    postWindowError.value = error instanceof Error ? error.message : t('api.repliesLoadFailed')
    navigationPhase.value = 'idle'
    navigationTargetPostNo.value = 0
    navigationTargetPostId.value = 0
  } finally {
    loadingPostWindow.value = false
    loadingPostDirection.value = null
  }
}

async function jumpToLatestPost() {
  await jumpToPostNo(postMaxRange.value)
}

function jumpToTopicBody() {
  void jumpToPostNo(1)
}

function focusPostComposer() {
  mobilePostRailOpen.value = false
  composerOpen.value = true
}

function updateComposerOpen(open: boolean) {
  composerOpen.value = open
  if (!open && editingPostId.value) {
    cancelEditPost()
  }
}

function openFloatingPostComposer() {
  if (editingPostId.value) {
    cancelEditPost()
  }
  targetPostId.value = 0
  focusPostComposer()
}

function closeMobilePostRail() {
  mobilePostRailOpen.value = false
}

async function selectPostFromRail(postNo: number) {
  closeMobilePostRail()
  if (postNo <= 1) {
    jumpToTopicBody()
    return
  }
  await jumpToPostNo(postNo)
}

async function jumpToLatestPostFromRail() {
  closeMobilePostRail()
  await jumpToLatestPost()
}

function jumpToTopicBodyFromRail() {
  closeMobilePostRail()
  jumpToTopicBody()
}

function isElementMostlyVisible(element: HTMLElement) {
  const rect = element.getBoundingClientRect()
  return rect.top >= 96 && rect.bottom <= window.innerHeight - 120
}

function scrollPostIntoComfortView(element: HTMLElement, behavior: ScrollBehavior = 'smooth') {
  const targetTop = element.getBoundingClientRect().top + window.scrollY - postNavigationTargetTop
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

async function findPostElementAfterLayout(postId: number) {
  let lastTop: number | null = null
  let stableFrames = 0
  for (let attempts = 0; attempts < 10; attempts += 1) {
    await nextTick()
    await waitForAnimationFrame()
    const element = document.getElementById(`post-${postId}`)
    if (!element) continue

    const top = element.getBoundingClientRect().top
    if (lastTop !== null && Math.abs(top - lastTop) < 1) {
      stableFrames += 1
      if (stableFrames >= 2) return element
    } else {
      stableFrames = 0
      lastTop = top
    }
  }
  return document.getElementById(`post-${postId}`)
}

async function revealCreatedPost(postId: number) {
  if (!postId) return

  navigationPhase.value = 'loading'
  const payload = await getPostWindow({
    topicId: page.props.topic.id,
    anchorPostId: postId,
    limit: 20,
  })
  applyPostWindowPayload(payload, 'replace')
  const createdPost = payload.posts.find((post) => post.id === postId)
  if (createdPost?.postNo) {
    navigationTargetPostNo.value = createdPost.postNo
    activePostNo.value = createdPost.postNo
    syncProgressForPostNo(createdPost.postNo)
  }
  highlightPost(postId)
  const element = await findPostElementAfterLayout(postId)
  if (element && !isElementMostlyVisible(element)) {
    navigationTargetPostId.value = postId
    scrollPostIntoComfortView(element)
    resumePostRailSyncWhenSettled()
    return
  }
  navigationPhase.value = 'idle'
  navigationTargetPostNo.value = 0
  navigationTargetPostId.value = 0
  collectPostElements()
  scheduleActivePostFromScroll()
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
    actionMessage.value = nextBookmarked ? t('topic.bookmarkAdded') : t('topic.bookmarkRemoved')
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
    actionMessage.value = nextWatched ? t('topic.watchAdded') : t('topic.watchRemoved')
  } catch (error) {
    isWatched.value = previousWatched
    actionMessage.value = error instanceof Error ? error.message : t('api.watchFailed')
  } finally {
    actingWatch.value = false
  }
}

function replyTo(post: PostPayload) {
  if (editingPostId.value) {
    cancelEditPost()
  }
  targetPostId.value = post.id
  errorMessage.value = ''
  successMessage.value = ''
  focusPostComposer()
}

function cancelPostTarget() {
  targetPostId.value = 0
  errorMessage.value = ''
}

function clearPostValidation() {
  errorMessage.value = ''
  successMessage.value = ''
}

function handlePostImageInserted(count: number) {
  errorMessage.value = ''
  successMessage.value = count > 1 ? t('publish.imagesInserted', { count }) : t('publish.imageInserted')
}

function handlePostImageError(message: string) {
  errorMessage.value = message
}

function isFirstPost(post: PostPayload) {
  return post.postNo === 1
}

function canEditPost(post: PostPayload) {
  return post.isOwnPost && !post.isHidden
}

function canDeleteRenderedPost(post: PostPayload) {
  return post.isOwnPost && !post.isHidden && !isFirstPost(post)
}

function startEditPost(post: PostPayload) {
  if (savingEditPostId.value || deletingPostId.value === post.id) return
  if (isFirstPost(post)) {
    window.location.href = `/publish?id=${page.props.topic.id}`
    return
  }
  if (!editingPostId.value) {
    postDraftBeforeEdit.value = postContent.value
    targetPostBeforeEdit.value = targetPostId.value
  }
  targetPostId.value = 0
  editingPostId.value = post.id
  postContent.value = post.content
  errorMessage.value = ''
  successMessage.value = ''
  focusPostComposer()
}

function cancelEditPost() {
  if (savingEditPostId.value) return
  editingPostId.value = 0
  errorMessage.value = ''
  postContent.value = postDraftBeforeEdit.value
  targetPostId.value = targetPostBeforeEdit.value
  postDraftBeforeEdit.value = ''
  targetPostBeforeEdit.value = 0
}

async function savePostEdit() {
  if (savingEditPostId.value) return

  const post = posts.value.find((item) => item.id === editingPostId.value)
  if (!post) {
    cancelEditPost()
    return
  }

  const content = postContent.value.trim()
  if (!content) {
    errorMessage.value = t('topic.replyRequired')
    return
  }
  if (content === post.content.trim()) {
    cancelEditPost()
    composerOpen.value = false
    return
  }

  savingEditPostId.value = post.id
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const updated = await updatePost(post.id, content)
    const index = posts.value.findIndex((item) => item.id === post.id)
    if (index >= 0) {
      posts.value[index] = {
        ...posts.value[index],
        content: updated.content,
        renderedContent: updated.renderedContent,
        updatedAt: updated.updatedAt,
      }
    }
    editingPostId.value = 0
    postContent.value = postDraftBeforeEdit.value
    targetPostId.value = targetPostBeforeEdit.value
    postDraftBeforeEdit.value = ''
    targetPostBeforeEdit.value = 0
    composerOpen.value = false
    pushFlash(t('topic.replyUpdated'), 'success')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : t('api.replyUpdateFailed')
  } finally {
    savingEditPostId.value = 0
  }
}

async function submitPost() {
  if (editingPostId.value) {
    await savePostEdit()
    return
  }

  const postId = targetPost.value?.id || 0
  const content = postContent.value.trim()
  if (submitting.value) return

  if (!content) {
    errorMessage.value = t('topic.replyRequired')
    successMessage.value = ''
    return
  }

  submitting.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const createdPost = await createPost(page.props.topic.id, content, postId)
    postContent.value = ''
    targetPostId.value = 0
    successMessage.value = t('topic.replyPosted')
    const createdPostId = typeof createdPost === 'object' && createdPost !== null ? createdPost.id : createdPost
    if (typeof createdPostId === 'number') {
      await revealCreatedPost(createdPostId)
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

function requestDeletePost(post: PostPayload) {
  if (savingEditPostId.value === post.id) return
  pendingDeletePost.value = post
  deleteErrorMessage.value = ''
}

function closeDeleteDialog() {
  if (deletingPostId.value) return
  pendingDeletePost.value = null
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
    actionMessage.value = action === 'ban' ? t('topic.moderationBanSuccess') : t('topic.moderationUnbanSuccess')
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

  const image = target.closest('.gf-prose-post img')
  if (!(image instanceof HTMLImageElement)) return

  const imageSrc = image.currentSrc || image.src
  if (!imageSrc) return

  const anchor = image.closest('a')
  if (anchor && !sameUrl(anchor.href, imageSrc)) return

  event.preventDefault()
  event.stopPropagation()

  const markdownImages = Array.from(document.querySelectorAll<HTMLImageElement>('.gf-prose-post img'))
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

function requestPostReport(post: PostPayload) {
  requestReport({
    targetType: 'post',
    targetId: post.id,
    title: t('topic.replyReportTitle', { no: post.postNo || post.id }),
    excerpt: post.content,
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
    pushFlash(t('topic.reportSubmitted'), 'success')
  } catch (error) {
    reportError.value = error instanceof Error ? error.message : t('api.reportFailed')
  } finally {
    reportSubmitting.value = false
  }
}

function postModerationBusy(postId: number) {
  return moderatingPostIds.value.includes(postId)
}

async function moderatePost(post: PostPayload, action: 'ban' | 'unban') {
  if (postModerationBusy(post.id)) return
  moderatingPostIds.value = [...moderatingPostIds.value, post.id]
  try {
    await updateModerationPostStatus(post.id, action)
    post.processStatus = action === 'ban' ? 1 : 0
    post.isHidden = action === 'ban'
    pushFlash(action === 'ban' ? t('topic.replyModerationBanSuccess') : t('topic.replyModerationUnbanSuccess'), 'success')
  } catch (error) {
    pushFlash(error instanceof Error ? error.message : t('api.moderationActionFailed'), 'error')
  } finally {
    moderatingPostIds.value = moderatingPostIds.value.filter(id => id !== post.id)
  }
}

async function removePost(postId: number) {
  if (deletingPostId.value || savingEditPostId.value === postId) return

  deletingPostId.value = postId
  errorMessage.value = ''
  successMessage.value = ''
  deleteErrorMessage.value = ''
  try {
    const removedPost = posts.value.find((post) => post.id === postId)
    await deletePost(postId)
    posts.value = posts.value.filter((post) => post.id !== postId)
    if (targetPostId.value === postId) {
      targetPostId.value = 0
    }
    if (editingPostId.value === postId) {
      editingPostId.value = 0
      postContent.value = postDraftBeforeEdit.value
      targetPostId.value = targetPostBeforeEdit.value
      postDraftBeforeEdit.value = ''
      targetPostBeforeEdit.value = 0
    }
    if (removedPost?.postNo && activePostNo.value === removedPost.postNo) {
      const closest = findClosestLoadedPost(removedPost.postNo)
      activePostNo.value = closest?.postNo || lastPostNo(posts.value) || firstPostNo(posts.value) || 1
    }
    syncLoadedPostWindowBounds()
    syncProgressForPostNo(activePostNo.value || 1)
    await nextTick()
    collectPostElements()
    scheduleActivePostFromScroll()
    successMessage.value = t('topic.replyDeleted')
    pendingDeletePost.value = null
  } catch (error) {
    deleteErrorMessage.value = error instanceof Error ? error.message : t('api.replyDeleteFailed')
  } finally {
    deletingPostId.value = 0
  }
}
</script>

<template>
  <div class="min-w-0">
    <div class="min-w-0" @click="handleMarkdownImageClick">
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
          <span class="inline-flex items-center gap-1.5">
            <MessageSquare class="h-3.5 w-3.5" />
            {{ formatNumber(page.props.topic.replyCount) }}
          </span>
          <span class="inline-flex items-center gap-1.5">
            <Eye class="h-3.5 w-3.5" />
            {{ formatNumber(page.props.topic.viewCount) }}
          </span>
          <span class="inline-flex items-center gap-1.5">
            <Heart class="h-3.5 w-3.5" />
            {{ formatNumber(likeCount) }}
          </span>
        </div>
      </header>

      <section class="gf-card xl:w-[calc(100%+292px)]">
        <div class="min-w-0 xl:grid xl:grid-cols-[minmax(0,1fr)_256px]">
          <div class="min-w-0">
            <span v-if="posts.length" id="posts" class="block scroll-mt-20" aria-hidden="true" />

            <div v-if="postHasBefore" class="relative px-4 py-3 text-center">
              <button
                v-if="postHasBefore"
                type="button"
                class="inline-flex h-8 items-center gap-1.5 rounded-md px-2 text-xs font-semibold text-primary transition-colors hover:bg-info/10 hover:text-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60"
                :disabled="loadingPostWindow"
                @click="loadPostWindow('before')"
              >
                <Loader2 v-if="loadingPostDirection === 'before'" class="h-3.5 w-3.5 animate-spin" />
                <ChevronsUp v-else class="h-3.5 w-3.5" />
                {{ t('topic.loadEarlierReplies') }}
              </button>
            </div>
            <article
              v-for="(post, index) in posts"
              :id="`post-${post.id}`"
              :key="post.id"
              :data-post-no="post.postNo"
              class="group relative grid scroll-mt-20 grid-cols-[40px_minmax(0,1fr)] gap-2.5 px-3 py-4 transition-[background-color] sm:grid-cols-[52px_minmax(0,1fr)] sm:gap-4 sm:p-5"
              :class="{
                'border-t border-line xl:border-t-transparent': index > 0,
                'bg-info/10': highlightedPostId === post.id,
                '[border-top-left-radius:calc(var(--gf-radius-box)-var(--gf-border))] [border-top-right-radius:calc(var(--gf-radius-box)-var(--gf-border))]': index === 0 && !postHasBefore,
              }"
            >
              <div v-if="index > 0" class="pointer-events-none absolute left-5 right-5 top-0 hidden border-t border-line xl:block" aria-hidden="true" />
              <a
                :href="`/u/${post.author.id}`"
                class="sticky top-19 self-start pt-1"
                @click="showUserCard(post.author, $event)"
              >
                <UserAvatar :src="post.author.avatarUrl" :alt="post.author.username" :badge="post.author.wornBadge" class="h-9 w-9 rounded-full ring-1 ring-line sm:h-10 sm:w-10" img-class="rounded-full" />
              </a>
              <div class="min-w-0">
                <div class="mb-1.5 flex min-w-0 items-start justify-between gap-2">
                  <div class="min-w-0">
                    <div class="flex min-w-0 items-center gap-2">
                      <a :href="`/u/${post.author.id}`" class="min-w-0 truncate font-semibold text-base-content hover:text-primary">{{ post.author.username }}</a>
                      <span v-if="isFirstPost(post)" class="rounded bg-base-200 px-1.5 py-0.5 text-xs font-semibold text-base-content/55">{{ t('topic.originalPost') }}</span>
                      <span v-if="post.postNo" class="hidden shrink-0 text-xs font-semibold tabular-nums text-base-content/55 sm:inline">#{{ formatNumber(post.postNo) }}</span>
                    </div>
                    <div class="mt-0.5 flex items-center gap-2 text-xs text-base-content/55 sm:hidden">
                      <span v-if="post.postNo" class="font-semibold tabular-nums text-base-content/55">#{{ formatNumber(post.postNo) }}</span>
                      <time class="truncate">{{ formatDateTime(post.createdAt) }}</time>
                    </div>
                  </div>
                  <div class="flex shrink-0 items-center gap-0.5 sm:gap-1.5">
                    <button
                      v-if="canEditPost(post)"
                      type="button"
                      class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-icon-muted transition hover:bg-info/10 hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                      :disabled="savingEditPostId === post.id || deletingPostId === post.id"
                      :title="t('common.edit')"
                      @click="startEditPost(post)"
                    >
                      <PencilLine class="h-3.5 w-3.5" />
                      <span class="sr-only">{{ t('common.edit') }}</span>
                    </button>
                    <button
                      v-if="canDeleteRenderedPost(post)"
                      type="button"
                      class="gf-icon-button h-8 w-8 shrink-0 hover:bg-error/10 hover:text-error focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-error focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                      :disabled="deletingPostId === post.id"
                      :title="deletingPostId === post.id ? t('topic.deleting') : t('topic.delete')"
                      @click="requestDeletePost(post)"
                    >
                      <Trash2 class="h-3.5 w-3.5" />
                      <span class="sr-only">{{ deletingPostId === post.id ? t('topic.deleting') : t('topic.delete') }}</span>
                    </button>
                    <button
                      v-if="page.props.permissions.canPost && !post.isHidden"
                      type="button"
                      class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-icon-muted transition hover:bg-info/10 hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2"
                      :title="t('topic.reply')"
                      @click="replyTo(post)"
                    >
                      <CornerDownLeft class="h-3.5 w-3.5" />
                      <span class="sr-only">{{ t('topic.reply') }}</span>
                    </button>
                    <button
                      v-if="!isFirstPost(post) && !post.isOwnPost && !post.isHidden"
                      type="button"
                      class="gf-icon-button h-8 w-8 shrink-0 hover:bg-warning/10 hover:text-warning focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-warning focus-visible:ring-offset-2"
                      :title="t('topic.report')"
                      @click="requestPostReport(post)"
                    >
                      <Flag class="h-3.5 w-3.5" />
                      <span class="sr-only">{{ t('topic.report') }}</span>
                    </button>
                    <button
                      v-if="!isFirstPost(post) && post.canModerate && post.processStatus === 0"
                      type="button"
                      class="gf-icon-button h-8 w-8 shrink-0 hover:bg-error/10 hover:text-error focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-error focus-visible:ring-offset-2 disabled:opacity-50"
                      :disabled="postModerationBusy(post.id)"
                      :title="t('topic.moderationBan')"
                      @click="moderatePost(post, 'ban')"
                    >
                      <Ban class="h-3.5 w-3.5" />
                      <span class="sr-only">{{ t('topic.moderationBan') }}</span>
                    </button>
                    <button
                      v-else-if="!isFirstPost(post) && post.canModerate && post.processStatus === 1"
                      type="button"
                      class="gf-icon-button h-8 w-8 shrink-0 hover:bg-info/10 hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 disabled:opacity-50"
                      :disabled="postModerationBusy(post.id)"
                      :title="t('topic.moderationUnban')"
                      @click="moderatePost(post, 'unban')"
                    >
                      <RotateCcw class="h-3.5 w-3.5" />
                      <span class="sr-only">{{ t('topic.moderationUnban') }}</span>
                    </button>
                    <time class="hidden w-36 shrink-0 text-right text-xs text-base-content/55 sm:-ml-1 sm:block">{{ formatDateTime(post.createdAt) }}</time>
                  </div>
                </div>
                <p v-if="post.replyToUsername" class="mb-1.5 inline-flex max-w-full min-w-0 items-center gap-1 rounded bg-base-200 px-2 py-1 text-sm text-base-content/55">
                  <span class="shrink-0">{{ t('topic.reply') }}</span>
                  <a :href="`/u/${post.replyToUserId}`" class="min-w-0 truncate font-medium text-base-content/75 hover:text-primary">@{{ post.replyToUsername }}</a>
                </p>
                <div v-if="post.isHidden && !post.canModerate" class="rounded border border-line bg-base-200/60 px-3 py-2 text-sm text-base-content/45">
                  {{ t('topic.hiddenReplyPlaceholder') }}
                </div>
                <div v-else class="gf-prose gf-prose-post" v-html="post.renderedContent" />
                <div v-if="post.isHidden && post.canModerate" class="mt-2 inline-flex rounded bg-base-200 px-2 py-1 text-xs font-semibold text-base-content/45">
                  {{ t('topic.hiddenReplyBadge') }}
                </div>
                <div v-if="post.updatedAt && post.updatedAt !== post.createdAt" class="mt-2 text-xs font-medium text-base-content/55">
                  {{ t('topic.editedAt', { time: formatDateTime(post.updatedAt) }) }}
                </div>
                <div v-if="isFirstPost(post)" class="mt-4 flex flex-wrap items-center gap-2 border-t border-line pt-3">
                  <button
                    type="button"
                    class="gf-button gf-button-sm px-2.5"
                    :class="isLiked ? 'bg-error/10 text-error hover:bg-error/10' : 'text-base-content/55 hover:bg-base-200 hover:text-base-content'"
                    :disabled="actingLike"
                    @click="toggleLike"
                  >
                    <Heart class="h-4 w-4" :fill="isLiked ? 'currentColor' : 'none'" />
                    {{ likeCount ? formatNumber(likeCount) : t('topic.like') }}
                  </button>
                  <button
                    type="button"
                    class="gf-button gf-button-sm px-2.5"
                    :class="isBookmarked ? 'bg-info/10 text-primary hover:bg-info/10' : 'text-base-content/55 hover:bg-base-200 hover:text-base-content'"
                    :disabled="actingBookmark"
                    @click="toggleBookmark"
                  >
                    <Bookmark class="h-4 w-4" :fill="isBookmarked ? 'currentColor' : 'none'" />
                    {{ isBookmarked ? t('topic.bookmarked') : t('topic.bookmark') }}
                  </button>
                  <button
                    type="button"
                    class="gf-button gf-button-sm px-2.5"
                    :class="isWatched ? 'bg-success/10 text-success hover:bg-success/15' : 'text-base-content/55 hover:bg-base-200 hover:text-base-content'"
                    :disabled="actingWatch"
                    @click="toggleWatch"
                  >
                    <Bell class="h-4 w-4" :fill="isWatched ? 'currentColor' : 'none'" />
                    {{ isWatched ? t('topic.watched') : t('topic.watch') }}
                  </button>
                  <button
                    v-if="!page.props.permissions.isOwnTopic"
                    type="button"
                    class="gf-button gf-button-sm px-2.5 text-base-content/55 hover:bg-warning/10 hover:text-warning"
                    @click="requestTopicReport"
                  >
                    <Flag class="h-4 w-4" />
                    {{ t('topic.report') }}
                  </button>
                  <button
                    v-if="page.props.permissions.canModerateTopic && topicProcessStatus === 0"
                    type="button"
                    class="gf-button gf-button-sm px-2.5 text-base-content/55 hover:bg-base-200 hover:text-base-content"
                    :disabled="actingModeration"
                    @click="requestTopicModeration('ban')"
                  >
                    <Ban class="h-4 w-4" />
                    {{ t('topic.moderationBan') }}
                  </button>
                  <button
                    v-else-if="page.props.permissions.canModerateTopic && topicProcessStatus === 1"
                    type="button"
                    class="gf-button gf-button-sm px-2.5 text-base-content/55 hover:bg-base-200 hover:text-base-content"
                    :disabled="actingModeration"
                    @click="requestTopicModeration('unban')"
                  >
                    <RotateCcw class="h-4 w-4" />
                    {{ t('topic.moderationUnban') }}
                  </button>
                  <span v-if="actionMessage" class="text-xs" :class="actionMessageSuccess ? 'text-base-content/75' : 'text-error'">{{ actionMessage }}</span>
                </div>
              </div>
            </article>

            <div v-if="postHasAfter || loadingPostDirection === 'after' || postWindowError || (!postHasAfter && posts.length)" ref="postLoadMoreEl" class="relative border-t border-line px-4 py-3 text-center xl:border-t-transparent">
              <div class="pointer-events-none absolute left-5 right-5 top-0 hidden border-t border-line xl:block" aria-hidden="true" />
              <button
                v-if="postHasAfter && postWindowError"
                type="button"
                class="gf-button gf-button-sm gf-button-secondary text-xs"
                :disabled="loadingPostWindow"
                @click="loadPostWindow('after')"
              >
                <Loader2 v-if="loadingPostDirection === 'after'" class="h-3.5 w-3.5 animate-spin" />
                {{ t('topic.retryLoadReplies') }}
              </button>
              <p v-else-if="postWindowError" class="text-xs text-error">{{ postWindowError }}</p>
              <p v-else-if="postHasAfter && loadingPostDirection === 'after'" class="inline-flex items-center justify-center gap-1.5 text-xs font-medium text-base-content/55">
                <Loader2 class="h-3.5 w-3.5 animate-spin" />
                {{ t('topic.loadingMoreReplies') }}
              </p>
              <button
                v-else-if="postHasAfter"
                type="button"
                class="gf-button gf-button-sm gf-button-secondary text-xs"
                :disabled="loadingPostWindow"
                @click="loadMoreRepliesManually"
              >
                {{ t('topic.loadMoreReplies') }}
              </button>
              <p v-else-if="!postHasAfter && posts.length" class="text-xs font-medium text-base-content/55">{{ t('topic.allRepliesShown') }}</p>
            </div>
            <span class="block h-px scroll-mb-28" aria-hidden="true" />
          </div>

          <aside class="hidden min-w-0 xl:block">
            <div
              class="sticky top-19"
            >
              <div class="px-4 py-4">
                <h2 class="text-sm font-semibold text-base-content/55">{{ t('topic.overview') }}</h2>
              </div>

              <dl class="space-y-4 border-t border-line px-4 py-5 text-sm">
                <div class="flex items-center justify-between gap-4">
                  <dt class="font-semibold text-base-content/55">{{ t('topic.replyCount') }}</dt>
                  <dd class="text-right font-semibold tabular-nums text-base-content">{{ formatNumber(page.props.topic.replyCount) }}</dd>
                </div>
                <div class="flex items-center justify-between gap-4">
                  <dt class="font-semibold text-base-content/55">{{ t('topic.viewCount') }}</dt>
                  <dd class="text-right font-semibold tabular-nums text-base-content">{{ formatNumber(page.props.topic.viewCount) }}</dd>
                </div>
                <div class="flex items-center justify-between gap-4">
                  <dt class="font-semibold text-base-content/55">{{ t('topic.participants') }}</dt>
                  <dd class="text-right font-semibold tabular-nums text-base-content">{{ page.props.topic.participants.length }}</dd>
                </div>
              </dl>

              <div v-if="page.props.topic.participants.length" class="border-t border-line px-4 py-4">
                <h3 class="mb-3 text-sm font-semibold text-base-content/55">{{ t('topic.activeParticipants') }}</h3>
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
                v-if="page.props.topic.replyCount > 0 && postMaxRange > 0"
                class="border-t border-line"
                :current="postRailCurrentNo"
                :max="postMaxRange"
                :start-label="postRailStartLabel"
                :end-label="postRailEndLabel"
                :current-label="postRailCurrentLabel"
                :busy="postRailBusy"
                :progress-current="postRailProgressCurrent"
                :progress-end="postRailProgressEnd"
                :progress-start="postRailProgressStart"
                @earliest="jumpToTopicBodyFromRail"
                @latest="jumpToLatestPostFromRail"
                @select="selectPostFromRail"
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

      <TopicFloatingControls
        v-model:mobile-rail-open="mobilePostRailOpen"
        :open="composerOpen"
        :actions="floatingTopicActions"
        :authenticated="page.layout.viewer.isAuthenticated"
        :can-post="page.props.permissions.canPost"
        :current-label="postRailCurrentLabel"
        :current-no="postRailCurrentNo"
        :end-label="postRailEndLabel"
        :has-rail="hasPostRail"
        :max-no="postMaxRange"
        :progress-current="postRailProgressCurrent"
        :progress-end="postRailProgressEnd"
        :progress-start="postRailProgressStart"
        :rail-busy="postRailBusy"
        :start-label="postRailStartLabel"
        @earliest="jumpToTopicBodyFromRail"
        @latest="jumpToLatestPostFromRail"
        @open-reply="openFloatingPostComposer"
        @select-rail="selectPostFromRail"
      />

      <PostComposer
        v-if="composerMounted"
        v-model="postContent"
        :open="composerOpen"
        :authenticated="page.layout.viewer.isAuthenticated"
        :error-message="errorMessage"
        :mode="composerMode"
        :submitting="editingPostId ? savingEditPostId > 0 : submitting"
        :success-message="successMessage"
        :target="targetPost"
        @clear-target="cancelPostTarget"
        @clear-validation="clearPostValidation"
        @image-error="handlePostImageError"
        @image-inserted="handlePostImageInserted"
        @submit="submitPost"
        @update:open="updateComposerOpen"
      />

    </div>

    <MarkdownImageViewer ref="markdownImageViewer" />

    <Teleport to="body">
      <Transition name="gf-modal">
        <div
          v-if="pendingDeletePost"
          class="fixed inset-0 z-[110] flex items-center justify-center bg-neutral/45 px-4 py-6 backdrop-blur-sm"
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-post-title"
          @click.self="closeDeleteDialog"
        >
          <div class="gf-menu-surface w-full max-w-sm p-4">
            <div class="flex items-start gap-3">
              <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-error/10 text-error">
                <AlertTriangle class="h-5 w-5" />
              </div>
              <div class="min-w-0 flex-1">
                <h2 id="delete-post-title" class="text-base font-bold text-base-content">{{ t('topic.deleteReplyTitle') }}</h2>
                <p class="mt-1 text-sm leading-6 text-base-content/55">{{ t('topic.deleteReplyDescription') }}</p>
              </div>
              <button
                type="button"
                class="rounded-md p-1 text-base-content/55 transition hover:bg-base-300 hover:text-base-content/75 disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="Boolean(deletingPostId)"
                @click="closeDeleteDialog"
              >
                <X class="h-4 w-4" />
              </button>
            </div>

            <div class="mt-4 border-l-2 border-error/35 pl-3">
              <div class="text-xs font-semibold text-base-content/45">@{{ pendingDeletePost.author.username }}</div>
              <p class="mt-1 line-clamp-3 whitespace-pre-wrap text-sm leading-6 text-base-content/70">{{ pendingDeletePost.content }}</p>
            </div>

            <p v-if="deleteErrorMessage" class="mt-3 text-sm text-error">{{ deleteErrorMessage }}</p>

            <div class="mt-4 flex justify-end gap-2">
              <button
                type="button"
                class="gf-button gf-button-md gf-button-muted"
                :disabled="Boolean(deletingPostId)"
                @click="closeDeleteDialog"
              >
                {{ t('common.cancel') }}
              </button>
              <button
                type="button"
                class="gf-button gf-button-md gf-button-danger"
                :disabled="Boolean(deletingPostId)"
                @click="removePost(pendingDeletePost.id)"
              >
                <Loader2 v-if="deletingPostId === pendingDeletePost.id" class="h-4 w-4 animate-spin" />
                <Trash2 v-else class="h-4 w-4" />
                {{ deletingPostId === pendingDeletePost.id ? t('topic.deleting') : t('topic.confirmDelete') }}
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
                <h2 id="report-title" class="text-base font-bold text-base-content">{{ t('topic.reportTitle') }}</h2>
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
                <span>{{ t(`topic.reportReasons.${reason}`) }}</span>
              </label>
              <textarea
                v-model="reportNote"
                class="gf-textarea min-h-24"
                maxlength="300"
                :placeholder="t('topic.reportNotePlaceholder')"
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
                {{ reportSubmitting ? t('common.loadingShort') : t('topic.submitReport') }}
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
                  {{ pendingModerationAction === 'ban' ? t('topic.moderationBanTitle') : t('topic.moderationUnbanTitle') }}
                </h2>
                <p class="mt-1 text-sm leading-6 text-base-content/55">
                  {{ pendingModerationAction === 'ban' ? t('topic.moderationBanDescription') : t('topic.moderationUnbanDescription') }}
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
                {{ actingModeration ? t('common.loadingShort') : (pendingModerationAction === 'ban' ? t('topic.confirmModerationBan') : t('topic.confirmModerationUnban')) }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
