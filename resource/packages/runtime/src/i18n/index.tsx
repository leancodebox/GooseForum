"use client";
import { useEffect, useState, type ReactNode } from "react";
import i18next, { type i18n, type Resource } from "i18next";
import { I18nextProvider, initReactI18next } from "react-i18next";
import { supportedLocales, type Locale } from "@gooseforum/client/i18n/locale";
import {
  cachedGooseResources,
  loadedGooseNamespaces,
  prepareGooseTranslations,
  subscribeGooseResources,
} from "./loader";
export {
  gooseNamespaces,
  commonNamespaces,
  goosePageNamespaces,
  loadedGooseNamespaces,
  prepareGooseTranslations,
} from "./loader";
export const defaultNamespace = "auth";

export function createGooseI18n(
  locale: Locale,
  initialResources: Resource = cachedGooseResources(),
): i18n {
  const instance = i18next.createInstance();
  void instance.use(initReactI18next).init({
    lng: locale,
    fallbackLng: "zh",
    supportedLngs: [...supportedLocales],
    load: "languageOnly",
    defaultNS: defaultNamespace,
    ns: [],
    resources: initialResources,
    interpolation: { escapeValue: false, prefix: "{", suffix: "}" },
    initAsync: false,
    react: { useSuspense: false, bindI18nStore: "added" },
  });
  return instance;
}

export function GooseI18nProvider({
  locale,
  initialResources,
  children,
}: {
  locale: Locale;
  initialResources?: Resource;
  children: ReactNode;
}) {
  const [instance] = useState(() =>
    createGooseI18n(locale, initialResources ?? cachedGooseResources()),
  );
  useEffect(() => {
    if (!initialResources) return;
    for (const [language, namespaces] of Object.entries(initialResources)) {
      for (const [namespace, dictionary] of Object.entries(namespaces)) {
        instance.addResourceBundle(
          language,
          namespace,
          dictionary,
          true,
          true,
        );
      }
    }
  }, [initialResources, instance]);
  useEffect(
    () =>
      subscribeGooseResources((language, namespace, dictionary) => {
        if (!instance.hasResourceBundle(language, namespace))
          instance.addResourceBundle(language, namespace, dictionary);
      }),
    [instance],
  );
  useEffect(() => {
    let cancelled = false;
    void prepareGooseTranslations(locale, loadedGooseNamespaces())
      .then(() => {
        if (!cancelled && instance.language !== locale)
          return instance.changeLanguage(locale);
      })
      .catch((error) =>
        console.warn(
          "Unable to prepare translations; retaining the current language.",
          error,
        ),
      );
    return () => {
      cancelled = true;
    };
  }, [instance, locale]);
  return <I18nextProvider i18n={instance}>{children}</I18nextProvider>;
}
