<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { CalendarDays, Loader2, Radio, UserPlus } from '@lucide/vue'
import { getUserHoverCard } from '../runtime/api'
import { formatDate, formatNumber, timeAgo } from '../runtime/format'
import type { UserCardShowDetail } from '../runtime/user-card-events'
import type { UserHoverCardPayload } from '../types/payload'

const visible = ref(false)
const loading = ref(false)
const error = ref('')
const fallbackUser = ref<UserCardShowDetail['user'] | null>(null)
const card = ref<UserHoverCardPayload | null>(null)
const position = ref({ left: 0, top: 0 })
const cache = new Map<number, UserHoverCardPayload>()
let hideTimer: number | undefined
let requestToken = 0

const displayName = computed(() => card.value?.nickname || fallbackUser.value?.username || card.value?.username || '')
const username = computed(() => card.value?.username || fallbackUser.value?.username || '')
const avatarUrl = computed(() => card.value?.avatarUrl || fallbackUser.value?.avatarUrl || '')
const profileUrl = computed(() => `/u/${card.value?.userId || fallbackUser.value?.id || 0}`)
const bioText = computed(() => card.value?.bio || card.value?.signature || '这个用户还没有留下简介。')

function clearHideTimer() {
  if (hideTimer) {
    window.clearTimeout(hideTimer)
    hideTimer = undefined
  }
}

function hideNow() {
  clearHideTimer()
  visible.value = false
}

function scheduleHide() {
  clearHideTimer()
  hideTimer = window.setTimeout(() => {
    visible.value = false
  }, 130)
}

function placeCard(target: HTMLElement) {
  const rect = target.getBoundingClientRect()
  const cardWidth = Math.min(320, window.innerWidth - 24)
  const estimatedCardHeight = 252
  const gap = 10
  const viewportPadding = 12
  const viewportWidth = window.innerWidth

  let left = rect.left
  left = Math.max(viewportPadding, Math.min(left, viewportWidth - cardWidth - viewportPadding))
  const belowTop = rect.bottom + gap
  const aboveTop = rect.top - estimatedCardHeight - gap
  const top = belowTop + estimatedCardHeight > window.innerHeight - viewportPadding && aboveTop > viewportPadding ? aboveTop : belowTop

  position.value = {
    left,
    top: Math.max(viewportPadding, Math.min(top, window.innerHeight - estimatedCardHeight - viewportPadding)),
  }
}

async function show(event: Event) {
  const detail = (event as CustomEvent<UserCardShowDetail>).detail
  if (!detail?.user?.id || !detail.target) return

  clearHideTimer()
  fallbackUser.value = detail.user
  visible.value = true
  error.value = ''
  placeCard(detail.target)

  const cached = cache.get(detail.user.id)
  if (cached) {
    card.value = cached
    loading.value = false
    return
  }

  const token = ++requestToken
  loading.value = true
  card.value = null
  try {
    const result = await getUserHoverCard(detail.user.id)
    if (token !== requestToken) return
    cache.set(detail.user.id, result)
    card.value = result
  } catch {
    if (token !== requestToken) return
    error.value = '用户资料暂时不可用'
  } finally {
    if (token === requestToken) loading.value = false
  }
}

onMounted(() => {
  window.addEventListener('goose:user-card-show', show)
  window.addEventListener('goose:user-card-hide', scheduleHide)
  window.addEventListener('scroll', hideNow, { passive: true })
  window.addEventListener('resize', hideNow)
  window.addEventListener('goose:page', hideNow)
})

onBeforeUnmount(() => {
  window.removeEventListener('goose:user-card-show', show)
  window.removeEventListener('goose:user-card-hide', scheduleHide)
  window.removeEventListener('scroll', hideNow)
  window.removeEventListener('resize', hideNow)
  window.removeEventListener('goose:page', hideNow)
  clearHideTimer()
})
</script>

