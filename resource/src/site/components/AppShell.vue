<script setup lang="ts">
import { computed, defineAsyncComponent, nextTick, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'
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
  Palette,
  PenSquare,
  Scale,
  TrendingUp,
  Search,
  Settings,
  Shield,
  Moon,
  Sun,
  UserRound,
} from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import GlobalFlash from './GlobalFlash.vue'
import { setLocale, supportedLocales, type Locale } from '@/runtime/i18n'
import { useSiteTheme } from '@/runtime/site-theme'
import { useNavigationState } from '@/runtime/navigation-state'
import { useUnreadStatus } from '@/runtime/unread-status'
import type { LayoutPayload } from '@/types/payload'
import type { UserCardShowDetail } from '@/runtime/user-card-events'
import UserAvatar from './UserAvatar.vue'
import type UserCardComponent from './UserCard.vue'

const props = defineProps<{
  layout: LayoutPayload
  rail?: boolean
  headerTitle?: string
  headerTags?: Array<{ id: number | string; name: string; color?: string }>
  showHeaderTitle?: boolean
}>()

interface SidebarNavItem {
  key: string
  label: string
  i18nLabel?: string
  url: string
  active: boolean
}

interface SidebarCategoryItem extends SidebarNavItem {
  id: number
  color: string
}

interface SidebarGroupItem {
  key: string
  title: string
  i18nLabel?: string
  items: SidebarNavItem[]
}

const MobileDrawer = defineAsyncComponent(() => import('./MobileDrawer.vue'))
const UserCard = shallowRef<typeof UserCardComponent | null>(null)
const drawerOpen = ref(false)
const headerElevated = ref(false)
const langMenuOpen = ref(false)
const userMenuOpen = ref(false)
const closeTimers: Record<'lang' | 'user', number | undefined> = {
  lang: undefined,
  user: undefined,
}
const { navigating } = useNavigationState()
const { t, te, locale } = useI18n()
const { isDark, toggleTheme } = useSiteTheme()
const unreadStatus = useUnreadStatus()
const hasUnreadNotification = computed(() => unreadStatus.notifications.value)
const hasUnreadMessage = computed(() => unreadStatus.messages.value)
const hasModerationReports = computed(() => unreadStatus.moderationReports.value)
const notificationTitle = computed(() => unreadStatus.notificationMessage.value)
const asArray = <T>(value: T[] | null | undefined): T[] => (Array.isArray(value) ? value : [])
const activeSidebarKey = computed(() => props.layout.sidebar.activeKey || 'topics')
const primaryItems = computed<SidebarNavItem[]>(() => {
  const items: SidebarNavItem[] = [
    sidebarItem('topics', t('shell.nav.topics'), '/'),
    sidebarItem('hot', t('shell.nav.hot'), '/?sort=hot'),
    sidebarItem('popular', t('shell.nav.popular'), '/?sort=popular'),
  ]
  if (props.layout.viewer.isAuthenticated) {
    items.push(
      sidebarItem('messages', t('shell.nav.messages'), '/messages'),
      sidebarItem('notifications', t('shell.nav.notifications'), '/notifications'),
      sidebarItem('drafts', t('shell.nav.drafts'), '/drafts'),
    )
  }
  if (props.layout.viewer.isModerator) {
    items.push(sidebarItem('moderation', t('shell.nav.moderation'), '/moderation'))
  }
  return [...items, ...serverSidebarItems(props.layout.sidebar.main)]
})
const resourceItems = computed<SidebarNavItem[]>(() => [
  sidebarItem('links', t('shell.nav.links'), '/links'),
  sidebarItem('sponsors', t('shell.nav.sponsors'), '/sponsors'),
  ...serverSidebarItems(props.layout.sidebar.resources),
])
const sidebarGroups = computed<SidebarGroupItem[]>(() =>
  asArray(props.layout.sidebar.groups)
    .map((group) => ({
      key: group.key,
      title: displayNavLabel(group),
      i18nLabel: group.i18nLabel,
      items: serverSidebarItems(group.items),
    }))
    .filter((group) => group.title && group.items.length > 0),
)
const categoryItems = computed<SidebarCategoryItem[]>(() =>
  asArray(props.layout.sidebar.categories).map((category) => {
    const key = `category_${category.id}`
    return {
      key,
      id: category.id,
      label: category.label,
      url: category.url,
      color: category.color,
      active: activeSidebarKey.value === key,
    }
  }),
)
const headerResourceItems = computed(() => {
  const configured = serverSidebarItems(props.layout.header)
  return configured.length > 0
    ? configured
    : ['sponsors', 'links']
      .map((key) => resourceItems.value.find((item) => item.key === key))
      .filter((item): item is NonNullable<typeof item> => Boolean(item))
})
const footerLinks = computed(() => asArray(props.layout.footer.links))
const footerPrimary = computed(() => asArray(props.layout.footer.primary))
const hasFooter = computed(() => footerLinks.value.length > 0 || footerPrimary.value.length > 0)
const brandType = computed(() => props.layout.site.brandType || 'default')
const brandText = computed(() => props.layout.site.brandText || props.layout.site.name)
const hasHeaderTitle = computed(() => Boolean(props.showHeaderTitle && props.headerTitle))
const sidebarIconMap = {
  topics: MessageCircle,
  hot: Flame,
  popular: TrendingUp,
  messages: Inbox,
  notifications: Bell,
  drafts: FileText,
  moderation: Scale,
  links: Link,
  sponsors: Heart,
} as const
let userCardLoading: Promise<void> | undefined

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
  updateHeaderElevated()
  window.addEventListener('scroll', updateHeaderElevated, { passive: true })
  window.addEventListener('goose:user-card-show', ensureUserCardForEvent)
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', updateHeaderElevated)
  window.removeEventListener('goose:user-card-show', ensureUserCardForEvent)
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

