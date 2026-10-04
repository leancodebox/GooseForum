"use client"

import { ToggleGroup, ToggleGroupItem } from '@gooseforum/ui/components/toggle-group'
import { authLocales, localeLabels } from '@gooseforum/runtime/i18n/auth'
import { useGooseRuntime } from '@gooseforum/runtime'

export function AuthLocaleSwitcher() {
  const runtime = useGooseRuntime()

  return (
    <ToggleGroup
      type="single"
      size="sm"
      value={runtime.locale}
      onValueChange={(value) => { if (value) void runtime.setLocale(value as typeof runtime.locale) }}
      className="absolute right-4 top-4 bg-muted p-1"
      aria-label="Language"
    >
        {authLocales.map((locale) => {
          return (
            <ToggleGroupItem key={locale} value={locale} title={localeLabels[locale].label} className="h-7 min-w-8 rounded-md px-2 data-[state=on]:bg-background data-[state=on]:text-foreground data-[state=on]:shadow-sm">
              {localeLabels[locale].short}
            </ToggleGroupItem>
          )
        })}
    </ToggleGroup>
  )
}
