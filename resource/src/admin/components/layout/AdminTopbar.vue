<script setup lang="ts">
import { adminText } from '@/admin/runtime/i18n-text'
import { computed, ref } from 'vue'
import { ArrowLeft, Languages, Search } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { Button } from '@/admin/components/ui/button'
import { Separator } from '@/admin/components/ui/separator'
import { SidebarTrigger } from '@/admin/components/ui/sidebar'
import { setLocale, supportedLocales, type Locale } from '@/runtime/i18n'
import type { LayoutPayload } from '@/types/payload'

defineProps<{
  layout: LayoutPayload
}>()

const { t, locale } = useI18n()
const languageMenuOpen = ref(false)
const currentLanguageLabel = computed(() => t(`locale.short.${locale.value}`))

function switchLocale(nextLocale: Locale) {
  setLocale(nextLocale)
  languageMenuOpen.value = false
}
</script>

<template>
  <header class="sticky top-0 z-50 flex h-16 shrink-0 items-center gap-3 bg-background p-4 transition-[width,height] ease-linear sm:gap-4">
    <SidebarTrigger class="-ml-1" />
    <Separator orientation="vertical" class="h-6" />

    <button
      type="button"
      class="hidden h-10 w-72 items-center gap-3 rounded-md border bg-background px-3 text-sm text-muted-foreground shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground md:flex"
    >
      <Search class="size-4" />
      <span class="flex-1 text-left">Search Menu</span>
      <kbd class="pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground">
        ⌘ + K
      </kbd>
    </button>

    <div class="flex-1" />

    <div class="ml-auto flex items-center space-x-4">
      <div class="relative">
        <button
          type="button"
          class="inline-flex h-9 items-center gap-2 rounded-md border bg-background px-2.5 text-sm font-medium text-muted-foreground shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground"
          :aria-label="t('shell.switchLanguage')"
          :title="t('shell.switchLanguage')"
          @click="languageMenuOpen = !languageMenuOpen"
        >
          <Languages class="size-4" />
          <span class="min-w-5 text-center">{{ currentLanguageLabel }}</span>
        </button>
        <div
          v-if="languageMenuOpen"
          class="absolute right-0 z-50 mt-2 w-36 overflow-hidden rounded-md border bg-popover py-1 text-popover-foreground shadow-lg"
        >
          <button
            v-for="item in supportedLocales"
            :key="item"
            type="button"
            class="block w-full px-3 py-1.5 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground"
            :class="locale === item ? 'font-semibold text-primary' : 'text-popover-foreground'"
            @click="switchLocale(item)"
          >
            {{ t(`locale.${item}`) }}
          </button>
        </div>
      </div>
      <Button as-child class="hidden md:inline-flex">
        <a href="/">
          <ArrowLeft class="size-4" />
          {{ adminText('k007y') }}
        </a>
      </Button>
      <img
        v-if="layout.viewer.isAuthenticated"
        :src="layout.viewer.avatarUrl"
        :alt="layout.viewer.username"
        class="size-10 rounded-full object-cover"
      />
    </div>
  </header>
</template>
