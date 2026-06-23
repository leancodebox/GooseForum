<script setup lang="ts">
import { adminText } from '@/admin/runtime/i18n-text'
import { ref } from 'vue'
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

function switchLocale(nextLocale: Locale) {
  setLocale(nextLocale)
  languageMenuOpen.value = false
}
</script>

<template>
  <header class="sticky top-0 z-50 flex h-16 shrink-0 items-center gap-3 bg-background p-4 transition-[width,height] ease-linear sm:gap-4">
    <SidebarTrigger class="-ml-1" />
    <Separator orientation="vertical" class="h-6" />

    <Button
      variant="outline"
      type="button"
      class="hidden h-10 w-72 justify-start px-3 text-muted-foreground md:inline-flex"
    >
      <Search class="size-4" />
      <span class="flex-1 text-left">Search Menu</span>
      <kbd class="pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground">
        ⌘ + K
      </kbd>
    </Button>

    <div class="flex-1" />

    <div class="ml-auto flex items-center space-x-4">
      <div class="relative">
        <Button
          variant="outline"
          size="sm"
          type="button"
          class="text-muted-foreground"
          :aria-label="t('shell.switchLanguage')"
          :title="t('shell.switchLanguage')"
          @click="languageMenuOpen = !languageMenuOpen"
        >
          <Languages class="size-4" />
        </Button>
        <div
          v-if="languageMenuOpen"
          class="absolute right-0 z-50 mt-2 w-36 overflow-hidden rounded-md border bg-popover py-1 text-popover-foreground shadow-lg"
        >
          <Button
            v-for="item in supportedLocales"
            :key="item"
            variant="ghost"
            size="sm"
            type="button"
            class="w-full justify-start rounded-none"
            :class="locale === item ? 'font-semibold text-primary hover:text-primary' : 'text-popover-foreground'"
            @click="switchLocale(item)"
          >
            {{ t(`locale.${item}`) }}
          </Button>
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
