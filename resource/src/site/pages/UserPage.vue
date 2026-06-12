<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  Bird,
  CalendarDays,
  FileText,
  Heart,
  MessageCircle,
  MessageSquare,
  PenLine,
  Radio,
  Settings,
  UserPlus,
} from '@lucide/vue'
import { followUser } from '@/runtime/api'
import { formatDate, formatDateTime, formatNumber, timeAgo } from '@/runtime/format'
import { topicDescription } from '@/runtime/topic-description'
import EmptyState from '@/site/components/EmptyState.vue'
import UserAvatar from '@/site/components/UserAvatar.vue'
import { socialIcons, socialLabels } from '@/site/utils/social-icons'
import type { LayoutPayload, TopicPayload, UserActivityPayload, UserConnectionPayload, UserProfileProps } from '@/types/payload'
import { useI18n } from 'vue-i18n'

const page = defineProps<{
  layout: LayoutPayload
  props: UserProfileProps
}>()

const { t } = useI18n()
const activeTab = ref<'topics' | 'activity' | 'following' | 'followers'>('topics')
const isFollowing = ref(page.props.user.isFollowing)
const followLoading = ref(false)
const followError = ref('')
const coverUrl = ref(page.props.user.profileCoverUrl || '')

const displayName = computed(() => page.props.user.nickname || page.props.user.username)
const bioText = computed(() => page.props.user.bio || page.props.user.signature || t('user.emptyBio'))
const visibleTopics = computed(() => page.props.topics)
const visibleBadges = computed(() => page.props.badges.slice(0, 8))
const socialKeys = ['github', 'twitter', 'linkedIn', 'weibo', 'bilibili', 'zhihu'] as const
const tabItems = computed(() => [
  { key: 'topics', label: t('user.tabs.topics'), count: page.props.topics.length },
  { key: 'activity', label: t('user.tabs.activity'), count: page.props.activities.length },
  { key: 'following', label: t('user.tabs.following'), count: page.props.user.followingCount },
  { key: 'followers', label: t('user.tabs.followers'), count: page.props.user.followerCount },
] as const)
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
  () => page.props.user.userId,
  () => {
    activeTab.value = 'topics'
    isFollowing.value = page.props.user.isFollowing
    coverUrl.value = page.props.user.profileCoverUrl || ''
    followError.value = ''
  },
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

function topicCategories(topic: TopicPayload) {
  return topic.categories.slice(0, 2)
}

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

function badgeClass(color: string, level: string) {
  if (color === 'blue') return 'bg-blue-100 text-blue-700 ring-blue-200'
  if (color === 'emerald') return 'bg-emerald-100 text-emerald-700 ring-emerald-200'
  if (color === 'teal') return 'bg-teal-100 text-teal-700 ring-teal-200'
  if (color === 'sky') return 'bg-sky-100 text-sky-700 ring-sky-200'
  if (color === 'cyan') return 'bg-cyan-100 text-cyan-700 ring-cyan-200'
  if (color === 'rose') return 'bg-rose-100 text-rose-700 ring-rose-200'
  if (color === 'violet') return 'bg-violet-100 text-violet-700 ring-violet-200'
  if (color === 'purple') return 'bg-purple-100 text-purple-700 ring-purple-200'
  if (color === 'fuchsia') return 'bg-fuchsia-100 text-fuchsia-700 ring-fuchsia-200'
  if (color === 'indigo') return 'bg-indigo-100 text-indigo-700 ring-indigo-200'
  if (color === 'amber') return 'bg-amber-100 text-amber-700 ring-amber-200'
  if (color === 'orange') return 'bg-orange-100 text-orange-700 ring-orange-200'
  if (color === 'yellow') return 'bg-yellow-100 text-yellow-700 ring-yellow-200'
  if (color === 'slate') return 'bg-slate-100 text-slate-700 ring-slate-200'
  if (level === 'gold') return 'bg-amber-100 text-amber-700 ring-amber-200'
  if (level === 'special') return 'bg-indigo-100 text-indigo-700 ring-indigo-200'
  return 'bg-blue-100 text-blue-700 ring-blue-200'
}

