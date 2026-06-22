<script setup lang="ts">
import { computed } from 'vue'
import { AnimatePresence, Motion } from 'motion-v'
import { X } from '@lucide/vue'
import { mobileDrawerMotion, motionTransitions, overlayMotion } from '@/runtime/motion'
import type { FooterPayload } from '@/types/payload'

interface SidebarNavItem {
  key: string
  label: string
  url: string
  active: boolean
  icon?: string
}

interface SidebarCategoryItem extends SidebarNavItem {
  id: number
  color: string
}

const props = defineProps<{
  open: boolean
  primaryItems: SidebarNavItem[]
  resourceItems: SidebarNavItem[]
  categoryItems: SidebarCategoryItem[]
  footer: FooterPayload
  hasUnreadMessages?: boolean
  hasUnreadNotifications?: boolean
  hasModerationReports?: boolean
  closeLabel: string
  menuLabel: string
  resourcesLabel: string
  categoriesLabel: string
  sidebarIcon: (key: string) => unknown
}>()

const emit = defineEmits<{
  close: []
}>()

const hasFooter = computed(() => props.footer.links.length > 0 || props.footer.primary.length > 0)

function close() {
  emit('close')
}
</script>

<template>
  <AnimatePresence>
    <div v-if="open" class="fixed inset-0 z-[60] lg:hidden" role="dialog" aria-modal="true">
      <Motion
        as="button"
        class="absolute inset-0 bg-neutral/40"
        :aria-label="closeLabel"
        v-bind="overlayMotion"
        :transition="motionTransitions.fast"
        @click="close"
      />
      <Motion
        as="nav"
        class="gf-drawer-surface relative h-full w-80 max-w-[85vw] overflow-y-auto p-3"
        v-bind="mobileDrawerMotion"
        :transition="motionTransitions.comfortable"
      >
        <div class="mb-3 flex h-10 items-center justify-between">
          <div class="font-bold text-base-content">{{ menuLabel }}</div>
          <button class="inline-flex h-8 w-8 items-center justify-center rounded-md text-icon-muted hover:bg-base-300 hover:text-base-content" type="button" :aria-label="closeLabel" @click="close">
            <X class="h-5 w-5" />
          </button>
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
              :is="sidebarIcon(item.key)"
              v-if="sidebarIcon(item.key)"
              class="h-4 w-4 shrink-0"
              aria-hidden="true"
            />
            <span v-else-if="item.icon" class="flex w-4 justify-center opacity-80" aria-hidden="true">{{ item.icon }}</span>
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
              :is="sidebarIcon(item.key)"
              v-if="sidebarIcon(item.key)"
              class="h-4 w-4 shrink-0"
              aria-hidden="true"
            />
            <span v-else-if="item.icon" class="flex w-4 justify-center opacity-80" aria-hidden="true">{{ item.icon }}</span>
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
      </Motion>
    </div>
  </AnimatePresence>
</template>
