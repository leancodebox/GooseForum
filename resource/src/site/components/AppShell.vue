<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, ref, watch } from 'vue'
import {
  Bell,
  FileText,
  Flame,
  Heart,
  Inbox,
  Link,
  MessageCircle,
  Languages,
  LogOut,
  Menu,
  PenSquare,
  TrendingUp,
  Search,
  Settings,
  Shield,
  UserRound,
} from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import GlobalFlash from './GlobalFlash.vue'
import UserHoverCard from './UserHoverCard.vue'
import { setLocale, supportedLocales, type Locale } from '@/runtime/i18n'
import { useNavigationState } from '@/runtime/navigation-state'
import { useUnreadStatus } from '@/runtime/unread-status'
import type { LayoutPayload } from '@/types/payload'

const props = defineProps<{
  layout: LayoutPayload
  rail?: boolean
  headerTitle?: string
  showHeaderTitle?: boolean
}>()

const MobileDrawer = defineAsyncComponent(() => import('./MobileDrawer.vue'))
const drawerOpen = ref(false)
const langMenuOpen = ref(false)
const userMenuOpen = ref(false)
const closeTimers: Record<'lang' | 'user', number | undefined> = {
  lang: undefined,
  user: undefined,
}
const { navigating } = useNavigationState()
const { t, locale } = useI18n()
const unreadStatus = useUnreadStatus()
const hasUnreadNotification = computed(() => unreadStatus.notifications.value)
const hasUnreadMessage = computed(() => unreadStatus.messages.value)
const notificationTitle = computed(() => unreadStatus.notificationMessage.value)
const asArray = <T>(value: T[] | null | undefined): T[] => (Array.isArray(value) ? value : [])
const primaryItems = computed(() => asArray(props.layout.sidebar.main))
const resourceItems = computed(() => asArray(props.layout.sidebar.resources))
const categoryItems = computed(() => asArray(props.layout.sidebar.categories))
const headerResourceItems = computed(() =>
  ['sponsors', 'links']
    .map((key) => resourceItems.value.find((item) => item.key === key))
    .filter((item): item is NonNullable<typeof item> => Boolean(item)),
)
const footerLinks = computed(() => asArray(props.layout.footer.links))
const footerPrimary = computed(() => asArray(props.layout.footer.primary))
const hasFooter = computed(() => footerLinks.value.length > 0 || footerPrimary.value.length > 0)
const brandType = computed(() => props.layout.site.brandType || 'default')
const brandText = computed(() => props.layout.site.brandText || props.layout.site.name)
const sidebarIconMap = {
  topics: MessageCircle,
  hot: Flame,
  popular: TrendingUp,
  messages: Inbox,
  notifications: Bell,
  drafts: FileText,
  links: Link,
  sponsors: Heart,
} as const

watch(
  () => props.layout.sidebar.activeKey,
  () => {
    drawerOpen.value = false
    langMenuOpen.value = false
    userMenuOpen.value = false
  },
)

onMounted(() => {
  if (props.layout.viewer.isAuthenticated) {
    unreadStatus.startPolling(props.layout.unread)
  }
})

watch(
  () => props.layout.unread,
  (unread) => {
    if (props.layout.viewer.isAuthenticated) {
      unreadStatus.applyUnread(unread)
    }
  },
  { deep: true },
)

function setLang(lang: Locale) {
  setLocale(lang)
  langMenuOpen.value = false
}

function openDrawer() {
  drawerOpen.value = true
}

function closeDrawer() {
  drawerOpen.value = false
}

async function logout() {
  await fetch('/api/logout', { method: 'POST' })
  window.location.reload()
}

