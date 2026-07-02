<script setup lang="ts">
import { computed, nextTick, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref, watch } from 'vue'
import {
  Award,
  Bird,
  CalendarDays,
  FileText,
  Heart,
  List,
  MessageCircle,
  MessageSquare,
  PenLine,
  Radio,
  Settings,
  UserRound,
  UserPlus,
} from '@lucide/vue'
import { followUser } from '@/runtime/api'
import { formatDate, formatDateTime, formatNumber, timeAgo } from '@/runtime/format'
import { fetchPage } from '@/runtime/router'
import { topicDescription } from '@/runtime/topic-description'
import EmptyState from '@/site/components/EmptyState.vue'
import TopicList from '@/site/components/TopicList.vue'
import TopicListFooter from '@/site/components/TopicListFooter.vue'
import UserAvatar from '@/site/components/UserAvatar.vue'
import { badgeClass, badgeIconURL } from '@/site/utils/badge-style'
import { socialIcons, socialLabels } from '@/site/utils/social-icons'
import type { LayoutPayload, PagePayload, TopicPayload, UserActivityPayload, UserLikePayload, UserProfileProps } from '@/types/payload'
import { useI18n } from 'vue-i18n'

const page = defineProps<{
  layout: LayoutPayload
  props: UserProfileProps
}>()

const { t } = useI18n()
const activityListMode = 'waterfall'
const isFollowing = ref(page.props.user.isFollowing)
const followLoading = ref(false)
const followError = ref('')
const coverUrl = ref(page.props.user.profileCoverUrl || '')
const activityTopics = ref<TopicPayload[]>([])
const activities = ref<UserActivityPayload[]>([])
const likes = ref<UserLikePayload[]>([])
const pagination = ref(page.props.pagination)
const loadingMore = ref(false)
const loadError = ref('')
const loadMoreSentinel = ref<HTMLElement | null>(null)
let observer: IntersectionObserver | undefined

const displayName = computed(() => page.props.user.nickname || page.props.user.username)
const bioText = computed(() => page.props.user.bio || page.props.user.signature || t('user.emptyBio'))
const visibleTopics = computed(() => page.props.topics)
const visibleBadges = computed(() => page.props.badges.slice(0, 8))
const activeConnections = computed(() => page.props.activityTab === 'following' ? page.props.following : page.props.followers)
const isWaterfallTab = computed(() => page.props.section === 'activity' && (page.props.activityTab === 'timeline' || page.props.activityTab === 'topics' || page.props.activityTab === 'likes'))
const hasActivityTopics = computed(() => activityTopics.value.length > 0)
const hasActivities = computed(() => activities.value.length > 0)
const hasLikes = computed(() => likes.value.length > 0)
const socialKeys = ['github', 'twitter', 'linkedIn', 'weibo', 'bilibili', 'zhihu'] as const
const tabItems = computed(() => [
  ...page.props.tabs.map(tab => ({ ...tab, label: userTabLabel(tab.key) })),
])
const activityTabItems = computed(() => [
  ...page.props.activityTabs.map(tab => ({ ...tab, label: userActivityTabLabel(tab.key) })),
])
const profileCoverStyle = computed(() => {
  const activeCoverUrl = coverUrl.value.trim()
  const defaultCover = 'linear-gradient(135deg, var(--gf-color-base-200) 0%, var(--gf-color-info-content) 52%, var(--gf-color-base-200) 100%)'
  if (!activeCoverUrl) {
    return {
      backgroundImage: defaultCover,
    }
  }
  return {
    backgroundImage: `url(${JSON.stringify(activeCoverUrl)}), ${defaultCover}`,
  }
})
const profileStats = computed(() => [
  { label: t('user.stats.reputation'), value: page.props.user.prestige, featured: true },
  { label: t('user.stats.topics'), value: page.props.user.articleCount },
  { label: t('user.stats.replies'), value: page.props.user.replyCount },
  { label: t('user.stats.likesReceived'), value: page.props.user.likeReceivedCount },
  { label: t('user.stats.likesGiven'), value: page.props.user.likeGivenCount },
  { label: t('user.stats.followers'), value: page.props.user.followerCount },
  { label: t('user.stats.following'), value: page.props.user.followingCount },
  { label: t('user.stats.bookmarks'), value: page.props.user.collectionCount },
])
const websiteUrl = computed(() => safeProfileUrl(page.props.user.website))
const socialProfileLinks = computed(() => socialKeys
  .map((key) => {
    const href = safeProfileUrl(page.props.user.externalInformation?.[key]?.link)
    return href
      ? {
          key,
          href,
          label: socialLabels[key],
          icon: socialIcons[key],
        }
      : null
  })
  .filter((item): item is NonNullable<typeof item> => Boolean(item)))

