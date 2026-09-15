"use client"

import { Tabs, TabsList, TabsTrigger } from '@gooseforum/ui/components/tabs'
import { authLocales, localeLabels } from '@gooseforum/runtime/i18n/auth'
import { useGooseRuntime } from '@gooseforum/runtime'

export function AuthLocaleSwitcher() {
  const runtime = useGooseRuntime()

  return (
    <Tabs
      value={runtime.locale}
      onValueChange={(value) => void runtime.setLocale(value as typeof runtime.locale)}
      className="absolute right-4 top-4"
    >
      <TabsList aria-label="Language">
        {authLocales.map((locale) => {
          return (
            <TabsTrigger key={locale} value={locale} title={localeLabels[locale].label}>
              {localeLabels[locale].short}
            </TabsTrigger>
          )
        })}
      </TabsList>
    </Tabs>
  )
}
