"use client"

import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Check, Languages } from 'lucide-react'
import { Button } from '@gooseforum/ui/components/button'
import { Popover, PopoverContent, PopoverTrigger } from '@gooseforum/ui/components/popover'
import { authLocales, localeLabels } from '@gooseforum/runtime/i18n/auth'
import { useGooseRuntime } from '@gooseforum/runtime'

export function AuthLocaleSwitcher() {
  const runtime = useGooseRuntime()
  const { t } = useTranslation('shell')
  const [open, setOpen] = useState(false)

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button type="button" variant="ghost" size="icon" className="absolute right-4 top-4" aria-label={t('switchLanguage')}>
          <Languages />
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" aria-label={t('switchLanguage')} className="max-h-[min(20rem,var(--radix-popover-content-available-height))] w-44 gap-0 overflow-y-auto p-1">
        {authLocales.map(locale => (
          <Button key={locale} type="button" variant="ghost" lang={locale} aria-pressed={runtime.locale === locale}
            className="shrink-0 justify-between" onClick={() => {
              void runtime.setLocale(locale)
              setOpen(false)
            }}>
            {localeLabels[locale].label}
            {runtime.locale === locale && <Check aria-hidden="true" />}
          </Button>
        ))}
      </PopoverContent>
    </Popover>
  )
}