watch(
  () => [page.props.user.userId, page.props.section, page.props.activityTab, page.props.pagination.nextUrl],
  () => {
    isFollowing.value = page.props.user.isFollowing
    coverUrl.value = page.props.user.profileCoverUrl || ''
    followError.value = ''
    activityTopics.value = [...page.props.topics]
    activities.value = [...page.props.activities]
    likes.value = [...page.props.likes]
    pagination.value = page.props.pagination
    loadError.value = ''
    void nextTick(observeSentinel)
  },
  { immediate: true },
)

async function toggleFollow() {
  if (!page.props.canFollow || followLoading.value) return

  followLoading.value = true
  followError.value = ''
  try {
    await followUser(page.props.user.userId, isFollowing.value)
    isFollowing.value = !isFollowing.value
  } catch (error) {
    followError.value = error instanceof Error ? error.message : t('api.followFailed')
  } finally {
    followLoading.value = false
  }
}

function activityText(activity: UserActivityPayload) {
  const label = activityLabel(activity)
  return activity.contentPreview ? `${label}: ${activity.contentPreview}` : label
}

function activityLabel(activity: UserActivityPayload) {
  if (activity.action === 1 || activity.label === 'signup') return t('user.activity.signup')
  if (activity.action === 2 || activity.label === 'post') return t('user.activity.post')
  if (activity.action === 3 || activity.label === 'like') return t('user.activity.like')
  if (activity.action === 4 || activity.label === 'follow') return t('user.activity.follow')
  if (activity.action === 5 || activity.label === 'comment') return t('user.activity.comment')
  return activity.label || t('user.activity.default')
}

function userTabLabel(key: string) {
  if (key === 'summary') return t('user.tabs.summary')
  if (key === 'activity') return t('user.tabs.activity')
  if (key === 'badges') return t('user.tabs.badges')
  return key
}

function userActivityTabLabel(key: string) {
  if (key === 'timeline') return t('user.tabs.timeline')
  if (key === 'topics') return t('user.tabs.topics')
  if (key === 'likes') return t('user.tabs.likes')
  if (key === 'following') return t('user.tabs.following')
  if (key === 'followers') return t('user.tabs.followers')
  return key
}

function topicCategories(topic: TopicPayload) {
  return topic.categories.slice(0, 2)
}

async function loadMore() {
  if (!isWaterfallTab.value || loadingMore.value || !pagination.value.hasNext || !pagination.value.nextUrl) return

  loadingMore.value = true
  loadError.value = ''
  try {
    const payload = (await fetchPage(new URL(pagination.value.nextUrl, window.location.origin))) as PagePayload<UserProfileProps>
    if (page.props.activityTab === 'topics') {
      activityTopics.value = mergeTopics(activityTopics.value, payload.props.topics)
    } else if (page.props.activityTab === 'likes') {
      likes.value = mergeLikes(likes.value, payload.props.likes)
    } else {
      activities.value = mergeActivities(activities.value, payload.props.activities)
    }
    pagination.value = payload.props.pagination
  } catch (error) {
    loadError.value = error instanceof Error ? error.message : t('common.loadFailed')
  } finally {
    loadingMore.value = false
  }
}

function mergeTopics(current: TopicPayload[], incoming: TopicPayload[]) {
  const seen = new Set(current.map((topic) => topic.id))
  return [...current, ...incoming.filter((topic) => !seen.has(topic.id))]
}

function mergeActivities(current: UserActivityPayload[], incoming: UserActivityPayload[]) {
  const seen = new Set(current.map((activity) => activity.id))
  return [...current, ...incoming.filter((activity) => !seen.has(activity.id))]
}

function mergeLikes(current: UserLikePayload[], incoming: UserLikePayload[]) {
  const seen = new Set(current.map((like) => like.id))
  return [...current, ...incoming.filter((like) => !seen.has(like.id))]
}

function observeSentinel() {
  observer?.disconnect()
  if (!isWaterfallTab.value || !loadMoreSentinel.value || !('IntersectionObserver' in window)) return
  observer = new IntersectionObserver(
    (entries) => {
      if (entries.some((entry) => entry.isIntersecting)) void loadMore()
    },
    { rootMargin: '480px 0px' },
  )
  observer.observe(loadMoreSentinel.value)
}

