<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Bird,
  CalendarDays,
  ExternalLink,
  Loader2,
  Radio,
  UserPlus,
} from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { PopoverContent } from '@/components/ui/popover'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { getUserCard } from '@/runtime/api'
import { formatDate, formatNumber, timeAgo } from '@/runtime/format'
import { userCardCache, type UserCardSide, type UserCardTarget } from '@/runtime/user-card'
import type { UserCardPayload } from '@gooseforum/client'
import { socialIcons, socialLabels, type SimpleIcon } from '@/site/utils/social-icons'
import { badgeClass, badgeIconURL, badgeTooltip } from '@/site/utils/badge-style'
import UserAvatar from './UserAvatar.vue'

const { t } = useI18n()
const props = defineProps<{
  user: UserCardTarget
  side: UserCardSide
}>()
const loading = ref(true)
const error = ref('')
const card = ref<UserCardPayload | null>(null)

const displayName = computed(() => card.value?.nickname || props.user.username || card.value?.username || '')
const username = computed(() => card.value?.username || props.user.username || '')
const avatarUrl = computed(() => card.value?.avatarUrl || props.user.avatarUrl || '')
const wornBadge = computed(() => card.value?.wornBadge || props.user.wornBadge || null)
const profileUrl = computed(() => `/u/${card.value?.userId || props.user.id}`)
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

async function loadCard() {
  error.value = ''
  const cached = userCardCache.get(props.user.id)
  if (cached) {
    card.value = cached
    loading.value = false
    return
  }

  loading.value = true
  card.value = null
  try {
    const result = await getUserCard(props.user.id)
    userCardCache.set(props.user.id, result)
    card.value = result
  } catch {
    error.value = t('userCard.unavailable')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadCard()
})

</script>

