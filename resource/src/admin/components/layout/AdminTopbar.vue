<script setup lang="ts">
import { adminText } from '@/admin/runtime/i18n-text'
import { ref } from 'vue'
import { ArrowLeft, Languages } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { Button } from '@/admin/components/ui/button'
import { Separator } from '@/admin/components/ui/separator'
import { SidebarTrigger } from '@/admin/components/ui/sidebar'
import { setLocale, supportedLocales, type Locale } from '@/runtime/i18n'
import type { LayoutPayload } from '@gooseforum/client'

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
  <header class="sticky top-0 z-50 flex h-[4.5rem] shrink-0 items-center gap-3 border-b bg-background/95 px-4 backdrop-blur transition-[width,height] ease-linear supports-[backdrop-filter]:bg-background/85 sm:gap-4">
    <SidebarTrigger class="-ml-1 shrink-0" />
    <Separator orientation="vertical" class="h-8 shrink-0" />

    <div id="admin-topbar-page-context" class="min-w-0 flex-1" />

    <div
      id="admin-topbar-primary-action"
      class="flex max-w-[48vw] shrink-0 items-center gap-2 overflow-x-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
    />

    <div class="ml-auto flex shrink-0 items-center gap-2 sm:gap-3">
      <div class="relative hidden lg:block">
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
      <Button as-child class="hidden xl:inline-flex">
        <a href="/">
          <ArrowLeft class="size-4" />
          {{ adminText('k007y') }}
        </a>
      </Button>
      <img
        v-if="layout.viewer.isAuthenticated"
        :src="layout.viewer.avatarUrl"
        :alt="layout.viewer.username"
        class="hidden size-9 rounded-full object-cover sm:block"
      />
    </div>
  </header>
</template>
