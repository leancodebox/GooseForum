<script setup lang="ts">
import { adminText } from '@/admin/runtime/i18n-text'
import { computed } from 'vue'
import { ArrowLeft, Languages, Moon, Sun } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Separator } from '@/components/ui/separator'
import { SidebarTrigger } from '@/components/ui/sidebar'
import { setLocale, supportedLocales, type Locale } from '@/runtime/i18n'
import { useSiteTheme } from '@/runtime/site-theme'
import type { LayoutPayload } from '@gooseforum/client'

defineProps<{
  layout: LayoutPayload
}>()

const { t, locale } = useI18n()
const { isDark, toggleTheme } = useSiteTheme()
const themeButtonText = computed(() => isDark.value ? adminText('themeSwitchLight') : adminText('themeSwitchDark'))

function switchLocale(nextLocale: Locale) {
  setLocale(nextLocale)
}
</script>

<template>
  <header class="sticky top-0 z-40 flex h-(--header-height) shrink-0 items-center gap-1 border-b bg-background/92 px-3 text-foreground backdrop-blur transition-[width,height] ease-linear supports-[backdrop-filter]:bg-background/80 lg:gap-2 lg:px-4">
    <SidebarTrigger class="-ml-1 size-8 shrink-0" />
    <Separator orientation="vertical" class="mx-2 h-4 shrink-0" />

    <div id="admin-topbar-page-context" class="min-w-0 flex-1" />

    <div
      id="admin-topbar-primary-action"
      class="flex max-w-[48vw] shrink-0 items-center gap-2 overflow-x-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
    />

    <div class="ml-auto flex shrink-0 items-center gap-1">
      <Button
        size="icon-sm"
        variant="ghost"
        type="button"
        :aria-label="themeButtonText"
        :title="themeButtonText"
        @click="toggleTheme"
      >
        <Sun v-if="isDark" class="size-4" />
        <Moon v-else class="size-4" />
      </Button>
      <DropdownMenu :modal="false">
        <DropdownMenuTrigger as-child>
          <Button
            size="icon-sm"
            variant="ghost"
            type="button"
            :aria-label="t('shell.switchLanguage')"
            :title="t('shell.switchLanguage')"
          >
            <Languages class="size-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" class="w-36">
          <DropdownMenuItem
            v-for="item in supportedLocales"
            :key="item"
            :class="locale === item ? 'font-semibold text-primary focus:text-primary' : ''"
            @select="switchLocale(item)"
          >
            {{ t(`locale.${item}`) }}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <Button as-child size="sm" variant="ghost" class="hidden xl:inline-flex">
        <a href="/">
          <ArrowLeft class="size-4" />
          {{ adminText('k007y') }}
        </a>
      </Button>
    </div>
  </header>
</template>