<template>
  <PopoverContent
    :side="side"
    :side-flip="false"
    :side-offset="10"
    :collision-padding="12"
    :aria-label="displayName"
    class="max-h-[var(--reka-popover-content-available-height)] w-[min(20rem,calc(100vw-1.5rem))] overflow-y-auto border-line bg-base-100 p-3 text-base-content"
  >
      <div class="flex items-start gap-3">
        <a :href="profileUrl" class="shrink-0 rounded-full ring-2 ring-base-100">
          <UserAvatar :src="avatarUrl" :alt="username" :badge="wornBadge" size="medium" class="h-14 w-14 rounded-full ring-1 ring-line" img-class="rounded-full" />
        </a>
        <div class="min-w-0 flex-1">
          <div class="flex min-w-0 items-center gap-2">
            <a :href="profileUrl" class="truncate text-base font-bold text-base-content hover:text-primary">{{ displayName }}</a>
            <Badge v-if="card?.isAdmin" variant="warning" class="shrink-0 rounded text-[11px]">Admin</Badge>
          </div>
          <div class="mt-0.5 flex items-center gap-2 text-xs text-base-content/55">
            <span class="truncate">@{{ username }}</span>
            <span v-if="card?.isOnline" class="inline-flex items-center gap-1 text-success">
              <Radio class="h-3 w-3" />
              {{ t('userCard.online') }}
            </span>
            <span v-else-if="card?.lastActiveTime">{{ t('userCard.activeAt', { time: timeAgo(card.lastActiveTime) }) }}</span>
          </div>
        </div>
      </div>

        <Transition name="user-card-content" mode="out-in">
          <div v-if="loading" key="loading" class="mt-3 min-h-[164px]">
            <div class="space-y-2">
              <div class="h-4 w-full rounded bg-base-300" />
              <div class="h-4 w-3/4 rounded bg-base-300" />
            </div>
            <div class="mt-3 grid grid-cols-4 divide-x divide-line border-y border-line py-2">
              <div v-for="item in 4" :key="item" class="px-2 text-center">
                <div class="mx-auto h-4 w-7 rounded bg-base-300" />
                <div class="mx-auto mt-1 h-3 w-8 rounded bg-base-300" />
              </div>
            </div>
            <div class="mt-3 flex items-center justify-between gap-3">
              <div class="flex items-center gap-1.5 text-xs text-base-content/55">
                <Loader2 class="h-3.5 w-3.5 animate-spin" />
                {{ t('userCard.loading') }}
              </div>
              <div class="h-8 w-24 rounded-md bg-base-300" />
            </div>
          </div>
          <div v-else-if="error" key="error" class="gf-status-message gf-status-message-error mt-3 flex min-h-[164px] items-center">{{ error }}</div>
          <div v-else key="content">
        <p v-if="bioText" class="mt-3 line-clamp-2 text-sm leading-relaxed text-base-content/75">{{ bioText }}</p>

        <TooltipProvider v-if="visibleBadges.length">
          <div class="mt-3 flex gap-2">
            <Tooltip v-for="badge in visibleBadges" :key="badge.code">
              <TooltipTrigger as-child>
                <span
                  class="group flex h-8 w-8 shrink-0 cursor-default items-center justify-center"
                >
                  <span
                    class="flex h-8 w-8 items-center justify-center shadow-none ring-1 ring-inset transition duration-150 group-hover:-translate-y-0.5 group-hover:scale-110 group-hover:shadow-md"
                    :class="[badgeClass(badge.color, badge.level)]"
                    style="clip-path: polygon(25% 5%, 75% 5%, 100% 50%, 75% 95%, 25% 95%, 0 50%)"
                  >
                    <img :src="badgeIconURL(badge)" :alt="badge.name" class="h-4 w-4 object-contain" />
                  </span>
                </span>
              </TooltipTrigger>
              <TooltipContent side="bottom" class="max-w-48 leading-5">
                {{ badgeTooltip(badge) }}
              </TooltipContent>
            </Tooltip>
          </div>
        </TooltipProvider>

        <div class="mt-3 grid grid-cols-4 divide-x divide-line border-y border-line py-2">
          <div class="px-2 text-center">
            <div class="text-sm font-bold tabular-nums text-base-content">{{ formatNumber(card?.topicCount || 0) }}</div>
            <div class="mt-0.5 text-[11px] text-base-content/55">{{ t('userCard.stats.topics') }}</div>
          </div>
          <div class="px-2 text-center">
            <div class="text-sm font-bold tabular-nums text-base-content">{{ formatNumber(card?.replyCount || 0) }}</div>
            <div class="mt-0.5 text-[11px] text-base-content/55">{{ t('userCard.stats.replies') }}</div>
          </div>
          <div class="px-2 text-center">
            <div class="text-sm font-bold tabular-nums text-base-content">{{ formatNumber(card?.likeReceivedCount || 0) }}</div>
            <div class="mt-0.5 text-[11px] text-base-content/55">{{ t('userCard.stats.likes') }}</div>
          </div>
          <div class="px-2 text-center">
            <div class="text-sm font-bold tabular-nums text-base-content">{{ formatNumber(card?.followerCount || 0) }}</div>
            <div class="mt-0.5 text-[11px] text-base-content/55">{{ t('userCard.stats.followers') }}</div>
          </div>
        </div>

        <TooltipProvider v-if="externalLinks.length">
          <div class="mt-3 flex items-center gap-2 border-b border-line pb-3">
            <Tooltip v-for="link in externalLinks.slice(0, 8)" :key="`${link.key}-${link.url}`">
              <TooltipTrigger as-child>
                <a
                  :href="link.url"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="inline-flex h-7 w-7 items-center justify-center rounded-md text-icon-muted transition hover:bg-base-200 hover:text-primary"
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
                </a>
              </TooltipTrigger>
              <TooltipContent side="top" class="max-w-40 truncate">
                {{ link.label }}
              </TooltipContent>
            </Tooltip>
          </div>
        </TooltipProvider>

        <div class="mt-3 flex items-center justify-between gap-3">
          <div class="inline-flex items-center gap-1.5 text-xs text-base-content/55">
            <CalendarDays class="h-3.5 w-3.5" />
            {{ t('userCard.joinedAt', { date: card?.createdAt ? formatDate(card.createdAt) : '-' }) }}
          </div>
          <Button as-child variant="neutral" size="sm">
            <a :href="profileUrl">
              <UserPlus class="h-4 w-4" />
              {{ card?.isFollowing ? t('userCard.following') : t('userCard.viewProfile') }}
            </a>
          </Button>
        </div>
          </div>
        </Transition>
  </PopoverContent>
</template>