function navIcon(item: SidebarNavItem) {
  return sidebarIconMap[item.key as keyof typeof sidebarIconMap] || Link
}

function displayNavLabel(item: { label?: string; title?: string; i18nLabel?: string }) {
  return item.i18nLabel && te(item.i18nLabel) ? t(item.i18nLabel) : item.label || item.title || ''
}

function sidebarItem(key: string, label: string, url: string): SidebarNavItem {
  return {
    key,
    label,
    url,
    active: activeSidebarKey.value === key,
  }
}

function serverSidebarItems(items: typeof props.layout.sidebar.main): SidebarNavItem[] {
  return asArray(items).map((item) => ({
    key: item.key,
    label: displayNavLabel(item),
    i18nLabel: item.i18nLabel,
    url: item.url,
    active: activeSidebarKey.value === item.key,
  }))
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function updateHeaderElevated() {
  headerElevated.value = window.scrollY > 8
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

function ensureUserCardForEvent(event: Event) {
  if (UserCard.value) return
  const detail = (event as CustomEvent<UserCardShowDetail>).detail
  if (!detail?.user?.id || !detail.target) return
  void loadUserCard().then(async () => {
    await nextTick()
    window.dispatchEvent(new CustomEvent<UserCardShowDetail>('goose:user-card-show', { detail }))
  })
}

async function loadUserCard() {
  if (UserCard.value) return
  if (!userCardLoading) {
    userCardLoading = import('./UserCard.vue')
      .then((module) => {
        UserCard.value = module.default
      })
      .finally(() => {
        userCardLoading = undefined
      })
  }
  await userCardLoading
}
</script>

<template>
  <div class="min-h-screen bg-base-200 text-base-content">
    <div
      v-show="navigating"
      class="fixed left-0 top-0 z-[100] h-0.5 w-full overflow-hidden bg-info/10"
    >
      <div class="h-full w-24 animate-[gf-loading-bar_1s_ease-in-out_infinite] rounded-r-full bg-primary sm:w-36" />
    </div>

    <header
      class="sticky top-0 z-50 border-b border-line bg-base-100/95 backdrop-blur-sm transition-[background-color,border-color,box-shadow,backdrop-filter] duration-200"
      :class="headerElevated
        ? 'sm:border-line sm:bg-base-100/95 sm:shadow-[0_1px_10px_rgb(15_23_42/0.04)] sm:backdrop-blur-sm'
        : 'sm:border-transparent sm:bg-base-100/0 sm:shadow-none sm:backdrop-blur-none'"
    >
      <div class="mx-auto grid h-16 w-full max-w-[1600px] grid-cols-[minmax(0,1fr)_auto] items-center gap-2 px-3 sm:gap-4 sm:px-5 md:grid-cols-[auto_minmax(0,1fr)_auto] lg:gap-8 lg:px-8">
        <div class="flex min-w-0 items-center gap-2 sm:gap-4 lg:gap-8">
          <button
            type="button"
            class="inline-flex h-9 w-9 items-center justify-center rounded-md text-icon-muted hover:bg-base-300 hover:text-base-content lg:hidden"
            :aria-label="t('shell.openMenu')"
            @click="openDrawer"
          >
            <Menu class="h-5 w-5" />
          </button>
          <button
            v-if="hasHeaderTitle"
            type="button"
            class="flex min-w-0 flex-1 flex-col items-start justify-center gap-0.5 self-stretch text-left transition md:hidden"
            @click="scrollToTop"
          >
            <span class="block max-w-full truncate text-lg font-semibold leading-6 text-base-content hover:text-primary">
              {{ headerTitle }}
            </span>
            <span
              v-if="headerTags?.length"
              class="flex max-w-full items-center gap-1 overflow-hidden text-[11px] font-medium leading-4 text-base-content/55"
            >
              <span
                v-for="tag in headerTags"
                :key="tag.id"
                class="inline-flex min-w-0 shrink-0 items-center gap-1"
              >
                <span
                  class="h-1.5 w-1.5 rounded-[2px]"
                  :style="{ backgroundColor: tag.color || 'var(--gf-color-base-content)' }"
                />
                <span class="max-w-20 truncate">{{ tag.name }}</span>
              </span>
            </span>
          </button>
          <a
            href="/"
            class="min-w-0 items-center gap-2"
            :class="hasHeaderTitle ? 'hidden md:flex' : 'flex'"
          >
            <img
              v-if="brandType === 'image' && layout.site.brandImage"
              :src="layout.site.brandImage"
              :alt="layout.site.name"
              class="h-8 w-auto max-w-32 shrink-0 object-contain sm:max-w-40 sm:h-9"
            />
            <span
              v-else-if="brandType === 'text'"
              class="max-w-36 truncate text-xl font-semibold tracking-tighter text-primary sm:max-w-44 sm:text-2xl md:max-w-none"
            >
              {{ brandText }}
            </span>
            <span v-else class="max-w-36 truncate text-xl font-semibold tracking-tighter text-primary sm:max-w-44 sm:text-2xl md:max-w-none">
              Goose<span class="text-base-content">Forum</span>
            </span>
          </a>
          <nav
            v-if="!showHeaderTitle"
            class="hidden items-center gap-1 md:flex"
            aria-label="Header navigation"
          >
            <a
              v-for="item in headerResourceItems"
              :key="item.key"
              :href="item.url"
              class="inline-flex h-7 items-center rounded-md px-2 text-sm font-medium text-base-content/75 transition-colors duration-150 hover:bg-base-300 hover:text-base-content"
            >
              {{ item.label }}
            </a>
          </nav>
        </div>

        <div class="hidden min-w-0 md:block">
          <button
            v-if="hasHeaderTitle"
            type="button"
            class="flex h-16 max-w-full flex-col items-start justify-center gap-0.5 text-left transition"
            @click="scrollToTop"
          >
            <span class="block max-w-full truncate text-xl font-semibold leading-6 text-base-content hover:text-primary">
              {{ headerTitle }}
            </span>
            <span
              v-if="headerTags?.length"
              class="flex max-w-full items-center gap-2 overflow-hidden text-[11px] font-medium leading-4 text-base-content/55"
            >
              <span
                v-for="tag in headerTags"
                :key="tag.id"
                class="inline-flex min-w-0 shrink-0 items-center gap-1"
              >
                <span
                  class="h-1.5 w-1.5 rounded-[2px]"
                  :style="{ backgroundColor: tag.color || 'var(--gf-color-base-content)' }"
                />
                <span class="max-w-28 truncate">{{ tag.name }}</span>
              </span>
            </span>
          </button>
        </div>

        <div
          class="items-center justify-end gap-0.5 sm:gap-1"
          :class="hasHeaderTitle ? 'hidden md:flex' : 'flex'"
        >
          <a
            href="/search"
            class="hidden h-9 w-9 items-center justify-center rounded-full text-icon-muted transition-colors duration-150 hover:bg-base-300 hover:text-base-content sm:inline-flex"
            :aria-label="t('shell.search')"
            :title="t('shell.search')"
          >
            <Search class="h-5 w-5" />
          </a>

          <button
            type="button"
            class="inline-flex h-9 w-9 items-center justify-center rounded-full text-icon-muted transition-colors duration-150 hover:bg-base-300 hover:text-base-content"
            :aria-label="isDark ? 'Switch to light theme' : 'Switch to dark theme'"
            :title="isDark ? 'Light' : 'Dark'"
            @click="toggleTheme"
          >
            <Sun v-if="isDark" class="h-5 w-5" />
            <Moon v-else class="h-5 w-5" />
          </button>

          <div
            class="relative"
            @mouseenter="setHoverMenu('lang', true)"
            @mouseleave="closeHoverMenuSoon('lang')"
            @focusin="setHoverMenu('lang', true)"
            @focusout="closeHoverMenuSoon('lang')"
          >
            <button
              type="button"
              class="inline-flex h-9 w-9 items-center justify-center rounded-full text-icon-muted transition-colors duration-150 hover:bg-base-300 hover:text-base-content"
              :aria-label="t('shell.switchLanguage')"
              :title="t('shell.switchLanguage')"
              :aria-expanded="langMenuOpen"
              @click="langMenuOpen = !langMenuOpen"
            >
              <Languages class="h-5 w-5" />
            </button>
            <Transition name="gf-menu">
              <div
                v-if="langMenuOpen"
                class="absolute right-0 top-full z-[70] w-36 pt-2"
              >
                <div class="gf-menu-surface overflow-hidden py-1">
                  <button
                    v-for="item in supportedLocales"
                    :key="item"
                    class="block w-full px-3 py-1.5 text-left text-sm transition-colors duration-150 hover:bg-base-200"
                    :class="locale === item ? 'font-semibold text-primary' : 'text-base-content/75'"
                    type="button"
                    @click="setLang(item)"
                  >
                    {{ t(`locale.${item}`) }}
                  </button>
                </div>
              </div>
            </Transition>
          </div>

          <template v-if="layout.viewer.isAuthenticated">
            <div
              class="relative"
              @mouseenter="setHoverMenu('user', true)"
              @mouseleave="closeHoverMenuSoon('user')"
              @focusin="setHoverMenu('user', true)"
              @focusout="closeHoverMenuSoon('user')"
            >
              <button
                type="button"
                class="relative ml-1 flex h-10 w-10 items-center justify-center rounded-full transition-colors duration-150 hover:bg-base-300"
                :aria-label="t('shell.userMenu')"
                :aria-expanded="userMenuOpen"
              >
                <UserAvatar :src="layout.viewer.avatarUrl" :alt="layout.viewer.username" class="h-9 w-9 rounded-full object-cover ring-1 ring-line/80" />
                <span
                  v-show="hasUnreadMessage || hasUnreadNotification"
                  class="absolute right-0.5 top-0.5 h-2.5 w-2.5 rounded-full bg-error ring-2 ring-base-100"
                />
              </button>
              <Transition name="gf-menu">
                <div
                  v-if="userMenuOpen"
                  class="absolute right-0 top-full z-[70] w-56 pt-2"
                >
                  <div class="gf-menu-surface overflow-hidden">
                    <div class="border-b border-line/70 px-3 py-2.5">
                      <div class="truncate text-sm font-semibold text-base-content">{{ layout.viewer.username }}</div>
                    </div>
                    <div class="py-1">
                      <a :href="`/u/${layout.viewer.id}`" class="gf-menu-item">
                        <UserRound class="h-4 w-4 text-icon-muted" /> {{ t('shell.profile') }}
                      </a>
                      <a href="/messages" class="gf-menu-item">
                        <Inbox class="h-4 w-4 text-icon-muted" />
                        <span class="min-w-0 flex-1">{{ t('shell.nav.messages') }}</span>
                        <span v-show="hasUnreadMessage" class="h-2 w-2 rounded-full bg-error" />
                      </a>
                      <a href="/notifications" class="gf-menu-item" :title="notificationTitle">
                        <Bell class="h-4 w-4 text-icon-muted" />
                        <span class="min-w-0 flex-1">{{ t('shell.nav.notifications') }}</span>
                        <span v-show="hasUnreadNotification" class="h-2 w-2 rounded-full bg-error" />
                      </a>
                      <a href="/drafts" class="gf-menu-item">
                        <FileText class="h-4 w-4 text-icon-muted" /> {{ t('shell.nav.drafts') }}
                      </a>
                    </div>
                    <div class="border-t border-line/70 py-1">
                      <a href="/publish" class="gf-menu-item-primary">
                        <PenSquare class="h-4 w-4" /> {{ t('shell.publish') }}
                      </a>
                      <a href="/settings" class="gf-menu-item">
                        <Settings class="h-4 w-4 text-icon-muted" /> {{ t('shell.settings') }}
                      </a>
                      <a v-if="layout.viewer.canAccessAdmin" href="/theme-preview" class="gf-menu-item">
                        <Palette class="h-4 w-4 text-icon-muted" /> {{ t('shell.themePreview') }}
                      </a>
                      <a v-if="layout.viewer.canAccessAdmin" href="/admin" class="gf-menu-item-warning">
                        <Shield class="h-4 w-4" /> {{ t('shell.admin') }}
                      </a>
                    </div>
                    <div class="border-t border-line/70 py-1">
                      <button class="gf-menu-item-danger" type="button" @click="logout">
                        <LogOut class="h-4 w-4" /> {{ t('shell.logout') }}
                      </button>
                    </div>
                  </div>
                </div>
              </Transition>
            </div>
          </template>
          <template v-else>
            <a href="/login" class="rounded-md px-3 py-2 text-sm font-medium text-base-content/75 hover:bg-base-300">{{ t('shell.login') }}</a>
            <a href="/login?register=true" class="gf-button gf-button-md gf-button-neutral hidden sm:inline-flex">{{ t('shell.register') }}</a>
          </template>
        </div>
      </div>
    </header>

    <GlobalFlash />

    <main
      class="gf-shell-main mx-auto grid w-full max-w-[1600px] grid-cols-1 gap-0 px-0 py-0 sm:gap-3 sm:px-5 sm:py-3 lg:grid-cols-[210px_minmax(0,1fr)] lg:px-8 xl:grid-cols-[224px_minmax(0,1fr)]"
      :class="{ 'xl:grid-cols-[224px_minmax(0,1fr)_280px]': rail }"
    >
      <aside class="gf-scrollbar-none sticky top-16 -my-3 hidden h-[calc(100vh-4rem)] overflow-y-auto self-start lg:block" aria-label="Sidebar">
        <nav class="py-3">
          <div class="pb-2">
            <div class="space-y-0.5">
              <a
                v-for="item in primaryItems"
                :key="item.key"
                :href="item.url"
                class="flex h-8 items-center gap-2 rounded-md px-2 text-[13px] font-medium transition-colors duration-150"
                :class="item.active ? 'bg-info/10 text-primary' : 'text-base-content/75 hover:bg-base-300 hover:text-base-content'"
              >
                <component
                  :is="navIcon(item)"
                  v-if="navIcon(item)"
                  class="h-4 w-4 shrink-0"
                  aria-hidden="true"
                />
                <span class="min-w-0 flex-1 truncate">{{ item.label }}</span>
                <span
                  v-if="(item.key === 'messages' && hasUnreadMessage) || (item.key === 'notifications' && hasUnreadNotification) || (item.key === 'moderation' && hasModerationReports)"
                  class="h-2 w-2 shrink-0 rounded-full bg-error/100"
                  aria-hidden="true"
                />
              </a>
            </div>

            <div v-if="resourceItems.length" class="mt-2">
              <div class="mb-1 px-2 text-[10px] font-bold uppercase tracking-wide text-base-content/55">{{ t('shell.resources') }}</div>
              <div class="space-y-px">
                <a
                  v-for="item in resourceItems"
                  :key="item.key"
                  :href="item.url"
                  class="flex h-7 items-center gap-2 rounded-md px-2 text-[13px] font-medium transition-colors duration-150"
                  :class="item.active ? 'bg-info/10 text-primary' : 'text-base-content/75 hover:bg-base-300 hover:text-base-content'"
                >
                  <component
                    :is="navIcon(item)"
                    v-if="navIcon(item)"
                    class="h-4 w-4 shrink-0"
                    aria-hidden="true"
                  />
                  <span class="truncate">{{ item.label }}</span>
                </a>
              </div>
            </div>

            <div
              v-for="group in sidebarGroups"
              :key="group.key"
              class="mt-2"
            >
              <div class="mb-1 px-2 text-[10px] font-bold uppercase tracking-wide text-base-content/55">{{ group.title }}</div>
              <div class="space-y-px">
                <a
                  v-for="item in group.items"
                  :key="item.key"
                  :href="item.url"
                  class="flex h-7 items-center gap-2 rounded-md px-2 text-[13px] font-medium transition-colors duration-150"
                  :class="item.active ? 'bg-info/10 text-primary' : 'text-base-content/75 hover:bg-base-300 hover:text-base-content'"
                >
                  <component
                    :is="navIcon(item)"
                    v-if="navIcon(item)"
                    class="h-4 w-4 shrink-0"
                    aria-hidden="true"
                  />
                  <span class="truncate">{{ item.label }}</span>
                </a>
              </div>
            </div>

            <div v-if="categoryItems.length" class="mt-2">
              <div class="mb-1 px-2 text-[10px] font-bold uppercase tracking-wide text-base-content/55">{{ t('shell.categories') }}</div>
              <div class="space-y-px">
                <a
                  v-for="category in categoryItems"
                  :key="category.key"
                  :href="category.url"
                  class="flex h-7 items-center gap-2 rounded-md px-2 text-[13px] font-medium transition-colors duration-150"
                  :class="category.active ? 'bg-base-300 text-base-content' : 'text-base-content/75 hover:bg-base-300 hover:text-base-content'"
                >
                  <span class="h-2 w-2 rounded-[3px]" :style="{ backgroundColor: category.color }" />
                  <span class="truncate">{{ category.label }}</span>
                </a>
              </div>
            </div>
          </div>

          <footer v-if="hasFooter" class="mt-0 px-2 pt-0.5 text-xs leading-5 text-base-content/75">
            <div v-if="footerLinks.length" class="flex flex-wrap items-center gap-x-3 gap-y-0.5">
              <a
                v-for="link in footerLinks"
                :key="`${link.name}-${link.url}`"
                :href="link.url"
                class="inline-flex min-h-5 items-center rounded hover:text-primary"
              >
                {{ link.name }}
              </a>
            </div>
            <div v-if="footerPrimary.length" class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-0.5 text-base-content/75">
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

      <section class="gf-shell-content min-w-0">
        <slot />
      </section>

      <aside v-if="rail" id="goose-shell-rail" class="hidden min-w-0 xl:block">
        <slot name="rail" />
      </aside>

      <section
        id="goose-shell-wide-content"
        class="min-w-0 empty:hidden lg:col-start-2 lg:row-start-2 xl:col-start-2 xl:col-span-2"
      />
    </main>

    <MobileDrawer
      v-if="drawerOpen"
      :open="drawerOpen"
      :primary-items="primaryItems"
      :resource-items="resourceItems"
      :sidebar-groups="sidebarGroups"
      :category-items="categoryItems"
      :footer="layout.footer"
      :has-unread-messages="hasUnreadMessage"
      :has-unread-notifications="hasUnreadNotification"
      :has-moderation-reports="hasModerationReports"
      :close-label="t('shell.closeMenu')"
      :menu-label="t('shell.menu')"
      :resources-label="t('shell.resources')"
      :categories-label="t('shell.categories')"
      :sidebar-icon="navIcon"
      @close="closeDrawer"
    />

    <component :is="UserCard" v-if="UserCard" />
  </div>
</template>
