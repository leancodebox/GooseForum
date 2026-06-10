<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
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
import UserAvatar from '@/site/components/UserAvatar.vue'
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
      <section class="overflow-hidden rounded-lg border border-line/70 bg-base-100 shadow-[0_2px_12px_rgba(0,0,0,0.04)]">
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
                  <span v-if="page.props.user.isAdmin" class="rounded bg-warning/10 px-1.5 py-0.5 text-[11px] font-semibold text-warning">Admin</span>
                  <span v-if="page.props.user.isOnline" class="inline-flex items-center gap-1 rounded bg-success/10 px-1.5 py-0.5 text-[11px] font-semibold text-success">
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
                class="inline-flex h-9 items-center gap-1.5 rounded-md border border-line bg-base-100 px-3 text-sm font-semibold text-base-content/75 hover:bg-base-200"
              >
                <Settings class="h-4 w-4" />
                {{ t('user.editProfile') }}
              </a>
              <a
                v-else-if="page.props.canMessage"
                :href="page.props.messageUrl"
                class="inline-flex h-9 items-center gap-1.5 rounded-md border border-line bg-base-100 px-3 text-sm font-semibold text-base-content/75 hover:bg-base-200"
              >
                <MessageSquare class="h-4 w-4" />
                {{ t('shell.nav.messages') }}
              </a>
              <button
                v-if="page.props.canFollow"
                type="button"
                class="inline-flex h-9 items-center gap-1.5 rounded-md px-3 text-sm font-semibold"
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

          <div class="mt-5 grid grid-cols-4 border-y border-line py-3 lg:grid-cols-7 lg:py-4">
            <div class="px-1 py-2 text-center lg:px-0 lg:py-0 lg:text-left">
              <div class="text-lg font-bold tabular-nums text-base-content lg:text-xl">{{ formatNumber(page.props.user.articleCount) }}</div>
              <div class="mt-0.5 text-[11px] font-medium text-base-content/55 lg:text-xs">{{ t('user.stats.topics') }}</div>
            </div>
            <div class="px-1 py-2 text-center lg:px-0 lg:py-0 lg:text-left">
              <div class="text-lg font-bold tabular-nums text-base-content lg:text-xl">{{ formatNumber(page.props.user.replyCount) }}</div>
              <div class="mt-0.5 text-[11px] font-medium text-base-content/55 lg:text-xs">{{ t('user.stats.replies') }}</div>
            </div>
            <div class="px-1 py-2 text-center lg:px-0 lg:py-0 lg:text-left">
              <div class="text-lg font-bold tabular-nums text-base-content lg:text-xl">{{ formatNumber(page.props.user.likeReceivedCount) }}</div>
              <div class="mt-0.5 text-[11px] font-medium text-base-content/55 lg:text-xs">{{ t('user.stats.likesReceived') }}</div>
            </div>
            <div class="px-1 py-2 text-center lg:px-0 lg:py-0 lg:text-left">
              <div class="text-lg font-bold tabular-nums text-base-content lg:text-xl">{{ formatNumber(page.props.user.likeGivenCount) }}</div>
              <div class="mt-0.5 text-[11px] font-medium text-base-content/55 lg:text-xs">{{ t('user.stats.likesGiven') }}</div>
            </div>
            <div class="px-1 py-2 text-center lg:px-0 lg:py-0 lg:text-left">
              <div class="text-lg font-bold tabular-nums text-base-content lg:text-xl">{{ formatNumber(page.props.user.followerCount) }}</div>
              <div class="mt-0.5 text-[11px] font-medium text-base-content/55 lg:text-xs">{{ t('user.stats.followers') }}</div>
            </div>
            <div class="px-1 py-2 text-center lg:px-0 lg:py-0 lg:text-left">
              <div class="text-lg font-bold tabular-nums text-base-content lg:text-xl">{{ formatNumber(page.props.user.followingCount) }}</div>
              <div class="mt-0.5 text-[11px] font-medium text-base-content/55 lg:text-xs">{{ t('user.stats.following') }}</div>
            </div>
            <div class="px-1 py-2 text-center lg:px-0 lg:py-0 lg:text-left">
              <div class="text-lg font-bold tabular-nums text-base-content lg:text-xl">{{ formatNumber(page.props.user.collectionCount) }}</div>
              <div class="mt-0.5 text-[11px] font-medium text-base-content/55 lg:text-xs">{{ t('user.stats.bookmarks') }}</div>
            </div>
          </div>

          <div class="mt-4 flex flex-wrap items-center gap-x-5 gap-y-2 text-xs text-base-content/55">
            <span class="inline-flex items-center gap-1.5"><CalendarDays class="h-3.5 w-3.5" /> {{ t('user.joinedAt', { date: formatDate(page.props.user.createdAt) }) }}</span>
            <span v-if="page.props.user.lastActiveTime">{{ t('user.lastActive', { time: timeAgo(page.props.user.lastActiveTime) }) }}</span>
          </div>

          <div v-if="visibleBadges.length" class="mt-5 border-t border-line pt-4">
            <div class="flex flex-wrap gap-3">
              <div
                v-for="badge in visibleBadges"
                :key="badge.code"
                class="group flex w-16 flex-col items-center gap-1.5"
                :title="badge.description"
              >
                <span
                  class="flex h-12 w-12 items-center justify-center ring-1 ring-inset transition group-hover:-translate-y-0.5 group-hover:shadow-sm"
                  :class="badgeClass(badge.color, badge.level)"
                  style="clip-path: polygon(25% 5%, 75% 5%, 100% 50%, 75% 95%, 25% 95%, 0 50%)"
                >
                  <img :src="badgeIconURL(badge)" :alt="badge.name" class="h-6 w-6 object-contain" />
                </span>
                <span class="w-full truncate text-center text-[11px] font-semibold text-base-content/75">{{ badge.name }}</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="mt-3 overflow-hidden rounded-lg border border-line/70 bg-base-100">
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
                  class="inline-flex h-5 items-center gap-1 rounded-full bg-base-300 px-2 text-[11px] text-base-content/55"
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
          <div v-if="!visibleTopics.length" class="px-4 py-14 text-center text-sm text-base-content/55">{{ t('user.emptyTopics') }}</div>
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
            <div v-if="!page.props.activities.length" class="py-14 text-center text-sm text-base-content/55">{{ t('user.emptyActivity') }}</div>
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
          <div v-if="(activeTab === 'following' ? page.props.following : page.props.followers).length === 0" class="py-14 text-center text-sm text-base-content/55">
            {{ t('user.emptyData') }}
          </div>
        </div>
      </section>
    </article>
</template>