function sidebarIcon(key: string) {
  return sidebarIconMap[key as keyof typeof sidebarIconMap]
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function setHoverMenu(menu: 'lang' | 'user', open: boolean) {
  window.clearTimeout(closeTimers[menu])
  closeTimers[menu] = undefined
  if (menu === 'lang') langMenuOpen.value = open
  else userMenuOpen.value = open
}

function closeHoverMenuSoon(menu: 'lang' | 'user') {
  window.clearTimeout(closeTimers[menu])
  closeTimers[menu] = window.setTimeout(() => {
    if (menu === 'lang') langMenuOpen.value = false
    else userMenuOpen.value = false
  }, 120)
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 text-gray-900">
    <div
      v-show="navigating"
      class="fixed left-0 top-0 z-[100] h-0.5 w-full overflow-hidden bg-blue-100"
    >
      <div class="h-full w-24 animate-[gf-loading-bar_1s_ease-in-out_infinite] rounded-r-full bg-blue-600 sm:w-36" />
    </div>

    <header class="sticky top-0 z-50 border-b border-gray-200/70 bg-white/95 backdrop-blur">
      <div class="mx-auto grid h-16 w-full max-w-[1600px] grid-cols-[minmax(0,1fr)_auto] items-center gap-2 px-3 sm:gap-4 sm:px-5 md:grid-cols-[auto_minmax(0,1fr)_auto] lg:gap-8 lg:px-8">
        <div class="flex min-w-0 items-center gap-2 sm:gap-4 lg:gap-8">
          <button
            type="button"
            class="inline-flex h-9 w-9 items-center justify-center rounded-md text-gray-500 hover:bg-gray-100 hover:text-gray-900 lg:hidden"
            :aria-label="t('shell.openMenu')"
            @click="openDrawer"
          >
            <Menu class="h-5 w-5" />
          </button>
          <a href="/" class="flex min-w-0 items-center gap-2">
            <img
              v-if="brandType === 'image' && layout.site.brandImage"
              :src="layout.site.brandImage"
              :alt="layout.site.name"
              class="h-8 w-auto max-w-32 shrink-0 object-contain sm:max-w-40 sm:h-9"
            />
            <span
              v-else-if="brandType === 'text'"
              class="max-w-36 truncate text-xl font-semibold tracking-tighter text-blue-600 sm:max-w-44 sm:text-2xl md:max-w-none"
            >
              {{ brandText }}
            </span>
            <span v-else class="max-w-36 truncate text-xl font-semibold tracking-tighter text-blue-600 sm:max-w-44 sm:text-2xl md:max-w-none">
              Goose<span class="text-gray-950">Forum</span>
            </span>
          </a>
          <nav
            v-if="!showHeaderTitle"
            class="hidden items-center gap-6 md:flex"
            aria-label="Header navigation"
          >
            <a
              v-for="item in headerResourceItems"
              :key="item.key"
              :href="item.url"
              class="text-sm font-medium text-gray-500 transition hover:text-blue-600"
            >
              {{ item.label }}
            </a>
          </nav>
        </div>

        <div class="hidden min-w-0 md:block">
          <button
            v-if="showHeaderTitle && headerTitle"
            type="button"
            class="block max-w-full truncate text-left text-base font-semibold text-gray-900 transition hover:text-blue-600 sm:text-lg"
            @click="scrollToTop"
          >
            {{ headerTitle }}
          </button>
        </div>

        <div class="flex items-center justify-end gap-0.5 sm:gap-1">
          <a
            href="/search"
            class="hidden h-9 w-9 items-center justify-center rounded-md text-gray-500 hover:bg-gray-100 hover:text-gray-900 sm:inline-flex"
            :aria-label="t('shell.search')"
            :title="t('shell.search')"
          >
            <Search class="h-5 w-5" />
          </a>

          <div
            class="relative hidden md:block"
            @mouseenter="setHoverMenu('lang', true)"
            @mouseleave="closeHoverMenuSoon('lang')"
            @focusin="setHoverMenu('lang', true)"
            @focusout="closeHoverMenuSoon('lang')"
          >
            <button
              type="button"
              class="inline-flex h-9 w-9 items-center justify-center rounded-md text-gray-500 hover:bg-gray-100 hover:text-gray-900"
              :aria-label="t('shell.switchLanguage')"
              :title="t('shell.switchLanguage')"
              :aria-expanded="langMenuOpen"
            >
              <Languages class="h-5 w-5" />
            </button>
            <div
              v-if="langMenuOpen"
              class="absolute right-0 top-full z-[70] w-36 pt-2"
            >
              <div class="rounded-md border border-gray-100 bg-white py-1 shadow-[0_4px_20px_-4px_rgba(0,0,0,0.1)]">
                <button
                  v-for="item in supportedLocales"
                  :key="item"
                  class="block w-full px-3 py-1.5 text-left text-sm hover:bg-gray-50"
                  :class="locale === item ? 'font-semibold text-blue-600' : 'text-gray-700'"
                  type="button"
                  @click="setLang(item)"
                >
                  {{ t(`locale.${item}`) }}
                </button>
              </div>
            </div>
          </div>

          <template v-if="layout.viewer.isAuthenticated">
            <a
              href="/messages"
              class="relative hidden h-9 w-9 items-center justify-center rounded-md text-gray-500 hover:bg-gray-100 hover:text-gray-900 sm:inline-flex"
              aria-label="私信"
              title="私信"
            >
              <Inbox class="h-5 w-5" />
              <span
                v-show="hasUnreadMessage"
                class="absolute right-2 top-2 h-2 w-2 rounded-full bg-red-500 ring-2 ring-white"
              />
            </a>
            <a
              href="/notifications"
              class="relative inline-flex h-9 w-9 items-center justify-center rounded-md text-gray-500 hover:bg-gray-100 hover:text-gray-900"
              :aria-label="t('shell.notifications')"
              :title="notificationTitle"
            >
              <Bell class="h-5 w-5" />
              <span
                v-show="hasUnreadNotification"
                class="absolute right-2 top-2 h-2 w-2 rounded-full bg-red-500 ring-2 ring-white"
              />
            </a>
            <a
              href="/publish"
              class="inline-flex h-9 w-9 items-center justify-center rounded-md text-blue-600 hover:bg-blue-50 hover:text-blue-700"
              :aria-label="t('shell.publish')"
              :title="t('shell.publish')"
            >
              <PenSquare class="h-5 w-5" />
            </a>
            <div
              class="relative"
              @mouseenter="setHoverMenu('user', true)"
              @mouseleave="closeHoverMenuSoon('user')"
              @focusin="setHoverMenu('user', true)"
              @focusout="closeHoverMenuSoon('user')"
            >
              <button
                type="button"
                class="ml-1 flex h-10 w-10 items-center justify-center rounded-full hover:bg-gray-100"
                :aria-label="t('shell.userMenu')"
                :aria-expanded="userMenuOpen"
              >
                <img :src="layout.viewer.avatarUrl" :alt="layout.viewer.username" class="h-9 w-9 rounded-full object-cover" />
              </button>
              <div
                v-if="userMenuOpen"
                class="absolute right-0 top-full z-[70] w-52 pt-2"
              >
                <div class="rounded-md border border-gray-100 bg-white py-1 shadow-[0_4px_20px_-4px_rgba(0,0,0,0.1)]">
                  <div class="border-b border-gray-50 px-3 py-2">
                    <div class="truncate text-sm font-semibold text-gray-900">{{ layout.viewer.username }}</div>
                  </div>
                  <a :href="`/u/${layout.viewer.id}`" class="flex items-center gap-2 px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-50">
                    <UserRound class="h-4 w-4 text-gray-400" /> {{ t('shell.profile') }}
                  </a>
                  <a href="/settings" class="flex items-center gap-2 px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-50">
                    <Settings class="h-4 w-4 text-gray-400" /> {{ t('shell.settings') }}
                  </a>
                  <a v-if="layout.viewer.isAdmin" href="/admin" class="flex items-center gap-2 px-3 py-1.5 text-sm text-amber-700 hover:bg-amber-50">
                    <Shield class="h-4 w-4" /> {{ t('shell.admin') }}
                  </a>
                  <button class="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm text-red-600 hover:bg-red-50" type="button" @click="logout">
                    <LogOut class="h-4 w-4" /> {{ t('shell.logout') }}
                  </button>
                </div>
              </div>
            </div>
          </template>
          <template v-else>
            <a href="/login" class="rounded-md px-3 py-2 text-sm font-medium text-gray-600 hover:bg-gray-100">{{ t('shell.login') }}</a>
            <a href="/login?register=true" class="hidden rounded-md bg-gray-900 px-3 py-2 text-sm font-semibold text-white hover:bg-gray-800 sm:inline-flex">{{ t('shell.register') }}</a>
          </template>
        </div>
      </div>
    </header>

    <GlobalFlash />

    <main
      class="mx-auto grid w-full max-w-[1600px] grid-cols-1 gap-3 px-3 py-3 sm:px-5 lg:grid-cols-[210px_minmax(0,1fr)] lg:px-8 xl:grid-cols-[224px_minmax(0,1fr)]"
      :class="{ 'xl:grid-cols-[224px_minmax(0,1fr)_280px]': rail }"
    >
      <aside class="gf-scrollbar-none sticky top-19 hidden h-[calc(100vh-5.5rem)] overflow-y-auto self-start lg:block" aria-label="Sidebar">
        <nav>
          <div class="pb-2">
            <div class="space-y-0.5">
              <a
                v-for="item in primaryItems"
                :key="item.key"
                :href="item.url"
                class="flex h-8 items-center gap-2 rounded-md px-2 text-[13px] font-medium"
                :class="item.active ? 'bg-blue-50 text-blue-700' : 'text-gray-700 hover:bg-gray-100 hover:text-gray-950'"
              >
                <component
                  :is="sidebarIcon(item.key)"
                  v-if="sidebarIcon(item.key)"
                  class="h-4 w-4 shrink-0"
                  aria-hidden="true"
                />
                <span v-else class="flex w-4 justify-center text-[13px] opacity-80" aria-hidden="true">{{ item.icon }}</span>
                <span class="min-w-0 flex-1 truncate">{{ item.label }}</span>
                <span
                  v-if="(item.key === 'messages' && hasUnreadMessage) || (item.key === 'notifications' && hasUnreadNotification)"
                  class="h-2 w-2 shrink-0 rounded-full bg-red-500"
                  aria-hidden="true"
                />
              </a>
            </div>

            <div v-if="resourceItems.length" class="mt-2">
              <div class="mb-1 px-2 text-[10px] font-bold uppercase tracking-wide text-gray-500">{{ t('shell.resources') }}</div>
              <div class="space-y-px">
                <a
                  v-for="item in resourceItems"
                  :key="item.key"
                  :href="item.url"
                  class="flex h-7 items-center gap-2 rounded-md px-2 text-[13px] font-medium"
                  :class="item.active ? 'bg-blue-50 text-blue-700' : 'text-gray-700 hover:bg-gray-100 hover:text-gray-950'"
                >
                  <component
                    :is="sidebarIcon(item.key)"
                    v-if="sidebarIcon(item.key)"
                    class="h-4 w-4 shrink-0"
                    aria-hidden="true"
                  />
                  <span v-else class="flex w-4 justify-center text-[13px] opacity-80" aria-hidden="true">{{ item.icon }}</span>
                  <span class="truncate">{{ item.label }}</span>
                </a>
              </div>
            </div>

            <div v-if="categoryItems.length" class="mt-2">
              <div class="mb-1 px-2 text-[10px] font-bold uppercase tracking-wide text-gray-500">{{ t('shell.categories') }}</div>
              <div class="space-y-px">
                <a
                  v-for="category in categoryItems"
                  :key="category.key"
                  :href="category.url"
                  class="flex h-7 items-center gap-2 rounded-md px-2 text-[13px] font-medium"
                  :class="category.active ? 'bg-gray-100 text-gray-950' : 'text-gray-700 hover:bg-gray-100 hover:text-gray-950'"
                >
                  <span class="h-2 w-2 rounded-[3px]" :style="{ backgroundColor: category.color }" />
                  <span class="truncate">{{ category.label }}</span>
                </a>
              </div>
            </div>
          </div>

          <footer v-if="hasFooter" class="mt-0 px-2 pt-0.5 text-xs leading-5 text-gray-600">
            <div v-if="footerLinks.length" class="flex flex-wrap items-center gap-x-3 gap-y-0.5">
              <a
                v-for="link in footerLinks"
                :key="`${link.name}-${link.url}`"
                :href="link.url"
                class="inline-flex min-h-5 items-center rounded hover:text-blue-600"
              >
                {{ link.name }}
              </a>
            </div>
            <div v-if="footerPrimary.length" class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-0.5 text-gray-600">
              <span
                v-for="item in footerPrimary"
                :key="item"
                class="inline-flex min-h-5 items-center rounded"
              >
                {{ item }}
              </span>
            </div>
          </footer>
        </nav>
      </aside>

      <section class="min-w-0">
        <slot />
      </section>

      <aside v-if="rail" id="goose-shell-rail" class="hidden min-w-0 xl:block">
        <slot name="rail" />
      </aside>
    </main>

    <MobileDrawer
      :open="drawerOpen"
      :primary-items="primaryItems"
      :resource-items="resourceItems"
      :category-items="categoryItems"
      :footer="layout.footer"
      :has-unread-messages="hasUnreadMessage"
      :has-unread-notifications="hasUnreadNotification"
      :close-label="t('shell.closeMenu')"
      :menu-label="t('shell.menu')"
      :categories-label="t('shell.categories')"
      :sidebar-icon="sidebarIcon"
      @close="closeDrawer"
    />

    <UserHoverCard />
  </div>
</template>
