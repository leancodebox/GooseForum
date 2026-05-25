<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import {
  Bird,
  CalendarDays,
  ExternalLink,
  Loader2,
  Radio,
  UserPlus,
} from '@lucide/vue'
import { getUserHoverCard } from '@/runtime/api'
import { formatDate, formatNumber, timeAgo } from '@/runtime/format'
import type { UserCardShowDetail } from '@/runtime/user-card-events'
import type { UserHoverCardPayload } from '@/types/payload'
import { socialIcons, socialLabels, type SimpleIcon } from '@/site/utils/social-icons'

const visible = ref(false)
const loading = ref(false)
const error = ref('')
const fallbackUser = ref<UserCardShowDetail['user'] | null>(null)
const card = ref<UserHoverCardPayload | null>(null)
const position = ref({ left: 0, top: 0 })
const cardEl = ref<HTMLElement | null>(null)
const activeBadgeCode = ref('')
const cache = new Map<number, UserHoverCardPayload>()
let requestToken = 0
let preferredSide: 'top' | 'bottom' | null = null

const displayName = computed(() => card.value?.nickname || fallbackUser.value?.username || card.value?.username || '')
const username = computed(() => card.value?.username || fallbackUser.value?.username || '')
const avatarUrl = computed(() => card.value?.avatarUrl || fallbackUser.value?.avatarUrl || '')
const profileUrl = computed(() => `/u/${card.value?.userId || fallbackUser.value?.id || 0}`)
const bioText = computed(() => card.value?.bio || card.value?.signature || '')
const externalLinks = computed(() => {
  const links: Array<{ key: string; label: string; url: string; icon?: SimpleIcon }> = []
  const primaryUrl = normalizeWebsiteURL(card.value?.website || '')
  if (primaryUrl) {
    links.push({
      key: 'website',
      label: card.value?.websiteName || formatLinkLabel(primaryUrl),
      url: primaryUrl,
    })
  }
  const externalInformation = card.value?.externalInformation || {}
  for (const [key, item] of Object.entries(externalInformation)) {
    const url = normalizeWebsiteURL(item?.link || '')
    if (!url) continue
    links.push({ key, label: socialLabels[key] || formatLinkLabel(url), url, icon: socialIcons[key] })
  }
  return links
})
const visibleBadges = computed(() => (card.value?.badges || []).slice(0, 5))

