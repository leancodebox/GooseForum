<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { CalendarDays, FileText, Heart, MessageCircle, MessageSquare, PenLine, Radio, Settings, UserPlus, UsersRound } from '@lucide/vue'
import { followUser } from '@/runtime/api'
import { formatDate, formatDateTime, formatNumber, timeAgo } from '@/runtime/format'
import type { LayoutPayload, TopicPayload, UserActivityPayload, UserConnectionPayload, UserProfileProps } from '@/types/payload'

const page = defineProps<{
  layout: LayoutPayload
  props: UserProfileProps
}>()

const activeTab = ref<'topics' | 'activity' | 'following' | 'followers'>('topics')
const isFollowing = ref(page.props.user.isFollowing)
const followLoading = ref(false)
const followError = ref('')

const displayName = computed(() => page.props.user.nickname || page.props.user.username)
const bioText = computed(() => page.props.user.bio || page.props.user.signature || '这个用户还没有留下简介。')
const visibleTopics = computed(() => page.props.topics)
const tabItems = computed(() => [
  { key: 'topics', label: '主题', count: page.props.topics.length },
  { key: 'activity', label: '动态', count: page.props.activities.length },
  { key: 'following', label: '关注', count: page.props.user.followingCount },
  { key: 'followers', label: '粉丝', count: page.props.user.followerCount },
] as const)

watch(
  () => page.props.user.userId,
  () => {
    activeTab.value = 'topics'
    isFollowing.value = page.props.user.isFollowing
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
    followError.value = error instanceof Error ? error.message : '关注操作失败'
  } finally {
    followLoading.value = false
  }
}

function activityText(activity: UserActivityPayload) {
  return activity.contentPreview ? `${activity.label}：${activity.contentPreview}` : activity.label
}

function topicCategories(topic: TopicPayload) {
  return topic.categories.slice(0, 2)
}
</script>