onMounted(observeSentinel)
onActivated(() => {
  void nextTick(observeSentinel)
})
onDeactivated(() => {
  observer?.disconnect()
})
onBeforeUnmount(() => {
  observer?.disconnect()
})

function safeProfileUrl(value?: string) {
  const rawValue = value?.trim()
  if (!rawValue) return ''

  try {
    const parsed = new URL(rawValue)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:' ? parsed.toString() : ''
  } catch {
    return ''
  }
}

</script>

<template>
    <article class="pb-12">
      <section class="gf-card overflow-hidden">
        <div class="h-20 border-b border-line bg-base-300 bg-cover bg-center sm:h-24" :style="profileCoverStyle" />
        <div class="px-4 pb-4 sm:px-5">
          <div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
            <div class="flex min-w-0 gap-4">
              <UserAvatar
                :src="page.props.user.avatarUrl"
                :alt="page.props.user.username"
                :badge="page.props.user.wornBadge"
                size="large"
                img-class="rounded-full"
                class="-mt-9 h-24 w-24 rounded-full border-2 border-base-100 bg-base-100 shadow-sm sm:-mt-10 sm:h-28 sm:w-28"
              />
              <div class="min-w-0 pt-3">
                <div class="flex min-w-0 flex-wrap items-center gap-2">
                  <h1 class="truncate text-2xl font-bold leading-tight text-base-content">{{ displayName }}</h1>
                  <span v-if="page.props.user.isAdmin" class="gf-badge gf-badge-warning rounded text-[11px]">Admin</span>
                  <span v-if="page.props.user.isOnline" class="gf-badge gf-badge-success rounded text-[11px]">
                    <Radio class="h-3 w-3" /> {{ t('user.online') }}
                  </span>
                </div>
                <p class="mt-1 text-sm font-medium text-base-content/55">@{{ page.props.user.username }}</p>
                <p class="mt-2 max-w-3xl text-sm leading-relaxed text-base-content/75">{{ bioText }}</p>
              </div>
            </div>

            <div class="flex shrink-0 flex-wrap items-center gap-2">
              <a
                v-if="page.props.isOwnProfile"
                :href="page.props.settingsUrl"
                class="gf-button gf-button-md gf-button-secondary"
              >
                <Settings class="h-4 w-4" />
                {{ t('user.editProfile') }}
              </a>
              <a
                v-else-if="page.props.canMessage"
                :href="page.props.messageUrl"
                class="gf-button gf-button-md gf-button-secondary"
              >
                <MessageSquare class="h-4 w-4" />
                {{ t('shell.nav.messages') }}
              </a>
              <button
                v-if="page.props.canFollow"
                type="button"
                class="gf-button gf-button-md"
                :class="isFollowing ? 'bg-base-300 text-base-content hover:bg-base-300' : 'bg-primary text-primary-content hover:bg-primary'"
                :disabled="followLoading"
                @click="toggleFollow"
              >
                <UserPlus class="h-4 w-4" />
                {{ followLoading ? t('common.loading') : isFollowing ? t('user.following') : t('user.follow') }}
              </button>
            </div>
          </div>

          <p v-if="followError" class="mt-3 text-sm text-error">{{ followError }}</p>

          <div class="mt-3 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex flex-wrap items-center gap-x-5 gap-y-2 text-xs text-base-content/55">
              <span class="inline-flex items-center gap-1.5"><CalendarDays class="h-3.5 w-3.5" /> {{ t('user.joinedAt', { date: formatDate(page.props.user.createdAt) }) }}</span>
              <span v-if="page.props.user.lastActiveTime">{{ t('user.lastActive', { time: timeAgo(page.props.user.lastActiveTime) }) }}</span>
            </div>

            <div v-if="websiteUrl || socialProfileLinks.length" class="flex flex-wrap items-center gap-0.5 sm:justify-end">
              <a
                v-if="websiteUrl"
                :href="websiteUrl"
                target="_blank"
                rel="noopener noreferrer ugc"
                class="group relative inline-flex h-8 w-8 items-center justify-center rounded-md text-icon-muted transition hover:bg-base-200 hover:text-primary"
                :title="page.props.user.websiteName || page.props.user.website"
                :aria-label="page.props.user.websiteName || page.props.user.website"
              >
                <Bird class="h-5 w-5" />
                <span class="gf-tooltip pointer-events-none absolute bottom-full left-1/2 z-10 mb-2 max-w-40 -translate-x-1/2 truncate opacity-0 transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100">
                  {{ page.props.user.websiteName || page.props.user.website }}
                </span>
              </a>
              <a
                v-for="item in socialProfileLinks"
                :key="item.key"
                :href="item.href"
                target="_blank"
                rel="noopener noreferrer ugc"
                class="group relative inline-flex h-8 w-8 items-center justify-center rounded-md text-icon-muted transition hover:bg-base-200 hover:text-primary"
                :title="item.label"
                :aria-label="item.label"
              >
                <svg class="h-4 w-4 fill-current" role="img" viewBox="0 0 24 24" aria-hidden="true">
                  <path :d="item.icon.path" />
                </svg>
                <span class="gf-tooltip pointer-events-none absolute bottom-full left-1/2 z-10 mb-2 max-w-40 -translate-x-1/2 truncate opacity-0 transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100">
                  {{ item.label }}
                </span>
              </a>
            </div>
          </div>

        </div>

        <div class="grid grid-cols-3 border-y border-line">
          <a
            v-for="tab in tabItems"
            :key="tab.key"
            :href="tab.url"
            class="inline-flex h-11 min-w-0 items-center justify-center gap-2 px-2 text-sm font-semibold"
            :class="tab.active ? 'text-primary shadow-[inset_0_-2px_0_var(--gf-color-primary)]' : 'text-base-content/55 hover:text-base-content'"
          >
            <UserRound v-if="tab.key === 'summary'" class="h-4 w-4 shrink-0" />
            <List v-else-if="tab.key === 'activity'" class="h-4 w-4 shrink-0" />
            <Award v-else class="h-4 w-4 shrink-0" />
            {{ tab.label }}
          </a>
        </div>

        <div v-if="page.props.section === 'summary'" class="p-4">
          <section class="grid grid-cols-4 gap-y-4 border-b border-line pb-4 lg:grid-cols-8">
            <div v-for="item in profileStats" :key="item.label" class="min-w-0 text-center">
              <div class="text-base font-bold tabular-nums lg:text-lg" :class="item.featured ? 'text-primary' : 'text-base-content'">{{ formatNumber(item.value) }}</div>
              <div class="mt-0.5 truncate text-[11px] font-medium lg:text-xs" :class="item.featured ? 'text-primary/80' : 'text-base-content/55'">{{ item.label }}</div>
            </div>
          </section>

          <div class="grid gap-5 pt-4 lg:grid-cols-[minmax(0,1fr)_300px]">
            <section class="min-w-0">
              <h2 class="mb-2 text-sm font-semibold text-base-content/75">{{ t('user.summarySections.recentTopics') }}</h2>
              <div class="divide-y divide-line">
                <a
                  v-for="topic in visibleTopics"
                  :key="topic.id"
                  :href="topic.url"
                  class="grid gap-2 py-3 sm:grid-cols-[minmax(0,1fr)_72px_88px] sm:items-center"
                >
                  <div class="min-w-0">
                    <div class="flex min-w-0 flex-wrap items-center gap-2">
                      <span class="truncate text-[15px] font-semibold text-base-content">{{ topic.title }}</span>
                      <span
                        v-for="category in topicCategories(topic)"
                        :key="category.id"
                        class="gf-badge gf-badge-muted h-5 gap-1 text-[11px] font-normal"
                      >
                        <span class="h-1.5 w-1.5 rounded-full" :style="{ backgroundColor: category.color }" />
                        {{ category.name }}
                      </span>
                    </div>
                    <p class="mt-1 truncate text-sm text-base-content/55">{{ topicDescription(topic) }}</p>
                  </div>
                  <span class="hidden text-center text-sm font-semibold tabular-nums text-base-content/75 sm:block">{{ formatNumber(topic.replyCount) }}</span>
                  <span class="hidden text-right text-xs font-medium text-base-content/55 sm:block">{{ timeAgo(topic.lastUpdateTime) }}</span>
                </a>
              </div>
              <EmptyState v-if="!visibleTopics.length" :icon="FileText" :title="t('user.emptyTopics')" />
            </section>

            <aside class="min-w-0 space-y-5">
              <section v-if="visibleBadges.length" class="border-b border-line pb-5 lg:border-b-0">
                <h2 class="mb-3 text-sm font-semibold text-base-content/75">{{ t('user.summarySections.recentBadges') }}</h2>
                <div class="flex flex-wrap gap-x-3 gap-y-2">
                  <div
                    v-for="badge in visibleBadges"
                    :key="badge.code"
                    class="group flex w-16 flex-col items-center gap-1"
                    :title="badge.description"
                  >
                    <span
                      class="flex h-10 w-10 items-center justify-center ring-1 ring-inset transition group-hover:-translate-y-0.5 group-hover:shadow-sm"
                      :class="badgeClass(badge.color, badge.level)"
                      style="clip-path: polygon(25% 5%, 75% 5%, 100% 50%, 75% 95%, 25% 95%, 0 50%)"
                    >
                      <img :src="badgeIconURL(badge)" :alt="badge.name" class="h-5 w-5 object-contain" />
                    </span>
                    <span class="w-full truncate text-center text-[10px] font-semibold text-base-content/75">{{ badge.name }}</span>
                  </div>
                </div>
              </section>

              <section v-if="page.props.activities.length">
                <h2 class="mb-2 text-sm font-semibold text-base-content/75">{{ t('user.summarySections.recentActivity') }}</h2>
                <div class="divide-y divide-line">
                  <a
                    v-for="activity in page.props.activities"
                    :key="activity.id"
                    :href="activity.url || '#'"
                    class="flex gap-3 py-3"
                  >
                    <span class="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-base-300 text-base-content/55">
                      <PenLine v-if="activity.action === 2" class="h-4 w-4" />
                      <Heart v-else-if="activity.action === 3" class="h-4 w-4" />
                      <UserPlus v-else-if="activity.action === 4" class="h-4 w-4" />
                      <MessageCircle v-else-if="activity.action === 5" class="h-4 w-4" />
                      <FileText v-else class="h-4 w-4" />
                    </span>
                    <span class="min-w-0">
                      <span class="block truncate text-sm font-semibold text-base-content">{{ activityText(activity) }}</span>
                      <time class="mt-1 block text-xs text-base-content/55">{{ formatDateTime(activity.createdAt) }}</time>
                    </span>
                  </a>
                </div>
              </section>
            </aside>
          </div>
        </div>

        <div v-else-if="page.props.section === 'activity'">
          <div class="grid grid-cols-5 border-b border-line">
            <a
              v-for="tab in activityTabItems"
              :key="tab.key"
              :href="tab.url"
              class="inline-flex h-10 min-w-0 items-center justify-center gap-2 px-2 text-sm font-medium"
              :class="tab.active ? 'text-primary shadow-[inset_0_-2px_0_var(--gf-color-primary)]' : 'text-base-content/55 hover:text-base-content'"
            >
              <List v-if="tab.key === 'timeline'" class="h-4 w-4 shrink-0" />
              <FileText v-else-if="tab.key === 'topics'" class="h-4 w-4 shrink-0" />
              <Heart v-else-if="tab.key === 'likes'" class="h-4 w-4 shrink-0" />
              <UserPlus v-else-if="tab.key === 'following'" class="h-4 w-4 shrink-0" />
              <UserRound v-else class="h-4 w-4 shrink-0" />
              {{ tab.label }}
            </a>
          </div>

          <div v-if="page.props.activityTab === 'timeline'">
            <div class="space-y-3 p-4">
              <a
                v-for="activity in activities"
                :key="activity.id"
                :href="activity.url || '#'"
                class="flex gap-3 rounded-md border border-line p-3 hover:border-primary/20 hover:bg-info/10"
              >
                <span class="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-base-300 text-base-content/55">
                  <PenLine v-if="activity.action === 2" class="h-4 w-4" />
                  <Heart v-else-if="activity.action === 3" class="h-4 w-4" />
                  <UserPlus v-else-if="activity.action === 4" class="h-4 w-4" />
                  <MessageCircle v-else-if="activity.action === 5" class="h-4 w-4" />
                  <FileText v-else class="h-4 w-4" />
                </span>
                <span class="min-w-0">
                  <span class="block truncate text-sm font-semibold text-base-content">{{ activityText(activity) }}</span>
                  <time class="mt-1 block text-xs text-base-content/55">{{ formatDateTime(activity.createdAt) }}</time>
                </span>
              </a>
              <EmptyState v-if="!hasActivities" :icon="MessageCircle" :title="t('user.emptyActivity')" />
            </div>
            <div v-if="pagination.hasNext || hasActivities" ref="loadMoreSentinel">
              <TopicListFooter
                :pagination="pagination"
                :mode="activityListMode"
                :loading-more="loadingMore"
                :has-topics="hasActivities"
                :load-error="loadError"
                @load-more="loadMore"
              />
            </div>
          </div>

          <div v-else-if="page.props.activityTab === 'topics'">
            <TopicList :topics="activityTopics">
              <template #empty>
                <EmptyState v-if="!hasActivityTopics" :icon="FileText" :title="t('user.emptyTopics')" />
              </template>
            </TopicList>
            <div v-if="pagination.hasNext || hasActivityTopics" ref="loadMoreSentinel">
              <TopicListFooter
                :pagination="pagination"
                :mode="activityListMode"
                :loading-more="loadingMore"
                :has-topics="hasActivityTopics"
                :load-error="loadError"
                @load-more="loadMore"
              />
            </div>
          </div>

          <div v-else-if="page.props.activityTab === 'likes'">
            <div class="space-y-3 p-4">
              <a
                v-for="like in likes"
                :key="like.id"
                :href="like.url"
                class="flex min-w-0 gap-3 rounded-md border border-line p-3 hover:border-primary/20 hover:bg-info/10"
              >
                <span class="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-base-300 text-base-content/55">
                  <Heart class="h-4 w-4" />
                </span>
                <span class="min-w-0">
                  <span class="block text-xs font-medium text-base-content/55">{{ t('user.activity.like') }}</span>
                  <span class="mt-0.5 block truncate text-sm font-semibold text-base-content">{{ like.title }}</span>
                  <time class="mt-1 block text-xs text-base-content/55">{{ formatDateTime(like.likedAt) }}</time>
                </span>
              </a>
              <EmptyState v-if="!hasLikes" :icon="Heart" :title="t('user.emptyData')" />
            </div>
            <div v-if="pagination.hasNext || hasLikes" ref="loadMoreSentinel">
              <TopicListFooter
                :pagination="pagination"
                :mode="activityListMode"
                :loading-more="loadingMore"
                :has-topics="hasLikes"
                :load-error="loadError"
                @load-more="loadMore"
              />
            </div>
          </div>

          <div v-else class="grid gap-3 p-4 sm:grid-cols-2 xl:grid-cols-3">
            <a
              v-for="item in activeConnections"
              :key="item.id"
              :href="item.url"
              class="flex min-w-0 gap-3 rounded-md border border-line p-3 hover:border-primary/20 hover:bg-info/10"
            >
              <UserAvatar :src="item.avatarUrl" :alt="item.username" class="h-10 w-10 rounded-full object-cover" />
              <span class="min-w-0">
                <span class="block truncate text-sm font-semibold text-base-content">{{ item.nickname || item.username }}</span>
                <span class="block truncate text-xs text-base-content/55">@{{ item.username }}</span>
                <span class="mt-1 block truncate text-xs text-base-content/55">{{ item.bio || t('user.noBio') }}</span>
              </span>
            </a>
          </div>
          <EmptyState v-if="(page.props.activityTab === 'following' || page.props.activityTab === 'followers') && activeConnections.length === 0" :icon="UserPlus" :title="t('user.emptyData')" />
        </div>

        <div v-else class="p-4">
          <div class="grid gap-3 sm:grid-cols-3 lg:grid-cols-4">
            <div
              v-for="badge in page.props.badges"
              :key="badge.code"
              class="flex min-w-0 items-center gap-3 rounded-md border border-line p-3"
              :title="badge.description"
            >
              <span
                class="flex h-11 w-11 shrink-0 items-center justify-center ring-1 ring-inset"
                :class="badgeClass(badge.color, badge.level)"
                style="clip-path: polygon(25% 5%, 75% 5%, 100% 50%, 75% 95%, 25% 95%, 0 50%)"
              >
                <img :src="badgeIconURL(badge)" :alt="badge.name" class="h-5 w-5 object-contain" />
              </span>
              <span class="min-w-0">
                <span class="block truncate text-sm font-semibold text-base-content">{{ badge.name }}</span>
                <span class="mt-0.5 block truncate text-xs text-base-content/55">{{ badge.description }}</span>
              </span>
            </div>
          </div>
          <EmptyState v-if="!page.props.badges.length" :icon="FileText" :title="t('user.emptyData')" />
        </div>
      </section>
    </article>
</template>