function badgeIconURL(badge: UserProfileProps['badges'][number]) {
  return badge.iconUrl || '/static/badges/contributor.svg'
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
                size="large"
                class="-mt-9 h-24 w-24 rounded-lg border-4 border-base-100 bg-base-100 object-cover shadow-sm sm:-mt-10 sm:h-28 sm:w-28"
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

          <div class="mt-5 grid grid-cols-4 border-b border-t border-line/70 py-2.5 lg:grid-cols-8">
            <div v-for="item in profileStats" :key="item.label" class="px-1 py-2 text-center lg:px-0 lg:py-0">
              <div class="text-base font-bold tabular-nums lg:text-lg" :class="item.featured ? 'text-primary' : 'text-base-content'">{{ formatNumber(item.value) }}</div>
              <div class="mt-0.5 text-[11px] font-medium lg:text-xs" :class="item.featured ? 'text-primary/80' : 'text-base-content/55'">{{ item.label }}</div>
            </div>
          </div>

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

          <div v-if="visibleBadges.length" class="mt-3 border-t border-line/60 pt-3">
            <div class="flex flex-wrap gap-x-2.5 gap-y-2">
              <div
                v-for="badge in visibleBadges"
                :key="badge.code"
                class="group flex w-14 flex-col items-center gap-1"
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
          </div>
        </div>
      </section>

      <section class="gf-panel mt-3 overflow-hidden">
        <div class="flex overflow-x-auto border-b border-line px-3">
          <button
            v-for="tab in tabItems"
            :key="tab.key"
            type="button"
            class="inline-flex h-11 shrink-0 items-center gap-2 border-b-2 px-3 text-sm font-semibold"
            :class="activeTab === tab.key ? 'border-primary text-primary' : 'border-transparent text-base-content/55 hover:text-base-content'"
            @click="activeTab = tab.key"
          >
            {{ tab.label }}
            <span class="rounded bg-base-300 px-1.5 py-0.5 text-[11px] text-base-content/55">{{ formatNumber(tab.count) }}</span>
          </button>
        </div>

        <div v-if="activeTab === 'topics'" class="divide-y divide-line">
          <a
            v-for="topic in visibleTopics"
            :key="topic.id"
            :href="topic.url"
            class="group grid gap-2 px-4 py-3 hover:bg-base-200 sm:grid-cols-[minmax(0,1fr)_80px_80px_104px] sm:items-center"
          >
            <div class="min-w-0">
              <div class="flex min-w-0 flex-wrap items-center gap-2">
                <span class="truncate text-[15px] font-semibold text-base-content group-hover:text-primary">{{ topic.title }}</span>
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
            <span class="hidden text-center text-sm tabular-nums text-base-content/55 sm:block">{{ formatNumber(topic.viewCount) }}</span>
            <span class="hidden text-right text-xs font-medium text-base-content/55 sm:block">{{ timeAgo(topic.lastUpdateTime) }}</span>
          </a>
          <EmptyState v-if="!visibleTopics.length" :icon="FileText" :title="t('user.emptyTopics')" />
        </div>

        <div v-else-if="activeTab === 'activity'" class="p-4">
          <div class="space-y-3">
            <a
              v-for="activity in page.props.activities"
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
            <EmptyState v-if="!page.props.activities.length" :icon="MessageCircle" :title="t('user.emptyActivity')" />
          </div>
        </div>

        <div v-else class="p-4">
          <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
            <a
              v-for="item in activeTab === 'following' ? page.props.following : page.props.followers"
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
          <EmptyState v-if="(activeTab === 'following' ? page.props.following : page.props.followers).length === 0" :icon="UserPlus" :title="t('user.emptyData')" />
        </div>
      </section>
    </article>
</template>