<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="fixed z-[90] w-[min(20rem,calc(100vw-1.5rem))] rounded-lg border border-gray-200 bg-white p-3 text-gray-900 shadow-[0_18px_50px_-24px_rgba(15,23,42,0.45),0_8px_24px_-16px_rgba(15,23,42,0.25)]"
      :style="{ left: `${position.left}px`, top: `${position.top}px` }"
      @mouseenter="clearHideTimer"
      @mouseleave="scheduleHide"
    >
      <div class="flex items-start gap-3">
        <a :href="profileUrl" class="shrink-0 rounded-full ring-2 ring-white">
          <img :src="avatarUrl" :alt="username" class="h-14 w-14 rounded-full object-cover ring-1 ring-gray-100" />
        </a>
        <div class="min-w-0 flex-1">
          <div class="flex min-w-0 items-center gap-2">
            <a :href="profileUrl" class="truncate text-base font-bold text-gray-950 hover:text-blue-600">{{ displayName }}</a>
            <span v-if="card?.isAdmin" class="shrink-0 rounded bg-amber-50 px-1.5 py-0.5 text-[11px] font-semibold text-amber-700">Admin</span>
          </div>
          <div class="mt-0.5 flex items-center gap-2 text-xs text-gray-400">
            <span class="truncate">@{{ username }}</span>
            <span v-if="card?.isOnline" class="inline-flex items-center gap-1 text-emerald-600">
              <Radio class="h-3 w-3" />
              在线
            </span>
            <span v-else-if="card?.lastActiveTime">活跃于 {{ timeAgo(card.lastActiveTime) }}</span>
          </div>
        </div>
      </div>

      <div v-if="loading" class="mt-3 min-h-[164px]">
        <div class="space-y-2">
          <div class="h-4 w-full rounded bg-gray-100" />
          <div class="h-4 w-3/4 rounded bg-gray-100" />
        </div>
        <div class="mt-3 grid grid-cols-4 divide-x divide-gray-100 border-y border-gray-100 py-2">
          <div v-for="item in 4" :key="item" class="px-2 text-center">
            <div class="mx-auto h-4 w-7 rounded bg-gray-100" />
            <div class="mx-auto mt-1 h-3 w-8 rounded bg-gray-100" />
          </div>
        </div>
        <div class="mt-3 flex items-center justify-between gap-3">
          <div class="flex items-center gap-1.5 text-xs text-gray-400">
            <Loader2 class="h-3.5 w-3.5 animate-spin" />
            加载用户资料
          </div>
          <div class="h-8 w-24 rounded-md bg-gray-100" />
        </div>
      </div>
      <div v-else-if="error" class="mt-3 flex min-h-[164px] items-center rounded-md bg-red-50 px-3 py-2 text-sm text-red-600">{{ error }}</div>
      <template v-else>
        <p class="mt-3 h-10 line-clamp-2 text-sm leading-relaxed text-gray-600">{{ bioText }}</p>

        <div class="mt-3 grid grid-cols-4 divide-x divide-gray-100 border-y border-gray-100 py-2">
          <div class="px-2 text-center">
            <div class="text-sm font-bold tabular-nums text-gray-950">{{ formatNumber(card?.articleCount || 0) }}</div>
            <div class="mt-0.5 text-[11px] text-gray-400">主题</div>
          </div>
          <div class="px-2 text-center">
            <div class="text-sm font-bold tabular-nums text-gray-950">{{ formatNumber(card?.replyCount || 0) }}</div>
            <div class="mt-0.5 text-[11px] text-gray-400">回复</div>
          </div>
          <div class="px-2 text-center">
            <div class="text-sm font-bold tabular-nums text-gray-950">{{ formatNumber(card?.likeReceivedCount || 0) }}</div>
            <div class="mt-0.5 text-[11px] text-gray-400">获赞</div>
          </div>
          <div class="px-2 text-center">
            <div class="text-sm font-bold tabular-nums text-gray-950">{{ formatNumber(card?.followerCount || 0) }}</div>
            <div class="mt-0.5 text-[11px] text-gray-400">关注者</div>
          </div>
        </div>

        <div class="mt-3 flex items-center justify-between gap-3">
          <div class="inline-flex items-center gap-1.5 text-xs text-gray-400">
            <CalendarDays class="h-3.5 w-3.5" />
            加入于 {{ card?.createdAt ? formatDate(card.createdAt) : '-' }}
          </div>
          <a
            :href="profileUrl"
            class="inline-flex h-8 items-center gap-1.5 rounded-md bg-gray-900 px-3 text-sm font-semibold text-white hover:bg-gray-800"
          >
            <UserPlus class="h-4 w-4" />
            {{ card?.isFollowing ? '已关注' : '查看主页' }}
          </a>
        </div>
      </template>
    </div>
  </Teleport>
</template>
