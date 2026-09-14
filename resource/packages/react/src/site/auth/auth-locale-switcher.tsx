"use client"

import { useTranslation } from 'react-i18next'
import { Tabs, TabsList, TabsTrigger } from '../../components/ui/tabs'
import { authLocales } from '../../i18n/auth'
import { useGooseRuntime } from '../../runtime'

export function AuthLocaleSwitcher() {
  const runtime = useGooseRuntime()
  const { i18n } = useTranslation('auth')

  return (
    <Tabs
      value={runtime.locale}
      onValueChange={(value) => runtime.setLocale(value as typeof runtime.locale)}
      className="absolute right-4 top-4"
    >
      <TabsList aria-label="Language">
        {authLocales.map((locale) => {
          const localeT = i18n.getFixedT(locale, 'auth')
          return (
            <TabsTrigger key={locale} value={locale} title={localeT('locale.label')}>
              {localeT('locale.short')}
            </TabsTrigger>
          )
        })}
      </TabsList>
    </Tabs>
  )
}
