"use client";

import { useEffect, useState, type ReactNode } from "react";
import i18next, { type i18n } from "i18next";
import { I18nextProvider, initReactI18next } from "react-i18next";
import {
  authResources,
  oidcConsentResources,
  siteResources,
  settingsResources,
  notificationResources,
  supportedLocales,
  type Locale,
} from "@gooseforum/client/i18n";

export const defaultNamespace = "auth";
const namespaces = [
  defaultNamespace,
  "oidcConsent",
  "shell",
  "links",
  "sponsors",
  "categories",
  "members",
  "home",
  "search",
  "user",
  "settings",
  "userCard",
  "notifications",
];

export function createGooseI18n(locale: Locale): i18n {
  const instance = i18next.createInstance();
  void instance.use(initReactI18next).init({
    lng: locale,
    fallbackLng: "zh",
    supportedLngs: [...supportedLocales],
    load: "languageOnly",
    defaultNS: defaultNamespace,
    ns: namespaces,
    resources: Object.fromEntries(
      supportedLocales.map((language) => [
        language,
        {
          [defaultNamespace]: authResources[language],
          oidcConsent: oidcConsentResources[language],
          shell: siteResources[language].shell,
          links: siteResources[language].links,
          sponsors: siteResources[language].sponsors,
          categories: siteResources[language].categories,
          members: siteResources[language].members,
          home: siteResources[language].home,
          search: siteResources[language].search,
          user: siteResources[language].user,
          settings: settingsResources[language],
          userCard: siteResources[language].userCard,
          notifications: notificationResources[language],
        },
      ]),
    ),
    interpolation: {
      escapeValue: false,
      prefix: "{",
      suffix: "}",
    },
    initAsync: false,
    react: { useSuspense: false },
  });
  return instance;
}

export function GooseI18nProvider({
  locale,
  children,
}: {
  locale: Locale;
  children: ReactNode;
}) {
  const [instance] = useState(() => createGooseI18n(locale));
  useEffect(() => {
    if (instance.language !== locale) void instance.changeLanguage(locale);
  }, [instance, locale]);
  return <I18nextProvider i18n={instance}>{children}</I18nextProvider>;
}
