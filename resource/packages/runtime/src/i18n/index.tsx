"use client";
import { useEffect, useState, type ReactNode } from "react";
import type { i18n, Resource } from "i18next";
import { I18nextProvider, initReactI18next } from "react-i18next";
import type { Locale } from "@gooseforum/client/i18n/locale";
import { createI18n } from "@gooseforum/i18n/react-runtime";
import {
  goosePageNamespaces,
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
export { useServerErrorMessage } from "./server-error";
export function createGooseI18n(
  locale: Locale,
  initialResources: Resource = cachedGooseResources(),
): i18n {
  return createI18n(locale, initialResources, initReactI18next);
}

export function GooseI18nProvider({
  locale,
  component,
  initialResources,
  children,
}: {
  locale: Locale;
  component?: string;
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
    void prepareGooseTranslations(locale, component ? goosePageNamespaces(component) : loadedGooseNamespaces())
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
  }, [instance, locale, component]);
  return <I18nextProvider i18n={instance}>{children}</I18nextProvider>;
}