function normalizeWebsiteURL(value: string) {
  const url = value.trim()
  if (!url) return ''
  if (/^https?:\/\//i.test(url)) return url
  return `https://${url}`
}

function formatLinkLabel(url: string) {
  return url.replace(/^https?:\/\//i, '').replace(/^www\./i, '').replace(/\/$/, '')
}

function hideNow() {
  visible.value = false
  activeBadgeCode.value = ''
  preferredSide = null
}

function placeCard(target: HTMLElement) {
  const rect = target.getBoundingClientRect()
  const cardWidth = Math.min(320, window.innerWidth - 24)
  const measuredHeight = cardEl.value?.offsetHeight || 0
  const cardHeight = Math.max(measuredHeight, 220)
  const gap = 10
  const viewportPadding = 12
  const viewportWidth = window.innerWidth

  let left = rect.left
  left = Math.max(viewportPadding, Math.min(left, viewportWidth - cardWidth - viewportPadding))
  const belowTop = rect.bottom + gap
  const aboveTop = rect.top - cardHeight - gap
  if (!preferredSide) {
    const belowSpace = window.innerHeight - rect.bottom - gap - viewportPadding
    const aboveSpace = rect.top - gap - viewportPadding
    preferredSide = belowSpace >= cardHeight || belowSpace >= aboveSpace ? 'bottom' : 'top'
  }
  const top = preferredSide === 'top' ? aboveTop : belowTop

  position.value = {
    left,
    top: Math.max(viewportPadding, Math.min(top, window.innerHeight - cardHeight - viewportPadding)),
  }
}

async function show(event: Event) {
  const detail = (event as CustomEvent<UserCardShowDetail>).detail
  if (!detail?.user?.id || !detail.target) return

  fallbackUser.value = detail.user
  visible.value = true
  error.value = ''
  requestAnimationFrame(() => placeCard(detail.target))

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
    requestAnimationFrame(() => placeCard(detail.target))
  } catch {
    if (token !== requestToken) return
    error.value = '用户资料暂时不可用'
  } finally {
    if (token === requestToken) loading.value = false
  }
}

function onDocumentPointerDown(event: PointerEvent) {
  if (!visible.value) return
  const target = event.target
  if (target instanceof Node && cardEl.value?.contains(target)) return
  hideNow()
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') hideNow()
}

onMounted(() => {
  window.addEventListener('goose:user-card-show', show)
  document.addEventListener('pointerdown', onDocumentPointerDown)
  window.addEventListener('keydown', onKeydown)
  window.addEventListener('scroll', hideNow, { passive: true })
  window.addEventListener('resize', hideNow)
  window.addEventListener('goose:page', hideNow)
})

onBeforeUnmount(() => {
  window.removeEventListener('goose:user-card-show', show)
  document.removeEventListener('pointerdown', onDocumentPointerDown)
  window.removeEventListener('keydown', onKeydown)
  window.removeEventListener('scroll', hideNow)
  window.removeEventListener('resize', hideNow)
  window.removeEventListener('goose:page', hideNow)
})

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

function badgeTooltip(badge: UserHoverCardPayload['badges'][number]) {
  return badge.description ? `${badge.name}：${badge.description}` : badge.name
}

function badgeIconURL(badge: UserHoverCardPayload['badges'][number]) {
  return badge.iconUrl || '/static/badges/contributor.svg'
}
</script>

<template>
  <Teleport to="body">
    <Transition name="user-card-pop">
      <div
        v-if="visible"
        ref="cardEl"
        class="fixed z-[90] w-[min(20rem,calc(100vw-1.5rem))] rounded-lg border border-gray-200 bg-white p-3 text-gray-900 shadow-[0_18px_50px_-24px_rgba(15,23,42,0.45),0_8px_24px_-16px_rgba(15,23,42,0.25)]"
        :style="{ left: `${position.left}px`, top: `${position.top}px` }"
        @click.stop
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

        <Transition name="user-card-content" mode="out-in">
          <div v-if="loading" key="loading" class="mt-3 min-h-[164px]">
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
          <div v-else-if="error" key="error" class="mt-3 flex min-h-[164px] items-center rounded-md bg-red-50 px-3 py-2 text-sm text-red-600">{{ error }}</div>
          <div v-else key="content">
        <p v-if="bioText" class="mt-3 line-clamp-2 text-sm leading-relaxed text-gray-600">{{ bioText }}</p>

        <div v-if="visibleBadges.length" class="mt-3 flex gap-2">
          <span
            v-for="badge in visibleBadges"
            :key="badge.code"
            class="group relative flex h-8 w-8 shrink-0 items-center justify-center"
            tabindex="0"
            @mouseenter="activeBadgeCode = badge.code"
            @mouseleave="activeBadgeCode = ''"
            @focus="activeBadgeCode = badge.code"
            @blur="activeBadgeCode = ''"
          >
            <span
              class="flex h-8 w-8 items-center justify-center ring-1 ring-inset transition duration-150"
              :class="[badgeClass(badge.color, badge.level), activeBadgeCode === badge.code ? '-translate-y-0.5 scale-110 shadow-md' : 'shadow-none']"
              style="clip-path: polygon(25% 5%, 75% 5%, 100% 50%, 75% 95%, 25% 95%, 0 50%)"
            >
              <img :src="badgeIconURL(badge)" :alt="badge.name" class="h-4 w-4 object-contain" />
            </span>
            <span
              v-if="activeBadgeCode === badge.code"
              class="pointer-events-none absolute left-1/2 top-full z-10 mt-2 w-max max-w-48 -translate-x-1/2 rounded-md bg-gray-950 px-2 py-1 text-xs font-medium leading-5 text-white shadow-lg"
            >
              {{ badgeTooltip(badge) }}
            </span>
          </span>
        </div>

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

        <div v-if="externalLinks.length" class="mt-3 flex items-center gap-2 border-b border-gray-100 pb-3">
          <a
            v-for="link in externalLinks.slice(0, 8)"
            :key="`${link.key}-${link.url}`"
            :href="link.url"
            target="_blank"
            rel="noopener noreferrer"
            class="group relative inline-flex h-7 w-7 items-center justify-center rounded-md text-gray-400 transition hover:bg-gray-50 hover:text-blue-600"
            :title="link.label"
            :aria-label="link.label"
          >
            <Bird v-if="link.key === 'website'" class="h-4 w-4" />
            <svg
              v-else-if="link.icon"
              class="h-4 w-4"
              viewBox="0 0 24 24"
              fill="currentColor"
              aria-hidden="true"
            >
              <path :d="link.icon.path" />
            </svg>
            <ExternalLink v-else class="h-4 w-4" />
            <span class="pointer-events-none absolute bottom-full left-1/2 z-10 mb-2 max-w-40 -translate-x-1/2 truncate rounded-md bg-gray-950 px-2 py-1 text-xs font-medium text-white opacity-0 shadow-lg transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100">
              {{ link.label }}
            </span>
          </a>
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
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.user-card-pop-enter-active,
.user-card-pop-leave-active {
  transition:
    opacity 140ms ease,
    transform 140ms ease;
  transform-origin: top left;
}

.user-card-pop-enter-from,
.user-card-pop-leave-to {
  opacity: 0;
  transform: translateY(4px) scale(0.98);
}

.user-card-content-enter-active,
.user-card-content-leave-active {
  transition:
    opacity 120ms ease,
    transform 120ms ease;
}

.user-card-content-enter-from,
.user-card-content-leave-to {
  opacity: 0;
  transform: translateY(3px);
}

@media (prefers-reduced-motion: reduce) {
  .user-card-pop-enter-active,
  .user-card-pop-leave-active,
  .user-card-content-enter-active,
  .user-card-content-leave-active {
    transition: none;
  }
}
</style>