<template>
    <article class="pb-12">
      <section class="overflow-hidden rounded-lg border border-gray-200/70 bg-white shadow-[0_2px_12px_rgba(0,0,0,0.04)]">
        <div class="h-24 border-b border-gray-100 bg-[linear-gradient(135deg,#f8fafc_0%,#eff6ff_48%,#f8fafc_100%)]" />
        <div class="px-4 pb-4 sm:px-5">
          <div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
            <div class="flex min-w-0 gap-4">
              <img
                :src="page.props.user.avatarUrl"
                :alt="page.props.user.username"
                class="-mt-10 h-20 w-20 rounded-lg border-4 border-white bg-white object-cover shadow-sm sm:h-24 sm:w-24"
              />
              <div class="min-w-0 pt-3">
                <div class="flex min-w-0 flex-wrap items-center gap-2">
                  <h1 class="truncate text-2xl font-bold leading-tight text-gray-950">{{ displayName }}</h1>
                  <span v-if="page.props.user.isAdmin" class="rounded bg-amber-50 px-1.5 py-0.5 text-[11px] font-semibold text-amber-700">Admin</span>
                  <span v-if="page.props.user.isOnline" class="inline-flex items-center gap-1 rounded bg-emerald-50 px-1.5 py-0.5 text-[11px] font-semibold text-emerald-700">
                    <Radio class="h-3 w-3" /> 在线
                  </span>
                </div>
                <p class="mt-1 text-sm font-medium text-gray-400">@{{ page.props.user.username }}</p>
                <p class="mt-2 max-w-3xl text-sm leading-relaxed text-gray-600">{{ bioText }}</p>
              </div>
            </div>

            <div class="flex shrink-0 flex-wrap items-center gap-2">
              <a
                v-if="page.props.isOwnProfile"
                :href="page.props.settingsUrl"
                class="inline-flex h-9 items-center gap-1.5 rounded-md border border-gray-200 bg-white px-3 text-sm font-semibold text-gray-700 hover:bg-gray-50"
              >
                <Settings class="h-4 w-4" />
                编辑资料
              </a>
              <a
                v-else-if="page.props.canMessage"
                :href="page.props.messageUrl"
                class="inline-flex h-9 items-center gap-1.5 rounded-md border border-gray-200 bg-white px-3 text-sm font-semibold text-gray-700 hover:bg-gray-50"
              >
                <MessageSquare class="h-4 w-4" />
                私信
              </a>
              <button
                v-if="page.props.canFollow"
                type="button"
                class="inline-flex h-9 items-center gap-1.5 rounded-md px-3 text-sm font-semibold"
                :class="isFollowing ? 'bg-gray-100 text-gray-800 hover:bg-gray-200' : 'bg-blue-600 text-white hover:bg-blue-700'"
                :disabled="followLoading"
                @click="toggleFollow"
              >
                <UserPlus class="h-4 w-4" />
                {{ followLoading ? '处理中...' : isFollowing ? '已关注' : '关注' }}
              </button>
            </div>
          </div>

          <p v-if="followError" class="mt-3 text-sm text-red-600">{{ followError }}</p>

          <div class="mt-5 grid grid-cols-2 gap-y-4 border-y border-gray-100 py-4 sm:grid-cols-4 lg:grid-cols-7">
            <div>
              <div class="text-xl font-bold tabular-nums text-gray-950">{{ formatNumber(page.props.user.articleCount) }}</div>
              <div class="mt-0.5 text-xs font-medium text-gray-400">主题</div>
            </div>
            <div>
              <div class="text-xl font-bold tabular-nums text-gray-950">{{ formatNumber(page.props.user.replyCount) }}</div>
              <div class="mt-0.5 text-xs font-medium text-gray-400">回复</div>
            </div>
            <div>
              <div class="text-xl font-bold tabular-nums text-gray-950">{{ formatNumber(page.props.user.likeReceivedCount) }}</div>
              <div class="mt-0.5 text-xs font-medium text-gray-400">获赞</div>
            </div>
            <div>
              <div class="text-xl font-bold tabular-nums text-gray-950">{{ formatNumber(page.props.user.likeGivenCount) }}</div>
              <div class="mt-0.5 text-xs font-medium text-gray-400">点赞</div>
            </div>
            <div>
              <div class="text-xl font-bold tabular-nums text-gray-950">{{ formatNumber(page.props.user.followerCount) }}</div>
              <div class="mt-0.5 text-xs font-medium text-gray-400">粉丝</div>
            </div>
            <div>
              <div class="text-xl font-bold tabular-nums text-gray-950">{{ formatNumber(page.props.user.followingCount) }}</div>
              <div class="mt-0.5 text-xs font-medium text-gray-400">关注</div>
            </div>
            <div>
              <div class="text-xl font-bold tabular-nums text-gray-950">{{ formatNumber(page.props.user.collectionCount) }}</div>
              <div class="mt-0.5 text-xs font-medium text-gray-400">收藏</div>
            </div>
          </div>

          <div class="mt-4 flex flex-wrap items-center gap-x-5 gap-y-2 text-xs text-gray-400">
            <span class="inline-flex items-center gap-1.5"><CalendarDays class="h-3.5 w-3.5" /> 加入于 {{ formatDate(page.props.user.createdAt) }}</span>
            <span v-if="page.props.user.lastActiveTime">最后活跃 {{ timeAgo(page.props.user.lastActiveTime) }}</span>
          </div>
        </div>
      </section>

      <section class="mt-3 overflow-hidden rounded-lg border border-gray-200/70 bg-white">
        <div class="flex overflow-x-auto border-b border-gray-100 px-3">
          <button
            v-for="tab in tabItems"
            :key="tab.key"
            type="button"
            class="inline-flex h-11 shrink-0 items-center gap-2 border-b-2 px-3 text-sm font-semibold"
            :class="activeTab === tab.key ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-900'"
            @click="activeTab = tab.key"
          >
            {{ tab.label }}
            <span class="rounded bg-gray-100 px-1.5 py-0.5 text-[11px] text-gray-500">{{ formatNumber(tab.count) }}</span>
          </button>
        </div>

        <div v-if="activeTab === 'topics'" class="divide-y divide-gray-100">
          <a
            v-for="topic in visibleTopics"
            :key="topic.id"
            :href="topic.url"
            class="group grid gap-2 px-4 py-3 hover:bg-gray-50 sm:grid-cols-[minmax(0,1fr)_80px_80px_104px] sm:items-center"
          >
            <div class="min-w-0">
              <div class="flex min-w-0 flex-wrap items-center gap-2">
                <span class="truncate text-[15px] font-semibold text-gray-950 group-hover:text-blue-600">{{ topic.title }}</span>
                <span
                  v-for="category in topicCategories(topic)"
                  :key="category.id"
                  class="inline-flex h-5 items-center gap-1 rounded-full bg-gray-100 px-2 text-[11px] text-gray-500"
                >
                  <span class="h-1.5 w-1.5 rounded-full" :style="{ backgroundColor: category.color }" />
                  {{ category.name }}
                </span>
              </div>
              <p v-if="topic.description" class="mt-1 truncate text-sm text-gray-500">{{ topic.description }}</p>
            </div>
            <span class="hidden text-center text-sm font-semibold tabular-nums text-gray-700 sm:block">{{ formatNumber(topic.replyCount) }}</span>
            <span class="hidden text-center text-sm tabular-nums text-gray-500 sm:block">{{ formatNumber(topic.viewCount) }}</span>
            <span class="hidden text-right text-xs font-medium text-gray-400 sm:block">{{ timeAgo(topic.lastUpdateTime) }}</span>
          </a>
          <div v-if="!visibleTopics.length" class="px-4 py-14 text-center text-sm text-gray-500">还没有发布主题。</div>
        </div>

        <div v-else-if="activeTab === 'activity'" class="p-4">
          <div class="space-y-3">
            <a
              v-for="activity in page.props.activities"
              :key="activity.id"
              :href="activity.url || '#'"
              class="flex gap-3 rounded-md border border-gray-100 p-3 hover:border-blue-100 hover:bg-blue-50/30"
            >
              <span class="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-gray-100 text-gray-500">
                <PenLine v-if="activity.action === 2" class="h-4 w-4" />
                <Heart v-else-if="activity.action === 3" class="h-4 w-4" />
                <UserPlus v-else-if="activity.action === 4" class="h-4 w-4" />
                <MessageCircle v-else-if="activity.action === 5" class="h-4 w-4" />
                <FileText v-else class="h-4 w-4" />
              </span>
              <span class="min-w-0">
                <span class="block truncate text-sm font-semibold text-gray-900">{{ activityText(activity) }}</span>
                <time class="mt-1 block text-xs text-gray-400">{{ formatDateTime(activity.createdAt) }}</time>
              </span>
            </a>
            <div v-if="!page.props.activities.length" class="py-14 text-center text-sm text-gray-500">暂无动态。</div>
          </div>
        </div>

        <div v-else class="p-4">
          <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
            <a
              v-for="item in activeTab === 'following' ? page.props.following : page.props.followers"
              :key="item.id"
              :href="item.url"
              class="flex min-w-0 gap-3 rounded-md border border-gray-100 p-3 hover:border-blue-100 hover:bg-blue-50/30"
            >
              <img :src="item.avatarUrl" :alt="item.username" class="h-10 w-10 rounded-full object-cover" />
              <span class="min-w-0">
                <span class="block truncate text-sm font-semibold text-gray-900">{{ item.nickname || item.username }}</span>
                <span class="block truncate text-xs text-gray-400">@{{ item.username }}</span>
                <span class="mt-1 block truncate text-xs text-gray-500">{{ item.bio || '暂无简介' }}</span>
              </span>
            </a>
          </div>
          <div v-if="(activeTab === 'following' ? page.props.following : page.props.followers).length === 0" class="py-14 text-center text-sm text-gray-500">
            暂无数据。
          </div>
        </div>
      </section>
    </article>
</template>
