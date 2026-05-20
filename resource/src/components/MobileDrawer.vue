<script setup lang="ts">
import { computed } from 'vue'
import { AnimatePresence, Motion } from 'motion-v'
import { X } from '@lucide/vue'
import { mobileDrawerMotion, motionTransitions, overlayMotion } from '@/runtime/motion'
import type { CategoryNavPayload, FooterPayload, NavItemPayload } from '@/types/payload'

const props = defineProps<{
  open: boolean
  primaryItems: NavItemPayload[]
  resourceItems: NavItemPayload[]
  categoryItems: CategoryNavPayload[]
  footer: FooterPayload
  hasUnreadMessages?: boolean
  hasUnreadNotifications?: boolean
  closeLabel: string
  menuLabel: string
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
        class="absolute inset-0 bg-gray-900/40"
        :aria-label="closeLabel"
        v-bind="overlayMotion"
        :transition="motionTransitions.fast"
        @click="close"
      />
      <Motion
        as="nav"
        class="relative h-full w-80 max-w-[85vw] overflow-y-auto bg-white p-3 shadow-xl"
        v-bind="mobileDrawerMotion"
        :transition="motionTransitions.comfortable"
      >
        <div class="mb-3 flex h-10 items-center justify-between">
          <div class="font-bold text-gray-900">{{ menuLabel }}</div>
          <button class="inline-flex h-8 w-8 items-center justify-center rounded-md text-gray-500 hover:bg-gray-100" type="button" :aria-label="closeLabel" @click="close">
            <X class="h-5 w-5" />
          </button>
        </div>
        <div class="space-y-0.5">
          <a
            v-for="item in [...primaryItems, ...resourceItems]"
            :key="item.key"
            :href="item.url"
            class="flex h-9 items-center gap-2 rounded-md px-2 text-sm font-medium"
            :class="item.active ? 'bg-blue-50 text-blue-700' : 'text-gray-700 hover:bg-gray-100 hover:text-gray-950'"
          >
            <component
              :is="sidebarIcon(item.key)"
              v-if="sidebarIcon(item.key)"
              class="h-4 w-4 shrink-0"
              aria-hidden="true"
            />
            <span v-else class="flex w-4 justify-center opacity-80" aria-hidden="true">{{ item.icon }}</span>
            <span class="min-w-0 flex-1 truncate">{{ item.label }}</span>
            <span
              v-if="(item.key === 'messages' && hasUnreadMessages) || (item.key === 'notifications' && hasUnreadNotifications)"
              class="h-2 w-2 shrink-0 rounded-full bg-red-500"
              aria-hidden="true"
            />
          </a>
        </div>
        <div v-if="categoryItems.length" class="mt-5 space-y-0.5">
          <div class="px-2 text-[10px] font-bold uppercase tracking-wide text-gray-500">{{ categoriesLabel }}</div>
          <a
            v-for="category in categoryItems"
            :key="category.key"
            :href="category.url"
            class="flex h-9 items-center gap-2 rounded-md px-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
          >
            <span class="h-2 w-2 rounded-[3px]" :style="{ backgroundColor: category.color }" />
            <span>{{ category.label }}</span>
          </a>
        </div>
        <footer v-if="hasFooter" class="mt-2 border-t border-gray-100 px-2 pt-2 text-xs leading-5 text-gray-600">
          <div v-if="footer.links.length" class="flex flex-wrap items-center gap-x-3 gap-y-0.5">
            <a
              v-for="link in footer.links"
              :key="`${link.name}-${link.url}`"
              :href="link.url"
              class="inline-flex min-h-6 items-center rounded hover:text-blue-600"
            >
              {{ link.name }}
            </a>
          </div>
          <div v-if="footer.primary.length" class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-0.5 text-gray-600">
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
