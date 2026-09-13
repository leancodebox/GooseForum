<script setup lang="ts">
import { computed } from 'vue'
import { Sheet, SheetContent, SheetTitle } from '@/components/ui/sheet'
import type { FooterPayload } from '@gooseforum/client'

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

const props = defineProps<{
  open: boolean
  primaryItems: SidebarNavItem[]
  resourceItems: SidebarNavItem[]
  sidebarGroups: SidebarGroupItem[]
  categoryItems: SidebarCategoryItem[]
  footer: FooterPayload
  hasUnreadMessages?: boolean
  hasUnreadNotifications?: boolean
  hasModerationReports?: boolean
  closeLabel: string
  menuLabel: string
  resourcesLabel: string
  categoriesLabel: string
  sidebarIcon: (item: SidebarNavItem) => unknown
}>()

const emit = defineEmits<{
  close: []
}>()

const hasFooter = computed(() => props.footer.links.length > 0 || props.footer.primary.length > 0)

function close() {
  emit('close')
}

function updateOpen(open: boolean) {
  if (!open) close()
}
</script>

<template>
  <Sheet :open="open" @update:open="updateOpen">
    <SheetContent
      side="left"
      :close-label="closeLabel"
      overlay-class="z-[60] bg-neutral/40 lg:hidden"
      class="gf-drawer-surface z-[60] w-80 max-w-[85vw] gap-0 overflow-hidden border-r border-line bg-base-100 p-0 sm:max-w-[85vw] lg:hidden [&>button]:right-4 [&>button]:top-4 [&>button]:grid [&>button]:size-8 [&>button]:place-items-center [&>button]:rounded-md [&>button]:text-icon-muted [&>button]:opacity-100 [&>button]:hover:bg-base-300 [&>button]:hover:text-base-content [&>button_svg]:size-5"
    >
      <nav class="h-full overflow-y-auto p-3">
        <div class="mb-3 flex h-10 items-center justify-between">
          <SheetTitle class="font-bold text-base-content">{{ menuLabel }}</SheetTitle>
        </div>
        <div class="space-y-0.5">
          <a
            v-for="item in primaryItems"
            :key="item.key"
            :href="item.url"
            class="flex h-9 items-center gap-2 rounded-md px-2 text-sm font-medium"
            :class="item.active ? 'bg-info/10 text-primary' : 'text-base-content/75 hover:bg-base-300 hover:text-base-content'"
          >
            <component
              :is="sidebarIcon(item)"
              v-if="sidebarIcon(item)"
              class="h-4 w-4 shrink-0"
              aria-hidden="true"
            />
            <span class="min-w-0 flex-1 truncate">{{ item.label }}</span>
            <span
              v-if="(item.key === 'messages' && hasUnreadMessages) || (item.key === 'notifications' && hasUnreadNotifications) || (item.key === 'moderation' && hasModerationReports)"
              class="h-2 w-2 shrink-0 rounded-full bg-error/100"
              aria-hidden="true"
            />
          </a>
        </div>
        <div v-if="resourceItems.length" class="mt-4 space-y-0.5">
          <div class="px-2 text-[10px] font-bold uppercase tracking-wide text-base-content/55">{{ resourcesLabel }}</div>
          <a
            v-for="item in resourceItems"
            :key="item.key"
            :href="item.url"
            class="flex h-9 items-center gap-2 rounded-md px-2 text-sm font-medium"
            :class="item.active ? 'bg-info/10 text-primary' : 'text-base-content/75 hover:bg-base-300 hover:text-base-content'"
          >
            <component
              :is="sidebarIcon(item)"
              v-if="sidebarIcon(item)"
              class="h-4 w-4 shrink-0"
              aria-hidden="true"
            />
            <span class="min-w-0 flex-1 truncate">{{ item.label }}</span>
          </a>
        </div>
        <div
          v-for="group in sidebarGroups"
          :key="group.key"
          class="mt-4 space-y-0.5"
        >
          <div class="px-2 text-[10px] font-bold uppercase tracking-wide text-base-content/55">{{ group.title }}</div>
          <a
            v-for="item in group.items"
            :key="item.key"
            :href="item.url"
            class="flex h-9 items-center gap-2 rounded-md px-2 text-sm font-medium"
            :class="item.active ? 'bg-info/10 text-primary' : 'text-base-content/75 hover:bg-base-300 hover:text-base-content'"
          >
            <component
              :is="sidebarIcon(item)"
              v-if="sidebarIcon(item)"
              class="h-4 w-4 shrink-0"
              aria-hidden="true"
            />
            <span class="min-w-0 flex-1 truncate">{{ item.label }}</span>
          </a>
        </div>
        <div v-if="categoryItems.length" class="mt-4 space-y-0.5">
          <div class="px-2 text-[10px] font-bold uppercase tracking-wide text-base-content/55">{{ categoriesLabel }}</div>
          <a
            v-for="category in categoryItems"
            :key="category.key"
            :href="category.url"
            class="flex h-9 items-center gap-2 rounded-md px-2 text-sm font-medium"
            :class="category.active ? 'bg-base-300 text-base-content' : 'text-base-content/75 hover:bg-base-300 hover:text-base-content'"
          >
            <span class="h-2 w-2 rounded-[3px]" :style="{ backgroundColor: category.color }" />
            <span class="min-w-0 flex-1 truncate">{{ category.label }}</span>
          </a>
        </div>
        <footer v-if="hasFooter" class="mt-2 border-t border-line px-2 pt-2 text-xs leading-5 text-base-content/75">
          <div v-if="footer.links.length" class="flex flex-wrap items-center gap-x-3 gap-y-0.5">
            <a
              v-for="link in footer.links"
              :key="`${link.name}-${link.url}`"
              :href="link.url"
              class="inline-flex min-h-6 items-center rounded hover:text-primary"
            >
              {{ link.name }}
            </a>
          </div>
          <div v-if="footer.primary.length" class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-0.5 text-base-content/75">
            <span
              v-for="item in footer.primary"
              :key="item"
              class="inline-flex min-h-6 items-center rounded"
            >
              {{ item }}
            </span>
          </div>
        </footer>
      </nav>
    </SheetContent>
  </Sheet>
</template>
